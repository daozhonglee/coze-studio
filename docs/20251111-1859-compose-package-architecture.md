# Compose 包架构分析

## 概述

`backend/domain/workflow/internal/compose` 是 Coze Studio 工作流执行引擎的**核心组件**，负责将前端画布配置转换为可执行的工作流实例，并管理整个执行过程。

## 核心定位

### 桥梁作用
```
前端画布 Canvas
    ↓ CanvasToWorkflowSchema
WorkflowSchema
    ↓ compose.NewWorkflow
Compose.Workflow (Coze封装)
    ↓ 基于
Eino.Workflow (底层执行引擎)
    ↓ 实际执行
工作流节点调度和执行
```

### 三层架构
1. **适配层**: 将Coze特有的画布模型转换为Eino标准格式
2. **增强层**: 在Eino基础上添加中断恢复、流式输出等高级功能
3. **管理层**: 统一管理工作流生命周期、状态管理和错误处理

## 目录结构

```
compose/
├── workflow.go                    # 工作流实例创建和管理
├── workflow_run.go               # 工作流运行器和执行环境准备
├── workflow_from_node.go         # 从单个节点创建工作流
├── workflow_tool.go              # 工作流工具函数
├── node_builder.go               # 节点实例化工厂
├── node_runner.go                # 节点执行器
├── state.go                      # 执行状态管理
├── stream.go                     # 流式处理和数据源计算
├── field_fill.go                 # 字段填充和类型转换
├── designate_option.go           # 选项配置
└── test/                         # 测试用例
    ├── workflow_test.go
    ├── batch_test.go
    ├── loop_test.go
    └── question_answer_test.go
```

## 核心组件

### 1. Workflow - 工作流实例

**文件**: `workflow.go`

**核心结构**:
```go
type Workflow struct {
    *workflow                      // 继承Eino的Workflow
    hierarchy   map[vo.NodeKey]vo.NodeKey  // 节点层级关系
    connections []*schema.Connection        // 节点连接关系
    requireCheckpoint bool                  // 是否需要检查点
    streamRun   bool                       // 是否流式执行
    input       map[string]*vo.TypeInfo    // 输入类型定义
    output      map[string]*vo.TypeInfo    // 输出类型定义
    schema      *schema.WorkflowSchema     // 工作流Schema
}
```

**主要功能**:
- 创建工作流实例 (`NewWorkflow`)
- 添加节点和连接
- 管理复合节点（循环、条件等）
- 配置输入输出类型
- 支持同步/异步/流式执行

### 2. WorkflowRunner - 工作流运行器

**文件**: `workflow_run.go`

**核心结构**:
```go
type WorkflowRunner struct {
    basic     *entity.WorkflowBasic        // 工作流基本信息
    input     string                       // 输入参数JSON
    resumeReq *entity.ResumeRequest        // 恢复请求
    schema    *schema2.WorkflowSchema      // 工作流Schema
    sw        *schema.StreamWriter          // 流式写入器
    container *execute.StreamContainer     // 流式容器
    config    model.ExecuteConfig          // 执行配置
    
    executeID      int64                   // 执行ID
    eventChan      chan *execute.Event     // 事件通道
    interruptEvent *entity.InterruptEvent  // 中断事件
}
```

**主要功能**:
- 准备执行环境 (`Prepare`)
- 生成执行ID
- 处理中断恢复
- 创建执行记录
- 管理事件通道
- 设置超时和取消机制

### 3. State - 执行状态

**文件**: `state.go`

**核心结构**:
```go
type State struct {
    NodeExeContexts      map[vo.NodeKey]*execute.Context      // 节点执行上下文
    WorkflowExeContext   *execute.Context                     // 工作流执行上下文
    ExecutedNodes        map[vo.NodeKey]bool                  // 已执行节点
    SourceInfos          map[vo.NodeKey]map[string]*schema2.SourceInfo  // 数据源信息
    Inputs               map[vo.NodeKey]map[string]any        // 节点输入
    NestedWorkflowStates map[vo.NodeKey]*nodes.NestedWorkflowState  // 嵌套工作流状态
    ResumeData           map[vo.NodeKey]string                // 恢复数据
    IntermediateResult   map[vo.NodeKey]map[string]any        // 中间结果
}
```

**主要功能**:
- 维护执行状态
- 管理节点上下文
- 支持状态序列化/反序列化
- 提供状态恢复能力

### 4. Node Builder - 节点构建器

**文件**: `node_builder.go`

**主要功能**:
- 将 `NodeSchema` 转换为实际可执行的节点
- 处理复合节点（循环、批量等）
- 注入节点依赖
- 配置节点选项

### 5. Stream - 流式处理

**文件**: `stream.go`

**主要功能**:
- 计算节点的完整输入源
- 处理嵌套工作流的数据映射
- 支持流式数据输出
- 管理数据流管道

### 6. Field Fill - 字段填充

**文件**: `field_fill.go`

**主要功能**:
- 处理节点间的数据类型转换
- 为缺失字段提供默认值
- 确保类型安全
- 支持复杂类型的序列化

## 工作流程

### 创建工作流
```
1. Canvas (前端画布)
    ↓
2. CanvasToWorkflowSchema (adaptor)
    ↓ 转换
3. WorkflowSchema
    ↓
4. compose.NewWorkflow()
    ↓ 创建
5. Workflow 实例
    ↓ 配置
6. 添加节点和连接
    ↓
7. 编译为可执行形式
```

### 执行工作流
```
1. NewWorkflowRunner()
    ↓ 创建运行器
2. Prepare()
    ↓ 准备环境
    ├─ 生成执行ID
    ├─ 处理中断恢复
    ├─ 创建执行记录
    └─ 启动事件处理
3. AsyncRun() / SyncRun()
    ↓ 执行
4. Eino引擎调度
    ↓
5. 节点逐个执行
    ↓
6. 事件通知和状态更新
```

## 关键技术特点

### 1. 基于Eino框架
- 使用云音乐开源的Eino作为底层执行引擎
- 通过组合模式扩展功能
- 保持与Eino框架的兼容性

### 2. 状态持久化
- 支持执行状态的序列化
- 实现工作流中断后的恢复
- 管理嵌套工作流的状态

### 3. 流式支持
- 自动检测是否需要流式输出
- 管理流式数据管道
- 支持实时进度反馈

### 4. 中断恢复
- 记录中断点信息
- 状态修改器机制
- 支持嵌套节点的恢复

### 5. 类型安全
- 完整的类型信息传递
- 节点间类型检查
- 运行时类型转换

### 6. 选项模式
- 函数选项模式配置
- 灵活的运行时配置
- 支持链式调用

## 与其他模块的关系

### 依赖关系
```
compose (本包)
    ↓ 依赖
├── internal/schema         # WorkflowSchema定义
├── internal/nodes          # 节点实现
├── internal/execute        # 执行引擎
├── entity/vo               # 值对象
└── eino/compose            # Eino框架
```

### 被依赖关系
```
service/executable_impl.go  # 领域服务
    ↓ 调用
compose.NewWorkflow()
compose.NewWorkflowRunner()
```

## 使用示例

### 创建并执行工作流
```go
// 1. 转换画布为Schema
workflowSC, err := adaptor.CanvasToWorkflowSchema(ctx, canvas)

// 2. 创建工作流实例
wf, err := compose.NewWorkflow(ctx, workflowSC,
    compose.WithIDAsName(workflowID),
    compose.WithMaxNodeCount(100))

// 3. 转换输入参数
convertedInput, err := nodes.ConvertInputs(ctx, input, wf.Inputs())

// 4. 创建运行器并准备环境
cancelCtx, executeID, opts, eventChan, err := 
    compose.NewWorkflowRunner(basic, workflowSC, config,
        compose.WithInput(inputJSON)).Prepare(ctx)

// 5. 异步执行
wf.AsyncRun(cancelCtx, convertedInput, opts...)

// 6. 处理执行事件
for event := range eventChan {
    // 处理事件...
}
```

## 设计模式

### 1. 组合模式
- `Workflow` 组合 `eino.Workflow`
- 通过组合而非继承扩展功能

### 2. 工厂模式
- `NewWorkflow()` - 工作流工厂
- `New()` - 节点构建工厂

### 3. 选项模式
- `WorkflowOption` 函数选项
- `WorkflowRunnerOption` 函数选项

### 4. 建造者模式
- `WorkflowRunner` 构建执行环境
- 分步准备和配置

### 5. 策略模式
- 不同的执行策略（同步/异步/流式）
- 可插拔的状态修改器

## 性能优化

1. **惰性初始化**: Schema只在需要时初始化
2. **并发安全**: 使用sync.Once确保单次初始化
3. **内存复用**: 状态对象池化
4. **流式处理**: 支持大数据量的流式输出
5. **上下文取消**: 支持执行的优雅取消

## 总结

`compose` 包是 Coze Studio 工作流系统的**执行心脏**：

- 🎯 **连接**: 前端画布设计与后端执行引擎
- 🔧 **转换**: 业务模型为技术实现
- 🚀 **增强**: 基础执行能力，添加高级功能
- 🛡️ **管理**: 复杂的工作流生命周期和状态
- 🔄 **恢复**: 支持中断后的状态恢复
- 📊 **监控**: 完整的执行事件和状态追踪

这个包的设计体现了领域驱动设计原则，将工作流执行的复杂性封装在清晰的接口后面，为上层应用提供了简单而强大的工作流执行能力。

