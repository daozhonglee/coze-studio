# API /workflow_api/save → SaveWorkflow 执行路径图

## 入口函数执行路径

```
入口函数  api/workflow_api/save  → SaveWorkflow

// SaveWorkflow 保存工作流画布的完整流程：

前端请求 (SaveWorkflowRequest)
    ↓ HTTP请求处理
API Handler (SaveWorkflow)
    ↓ 参数绑定和验证
Application Layer (SaveWorkflow)
    ↓ 用户身份验证 → ctxutil.MustGetUIDFromCtx()
    ↓ 权限验证 → checkUserSpace()
    ↓ 调用领域层保存 → GetWorkflowDomainSVC().Save()
    ↓
Domain Layer (Save)
    ↓ 画布反序列化 → sonic.UnmarshalString()
    ↓ 参数提取 → extractInputsAndOutputsNamedInfoList()
    ↓ 输入参数序列化 → sonic.MarshalString(inputs)
    ↓ 输出参数序列化 → sonic.MarshalString(outputs)
    ↓ 测试运行状态计算 → calculateTestRunSuccess()
    ↓ 生成提交ID → repo.GenID()
    ↓ 保存草稿到数据库 → repo.CreateOrUpdateDraft()
    ↓
Repository Layer
    ↓ 数据库事务开始
    ↓ 更新workflow_drafts表
    ↓ 事务提交
    ↓
响应返回 (SaveWorkflowResponse)
    ↓ 返回空数据对象 → SaveWorkflowData{}
```

## 关键数据流

### 请求数据
```
SaveWorkflowRequest
├── workflow_id: 工作流ID
├── schema: 画布JSON配置
├── space_id: 空间ID
├── name: 工作流名称 [可选]
├── desc: 工作流描述 [可选]
├── icon_uri: 图标URI [可选]
├── submit_commit_id: 提交ID
└── ignore_status_transfer: 忽略状态转移 [可选]
```

### 响应数据
```
SaveWorkflowResponse
└── data: SaveWorkflowData
    ├── name: 工作流名称
    ├── url: 工作流URL
    ├── status: 开发状态
    └── workflow_status: 工作流状态
```

### 数据库更新
```
workflow_drafts表 (更新操作)
├── workflow_id: 工作流ID (WHERE条件)
├── commit_id: 新生成的提交ID (版本控制)
├── canvas: 更新的画布JSON配置
├── input_params: 序列化的输入参数定义 (JSON)
├── output_params: 序列化的输出参数定义 (JSON)
├── test_run_success: 重新计算的测试运行状态
├── modified: true (标记为已修改)
└── updated_at: 更新时间戳
```

## 参数提取逻辑

### 输入输出参数提取
```
extractInputsAndOutputsNamedInfoList()
    ↓
查找Entry节点 → 从输出中提取输入参数
    ↓
查找Exit节点 → 从输入中提取输出参数
    ↓
转换为NamedTypeInfo列表
    ↓
序列化为JSON字符串
```

## 测试状态计算

```
calculateTestRunSuccess()
    ↓
将当前画布转换为WorkflowSchema
    ↓
获取之前保存的草稿版本
    ↓
比较新旧Schema的执行逻辑
    ↓
执行逻辑相同 → 继承之前的测试状态
执行逻辑不同 → 重置测试状态为false
```

## 关键技术特点

- **版本控制**: 每次保存生成新的提交ID
- **智能状态**: 自动判断是否需要重新测试
- **参数同步**: 输入输出参数与画布保持同步
- **异常恢复**: 参数提取失败不影响保存流程
- **数据一致性**: 画布和参数在同一事务中更新

## 与其他操作的区别

- **SaveWorkflow**: 只保存画布，不更新元数据，无事件发布
- **UpdateWorkflowMeta**: 更新元数据，会异步发布事件到搜索索引
- **CreateWorkflow**: 创建新工作流，生成ID，初始化元数据和草稿

