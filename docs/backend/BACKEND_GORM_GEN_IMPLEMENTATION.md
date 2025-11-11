# GORM Gen 实现步骤详解

> 📅 **最后更新**: 2025-01-27  
> 🎯 **目标**: 详解项目中 GORM Gen 从配置到生成的完整流程

---

## 📋 目录

- [概述](#概述)
- [生成脚本位置](#生成脚本位置)
- [配置说明](#配置说明)
- [生成步骤详解](#生成步骤详解)
- [生成的代码结构](#生成的代码结构)
- [实际使用](#实际使用)
- [添加新表](#添加新表)
- [常见问题](#常见问题)

---

## 概述

项目使用 **GORM Gen** 从数据库表自动生成类型安全的查询代码。

### 核心流程

```
数据库表 (MySQL)
    ↓
生成脚本 (gen_orm_query.go)
    ↓
生成代码
    ├── Model (model/*.gen.go)
    └── Query (query/*.gen.go)
    ↓
DAO 实现 (使用生成的 Query)
```

---

## 生成脚本位置

**核心文件**: `backend/types/ddl/gen_orm_query.go`

这是项目中**唯一的** GORM Gen 生成脚本，负责为所有 Domain 生成代码。

---

## 配置说明

### 1. 配置文件结构

生成脚本的核心配置是 `path2Table2Columns2Model`，这是一个三层嵌套的 Map：

```go
var path2Table2Columns2Model = map[string]map[string]map[string]any{
    "domain/user/internal/dal/query": {           // ← 输出路径
        "user": {                                  // ← 数据库表名
            // 空 map = 使用默认类型
        },
        "space": {},
        "space_user": {},
    },
    "domain/plugin/internal/dal/query": {
        "plugin": {
            "manifest":    &plugin.PluginManifest{},  // ← 字段类型映射
            "openapi_doc": &plugin.Openapi3T{},
            "ext":         map[string]any{},
        },
    },
}
```

### 2. 配置层级说明

```
path2Table2Columns2Model
├── 第一层: 输出路径 (string)
│   └── "domain/user/internal/dal/query"
│
├── 第二层: 数据库表名 (string)
│   └── "user"
│
└── 第三层: 字段类型映射 (map[string]any)
    └── "manifest" → &plugin.PluginManifest{}
```

### 3. 字段类型映射

**默认情况**（空 map）：
- 使用 GORM Gen 自动推断的字段类型

**自定义类型**：
```go
"plugin": {
    "manifest":    &plugin.PluginManifest{},  // JSON 字段映射到自定义类型
    "openapi_doc": &plugin.Openapi3T{},
    "ext":         map[string]any{},          // JSON 字段映射到 map
}
```

**实际示例**：
```go
// 数据库表: plugin
// 字段: manifest (JSON 类型)
// 映射到: &plugin.PluginManifest{}

// 生成后的 Model:
type Plugin struct {
    // ...
    Manifest *plugin.PluginManifest `gorm:"column:manifest;serializer:json"`
    // ...
}
```

### 4. 字段可空性配置

```go
var fieldNullablePath = map[string]bool{
    "domain/agent/singleagent/internal/dal/query": true,
    // true = 所有字段都是可选的 (*string, *int64 等)
    // false 或不配置 = 根据数据库表结构决定
}
```

---

## 生成步骤详解

### 步骤 1: 准备数据库连接

```go
// backend/types/ddl/gen_orm_query.go
func main() {
    // 1. 获取数据库连接字符串
    dsn := os.Getenv("MYSQL_DSN")
    if dsn == "" {
        dsn = "root:root@tcp(localhost:3306)/opencoze?charset=utf8mb4&parseTime=True"
    }
    
    // 2. 连接数据库
    gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        NamingStrategy: schema.NamingStrategy{
            SingularTable: true,  // 使用单数表名
        },
    })
}
```

**关键点**：
- 使用环境变量 `MYSQL_DSN` 或默认值
- 配置 `SingularTable: true`（表名使用单数形式）

### 步骤 2: 遍历配置并生成

```go
for path, mapping := range path2Table2Columns2Model {
    // path = "domain/user/internal/dal/query"
    // mapping = {"user": {}, "space": {}, ...}
    
    // 1. 创建 Generator
    g := gen.NewGenerator(gen.Config{
        OutPath:       filepath.Join(rootPath, path),
        Mode:          gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
        FieldNullable: fieldNullablePath[path],
    })
    
    // 2. 连接数据库
    g.UseDB(gormDB)
    
    // 3. 配置特殊字段类型
    g.WithOpts(gen.FieldType("deleted_at", "gorm.DeletedAt"))
}
```

**配置说明**：
- `OutPath`: 生成代码的输出路径
- `Mode`: 
  - `WithoutContext`: 不强制使用 context
  - `WithDefaultQuery`: 生成默认查询方法
  - `WithQueryInterface`: 生成查询接口
- `FieldNullable`: 字段是否可空

### 步骤 3: 字段类型解析

```go
var resolveType func(typ reflect.Type, required bool) string
resolveType = func(typ reflect.Type, required bool) string {
    switch typ.Kind() {
    case reflect.Ptr:
        return resolveType(typ.Elem(), false)
    case reflect.Slice:
        return "[]" + resolveType(typ.Elem(), required)
    default:
        prefix := "*"
        if required {
            prefix = ""
        }
        
        // 如果是当前包的模型，直接使用名称
        if strings.HasSuffix(typ.PkgPath(), modelPath) {
            return prefix + typ.Name()
        }
        
        return prefix + typ.String()
    }
}
```

**作用**：
- 将 Go 类型转换为字符串表示
- 处理指针、切片等复杂类型
- 处理自定义类型

### 步骤 4: 字段修改器

```go
// 自定义字段类型修改器
genModify := func(col string, model any) func(f gen.Field) gen.Field {
    return func(f gen.Field) gen.Field {
        if f.ColumnName != col {
            return f  // 不是目标字段，不修改
        }
        
        st := reflect.TypeOf(model)
        f.Type = resolveType(st, true)
        f.GORMTag.Set("serializer", "json")  // 添加 JSON 序列化标签
        return f
    }
}

// 时间字段修改器
timeModify := func(f gen.Field) gen.Field {
    if f.ColumnName == "updated_at" {
        f.GORMTag.Set("autoUpdateTime", "milli")
    }
    if f.ColumnName == "created_at" {
        f.GORMTag.Set("autoCreateTime", "milli")
    }
    return f
}
```

**作用**：
- `genModify`: 将数据库字段映射到自定义 Go 类型
- `timeModify`: 自动设置时间字段的 GORM 标签

### 步骤 5: 生成模型

```go
var models []any
for table, col2Model := range mapping {
    // table = "user"
    // col2Model = {} 或 {"manifest": &plugin.PluginManifest{}}
    
    opts := make([]gen.ModelOpt, 0, len(col2Model))
    
    // 为每个字段添加修改器
    for column, m := range col2Model {
        cp := m
        opts = append(opts, gen.FieldModify(genModify(column, cp)))
    }
    
    // 添加时间字段修改器
    opts = append(opts, gen.FieldModify(timeModify))
    
    // 生成模型
    models = append(models, g.GenerateModel(table, opts...))
}

// 应用所有模型
g.ApplyBasic(models...)

// 执行生成
g.Execute()
```

**流程**：
1. 遍历每个表配置
2. 为每个字段创建修改器
3. 生成模型结构
4. 应用到 Generator
5. 执行生成

---

## 生成的代码结构

### 1. Model (模型)

**位置**: `backend/domain/user/internal/dal/model/user.gen.go`

```go
// Code generated by gorm.io/gen. DO NOT EDIT.
package model

import (
    "gorm.io/gorm"
)

const TableNameUser = "user"

type User struct {
    ID           int64          `gorm:"column:id;primaryKey;autoIncrement:true"`
    Name         string         `gorm:"column:name;not null"`
    Email        string         `gorm:"column:email;not null"`
    CreatedAt    int64          `gorm:"column:created_at;not null;autoCreateTime:milli"`
    UpdatedAt    int64          `gorm:"column:updated_at;not null;autoUpdateTime:milli"`
    DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (*User) TableName() string {
    return TableNameUser
}
```

**特点**：
- ✅ 包含完整的 GORM 标签
- ✅ 自动设置主键、时间戳等
- ✅ 支持软删除（`DeletedAt`）

### 2. Query (查询)

**位置**: `backend/domain/user/internal/dal/query/user.gen.go`

```go
// Code generated by gorm.io/gen. DO NOT EDIT.
package query

type user struct {
    userDo
    
    ALL          field.Asterisk
    ID           field.Int64
    Name         field.String
    Email        field.String
    // ...
}

// 类型安全的查询方法
func (u user) Where(conds ...gen.Condition) *userDo {
    // ...
}

func (u user) First() (*model.User, error) {
    // ...
}
```

**特点**：
- ✅ 类型安全的字段访问 (`u.ID`, `u.Email`)
- ✅ 链式查询方法
- ✅ 编译时检查

### 3. Query 入口

**位置**: `backend/domain/user/internal/dal/query/gen.go`

```go
// Code generated by gorm.io/gen. DO NOT EDIT.
package query

func Use(db *gorm.DB, opts ...gen.DOOption) *Query {
    return &Query{
        db:        db,
        User:      newUser(db, opts...),
        Space:     newSpace(db, opts...),
        SpaceUser: newSpaceUser(db, opts...),
    }
}

type Query struct {
    db *gorm.DB
    
    User      user
    Space     space
    SpaceUser spaceUser
}

func (q *Query) WithContext(ctx context.Context) *queryCtx {
    return &queryCtx{
        User:      q.User.WithContext(ctx),
        Space:     q.Space.WithContext(ctx),
        SpaceUser: q.SpaceUser.WithContext(ctx),
    }
}
```

**作用**：
- 提供统一的查询入口
- 管理所有表的查询对象
- 支持 Context 传递

---

## 实际使用

### 在 DAO 中使用生成的代码

```go
// backend/domain/user/internal/dal/user.go
package dal

import (
    "context"
    "gorm.io/gorm"
    "github.com/coze-dev/coze-studio/backend/domain/user/internal/dal/query"
    "github.com/coze-dev/coze-studio/backend/domain/user/internal/dal/model"
)

// 创建 DAO
func NewUserDAO(db *gorm.DB) *UserDAO {
    return &UserDAO{
        query: query.Use(db),  // ← 使用生成的 Query
    }
}

type UserDAO struct {
    query *query.Query  // ← 生成的查询对象
}

// 查询用户
func (dao *UserDAO) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
    // ✅ 类型安全的查询
    return dao.query.User.WithContext(ctx).
        Where(dao.query.User.ID.Eq(userID)).  // ← 类型安全的字段访问
        First()
}

// 根据邮箱查询
func (dao *UserDAO) GetUsersByEmail(ctx context.Context, email string) (*model.User, bool, error) {
    user, err := dao.query.User.WithContext(ctx).
        Where(dao.query.User.Email.Eq(email)).  // ← 类型安全
        First()
    
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, false, nil
    }
    return user, true, err
}

// 更新用户
func (dao *UserDAO) UpdateProfile(ctx context.Context, userID int64, updates map[string]interface{}) error {
    _, err := dao.query.User.WithContext(ctx).
        Where(dao.query.User.ID.Eq(userID)).
        Updates(updates)
    return err
}
```

**优势**：
- ✅ 类型安全：`dao.query.User.Email` 编译时检查
- ✅ 自动补全：IDE 可以自动补全字段名
- ✅ 避免错误：不会出现字段名拼写错误

---

## 添加新表

### 场景：为 `note` 表生成代码

#### 步骤 1: 在配置中添加表

编辑 `backend/types/ddl/gen_orm_query.go`：

```go
var path2Table2Columns2Model = map[string]map[string]map[string]any{
    // ... 现有配置 ...
    
    "domain/note/internal/dal/query": {  // ← 新增路径
        "note": {                         // ← 表名
            // 空 map = 使用默认类型
        },
    },
}
```

#### 步骤 2: 运行生成脚本

```bash
# 设置数据库连接
export MYSQL_DSN="root:root@tcp(localhost:3306)/opencoze?charset=utf8mb4&parseTime=True"

# 运行生成脚本
cd backend/types/ddl
go run gen_orm_query.go
```

#### 步骤 3: 检查生成的代码

生成后应该看到：

```
backend/domain/note/internal/dal/
├── model/
│   └── note.gen.go          ← 自动生成
└── query/
    ├── gen.go               ← 自动生成
    └── note.gen.go          ← 自动生成
```

#### 步骤 4: 在 DAO 中使用

```go
// backend/domain/note/internal/dal/note.go
package dal

import (
    "context"
    "gorm.io/gorm"
    "github.com/coze-dev/coze-studio/backend/domain/note/internal/dal/query"
)

func NewNoteDAO(db *gorm.DB) *NoteDAO {
    return &NoteDAO{
        query: query.Use(db),  // ← 使用生成的 Query
    }
}

type NoteDAO struct {
    query *query.Query
}

func (dao *NoteDAO) GetByID(ctx context.Context, noteID int64) (*model.Note, error) {
    return dao.query.Note.WithContext(ctx).
        Where(dao.query.Note.ID.Eq(noteID)).
        First()
}
```

---

## 自定义字段类型

### 场景：`plugin` 表的 `manifest` 字段是 JSON，需要映射到自定义类型

#### 步骤 1: 定义自定义类型

```go
// backend/crossdomain/plugin/model/plugin.go
package plugin

type PluginManifest struct {
    Name        string `json:"name"`
    Version     string `json:"version"`
    Description string `json:"description"`
}
```

#### 步骤 2: 在配置中映射

```go
// backend/types/ddl/gen_orm_query.go
var path2Table2Columns2Model = map[string]map[string]map[string]any{
    "domain/plugin/internal/dal/query": {
        "plugin": {
            "manifest": &plugin.PluginManifest{},  // ← 字段类型映射
        },
    },
}
```

#### 步骤 3: 生成代码

运行生成脚本后，生成的 Model 会是：

```go
// backend/domain/plugin/internal/dal/model/plugin.gen.go
type Plugin struct {
    // ...
    Manifest *plugin.PluginManifest `gorm:"column:manifest;serializer:json"`
    // ...
}
```

#### 步骤 4: 使用

```go
// 创建插件
plugin := &model.Plugin{
    Manifest: &plugin.PluginManifest{
        Name:    "My Plugin",
        Version: "1.0.0",
    },
}
dao.query.Plugin.Create(plugin)

// 查询插件
p, err := dao.query.Plugin.First()
fmt.Println(p.Manifest.Name)  // ✅ 类型安全访问
```

---

## 常见问题

### Q1: 如何运行生成脚本？

**A**: 
```bash
# 方式 1: 使用环境变量
export MYSQL_DSN="root:root@tcp(localhost:3306)/opencoze?charset=utf8mb4&parseTime=True"
cd backend/types/ddl
go run gen_orm_query.go

# 方式 2: 直接使用默认值（如果数据库在本机）
cd backend/types/ddl
go run gen_orm_query.go
```

### Q2: 生成的代码在哪里？

**A**: 
根据配置的 `path`：
- Model: `{path}/../model/`
- Query: `{path}/`

例如：
- `domain/user/internal/dal/query` → Model 在 `domain/user/internal/dal/model/`

### Q3: 可以手动修改生成的代码吗？

**A**: ❌ **不可以！**

所有生成的代码都有注释：
```go
// Code generated by gorm.io/gen. DO NOT EDIT.
```

手动修改会在下次生成时被覆盖。

### Q4: 如何为 JSON 字段指定类型？

**A**: 在配置中添加字段映射：

```go
"plugin": {
    "manifest": &plugin.PluginManifest{},  // JSON 字段 → 自定义类型
    "ext":      map[string]any{},           // JSON 字段 → map
}
```

### Q5: 生成后需要做什么？

**A**: 
1. ✅ 检查生成的代码是否能编译
2. ✅ 在 DAO 中使用生成的 Query
3. ✅ 测试数据访问功能

### Q6: 为什么查询时需要 `WithContext(ctx)`？

**A**: 
- Context 用于传递请求上下文（超时、取消等）
- GORM Gen 支持 Context，但不是必须的
- 项目中统一使用 `WithContext(ctx)` 保持一致性

### Q7: 如何添加时间字段自动更新？

**A**: 
生成脚本已经自动处理：

```go
timeModify := func(f gen.Field) gen.Field {
    if f.ColumnName == "updated_at" {
        f.GORMTag.Set("autoUpdateTime", "milli")
    }
    if f.ColumnName == "created_at" {
        f.GORMTag.Set("autoCreateTime", "milli")
    }
    return f
}
```

只要数据库表有 `created_at` 和 `updated_at` 字段，就会自动设置。

---

## 📚 相关文档

- [GORM Gen 官方文档](https://gorm.io/gen/) - 官方使用指南
- [后端快速入门](./BACKEND_QUICKSTART.md) - 后端开发流程
- [后端实战练习](./BACKEND_PRACTICE.md) - 实际开发案例
- [后端 GORM Gen 指南](./BACKEND_GORM_GEN_GUIDE.md) - GORM Gen 使用指南

---

## 🎯 总结

### 核心要点

1. ✅ **统一生成脚本** - `backend/types/ddl/gen_orm_query.go`
2. ✅ **配置驱动** - `path2Table2Columns2Model` 配置所有表
3. ✅ **类型安全** - 生成的 Query 提供编译时检查
4. ✅ **自动处理** - 时间字段、软删除等自动配置

### 生成流程

```
1. 配置 path2Table2Columns2Model
   ↓
2. 运行生成脚本 (go run gen_orm_query.go)
   ↓
3. 生成 Model 和 Query
   ↓
4. 在 DAO 中使用 query.Use(db)
```

### 关键代码位置

| 文件 | 说明 |
|------|------|
| `backend/types/ddl/gen_orm_query.go` | 生成脚本 |
| `backend/domain/*/internal/dal/model/` | 生成的 Model |
| `backend/domain/*/internal/dal/query/` | 生成的 Query |
| `backend/domain/*/internal/dal/*.go` | DAO 实现（手写）|

---

**💡 提示**: 修改数据库表结构后，记得重新运行生成脚本更新代码！

