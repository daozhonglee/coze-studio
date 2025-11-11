# API /workflow_api/create → CreateWorkflow 执行路径图

## 入口函数执行路径

```
入口函数  api/workflow_api/create  → CreateWorkflow

// CreateWorkflow 创建工作流的完整流程：

前端请求 (CreateWorkflowRequest)
    ↓ HTTP请求处理
API Handler (WorkFlowCreate)
    ↓ 参数绑定和验证
Application Layer (CreateWorkflow)
    ↓ 用户身份验证 → ctxutil.MustGetUIDFromCtx()
    ↓ 权限验证 → checkUserSpace()
    ↓ 会话模板创建 → CreateDraftConversationTemplate() [ChatFlow模式可选]
    ↓ 元数据构建 → vo.MetaCreate{}
    ↓ 画布初始化 → GetDefaultInitCanvasJsonSchema()
    ↓ 工作流模式判断 → ChatFlow vs Workflow
    ↓
Domain Layer (Create)
    ↓ 元数据创建 → repo.CreateMeta()
    ↓ 生成工作流ID
    ↓ 草稿版本保存 → Save()
    ↓ 画布解析 → extractInputsAndOutputsNamedInfoList()
    ↓ 参数序列化 → sonic.MarshalString()
    ↓ 测试状态计算 → calculateTestRunSuccess()
    ↓ 提交ID生成 → repo.CreateOrUpdateDraft()
    ↓
Repository Layer
    ↓ 数据库事务开始
    ↓ 写入workflow_meta表
    ↓ 写入workflow_drafts表
    ↓ 事务提交
    ↓
Event Publishing
    ↓ 资源事件构建 → search.ResourceDocument{}
    ↓ 事件发布 → search.ResourceEventBus.PublishResources()
    ↓ 搜索索引更新（异步）
    ↓
响应返回 (CreateWorkflowResponse)
    ↓ 返回工作流ID → WorkflowID: strconv.FormatInt(id, 10)
```

## 关键数据流

### 请求数据
```
CreateWorkflowRequest
├── name: 工作流名称
├── desc: 工作流描述
├── icon_uri: 图标URI
├── space_id: 空间ID
├── flow_mode: 工作流模式 (Workflow/ChatFlow)
├── schema_type: 模式类型
├── project_id: 项目ID [可选]
├── create_conversation: 是否创建会话 [ChatFlow专用]
└── bind_biz_id/type: 业务绑定信息 [可选]
```

### 响应数据
```
CreateWorkflowResponse
└── data: CreateWorkflowData
    └── workflow_id: 新创建的工作流ID (字符串格式)
```

### 数据库写入
```
workflow_meta表
├── id: 工作流ID (自增主键)
├── creator_id: 创建者用户ID
├── space_id: 所属空间ID
├── content_type: 内容类型 (User)
├── name: 工作流名称
├── description: 工作流描述
├── icon_uri: 图标URI
├── app_id: 应用ID [可选]
├── mode: 工作流模式
└── timestamps: 创建/更新时间

workflow_drafts表
├── workflow_id: 工作流ID
├── commit_id: 提交ID (版本控制)
├── canvas: 画布JSON配置
├── input_params: 输入参数定义 (JSON)
├── output_params: 输出参数定义 (JSON)
├── test_run_success: 测试运行状态
└── updated_at: 更新时间
```

## 分支逻辑

### 工作流模式分支
```
flow_mode == ChatFlow?
├── YES → 使用聊天画布模板
│   ├── create_conversation == true?
│   │   ├── YES → 创建会话模板
│   │   └── NO → 使用默认会话名
│   └── 初始化聊天画布
└── NO → 使用普通工作流画布模板
```

## 关键技术特点

- **事务一致性**: 元数据和草稿在同一事务中创建
- **版本控制**: 每次保存生成新的提交ID
- **异步事件**: 搜索索引更新不阻塞主流程
- **模板系统**: 支持不同类型的画布模板
- **多租户**: 通过space_id实现租户隔离

