# Coze Studio 后端快速入门

> 通过实际代码示例，快速理解项目架构

## ⚠️ 重要说明

**本文档目的**: 帮助你理解 DDD 分层架构的核心思想

**请注意**:
- ✅ 架构分层和依赖关系是准确的
- ⚠️ 部分代码示例为**简化版**，便于理解
- 📖 学习完架构后，请参考 **实际代码** 了解具体实现
- 📄 查看 [`BACKEND_ERRATA.md`](./BACKEND_ERRATA.md) 了解文档与实际代码的差异
- 📄 查看 [`BACKEND_GORM_GEN_GUIDE.md`](./BACKEND_GORM_GEN_GUIDE.md) 了解 GORM Gen 的实际使用

## 📝 从一个 User 请求说起

让我们通过一个真实的 **"获取用户信息"** 请求，了解代码是如何在各层之间流转的。

### 完整请求流程图

```
用户浏览器
    │ HTTP GET /api/user/profile?user_id=123
    ▼
┌─────────────────────────────────────────────┐
│  1. API Layer (api/)                        │
│  ▸ Middleware: 认证、日志、权限检查         │
│  ▸ Handler: 解析请求参数                    │
└─────────────────┬───────────────────────────┘
                  │ GetUserInfo(ctx, userID)
                  ▼
┌─────────────────────────────────────────────┐
│  2. Application Layer (application/)        │
│  ▸ 协调领域服务                             │
│  ▸ 事务管理                                 │
└─────────────────┬───────────────────────────┘
                  │ GetUserInfo(ctx, userID)
                  ▼
┌─────────────────────────────────────────────┐
│  3. Domain Layer (domain/)                  │
│  ▸ Entity: User 实体                        │
│  ▸ Service: 业务逻辑                        │
│  ▸ Repository: 数据访问接口                 │
└─────────────────┬───────────────────────────┘
                  │ FindByID(ctx, userID)
                  ▼
┌─────────────────────────────────────────────┐
│  4. Infrastructure Layer (infra/)           │
│  ▸ GORM: 数据库查询                         │
│  ▸ Redis: 缓存                              │
└─────────────────┬───────────────────────────┘
                  │ MySQL Query
                  ▼
                MySQL
```

---

## 🔍 代码分层详解

### Layer 1: Entity (领域实体)

**位置**: `domain/user/entity/user.go`

```go
// User 实体定义了用户的核心属性
type User struct {
    UserID       int64  // 用户 ID
    Name         string // 昵称
    UniqueName   string // 唯一用户名
    Email        string // 邮箱
    Description  string // 描述
    IconURI      string // 头像 URI
    IconURL      string // 头像 URL
    UserVerified bool   // 是否验证
    Locale       string // 语言设置
    SessionKey   string // 会话密钥
    CreatedAt    int64  // 创建时间
    UpdatedAt    int64  // 更新时间
}

// 💡 关键点:
// 1. Entity 是纯数据对象，包含核心领域概念
// 2. 不包含任何基础设施相关的代码 (如数据库标签)
// 3. 可以包含简单的业务逻辑方法
```

**理解要点**:
- ✅ Entity 是领域模型的核心
- ✅ 反映业务概念，而非数据库表结构
- ✅ 可以包含业务规则验证方法

### Layer 2: Service Interface (领域服务接口)

**位置**: `domain/user/service/user.go`

```go
// User 接口定义了用户领域的所有业务操作
type User interface {
    // 创建用户
    Create(ctx context.Context, req *CreateUserRequest) (user *entity.User, err error)
    
    // 登录
    Login(ctx context.Context, email, password string) (user *entity.User, err error)
    
    // 登出
    Logout(ctx context.Context, userID int64) (err error)
    
    // 重置密码
    ResetPassword(ctx context.Context, email, password string) (err error)
    
    // 获取用户信息 ⭐️ 我们关注的方法
    GetUserInfo(ctx context.Context, userID int64) (user *entity.User, err error)
    
    // 更新头像
    UpdateAvatar(ctx context.Context, userID int64, ext string, imagePayload []byte) (url string, err error)
    
    // 更新资料
    UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (err error)
    
    // 验证会话
    ValidateSession(ctx context.Context, sessionKey string) (session *entity.Session, exist bool, err error)
    
    // 获取用户空间列表
    GetUserSpaceList(ctx context.Context, userID int64) (spaces []*entity.Space, err error)
    
    // ... 其他方法
}

// 💡 关键点:
// 1. 定义接口而非实现，遵循依赖倒置原则
// 2. 方法签名清晰，参数和返回值明确
// 3. 所有方法都接收 context.Context，支持取消和超时
```

**理解要点**:
- ✅ 接口定义了领域的能力边界
- ✅ 便于测试 (可以 mock)
- ✅ 便于替换实现

### Layer 3: Service Implementation (领域服务实现)

**位置**: `domain/user/service/user_impl.go`

```go
// ServiceImpl 实现了 User 接口
type ServiceImpl struct {
    userRepo  repository.UserRepository   // 用户仓储
    spaceRepo repository.SpaceRepository  // 空间仓储
    iconOSS   storage.Storage             // 对象存储
    idgen     idgen.IDGenerator           // ID 生成器
}

// GetUserInfo 获取用户信息的具体实现
func (s *ServiceImpl) GetUserInfo(ctx context.Context, userID int64) (*entity.User, error) {
    // 1. 参数验证
    if userID <= 0 {
        return nil, errors.New("invalid user id")
    }
    
    // 2. 从仓储获取用户
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // 3. 业务逻辑处理 (如果需要)
    // 例如: 加载用户头像 URL
    if user.IconURI != "" {
        user.IconURL = s.iconOSS.GetURL(user.IconURI)
    }
    
    // 4. 返回结果
    return user, nil
}

// 💡 关键点:
// 1. 通过依赖注入获取 Repository、Storage 等依赖
// 2. 包含核心业务逻辑
// 3. 不关心数据从哪里来 (Database? Cache? API?)
```

**理解要点**:
- ✅ 实现业务逻辑，不关心技术细节
- ✅ 依赖接口而非具体实现
- ✅ 纯粹的业务代码，易于测试

### Layer 4: Repository Interface (仓储接口)

**位置**: `domain/user/repository/repository.go`

```go
// UserRepository 定义了用户数据访问接口
type UserRepository interface {
    // 通过 ID 获取用户
    GetByID(ctx context.Context, id int64) (*entity.User, error)
    
    // 通过邮箱获取用户
    GetByEmail(ctx context.Context, email string) (*entity.User, error)
    
    // 保存用户
    Save(ctx context.Context, user *entity.User) error
    
    // 更新用户
    Update(ctx context.Context, user *entity.User) error
    
    // 删除用户
    Delete(ctx context.Context, id int64) error
    
    // 批量获取用户
    MGetByIDs(ctx context.Context, ids []int64) ([]*entity.User, error)
    
    // ... 其他数据访问方法
}

// 💡 关键点:
// 1. Repository 是 Domain 层定义的，但在 Infrastructure 层实现
// 2. 接口返回 Domain Entity，而非数据库模型
// 3. 抽象了数据访问，Domain 层不关心数据来源
```

**理解要点**:
- ✅ 数据访问的抽象
- ✅ Domain 层定义接口，Infrastructure 层实现
- ✅ 可以有多种实现 (MySQL、MongoDB、内存等)

### Layer 5: Repository Implementation (仓储实现)

**位置**: `domain/user/internal/dal/user.go`

```go
// UserDAO 实现了 UserRepository 接口
// 💡 注意：这个项目使用 GORM Gen 自动生成 DAO 代码
type UserDAO struct {
    query *query.Query  // GORM Gen 生成的查询对象
}

// NewUserDAO 创建用户 DAO 实例
func NewUserDAO(db *gorm.DB) *UserDAO {
    return &UserDAO{
        query: query.Use(db),  // 使用 GORM Gen
    }
}

// GetUserByID 通过 ID 获取用户
func (dao *UserDAO) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
    // 使用 GORM Gen 生成的类型安全查询
    return dao.query.User.WithContext(ctx).
        Where(dao.query.User.ID.Eq(userID)).
        First()
}

// GetUsersByEmail 通过邮箱获取用户
func (dao *UserDAO) GetUsersByEmail(ctx context.Context, email string) (*model.User, bool, error) {
    user, err := dao.query.User.WithContext(ctx).
        Where(dao.query.User.Email.Eq(email)).
        First()
    
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, false, nil
    }
    
    if err != nil {
        return nil, false, err
    }
    
    return user, true, nil
}

// CreateUser 创建新用户
func (dao *UserDAO) CreateUser(ctx context.Context, user *model.User) error {
    return dao.query.User.WithContext(ctx).Create(user)
}

// UpdateProfile 更新用户资料
func (dao *UserDAO) UpdateProfile(ctx context.Context, userID int64, updates map[string]interface{}) error {
    if _, ok := updates["updated_at"]; !ok {
        updates["updated_at"] = time.Now().UnixMilli()
    }
    
    _, err := dao.query.User.WithContext(ctx).
        Where(dao.query.User.ID.Eq(userID)).
        Updates(updates)
    return err
}

// 💡 关键点:
// 1. 使用 GORM Gen 自动生成的类型安全查询
// 2. model.User 是 GORM Gen 生成的模型 (在 internal/dal/model/ 目录)
// 3. 不需要手写 DO → Entity 转换，直接使用生成的模型
// 4. 类型安全：dao.query.User.ID.Eq(userID) 有编译时检查
```

**GORM Gen 的优势**:
- ✅ 自动生成类型安全的查询代码
- ✅ 避免字符串拼接 SQL
- ✅ 编译时发现错误
- ✅ 减少手写重复代码

**模型文件位置**:
```
domain/user/internal/dal/
├── model/              # GORM Gen 生成的模型
│   ├── user.gen.go    # User 模型
│   └── space.gen.go   # Space 模型
├── query/              # GORM Gen 生成的查询代码
│   ├── user.gen.go    # User 查询
│   └── gen.go         # 查询入口
└── user.go             # 手写的 DAO 实现
```

**理解要点**:
- ✅ 实现具体的数据访问逻辑
- ✅ DO (Data Object) 与数据库表对应
- ✅ Entity 与业务概念对应
- ✅ 通过转换隔离技术细节

### Layer 6: Application Service (应用服务)

**位置**: `application/user/user.go`

> ⚠️ **注意**: 以下为**简化的概念示例**，实际项目使用全局变量单例模式 + IDL 生成的类型

```go
// UserApplicationService 应用服务
// 实际：var UserApplicationSVC = &UserApplicationService{} (全局变量)
type UserApplicationService struct {
    DomainSVC service.User    // 领域服务
    oss       storage.Storage // 对象存储
}

// GetUserInfo 获取用户信息的应用层实现（简化示例）
// 实际方法：PassportAccountInfoV2，使用 IDL 生成的类型
func (s *UserApplicationService) GetUserInfo(ctx context.Context, userID int64) (*GetUserInfoResponse, error) {
    // 1. 调用领域服务
    user, err := s.DomainSVC.GetUserInfo(ctx, userID)
    if err != nil {
        logs.Errorf("[UserApplicationService] GetUserInfo failed, userID=%d, err=%v", userID, err)
        return nil, err
    }
    
    // 2. 转换为应用层响应对象
    resp := &GetUserInfoResponse{
        UserID:       user.UserID,
        Name:         user.Name,
        UniqueName:   user.UniqueName,
        Email:        user.Email,
        Description:  user.Description,
        IconURL:      user.IconURL,
        UserVerified: user.UserVerified,
        Locale:       user.Locale,
    }
    
    // 3. 可以在这里添加额外的协调逻辑
    // 例如: 记录访问日志、发送事件、调用其他服务等
    
    return resp, nil
}

// GetUserInfoResponse 应用层响应对象
type GetUserInfoResponse struct {
    UserID       int64  `json:"user_id"`
    Name         string `json:"name"`
    UniqueName   string `json:"unique_name"`
    Email        string `json:"email"`
    Description  string `json:"description"`
    IconURL      string `json:"icon_url"`
    UserVerified bool   `json:"user_verified"`
    Locale       string `json:"locale"`
}

// 💡 关键点:
// 1. Application 层协调 Domain 服务
// 2. 可以涉及多个 Domain 服务的编排
// 3. 管理事务边界
// 4. 转换为适合 API 的响应格式
```

**理解要点**:
- ✅ 协调多个领域服务
- ✅ 处理用例逻辑
- ✅ 管理事务
- ✅ 转换数据格式

### Layer 7: Initialization (依赖注入)

**位置**: `application/user/init.go`

> ⚠️ **重要**: 实际项目使用**全局变量单例模式**，不是每次创建新实例！

```go
// 实际代码：全局变量
var UserApplicationSVC = &UserApplicationService{}

// InitService 初始化用户应用服务（实际代码）
func InitService(
    ctx context.Context,
    db *gorm.DB,
    oss storage.Storage,
    idgen idgen.IDGenerator,
) *UserApplicationService {
    // 直接修改全局变量的字段，而不是创建新实例
    UserApplicationSVC.DomainSVC = service.NewUserDomain(ctx, &service.Components{
        IconOSS:   oss,
        IDGen:     idgen,
        UserRepo:  repository.NewUserRepo(db),
        SpaceRepo: repository.NewSpaceRepo(db),
    })
    UserApplicationSVC.oss = oss
    
    return UserApplicationSVC  // 返回全局变量
}

// 💡 关键点:
// 1. 依赖注入：从外部传入依赖
// 2. 自底向上构建：Repository → Domain Service → Application Service
// 3. 所有依赖都是接口类型
```

**理解要点**:
- ✅ 依赖注入解耦
- ✅ 使用**全局变量单例**模式（项目特点）
- ✅ 初始化时修改全局变量字段
- ⚠️ 与标准构造函数模式不同，注意区别

### Layer 8: API Handler（⚠️ IDL 自动生成）

**⚠️ 重要**: Handler 代码是由 Thrift IDL 自动生成的，不是手写的！

**实际位置**: `api/handler/coze/passport_service.go`（由 `idl/passport/passport.thrift` 生成）

```go
// ⚠️ 以下代码由 IDL 自动生成（不要手动修改）

package coze

import (
    "github.com/coze-dev/coze-studio/backend/api/model/passport"
    "github.com/coze-dev/coze-studio/backend/application/user"
)

// PassportAccountInfoV2 获取用户信息
// @router /api/passport/account/info/v2/ [POST]
func PassportAccountInfoV2(ctx context.Context, c *app.RequestContext) {
    var req passport.PassportAccountInfoV2Request
    
    // 1. 绑定和验证请求参数（自动）
    err := c.BindAndValidate(&req)
    if err != nil {
        invalidParamRequestResponse(c, err.Error())
        return
    }

    // 2. 调用应用服务（⚠️ 使用全局变量）
    resp, err := user.UserApplicationSVC.PassportAccountInfoV2(ctx, &req)
    if err != nil {
        internalServerErrorResponse(ctx, c, err)
        return
    }

    // 3. 返回 JSON 响应
    c.JSON(http.StatusOK, resp)
}

// 💡 关键点:
// 1. ⚠️ Handler 由 IDL 生成，不要手动编写或修改
// 2. ⚠️ 直接使用全局变量 user.UserApplicationSVC
// 3. ✅ BindAndValidate 自动验证请求参数
// 4. ✅ 使用辅助函数处理错误（invalidParamRequestResponse, internalServerErrorResponse）
```

**理解要点**:
- ⚠️ **代码自动生成**：由 `idl/passport/passport.thrift` 自动生成
- ⚠️ **无 Handler 结构体**：没有 `UserHandler` 这样的结构体
- ⚠️ **全局变量调用**：直接使用 `user.UserApplicationSVC`
- ✅ 负责 HTTP 请求解析和响应封装
- ✅ 不包含业务逻辑

### Layer 9: Router (路由注册 - ⚠️ IDL 自动生成)

**⚠️ 重要**: 路由注册代码也是由 IDL 自动生成的！

**实际位置**: `api/router/coze/api.go`（由 IDL 自动生成）

```go
// ⚠️ 以下代码由 IDL 自动生成（不要手动修改）

package coze

import "github.com/cloudwego/hertz/pkg/app/server"

// Register 注册所有路由
func Register(r *server.Hertz) {
    // PassportService 路由（来自 idl/passport/passport.thrift）
    r.POST("/api/passport/web/email/register/v2/", PassportWebEmailRegisterV2Post)
    r.GET("/api/passport/web/logout/", PassportWebLogoutGet)
    r.POST("/api/passport/web/email/login/", PassportWebEmailLoginPost)
    r.POST("/api/passport/account/info/v2/", PassportAccountInfoV2)
    r.POST("/api/web/user/update/upload_avatar/", UserUpdateAvatar)
    r.POST("/api/user/update_profile", UserUpdateProfile)
    
    // ... 其他服务的路由
}

// 💡 关键点:
// 1. ⚠️ 路由由 IDL 自动生成，不要手动注册
// 2. ✅ 直接绑定到生成的 Handler 函数
// 3. ✅ 路由路径来自 IDL 注解（如 api.post="/api/passport/..."）
```

**主入口调用**:

```go
// api/router/register.go
func GeneratedRegister(r *server.Hertz) {
    // INSERT_POINT: DO NOT DELETE THIS LINE!
    coze.Register(r)  // ⚠️ 调用自动生成的路由注册
    staticFileRegister(r)
}
```

---

## 🎯 关键概念总结

### 1. 分层依赖方向

```
API Layer (api/)
    ↓ depends on
Application Layer (application/)
    ↓ depends on
Domain Layer (domain/)
    ↑ implements
Infrastructure Layer (infra/)
```

### 2. 数据对象转换

```
HTTP Request (JSON)
    ↓ bind
API Model
    ↓ convert
Application DTO
    ↓ convert
Domain Entity
    ↓ convert
Data Object (DO)
    ↓ save
Database Table
```

### 3. 关键模式

#### Repository Pattern (仓储模式)
```
Domain defines Interface → Infrastructure implements
```

#### Dependency Injection (依赖注入)
```
Depends on Interface, not Implementation
```

#### DTO Pattern (数据传输对象)
```
Layer boundaries use DTOs for data transfer
```

---

## 🚀 动手实践

### 任务 1: 追踪一个请求

1. **启动项目**
```bash
make server
```

2. **发送请求**
```bash
# 用户登录请求（实际存在的 API）
curl -X POST http://localhost:8080/api/passport/web/email/login/ \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

3. **在代码中添加日志**，追踪请求流程:

```go
// ⚠️ 注意：使用实际的文件路径

// api/handler/coze/passport_service.go（IDL 生成，不建议修改）
func PassportWebEmailLoginPost(ctx context.Context, c *app.RequestContext) {
    logs.Infof("👉 [API Layer] PassportWebEmailLoginPost called")
    // ... 实际代码
}

// application/user/user.go
func (u *UserApplicationService) PassportWebEmailLoginPost(ctx context.Context, req *passport.PassportWebEmailLoginPostRequest) (...) {
    logs.Infof("👉 [Application Layer] Login called, email=%s", req.GetEmail())
    userInfo, err := u.DomainSVC.Login(ctx, req.GetEmail(), req.GetPassword())
    // ...
}

// domain/user/service/user_impl.go
func (u *userImpl) Login(ctx context.Context, email, password string) (*userEntity.User, error) {
    logs.Infof("👉 [Domain Layer] Login called, email=%s", email)
    userModel, exist, err := u.UserRepo.GetUsersByEmail(ctx, email)
    // ...
}

// domain/user/internal/dal/user.go
func (dao *UserDAO) GetUsersByEmail(ctx context.Context, email string) (*model.User, bool, error) {
    logs.Infof("👉 [Infrastructure Layer] GetUsersByEmail called, email=%s", email)
    user, err := dao.query.User.WithContext(ctx).Where(dao.query.User.Email.Eq(email)).First()
    // ...
}
```

4. **观察日志输出**，理解请求在各层的流转

### 任务 2: 修改代码，添加缓存

在 Repository 实现中添加 Redis 缓存:

```go
// ⚠️ 实际文件: domain/user/internal/dal/user.go

type UserDAO struct {
    query *query.Query
    cache *redis.Client  // 新增缓存（需要在初始化时注入）
}

func (dao *UserDAO) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
    // 1. 尝试从缓存读取
    cacheKey := fmt.Sprintf("user:%d", userID)
    cachedData, err := dao.cache.Get(ctx, cacheKey).Result()
    if err == nil {
        // 缓存命中
        var user model.User
        json.Unmarshal([]byte(cachedData), &user)
        logs.Infof("✅ Cache hit for user %d", userID)
        return &user, nil
    }
    
    // 2. 缓存未命中，从数据库读取
    var userDO UserDO
    err = r.db.WithContext(ctx).Where("user_id = ?", id).First(&userDO).Error
    if err != nil {
        return nil, err
    }
    
    // 3. 转换为 Entity
    user := convertDOToEntity(&userDO)
    
    // 4. 写入缓存
    data, _ := json.Marshal(user)
    r.cache.Set(ctx, cacheKey, data, 5*time.Minute)
    
    return user, nil
}
```

**思考**: 
- ✅ Domain **接口**不需要修改（`repository.UserRepository` 接口保持不变）
- ✅ Domain **Service** 层不需要修改（业务逻辑不变）
- ✅ Application 层不需要修改
- ✅ API 层不需要修改
- ✅ 只修改 Repository **实现**（`UserDAO` 的内部实现）

💡 **关键理解 - 为什么说 "Domain 层不需要修改"？**

这里需要区分**物理位置**和**逻辑分层**：

**物理位置**（文件在哪里）：
- `UserDAO` 确实在 `domain/user/internal/dal/` 目录下

**逻辑分层**（代码的职责）：
- `UserDAO` 是**数据访问实现**，属于 Infrastructure 层的职责
- 虽然文件在 `domain/` 目录，但 `internal/dal/` 表示这是内部实现细节

---

### 🤔 为什么 UserDAO 属于 Infrastructure 层？

这是 DDD 中的核心概念，让我详细解释：

**🎯 DDD 的分层原则**：

```
┌─────────────────────────────────────────┐
│ Domain 层：业务逻辑                      │
│ - 关注"做什么"（What）                   │
│ - 业务规则、业务概念                    │
│ - 不关心技术细节                         │
└─────────────────────────────────────────┘
              ↓ 定义需求
┌─────────────────────────────────────────┐
│ Infrastructure 层：技术实现              │
│ - 关注"怎么做"（How）                    │
│ - 数据库、网络、文件系统                │
│ - 具体的技术选型和实现                  │
└─────────────────────────────────────────┘
```

**💡 关键区别**：

| 层次 | 关注点 | 示例 |
|------|--------|------|
| **Domain 层** | 业务概念 | "我需要存取用户信息" |
| **Infrastructure 层** | 技术手段 | "我用 MySQL/Redis/文件来存取" |

**🏪 举个生活例子**：

```
你去餐厅吃饭：

Domain 层（业务需求）：
├── 你说："我要一份宫保鸡丁"  ← 你关心的是"吃什么"
└── 接口定义：Restaurant.orderFood(dishName)

Infrastructure 层（技术实现）：
├── 厨房怎么做？用炒锅还是烤箱？  ← 技术细节
├── 食材从哪采购？              ← 你不关心
└── 具体实现：Chef.cook(ingredients, method)

你（调用方）只需要知道接口，不需要知道厨房怎么运作！
```

**📝 代码层面的详细说明**：

```go
// ==========================================
// Domain 层：定义业务需要什么
// ==========================================
// domain/user/repository/repository.go

type UserRepository interface {
    GetUserByID(ctx context.Context, userID int64) (*model.User, error)
}

💡 这是 Domain 层的一部分，因为：
- 它表达了业务需求："我需要能够根据 ID 获取用户"
- 它不关心技术实现：
  ❌ 不关心用 MySQL 还是 PostgreSQL
  ❌ 不关心用 GORM 还是 SQL Builder
  ❌ 不关心有没有缓存
  ✅ 只关心"能获取到用户"这个业务概念

// ==========================================
// Infrastructure 层：实现技术细节
// ==========================================
// domain/user/internal/dal/user.go

type UserDAO struct {
    query *query.Query        // ← 技术细节：用的是 GORM Gen
    cache *redis.Client       // ← 技术细节：用的是 Redis
}

func (dao *UserDAO) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
    // ← 这些都是技术实现细节（Infrastructure）：
    
    // 1. 查 Redis 缓存
    if cached, err := dao.cache.Get(ctx, key).Result(); err == nil {
        return cached, nil  // ← 技术决策：用 Redis
    }
    
    // 2. 查 MySQL 数据库
    user, err := dao.query.User.WithContext(ctx).  // ← 技术决策：用 GORM Gen
        Where(dao.query.User.ID.Eq(userID)).       // ← 技术决策：用 MySQL
        First()
    
    // 3. 写入缓存
    dao.cache.Set(ctx, key, user, 5*time.Minute)  // ← 技术决策：缓存 5 分钟
    
    return user, err
}

💡 这是 Infrastructure 层，因为：
- 包含了大量技术决策
- 依赖具体的技术栈（GORM、Redis、MySQL）
- 如果换技术（如换成 MongoDB），只需要改这里
```

**🔑 判断标准：问自己这个问题**

```
这段代码在说：
A. "业务上需要做什么"  → Domain 层
B. "技术上怎么实现"    → Infrastructure 层

示例：
GetUserByID(userID)       ← A，业务需求（Domain）
dao.query.User.First()    ← B，技术实现（Infrastructure）
```

**📊 对比表格：Domain vs Infrastructure**

| 特征 | Domain 层 (接口) | Infrastructure 层 (实现) |
|------|-----------------|------------------------|
| **关注点** | 业务概念 | 技术细节 |
| **代码示例** | `GetUserByID(userID)` | `db.Query("SELECT * FROM user WHERE id=?")` |
| **依赖** | 不依赖具体技术 | 依赖 GORM、Redis、MySQL 等 |
| **改动原因** | 业务规则变化 | 技术优化、换数据库 |
| **测试方式** | Mock 接口 | 需要数据库环境 |
| **可替换性** | 不可替换（业务核心） | 可替换（技术选型） |

**🎭 实际场景举例**：

假设你要把数据库从 **MySQL 换成 MongoDB**：

```go
// ✅ Domain 层（接口）- 不需要改
type UserRepository interface {
    GetUserByID(ctx context.Context, userID int64) (*model.User, error)
    // 业务需求没变：还是"根据 ID 获取用户"
}

// 🔄 Infrastructure 层 - 需要改实现

// 原来：MySQL 实现
type UserDAO struct {
    db *gorm.DB  // ← 改之前：用 MySQL
}

func (dao *UserDAO) GetUserByID(...) (*model.User, error) {
    return dao.db.Where("id = ?", userID).First(&user).Error
    // ← 改之前：用 SQL 查询
}

// 改成：MongoDB 实现
type UserMongoDAO struct {
    mongo *mongo.Client  // ← 改之后：用 MongoDB
}

func (dao *UserMongoDAO) GetUserByID(...) (*model.User, error) {
    return dao.mongo.Collection("users").FindOne(ctx, bson.M{"_id": userID})
    // ← 改之后：用 MongoDB 查询
}

// ✅ Service 层 - 不需要改
type userImpl struct {
    UserRepo repository.UserRepository  // ← 还是依赖接口
}
// 无论底层用 MySQL 还是 MongoDB，Service 层代码不变！
```

**💡 这就是为什么说 UserDAO 是 Infrastructure 层**：

1. **包含技术决策**：用什么数据库、怎么查询、怎么缓存
2. **可以替换实现**：换数据库只需要改 DAO，不影响业务逻辑
3. **依赖具体技术**：代码里写死了 GORM、Redis、MySQL 等
4. **不是业务概念**：`dao.query.User.First()` 不是业务语言，是技术语言

---

### 🎯 最终总结：一句话理解

```
Domain 层说：   "我需要获取用户信息"        （业务需求）
                      ↓
Infrastructure 层说： "我用 MySQL+Redis 获取"  （技术实现）
```

**为什么 `UserDAO` 在 `domain/user/internal/dal/` 目录？**

- **物理位置**：在 `domain/user/` 下，是为了组织方便（和 User 相关的代码都在一起）
- **逻辑职责**：在 `internal/dal/` 下，表示它是**内部实现细节**，不对外暴露
- **分层归属**：虽然在 `domain/` 目录，但职责是 **Infrastructure 层**

就像：
- 厨房虽然在餐厅里（物理位置），但它是**后勤支撑**（职责归属）
- 你点菜只和服务员交互（Domain 接口），不需要直接和厨师交流（Infrastructure 实现）

---

### 🤔 等等！那 `infra/` 目录和这个 Infrastructure 有什么区别？

非常好的问题！这是两个不同层面的 Infrastructure：

#### 📦 两种 Infrastructure 的区别

```
backend/
├── infra/                           ← 🌍 全局 Infrastructure
│   ├── storage/                     ← 对象存储（MinIO/S3）
│   ├── idgen/                       ← ID 生成器
│   ├── rdb/                         ← 关系数据库抽象
│   ├── cache/                       ← 缓存服务
│   ├── es/                          ← Elasticsearch
│   └── eventbus/                    ← 事件总线（NSQ）
│   └── ...
│   💡 这些是通用的、可复用的基础设施服务
│
└── domain/
    └── user/
        └── internal/dal/            ← 🏠 领域内 Infrastructure
            └── user.go (UserDAO)    ← User 领域的数据访问实现
            💡 这是特定于 User 业务的数据访问层
```

#### 🔑 核心区别

| 特征 | `infra/` 目录 | `domain/*/internal/dal/` |
|------|---------------|-------------------------|
| **层级** | 全局/项目级 | 领域级/模块级 |
| **作用** | 提供通用基础设施服务 | 实现特定领域的数据访问 |
| **复用性** | 跨领域复用 | 只在当前领域使用 |
| **抽象层次** | 高层抽象（接口） | 具体实现 |
| **示例** | `Storage` 接口 | `UserDAO` 实现 |
| **依赖关系** | 不依赖 Domain | 使用 `infra/` 的组件 |

#### 💡 用代码说明依赖关系

```go
// ==========================================
// 1️⃣ infra/ - 全局基础设施（最底层）
// ==========================================
// backend/infra/storage/storage.go
package storage

type Storage interface {
    PutObject(ctx context.Context, key string, content []byte) error
    GetObject(ctx context.Context, key string) ([]byte, error)
}
// 💡 这是全局通用的对象存储接口
// 可以被任何 Domain 使用（User、Plugin、Knowledge 等）

// backend/infra/idgen/idgen.go
package idgen

type IDGenerator interface {
    GenID(ctx context.Context) (int64, error)
}
// 💡 这是全局通用的 ID 生成器
// 可以被任何 Domain 使用

// ==========================================
// 2️⃣ domain/*/internal/dal/ - 领域内数据访问（中间层）
// ==========================================
// backend/domain/user/internal/dal/user.go
package dal

import (
    "github.com/coze-dev/coze-studio/backend/infra/storage"  // ← 依赖 infra
    "gorm.io/gorm"
)

type UserDAO struct {
    db      *gorm.DB           // ← 使用数据库（来自 infra/orm）
    storage storage.Storage    // ← 使用对象存储（来自 infra/storage）
}

func (dao *UserDAO) SaveUserAvatar(ctx context.Context, userID int64, avatar []byte) error {
    // 1. 上传头像到对象存储（使用 infra 的 Storage）
    key := fmt.Sprintf("avatar/%d.jpg", userID)
    err := dao.storage.PutObject(ctx, key, avatar)
    
    // 2. 保存用户记录到数据库（使用 GORM）
    err = dao.db.WithContext(ctx).
        Model(&User{}).
        Where("id = ?", userID).
        Update("avatar_url", key).Error
    
    return err
}
// 💡 这是 User 领域专用的数据访问实现
// 它组合使用了 infra/ 提供的基础设施组件

// ==========================================
// 3️⃣ domain/user/service - 领域服务（业务层）
// ==========================================
// backend/domain/user/service/user_impl.go
package service

type userImpl struct {
    UserRepo repository.UserRepository  // ← 依赖 Repository 接口
}

func (u *userImpl) UpdateAvatar(ctx context.Context, userID int64, avatar []byte) error {
    return u.UserRepo.SaveUserAvatar(ctx, userID, avatar)
}
// 💡 业务逻辑层，不关心底层用什么数据库或存储
```

#### 🏗️ 完整的依赖层次

```
┌────────────────────────────────────────┐
│  API Layer                             │
│  api/handler/                          │
└────────────────┬───────────────────────┘
                 ↓ 调用
┌────────────────────────────────────────┐
│  Application Layer                     │
│  application/user/                     │
└────────────────┬───────────────────────┘
                 ↓ 调用
┌────────────────────────────────────────┐
│  Domain Service Layer                  │
│  domain/user/service/                  │  ← 业务逻辑
└────────────────┬───────────────────────┘
                 ↓ 依赖接口
┌────────────────────────────────────────┐
│  Domain Repository Interface           │
│  domain/user/repository/               │  ← 接口定义
└────────────────┬───────────────────────┘
                 ↓ 实现
┌────────────────────────────────────────┐
│  领域内 Infrastructure                  │
│  domain/user/internal/dal/             │  ← 数据访问实现
│  (UserDAO)                             │
└────────────────┬───────────────────────┘
                 ↓ 使用
┌────────────────────────────────────────┐
│  全局 Infrastructure                    │
│  infra/                                │  ← 基础设施服务
│  - storage/ (对象存储)                  │
│  - idgen/ (ID生成)                     │
│  - cache/ (缓存)                       │
│  - orm/ (数据库)                       │
└────────────────────────────────────────┘
```

#### 🎯 总结

1. **`infra/`** = **通用工具箱**
   - 提供 Storage、IDGen、Cache、DB 等通用服务
   - 所有 Domain 都可以使用
   - 不包含业务逻辑

2. **`domain/*/internal/dal/`** = **领域专用实现**
   - 使用 `infra/` 的工具来实现具体的数据访问
   - 包含该领域的数据访问逻辑
   - 只为当前 Domain 服务

**比喻**：
- `infra/` = 五金店（提供锤子、螺丝刀、电钻）
- `domain/user/internal/dal/` = 木匠（用五金店的工具来做家具）

**🎨 形象图示**：

```
               UserService (业务层)
                    ↓ 需要存储用户头像
               UserRepository 接口
                    ↓ 实现
               UserDAO (领域内 Infrastructure)
                    ↓ 使用工具
        ┌───────────┴───────────┐
        ↓                       ↓
   Storage (infra)         IDGen (infra)
   存储头像文件             生成用户ID
        ↓                       ↓
   MinIO/S3               Snowflake
   (具体技术实现)          (具体技术实现)
```

**实际调用链**：
```go
// 1. 业务层调用
userService.UpdateAvatar(avatar)
    ↓
// 2. Repository 调用
userRepo.SaveUserAvatar(avatar)
    ↓
// 3. DAO 实现（使用 infra 工具）
userDAO.SaveUserAvatar() {
    // 3.1 使用 infra/storage 上传文件
    storage.PutObject("avatar/123.jpg", avatar)
    
    // 3.2 使用 GORM 保存数据库记录
    db.Update("avatar_url", "avatar/123.jpg")
}
```

现在明白了吗？😊

---

**🔌 再用插座和充电器的比喻**

```
你的手机 (Service 层)
    ↓ 需要充电
只认"插座标准" (UserRepository 接口)
    - 必须有两个孔
    - 必须提供 220V 电压
    
    ↓ 不关心插座后面是什么
    
插座后面的实现 (UserDAO)
    - 方案 A: 直接连电网 (原来：直接查数据库)
    - 方案 B: 加个 UPS (现在：加缓存)
    ✅ 不管怎么改，只要插座接口不变，手机不需要知道
```

**具体代码说明**:

```go
// 📜 第1步：定义"合同"（接口）
// domain/user/repository/repository.go
type UserRepository interface {
    GetUserByID(ctx context.Context, userID int64) (*model.User, error)
    // 这是"合同"，说好了提供什么服务
}
✅ 不需要修改

// 🔧 第2步：实现"合同"（具体怎么做）
// domain/user/internal/dal/user.go
type UserDAO struct {
    query *query.Query
    cache *redis.Client  // ← 新增：内部加个缓存
}

func (dao *UserDAO) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
    // ← 内部实现改了（先查缓存再查数据库）
    // 但是！方法签名没变，还是返回 (*model.User, error)
    // 1. 先查缓存 ← 新增逻辑
    // 2. 缓存没有，查数据库
    // 3. 写入缓存 ← 新增逻辑
}
✅ 可以修改内部实现

// 👤 第3步：使用"合同"（Service 层）
// domain/user/service/user_impl.go
type userImpl struct {
    UserRepo repository.UserRepository  // ← 只知道"合同"，不知道具体实现
}

func (u *userImpl) GetUserInfo(...) {
    // 调用接口方法
    // Service 不知道也不关心：
    // - 有没有缓存
    // - 用的是 MySQL 还是 PostgreSQL
    // - 内部怎么实现的
    user, err := u.UserRepo.GetUserByID(ctx, userID)
}
✅ 不需要修改
```

**🎯 核心要点**：

| 层次 | 位置 | 修改？ | 原因 |
|------|------|--------|------|
| 接口定义 | `repository/repository.go` | ❌ 不改 | "合同"不变 |
| 接口实现 | `internal/dal/user.go` | ✅ 可以改 | 内部优化 |
| Service 层 | `service/user_impl.go` | ❌ 不改 | 只依赖接口 |

所以当我们说"Domain 层不需要修改"时，准确的说法是：
- ✅ Domain **核心逻辑**（接口 + Service）不需要修改
- ✅ 只修改 **Infrastructure 实现**（`UserDAO` 的内部实现）
- ⚠️ 虽然 `UserDAO` 文件在 `domain/` 目录，但它是**实现细节**

### 任务 3: 编写单元测试

为 Domain Service 编写测试:

```go
// domain/user/service/user_test.go
func TestServiceImpl_GetUserInfo(t *testing.T) {
    // 1. 创建 Mock Repository
    mockRepo := mock.NewMockUserRepository(t)
    
    // 2. 设置 Mock 行为
    expectedUser := &entity.User{
        UserID: 123,
        Name:   "Test User",
        Email:  "test@example.com",
    }
    mockRepo.EXPECT().
        GetByID(gomock.Any(), int64(123)).
        Return(expectedUser, nil)
    
    // 3. 创建 Service
    svc := &ServiceImpl{
        userRepo: mockRepo,
    }
    
    // 4. 调用方法
    user, err := svc.GetUserInfo(context.Background(), 123)
    
    // 5. 断言
    assert.NoError(t, err)
    assert.Equal(t, int64(123), user.UserID)
    assert.Equal(t, "Test User", user.Name)
}
```

---

## 📖 下一步学习

1. **阅读完整的学习指南**: `BACKEND_LEARNING_GUIDE.md`
2. **选择一个领域深入学习**: 推荐从 Knowledge 或 Workflow 开始
3. **完成实战练习**: 添加新的 API、实现新的 Workflow 节点
4. **阅读测试代码**: 理解最佳实践

---

## ❓ 常见问题

### Q: 为什么要分这么多层？

**A**: 每层有明确的职责，便于:
- 📝 理解和维护
- 🧪 测试 (可以 mock)
- 🔄 替换实现 (如切换数据库)
- 👥 团队协作 (不同层可以并行开发)

### Q: Entity 和 DO 有什么区别？

**A**: 
- **Entity**: 领域概念，反映业务逻辑
- **DO (Data Object)**: 数据库表结构，带 GORM 标签

**示例**:
```go
// Entity: 业务概念
type User struct {
    UserID int64
    Name   string
}

// DO: 数据库映射
type UserDO struct {
    UserID int64 `gorm:"column:user_id;primaryKey"`
    Name   string `gorm:"column:name"`
}
```

### Q: 什么时候使用 Application Service？

**A**: 当需要:
- 协调多个 Domain Service
- 管理事务
- 发布事件
- 转换数据格式

简单的 CRUD 可以直接在 Domain Service 处理。

---

🎉 **恭喜！** 你已经理解了 Coze Studio 后端的核心架构！

现在可以开始深入学习各个领域模块了。加油！💪

