## 二、整体架构

### 2.1 项目目录结构

```
cwgo/
├── cwgo.go                          # 程序入口，初始化插件模式
├── cmd/                             # CLI 命令定义
│   └── static/                      # 静态命令配置
│       ├── cmd.go                   # 命令路由器（核心）
│       ├── server_flags.go          # server 命令参数
│       ├── client_flags.go          # client 命令参数
│       ├── model_flags.go           # model 命令参数
│       ├── doc_flags.go             # doc 命令参数
│       ├── job_flags.go             # job 命令参数
│       └── api_list_flags.go        # api_list 命令参数
├── config/                          # 配置结构体定义
│   ├── argument.go                  # 通用参数结构
│   ├── server.go                    # 服务器配置
│   ├── client.go                    # 客户端配置
│   ├── model.go                     # 模型配置
│   ├── doc.go                       # 文档配置
│   └── job.go                       # 任务配置
├── pkg/                             # 核心业务逻辑
│   ├── server/                      # 服务器代码生成
│   │   ├── server.go                # 服务器生成入口
│   │   ├── kitex.go                 # Kitex RPC 生成逻辑
│   │   ├── hz.go                    # Hertz HTTP 生成逻辑
│   │   └── check.go                 # 参数校验
│   ├── client/                      # 客户端代码生成
│   │   ├── client.go                # 客户端生成入口
│   │   ├── kitex.go                 # RPC 客户端生成
│   │   └── hz.go                    # HTTP 客户端生成
│   ├── model/                       # 数据库模型生成
│   │   ├── model.go                 # 模型生成核心逻辑
│   │   └── check.go                 # 参数校验
│   ├── config_generator/            # 配置文件转代码
│   │   ├── sdk.go                   # 配置生成主逻辑
│   │   ├── yaml2go.go               # YAML 转 Go 结构体
│   │   └── metadata.go              # 元数据定义
│   ├── curd/                        # CRUD 代码生成
│   │   └── doc/                     # MongoDB 支持
│   ├── api_list/                    # Hertz 路由分析工具
│   ├── job/                         # 批处理任务生成
│   ├── common/                      # 通用工具
│   │   ├── utils/                   # 工具函数
│   │   └── kx_registry/             # Kitex 注册表处理
│   └── consts/                      # 常量定义
├── tpl/                             # 代码模板（核心）
│   ├── init.go                      # 模板初始化
│   ├── kitex/                       # Kitex RPC 模板
│   │   ├── server/                  # 服务器端模板
│   │   │   └── standard/            # 标准布局
│   │   │       ├── main_tpl.yaml    # 主程序入口
│   │   │       ├── handler_tpl.yaml # Handler 实现
│   │   │       ├── service.yaml     # 业务服务层
│   │   │       ├── conf_tpl.yaml    # 配置结构
│   │   │       ├── conf_*_tpl.yaml  # 多环境配置
│   │   │       ├── dal_init.yaml    # DAL 初始化
│   │   │       ├── kitex_yaml.yaml  # Kitex 配置
│   │   │       ├── mysql.yaml       # MySQL 配置
│   │   │       ├── redis.yaml       # Redis 配置
│   │   │       ├── bootstrap_sh_tpl.yaml  # 启动脚本
│   │   │       ├── build_sh_tpl.yaml       # 构建脚本
│   │   │       ├── docker_compose.yaml     # Docker 配置
│   │   │       └── readme_tpl.yaml  # 项目文档
│   │   └── client/                  # 客户端模板
│   │       └── standard/
│   │           ├── client_tpl.yaml  # 客户端接口
│   │           ├── default_tpl.yaml # 默认实现
│   │           └── init_tpl.yaml    # 初始化代码
│   └── hertz/                        # Hertz HTTP 模板
│       ├── server/                  # 服务器端模板
│       │   ├── standard/            # 标准版本
│       │   │   ├── layout.yaml      # 项目布局
│       │   │   └── package.yaml     # 包配置
│       │   └── standard_v2/         # V2 版本
│       └── client/                  # 客户端模板
├── hack/                            # 构建脚本
│   ├── tools.sh                     # 测试和检查脚本
│   ├── util.sh                      # 工具函数
│   └── resolve-modules.sh           # 模块解析
├── Makefile                         # 构建配置
└── go.mod                           # Go 模块定义
```

### 2.2 架构设计原则

1. **模块化设计**：各组件职责清晰，低耦合高内聚
2. **模板驱动**：使用 YAML 模板灵活控制代码生成
3. **插件化架构**：通过插件模式与 Kitex/Hz 集成
4. **配置化**：支持丰富的配置选项
5. **多框架支持**：统一的接口适配不同框架

---

