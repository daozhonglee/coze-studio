# API /workflow_api/test_run 调用链路图

## 入口函数执行路径

```
入口函数  api/workflow_api/test_run  → WorkFlowTestRun

// WorkFlowTestRun 工作流测试运行API的完整调用链：

前端请求 (WorkFlowTestRunRequest)
    ↓ HTTP请求处理
API Handler (WorkFlowTestRun)
    ↓ 参数绑定和验证
Application Layer (TestRun)
    ↓ 用户身份验证 → ctxutil.MustGetUIDFromCtx()
    ↓ 权限验证 → checkUserSpace()
    ↓ 构建ExecuteConfig → ExecuteModeDebug, FromDraft, SyncPatternAsync
    ↓ 参数验证 → 检查project_id和bot_id不冲突
    ↓ 调用领域层执行 → GetWorkflowDomainSVC().AsyncExecute()
    ↓
Domain Layer (AsyncExecute)
    ↓ 获取工作流实体 → i.Get() → repo.GetWorkflow()
    ↓ 应用版本检查 → checkApplicationWorkflowReleaseVersion()
    ↓ 画布反序列化 → sonic.UnmarshalString(wfEntity.Canvas)
    ↓ Schema转换 → adaptor.CanvasToWorkflowSchema()
    ↓ 文件字段映射 → GetAllNodesInputFileFields()
    ↓ 创建工作流 → compose.NewWorkflow()
    ↓ 配置处理 → AppID设置，CommitID设置
    ↓ 输入转换 → nodes.ConvertInputs()
    ↓ 运行器准备 → compose.NewWorkflowRunner().Prepare()
    ↓ 设置测试执行ID → repo.SetTestRunLatestExeID()
    ↓ 异步执行 → wf.AsyncRun()
    ↓
Compose Layer (Prepare)
    ↓ 生成执行ID → repo.GenID()
    ↓ 处理中断事件 → repo.GetFirstInterruptEvent()
    ↓ 启动流容器 → execute.NewStreamContainer()
    ↓ 获取执行选项 → r.designateOptions()
    ↓ 状态修改器 → GenStateModifierByEventType()
    ↓ 创建执行记录 → repo.CreateWorkflowExecution()
    ↓ 上下文设置 → context.WithCancel() + WithTimeout()
    ↓ 启动事件处理 → go HandleExecuteEvent()
    ↓
Eino Framework (AsyncRun)
    ↓ 工作流节点调度
    ↓ 数据流处理
    ↓ 节点执行
    ↓ 结果收集
    ↓ 事件通知 → eventChan
    ↓
Event Processing (HandleExecuteEvent)
    ↓ 监听事件通道 → eventChan
    ↓ 状态更新 → repo.UpdateWorkflowExecution()
    ↓ 事件发布 → 可能的外部通知
    ↓ 执行完成处理
    ↓
响应返回 (WorkFlowTestRunResponse)
    ↓ 返回执行ID → WorkFlowTestRunData{ExecuteID}
```

## 关键数据流

### 请求参数
```
WorkFlowTestRunRequest
├── workflow_id: "123"      // 工作流ID
├── input: {...}            // 执行输入参数
├── space_id: "456"         // 空间ID
├── commit_id: "abc123"     // 提交ID [可选]
├── project_id: "789"       // 项目ID [可选]
└── bot_id: "101"           // 机器人ID [可选]
```

### 响应数据
```
WorkFlowTestRunResponse
└── data: WorkFlowTestRunData
    ├── workflow_id: "123"
    └── execute_id: "456"    // 用于后续查询执行结果
```

## 并发执行模型

```
主线程 (API → Application → Domain准备) [同步，毫秒级]
    ↓
异步执行线程 (Eino Framework执行) [异步，秒级~分钟级]
    ↓
事件处理线程 (状态更新和通知) [异步，事件驱动]
```

## 关键技术特点

- **异步架构**: API立即返回，实际执行异步进行
- **容错设计**: 多层异常处理和资源清理
- **状态管理**: 完整的执行状态追踪和中断恢复
- **流式支持**: 支持实时输出和进度反馈
- **版本控制**: 通过CommitID确保执行版本一致性
