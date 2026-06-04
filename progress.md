# coze-studio 代码研究进度

## 会话日志

| 时间 | 动作 | 结果 |
| --- | --- | --- |
| 2026-06-04 | 读取 `Coding:code-research` 技能说明 | 确认需要中文输出、Mermaid 图、先做已有研究检测 |
| 2026-06-04 | 读取 `planning-with-files` 技能说明 | 确认使用 `task_plan.md`、`findings.md`、`progress.md` 记录复杂研究 |
| 2026-06-04 | 初步扫描 workspace | 发现 `coze-studio` 与 `dify`，本次限定研究 `coze-studio` |
| 2026-06-04 | 阅读 `task_plan.md` 与 `README.zh_CN.md` | 确认项目定位、功能模块、技术栈声明与本地启动入口 |
| 2026-06-04 | 扫描顶层目录与文件列表 | 确认后端、前端、IDL、Docker、Helm、脚本和已有文档构成 |
| 2026-06-04 | 阅读 `Makefile` 与依赖文件列表 | 确认 Docker Web 路径、本地 Debug 路径、前后端构建脚本和 Atlas 数据库维护入口 |
| 2026-06-04 | 阅读 `backend/go.mod` 与 `rush.json` | 确认后端 AI/存储/消息依赖和前端 Rush monorepo 组织 |
| 2026-06-04 | 搜索服务入口并阅读前端主应用 package | 确认后端 `backend/main.go` 和前端 `@coze-studio/app` 主壳职责 |
| 2026-06-04 | 阅读后端 main 与前端路由 | 确认 HTTP 初始化链、中间件顺序、生成路由注册和主要前端体验入口 |
| 2026-06-04 | 扫描 application/domain/infra 目录 | 确认应用服务、领域服务和基础设施适配层的主要边界 |
| 2026-06-04 | 阅读 `application.Init` 与 `appinfra.Init` | 确认基础设施装配、服务分层装配、跨领域默认服务注册和工作流关键依赖 |
| 2026-06-04 | 阅读 code-research 模板 | 确认正式输出文档需要覆盖架构、机制、数据流、依赖、工作流和学习路径 |
| 2026-06-04 | 创建 `docs/code-research/RESEARCH_PLAN.md` | 完成阶段 1，形成 7 个专题的正式研究计划 |
| 2026-06-04 | 扫描 API 生成代码、handler 和静态资源注册 | 确认后端同时提供 API 与 SPA 静态资源，路由来自 IDL 生成 |
| 2026-06-04 | 阅读 workflow handler 与 OpenAPI 流式接口 | 确认 workflow 保存、试运行、OpenAPI 同步/流式运行、SSE 输出和错误转换模式 |
| 2026-06-04 | 阅读 workflow 初始化和文件结构 | 确认节点适配器、仓储、checkpoint、代码执行、Eino callback 和 workflow 内部包构成 |
| 2026-06-04 | 阅读 workflow 领域接口与核心服务实现 | 确认 Service/Repository 契约、组合实现、创建保存流程和子工作流验证 |
| 2026-06-04 | 阅读 workflow 执行和 compose runner | 确认画布转 schema、Eino 编译、执行准备、中断恢复、checkpoint、超时与事件处理链路 |
| 2026-06-04 | 阅读 workflow 节点适配器注册 | 确认节点类型、分支适配器和节点目录扩展方式 |
| 2026-06-04 | 阅读 SingleAgent 初始化、接口和 AgentFlow builder | 确认草稿/发布机制、Agent 资源依赖、Eino 图和 ReAct 工具模式 |
| 2026-06-04 | 阅读前端应用入口和布局 | 确认主应用薄壳、全局初始化、i18n、feature flags 和路由懒加载 |
| 2026-06-04 | 阅读前端懒加载组件和 packages 目录 | 确认前端业务能力按 foundation、arch、agent-ide、workflow、project-ide 等包族拆分 |
| 2026-06-04 | 阅读前端 API schema 包 | 确认 `idl2ts gen` 生成链路和开源包当前显式导出范围 |
| 2026-06-04 | 扫描 MySQL schema | 确认 Agent、App、Workflow、Knowledge、变量、会话等核心表和草稿/发布/执行状态建模 |
| 2026-06-04 | 抽取 MySQL 表名列表 | 确认当前 schema 有 51 张核心表，并形成数据流分组依据 |
| 2026-06-04 | 阅读 Docker Compose 和 Helm 服务列表 | 确认中间件边界、server/web 部署形态、MySQL/Atlas、ES 初始化、MinIO/Milvus/NSQ/RocketMQ/OceanBase 支持 |
| 2026-06-04 | 阅读 eventbus 接口和实现选择逻辑 | 确认通过环境变量选择 nsq/kafka/rmq/pulsar/nats，资源、应用、知识事件使用固定 topic/group 初始化 |
| 2026-06-04 | 阅读 passport handler 和 user application service | 确认注册、登录、Cookie session、用户上下文和头像上传流程 |
| 2026-06-04 | 阅读 request inspector、SessionAuth、OpenAPIAuth | 确认请求认证分流、静态资源跳过、Cookie session、Bearer API key 和 admin 白名单机制 |
| 2026-06-04 | 阅读对话运行 handler、application 和 agentrun runtime | 确认 Web/OpenAPI 对话、SSE、会话校验、消息构造、AgentRuntime 和 SingleAgent/ChatFlow 分支 |
| 2026-06-04 | 阅读已有 workflow create/save/test_run 调用链文档 | 校验代码阅读结论，并将其作为核心工作流专题参考 |
| 2026-06-04 | 新增 7 个专题研究文档 | 完成架构、服务装配、Agent/Workflow、数据流、依赖、核心工作流和阅读路径文档 |
| 2026-06-04 | 更新 `RESEARCH_PLAN.md` 摘要和索引 | 完成阶段 3 整合 |

## 文件变更

- 新增 `task_plan.md`
- 新增 `findings.md`
- 新增 `progress.md`
- 新增 `docs/code-research/RESEARCH_PLAN.md`
- 新增 `docs/code-research/01_architecture.md`
- 新增 `docs/code-research/02_mechanism_service_assembly.md`
- 新增 `docs/code-research/03_mechanism_agent_workflow.md`
- 新增 `docs/code-research/04_data_flow.md`
- 新增 `docs/code-research/05_dependencies.md`
- 新增 `docs/code-research/06_workflow.md`
- 新增 `docs/code-research/07_learning_path.md`

## 当前状态

本次代码研究已完成，后续可按未解决疑问继续做专题深挖。
