## 六、定制化开发指南

### 6.1 修改生成的代码结构（最常见）

#### 场景 1：为服务添加统一的错误处理

**修改文件**：`tpl/kitex/server/standard/service.yaml`

```yaml
# 原始模板
body: |-
  func (s *{{.Name}}Service) Run({{range .Args}}{{LowerFirst .Name}} {{.Type}}, {{end}}) (resp {{.Resp.Type}}, err error) {
    // Finish your business logic.
    return
  }

# 修改后：添加统一错误处理
body: |-
  import (
    "context"
    "your-project/pkg/errors"
    "github.com/cloudwego/hertz/pkg/common/hlog"
  )

  func (s *{{.Name}}Service) Run({{range .Args}}{{LowerFirst .Name}} {{.Type}}, {{end}}) (resp {{.Resp.Type}}, err error) {
    // 添加 panic 恢复
    defer func() {
      if r := recover(); r != nil {
        err = errors.NewPanicError(r)
        hlog.Errorf("panic in {{.Name}}: %v", r)
      }
    }()

    // 添加方法进入日志
    hlog.Infof("{{.Name}} called with args: %+v", {{range $i, $arg := .Args}}{{if $i}}, {{end}}{{$arg.Name}}{{end}})

    // Finish your business logic.

    // 添加方法返回日志
    hlog.Infof("{{.Name}} returned: resp=%+v, err=%v", resp, err)
    return
  }
```

#### 场景 2：添加数据库事务支持

**修改文件**：`tpl/kitex/server/standard/service.yaml`

```yaml
body: |-
  import (
    "context"
    "gorm.io/gorm"
  )

  type {{.Name}}Service struct {
    ctx    context.Context
    db     *gorm.DB
  }

  func New{{.Name}}Service(ctx context.Context, db *gorm.DB) *{{.Name}}Service {
    return &{{.Name}}Service{ctx: ctx, db: db}
  }

  func (s *{{.Name}}Service) Run({{range .Args}}{{LowerFirst .Name}} {{.Type}}, {{end}}) (resp {{.Resp.Type}}, err error) {
    // 使用事务
    tx := s.db.Begin()
    defer func() {
      if r := recover(); r != nil {
        tx.Rollback()
        panic(r)
      } else if err != nil {
        tx.Rollback()
      } else {
        tx.Commit()
      }
    }()

    // 使用 tx 执行业务逻辑...

    return
  }
```

#### 场景 3：修改 Hertz 主程序结构

**修改文件**：`tpl/hertz/server/standard/layout.yaml`

```yaml
# 添加优雅关闭
body: |-
  import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"
  )

  func main() {
    address := conf.GetConf().Hertz.Address
    h := server.New(server.WithHostPorts(address))

    registerMiddleware(h)
    router.GeneratedRegister(h)

    // 优雅关闭
    go func() {
      if err := h.Spin(); err != nil {
        hlog.Errorf("Hertz server error: %v", err)
      }
    }()

    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    hlog.Info("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := h.Shutdown(ctx); err != nil {
      hlog.Errorf("Server forced to shutdown: %v", err)
    }

    hlog.Info("Server exited")
  }
```

### 6.2 添加新的命令参数

#### 步骤 1：修改配置结构

**文件**：`config/server.go`

```go
type ServerArgument struct {
    *CommonParam

    // 现有字段...
    Template   string
    Branch     string
    Verbose    bool
    Hex        bool

    // 添加你的自定义字段
    EnableCache      bool    // 启用缓存
    CacheType        string  // 缓存类型（redis/memory）
    EnableTracing    bool    // 启用链路追踪
    TracingExporter  string  // 追踪导出器
    EnableMetrics    bool    // 启用指标
    MetricsPort      int     // 指标端口
}
```

#### 步骤 2：添加 CLI 标志

**文件**：`cmd/static/server_flags.go`

```go
func serverFlags() []cli.Flag {
    return []cli.Flag{
        // 现有标志...

        // 添加你的标志
        &cli.BoolFlag{
            Name:  "enable-cache",
            Usage: "Enable cache in generated code",
            Value: false,
        },
        &cli.StringFlag{
            Name:  "cache-type",
            Usage: "Cache type: redis or memory",
            Value: "memory",
        },
        &cli.BoolFlag{
            Name:  "enable-tracing",
            Usage: "Enable distributed tracing",
            Value: false,
        },
        &cli.StringFlag{
            Name:  "tracing-exporter",
            Usage: "Tracing exporter: jaeger, zipkin, stdout",
            Value: "stdout",
        },
        &cli.BoolFlag{
            Name:  "enable-metrics",
            Usage: "Enable metrics collection",
            Value: false,
        },
        &cli.IntFlag{
            Name:  "metrics-port",
            Usage: "Metrics server port",
            Value: 9090,
        },
    }
}
```

#### 步骤 3：在生成逻辑中使用

**文件**：`pkg/server/kitex.go` 或 `hz.go`

```go
func convertKitexArgs(sa *config.ServerArgument, kitexArgument *kargs.Arguments) error {
    // ... 现有代码 ...

    // 根据配置设置不同的生成选项
    if sa.EnableCache {
        // 设置缓存相关的模板变量
        // 或者选择不同的模板目录
    }

    if sa.EnableTracing {
        // 添加追踪相关的导入和代码
    }

    return nil
}
```

#### 步骤 4：在模板中使用

**文件**：`tpl/kitex/server/standard/service.yaml`

```yaml
body: |-
  import (
    "context"
    {{- if .EnableCache }}
    "github.com/go-redis/redis/v8"
    {{- end}}
    {{- if .EnableTracing }}
    "go.opentelemetry.io/otel"
    {{- end}}
  )

  type {{.Name}}Service struct {
    ctx context.Context
    {{- if .EnableCache }}
    cache *redis.Client
    {{- end}}
  }

  func New{{.Name}}Service(ctx context.Context{{if .EnableCache}}, cache *redis.Client{{end}}) *{{.Name}}Service {
    return &{{.Name}}Service{
      ctx: ctx{{if .EnableCache}},
      cache: cache{{end}}
    }
  }

  func (s *{{.Name}}Service) Run({{range .Args}}{{LowerFirst .Name}} {{.Type}}, {{end}}) (resp {{.Resp.Type}}, err error) {
    {{- if .EnableTracing}}
    ctx, span := otel.Tracer("service").Start(s.ctx, "{{.Name}}")
    defer span.End()
    {{- end}}

    // 业务逻辑...

    return
  }
```

### 6.3 创建全新的模板布局

#### ⚠️ 重要提示

在创建自定义模板时，**最关键的要点**是：

1. **必须保留 `FilterImports` 动态导入逻辑**，否则生成的代码无法编译
2. **先复制原始模板，再进行修改**，避免遗漏关键部分
3. **理解模板变量的作用**，正确使用模板语法

#### 步骤 1：创建新模板目录

```bash
# 创建自定义布局目录
mkdir -p tpl/kitex/server/my_custom_layout

# 或者基于标准模板复制
cp -r tpl/kitex/server/standard tpl/kitex/server/my_custom_layout
```

#### 步骤 2：创建正确的 service.yaml 模板

**文件**：`tpl/kitex/server/my_custom_layout/service.yaml`

```yaml
path: internal/service/{{SnakeString .ServiceName}}/{{ SnakeString (index .Methods 0).Name }}.go
loop_method: true
update_behavior:
  type: skip
body: |-
  package {{SnakeString .ServiceName}}

  import (
    "context"

    {{- if .EnableCustomLogger }}
    "{{.GoModule}}/pkg/logger"
    {{- end}}
    {{- if .EnableCustomErrors }}
    "{{.GoModule}}/pkg/errors"
    {{- end}}

    {{- /* ⚠️ 关键：保留动态导入逻辑，否则会编译失败 */}}
    {{- range $path, $aliases := ( FilterImports .Imports .Methods )}}
        {{- if not $aliases }}
            "{{$path}}"
        {{- else if or (eq $path "github.com/cloudwego/kitex/client") (eq $path "github.com/cloudwego/kitex/pkg/serviceinfo")}}
        {{- else}}
            {{- range $alias, $is := $aliases}}
                {{$alias}} "{{$path}}"
            {{- end}}
        {{- end}}
    {{- end}}
  )

  {{range .Methods}}

  type {{.Name}}Service struct {
    ctx context.Context
    {{- if .EnableCustomLogger }}
    log *logger.Logger
    {{- end}}
  }

  {{- if or .ClientStreaming .ServerStreaming}}

  // New{{.Name}}Service new {{.Name}}Service
  func New{{.Name}}Service(ctx context.Context{{if .EnableCustomLogger}}, log *logger.Logger{{end}}) *{{.Name}}Service {
    return &{{.Name}}Service{
      ctx: ctx{{if .EnableCustomLogger}},
      log: log{{end}}
    }
  }

  func (s *{{.Name}}Service) Run({{if not .ClientStreaming}}{{range .Args}}{{LowerFirst .Name}} {{.Type}}, {{end}}{{end}}stream {{.PkgRefName}}.{{.ServiceName}}_{{.RawName}}Server) (err error) {
    {{- if .EnableCustomLogger }}
    s.log.Infof("[{{.Name}}] Service called (streaming)")
    {{- end}}

    defer func() {
      {{- if .EnableCustomLogger }}
      if r := recover(); r != nil {
        s.log.Errorf("[{{.Name}}] panic: %v", r)
        panic(r)
      }
      {{- end}}
    }()

    // Finish your business logic.
    return
  }
  {{- else}}
  {{- if .Void}}

  func New{{.Name}}Service(ctx context.Context{{if .EnableCustomLogger}}, log *logger.Logger{{end}}) *{{.Name}}Service {
    return &{{.Name}}Service{
      ctx: ctx{{if .EnableCustomLogger}},
      log: log{{end}}
    }
  }

  func (s *{{.Name}}Service) Run({{range .Args}}{{LowerFirst .Name}} {{.Type}}, {{end}}) error {
    {{- if .EnableCustomLogger }}
    s.log.Infof("[{{.Name}}] Service called with args: %+v", {{range $i, $arg := .Args}}{{if $i}}, {{end}}{{$arg.Name}}{{end}})
    {{- end}}

    defer func() {
      {{- if .EnableCustomLogger }}
      if r := recover(); r != nil {
        s.log.Errorf("[{{.Name}}] panic: %v", r)
        panic(r)
      }
      {{- end}}
    }()

    // Finish your business logic.

    return nil
  }
  {{else}}

  func New{{.Name}}Service(ctx context.Context{{if .EnableCustomLogger}}, log *logger.Logger{{end}}) *{{.Name}}Service {
    return &{{.Name}}Service{
      ctx: ctx{{if .EnableCustomLogger}},
      log: log{{end}}
    }
  }

  func (s *{{.Name}}Service) Run({{range .Args}}{{LowerFirst .Name}} {{.Type}}, {{end}}) (resp {{.Resp.Type}}, err error) {
    {{- if .EnableCustomLogger }}
    s.log.Infof("[{{.Name}}] Service called with args: %+v", {{range $i, $arg := .Args}}{{if $i}}, {{end}}{{$arg.Name}}{{end}})

    defer func() {
      if r := recover(); r != nil {
        {{- if .EnableCustomErrors }}
        err = errors.NewPanicError(r)
        {{- end}}
        s.log.Errorf("[{{.Name}}] panic: %v", r)
        panic(r)
      } else if err != nil {
        s.log.Errorf("[{{.Name}}] failed: %v", err)
        {{- if .EnableCustomErrors }}
        err = errors.Wrap(err, "{{.Name}} failed")
        {{- end}}
      } else {
        s.log.Infof("[{{.Name}}] succeeded")
      }
    }()
    {{- end}}

    // Finish your business logic.

    return
  }
  {{end}}
  {{end}}
  {{end}}
```

#### 📊 模板关键要素说明

| 要素 | 说明 | 错误示例 ✗ | 正确示例 ✓ |
|------|------|-----------|----------|
| **动态导入** | 使用 `FilterImports` 自动导入 IDL 类型 | `import ("context")` | `{{range $path, $aliases := ( FilterImports .Imports .Methods )}}` |
| **路径变量** | 使用正确的模板变量 | `path: service.go` | `path: internal/service/{{SnakeString .ServiceName}}/...` |
| **条件渲染** | 根据配置条件生成代码 | `log *logger.Logger` | `{{if .EnableCustomLogger}}log *logger.Logger{{end}}` |
| **循环方法** | 为每个方法生成独立文件 | 删除 `loop_method` | `loop_method: true` |
| **更新行为** | 控制文件覆盖策略 | `type: cover` | `type: skip` |

#### 步骤 3：创建配套的 main 模板

**文件**：`tpl/kitex/server/my_custom_layout/main_tpl.yaml`

```yaml
path: cmd/server/main.go
body: |-
  package main

  import (
    "context"
    "flag"

    {{- if .EnableCustomConfig }}
    "{{.GoModule}}/pkg/config"
    {{- end}}
    {{- if .EnableCustomLogger }}
    "{{.GoModule}}/pkg/logger"
    {{- end}}

    "{{.GoModule}}/kitex_gen/{{SnakeString .ServiceName}}/{{.ServiceName}}"
    "{{.GoModule}}/internal/service/{{SnakeString .ServiceName}}"
  )

  var (
    {{- if .EnableCustomConfig }}
    configFile = flag.String("c", "configs/config.yaml", "config file path")
    {{- end}}
  )

  func main() {
    flag.Parse()

    {{- if .EnableCustomConfig }}
    // 初始化配置
    cfg, err := config.Load(*configFile)
    if err != nil {
      panic(err)
    }

    // 初始化日志
    {{- if .EnableCustomLogger }}
    log := logger.New(cfg.Log)
    {{- end}}
    {{- end}}

    // 创建服务实现
    {{if .EnableCustomLogger}}
    svr := {{.ServiceName}}.NewServer(new({{SnakeString .ServiceName}}.{{.ServiceName}}ServiceImpl))
    {{- else}}
    svr := {{.ServiceName}}.NewServer(new({{SnakeString .ServiceName}}.{{.ServiceName}}ServiceImpl))
    {{- end}}

    // 启动服务
    if err := svr.Run(); err != nil {
      panic(err)
    }
  }
```

#### 步骤 4：注册模板（可选）

如果不想通过 `-template` 参数指定，可以在 `tpl/init.go` 中注册：

```go
// ⚠️ 注意：cwgo 的模板系统通过 -template 参数动态指定
// 不需要修改 tpl/init.go，直接通过命令行参数使用即可
```

#### 步骤 5：使用自定义模板

```bash
# 方式 1：通过 -template 参数指定（推荐）
cwgo server -type rpc \
  -module github.com/your/project \
  -service user.service \
  -idl idl/user.thrift \
  -template tpl/kitex/server/my_custom_layout

# 方式 2：使用配置文件启用自定义功能
cwgo server -type rpc \
  -module github.com/your/project \
  -service user.service \
  -idl idl/user.thrift \
  -enable-custom-logger \
  -enable-custom-errors \
  -template tpl/kitex/server/my_custom_layout
```

#### 步骤 6：添加配置参数支持

为了让模板中的条件变量生效，需要添加对应的配置参数：

**文件**：`config/server.go`

```go
type ServerArgument struct {
    *CommonParam

    Template   string
    Branch     string
    Verbose    bool
    Hex        bool

    // 添加自定义功能开关
    EnableCustomLogger  bool  // 是否生成带日志的代码
    EnableCustomErrors  bool  // 是否使用自定义错误包
    EnableCustomConfig  bool  // 是否使用自定义配置
}
```

**文件**：`cmd/static/server_flags.go`

```go
func serverFlags() []cli.Flag {
    return []cli.Flag{
        // 现有标志...

        // 自定义功能标志
        &cli.BoolFlag{
            Name:  "enable-custom-logger",
            Usage: "Generate code with custom logger integration",
            Value: false,
        },
        &cli.BoolFlag{
            Name:  "enable-custom-errors",
            Usage: "Generate code with custom error wrapping",
            Value: false,
        },
        &cli.BoolFlag{
            Name:  "enable-custom-config",
            Usage: "Generate code with custom config loading",
            Value: false,
        },
    }
}
```

#### 🎯 完整示例对比

**原始模板生成**：
```go
package userservice

import (
    "context"
    // ... IDL 相关导入
)

type GetUserService struct {
    ctx context.Context
}

func NewGetUserService(ctx context.Context) *GetUserService {
    return &GetUserService{ctx: ctx}
}

func (s *GetUserService) Run(req Request) (resp *Response, err error) {
    // Finish your business logic.
    return
}
```

**自定义模板生成（启用日志和错误处理）**：
```go
package userservice

import (
    "context"
    "github.com/your/project/pkg/logger"
    "github.com/your/project/pkg/errors"
    // ... IDL 相关导入（自动生成）
)

type GetUserService struct {
    ctx context.Context
    log *logger.Logger
}

func NewGetUserService(ctx context.Context, log *logger.Logger) *GetUserService {
    return &GetUserService{
        ctx: ctx,
        log: log,
    }
}

func (s *GetUserService) Run(req Request) (resp *Response, err error) {
    s.log.Infof("[GetUser] Service called with args: %+v", req)

    defer func() {
      if r := recover(); r != nil {
        err = errors.NewPanicError(r)
        s.log.Errorf("[GetUser] panic: %v", r)
        panic(r)
      } else if err != nil {
        s.log.Errorf("[GetUser] failed: %v", err)
        err = errors.Wrap(err, "GetUser failed")
      } else {
        s.log.Infof("[GetUser] succeeded")
      }
    }()

    // Finish your business logic.
    return
}
```

#### 💡 最佳实践

1. **渐进式定制**：
   ```bash
   # 步骤 1：先复制标准模板
   cp -r tpl/kitex/server/standard tpl/kitex/server/custom

   # 步骤 2：修改单个文件测试
   vim tpl/kitex/server/custom/service.yaml

   # 步骤 3：测试生成
   ./cwgo server -type rpc -template tpl/kitex/server/custom ...

   # 步骤 4：验证编译
   cd output && go build
   ```

2. **版本控制**：
   ```bash
   # 使用 Git 管理自定义模板
   git add tpl/kitex/server/custom/
   git commit -m "Add custom template with logging"
   ```

3. **文档化**：
   ```markdown
   # 自定义模板说明
   ## 路径
   tpl/kitex/server/my_custom_layout/

   ## 特性
   - 集成统一日志
   - 错误包装
   - Panic 恢复
   - 方法调用追踪

   ## 使用
   cwgo server -template tpl/kitex/server/my_custom_layout ...
   ```

#### ⚠️ 常见错误

| 错误现象 | 原因 | 解决方法 |
|---------|------|---------|
| `undefined: Request` | 缺少动态导入 | 保留 `FilterImports` 逻辑 |
| `undefined: logger` | 未启用自定义日志 | 添加条件判断或导入包 |
| 模板变量不生效 | 拼写错误或未传递 | 检查变量名和配置参数 |
| 生成路径错误 | path 模板语法错误 | 检查模板变量和语法 |

### 6.4 添加新的生成命令

#### 场景：创建一个完整的微服务生成命令

**步骤 1**：定义新命令

**文件**：`cmd/static/cmd.go`

```go
const (
    ServerName   = "server"
    ClientName   = "client"
    ModelName    = "model"
    DocName      = "doc"
    JobName      = "job"
    ApiListName  = "api_list"
    MicroServiceName = "microservice"  // 新增
)

func Init() *cli.App {
    // ...

    app.Commands = []*cli.Command{
        // 现有命令...

        {
            Name:  MicroServiceName,
            Usage: "Generate a complete microservice with server, client, and models",
            Flags: microserviceFlags(),
            Action: func(c *cli.Context) error {
                err := globalArgs.MicroServiceArgument.ParseCli(c)
                if err != nil {
                    return err
                }
                return microservice.Generate(globalArgs.MicroServiceArgument)
            },
        },
    }

    return app
}
```

### 6.5 高级定制：Service 编排与 Processor 算子化

#### 6.5.1 核心理念

为了解决复杂业务场景下 Service 层代码臃肿、难以测试的问题，推荐采用**逻辑编排与算子化**的设计模式。

**注意**：此模式主要适用于承载核心业务逻辑的 **Kitex (RPC)** 服务。对于 **Hertz (HTTP)** 服务，通常推荐采用 API Gateway 模式，直接调用下游 RPC 服务，保持轻量级，不需要复杂的本地编排。

核心思想是将 Kitex Service 层做薄，仅负责策略分发；将业务逻辑下沉为无状态的原子算子（Processor）；通过独立的 Strategy 层在 `init` 阶段完成逻辑编排（串行或 DAG）。

#### 6.5.2 架构分层 (Kitex)

| 层级 | 职责 | 生成策略 | 说明 |
|------|------|----------|------|
| **Service 层** | **分发** | `Skip` (推荐) | 定义接口标准和全局策略 Map，运行时查表执行。保留 `Skip` 以允许用户编写动态路由逻辑。 |
| **Strategy 层** | **编排** | `Skip` (不覆盖) | 用户在此编写 `init` 函数，将 Processor 算子组装成执行流，注册到 Map 中。 |
| **Processor 层** | **执行** | 仅生成 Doc | 用户自研的原子业务逻辑，纯 Go 代码，与框架解耦。 |

#### 6.5.3 目录结构示例

```text
biz/
├── service/              # [自动生成] Service 壳，负责分发
│   └── user/
│       └── get_user.go   
├── strategy/             # [自动生成一次] 策略编排层，用户在此处由 init 组装 Processor
│   └── user/
│       └── get_user.go   
└── processor/            # [用户自研] 原子业务逻辑，无代码生成干扰
    └── doc.go            # 仅生成文档占位
    └── user/
        ├── check_permission.go
        └── query_db.go
```

#### 6.5.4 实现代码示例

**1. Service 层（生成的壳）**

```go
// biz/service/user/get_user.go
package user_service

import (
    "context"
    "errors"
    user "example/kitex_gen/user" 
)

// 定义统一的 Handler 签名
type GetUserHandler func(ctx context.Context, req *user.GetUserRequest) (resp *user.GetUserResponse, err error)

// 全局策略 Map，用于存储编排好的逻辑
var GetUserStrategies = make(map[string]GetUserHandler)

type GetUserService struct {
    ctx context.Context
}

// Run 方法负责查表执行
func (s *GetUserService) Run(req *user.GetUserRequest) (resp *user.GetUserResponse, err error) {
    // --- 路由选择逻辑 (用户可扩展) ---
    strategyName := "default" 
    // if req.Type == "vip" { strategyName = "vip_channel" }
    
    if handler, ok := GetUserStrategies[strategyName]; ok {
        return handler(s.ctx, req)
    }
    
    return nil, errors.New("strategy not found")
}
```

**2. Strategy 层（编排入口）**

```go
// biz/strategy/user/get_user.go
package user_strategy

import (
    "context"
    "example/biz/service/user"             // 导入 Service 定义
    "example/biz/processor/user_processor" // 导入 Processor 算子
)

func init() {
    // 在 init 中注册具体的编排逻辑
    user_service.GetUserStrategies["default"] = func(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
        
        // --- 编排逻辑 (串行示例) ---
        
        // 1. 算子 A: 检查权限
        if err := user_processor.CheckPermission(ctx, req.UserId); err != nil {
            return nil, err
        }
        
        // 2. 算子 B: 查询数据
        userInfo, err := user_processor.QueryUserInfo(ctx, req.UserId)
        if err != nil {
            return nil, err
        }
        
        return &user.GetUserResponse{User: userInfo}, nil
    }
}
```

**3. Processor 层（原子算子）**

```go
// biz/processor/user/check_permission.go
package user_processor

// 纯函数，易于测试和复用
func CheckPermission(ctx context.Context, userID int64) error {
    // 具体实现...
    return nil
}
```

#### 6.5.5 Hertz 推荐架构 (Gateway 模式)

对于 Hertz 服务，建议保持轻量级：

*   **Handler**: 直接调用 RPC Client。
*   **Service**: 可选，仅作简单封装。
*   **Processor**: 不推荐在 Gateway 层实现复杂逻辑。

这样形成了 **Hertz (网关) -> Kitex (业务编排)** 的清晰分层。

**步骤 2**：定义配置结构

**文件**：`config/microservice.go`

```go
package config

import "github.com/urfave/cli/v2"

type MicroServiceArgument struct {
    *CommonParam

    // 数据库配置
    EnableDatabase bool
    DbType         string
    DSN            string
    Tables         []string

    // 缓存配置
    EnableCache bool
    CacheType   string

    // 追踪配置
    EnableTracing   bool
    TracingExporter string

    // 指标配置
    EnableMetrics bool
    MetricsPort   int
}

func NewMicroServiceArgument() *MicroServiceArgument {
    return &MicroServiceArgument{
        CommonParam: &CommonParam{},
    }
}

func (m *MicroServiceArgument) ParseCli(ctx *cli.Context) error {
    // 解析参数
    m.ServerName = ctx.String("service")
    m.GoMod = ctx.String("module")
    m.IdlPath = ctx.String("idl")
    m.OutDir = ctx.String("out-dir")

    m.EnableDatabase = ctx.Bool("enable-database")
    m.DbType = ctx.String("db-type")
    m.DSN = ctx.String("dsn")

    // ... 其他参数

    return nil
}
```

**步骤 3**：实现生成逻辑

**文件**：`pkg/microservice/generator.go`

```go
package microservice

import (
    "github.com/cloudwego/cwgo/config"
    "github.com/cloudwego/cwgo/pkg/client"
    "github.com/cloudwego/cwgo/pkg/model"
    "github.com/cloudwego/cwgo/pkg/server"
)

func Generate(c *config.MicroServiceArgument) error {
    // 1. 生成服务器
    serverArg := convertToServerArg(c)
    if err := server.Server(serverArg); err != nil {
        return err
    }

    // 2. 生成客户端
    clientArg := convertToClientArg(c)
    if err := client.Client(clientArg); err != nil {
        return err
    }

    // 3. 生成数据库模型
    if c.EnableDatabase {
        modelArg := convertToModelArg(c)
        if err := model.Model(modelArg); err != nil {
            return err
        }
    }

    // 4. 生成 Docker 配置
    if err := generateDockerCompose(c); err != nil {
        return err
    }

    // 5. 生成部署脚本
    if err := generateDeployScripts(c); err != nil {
        return err
    }

    return nil
}

func convertToServerArg(c *config.MicroServiceArgument) *config.ServerArgument {
    return &config.ServerArgument{
        CommonParam: c.CommonParam,
        // ... 映射其他字段
    }
}
```

**步骤 4**：定义命令标志

**文件**：`cmd/static/microservice_flags.go`

```go
package static

import "github.com/urfave/cli/v2"

const (
    MicroServiceUsage = "Generate a complete microservice"
)

func microserviceFlags() []cli.Flag {
    return []cli.Flag{
        // 基础参数
        &cli.StringFlag{
            Name:     "service",
            Required: true,
            Usage:    "Service name",
        },
        &cli.StringFlag{
            Name:     "module",
            Required: true,
            Usage:    "Go module name",
        },
        &cli.StringFlag{
            Name:     "idl",
            Required: true,
            Usage:    "IDL file path",
        },

        // 数据库参数
        &cli.BoolFlag{
            Name:  "enable-database",
            Usage: "Enable database model generation",
            Value: false,
        },
        &cli.StringFlag{
            Name:  "db-type",
            Usage: "Database type: mysql, postgres, sqlite",
            Value: "mysql",
        },
        &cli.StringFlag{
            Name:  "dsn",
            Usage: "Database DSN",
        },

        // 缓存参数
        &cli.BoolFlag{
            Name:  "enable-cache",
            Usage: "Enable cache",
            Value: false,
        },
        &cli.StringFlag{
            Name:  "cache-type",
            Usage: "Cache type: redis, memory",
            Value: "redis",
        },

        // 追踪参数
        &cli.BoolFlag{
            Name:  "enable-tracing",
            Usage: "Enable distributed tracing",
            Value: false,
        },
    }
}
```

---

