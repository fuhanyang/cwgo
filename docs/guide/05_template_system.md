## 五、模板系统

### 5.1 模板系统架构

#### 技术选型

- **格式**：YAML
- **引擎**：Go template (text/template)
- **存储**：embed.FS（内嵌到二进制）
- **函数库**：Sprig v3 + 自定义函数

#### 模板文件结构

```yaml
path: biz/service/{{SnakeString .ServiceName}}/{{ SnakeString (index .Methods 0).Name }}.go
loop_method: true
update_behavior:
  type: skip
delims:
  - "{{"
  - "}}"
body: |-
  // 模板内容
  package {{SnakeString .ServiceName}}

  import (
      "context"
  )

  type {{.Name}}Service struct {
      ctx context.Context
  }
```

#### 模板字段说明

| 字段 | 说明 |
|------|------|
| `path` | 输出文件路径，支持模板变量 |
| `loop_method` | 是否为每个方法生成一个文件 |
| `update_behavior` | 更新行为（skip/cover/append） |
| `delims` | 模板定界符（可选） |
| `body` | 模板内容 |
| `import_tpl` | 导入其他模板（可选） |

### 5.2 Kitex 模板详解

#### 服务器端模板

| 文件 | 生成路径 | 说明 |
|------|---------|------|
| `main_tpl.yaml` | `main.go` | 主程序入口 |
| `handler_tpl.yaml` | `biz/handler/*.go` | Handler 实现 |
| `service.yaml` | `biz/service/*/*.go` | 业务服务层 |
| `conf_tpl.yaml` | `conf/conf.go` | 配置结构 |
| `conf_*_tpl.yaml` | `conf/config_*.yaml` | 多环境配置 |
| `dal_init.yaml` | `dal/init.go` | 数据访问层初始化 |
| `kitex_yaml.yaml` | `kitex.yaml` | Kitex 工具配置 |
| `mysql.yaml` | `docker-compose.yml` | MySQL 配置 |
| `redis.yaml` | `docker-compose.yml` | Redis 配置 |
| `bootstrap_sh_tpl.yaml` | `bootstrap.sh` | 启动脚本 |
| `build_sh_tpl.yaml` | `build.sh` | 构建脚本 |
| `docker_compose.yaml` | `docker-compose.yml` | Docker 编排 |
| `readme_tpl.yaml` | `README.md` | 项目文档 |

#### 客户端模板

| 文件 | 生成路径 | 说明 |
|------|---------|------|
| `client_tpl.yaml` | `client/rpc_client.go` | 客户端接口 |
| `default_tpl.yaml` | `client/default_rpc_client.go` | 默认实现 |
| `init_tpl.yaml` | `client/init.go` | 初始化函数 |

### 5.3 Hertz 模板详解

#### 标准版本 (standard)

| 文件 | 说明 |
|------|------|
| `layout.yaml` | 项目布局定义，包含 main.go、router、handler 等 |
| `package.yaml` | 包信息配置 |

#### V2 版本 (standard_v2)

优化后的布局，提供了更好的组织结构。

### 5.4 模板变量

#### 常用模板变量

```go
// 服务相关
.ServiceName       // 服务名称
.Type              // 服务类型（RPC/HTTP）
.GoModule          // Go 模块名

// 方法相关
.Methods           // 方法列表
.Name              // 方法名
.Args              // 方法参数
.Resp              // 返回值
.Void              // 是否无返回值
.Oneway            // 是否单向调用

// IDL 相关
.IdlPath           // IDL 文件路径
.Imports           // 导入包

// 自定义配置
.OutDir            // 输出目录
.Registry          // 注册中心类型
```

#### 常用模板函数

```go
// Sprig 函数库
{{ toLower "String" }}        // 转小写
{{ toUpper "String" }}        // 转大写
{{ title "String" }}          // 首字母大写
{{ snakeCase "String" }}      // 蛇形命名
{{ camelCase "String" }}      // 驼峰命名

// 自定义函数
{{ ToCamel "string" }}        // 转驼峰
{{ SnakeString "String" }}    // 转蛇形
{{ LowerFirst "String" }}     // 首字母小写
```

---

