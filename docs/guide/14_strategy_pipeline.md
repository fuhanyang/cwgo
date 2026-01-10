配置驱动策略编排方案总结
========================

### 需求背景与问题抛出

在 Kitex server 标准模板的策略模式中，最初生成代码采用“每个方法一个全局 `map[string]Handler`”来存储策略，并依赖 `init()` + 空导入触发注册。这个实现能跑通 demo，但在真实工程中会暴露一组典型问题：

- **工程化不足**：策略注册与加载高度隐式（`init()` 副作用 + blank import），排查问题和阅读代码成本高。
- **扩展与治理困难**：只有一个 map，缺少统一的校验、并发安全约束、错误信息聚合、可观测性扩展点等。
- **配置化能力弱**：只能“写死一个 handler”，无法把策略编排表达为可配置的流程（例如 login/logout/addfriend 等策略由多个 processor 组成）。
- **灵活度不足**：无法在每次请求中按请求内容/上下文动态选择策略（灰度、AB、权限、租户等场景常见）。

### 现状梳理

对话开始时的模板（位于 `tpl/kitex/server/standard/`）主要形态如下：

- `service.yaml`
  - 每个方法生成 `{{Method}}Strategies map[string]{{Method}}Handler`
  - `Run()` 固定执行 `"default"` 策略
- `strategy.yaml`
  - 在 `init()` 中把 `"default"` handler 写入 map
- `strategy_init_tpl.yaml`
  - `package main` 里空导入 `biz/strategy/...`，通过 `init()` 完成注册

这套机制的核心特征是：**策略内容由代码写死，加载通过 init 隐式触发，运行时只能固定选 default**。

### 解决思路与方案演进

本次讨论的关键，是把“策略系统”从“一个 map”升级为可治理、可扩展、可配置的模块，并最终落到一个更贴近实际业务的模式：

1. **从 map 到标准 Registry**
   - 先把存储结构封装为一个独立模块（Registry），提供统一的 Register/查询/错误输出入口。
   - 目的：去隐式副作用，提升可测试性与扩展点（并发、观测、校验）。

2. **从“代码注册策略”到“配置驱动策略编排”**
   - 将 strategy 的“内容”从代码 handler 中剥离出来，用 conf 表达“pipeline = 处理器序列”。
   - 代码侧只做两件事：
     - 注册 processor（原子步骤）
     - 在启动时把配置 pipeline 编译为可执行 handler

3. **从“启动时固定 active 策略”到“每请求动态选策略”**
   - 你最终选择了更灵活的 B：由 service 层按每次请求动态选择 strategyName。
   - 这让生成代码天然支持灰度/AB/场景化流程，而不必改动框架底层。

### 经历的关键流程（最终落地的运行链路）

最终落地的关键流程可以概括为“注册 processor → 从每个服务的策略配置文件编译 strategy → 请求时选择并执行 strategy”：

1. **用户注册 processor（显式）**
   - 每个方法有一个 `{{Method}}Processors`（ProcessorRegistry），存储 `processorName -> processor`。
   - processor 支持两种风格：
     - **接口式**：用户自定义类型实现 `Name() string` + `Process(*State) error`，然后注册（显式、可复用、便于测试）。
     - **函数式**：用 `RegisterFunc(name, func(*State) error)` 直接注册函数（更轻量）。
   - 用户在生成的 `biz/strategy/<service>_strategy/` 下实现 `Register{{Method}}Processors()` 并注册具体 processor。

2. **启动时从每个服务的策略配置文件编译并注册策略（配置驱动）**
   - `main.go` 启动前调用 `initStrategies()`。
   - `initStrategies()` 读取 `biz/strategy/<service>_strategy/strategy.yaml`（或 `strategy.<env>.yaml`）：
     - 配置结构：`operationKey -> strategyName -> [processorName...]`
     - 对每条 pipeline 生成一个 handler：
       - 初始化 `{{Method}}State`
       - 按顺序执行 processor
       -（非 Void 方法）从 state 中取出 Resp 返回
     - 注册到 `{{Method}}Strategies`：`strategyName -> compiled handler`

3. **请求到来时动态选策略并执行**
   - `Run()` 调用 `Choose{{Method}}Strategy(...)` 获取本次请求的 strategyName（空则 fallback `default`）。
   - 用 `{{Method}}Strategies.Get(name)` 获取 handler 并执行。

### 关键知识点与设计要点

- **将“策略”定义为“可配置的 pipeline”**
  - strategy 不再是写死的 handler，而是“processor 的有序组合”。
  - 优点：业务流程可以在不改代码的情况下通过配置调整（前提是 processor 已注册）。

- **State mutation 模型**
  - 每个方法生成 `{{Method}}State`，用于在 processor 间传递请求数据、响应数据与中间变量（`Vars map[string]any`）。
  - 这比“纯函数式 processor(ctx,args…)”更适配编排场景：可以逐步构建 resp、写入中间结果、做跨步骤共享。

- **Registry 分层：processor registry 与 strategy registry**
  - `{{Method}}Processors`：名字到 processor 的映射（用户代码显式注册）
  - `{{Method}}Strategies`：名字到 handler 的映射（启动时由配置编译生成）
  - 这样避免了“策略和处理器混在一起”的边界模糊。

- **显式初始化优先**
  - 生成代码尽量避免 `init()` 副作用：注册发生在明确的 `initStrategies()` 执行路径中，便于排错与测试。

- **模板渲染中的空白控制**
  - Go template 的 `{{-`/`-}}` 会裁剪换行与空格；在 Go 源码生成里不当使用可能导致注释/导入/代码块被“粘连”而产生语法问题。
  - 因此在关键逻辑块处应尽量使用非裁剪形式 `{{ ... }}` 或谨慎控制裁剪范围。

### 结果与当前能力

本次对 Kitex server 标准模板的策略系统，最终实现了这些能力：

- **配置定义策略内容**：策略由每个服务的 `biz/strategy/<service>_strategy/strategy*.yaml` 描述 processor pipeline，而非写死在代码 handler 中。
- **processor 可复用**：同一 processor 可被多个 strategy 复用，减少重复逻辑。
- **service 动态选策略**：每次请求可按业务/灰度/AB/权限等条件动态选择 strategyName。
- **清晰的失败报错**：当 processor/strategy 缺失时，错误信息可包含 method、strategyName、缺失项及可用列表。

涉及的核心模板文件包括（均在 `tpl/kitex/server/standard/`）：

- `strategy_registry.yaml`：Registry + ProcessorRegistry（Get/Names/并发安全等）
- `strategy_config_loader.yaml`：从策略目录加载 `strategy.yaml` / `strategy.<env>.yaml`
- `strategy_service_config.yaml`：为每个服务生成 `LoadStrategyConfig()`（定位包目录并读取策略 YAML）
- `strategy_service_yaml.yaml`：为每个服务生成默认 `strategy.yaml` 骨架
- `service.yaml`：生成 State/Processor/Processors + 动态 ChooseStrategy + `Strategies.Get`
- `strategy_init_tpl.yaml`：从 per-service 策略 YAML 编译 pipeline 并注册策略；不再依赖 SetActive
- `strategy.yaml`：改为 processor 注册引导（不再注册 handler）

### 功能展望与可继续演进方向

- **分支与条件编排**：当前 pipeline 为线性列表；可扩展为带条件、并行、短路、重试、降级等能力。
- **可观测性增强**：在编译出的 handler 中自动埋点（每个 processor 的耗时、错误率、trace span）。
- **热更新策略**：支持配置变更触发策略重编译（需要并发与一致性设计）。
- **类型更强的中间结果**：`Vars map[string]any` 可升级为更结构化的 state 或 codegen 的 typed slots，减少运行时断言。

