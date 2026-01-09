## 十三、多微服务工作区与中心化 IDL 仓库实践（go.work）

本章总结一种与 cwgo 适配度很高的落地方式：**每个服务一个 go.mod**、仓库拆分、微服务之间用 **go.work 联调**，并通过一个中心化的 **IDL 仓库**管理 Protobuf/生成产物。

---

### 13.1 目标形态（推荐）

- **服务仓库拆分**：`gateway`、`user-svc`、`order-svc`、`agent-svc` 等各自独立仓库，各自维护 `go.mod`
- **中心化 IDL 仓库**：建议命名为 **`rpc-contracts`**
  - 存放 **原始 `.proto`**（source of truth）
  - 可选：存放 **生成的 Go 代码产物**（如 `pb.go` / `kitex_gen`），用于分发复用
- **统一网关**：网关使用 Hertz（HTTP），内部调用下游 **Kitex client**（RPC）
- **Agent 微服务**：允许同时提供 **RPC + HTTP**（同进程双协议）

---

### 13.2 `rpc-contracts` 仓库推荐结构

> 目标：路径稳定、版本可控、各服务可通过 submodule/固定 clone 路径引用。

建议目录：

```text
rpc-contracts/
├── proto/                  # 原始 .proto（source of truth）
│   ├── user/v1/user.proto
│   └── agent/v1/agent.proto
└── gen/                    # 可选：生成产物（分发复用）
    └── go/
        ├── pb/             # protoc-gen-go 输出
        └── kitex_gen/      # kitex 生成代码（如使用 kitex/protobuf 体系）
```

**约定建议**：
- 原始 `.proto` 与生成产物分目录，避免“手改生成代码”
- 对 `rpc-contracts` 打 tag（例如 `rpc-contracts/v1.3.0`），服务升级 IDL 时更新 submodule/tag 并重新生成

---

### 13.3 各服务仓库如何引用 IDL（推荐：git submodule）

在每个服务仓库固定一个路径（示例）：

```text
<service-repo>/
├── go.mod
└── third_party/
    └── rpc-contracts/      # git submodule 指向 rpc-contracts
```

这样 `-idl` / `-I` 路径稳定，不依赖 GOPATH 或 module cache。

---

### 13.4 go.work 联调（多仓库协作）

在任意一个“联调工作区”目录创建 `go.work`，把多个服务仓库加入：

```text
workspace/
├── go.work
├── gateway/        # repo clone
├── user-svc/       # repo clone
└── agent-svc/      # repo clone
```

`go.work` 示例（思路展示）：

```text
go 1.22

use ./gateway
use ./user-svc
use ./agent-svc
```

> 这样 gateway 引用本地的 client 包/下游模块时，会优先使用 `go.work` 中的本地代码，便于联调与迭代。

---

### 13.5 用 cwgo 生成服务（命令范式）

#### 13.5.1 生成 Kitex 微服务（RPC）

在服务仓库根目录执行：

```bash
cwgo server -type rpc \
  -module github.com/your-org/user-svc \
  -service user \
  -idl third_party/rpc-contracts/proto/user/v1/user.proto \
  -I third_party/rpc-contracts/proto
```

#### 13.5.2 生成 agent 微服务（RPC + HTTP 同进程）

使用 `--hex` 让 Kitex 服务同进程额外挂 HTTP（基于 Hertz 路由引擎）：

```bash
cwgo server -type rpc --hex \
  -module github.com/your-org/agent-svc \
  -service agent \
  -idl third_party/rpc-contracts/proto/agent/v1/agent.proto \
  -I third_party/rpc-contracts/proto
```

#### 13.5.3 生成网关（Hertz HTTP）

网关建议使用 Hertz 作为对外 HTTP 层：

```bash
cwgo server -type http \
  -module github.com/your-org/gateway \
  -service gateway \
  -idl third_party/rpc-contracts/proto/gateway/v1/gateway.proto \
  -I third_party/rpc-contracts/proto
```

#### 13.5.4 网关生成下游 Kitex client（推荐）

网关通过 client 调用下游 RPC：

```bash
cwgo client -type rpc \
  -module github.com/your-org/gateway \
  -service user \
  -idl third_party/rpc-contracts/proto/user/v1/user.proto \
  -I third_party/rpc-contracts/proto
```

---

### 13.6 基础设施（MySQL + Redis）

默认模板已包含 MySQL/Redis 的配置结构与 DAL 初始化文件（但是否在 `main` 中启用 DAL 初始化，通常需要你在模板或业务代码里接上）。

---

### 13.7 更新策略：可填充逻辑的文件一律 skip

你的目标通常是：
- **可填充业务逻辑的文件**：永不覆盖（`skip`）
- **纯生成产物**：允许覆盖或追加（`cover` / `append`）

在 cwgo 默认模板下：
- Kitex 侧大多数文件已采用 `skip`
- 个别文件可能是 `cover`（元信息）或 `append`（聚合式 handler）

如果你想更严格地控制覆盖行为，建议：
1. 复制标准模板到自定义目录（例如 `tpl/kitex/server/custom/`）
2. 将“你不希望被覆盖”的模板文件 `update_behavior.type` 统一改为 `skip`
3. 以后生成时统一加 `--template <自定义模板目录>`

---

### 13.8 Eino（AI）集成提醒

`--enable-eino` 当前会生成一个 `internal/agent/agent.go` 的占位实现（可作为 AI 能力模块的起点），但不会自动把该 agent 挂到 HTTP/RPC 的 handler/service 调用链上；你需要根据自己的接口协议在业务层完成接入。


