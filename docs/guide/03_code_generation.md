## 三、代码生成流程

### 3.1 整体流程图

```
┌─────────────────────────────────────────────────────────────┐
│                         用户执行命令                          │
│                    cwgo server -type rpc ...                 │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    CLI 命令解析                              │
│                  cmd/static/cmd.go                           │
│  - 解析命令参数                                                │
│  - 路由到对应的处理函数                                        │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    参数验证和转换                              │
│                  config/*.go                                 │
│  - 验证参数合法性                                              │
│  - 转换为内部配置结构                                          │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    选择生成策略                                │
│                  pkg/server/server.go                        │
│                                                              │
│   ┌──────────────┐         ┌──────────────┐                │
│   │  RPC 类型     │         │  HTTP 类型    │                │
│   │  (Kitex)     │         │  (Hertz)     │                │
│   └──────┬───────┘         └──────┬───────┘                │
│          │                        │                          │
│          ▼                        ▼                          │
│   pkg/server/          pkg/server/hz.go                     │
│   kitex.go                                                  │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│              调用底层工具（Kitex/Hz）                          │
│  - Kitex: 调用 kitex 命令行工具                               │
│  - Hertz: 调用 hz 插件模式                                    │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                    读取并渲染模板                              │
│                  tpl/kitex or tpl/hertz                      │
│  - 从 embed.FS 读取模板                                      │
│  - 使用 Sprig 模板函数                                       │
│  - 渲染生成代码                                              │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                    后处理和文件生成                             │
│  - 替换 Thrift 版本                                          │
│  - 升级 Protobuf                                             │
│  - Hessian2 后处理                                           │
│  - 生成最终文件到输出目录                                      │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 详细执行流程

#### Kitex RPC 服务生成流程

```go
// pkg/server/server.go
func Server(c *config.ServerArgument) error {
    // 1. 参数校验
    err = check(c)
    if err != nil {
        return err
    }

    // 2. 根据 Type 选择策略
    switch c.Type {
    case consts.RPC:
        // 3. 转换 Kitex 参数
        var args kargs.Arguments
        log.Verbose = c.Verbose
        err = convertKitexArgs(c, &args)

        // 4. 处理注册表
        kx_registry.HandleRegistry(c.CommonParam, args.TemplateDir)
        defer kx_registry.RemoveExtension()

        // 5. 构建并执行命令
        out := new(bytes.Buffer)
        cmd := args.BuildCmd(out)
        err = cmd.Run()

        // 6. 后处理
        utils.ReplaceThriftVersion()
        utils.UpgradeGolangProtobuf()
        utils.Hessian2PostProcessing(args)

    case consts.HTTP:
        // Hertz 生成流程...
    }

    return nil
}
```

#### Hertz HTTP 服务生成流程

```go
// pkg/server/hz.go（简化）
func convertHzArgument(c *config.ServerArgument, args *hzConfig.Argument) error {
    // 1. 判断项目类型
    if utils.IsHzNew(c.OutDir) {
        // 新项目：生成完整布局
        args.CmdType = meta.CmdNew
        err = app.GenerateLayout(args)
    } else {
        // 已存在项目：更新
        args.CmdType = meta.CmdUpdate
        err = manifest.InitAndValidate(args.OutDir)
    }

    // 2. 触发插件生成路由
    err = app.TriggerPlugin(args)

    return nil
}
```

### 3.3 模板渲染流程

```go
// tpl/init.go
func Init() {
    // 1. 清理临时目录
    os.RemoveAll(KitexDir)
    os.RemoveAll(HertzDir)

    // 2. 创建临时目录
    os.Mkdir(KitexDir, 0o755)
    os.Mkdir(HertzDir, 0o755)

    // 3. 从 embed.FS 提取模板到临时目录
    initDir(kitexTpl, consts.Kitex, KitexDir)
    initDir(hertzTpl, consts.Hertz, HertzDir)
}

// 注册模板函数（Sprig + 自定义）
func RegisterTemplateFunc() {
    // 1. 注册 Sprig 函数库
    for k, f := range sprig.FuncMap() {
        generator.AddTemplateFunc(k, f)
    }

    // 2. 注册自定义函数
    generator.AddTemplateFunc("ToCamel", func(name string) string {
        name = strings.Replace(name, "_", " ", -1)
        name = strings.Title(name)
        return strings.Replace(name, " ", "", -1)
    })
}
```

---

