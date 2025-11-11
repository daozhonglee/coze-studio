# CreateWorkflow API 工作流程文档

## 概述

本文档详细描述了 `/api/workflow_api/create` 接口的完整工作流程，从前端请求到数据库存储的完整链路。

## API 接口定义

### 接口信息
- **URL**: `/api/workflow_api/create`
- **方法**: POST
- **服务**: WorkflowService.CreateWorkflow
- **请求体**: CreateWorkflowRequest
- **响应体**: CreateWorkflowResponse

### 请求参数 (CreateWorkflowRequest)
```thrift
struct CreateWorkflowRequest {
    1: required string       name           // 工作流名称
    2: required string       desc           // 工作流描述
    3: required string       icon_uri       // 工作流图标URI
    4: required string       space_id       // 所属空间ID
    5: optional WorkflowMode flow_mode      // 工作流模式（Workflow/ChatFlow）
    6: optional SchemaType   schema_type    // 模式类型
    7: optional string       bind_biz_id    // 绑定业务ID
    8: optional i32          bind_biz_type  // 绑定业务类型
    9: optional string       project_id     // 项目ID（可选）
    10: optional bool        create_conversation // 是否创建会话（仅ChatFlow模式）
}
```

## 工作流程详解

### 1. 请求入口 (Application Layer)

**文件**: `backend/application/workflow/workflow.go`
**方法**: `ApplicationService.CreateWorkflow()`

#### 1.1 权限验证
```go
uID := ctxutil.MustGetUIDFromCtx(ctx)
spaceID := mustParseInt64(req.GetSpaceID())
if err := checkUserSpace(ctx, uID, spaceID); err != nil {
    return nil, err
}
```

**流程说明**:
- 从请求上下文中获取用户ID
- 解析并验证空间ID
- 调用 `checkUserSpace()` 检查用户是否有权限在该空间创建工作流
- `checkUserSpace()` 会调用跨用户服务获取用户的空间列表并验证权限

#### 1.2 会话模板创建 (仅ChatFlow模式)
```go
if req.ProjectID != nil && req.IsSetFlowMode() &&
   req.GetFlowMode() == workflow.WorkflowMode_ChatFlow &&
   req.IsSetCreateConversation() && req.GetCreateConversation() {
    // 创建会话模板
    _, err := GetWorkflowDomainSVC().CreateDraftConversationTemplate(ctx, &vo.CreateConversationTemplateMeta{
        AppID:   mustParseInt64(req.GetProjectID()),
        UserID:  uID,
        SpaceID: spaceID,
        Name:    req.Name,
    })
}
```

**流程说明**:
- 仅在ChatFlow模式且指定创建会话时执行
- 调用领域服务创建草稿会话模板
- 为聊天工作流预设会话配置

#### 1.3 工作流元数据构建
```go
wf := &vo.MetaCreate{
    CreatorID:        uID,
    SpaceID:          spaceID,
    ContentType:      workflow.WorkFlowType_User,
    Name:             req.Name,
    Desc:             req.Desc,
    IconURI:          req.IconURI,
    AppID:            parseInt64(req.ProjectID),
    Mode:             ternary.IFElse(req.IsSetFlowMode(), req.GetFlowMode(), workflow.WorkflowMode_Workflow),
    InitCanvasSchema: vo.GetDefaultInitCanvasJsonSchema(i18n.GetLocale(ctx)),
}
```

**流程说明**:
- 构建工作流元数据对象
- 设置创建者、空间、类型等基本信息
- 根据语言环境获取默认画布配置

#### 1.4 画布初始化
```go
if req.IsSetFlowMode() && req.GetFlowMode() == workflow.WorkflowMode_ChatFlow {
    conversationName := req.Name
    if !req.IsSetProjectID() || mustParseInt64(req.GetProjectID()) == 0 || !createConversation {
        conversationName = "Default"
    }
    wf.InitCanvasSchema = vo.GetDefaultInitCanvasJsonSchemaChat(i18n.GetLocale(ctx), conversationName)
}
```

**流程说明**:
- ChatFlow模式使用聊天画布模板
- 普通工作流使用标准画布模板
- 模板包含开始和结束节点的基础配置

### 2. 领域层处理 (Domain Layer)

**文件**: `backend/domain/workflow/service/service_impl.go`
**方法**: `impl.Create()`

#### 2.1 元数据创建
```go
id, err := i.repo.CreateMeta(ctx, &vo.Meta{
    CreatorID:   meta.CreatorID,
    SpaceID:     meta.SpaceID,
    ContentType: meta.ContentType,
    Name:        meta.Name,
    Desc:        meta.Desc,
    IconURI:     meta.IconURI,
    AppID:       meta.AppID,
    Mode:        meta.Mode,
})
```

**流程说明**:
- 调用Repository层创建工作流元数据
- 生成唯一的工作流ID
- 存储基本信息到数据库

#### 2.2 草稿版本创建
```go
if err = i.Save(ctx, id, meta.InitCanvasSchema); err != nil {
    return 0, err
}
```

**流程说明**:
- 调用 `Save()` 方法保存初始化画布
- 解析画布JSON，提取输入输出参数
- 生成新的提交ID用于版本控制
- 创建草稿版本记录

### 3. 数据访问层 (Repository Layer)

**涉及文件**:
- `backend/domain/workflow/repository.go`
- `backend/domain/workflow/entity/*.go`

#### 3.1 元数据存储
- 插入工作流基础信息到 `workflow_meta` 表
- 包含ID、名称、描述、创建者、空间等信息

#### 3.2 草稿数据存储
- 解析画布JSON结构
- 提取输入输出参数并序列化为JSON
- 计算测试运行状态
- 生成提交ID
- 存储到 `workflow_drafts` 表

### 4. 事件发布 (Event Publishing)

**文件**: `backend/application/workflow/eventbus.go`
**方法**: `PublishWorkflowResource()`

```go
err = PublishWorkflowResource(ctx, id, ptr.Of(int32(wf.Mode)), search.Created, &search.ResourceDocument{
    Name:          &wf.Name,
    APPID:         wf.AppID,
    SpaceID:       &wf.SpaceID,
    OwnerID:       &wf.CreatorID,
    PublishStatus: ptr.Of(resource.PublishStatus_UnPublished),
    CreateTimeMS:  ptr.Of(time.Now().UnixMilli()),
})
```

**流程说明**:
- 创建资源变更事件
- 设置工作流类型和子类型（模式）
- 包含工作流的元数据信息
- 发布到搜索索引系统用于全文检索

### 5. 响应返回

```go
return &workflow.CreateWorkflowResponse{
    Data: &workflow.CreateWorkflowData{
        WorkflowID: strconv.FormatInt(id, 10),
    },
}, nil
```

**流程说明**:
- 返回新创建的工作流ID
- ID转换为字符串格式返回给前端

## 数据流图

```
前端请求
    ↓
Application Layer (CreateWorkflow)
    ↓ 权限验证 → checkUserSpace()
    ↓ 会话创建 → CreateDraftConversationTemplate() [可选]
    ↓ 元数据构建 → vo.MetaCreate{}
    ↓ 画布初始化 → GetDefaultInitCanvasJsonSchema()
    ↓
Domain Layer (Create)
    ↓ 元数据创建 → repo.CreateMeta()
    ↓ 草稿保存 → Save() → repo.CreateOrUpdateDraft()
    ↓
Repository Layer
    ↓ 数据库写入：workflow_meta, workflow_drafts
    ↓
Event Publishing
    ↓ 发布事件 → search.ResourceEventBus.PublishResources()
    ↓
响应返回
    ↓ 返回工作流ID
```

## 关键组件说明

### 1. 权限验证
- **目的**: 确保用户只能在有权限的空间创建工作流
- **实现**: 调用跨用户服务验证用户-空间关系
- **失败处理**: 返回权限错误，阻止创建

### 2. 画布初始化
- **普通工作流**: 包含开始和结束节点的空画布
- **聊天工作流**: 包含会话管理相关的预设节点
- **本地化**: 根据用户语言环境选择合适的模板

### 3. 版本控制
- **草稿版本**: 每次保存都会创建新的提交ID
- **状态追踪**: 记录是否可以成功运行测试
- **历史回溯**: 支持查看和恢复历史版本

### 4. 事件驱动
- **搜索索引**: 新创建的工作流会自动加入搜索索引
- **缓存更新**: 相关缓存会通过事件机制更新
- **异步处理**: 资源变更事件异步处理，不阻塞主流程

## 错误处理

### 异常捕获
```go
defer func() {
    if panicErr := recover(); panicErr != nil {
        err = safego.NewPanicErr(panicErr, debug.Stack())
    }
    if err != nil {
        err = vo.WrapIfNeeded(errno.ErrWorkflowOperationFail, err, errorx.KV("cause", vo.UnwrapRootErr(err).Error()))
    }
}()
```

### 关键错误场景
1. **权限验证失败**: 用户无权访问指定空间
2. **会话模板创建失败**: ChatFlow模式下会话创建异常
3. **数据库操作失败**: 元数据或草稿保存失败
4. **事件发布失败**: 搜索索引更新失败（不影响主流程）

## 性能考虑

1. **异步事件**: 资源事件发布是异步的，不影响API响应时间
2. **批量操作**: 画布解析和参数提取一次性完成
3. **缓存预热**: 节点图标URL缓存会在服务启动时预加载
4. **连接池**: 数据库操作使用连接池，提高并发性能

## 扩展点

1. **自定义模板**: 可以扩展更多的工作流模板类型
2. **插件集成**: 支持通过插件扩展创建时的逻辑
3. **业务绑定**: 通过 `bind_biz_id` 和 `bind_biz_type` 支持业务关联
4. **多租户**: 通过空间ID实现多租户隔离
