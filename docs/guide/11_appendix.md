## 十一、附录

### 11.1 完整的模板变量列表

```go
// 服务相关
.ServiceName        string  // 服务名称
.Type              string  // 服务类型（RPC/HTTP）
.GoModule          string  // Go 模块名
.IdlPath           string  // IDL 文件路径
.OutDir            string  // 输出目录

// 方法相关
.Methods           []Method // 所有方法
.Name              string   // 当前方法名
.Args              []Arg    // 方法参数
.Resp              Response // 返回值
.Void              bool     // 是否无返回值
.Oneway            bool     // 是否单向调用
.ClientStreaming   bool     // 是否客户端流
.ServerStreaming   bool     // 是否服务端流

// 配置相关
.Registry          string   // 注册中心类型
.Template          string   // 模板路径
.Verbose           bool     // 详细模式

// 自定义配置（需要添加）
.EnableCache       bool
.CacheType         string
.EnableTracing     bool
// ... 更多自定义字段
```

### 11.2 常用模板函数速查

```go
// 字符串处理
{{ toLower "String" }}        // "string"
{{ toUpper "String" }}        // "STRING"
{{ title "string" }}          // "String"
{{ snakeCase "String" }}      // "string"
{{ camelCase "string_name" }} // "stringName"
{{ kebabCase "StringName" }}  // "string-name"

// 字符串操作
{{ trim "  hello  " }}        // "hello"
{{ split "a,b,c" "," }}       // ["a", "b", "c"]
{{ join ["a","b"] "," }}      // "a,b"

// 列表操作
{{ list "a" "b" "c" }}        // ["a", "b", "c"]
{{ first ["a","b"] }}         // "a"
{{ last ["a","b"] }}          // "b"
{{ slice ["a","b","c"] 1 2 }} // ["b"]

// 数学运算
{{ add 1 2 }}                 // 3
{{ sub 3 1 }}                 // 2
{{ mul 2 3 }}                 // 6
{{ div 6 2 }}                 // 3

// 默认值
{{ default "default" .Value }} // 如果 Value 为空则使用 "default"

// 条件判断
{{ if .Enable }}enabled{{end}}
{{ if eq .Type "rpc" }}RPC{{else}}HTTP{{end}}

// 循环
{{ range .Methods }}
  Method: {{.Name}}
{{ end }}

// 自定义函数
{{ ToCamel "string_name" }}   // "StringName"
{{ SnakeString "StringName" }} // "string_name"
{{ LowerFirst "String" }}      // "string"
```

### 11.3 更新行为说明

```yaml
update_behavior:
  type: skip    # 跳过已存在的文件
  type: cover   # 覆盖已存在的文件
  type: append  # 追加到已存在的文件
```

### 11.4 资源链接

- [CloudWeGo 官方文档](https://www.cloudwego.io/)
- [Kitex 文档](https://www.cloudwego.io/docs/kitex/)
- [Hertz 文档](https://www.cloudwego.io/docs/hertz/)
- [Go Template 文档](https://pkg.go.dev/text/template)
- [Sprig 函数库](https://masterminds.github.io/sprig/)
- [GORM 文档](https://gorm.io/docs/)

---

