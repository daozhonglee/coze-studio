# IDL 目录分析文档

> 📅 **生成日期**: 2025-10-27  
> 📊 **统计**: 49 个 Thrift 文件，约 11,580 行代码

## 📋 目录
- [概述](#概述)
- [目录结构](#目录结构)
- [核心文件](#核心文件)
- [服务分类](#服务分类)
- [IDL 组织模式](#idl-组织模式)
- [API 路由规范](#api-路由规范)
- [数据结构规范](#数据结构规范)
- [服务间依赖](#服务间依赖)
- [开发指南](#开发指南)
- [最佳实践](#最佳实践)

---

## 概述

本项目使用 **Apache Thrift IDL** 作为接口定义语言，采用契约优先（Contract-First）的开发模式。所有 API 接口、数据结构和服务定义都通过 Thrift 文件描述，然后自动生成 Go 代码和 API 路由。

### 🎯 IDL 的作用

1. **接口契约** - 定义前后端通信协议
2. **代码生成** - 自动生成 API Handler 和数据模型
3. **文档化** - 接口即文档，保持一致性
4. **类型安全** - 编译时检查，减少运行时错误
5. **跨语言支持** - Thrift 支持多种编程语言

### 📊 统计信息

| 项目 | 数量 |
|------|------|
| Thrift 文件总数 | 49 个 |
| 代码总行数 | ~11,580 行 |
| 顶级模块数量 | 12 个 |
| 定义的服务数量 | 15+ 个 |

---

## 目录结构

```
idl/
├── api.thrift                    # 🔥 主入口文件（聚合所有服务）
├── base.thrift                   # 🔥 基础类型定义
│
├── admin/                        # 管理后台
│   └── config.thrift            # 系统配置管理
│
├── app/                          # 应用/智能体模块
│   ├── bot_common.thrift        # Bot 通用结构
│   ├── bot_open_api.thrift      # Bot OpenAPI
│   ├── developer_api.thrift     # 开发者 API
│   ├── intelligence.thrift      # 🔥 智能体主服务
│   ├── project.thrift           # 项目管理
│   ├── publish.thrift           # 发布管理
│   ├── search.thrift            # 搜索服务
│   ├── task.thrift              # 任务管理
│   └── common_struct/           # 通用结构
│       ├── common_struct.thrift
│       ├── intelligence_common_struct.thrift
│       └── task_struct.thrift
│
├── conversation/                 # 对话模块
│   ├── agentrun_service.thrift  # Agent 运行服务
│   ├── common.thrift            # 对话通用定义
│   ├── conversation.thrift      # 对话实体
│   ├── conversation_service.thrift # 🔥 对话服务
│   ├── message.thrift           # 消息实体
│   ├── message_service.thrift   # 消息服务
│   └── run.thrift               # 运行记录
│
├── data/                         # 数据管理模块
│   ├── database/                # 数据库管理
│   │   ├── database_svc.thrift  # 数据库服务
│   │   └── table.thrift         # 表管理
│   ├── knowledge/               # 知识库管理
│   │   ├── common.thrift        # 知识库通用定义
│   │   ├── document.thrift      # 文档管理
│   │   ├── knowledge.thrift     # 知识库实体
│   │   ├── knowledge_svc.thrift # 🔥 知识库服务
│   │   ├── review.thrift        # 审核管理
│   │   └── slice.thrift         # 切片管理
│   └── variable/                # 变量/内存管理
│       ├── kvmemory.thrift      # KV 内存
│       ├── project_memory.thrift # 项目内存
│       └── variable_svc.thrift  # 变量服务
│
├── marketplace/                  # 市场模块
│   ├── marketplace_common.thrift
│   ├── product_common.thrift
│   └── public_api.thrift
│
├── passport/                     # 用户认证模块
│   └── passport.thrift          # 🔥 用户登录/注册服务
│
├── permission/                   # 权限管理模块
│   ├── openapiauth.thrift       # OpenAPI 认证实体
│   └── openapiauth_service.thrift # OpenAPI 认证服务
│
├── playground/                   # 游乐场模块
│   ├── playground.thrift        # 🔥 游乐场主服务
│   ├── prompt_resource.thrift   # Prompt 资源
│   └── shortcut_command.thrift  # 快捷命令
│
├── plugin/                       # 插件模块
│   ├── plugin_develop.thrift    # 🔥 插件开发服务
│   └── plugin_develop_common.thrift # 插件通用结构
│
├── resource/                     # 资源管理模块
│   ├── resource.thrift          # 资源服务
│   └── resource_common.thrift   # 资源通用定义
│
├── upload/                       # 文件上传模块
│   └── upload.thrift            # 上传服务
│
└── workflow/                     # 工作流模块
    ├── trace.thrift             # 追踪服务
    ├── workflow.thrift          # 工作流实体
    └── workflow_svc.thrift      # 🔥 工作流主服务
```

### 📁 模块说明

| 模块 | 文件数 | 主要功能 |
|------|--------|---------|
| `app/` | 8 | 智能体/应用管理、发布、搜索 |
| `conversation/` | 7 | 对话、消息、运行记录管理 |
| `data/` | 10 | 知识库、数据库、变量管理 |
| `workflow/` | 3 | 工作流创建、执行、追踪 |
| `plugin/` | 2 | 插件开发和管理 |
| `playground/` | 3 | 游乐场、Prompt、快捷命令 |
| `passport/` | 1 | 用户认证（登录/注册） |
| `permission/` | 2 | OpenAPI 权限管理 |
| `resource/` | 2 | 资源文件管理 |
| `upload/` | 1 | 文件上传服务 |
| `marketplace/` | 3 | 市场和产品管理 |
| `admin/` | 1 | 系统配置管理 |

---

## 核心文件

### 1. `api.thrift` - 主入口文件

**作用**: 聚合所有服务，作为代码生成的入口点。

```thrift
namespace go coze

// 包含所有子服务
include "./plugin/plugin_develop.thrift"
include "./marketplace/public_api.thrift"
include "./data/knowledge/knowledge_svc.thrift"
// ... 更多 include

// 定义服务（通过继承的方式聚合）
service IntelligenceService extends intelligence.IntelligenceService {}
service ConversationService extends conversation_service.ConversationService {}
service MessageService extends message_service.MessageService {}
service AgentRunService extends agentrun_service.AgentRunService {}
service OpenAPIAuthService extends openapiauth_service.OpenAPIAuthService {}
service MemoryService extends variable_svc.MemoryService {}
service PluginDevelopService extends plugin_develop.PluginDevelopService {}
service PublicProductService extends public_api.PublicProductService {}
service DeveloperApiService extends developer_api.DeveloperApiService {}
service PlaygroundService extends playground.PlaygroundService {}
service DatabaseService extends database_svc.DatabaseService {}
service ResourceService extends resource.ResourceService {}
service PassportService extends passport.PassportService {}
service WorkflowService extends workflow_svc.WorkflowService {}
service KnowledgeService extends knowledge_svc.DatasetService {}
service BotOpenApiService extends bot_open_api.BotOpenApiService {}
service UploadService extends upload.UploadService {}
service ConfigService extends config.ConfigService {}
```

**关键点**:
- ✅ 所有服务通过 `extends` 继承子服务
- ✅ 命名空间统一为 `go coze`
- ✅ 自动生成路由注册代码

---

### 2. `base.thrift` - 基础类型定义

**作用**: 定义所有 IDL 共享的基础结构。

```thrift
namespace go base

// RPC 基础请求
struct Base {
    1: string             LogID
    2: string             Caller
    3: string             Addr
    4: string             Client
    5: optional TrafficEnv         TrafficEnv
    6: optional map<string,string> Extra
}

// RPC 基础响应
struct BaseResp {
    1: string             StatusMessage = ""
    2: i32                StatusCode    = 0
    3: optional map<string,string> Extra
}

// 空请求
struct EmptyReq {}

// 空响应
struct EmptyResp {
    1: i64       code
    2: string    msg
    3: EmptyData data
}
```

**使用场景**:
- 所有请求可选包含 `Base` 字段（用于 RPC 调用）
- 所有响应可选包含 `BaseResp` 字段
- 统一的错误码和消息格式

---

## 服务分类

### 🤖 核心业务服务

| 服务名 | Thrift 文件 | 主要功能 |
|--------|------------|---------|
| IntelligenceService | `app/intelligence.thrift` | 智能体/项目的增删改查、发布管理 |
| ConversationService | `conversation/conversation_service.thrift` | 对话创建、清空、列表 |
| MessageService | `conversation/message_service.thrift` | 消息发送、查询、管理 |
| WorkflowService | `workflow/workflow_svc.thrift` | 工作流创建、保存、执行、追踪 |

### 📊 数据管理服务

| 服务名 | Thrift 文件 | 主要功能 |
|--------|------------|---------|
| KnowledgeService | `data/knowledge/knowledge_svc.thrift` | 知识库和文档管理 |
| DatabaseService | `data/database/database_svc.thrift` | 数据库表管理 |
| MemoryService | `data/variable/variable_svc.thrift` | KV 内存和变量管理 |

### 🔌 扩展服务

| 服务名 | Thrift 文件 | 主要功能 |
|--------|------------|---------|
| PluginDevelopService | `plugin/plugin_develop.thrift` | 插件开发和调试 |
| ResourceService | `resource/resource.thrift` | 资源文件管理 |
| UploadService | `upload/upload.thrift` | 文件上传 |

### 👤 用户相关服务

| 服务名 | Thrift 文件 | 主要功能 |
|--------|------------|---------|
| PassportService | `passport/passport.thrift` | 用户注册、登录、登出 |
| OpenAPIAuthService | `permission/openapiauth_service.thrift` | OpenAPI 认证和授权 |

### 🛒 市场服务

| 服务名 | Thrift 文件 | 主要功能 |
|--------|------------|---------|
| PublicProductService | `marketplace/public_api.thrift` | 市场产品展示 |

### ⚙️ 系统服务

| 服务名 | Thrift 文件 | 主要功能 |
|--------|------------|---------|
| ConfigService | `admin/config.thrift` | 系统配置管理 |
| PlaygroundService | `playground/playground.thrift` | 游乐场功能 |

---

## IDL 组织模式

### 📐 文件组织原则

1. **按领域模块分类**
   - 每个业务领域有独立的目录
   - 例如：`conversation/`、`workflow/`、`data/`

2. **服务与实体分离**
   - `*_service.thrift` - 服务定义（接口方法）
   - `*.thrift` - 数据结构定义（实体、请求、响应）
   - `*_common.thrift` - 通用定义和枚举

3. **分层结构**
   ```
   domain/
   ├── service.thrift      # 服务接口
   ├── entity.thrift       # 领域实体
   ├── common.thrift       # 通用定义
   └── subdomain/          # 子领域
   ```

### 📝 命名规范

#### 服务命名
```thrift
service <Domain>Service {
    // 方法命名: <Action><Resource><Version?>
    CreateWorkflowResponse CreateWorkflow(1: CreateWorkflowRequest request)
    GetWorkflowInfoResponse GetWorkflowInfo(1: GetWorkflowInfoRequest request)
    UpdateWorkflowResponse UpdateWorkflow(1: UpdateWorkflowRequest request)
    DeleteWorkflowResponse DeleteWorkflow(1: DeleteWorkflowRequest request)
}
```

#### 结构体命名
```thrift
// 请求: <Action><Resource>Request
struct CreateWorkflowRequest {
    // ...
}

// 响应: <Action><Resource>Response
struct CreateWorkflowResponse {
    1: required WorkflowData data
    253: required i32 code
    254: required string msg
}

// 实体: <Resource>Info / <Resource>Data
struct WorkflowInfo {
    1: required i64 workflow_id
    2: required string name
    // ...
}
```

---

## API 路由规范

### 🌐 路由定义格式

Thrift IDL 通过注解定义 HTTP 路由：

```thrift
service WorkflowService {
    // POST 请求
    CreateWorkflowResponse CreateWorkflow(1: CreateWorkflowRequest request) 
        (api.post='/api/workflow_api/create', 
         api.category="workflow_api", 
         api.gen_path="workflow_api", 
         agw.preserve_base="true")
    
    // GET 请求
    GetWorkflowInfoResponse GetWorkflowInfo(1: GetWorkflowInfoRequest request)
        (api.get='/api/workflow_api/:workflow_id',
         api.category="workflow_api")
    
    // PUT 请求
    UpdateWorkflowResponse UpdateWorkflow(1: UpdateWorkflowRequest request)
        (api.put='/api/workflow_api/update',
         api.category="workflow_api")
    
    // DELETE 请求
    DeleteWorkflowResponse DeleteWorkflow(1: DeleteWorkflowRequest request)
        (api.delete='/api/workflow_api/:workflow_id',
         api.category="workflow_api")
}
```

### 🏷️ 注解说明

| 注解 | 作用 | 示例 |
|------|------|------|
| `api.post` | 定义 POST 路由 | `api.post='/api/note/create'` |
| `api.get` | 定义 GET 路由 | `api.get='/api/note/:note_id'` |
| `api.put` | 定义 PUT 路由 | `api.put='/api/note/update'` |
| `api.delete` | 定义 DELETE 路由 | `api.delete='/api/note/:note_id'` |
| `api.category` | API 分类 | `api.category="note_api"` |
| `api.gen_path` | 代码生成路径 | `api.gen_path="note_api"` |
| `api.tag` | API 标签 | `api.tag="openapi"` |
| `agw.preserve_base` | 保留 Base 字段 | `agw.preserve_base="true"` |
| `api.js_conv` | JS 类型转换 | `api.js_conv="true"` (i64 -> string) |
| `agw.js_conv` | 网关 JS 转换 | `agw.js_conv="str"` |

### 📍 路由模式

#### 1. RESTful 风格
```thrift
// 创建资源
api.post='/api/resources'

// 获取资源详情
api.get='/api/resources/:resource_id'

// 更新资源
api.put='/api/resources/:resource_id'

// 删除资源
api.delete='/api/resources/:resource_id'

// 列表查询
api.get='/api/resources'
```

#### 2. RPC 风格
```thrift
// 所有操作都用 POST
api.post='/api/resource_api/create'
api.post='/api/resource_api/get_info'
api.post='/api/resource_api/update'
api.post='/api/resource_api/delete'
```

#### 3. 路径参数
```thrift
struct GetNoteRequest {
    1: required i64 note_id (api.path="note_id")
}

service NoteService {
    GetNoteResponse GetNote(1: GetNoteRequest req) 
        (api.get='/api/note/:note_id')
}
```

---

## 数据结构规范

### 📦 响应结构

所有响应必须包含 `code` 和 `msg` 字段：

```thrift
struct StandardResponse {
    1: required DataType data      // 业务数据
    253: required i32 code          // 状态码（0 表示成功）
    254: required string msg        // 状态消息
}
```

**示例**:
```thrift
struct CreateNoteResponse {
    1: required NoteInfo data
    253: required i32 code
    254: required string msg
}
```

### 🔢 字段编号规范

- `1-252`: 业务字段
- `253`: 状态码 (`code`)
- `254`: 状态消息 (`msg`)
- `255`: RPC 基础字段 (`Base` / `BaseResp`)

```thrift
struct ExampleRequest {
    1: required string field1
    2: optional i64 field2
    // ... 业务字段 ...
    255: optional base.Base Base (api.none="true")
}

struct ExampleResponse {
    1: required DataType data
    // ... 业务字段 ...
    253: required i32 code
    254: required string msg
    255: required base.BaseResp BaseResp (api.none="true")
}
```

### 🎯 类型转换注解

#### int64 转 string（避免 JS 精度丢失）

```thrift
struct NoteInfo {
    // Go 后端: int64
    // JS 前端: string
    1: required i64 note_id (agw.js_conv="str", api.js_conv="true")
    2: required i64 user_id (agw.js_conv="str", api.js_conv="true")
}
```

#### 路径参数绑定

```thrift
struct GetNoteRequest {
    1: required i64 note_id (
        agw.js_conv="str", 
        api.js_conv="true",
        api.path="note_id"  // 从路径参数绑定
    )
}
```

#### Query 参数绑定

```thrift
struct ListNotesRequest {
    1: optional i32 offset (api.query="offset")
    2: optional i32 limit (api.query="limit")
}
```

#### Header 参数绑定

```thrift
struct UploadFileRequest {
    1: required string ContentType (
        api.header="Content-Type",
        agw.source="header",
        agw.key="Content-Type"
    )
}
```

### 📋 常见数据类型

| Thrift 类型 | Go 类型 | 说明 |
|-------------|---------|------|
| `bool` | `bool` | 布尔值 |
| `i8` | `int8` | 8位整数 |
| `i16` | `int16` | 16位整数 |
| `i32` | `int32` | 32位整数 |
| `i64` | `int64` | 64位整数 |
| `double` | `float64` | 双精度浮点数 |
| `string` | `string` | 字符串 |
| `binary` | `[]byte` | 二进制数据 |
| `list<T>` | `[]T` | 列表 |
| `map<K,V>` | `map[K]V` | 映射 |
| `set<T>` | `[]T` | 集合 |

---

## 服务间依赖

### 📊 依赖关系图

```
api.thrift (根)
├── passport.thrift (用户认证)
│
├── app/intelligence.thrift (智能体)
│   ├── app/project.thrift
│   ├── app/publish.thrift
│   ├── app/search.thrift
│   └── app/task.thrift
│
├── conversation/conversation_service.thrift (对话)
│   ├── conversation/conversation.thrift
│   └── conversation/common.thrift
│
├── workflow/workflow_svc.thrift (工作流)
│   ├── workflow/workflow.thrift
│   └── workflow/trace.thrift
│
├── data/knowledge/knowledge_svc.thrift (知识库)
│   ├── data/knowledge/knowledge.thrift
│   ├── data/knowledge/document.thrift
│   └── data/knowledge/slice.thrift
│
├── plugin/plugin_develop.thrift (插件)
│   └── plugin/plugin_develop_common.thrift
│
└── ... (其他服务)
```

### 🔗 Include 依赖规范

1. **相对路径引用**
   ```thrift
   // 同级目录
   include "common.thrift"
   
   // 父级目录
   include "../base.thrift"
   
   // 子目录
   include "subdomain/entity.thrift"
   ```

2. **避免循环依赖**
   - ❌ 不要: A includes B, B includes A
   - ✅ 应该: 提取公共部分到 `common.thrift`

3. **依赖层次**
   ```
   base.thrift          (最底层，无依赖)
     ↓
   *_common.thrift      (领域通用定义)
     ↓
   entity.thrift        (实体定义)
     ↓
   service.thrift       (服务定义)
     ↓
   api.thrift           (顶层聚合)
   ```

---

## 开发指南

### 🚀 新增服务的完整流程

#### Step 1: 创建目录结构

```bash
mkdir -p idl/note
cd idl/note
```

#### Step 2: 定义实体和请求响应

创建 `note.thrift`:

```thrift
namespace go note

// 笔记信息
struct NoteInfo {
    1: required i64 note_id (agw.js_conv="str", api.js_conv="true")
    2: required i64 user_id (agw.js_conv="str", api.js_conv="true")
    3: required string title
    4: required string content
    5: required i64 created_at
    6: required i64 updated_at
}

// 创建笔记请求
struct CreateNoteRequest {
    1: required string title
    2: required string content
}

// 创建笔记响应
struct CreateNoteResponse {
    1: required NoteInfo data
    253: required i32 code
    254: required string msg
}

// ... 其他请求响应定义
```

#### Step 3: 定义服务接口

创建 `note_service.thrift`:

```thrift
include "./note.thrift"

namespace go note

service NoteService {
    // 创建笔记
    note.CreateNoteResponse CreateNote(1: note.CreateNoteRequest req) 
        (api.post="/api/note/create", api.category="note")
    
    // 获取笔记详情
    note.GetNoteResponse GetNote(1: note.GetNoteRequest req)
        (api.get="/api/note/:note_id", api.category="note")
    
    // 更新笔记
    note.UpdateNoteResponse UpdateNote(1: note.UpdateNoteRequest req)
        (api.post="/api/note/update", api.category="note")
    
    // 删除笔记
    note.DeleteNoteResponse DeleteNote(1: note.DeleteNoteRequest req)
        (api.delete="/api/note/:note_id", api.category="note")
    
    // 列表查询
    note.ListNotesResponse ListNotes(1: note.ListNotesRequest req)
        (api.get="/api/note/list", api.category="note")
}
```

#### Step 4: 在 api.thrift 中注册

编辑 `api.thrift`:

```thrift
// 添加 include
include "./note/note_service.thrift"

// ... 其他 include ...

// 添加服务定义
service NoteService extends note_service.NoteService {}

// ... 其他服务 ...
```

#### Step 5: 生成代码

```bash
# 进入项目根目录

# 生成 API 代码
make gen_api
```

生成的文件：
- `backend/api/model/note/*.go` - 数据模型
- `backend/api/handler/coze/note_service.go` - API Handler

#### Step 6: 实现 Handler

编辑生成的 Handler 文件，调用应用服务：

```go
package coze

import (
    "context"
    "net/http"

    "github.com/cloudwego/hertz/pkg/app"
    
    "github.com/coze-dev/coze-studio/backend/api/model/note"
    noteApp "github.com/coze-dev/coze-studio/backend/application/note"
)

// CreateNote .
// @router /note/create [POST]
func CreateNote(ctx context.Context, c *app.RequestContext) {
    var req note.CreateNoteRequest
    err := c.BindAndValidate(&req)
    if err != nil {
        invalidParamRequestResponse(c, err.Error())
        return
    }

    resp, err := noteApp.NoteApplicationSVC.CreateNote(ctx, &req)
    if err != nil {
        internalServerErrorResponse(ctx, c, err)
        return
    }

    c.JSON(http.StatusOK, resp)
}
```

---

## 最佳实践

### ✅ DO - 应该做的

1. **命名清晰规范**
   ```thrift
   ✅ CreateNoteRequest
   ✅ GetNoteDetailResponse
   ✅ UpdateNoteRequest
   
   ❌ NoteReq
   ❌ Resp1
   ❌ UpdateReq
   ```

2. **字段注释完整**
   ```thrift
   ✅ struct NoteInfo {
       1: required i64 note_id  // 笔记 ID
       2: required string title // 笔记标题
   }
   
   ❌ struct NoteInfo {
       1: required i64 note_id
       2: required string title
   }
   ```

3. **使用语义化的类型**
   ```thrift
   ✅ i64 created_at  // 时间戳（毫秒）
   ✅ i32 status      // 状态：1-正常 2-删除
   
   ❌ string created_at  // 字符串表示时间
   ❌ i64 flag           // 含义不明确
   ```

4. **统一响应格式**
   ```thrift
   ✅ struct Response {
       1: required DataType data
       253: required i32 code
       254: required string msg
   }
   
   ❌ struct Response {
       1: optional DataType result
       2: i32 status
   }
   ```

5. **int64 字段加转换注解**
   ```thrift
   ✅ 1: required i64 id (agw.js_conv="str", api.js_conv="true")
   
   ❌ 1: required i64 id
   ```

6. **服务方法归类**
   ```thrift
   ✅ service NoteService {
       // CRUD 操作
       CreateNoteResponse CreateNote(...)
       GetNoteResponse GetNote(...)
       UpdateNoteResponse UpdateNote(...)
       DeleteNoteResponse DeleteNote(...)
       
       // 列表查询
       ListNotesResponse ListNotes(...)
   }
   ```

### ❌ DON'T - 不应该做的

1. **不要定义冗余结构**
   ```thrift
   ❌ // 重复定义
   struct CreateNoteData {
       1: string title
       2: string content
   }
   struct UpdateNoteData {
       1: string title
       2: string content
   }
   
   ✅ // 复用结构
   struct NoteInput {
       1: string title
       2: string content
   }
   ```

2. **不要使用魔法数字**
   ```thrift
   ❌ 1: i32 status  // 1=正常 2=删除 ???
   
   ✅ enum NoteStatus {
       NORMAL = 1,
       DELETED = 2
   }
   1: NoteStatus status
   ```

3. **不要省略 required/optional**
   ```thrift
   ❌ struct Request {
       1: string title  // 不明确
   }
   
   ✅ struct Request {
       1: required string title  // 明确必填
       2: optional string desc   // 明确可选
   }
   ```

4. **不要创建循环依赖**
   ```thrift
   ❌ // a.thrift
   include "b.thrift"
   
   // b.thrift
   include "a.thrift"
   ```

5. **不要直接修改生成的代码**
   ```go
   ❌ // backend/api/model/note/note_gen.go
   // 手动修改生成的文件（会被覆盖）
   
   ✅ // 修改 IDL，重新生成
   ```

### 📝 注释规范

```thrift
// ========================================
// 笔记管理服务
// ========================================

namespace go note

/**
 * 笔记信息
 * 用于存储用户创建的笔记
 */
struct NoteInfo {
    1: required i64 note_id      // 笔记唯一标识
    2: required i64 user_id      // 所属用户 ID
    3: required i64 space_id     // 所属空间 ID
    4: required string title     // 笔记标题（最大 255 字符）
    5: required string content   // 笔记内容（支持 Markdown）
    6: required i32 status       // 状态：1-正常 2-已删除
    7: required i64 created_at   // 创建时间（毫秒时间戳）
    8: required i64 updated_at   // 更新时间（毫秒时间戳）
}
```

### 🔧 调试技巧

1. **检查 IDL 语法**
   ```bash
   # 使用 thrift 编译器检查语法
   thrift --gen go idl/note/note.thrift
   ```

2. **查看生成的路由**
   ```bash
   # 查看路由注册代码
   cat backend/api/router/register.go
   ```

3. **验证生成的模型**
   ```bash
   # 查看生成的数据模型
   ls -la backend/api/model/note/
   ```

4. **测试 API**
   ```bash
   # 使用 curl 测试
   curl -X POST http://localhost:8080/api/note/create \
     -H "Content-Type: application/json" \
     -d '{"title":"测试","content":"内容"}'
   ```

---

## 📚 参考资源

### 官方文档
- [Apache Thrift 官方文档](https://thrift.apache.org/docs/)
- [Thrift IDL 语法参考](https://thrift.apache.org/docs/idl)
- [Hertz 框架文档](https://www.cloudwego.io/docs/hertz/)

### 项目相关文档
- `BACKEND_ERRATA.md` - 后端勘误表
- `BACKEND_PRACTICE.md` - 后端实战练习
- `BACKEND_QUICKSTART.md` - 快速入门

### 工具推荐
- [Thrift Compiler](https://thrift.apache.org/download) - Thrift 编译器
- [VS Code Thrift Extension](https://marketplace.visualstudio.com/items?itemName=faustinoaq.thrift-language) - VS Code 插件
- [Postman](https://www.postman.com/) - API 测试工具

---

## 🎯 总结

### IDL 的核心价值

1. **契约优先** - API 定义先于实现，确保前后端一致
2. **自动生成** - 减少手写代码，提高开发效率
3. **类型安全** - 编译时检查，减少运行时错误
4. **文档化** - 接口即文档，保持最新
5. **跨语言** - 支持多种编程语言

### 开发流程

```
定义 IDL → 生成代码 → 实现 Handler → 测试 API
   ↓          ↓           ↓            ↓
 .thrift    make gen   handler.go   curl/postman
```

### 关键要点

- ✅ 所有 API 必须通过 IDL 定义
- ✅ 遵循命名和结构规范
- ✅ int64 字段必须添加 JS 转换注解
- ✅ 响应格式必须统一（code + msg）
- ✅ 不要手动修改生成的代码
- ✅ 保持 IDL 文件的清晰和可维护性

---

<div align="center">
  <strong>📖 完整的 IDL 分析文档 📖</strong><br>
  <em>Contract-First Development with Apache Thrift</em>
</div>

