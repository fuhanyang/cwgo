## 四、核心模块详解

### 4.1 服务器生成模块 (pkg/server)

#### 功能职责

- 生成 Kitex RPC 服务代码
- 生成 Hertz HTTP 服务代码
- 支持自定义模板和布局
- 处理注册中心和配置

#### 关键文件

| 文件 | 职责 |
|------|------|
| `server.go` | 服务器生成的主入口，根据类型路由到不同的生成逻辑 |
| `kitex.go` | Kitex RPC 服务生成，参数转换和命令构建 |
| `hz.go` | Hertz HTTP 服务生成，调用 hz 插件 |
| `check.go` | 参数校验，验证用户输入的合法性 |

#### Kitex 参数转换示例

```go
// pkg/server/kitex.go
func convertKitexArgs(sa *config.ServerArgument, kitexArgument *kargs.Arguments) error {
    // 基础参数
    kitexArgument.ModuleName = sa.GoMod
    kitexArgument.ServiceName = sa.ServerName
    kitexArgument.IDL = sa.IdlPath

    // Thrift 编译选项
    kitexArgument.ThriftOptions = append(kitexArgument.ThriftOptions,
        "naming_style=golint",
        "ignore_initialisms",
        "gen_setter",
        "gen_deep_equal",
        "compatible_names",
        "frugal_tag",
    )

    // 模板相关
    if sa.Template != "" {
        kitexArgument.TemplateDir = tpl.KitexDir
    }

    return nil
}
```

### 4.2 客户端生成模块 (pkg/client)

#### 功能职责

- 生成统一的客户端调用接口
- 封装 Kitex 和 Hertz 客户端
- 支持连接池和配置管理
- 提供开箱即用的客户端

#### 生成的客户端结构

```
client/
├── rpc_client.go          # RPC 客户端接口
├── default_rpc_client.go  # 默认实现
└── init.go                # 初始化函数
```

#### 客户端接口示例

```go
// 生成的客户端接口
type RPCClient interface {
    // 用户服务
    UserService() userservice.Client
    // 订单服务
    OrderService() orderservice.Client
}

// 默认实现
type DefaultRPCClient struct {
    userService userservice.Client
    orderService orderservice.Client
}

func (c *DefaultRPCClient) UserService() userservice.Client {
    return c.userService
}
```

### 4.3 模型生成模块 (pkg/model)

#### 功能职责

- 从数据库表结构生成 GORM 模型
- 支持多种数据库（MySQL, PostgreSQL, SQLite, SQL Server）
- 生成 CRUD 方法
- 支持单元测试生成

#### 生成流程

```go
// pkg/model/model.go
func Model(c *config.ModelArgument) error {
    // 1. 连接数据库
    dialector := config.OpenTypeFuncMap[consts.DataBaseType(c.Type)]
    db, err = gorm.Open(dialector(c.DSN))

    // 2. 配置生成器
    genConfig := gen.Config{
        OutPath:           c.OutPath,
        OutFile:           c.OutFile,
        ModelPkgPath:      c.ModelPkgName,
        WithUnitTest:      c.WithUnitTest,
        FieldNullable:     c.FieldNullable,
        FieldSignable:     c.FieldSignable,
        FieldWithIndexTag: c.FieldWithIndexTag,
        FieldWithTypeTag:  c.FieldWithTypeTag,
    }

    // 3. 创建生成器
    g := gen.NewGenerator(genConfig)
    g.UseDB(db)

    // 4. 生成模型
    models, err := genModels(g, db, c.Tables)
    if !c.OnlyModel {
        g.ApplyBasic(models...)
    }

    // 5. 执行生成
    g.Execute()

    return nil
}
```

#### 生成的模型示例

```go
// 生成的模型
type User struct {
    ID        uint   `gorm:"column:id;primaryKey;autoIncrement" gormType:"uint"`
    Name      string `gorm:"column:name;type:varchar(255);not null" gormType:"string"`
    Email     string `gorm:"column:email;type:varchar(255);uniqueIndex" gormType:"string"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (User) TableName() string {
    return "users"
}

// CRUD 方法
type userQuery struct {
    gen.DO
}

func newUserQuery(db *gorm.DB) *userQuery {
    return &userQuery{gen.DO{db: db}}
}

// 查询方法
func (u *userQuery) WhereByID(id uint) *userQuery {
    return u.Where(u.FieldByID().Eq(id))
}
```

### 4.4 配置生成模块 (pkg/config_generator)

#### 功能职责

- 将 YAML/JSON 配置文件转换为 Go 结构体
- 自动识别数据类型
- 生成嵌套结构
- 添加 JSON/YAML 标签

#### YAML 转 Go 示例

输入配置：
```yaml
server:
  host: "localhost"
  port: 8080
database:
  type: "mysql"
  connection:
    host: "127.0.0.1"
    port: 3306
```

生成代码：
```go
type Config struct {
    Server     ServerConfig     `yaml:"server"`
    Database   DatabaseConfig   `yaml:"database"`
}

type ServerConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

type DatabaseConfig struct {
    Type       string             `yaml:"type"`
    Connection ConnectionConfig   `yaml:"connection"`
}

type ConnectionConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}
```

---

