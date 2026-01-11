# 提案：基于契约的处理器编排与安全校验机制 (Contract-Based Processor Orchestration)

## 1. 背景与问题

当前的微服务架构（基于 CloudWeGo/Kitex + 策略模式）使用了 **Pipeline 模式** 来编排业务逻辑。
- **SearchState** 等上下文对象在多个 Processor 之间传递。
- Processor 之间的**数据依赖是隐式的**（例如：`Rank` Processor 假设 `Candidates` 字段已被 `Retrieve` Processor 填充）。
- **配置与代码脱节**：`conf.yaml` 定义了执行顺序，但无法感知数据依赖。如果配置顺序错误（如先排序再检索），只有在运行时才会触发 Panic 或逻辑错误。

## 2. 核心提案

我们建议引入 **"代码定义契约，启动时自动校验" (Code-First Contract, Startup Validation)** 的机制。

### 2.1 核心原则

1.  **配置极简**：YAML 配置文件仅保留 Processor 的**执行顺序**，不包含出入参定义。
2.  **代码即契约**：每个 Processor 在 Go 代码中显式声明其**输入依赖 (Reads)** 和**输出承诺 (Writes)**。
3.  **零运行时开销**：依赖校验仅在**服务启动阶段 (Startup Time)** 进行，构建好的 Pipeline 在运行时无额外性能损耗。
4.  **混合状态模型**：推荐使用 "核心数据强类型 (Req/Resp) + 临时数据弱类型 Map (Vars)" 的混合模式。

## 3. 实现规范

### 3.1 解耦依赖：引入 `biz/processor/types.go`

为了避免循环依赖，建议将基础类型定义在 `biz/processor` 包中：

```go
package processor

// DataField 代表上下文中的字段路径，如 "Req.Query", "Vars.UserScore"
type DataField string

// Contract 定义 Processor 的数据契约
type Contract struct {
    // Reads: 该 Processor 必须读取的字段。
    // 如果这些字段在 Pipeline 前序步骤中未产生，启动时应报错。
    // 支持层级匹配：如果上游提供了 "Req"，则 "Req.Query" 也会被视为满足。
    Reads []DataField 
    
    // Writes: 该 Processor 承诺会填充或修改的字段。
    // 该字段在后续的所有步骤中均可见。
    Writes []DataField
}

// Processor 接口扩展 Contract 方法
type Processor[S any] interface {
    Name() string
    Process(s *S) error
    
    // 新增：返回契约元数据
    Contract() *Contract
}
```

### 3.2 处理器实现示例

开发者编写 Processor 时，需实现 `Contract()` 方法：

```go
// biz/processor/search_demo/normalize_query.go

func (NormalizeQuery) Contract() *processor.Contract {
    return &processor.Contract{
        // 声明依赖：必须有原始查询词
        // 校验器支持层级推导：如果初始状态只有 "Req"，此依赖也满足。
        Reads:  []processor.DataField{"Req.Query"},
        // 声明产出：我会生成归一化后的查询词
        Writes: []processor.DataField{"Resp.Normalized", "Vars.terms"},
    }
}
```

### 3.3 启动时校验逻辑 (支持层级推导)

脚手架应生成通用的校验函数，在 `initStrategies()` 或 Pipeline 构建时调用。
校验器应支持**累积上下文 (Cumulative Context)** 和 **层级匹配 (Hierarchical Matching)**。

```go
func ValidatePipeline[S any](processors []Processor[S], initialFields []DataField) error {
	// 追踪当前可用的字段集合 (累积)
	availableFields := make(map[DataField]bool)
	for _, f := range initialFields {
		availableFields[f] = true
	}

	for _, p := range processors {
		c := p.Contract()
		if c == nil { continue }

		// 1. 检查依赖是否满足
		for _, read := range c.Reads {
			satisfied := false
			// 精确匹配
			if availableFields[read] {
				satisfied = true
			} else {
				// 层级匹配：如果需要 "Req.Query"，检查 "Req" 是否存在
				s := string(read)
				for i := 0; i < len(s); i++ {
					if s[i] == '.' {
						parent := DataField(s[:i])
						if availableFields[parent] {
							satisfied = true
							break
						}
					}
				}
			}

			if !satisfied {
				return fmt.Errorf("Pipeline Error: Processor [%s] requires field [%s], but it is not available yet.", p.Name(), read)
			}
		}

		// 2. 注册该步骤产生的字段 (供后续步骤使用)
		for _, write := range c.Writes {
			availableFields[write] = true
		}
	}
	return nil
}
```

## 4. 拓展性讨论

### 4.1 支持并行与DAG
该校验机制天然支持并行或DAG结构，只需调整校验逻辑：
- **并行块**：输入校验需满足所有并行任务的依赖；输出集合为所有并行任务产出的并集。
- **DAG**：只要确保节点的入度依赖在执行前满足即可。

### 4.2 性能影响
- **运行时**：零开销。校验在启动时完成，运行时仅执行 `Process()` 函数。
- **启动时**：微秒级开销，完全可忽略。

## 5. 高级模式：基于 DAG 的自动并行调度 (DAG-Based Auto-Parallelism)

为了最大化利用多核性能，同时避免手动编排的复杂性，我们将“并行执行”与“依赖分析”合二为一，提出**智能并行处理器 (Smart DAG Processor)**。

这个 Processor 本身也是一个标准的 `Processor`，可以被嵌套使用。它内部通过分析子任务的 `Contract`，自动构建有向无环图 (DAG) 并进行分层调度。

### 5.1 统一设计思路

不再需要开发者手动区分“串行”还是“并行”，而是提供一个容器（我们称为 `DAGProcessor`），只需放入一堆 Processor，它会自动：

1.  **解析依赖**：读取所有子 Processor 的 `Contract`。
2.  **构建图谱**：根据 Reads/Writes 关系建立依赖连线。
3.  **自动分层**：将没有依赖关系的 Processor 放在同一层（Layer）。
4.  **并行执行**：运行时，层内自动并发，层间自动等待。

### 5.2 核心实现 (DAGProcessor)

```go
// DAGProcessor 是一个智能容器，它表现为一个单一的 Processor，
// 但内部包含了一组根据依赖关系自动分层调度的子 Processors。
type DAGProcessor[S any] struct {
    NameStr string
    // Layers 是预计算好的执行计划。
    // 例如: [[A, B], [C], [D, E]] 表示 A/B 并行 -> 之后 C -> 之后 D/E 并行
    Layers  [][]Processor[S]
    
    // 缓存合并后的契约
    mergedContract *Contract
}

func (dp *DAGProcessor[S]) Name() string { return dp.NameStr }

// NewDAGProcessor 在初始化阶段完成核心的拓扑排序和分层逻辑
func NewDAGProcessor[S any](name string, tasks []Processor[S]) (*DAGProcessor[S], error) {
    // 1. 构建依赖图 (Graph Building)
    //    - 节点：每个 task
    //    - 边：如果 Task B Reads "x", Task A Writes "x", 则 A -> B
    //    - 注意：需处理层级匹配 (e.g. Reads "Req.Query" 可以被 Writes "Req" 满足)
    
    // 2. 拓扑分层 (Topological Layering)
    //    - Kahn 算法变体：
    //      Layer 0: 所有入度为 0 的节点
    //      Layer 1: 移除 Layer 0 后，新出现的入度为 0 的节点
    //      ...
    
    // 3. 计算合并契约 (Contract Merging)
    //    Reads = (Union(All Reads) - Union(All Internal Writes))
    //    Writes = Union(All Writes)
    
    return &DAGProcessor[S]{
        NameStr: name,
        Layers:  calculatedLayers,   // 伪代码：实际需填充计算结果
        mergedContract: calculatedContract, // 伪代码：实际需填充计算结果
    }, nil
}

func (dp *DAGProcessor[S]) Process(s *S) error {
    // 运行时极其高效：直接按层调度
    for _, layer := range dp.Layers {
        if len(layer) == 1 {
            // 串行优化：单任务直接执行，避免协程开销
            if err := layer[0].Process(s); err != nil {
                return err
            }
        } else {
            // 多任务层：并行执行
            var wg sync.WaitGroup
            var errs []error
            var mu sync.Mutex

            for _, task := range layer {
                wg.Add(1)
                go func(t Processor[S]) {
                    defer wg.Done()
                    if err := t.Process(s); err != nil {
                        mu.Lock()
                        errs = append(errs, err)
                        mu.Unlock()
                    }
                }(task)
            }
            wg.Wait()
            
            if len(errs) > 0 {
                return fmt.Errorf("DAG layer execution failed: %v", errs)
            }
        }
    }
    return nil
}

func (dp *DAGProcessor[S]) Contract() *Contract {
    return dp.mergedContract
}
```

### 5.3 优势与场景

*   **完全透明**：业务开发者只需关注单个 Processor 的输入输出，不需要关心谁和谁并行。
*   **包含普通并行 (Scatter-Gather)**：如果放入一组互不依赖的 Processor，DAG 调度器会将它们全部分配到第一层 (Layer 0) 并行执行。这意味着你不需要单独实现一个“普通并行”的 Processor，DAGProcessor 会自动退化为那种行为。
*   **自动优化**：如果删除了某个依赖字段，调度器会自动将原本串行的任务升级为并行，无需修改配置。
*   **混合编排**：由于 `DAGProcessor` 也是一个 Processor，它可以被嵌套在另一个 Pipeline 中，甚至嵌套在另一个 `DAGProcessor` 中，实现任意复杂的流控。

### 5.4 状态安全策略 (Context Safety)

在原有架构基础上引入并行，必须处理 `Context` 的线程安全问题。由于我们不想大规模重构 `Context` 结构，建议采用以下轻量级方案：

1.  **约定优于配置**：DAG 调度器通过 Contract 保证并行的 Processor 写入的是**不同的字段**（Writes 集合无交集），从逻辑上避免冲突。
2.  **底层安全**：
    *   对于 `Request/Response` 等强类型结构体字段，不同 Processor 修改不同字段是内存安全的。
    *   对于 `Vars` (Map) 类型的弱类型字段，**必须**将其替换为 `sync.Map` 或在 Set 方法中加锁，因为 Go 的普通 Map 不支持并发写（即使是不同的 Key）。

    ```go
    // 推荐改造 Context 基础结构
    type BaseContext struct {
        // ...
        Vars sync.Map // 仅需改动此处即可支持并行
    }
    ```

## 6. 对脚手架的修改建议

1.  **代码结构**：创建 `biz/processor` 包存放基础类型，避免循环依赖。
2.  **模板更新**：修改 `Processor` 接口模板，增加 `Contract() *Contract` 方法存根。
3.  **生成逻辑**：在生成 `strategy_init.go` 或类似初始化代码时，插入 `ValidatePipeline` 调用。
