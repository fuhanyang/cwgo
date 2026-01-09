## 九、开发工作流

### 9.1 推荐的定制化开发流程

```
1. 需求分析
   ↓
2. 确定要修改的文件
   ↓
3. 备份原始模板
   ↓
4. 修改模板文件
   ↓
5. 本地测试
   ↓
6. 构建和验证
   ↓
7. 提交代码
```

### 9.2 本地测试流程

```bash
# 1. 修改模板文件
vim tpl/kitex/server/standard/service.yaml

# 2. 构建工具
go build -o cwgo cwgo.go

# 3. 测试生成代码
./cwgo server -type rpc \
  -module test-project \
  -service test.service \
  -idl idl/test.thrift \
  -out-dir /tmp/test_output

# 4. 检查生成的代码
ls -la /tmp/test_output
cat /tmp/test_output/biz/service/testservice/method.go

# 5. 尝试编译生成的代码
cd /tmp/test_output
go mod tidy
go build
```

### 9.3 调试技巧

#### 1. 查看渲染后的模板

在模板中添加注释：

```yaml
body: |-
  // DEBUG: ServiceName = {{.ServiceName}}
  // DEBUG: Method = {{.Name}}
  package {{SnakeString .ServiceName}}
  // ...
```

#### 2. 使用详细模式

```bash
./cwgo server -vv -type rpc ...
```

#### 3. 检查中间产物

```bash
# 查看临时目录中的模板
ls -la $TMPDIR/kitex/
```

### 9.4 常见问题

#### Q1: 模板变量不生效

**原因**：模板变量名称错误或未传递

**解决**：
1. 检查变量名是否正确（区分大小写）
2. 确认变量在生成逻辑中已设置
3. 使用详细模式查看错误信息

#### Q2: 生成的代码编译失败

**原因**：模板语法错误或导入路径错误

**解决**：
1. 检查模板的 YAML 语法
2. 验证 import 语句的正确性
3. 确保所有必要的包都已导入

#### Q3: 如何添加条件渲染

**方法**：使用 Go template 的 if 语句

```yaml
{{if .EnableCache}}
import "github.com/go-redis/redis/v8"
{{end}}

type Service struct {
    {{if .EnableCache}}
    redis *redis.Client
    {{end}}
}
```

#### Q4: 如何循环生成代码

**方法**：使用 Go template 的 range 语句

```yaml
{{range .Methods}}
func (s *Service) {{.Name}}({{range .Args}}{{.Name}} {{.Type}}, {{end}}) {
    // 方法实现
}
{{end}}
```

---

