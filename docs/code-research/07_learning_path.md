# 代码阅读路径

## 快速理解

1. `README.zh_CN.md`：先读项目定位、功能范围和启动方式，读完后知道 Coze Studio 解决什么问题。
2. `backend/main.go`：理解后端启动、中间件和路由注册，读完后知道请求如何进入系统。
3. `backend/application/application.go`：理解服务装配顺序，读完后知道领域服务和基础设施如何被组合。
4. `frontend/apps/coze-studio/src/routes/index.tsx`：理解前端主体验入口，读完后知道页面如何映射到业务模块。
5. `backend/domain/workflow/interface.go`：理解 workflow 的完整领域能力，读完后知道为什么 workflow 是核心模块。

## 按目标索引

### 想理解整体架构

- `backend/main.go`：后端入口。
- `backend/application/application.go`：应用服务装配。
- `backend/application/base/appinfra/app_infra.go`：基础设施初始化。
- `frontend/apps/coze-studio/src/index.tsx`：前端入口。
- `frontend/apps/coze-studio/src/routes/async-components.tsx`：前端业务包懒加载关系。
- `docker/docker-compose.yml`：本地部署和外部依赖。

### 想理解工作流

- `backend/application/workflow/init.go`：workflow 应用服务初始化。
- `backend/domain/workflow/interface.go`：workflow service 和 repository 契约。
- `backend/domain/workflow/service/service_impl.go`：创建、保存、验证、元数据等核心逻辑。
- `backend/domain/workflow/service/executable_impl.go`：同步、异步、流式执行。
- `backend/domain/workflow/internal/canvas/adaptor/to_schema.go`：画布到执行 schema 的转换和节点适配器注册。
- `backend/domain/workflow/internal/compose/workflow.go`：Eino Workflow 编译。
- `backend/domain/workflow/internal/compose/workflow_run.go`：WorkflowRunner 执行准备、中断恢复和事件处理。

### 想理解智能体

- `backend/application/singleagent/init.go`：SingleAgent 应用层依赖。
- `backend/domain/agent/singleagent/service/single_agent.go`：草稿、发布和执行接口。
- `backend/domain/agent/singleagent/internal/agentflow/agent_flow_builder.go`：AgentFlow 图构建。
- `backend/domain/conversation/agentrun/internal/singleagent_run.go`：对话中调用 Agent 的流式执行。

### 想理解对话

- `backend/api/handler/coze/agent_run_service.go`：Web 和 OpenAPI 对话入口。
- `backend/application/conversation/agent_run.go`：会话校验、输入构造和 SSE chunk 转换。
- `backend/domain/conversation/agentrun/internal/run.go`：AgentRuntime 主流程。
- `backend/domain/conversation/message/service/message.go`：消息领域接口。
- `backend/domain/conversation/conversation/service/conversation.go`：会话领域接口。

### 想理解鉴权

- `backend/api/middleware/request_inspector.go`：请求分类。
- `backend/api/middleware/session.go`：Web session 鉴权和 admin 鉴权。
- `backend/api/middleware/openapi_auth.go`：OpenAPI Bearer key 鉴权。
- `backend/api/handler/coze/passport_service.go`：登录注册 HTTP 入口。
- `backend/application/user/user.go`：用户应用服务。

### 想理解前端

- `frontend/apps/coze-studio/src/index.tsx`：启动。
- `frontend/apps/coze-studio/src/app.tsx`：RouterProvider。
- `frontend/apps/coze-studio/src/layout.tsx`：全局布局初始化。
- `frontend/apps/coze-studio/src/routes/index.tsx`：空间、Agent IDE、Project IDE、Workflow、Explore 路由。
- `frontend/apps/coze-studio/src/routes/async-components.tsx`：业务包映射。
- `frontend/packages/agent-ide`：智能体编辑器。
- `frontend/packages/workflow`：工作流画布和试运行。
- `frontend/packages/project-ide`：应用式 IDE。

### 想理解数据模型

- `docker/volumes/mysql/schema.sql`：核心业务表。
- `backend/domain/*/entity`：领域实体。
- `backend/domain/*/repository`：仓储接口和实现。
- `backend/infra/eventbus`：异步事件。
- `backend/infra/storage`：对象存储。
- `backend/infra/document/searchstore`：知识库检索存储。

## 值得深入学习的代码片段

### `application.Init`

值得看，因为它把整个后端依赖图浓缩到一个函数里。读懂 Basic、Primary、Complex 三层后，很多领域服务之间的关系会自然清晰。

### `AgentFlow BuildAgent`

值得看，因为它展示了一个 Agent 如何从 Prompt、变量、知识库、插件、工作流、数据库工具动态组装成 Eino 图。它是“Agent 不只是模型调用”的最佳入口。

### `workflow.Save`

值得看，因为它展示了画布数据如何进入后端领域模型：解析 canvas、提取入口出口参数、计算测试状态、生成 commitID、保存草稿。

### `compose.NewWorkflow`

值得看，因为它是前端画布成为后端可执行图的关键：处理复合节点、普通节点、连接、checkpoint 和 Eino 编译。

### `WorkflowRunner.Prepare`

值得看，因为它聚合了执行 ID、恢复、中断、锁、超时、事件处理、执行记录和流式容器，是 workflow runtime 的心脏。

## 令人困惑的地方

### 全局服务变量较多

第一眼会困惑：为什么 handler 直接调用 `appworkflow.SVC`、`user.UserApplicationSVC` 这类全局服务。背后原因是项目偏运行时平台，启动时集中装配后通过默认服务简化调用。但二次开发和测试时需要注意替换和隔离。

### Repository 职责很宽

Workflow repository 不只是数据库访问，还包含 checkpoint、ID 生成、对象 URL、配置和建议能力。它的好处是 workflow runtime 拿一个 repository 就能完成执行；代价是接口变大，mock 和理解成本上升。

### 前端主应用很薄

刚开始看 `frontend/apps/coze-studio` 会觉得功能少。真正业务在 workspace 包里，主应用只负责启动、路由和懒加载。

### 消息队列默认值有多处表述

EventBus 注释说 RocketMQ 默认，但 Docker Compose 默认启动 NSQ，Helm 又提供 RocketMQ。实际选择由环境变量决定，阅读时要以部署配置为准。

