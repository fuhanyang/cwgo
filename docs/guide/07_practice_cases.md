## 七、实战案例

### 7.1 案例 1：集成公司内部组件库

**需求**：将生成的代码集成到公司内部的基础设施中，包括：
- 统一的日志库
- 统一的配置中心
- 统一的监控和追踪

#### 解决方案

1. **修改 main.go 模板**

```yaml
# tpl/kitex/server/standard/main_tpl.yaml
body: |-
  package main

  import (
    "context"
    "flag"

    "your-company/internal/config"
    "your-company/pkg/logger"
    "your-company/pkg/metrics"
    "your-company/pkg/tracing"
  )

  var (
    configFile = flag.String("c", "configs/config.yaml", "config file path")
  )

  func main() {
    flag.Parse()

    // 初始化配置中心
    cfg, err := config.LoadFromCenter(*configFile)
    if err != nil {
      panic(err)
    }

    // 初始化日志
    logger.Init(cfg.Log)

    // 初始化追踪
    if cfg.Tracing.Enabled {
      tracing.Init(cfg.Tracing)
    }

    // 初始化指标
    if cfg.Metrics.Enabled {
      metrics.Init(cfg.Metrics)
    }

    // 创建服务
    svr := {{.ServiceName}}.NewServer(new({{.ServiceName}}Impl))

    // 启动服务
    svr.Run()
  }
```

2. **修改 service 模板**

```yaml
# tpl/kitex/server/standard/service.yaml
body: |-
  import (
    "context"
    "your-company/pkg/logger"
    "your-company/pkg/errors"
  )

  type {{.Name}}Service struct {
    ctx context.Context
    log *logger.Logger
  }

  func New{{.Name}}Service(ctx context.Context) *{{.Name}}Service {
    return &{{.Name}}Service{
      ctx: ctx,
      log: logger.FromContext(ctx),
    }
  }

  func (s *{{.Name}}Service) Run({{range .Args}}{{LowerFirst .Name}} {{.Type}}, {{end}}) (resp {{.Resp.Type}}, err error) {
    s.log.Infof("[{{.Name}}] Service called")

    defer func() {
      if err != nil {
        s.log.Errorf("[{{.Name}}] Service failed: %v", err)
        err = errors.Wrap(err, "{{.Name}} failed")
      }
    }()

    // 业务逻辑...

    return
  }
}
```

### 7.2 案例 2：生成带缓存的服务代码

**需求**：自动为所有查询方法添加缓存支持

#### 解决方案

1. **添加配置参数**

```go
// config/server.go
type ServerArgument struct {
    // ...
    EnableCache bool
    CacheType   string // redis, memory
    CacheTTL    int    // 缓存过期时间（秒）
}
```

2. **修改 service 模板**

```yaml
# tpl/kitex/server/standard/service.yaml
body: |-
  import (
    "context"
    "encoding/json"
    "time"
    {{- if eq .CacheType "redis" }}
    "github.com/go-redis/redis/v8"
    {{- end}}
  )

  type {{.Name}}Service struct {
    ctx context.Context
    {{- if eq .CacheType "redis" }}
    redis *redis.Client
    {{- end}}
  }

  func New{{.Name}}Service(ctx context.Context{{if eq .CacheType "redis"}}, rdb *redis.Client{{end}}) *{{.Name}}Service {
    return &{{.Name}}Service{
      ctx: ctx{{if eq .CacheType "redis"}},
      redis: rdb{{end}}
    }
  }

  {{if not .Void}}func (s *{{.Name}}Service) Run({{range .Args}}{{LowerFirst .Name}} {{.Type}}, {{end}}) (resp {{.Resp.Type}}, err error) {
    {{- if eq .CacheType "redis"}}
    // 生成缓存键
    cacheKey := s.buildCacheKey({{range $i, $arg := .Args}}{{if $i}}, {{end}}{{$arg.Name}}{{end}})

    // 尝试从缓存获取
    cached, err := s.redis.Get(s.ctx, cacheKey).Result()
    if err == nil {
      json.Unmarshal([]byte(cached), &resp)
      return resp, nil
    }

    // 执行业务逻辑
    resp, err = s.runBusinessLogic({{range .Args}}{{LowerFirst .Name}}, {{end}})
    if err != nil {
      return resp, err
    }

    // 缓存结果
    data, _ := json.Marshal(resp)
    s.redis.Set(s.ctx, cacheKey, data, time.Duration({{.CacheTTL}})*time.Second)

    return resp, nil
  }

  func (s *{{.Name}}Service) runBusinessLogic({{range .Args}}{{LowerFirst .Name}} {{.Type}}, {{end}}) (resp {{.Resp.Type}}, err error) {
    // 原业务逻辑
    return
  }

  func (s *{{.Name}}Service) buildCacheKey({{range .Args}}{{LowerFirst .Name}} {{.Type}}, {{end}}) string {
    key := "{{.ServiceName}}:{{.Name}}:"
    {{- range .Args}}
    key += fmt.Sprintf("%v:", {{LowerFirst .Name}})
    {{- end}}
    return key
  }
  {{- else}}
  // 无缓存版本
  func (s *{{.Name}}Service) Run({{range .Args}}{{LowerFirst .Name}} {{.Type}}, {{end}}) (resp {{.Resp.Type}}, err error) {
    // 业务逻辑
    return
  }
  {{- end}}
  {{- end}}
}
```

### 7.3 案例 3：生成 DDD 架构的项目

**需求**：按照领域驱动设计（DDD）组织代码结构

#### 解决方案

创建新的模板布局：

```
tpl/kitex/server/ddd_layout/
├── domain/
│   ├── domain.yaml          # 领域模型
│   └── repository.yaml      # 仓储接口
├── application/
│   ├── service.yaml         # 应用服务
│   └── dto.yaml             # DTO
├── infrastructure/
│   ├── repository_impl.yaml # 仓储实现
│   └── dal_init.yaml        # 数据访问层
└── interfaces/
    ├── handler.yaml         # 处理器
    └── assembler.yaml       # 组装器
```

**示例 - domain/repository.yaml**：

```yaml
path: domain/{{SnakeString .ServiceName}}/repository.go
body: |-
  package {{SnakeString .ServiceName}}

  import "context"

  // Repository 仓储接口
  type Repository interface {
    {{range .Methods}}
    // {{.Name}} {{.Name}} 方法
    {{.Name}}(ctx context.Context, {{range .Args}}{{.Name}} {{.Type}}, {{end}}) ({{if .Void}}error{{else}}{{.Resp.Type}}, error{{end}})
    {{end}}
  }
```

**示例 - application/service.yaml**：

```yaml
path: application/{{SnakeString .ServiceName}}/service.go
body: |-
  package {{SnakeString .ServiceName}}

  import (
    "context"
    "your-project/domain/{{SnakeString .ServiceName}}"
  )

  // Service 应用服务
  type Service struct {
    repo repository.Repository
  }

  func NewService(repo repository.Repository) *Service {
    return &Service{repo: repo}
  }

  {{range .Methods}}
  func (s *Service) {{.Name}}(ctx context.Context, {{range .Args}}{{.Name}} {{.Type}}, {{end}}) ({{if .Void}}error{{else}}{{.Resp.Type}}, error{{end}}) {
    return s.repo.{{.Name}}(ctx, {{range .Args}}{{.Name}}, {{end}})
  }
  {{end}}
```

---

