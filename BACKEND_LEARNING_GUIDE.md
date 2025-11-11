# Coze Studio 后端学习指南

> 本指南帮助你系统学习 Coze Studio 后端代码,从架构理解到实战开发。

## 📚 目录

- [1. 项目概览](#1-项目概览)
- [2. 技术栈](#2-技术栈)
- [3. 架构设计](#3-架构设计)
- [4. 目录结构详解](#4-目录结构详解)
- [5. 启动流程分析](#5-启动流程分析)
- [6. DDD 分层架构](#6-ddd-分层架构)
- [7. 核心领域模块](#7-核心领域模块)
- [8. 学习路径](#8-学习路径)
- [9. 实战练习](#9-实战练习)
- [10. 最佳实践](#10-最佳实践)

---

## 1. 项目概览

Coze Studio 后端是一个基于 Go 语言开发的 **AI Agent 开发平台后端服务**,采用:
- ✅ **DDD (领域驱动设计)** 架构
- ✅ **微服务化** 设计思想
- ✅ **清晰的分层结构**
- ✅ **Hertz HTTP 框架** (字节跳动开源高性能框架)

---

## 2. 技术栈

### 核心框架
- **Web 框架**: Cloudwego Hertz (高性能 HTTP 框架)
- **ORM**: GORM v1.25.11 (数据库操作)
- **Go 版本**: 1.24.0

### 数据存储
- **关系数据库**: MySQL 8.4.5
- **缓存**: Redis 8.0
- **搜索引擎**: Elasticsearch 8.18.0
- **向量数据库**: Milvus v2.5.10 (用于 AI Embeddings)
- **对象存储**: MinIO

### 消息队列 & 配置
- **消息队列**: NSQ
- **配置中心**: etcd 3.5

### AI 相关
- **Eino**: Cloudwego AI 框架
- 支持多种 LLM: OpenAI, Claude, Gemini, 火山方舟, 千问, DeepSeek, Ollama

### 其他工具库
- **UUID**: `github.com/google/uuid`
- **HTTP Client**: `github.com/go-resty/resty/v2`
- **深拷贝**: `github.com/mohae/deepcopy`
- **环境变量**: `github.com/joho/godotenv`
- **JSON**: `github.com/bytedance/sonic` (高性能)

---

## 3. 架构设计

### 3.1 整体架构图

```
┌─────────────────────────────────────────────────────────┐
│                     API Layer (api/)                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Router     │  │  Middleware  │  │   Handler    │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│              Application Layer (application/)           │
│  ┌──────────────────────────────────────────────────┐  │
│  │  应用服务 (协调领域服务，处理用例)                │  │
│  └──────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│               Domain Layer (domain/)                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │  Entity  │  │  Service │  │Repository│              │
│  └──────────┘  └──────────┘  └──────────┘              │
│  核心业务逻辑，独立于基础设施                           │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│          Infrastructure Layer (infra/)                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │    DB    │  │  Cache   │  │  Storage │              │
│  └──────────┘  └──────────┘  └──────────┘              │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│          Cross Domain Layer (crossdomain/)              │
│  跨域服务协调，不同领域之间的通信                       │
└─────────────────────────────────────────────────────────┘
```

### 3.2 DDD 分层职责

| 层级 | 目录 | 职责 | 依赖方向 |
|------|------|------|----------|
| **API 层** | `api/` | HTTP 路由、中间件、请求处理 | → Application |
| **应用层** | `application/` | 用例编排、事务管理 | → Domain |
| **领域层** | `domain/` | 核心业务逻辑、实体、领域服务 | → 无 (最纯净) |
| **基础设施层** | `infra/` | 数据库、缓存、消息队列等技术实现 | → Domain (接口) |
| **跨域层** | `crossdomain/` | 不同领域间的协调和通信 | → Domain |

---

## 4. 目录结构详解

### 4.1 根目录结构

```
backend/
├── api/                    # API 层：HTTP 接口定义
│   ├── handler/           # 请求处理器
│   ├── middleware/        # 中间件 (认证、日志、CORS 等)
│   ├── model/             # API 模型定义 (请求/响应)
│   └── router/            # 路由注册
│
├── application/           # 应用层：用例编排
│   ├── app/              # 应用管理
│   ├── conversation/     # 对话管理
│   ├── knowledge/        # 知识库管理
│   ├── plugin/           # 插件管理
│   ├── workflow/         # 工作流管理
│   └── ...               # 其他应用服务
│
├── domain/               # 领域层：核心业务逻辑 ⭐️ 最重要
│   ├── agent/           # Agent 领域
│   ├── conversation/    # 对话领域
│   ├── knowledge/       # 知识库领域
│   ├── plugin/          # 插件领域
│   ├── workflow/        # 工作流领域
│   └── ...              # 其他领域
│
├── infra/               # 基础设施层：技术实现
│   ├── cache/          # 缓存实现
│   ├── es/             # Elasticsearch
│   ├── orm/            # ORM 封装
│   ├── rdb/            # 关系数据库
│   ├── storage/        # 对象存储
│   ├── eventbus/       # 事件总线
│   └── ...             # 其他基础设施
│
├── crossdomain/        # 跨域层：领域间协调
│   ├── agent/         # Agent 跨域服务
│   ├── conversation/  # 对话跨域服务
│   ├── plugin/        # 插件跨域服务
│   └── ...            # 其他跨域服务
│
├── bizpkg/            # 业务通用包
│   ├── config/       # 配置管理
│   └── llm/          # LLM 模型构建器
│
├── pkg/              # 通用工具包
│   ├── errorx/      # 错误处理
│   ├── logs/        # 日志工具
│   ├── lang/        # 语言工具
│   └── ...          # 其他工具
│
├── types/           # 类型定义
│   ├── consts/     # 常量
│   ├── errno/      # 错误码
│   └── ddl/        # 数据库定义
│
├── conf/           # 配置文件
│   ├── model/     # 模型配置
│   ├── plugin/    # 插件配置
│   └── ...        # 其他配置
│
├── main.go        # 程序入口 ⭐️ 启动点
├── go.mod         # Go 模块定义
└── Dockerfile     # Docker 镜像构建
```

### 4.2 Domain 层内部结构 (以 knowledge 为例)

```
domain/knowledge/
├── entity/              # 实体定义 (核心领域对象)
│   ├── knowledge.go    # 知识库实体
│   ├── document.go     # 文档实体
│   ├── slice.go        # 切片实体
│   └── ...
│
├── service/            # 领域服务 (核心业务逻辑)
│   ├── knowledge.go   # 知识库服务
│   ├── retrieve.go    # 检索服务
│   └── ...
│
├── repository/         # 仓储接口 (数据访问抽象)
│   └── repository.go  # 定义数据访问接口
│
├── internal/          # 内部实现 (不对外暴露)
│   ├── dal/          # 数据访问层实现
│   ├── convert/      # 转换器
│   └── ...
│
└── processor/         # 处理器 (特定业务逻辑)
    └── interface.go  # 处理器接口
```

---

## 5. 启动流程分析

### 5.1 主函数流程 (`main.go`)

```go
func main() {
    ctx := context.Background()
    
    // 1. 设置崩溃日志输出
    setCrashOutput()
    
    // 2. 加载环境变量 (.env 文件)
    if err := loadEnv(); err != nil {
        panic("loadEnv failed")
    }
    
    // 3. 设置日志级别
    setLogLevel()
    
    // 4. 初始化应用 (核心初始化) ⭐️
    if err := application.Init(ctx); err != nil {
        panic("InitializeInfra failed")
    }
    
    // 5. 启动 HTTP 服务器
    startHttpServer()
}
```

### 5.2 应用初始化流程 (`application/application.go`)

```go
func Init(ctx context.Context) error {
    // 1. 初始化上下文缓存
    ctx = ctxcache.Init(ctx)
    
    // 2. 初始化基础设施 (数据库、缓存、消息队列等)
    infra, err := appinfra.Init(ctx)
    
    // 3. 初始化事件总线
    eventbus := initEventBus(infra)
    
    // 4. 初始化基础服务 (user, connector, upload 等)
    basicServices, err := initBasicServices(ctx, infra, eventbus)
    
    // 5. 初始化主要服务 (knowledge, plugin, workflow 等)
    primaryServices, err := initPrimaryServices(ctx, basicServices)
    
    // 6. 初始化复杂服务 (conversation, agent 等)
    complexServices, err := initComplexServices(ctx, primaryServices)
    
    // 7. 设置跨域服务 (供其他领域调用)
    crossdomain.SetDefaultSVC(...)
    
    return nil
}
```

### 5.3 HTTP 服务器启动流程 (`main.go`)

```go
func startHttpServer() {
    // 1. 创建 Hertz 服务器
    s := server.Default(opts...)
    
    // 2. 注册中间件 (顺序很重要!)
    s.Use(middleware.ContextCacheMW())     // 必须第一个
    s.Use(middleware.RequestInspectorMW()) // 必须第二个
    s.Use(middleware.SetHostMW())
    s.Use(middleware.SetLogIDMW())
    s.Use(corsHandler)                     // CORS
    s.Use(middleware.AccessLogMW())        // 访问日志
    s.Use(middleware.OpenapiAuthMW())      // OpenAPI 认证
    s.Use(middleware.SessionAuthMW())      // Session 认证
    s.Use(middleware.I18nMW())             // 国际化
    
    // 3. 注册路由
    router.GeneratedRegister(s)
    
    // 4. 启动服务 (阻塞)
    s.Spin()
}
```

### 5.4 初始化依赖关系图

```
main()
  └─> application.Init()
       ├─> appinfra.Init()          [基础设施: DB, Redis, ES, Milvus]
       │    ├─> initDatabase()
       │    ├─> initCache()
       │    ├─> initElasticsearch()
       │    └─> initMilvus()
       │
       ├─> initEventBus()           [事件总线]
       │
       ├─> initBasicServices()      [基础服务]
       │    ├─> UserService
       │    ├─> ConnectorService
       │    └─> UploadService
       │
       ├─> initPrimaryServices()    [主要服务]
       │    ├─> KnowledgeService    (依赖 UserService)
       │    ├─> PluginService
       │    ├─> WorkflowService
       │    └─> MemoryService
       │
       └─> initComplexServices()    [复杂服务]
            ├─> ConversationService  (依赖多个服务)
            └─> SingleAgentService
```

---

## 6. DDD 分层架构

### 6.1 领域层 (Domain Layer)

**职责**: 包含核心业务逻辑，是整个系统的核心。

#### Entity (实体)
```go
// domain/knowledge/entity/knowledge.go
type Knowledge struct {
    ID          int64
    SpaceID     int64
    Name        string
    Description string
    Type        KnowledgeType
    Status      KnowledgeStatus
    CreateTime  time.Time
    UpdateTime  time.Time
}

// 实体方法 (业务逻辑)
func (k *Knowledge) CanDelete() bool {
    return k.Status != KnowledgeStatusDeleted
}
```

#### Service (领域服务)
```go
// domain/knowledge/service/knowledge.go
type Service interface {
    // 创建知识库
    CreateKnowledge(ctx context.Context, req *CreateKnowledgeReq) (*entity.Knowledge, error)
    
    // 检索知识
    Retrieve(ctx context.Context, req *RetrieveReq) (*RetrieveResp, error)
}
```

#### Repository (仓储接口)
```go
// domain/knowledge/repository/repository.go
type Repository interface {
    // 保存知识库
    Save(ctx context.Context, knowledge *entity.Knowledge) error
    
    // 查询知识库
    FindByID(ctx context.Context, id int64) (*entity.Knowledge, error)
}
```

### 6.2 应用层 (Application Layer)

**职责**: 协调领域服务，处理用例，管理事务。

```go
// application/knowledge/knowledge.go
type KnowledgeApplicationService struct {
    domainSVC domain.Service
    eventBus  eventbus.EventBus
}

func (s *KnowledgeApplicationService) CreateKnowledgeUseCase(
    ctx context.Context, 
    req *CreateKnowledgeRequest,
) (*CreateKnowledgeResponse, error) {
    // 1. 参数验证
    if err := validateRequest(req); err != nil {
        return nil, err
    }
    
    // 2. 调用领域服务
    knowledge, err := s.domainSVC.CreateKnowledge(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // 3. 发布领域事件
    s.eventBus.Publish(ctx, KnowledgeCreatedEvent{
        KnowledgeID: knowledge.ID,
    })
    
    // 4. 返回响应
    return convertToResponse(knowledge), nil
}
```

### 6.3 API 层 (API Layer)

**职责**: 处理 HTTP 请求，参数解析，响应封装。

```go
// api/handler/coze/knowledge.go
type KnowledgeHandler struct {
    appSVC *application.KnowledgeApplicationService
}

func (h *KnowledgeHandler) CreateKnowledge(
    ctx context.Context,
    c *app.RequestContext,
) {
    // 1. 解析请求
    var req CreateKnowledgeAPIRequest
    if err := c.BindAndValidate(&req); err != nil {
        c.JSON(400, ErrorResponse(err))
        return
    }
    
    // 2. 调用应用服务
    resp, err := h.appSVC.CreateKnowledgeUseCase(ctx, &req)
    if err != nil {
        c.JSON(500, ErrorResponse(err))
        return
    }
    
    // 3. 返回响应
    c.JSON(200, SuccessResponse(resp))
}
```

### 6.4 基础设施层 (Infrastructure Layer)

**职责**: 实现领域层定义的接口，提供技术实现。

```go
// domain/knowledge/internal/dal/knowledge_impl.go
type KnowledgeRepositoryImpl struct {
    db *gorm.DB
}

func (r *KnowledgeRepositoryImpl) Save(
    ctx context.Context,
    knowledge *entity.Knowledge,
) error {
    // 使用 GORM 保存到数据库
    return r.db.WithContext(ctx).Create(knowledge).Error
}
```

### 6.5 跨域层 (Cross Domain Layer)

**职责**: 协调不同领域之间的交互。

```go
// crossdomain/knowledge/contract.go
type Service interface {
    // 供其他领域调用的接口
    GetKnowledgeInfo(ctx context.Context, knowledgeID int64) (*model.KnowledgeInfo, error)
}

// crossdomain/knowledge/impl/service.go
type ServiceImpl struct {
    domainSVC domain.Service
}

func (s *ServiceImpl) GetKnowledgeInfo(
    ctx context.Context,
    knowledgeID int64,
) (*model.KnowledgeInfo, error) {
    return s.domainSVC.GetKnowledge(ctx, knowledgeID)
}
```

---

## 7. 核心领域模块

### 7.1 Agent 领域 (Single Agent)

**位置**: `domain/agent/singleagent/`

**职责**: 管理单个 AI Agent 的生命周期

**核心概念**:
- **Entity**: Agent 实体，包含 Agent 配置、能力
- **Service**: Agent 创建、更新、发布、运行
- **AgentFlow**: Agent 执行流程编排

**关键文件**:
```
domain/agent/singleagent/
├── entity/
│   ├── agent.go           # Agent 实体
│   ├── config.go          # Agent 配置
│   └── capability.go      # Agent 能力
├── service/
│   ├── agent_service.go   # Agent 服务
│   ├── publish.go         # 发布服务
│   └── run.go             # 运行服务
└── internal/
    └── agentflow/         # Agent 流程编排
```

### 7.2 Workflow 领域

**位置**: `domain/workflow/`

**职责**: 工作流引擎，支持复杂的流程编排

**核心概念**:
- **Node**: 工作流节点 (LLM、插件、条件、循环等)
- **Edge**: 节点连接
- **Execution**: 工作流执行
- **Variable**: 工作流变量

**关键文件**:
```
domain/workflow/
├── entity/
│   ├── workflow.go        # 工作流实体
│   ├── node_meta.go       # 节点元数据
│   └── workflow_execution.go  # 执行记录
├── internal/
│   ├── nodes/            # 各种节点实现
│   │   ├── llm_node.go
│   │   ├── plugin_node.go
│   │   ├── if_node.go
│   │   └── ...
│   ├── execute/          # 执行引擎
│   └── canvas/           # 画布 (前端可视化)
└── service/
    └── workflow_service.go
```

### 7.3 Knowledge 领域

**位置**: `domain/knowledge/`

**职责**: 知识库管理，文档解析、切片、检索

**核心概念**:
- **Knowledge**: 知识库
- **Document**: 文档
- **Slice**: 文档切片 (Chunk)
- **Retrieve**: 向量检索

**关键文件**:
```
domain/knowledge/
├── entity/
│   ├── knowledge.go      # 知识库实体
│   ├── document.go       # 文档实体
│   └── slice.go          # 切片实体
├── service/
│   ├── knowledge.go      # 知识库服务
│   ├── retrieve.go       # 检索服务
│   └── sheet.go          # 表格处理
├── processor/            # 文档处理器
│   └── impl/
│       ├── pdf_processor.go
│       ├── word_processor.go
│       └── excel_processor.go
└── internal/
    └── dal/              # 数据访问
```

### 7.4 Plugin 领域

**位置**: `domain/plugin/`

**职责**: 插件管理，API 工具管理

**核心概念**:
- **Plugin**: 插件
- **Tool**: API 工具
- **OAuth**: OAuth 认证
- **API Management**: API 管理

**关键文件**:
```
domain/plugin/
├── entity/
│   ├── plugin.go         # 插件实体
│   └── tool.go           # 工具实体
├── service/
│   ├── plugin_draft.go   # 草稿管理
│   ├── plugin_release.go # 发布管理
│   ├── plugin_oauth.go   # OAuth 认证
│   ├── exec_tool.go      # 工具执行
│   └── agent_tool.go     # Agent 工具
├── repository/
│   ├── plugin_repository.go
│   └── tool_repository.go
└── dto/                  # 数据传输对象
```

### 7.5 Conversation 领域

**位置**: `domain/conversation/`

**职责**: 对话管理、消息管理、Agent 运行管理

**核心子域**:
1. **Conversation**: 对话会话
2. **Message**: 消息
3. **AgentRun**: Agent 运行记录

**关键文件**:
```
domain/conversation/
├── conversation/         # 对话子域
│   ├── entity/
│   │   └── conversation.go
│   └── service/
│       └── conversation_service.go
│
├── message/             # 消息子域
│   ├── entity/
│   │   └── message.go
│   └── service/
│       └── message_service.go
│
└── agentrun/            # Agent 运行子域
    ├── entity/
    │   └── agent_run.go
    └── service/
        └── agent_run_service.go
```

### 7.6 Memory 领域

**位置**: `domain/memory/`

**职责**: 记忆管理 (变量、数据库)

**核心子域**:
1. **Variables**: 变量管理
2. **Database**: 数据库管理

**关键文件**:
```
domain/memory/
├── variables/           # 变量子域
│   ├── entity/
│   ├── service/
│   └── repository/
│
└── database/            # 数据库子域
    ├── entity/
    ├── service/
    └── repository/
```

### 7.7 User 领域

**位置**: `domain/user/`

**职责**: 用户管理、认证、空间管理

**核心概念**:
- **User**: 用户
- **Session**: 会话
- **Space**: 工作空间

---

## 8. 学习路径

### 阶段一: 基础理解 (1-2 天)

#### 1. 理解项目结构
- ✅ 阅读本文档的前 6 章
- ✅ 熟悉目录结构
- ✅ 了解 DDD 分层架构

#### 2. 运行项目
```bash
# 1. 启动中间件
make middleware

# 2. 启动后端服务
make server

# 3. 查看日志，理解启动流程
```

#### 3. 阅读启动代码
- `main.go` - 程序入口
- `application/application.go` - 应用初始化
- `application/base/appinfra/app_infra.go` - 基础设施初始化

#### 4. 理解请求流程
跟踪一个简单的 API 请求:
```
HTTP 请求 → 中间件 → 路由 → Handler → Application Service → Domain Service → Repository → 数据库
```

### 阶段二: 深入 Domain (3-5 天)

#### 1. 选择一个简单领域学习 (推荐: User)

**User 领域学习清单**:
```
1. domain/user/entity/user.go           # 理解用户实体
2. domain/user/service/user.go          # 理解用户服务接口
3. domain/user/service/user_impl.go     # 理解服务实现
4. domain/user/repository/repository.go # 理解仓储接口
5. domain/user/internal/dal/           # 理解数据访问实现
```

#### 2. 理解 Entity → Service → Repository 模式

**练习**: 画出 User 领域的类图:
```
┌─────────────┐
│   User      │  Entity (实体)
│ ─────────── │
│ + ID        │
│ + Username  │
│ + Email     │
└─────────────┘
       ▲
       │ uses
┌─────────────┐
│UserService  │  Service (服务)
│ ─────────── │
│ + Create()  │
│ + GetByID() │
└─────────────┘
       │ uses
       ▼
┌─────────────┐
│Repository   │  Repository (仓储)
│ ─────────── │
│ + Save()    │
│ + FindByID()│
└─────────────┘
```

#### 3. 学习一个复杂领域 (推荐: Knowledge)

**Knowledge 领域学习路线**:
1. 阅读 `entity/knowledge.go` - 理解知识库实体
2. 阅读 `entity/document.go` - 理解文档实体
3. 阅读 `service/knowledge.go` - 理解知识库服务
4. 阅读 `service/retrieve.go` - 理解向量检索
5. 阅读 `processor/` - 理解文档处理器

#### 4. 理解跨域调用

查看其他领域如何调用 Knowledge:
```go
// 在 agent 领域中调用 knowledge
import "github.com/coze-dev/coze-studio/backend/crossdomain/knowledge"

knowledgeInfo, err := knowledge.GetDefaultSVC().GetKnowledgeInfo(ctx, knowledgeID)
```

### 阶段三: 理解应用层 (2-3 天)

#### 1. 学习应用服务

以 `application/knowledge/knowledge.go` 为例:
```go
type KnowledgeApplicationService struct {
    // 依赖注入
    domainSVC    domain.Service      // 领域服务
    eventBus     eventbus.EventBus   // 事件总线
    uploadSVC    upload.Service      // 上传服务
}

// 用例：创建知识库
func (s *KnowledgeApplicationService) CreateKnowledge(...) {
    // 1. 参数校验
    // 2. 调用领域服务
    // 3. 发布事件
    // 4. 返回结果
}
```

#### 2. 理解依赖注入

查看 `application/knowledge/init.go`:
```go
func InitKnowledgeService(
    domainSVC domain.Service,
    eventBus eventbus.EventBus,
    uploadSVC upload.Service,
) *KnowledgeApplicationService {
    return &KnowledgeApplicationService{
        domainSVC: domainSVC,
        eventBus:  eventBus,
        uploadSVC: uploadSVC,
    }
}
```

#### 3. 理解事件总线

学习事件的发布和订阅:
```go
// 发布事件
s.eventBus.Publish(ctx, KnowledgeCreatedEvent{...})

// 订阅事件
eventBus.Subscribe(KnowledgeCreatedEvent, handleKnowledgeCreated)
```

### 阶段四: 理解 API 层 (1-2 天)

#### 1. 学习 Handler

查看 `api/handler/coze/knowledge.go`:
```go
type KnowledgeHandler struct {
    appSVC *application.KnowledgeApplicationService
}

func (h *KnowledgeHandler) CreateKnowledge(ctx context.Context, c *app.RequestContext) {
    // 1. 参数绑定
    var req CreateKnowledgeRequest
    c.BindAndValidate(&req)
    
    // 2. 调用应用服务
    resp, err := h.appSVC.CreateKnowledge(ctx, &req)
    
    // 3. 返回响应
    c.JSON(200, resp)
}
```

#### 2. 学习中间件

查看 `api/middleware/`:
- `session.go` - Session 认证
- `openapi_auth.go` - OpenAPI 认证
- `log.go` - 日志中间件
- `i18n.go` - 国际化中间件

#### 3. 学习路由注册

查看 `api/router/coze/router.go`:
```go
func Register(r *server.Hertz) {
    // 知识库路由组
    knowledge := r.Group("/api/knowledge")
    {
        knowledge.POST("/create", handler.CreateKnowledge)
        knowledge.GET("/list", handler.ListKnowledge)
        knowledge.POST("/delete", handler.DeleteKnowledge)
    }
}
```

### 阶段五: 核心业务理解 (5-7 天)

#### 1. 深入 Workflow 引擎

**学习重点**:
- 工作流节点类型
- 节点执行引擎
- 变量传递机制
- 条件判断和循环

**关键文件**:
```
domain/workflow/
├── internal/nodes/      # 各种节点实现 ⭐️
├── internal/execute/    # 执行引擎 ⭐️
└── variable/           # 变量系统 ⭐️
```

#### 2. 深入 Agent 执行流程

**学习重点**:
- Agent 配置加载
- Agent 执行流程
- 工具调用机制
- 流式输出

**关键文件**:
```
domain/agent/singleagent/
├── internal/agentflow/  # Agent 执行流程 ⭐️
└── service/            # Agent 服务
```

#### 3. 深入 Plugin 系统

**学习重点**:
- API 工具定义
- OAuth 认证流程
- 工具执行
- 参数验证

**关键文件**:
```
domain/plugin/
├── service/exec_tool.go      # 工具执行 ⭐️
├── service/plugin_oauth.go   # OAuth 认证 ⭐️
└── service/agent_tool.go     # Agent 工具集成
```

### 阶段六: 高级主题 (3-5 天)

#### 1. 事件驱动架构

学习事件的使用:
```go
// infra/eventbus/eventbus.go
type EventBus interface {
    Publish(ctx context.Context, event Event) error
    Subscribe(eventType string, handler EventHandler) error
}

// 事件定义
type KnowledgeCreatedEvent struct {
    KnowledgeID int64
    Name        string
    CreateTime  time.Time
}

// 事件处理
func handleKnowledgeCreated(ctx context.Context, event Event) error {
    // 处理知识库创建事件
    // 例如：更新搜索索引、发送通知等
}
```

#### 2. 向量检索 (Milvus)

学习向量数据库的使用:
```go
// infra/embedding/
- client.go           # Milvus 客户端
- collection.go       # Collection 管理
- search.go          # 向量检索
```

#### 3. LLM 集成

学习 LLM 模型的使用:
```go
// bizpkg/llm/modelbuilder/
- builder.go         # 模型构建器
- openai.go         # OpenAI 集成
- claude.go         # Claude 集成
```

#### 4. 流式输出 (SSE)

学习 Server-Sent Events:
```go
// infra/sse/
- sse.go            # SSE 实现
```

---

## 9. 实战练习

### 练习 1: 添加一个新的 API 端点

**目标**: 为 User 领域添加一个"获取用户统计信息"的 API

**步骤**:

#### 1. 在 Domain 层添加方法

```go
// domain/user/service/user.go
type Service interface {
    // ... 现有方法 ...
    
    // 获取用户统计信息
    GetUserStatistics(ctx context.Context, userID int64) (*UserStatistics, error)
}

// domain/user/entity/user.go
type UserStatistics struct {
    UserID         int64
    AgentCount     int
    WorkflowCount  int
    KnowledgeCount int
}
```

#### 2. 实现领域服务

```go
// domain/user/service/user_impl.go
func (s *ServiceImpl) GetUserStatistics(ctx context.Context, userID int64) (*entity.UserStatistics, error) {
    // 从数据库查询统计信息
    stats, err := s.repo.GetStatistics(ctx, userID)
    if err != nil {
        return nil, err
    }
    return stats, nil
}
```

#### 3. 在 Repository 添加方法

```go
// domain/user/repository/repository.go
type Repository interface {
    // ... 现有方法 ...
    GetStatistics(ctx context.Context, userID int64) (*entity.UserStatistics, error)
}
```

#### 4. 实现 Repository

```go
// domain/user/internal/dal/user_repo_impl.go
func (r *RepositoryImpl) GetStatistics(ctx context.Context, userID int64) (*entity.UserStatistics, error) {
    var stats entity.UserStatistics
    
    // 查询 Agent 数量
    err := r.db.WithContext(ctx).
        Model(&Agent{}).
        Where("user_id = ?", userID).
        Count(&stats.AgentCount).Error
    if err != nil {
        return nil, err
    }
    
    // 查询 Workflow 数量
    // ...
    
    return &stats, nil
}
```

#### 5. 在 Application 层添加用例

```go
// application/user/user.go
func (s *UserApplicationService) GetUserStatistics(ctx context.Context, userID int64) (*GetUserStatisticsResponse, error) {
    // 调用领域服务
    stats, err := s.domainSVC.GetUserStatistics(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // 转换为响应对象
    return &GetUserStatisticsResponse{
        UserID:         stats.UserID,
        AgentCount:     stats.AgentCount,
        WorkflowCount:  stats.WorkflowCount,
        KnowledgeCount: stats.KnowledgeCount,
    }, nil
}
```

#### 6. 在 API 层添加 Handler

```go
// api/handler/coze/user.go
func (h *UserHandler) GetUserStatistics(ctx context.Context, c *app.RequestContext) {
    // 从请求中获取 userID
    userID := c.Query("user_id")
    
    // 调用应用服务
    resp, err := h.appSVC.GetUserStatistics(ctx, conv.StrToInt64(userID))
    if err != nil {
        c.JSON(500, ErrorResponse(err))
        return
    }
    
    // 返回响应
    c.JSON(200, SuccessResponse(resp))
}
```

#### 7. 注册路由

```go
// api/router/coze/user.go
func registerUserRoutes(r *server.Hertz, handler *UserHandler) {
    user := r.Group("/api/user")
    {
        // ... 现有路由 ...
        user.GET("/statistics", handler.GetUserStatistics)
    }
}
```

#### 8. 测试

```bash
curl "http://localhost:8888/api/user/statistics?user_id=123"
```

### 练习 2: 实现一个简单的 Workflow 节点

**目标**: 实现一个"文本转大写"的 Workflow 节点

**步骤**:

#### 1. 定义节点类型

```go
// domain/workflow/entity/node_meta.go
const (
    NodeTypeUpperCase = "upper_case"
)
```

#### 2. 实现节点

```go
// domain/workflow/internal/nodes/upper_case_node.go
package nodes

import (
    "context"
    "strings"
)

type UpperCaseNode struct {
    BaseNode
}

func NewUpperCaseNode() *UpperCaseNode {
    return &UpperCaseNode{}
}

func (n *UpperCaseNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
    // 获取输入文本
    text, ok := input["text"].(string)
    if !ok {
        return nil, errors.New("invalid input: text is required")
    }
    
    // 转换为大写
    result := strings.ToUpper(text)
    
    // 返回结果
    return map[string]interface{}{
        "result": result,
    }, nil
}

func (n *UpperCaseNode) GetType() string {
    return entity.NodeTypeUpperCase
}
```

#### 3. 注册节点

```go
// domain/workflow/internal/nodes/registry.go
func init() {
    RegisterNode(entity.NodeTypeUpperCase, NewUpperCaseNode)
}
```

#### 4. 测试节点

```go
// domain/workflow/internal/nodes/upper_case_node_test.go
func TestUpperCaseNode(t *testing.T) {
    node := NewUpperCaseNode()
    
    input := map[string]interface{}{
        "text": "hello world",
    }
    
    output, err := node.Execute(context.Background(), input)
    assert.NoError(t, err)
    assert.Equal(t, "HELLO WORLD", output["result"])
}
```

### 练习 3: 实现一个事件处理器

**目标**: 当知识库创建时，自动创建搜索索引

**步骤**:

#### 1. 定义事件

```go
// domain/knowledge/entity/event.go
type KnowledgeCreatedEvent struct {
    KnowledgeID int64
    Name        string
    SpaceID     int64
    CreateTime  time.Time
}

func (e KnowledgeCreatedEvent) EventType() string {
    return "knowledge.created"
}
```

#### 2. 实现事件处理器

```go
// domain/search/service/handler_knowledge.go
type KnowledgeEventHandler struct {
    searchSVC Service
}

func (h *KnowledgeEventHandler) HandleKnowledgeCreated(ctx context.Context, event knowledge.KnowledgeCreatedEvent) error {
    logs.Infof("[KnowledgeEventHandler] handle knowledge created, id=%d", event.KnowledgeID)
    
    // 创建搜索索引
    err := h.searchSVC.CreateKnowledgeIndex(ctx, &CreateIndexRequest{
        KnowledgeID: event.KnowledgeID,
        Name:        event.Name,
        SpaceID:     event.SpaceID,
    })
    
    if err != nil {
        logs.Errorf("[KnowledgeEventHandler] create index failed, err=%v", err)
        return err
    }
    
    logs.Infof("[KnowledgeEventHandler] create index success, id=%d", event.KnowledgeID)
    return nil
}
```

#### 3. 订阅事件

```go
// application/search/init.go
func InitSearchService(eventBus eventbus.EventBus, searchSVC domain.Service) {
    handler := &KnowledgeEventHandler{
        searchSVC: searchSVC,
    }
    
    // 订阅知识库创建事件
    eventBus.Subscribe("knowledge.created", handler.HandleKnowledgeCreated)
}
```

#### 4. 发布事件

```go
// application/knowledge/knowledge.go
func (s *KnowledgeApplicationService) CreateKnowledge(ctx context.Context, req *CreateKnowledgeRequest) (*CreateKnowledgeResponse, error) {
    // 创建知识库
    knowledge, err := s.domainSVC.CreateKnowledge(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // 发布事件
    s.eventBus.Publish(ctx, knowledge.KnowledgeCreatedEvent{
        KnowledgeID: knowledge.ID,
        Name:        knowledge.Name,
        SpaceID:     knowledge.SpaceID,
        CreateTime:  knowledge.CreateTime,
    })
    
    return convertToResponse(knowledge), nil
}
```

---

## 10. 最佳实践

### 10.1 代码组织

#### ✅ DO: 遵循 DDD 分层

```go
// ✅ 好的做法：清晰的分层
// Domain 层只包含业务逻辑
package domain

type KnowledgeService interface {
    CreateKnowledge(ctx context.Context, req *CreateKnowledgeReq) (*Knowledge, error)
}

// Application 层协调领域服务
package application

type KnowledgeAppService struct {
    domainSVC domain.KnowledgeService
}

func (s *KnowledgeAppService) CreateKnowledge(ctx context.Context, req *CreateKnowledgeReq) (*CreateKnowledgeResp, error) {
    // 调用领域服务
    knowledge, err := s.domainSVC.CreateKnowledge(ctx, req)
    // 其他协调逻辑
}
```

#### ❌ DON'T: 跨层调用

```go
// ❌ 不好的做法：API 层直接调用 Repository
package api

func (h *Handler) CreateKnowledge(ctx context.Context, c *app.RequestContext) {
    // 直接调用 Repository，跳过了业务逻辑层
    knowledge := &Knowledge{}
    h.repo.Save(ctx, knowledge) // 不要这样做！
}
```

### 10.2 依赖注入

#### ✅ DO: 使用接口依赖

```go
// ✅ 好的做法：依赖接口而非具体实现
type KnowledgeService struct {
    repo       KnowledgeRepository    // 接口
    uploadSVC  upload.Service         // 接口
}

func NewKnowledgeService(
    repo KnowledgeRepository,
    uploadSVC upload.Service,
) *KnowledgeService {
    return &KnowledgeService{
        repo:      repo,
        uploadSVC: uploadSVC,
    }
}
```

#### ❌ DON'T: 依赖具体实现

```go
// ❌ 不好的做法：依赖具体实现
type KnowledgeService struct {
    repo *KnowledgeRepositoryImpl  // 具体实现，难以测试和替换
}
```

### 10.3 错误处理

#### ✅ DO: 使用自定义错误类型

```go
// ✅ 好的做法：使用自定义错误
var (
    ErrKnowledgeNotFound = errorx.New(errno.KnowledgeNotFound, "knowledge not found")
    ErrKnowledgeExists   = errorx.New(errno.KnowledgeExists, "knowledge already exists")
)

func (s *Service) GetKnowledge(ctx context.Context, id int64) (*Knowledge, error) {
    knowledge, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrKnowledgeNotFound
        }
        return nil, err
    }
    return knowledge, nil
}
```

#### ❌ DON'T: 直接返回原始错误

```go
// ❌ 不好的做法：直接返回 GORM 错误
func (s *Service) GetKnowledge(ctx context.Context, id int64) (*Knowledge, error) {
    knowledge, err := s.repo.FindByID(ctx, id)
    return knowledge, err  // 泄露了基础设施层的错误
}
```

### 10.4 日志记录

#### ✅ DO: 记录关键信息

```go
// ✅ 好的做法：记录关键操作和错误
func (s *Service) CreateKnowledge(ctx context.Context, req *CreateKnowledgeReq) (*Knowledge, error) {
    logs.Infof("[CreateKnowledge] start, name=%s, spaceID=%d", req.Name, req.SpaceID)
    
    knowledge, err := s.repo.Save(ctx, req)
    if err != nil {
        logs.Errorf("[CreateKnowledge] failed, err=%v", err)
        return nil, err
    }
    
    logs.Infof("[CreateKnowledge] success, knowledgeID=%d", knowledge.ID)
    return knowledge, nil
}
```

### 10.5 测试

#### ✅ DO: 编写单元测试

```go
// ✅ 好的做法：为领域服务编写测试
func TestKnowledgeService_CreateKnowledge(t *testing.T) {
    // 使用 mock
    repo := mock.NewMockKnowledgeRepository(t)
    svc := NewKnowledgeService(repo)
    
    // 设置 mock 行为
    repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
    
    // 测试
    knowledge, err := svc.CreateKnowledge(context.Background(), &CreateKnowledgeReq{
        Name: "test",
    })
    
    assert.NoError(t, err)
    assert.NotNil(t, knowledge)
}
```

### 10.6 数据库操作

#### ✅ DO: 使用事务

```go
// ✅ 好的做法：使用事务保证一致性
func (s *Service) CreateKnowledgeWithDocuments(ctx context.Context, req *CreateKnowledgeReq) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 创建知识库
        knowledge := &Knowledge{Name: req.Name}
        if err := tx.Create(knowledge).Error; err != nil {
            return err
        }
        
        // 创建文档
        for _, doc := range req.Documents {
            doc.KnowledgeID = knowledge.ID
            if err := tx.Create(doc).Error; err != nil {
                return err  // 自动回滚
            }
        }
        
        return nil
    })
}
```

### 10.7 性能优化

#### ✅ DO: 使用批量操作

```go
// ✅ 好的做法：批量插入
func (r *Repository) BatchCreate(ctx context.Context, documents []*Document) error {
    // 使用 GORM 的 CreateInBatches，每次插入 100 条
    return r.db.WithContext(ctx).CreateInBatches(documents, 100).Error
}
```

#### ❌ DON'T: 循环单条插入

```go
// ❌ 不好的做法：循环插入，性能差
func (r *Repository) BatchCreate(ctx context.Context, documents []*Document) error {
    for _, doc := range documents {
        if err := r.db.WithContext(ctx).Create(doc).Error; err != nil {
            return err
        }
    }
    return nil
}
```

### 10.8 并发安全

#### ✅ DO: 使用并发安全的数据结构

```go
// ✅ 好的做法：使用 sync.Map
type Cache struct {
    data sync.Map
}

func (c *Cache) Set(key string, value interface{}) {
    c.data.Store(key, value)
}

func (c *Cache) Get(key string) (interface{}, bool) {
    return c.data.Load(key)
}
```

### 10.9 Context 使用

#### ✅ DO: 正确传递 Context

```go
// ✅ 好的做法：始终传递 context
func (s *Service) ProcessKnowledge(ctx context.Context, id int64) error {
    // 传递 context
    knowledge, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return err
    }
    
    // 传递 context 到下游
    return s.processSVC.Process(ctx, knowledge)
}
```

#### ❌ DON'T: 使用 context.Background()

```go
// ❌ 不好的做法：在中间层使用 Background
func (s *Service) ProcessKnowledge(ctx context.Context, id int64) error {
    // 不要重新创建 context，会丢失上游的取消信号
    knowledge, err := s.repo.FindByID(context.Background(), id)
    return err
}
```

---

## 11. 常见问题

### Q1: Domain 层可以调用 Application 层吗？

**A**: ❌ 不可以。依赖方向是:
```
API → Application → Domain → Infrastructure (仅接口)
```

Domain 层是最纯净的，不应该依赖任何其他层。

### Q2: 如何在不同领域间共享逻辑？

**A**: 使用 **Cross Domain Layer**:
```go
// 在 crossdomain/knowledge 中定义接口
type Service interface {
    GetKnowledge(ctx context.Context, id int64) (*model.Knowledge, error)
}

// 在其他领域调用
knowledge, err := crossknowledge.GetDefaultSVC().GetKnowledge(ctx, id)
```

### Q3: Repository 应该返回 Entity 还是 DTO？

**A**: Repository 应该返回 **Domain Entity**:
```go
// ✅ 正确
type Repository interface {
    FindByID(ctx context.Context, id int64) (*entity.Knowledge, error)
}

// ❌ 错误
type Repository interface {
    FindByID(ctx context.Context, id int64) (*dto.KnowledgeDTO, error)
}
```

### Q4: 什么时候使用事件总线？

**A**: 当需要 **异步处理** 或 **解耦** 时使用:
- ✅ 知识库创建后，异步创建搜索索引
- ✅ 用户注册后，发送欢迎邮件
- ❌ 查询用户信息（同步操作，不需要事件）

### Q5: 如何处理分页查询？

**A**: 在 Repository 层定义分页接口:
```go
type PageQuery struct {
    Page     int
    PageSize int
}

type PageResult struct {
    Total int64
    Items interface{}
}

func (r *Repository) FindByPage(ctx context.Context, query *PageQuery) (*PageResult, error) {
    var total int64
    var items []*entity.Knowledge
    
    db := r.db.WithContext(ctx).Model(&entity.Knowledge{})
    
    // 查询总数
    db.Count(&total)
    
    // 查询数据
    db.Offset((query.Page - 1) * query.PageSize).
       Limit(query.PageSize).
       Find(&items)
    
    return &PageResult{
        Total: total,
        Items: items,
    }, nil
}
```

---

## 12. 扩展阅读

### 12.1 DDD 相关
- 《领域驱动设计》 - Eric Evans
- 《实现领域驱动设计》 - Vaughn Vernon

### 12.2 Go 相关
- 《Go 语言高级编程》
- 《Go 语言设计与实现》

### 12.3 架构相关
- 《Clean Architecture》 - Robert C. Martin
- 《微服务架构设计模式》

### 12.4 项目相关文档
- `CLAUDE.md` - 项目整体介绍
- `README.md` - 快速开始指南
- `docs/` - 详细文档

---

## 13. 总结

### 核心要点
1. ✅ Coze Studio 后端采用 **DDD 分层架构**
2. ✅ 依赖方向: `API → Application → Domain → Infrastructure`
3. ✅ Domain 层是核心，包含业务逻辑
4. ✅ 使用接口进行依赖注入，便于测试和替换
5. ✅ 通过 Cross Domain 层进行跨域通信

### 学习建议
1. **先理解架构**，再看代码
2. **从简单到复杂**，逐步深入
3. **动手实践**，完成练习题
4. **阅读测试代码**，理解最佳实践
5. **保持耐心**，DDD 架构需要时间消化

---

祝你学习愉快！如果有任何问题，请随时在项目 Issue 中提问。

**Happy Coding! 🚀**

