# Coze Studio 代码研究计划

## 研究摘要

Coze Studio 的核心价值是把 AI Agent 开发所需的模型、Prompt、RAG、插件、工作流、数据库、变量、调试、发布、会话和开放接口整合成一个可视化开发平台。它的后端以 Go/Hertz 为入口，以 `application.Init` 集中装配基础设施和领域服务，以 Eino 支撑 AgentFlow 与 Workflow runtime；前端以 React/Rush 多包组织，将主应用壳、基础能力、Agent IDE、Workflow 画布、Project IDE 和资源管理分开演进。

## 文档索引

- [01 架构全景](01_architecture.md)：整体模块、前后端边界、IDL、部署和核心抽象。
- [02 服务装配与跨领域调用](02_mechanism_service_assembly.md)：`application.Init`、Basic/Primary/Complex 服务分层、`crossdomain` 门面。
- [03 Agent、Workflow 与资源编排](03_mechanism_agent_workflow.md)：SingleAgent、AgentFlow、Workflow 生命周期、节点系统和中断恢复。
- [04 数据流与状态管理](04_data_flow.md)：MySQL 核心表、草稿/发布/执行状态、Agent 对话和 Workflow 数据生命周期。
- [05 依赖与生态](05_dependencies.md)：Hertz、Eino、GORM、Redis、Elasticsearch、Milvus、MinIO、EventBus、Rush 和部署生态。
- [06 核心工作流](06_workflow.md)：注册登录、创建/保存/试运行工作流、OpenAPI 运行、Agent 对话运行。
- [07 代码阅读路径](07_learning_path.md)：按目标推荐阅读顺序和关键代码片段。

## 设计亮点

- 后端初始化顺序清晰：先基础设施，再基础服务，再主要业务服务，再复杂聚合服务，最后注册跨领域门面。
- Agent 和 Workflow 都基于 Eino，但职责不同：AgentFlow 面向对话和工具调用，Workflow 面向确定性业务编排和可视化节点图。
- Workflow 的 `commitID` 连接草稿、快照、发布版本和执行记录，解决编辑态和运行态的一致性问题。
- 前端主应用很薄，业务复杂度主要沉到 Rush workspace 包里，便于按能力域拆分。
- 部署层把 MySQL/Redis/ES/Milvus/MinIO/MQ 等状态组件显式编排，适合本地和集群环境分别演进。

## 项目概述

**它是什么**：Coze Studio 是一个开源的一站式 AI Agent 可视化开发平台，提供智能体、应用、工作流、插件、知识库、数据库、模型服务和 OpenAPI/SDK 等能力。

**解决什么问题**：它把 AI Agent 开发中分散的模型接入、Prompt、RAG、插件调用、工作流编排、资源管理、调试发布和 API 集成放到同一个工程体系中。没有它，开发者通常需要分别搭建前端画布、后端编排服务、模型适配、知识库检索、插件系统、权限和部署基础设施。

**谁在使用**：面向希望用低代码或二次开发方式构建 AI Agent、工作流应用和企业 AI 产品的开发者与团队。

## 研究专题

### 专题 A：架构全景

- 目标：理解项目由哪些模块组成，每个模块是什么、为什么存在，以及前端、后端、IDL、部署目录之间如何协同。
- 范围：`README.zh_CN.md`、`Makefile`、`backend/main.go`、`backend/application`、`backend/domain`、`backend/infra`、`frontend/apps/coze-studio`、`frontend/packages`、`idl`、`docker`、`helm`。
- 输出：`docs/code-research/01_architecture.md`

### 专题 B：核心机制：服务装配与跨领域调用

- 目标：理解后端为什么把基础设施、应用服务、领域服务和跨领域门面分开，以及 `application.Init` 如何把系统装配起来。
- 范围：`backend/application/application.go`、`backend/application/base/appinfra`、`backend/crossdomain`、各 `application/*/init.go`。
- 输出：`docs/code-research/02_mechanism_service_assembly.md`

### 专题 C：核心机制：Agent、Workflow 与资源编排

- 目标：理解智能体、工作流、插件、知识库、变量、数据库等资源如何在前后端形成编排闭环。
- 范围：`backend/application/singleagent`、`backend/application/workflow`、`backend/domain/agent`、`backend/domain/workflow`、`backend/domain/plugin`、`backend/domain/knowledge`、`backend/domain/memory`、`frontend/packages/agent-ide`、`frontend/packages/workflow`、`frontend/apps/coze-studio/src/routes`。
- 输出：`docs/code-research/03_mechanism_agent_workflow.md`

### 专题 D：数据流与状态管理

- 目标：梳理核心业务对象、持久化表、缓存、事件总线、前端状态和端到端数据生命周期。
- 范围：`docker/volumes/mysql/schema.sql`、`docker/atlas/migrations`、`backend/domain/*/entity`、`backend/infra/rdb`、`backend/infra/eventbus`、`frontend/packages/*/stores`、`frontend/packages/*/context`。
- 输出：`docs/code-research/04_data_flow.md`

### 专题 E：依赖与生态

- 目标：理解核心依赖为什么被引入，它们分别解决哪些问题，以及外部系统边界如何划分。
- 范围：`backend/go.mod`、`rush.json`、`frontend/apps/coze-studio/package.json`、`docker/docker-compose.yml`、`helm/charts/opencoze`。
- 输出：`docs/code-research/05_dependencies.md`

### 专题 F：核心工作流

- 目标：追踪 3-5 个端到端流程：登录与鉴权、智能体编辑与发布、工作流创建/保存/试运行、知识库构建与检索、OpenAPI 对话或工作流调用。
- 范围：`backend/api/router`、`backend/api/handler`、`backend/application`、`backend/domain`、`idl`、`frontend/apps/coze-studio/src/routes`、既有 `docs/*workflow*` 文档。
- 输出：`docs/code-research/06_workflow.md`

### 专题 G：代码阅读路径

- 目标：为二次开发者总结最短阅读路径、按目标索引和最值得深入的代码片段。
- 范围：依赖以上专题结论。
- 输出：`docs/code-research/07_learning_path.md`

## 待解决疑问

- `router.GeneratedRegister` 由 Hertz/IDL 生成代码产生，但具体生成命令和开发流程仍需结合开发规范或脚本继续确认。
- `crossdomain` 默认服务门面用于降低领域间直接依赖，但也具有全局服务定位器特征，测试替换方式仍需继续验证。
- Python 环境在运行时承担什么职责，尤其是工作流代码节点还是其他执行逻辑？
- 前端 FlowGram 画布适配层与后端 workflow domain 的数据模型如何对齐？
- 事件总线在资源索引、项目搜索、知识库构建中的一致性保障和重试策略还需要继续细读消费者实现。
- `node_execution` 与 `workflow_execution` 的完整写入链路还需要继续追踪。
- OpenAPI path map 与 IDL 路由之间是否有自动同步机制，还是需要手工维护。
