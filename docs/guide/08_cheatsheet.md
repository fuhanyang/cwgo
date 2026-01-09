## 八、关键文件速查表

### 8.1 根据需求查找文件

| 定制需求 | 需要修改的文件 | 说明 |
|---------|---------------|------|
| **修改 Kitex 服务代码结构** | `tpl/kitex/server/standard/service.yaml` | 业务服务层模板 |
| **修改 Hertz 服务代码结构** | `tpl/hertz/server/standard/layout.yaml` | 项目布局模板 |
| **修改 Handler 实现** | `tpl/kitex/server/standard/handler_tpl.yaml` | Handler 模板 |
| **修改 main.go 入口** | `tpl/kitex/server/standard/main_tpl.yaml`<br>`tpl/hertz/server/standard/layout.yaml` | 主程序入口 |
| **修改配置结构** | `tpl/kitex/server/standard/conf_tpl.yaml` | 配置文件模板 |
| **添加 CLI 参数** | `cmd/static/*_flags.go`<br>`config/*.go` | 命令行参数和配置 |
| **修改生成逻辑** | `pkg/server/*.go`<br>`pkg/client/*.go`<br>`pkg/model/*.go` | 生成逻辑实现 |
| **创建新模板布局** | `tpl/kitex/server/your_layout/`<br>`tpl/hertz/server/your_layout/` | 自定义模板目录 |
| **修改数据库模型生成** | `pkg/model/model.go` | 模型生成逻辑 |
| **修改配置文件生成** | `pkg/config_generator/*.go` | 配置转换逻辑 |
| **添加新的生成命令** | `cmd/static/cmd.go`<br>`config/your_cmd.go`<br>`pkg/your_cmd/*.go` | 新命令实现 |
| **修改模板函数** | `tpl/init.go` | 注册模板函数 |
| **修改中间件** | `tpl/hertz/server/standard/layout.yaml` | 在 registerMiddleware 中添加 |

### 8.2 模板文件对照表

#### Kitex 服务器模板

| 模板文件 | 生成文件 | 用途 |
|---------|---------|------|
| `main_tpl.yaml` | `main.go` | 主程序入口 |
| `handler_tpl.yaml` | `biz/handler/*.go` | Handler 实现 |
| `service.yaml` | `biz/service/*/*.go` | 业务逻辑层 |
| `conf_tpl.yaml` | `conf/conf.go` | 配置结构体 |
| `conf_dev_tpl.yaml` | `conf/config_dev.yaml` | 开发环境配置 |
| `conf_test_tpl.yaml` | `conf/config_test.yaml` | 测试环境配置 |
| `conf_online_tpl.yaml` | `conf/config_online.yaml` | 生产环境配置 |
| `dal_init.yaml` | `dal/init.go` | 数据访问层初始化 |
| `kitex_yaml.yaml` | `kitex.yaml` | Kitex 工具配置 |
| `mysql.yaml` | `docker-compose.yml` | MySQL 容器配置 |
| `redis.yaml` | `docker-compose.yml` | Redis 容器配置 |
| `bootstrap_sh_tpl.yaml` | `bootstrap.sh` | 启动脚本 |
| `build_sh_tpl.yaml` | `build.sh` | 构建脚本 |
| `docker_compose.yaml` | `docker-compose.yml` | Docker 编排文件 |
| `readme_tpl.yaml` | `README.md` | 项目文档 |
| `ignore_tpl.yaml` | `.gitignore` | Git 忽略规则 |

#### Kitex 客户端模板

| 模板文件 | 生成文件 | 用途 |
|---------|---------|------|
| `client_tpl.yaml` | `client/rpc_client.go` | 客户端接口定义 |
| `default_tpl.yaml` | `client/default_rpc_client.go` | 默认客户端实现 |
| `init_tpl.yaml` | `client/init.go` | 客户端初始化 |

#### Hertz 服务器模板

| 模板文件 | 生成文件 | 用途 |
|---------|---------|------|
| `layout.yaml` | 多个文件 | 整个项目布局 |
| `package.yaml` | - | 包信息配置 |

### 8.3 代码生成模块对照表

| 模块 | 入口文件 | 核心文件 | 功能 |
|------|---------|---------|------|
| **Server** | `pkg/server/server.go` | `pkg/server/kitex.go`<br>`pkg/server/hz.go` | 生成服务端代码 |
| **Client** | `pkg/client/client.go` | `pkg/client/kitex.go`<br>`pkg/client/hz.go` | 生成客户端代码 |
| **Model** | `pkg/model/model.go` | `pkg/model/check.go` | 生成数据库模型 |
| **Doc** | `pkg/curd/doc/doc.go` | `pkg/curd/doc/mongo/` | 生成 MongoDB 代码 |
| **Config** | `pkg/config_generator/sdk.go` | `pkg/config_generator/yaml2go.go` | 配置转代码 |
| **API List** | `pkg/api_list/api_list.go` | `pkg/api_list/parser/` | 路由分析 |

---

