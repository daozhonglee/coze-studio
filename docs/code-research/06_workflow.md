# 核心工作流

## 端到端业务流程

### 用户注册与登录

**它是什么**：用户通过邮箱注册或登录 Web 平台，后端写入 session cookie，后续 Web API 通过 SessionAuth 校验登录态。

**触发方式**：前端 `/sign` 页面调用 passport API。

```mermaid
sequenceDiagram
    participant 用户
    participant Handler
    participant UserApp
    participant UserDomain
    participant DB
    participant Cookie

    用户->>Handler: 邮箱注册或登录
    Handler->>Handler: BindAndValidate
    Handler->>UserApp: Register/Login
    UserApp->>UserApp: 校验邮箱和注册配置
    UserApp->>UserDomain: Create 或 Login
    UserDomain->>DB: 写入或读取 user/session
    UserDomain-->>UserApp: 用户信息和 sessionKey
    UserApp-->>Handler: response + sessionKey
    Handler->>Cookie: 写入 SessionKey
    Handler-->>用户: 返回用户信息
```

**关键步骤说明**：

| 步骤 | 负责模块 | 做什么 | 为什么在这一步 |
| --- | --- | --- | --- |
| 1 | passport handler | 绑定请求并调用用户应用服务 | handler 只处理 HTTP 边界 |
| 2 | user application | 校验邮箱、注册开关和允许邮箱 | 注册策略属于应用用例 |
| 3 | user domain | 创建用户或校验登录 | 账号规则属于用户领域 |
| 4 | handler | 写 session cookie | HTTP cookie 是传输层行为 |
| 5 | SessionAuth | 后续请求校验 cookie | 保持 Web API 登录态 |

### 创建工作流

**它是什么**：用户在空间中创建一个普通 Workflow 或 ChatFlow，系统写入元数据和初始草稿画布。

**触发方式**：前端调用 `/api/workflow_api/create`。

```mermaid
sequenceDiagram
    participant 前端
    participant Handler
    participant WorkflowApp
    participant WorkflowDomain
    participant Repo
    participant EventBus

    前端->>Handler: CreateWorkflowRequest
    Handler->>WorkflowApp: CreateWorkflow
    WorkflowApp->>WorkflowApp: 获取用户并校验空间
    WorkflowApp->>WorkflowApp: 构建初始画布
    WorkflowApp->>WorkflowDomain: Create(MetaCreate)
    WorkflowDomain->>Repo: CreateMeta
    WorkflowDomain->>WorkflowDomain: Save(init canvas)
    WorkflowDomain->>Repo: CreateOrUpdateDraft
    WorkflowApp->>EventBus: 发布资源事件
    WorkflowApp-->>Handler: workflow_id
    Handler-->>前端: CreateWorkflowResponse
```

**关键步骤说明**：

| 步骤 | 负责模块 | 做什么 | 为什么在这一步 |
| --- | --- | --- | --- |
| 1 | application/workflow | 校验用户和空间 | 多租户资源必须归属空间 |
| 2 | application/workflow | 根据模式生成默认画布 | 普通 Workflow 和 ChatFlow 初始模板不同 |
| 3 | domain/workflow | 创建 meta 后保存 draft | 生命周期起点同时需要元数据和画布 |
| 4 | repository | 写 `workflow_meta` 和 `workflow_draft` | 保证可查询和可编辑 |
| 5 | search event | 异步更新资源索引 | 搜索不阻塞创建主流程 |

### 保存工作流画布

**它是什么**：用户编辑画布后保存，系统更新草稿 canvas、输入输出参数、测试状态和 commitID。

**触发方式**：前端调用 `/api/workflow_api/save`。

```mermaid
sequenceDiagram
    participant 前端
    participant Handler
    participant WorkflowApp
    participant WorkflowDomain
    participant Repo

    前端->>Handler: SaveWorkflowRequest
    Handler->>WorkflowApp: SaveWorkflow
    WorkflowApp->>WorkflowApp: 校验用户和空间
    WorkflowApp->>WorkflowDomain: Save(workflowID, schema)
    WorkflowDomain->>WorkflowDomain: 反序列化 canvas
    WorkflowDomain->>WorkflowDomain: 提取 Entry/Exit 参数
    WorkflowDomain->>WorkflowDomain: 计算测试运行状态
    WorkflowDomain->>Repo: 生成 commitID
    WorkflowDomain->>Repo: 更新 workflow_draft
    Handler-->>前端: SaveWorkflowResponse
```

**关键步骤说明**：

| 步骤 | 负责模块 | 做什么 | 为什么在这一步 |
| --- | --- | --- | --- |
| 1 | domain/workflow | 解析 canvas | 后端需要理解画布结构 |
| 2 | domain/workflow | 提取输入输出参数 | OpenAPI、发布和工具化需要参数定义 |
| 3 | domain/workflow | 计算测试状态 | 画布逻辑改变后需要重新测试 |
| 4 | repository | 写 commitID | 让执行、快照、发布能绑定具体草稿 |

### 工作流试运行

**它是什么**：用户在编辑态试运行 workflow，API 立即返回 executeID，后端异步执行并记录过程。

**触发方式**：前端调用 `/api/workflow_api/test_run`。

```mermaid
sequenceDiagram
    participant 前端
    participant Handler
    participant WorkflowApp
    participant WorkflowDomain
    participant Compose
    participant Eino
    participant Repo

    前端->>Handler: WorkFlowTestRunRequest
    Handler->>WorkflowApp: TestRun
    WorkflowApp->>WorkflowApp: 构建 ExecuteConfig
    WorkflowApp->>WorkflowDomain: AsyncExecute
    WorkflowDomain->>Repo: 读取 draft workflow
    WorkflowDomain->>Compose: CanvasToWorkflowSchema / NewWorkflow
    Compose->>Eino: 编译 Workflow Runner
    WorkflowDomain->>Compose: NewWorkflowRunner.Prepare
    Compose->>Repo: 创建 workflow_execution
    WorkflowDomain->>Eino: AsyncRun
    Handler-->>前端: executeID
    Eino-->>Repo: 异步更新执行事件和状态
```

**关键步骤说明**：

| 步骤 | 负责模块 | 做什么 | 为什么在这一步 |
| --- | --- | --- | --- |
| 1 | application/workflow | 构建 Debug/Draft/Async 配置 | 试运行不应影响发布态 |
| 2 | domain/workflow | 画布转 WorkflowSchema | 执行引擎不直接吃前端 canvas |
| 3 | compose | 编译 Eino Workflow | 获得可执行 Runner |
| 4 | WorkflowRunner | 创建 execution 和事件处理 | 支撑查询进度、中断和失败记录 |
| 5 | Eino | 异步执行节点图 | API 可以快速返回 executeID |

### OpenAPI 运行工作流

**它是什么**：外部开发者用 API key 调用发布后的 workflow，同步或流式获取执行结果。

**触发方式**：`/v1/workflow/run`、`/v1/workflow/stream_run`、`/v1/workflows/chat`。

```mermaid
sequenceDiagram
    participant 调用方
    participant OpenapiAuth
    participant Handler
    participant WorkflowApp
    participant WorkflowDomain
    participant SSE

    调用方->>OpenapiAuth: Bearer API key
    OpenapiAuth->>OpenapiAuth: MD5 并校验权限
    OpenapiAuth->>Handler: 写入 OpenAPI auth context
    调用方->>Handler: workflow run 请求
    Handler->>Handler: 预处理 parameters
    Handler->>WorkflowApp: OpenAPIRun 或 OpenAPIStreamRun
    WorkflowApp->>WorkflowDomain: 执行发布版本
    WorkflowDomain-->>Handler: 响应或 StreamReader
    Handler-->>调用方: JSON 或 SSE
```

**关键步骤说明**：

| 步骤 | 负责模块 | 做什么 | 为什么在这一步 |
| --- | --- | --- | --- |
| 1 | OpenapiAuth | 校验 Bearer API key | 外部 API 不使用 Web cookie |
| 2 | handler | 预处理 parameters | 兼容对象参数和字符串参数 |
| 3 | application/workflow | 根据 OpenAPI 上下文构造执行配置 | 需要 connector、用户和发布态信息 |
| 4 | handler | SSE 转换 | HTTP 层负责流式协议输出 |

### Agent 对话运行

**它是什么**：用户在 Web 端和 Agent 对话，后端创建或校验会话，运行 AgentFlow 或 ChatFlow，并通过 SSE 返回消息 chunk。

**触发方式**：`/api/conversation/chat` 或 `/v3/chat`。

```mermaid
sequenceDiagram
    participant 用户
    participant Handler
    participant ConversationApp
    participant AgentRunDomain
    participant AgentRuntime
    participant AgentFlow
    participant Message
    participant SSE

    用户->>Handler: AgentRunRequest
    Handler->>ConversationApp: Run
    ConversationApp->>ConversationApp: 校验 Agent 和会话
    ConversationApp->>ConversationApp: 构建 AgentRunMeta
    ConversationApp->>AgentRunDomain: AgentRun
    AgentRunDomain->>AgentRuntime: goroutine Run
    AgentRuntime->>Message: 加载历史并保存输入
    AgentRuntime->>AgentFlow: SingleAgent 或 ChatFlow
    AgentFlow-->>AgentRuntime: 模型/工具/知识消息流
    AgentRuntime->>Message: 保存回答和工具消息
    AgentRuntime-->>SSE: message/done/error
    SSE-->>用户: 流式响应
```

**关键步骤说明**：

| 步骤 | 负责模块 | 做什么 | 为什么在这一步 |
| --- | --- | --- | --- |
| 1 | ConversationApp | 校验 Agent 和会话归属 | 防止跨用户会话访问 |
| 2 | ConversationApp | 解析多模态输入 | AgentFlow 需要结构化输入 |
| 3 | AgentRunDomain | 创建 StreamReader/Writer | 分离运行和 HTTP SSE 输出 |
| 4 | AgentRuntime | 加载历史、创建 run record | 多轮对话和运行审计需要 |
| 5 | AgentFlow | 模型、工具、知识库执行 | 真正生成回复 |
| 6 | MessageDomain | 保存消息 | 支持历史、重试、审计和前端展示 |

## 模块内部执行流程

### WorkflowRunner 内部流程

```mermaid
flowchart TD
    A["Prepare"] --> B{"是否恢复执行"}
    B -->|否| C["生成新 executeID"]
    B -->|是| D["读取 interrupt event 并校验 eventID"]
    D --> E["生成 state modifier"]
    E --> F["锁定 execution"]
    C --> G["创建 eventChan"]
    F --> G
    G --> H["配置 Eino 执行选项"]
    H --> I{"是否流式"}
    I -->|是| J["启动 StreamContainer"]
    I -->|否| K["跳过流容器"]
    J --> L["创建或复用 execution 记录"]
    K --> L
    L --> M["设置取消和超时"]
    M --> N["启动事件处理 goroutine"]
    N --> O["返回 cancelCtx/executeID/options/lastEventChan"]
```

### AgentRuntime 内部流程

```mermaid
flowchart TD
    A["Run"] --> B["获取 Agent 信息"]
    B --> C["处理 additional messages"]
    C --> D["加载历史消息"]
    D --> E["创建 run record"]
    E --> F["发送 created/in_progress"]
    F --> G["保存用户输入"]
    G --> H{"BotMode 是否 WorkflowMode"}
    H -->|是| I["ChatflowRun"]
    H -->|否| J["AgentStreamExecute"]
    I --> K["处理消息事件"]
    J --> K
    K --> L{"是否出错"}
    L -->|是| M["StepToFailed"]
    L -->|否| N["StepToComplete"]
```

## 异常分支

| 异常场景 | 触发条件 | 处理方式 | 对用户的影响 |
| --- | --- | --- | --- |
| Web API 缺少 session | 非登录注册 Web API 没有 cookie | SessionAuth 返回未授权 | 页面需要登录或请求失败 |
| OpenAPI key 无效 | Bearer 缺失、格式错误或校验失败 | OpenapiAuth 返回鉴权错误 | API 调用失败 |
| Workflow 参数无法绑定 | 请求结构不符合 IDL 模型 | handler 返回 invalid param | 前端或调用方看到参数错误 |
| Workflow 运行业务错误 | 抛出 `vo.WorkflowError` | OpenAPI 返回业务错误码和可能的 debug URL | 调用方可定位失败节点 |
| Agent 运行出错 | AgentRuntime 或下游工具出错 | SSE 发送 error 或错误消息 chunk | 聊天流中出现错误提示 |
| 中断恢复并发冲突 | 同一 execution 被多个恢复请求操作 | TryLockWorkflowExecution 失败 | 恢复请求失败，需要重试或刷新状态 |

