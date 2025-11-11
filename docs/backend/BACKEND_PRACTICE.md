# 后端实战练习：笔记管理功能

> ⚠️ **重要提示**：本文档完全基于实际项目代码编写，所有示例都遵循项目的真实实现模式。

## 📋 目录
- [实战目标](#实战目标)
- [前置知识](#前置知识)
- [Step 1: 数据库设计](#step-1-数据库设计)
- [Step 2: 生成 GORM 代码](#step-2-生成-gorm-代码)
- [Step 3: 实现 Repository](#step-3-实现-repository)
- [Step 4: 实现 Domain Service](#step-4-实现-domain-service)
- [Step 5: 实现 Application Service](#step-5-实现-application-service)
- [Step 6: 定义 IDL](#step-6-定义-idl)
- [Step 7: 生成 API Handler](#step-7-生成-api-handler)
- [Step 8: 测试](#step-8-测试)
- [常见问题](#常见问题)

---

## 实战目标

我们将实现一个 **笔记管理功能**，包含以下 API：
- ✅ 创建笔记
- ✅ 获取笔记详情
- ✅ 更新笔记
- ✅ 删除笔记
- ✅ 获取用户的笔记列表

通过这个实战，你将学会：
1. 按照项目规范设计数据库表
2. 使用 GORM Gen 自动生成代码
3. 实现 Repository、Domain Service、Application Service
4. 定义 Thrift IDL 并生成 API 代码
5. 测试完整的请求流程

---

## 前置知识

在开始之前，请确保你已经阅读了以下文档：
- ✅ `BACKEND_ERRATA.md` - 了解项目的实际实现模式
- ✅ `BACKEND_QUICKSTART.md` - 理解请求流程
- ✅ `BACKEND_GORM_GEN_GUIDE.md` - 了解 GORM Gen 的使用

---

## Step 1: 数据库设计

### 1.1 设计笔记表

创建文件：`docker/atlas/migrations/YYYYMMDDHHMMSS_add_note.sql`

```sql
-- Create 'note' table
CREATE TABLE IF NOT EXISTS `note` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'Primary Key ID',
    `user_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'User ID',
    `space_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'Space ID',
    `title` varchar(255) NOT NULL DEFAULT '' COMMENT 'Note Title',
    `content` text NULL COMMENT 'Note Content',
    `status` tinyint unsigned NOT NULL DEFAULT 1 COMMENT 'Status: 1-Normal 2-Deleted',
    `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'Create Time in Milliseconds',
    `updated_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'Update Time in Milliseconds',
    `deleted_at` datetime(3) NULL COMMENT 'Delete Time',
    PRIMARY KEY (`id`),
    INDEX `idx_user_id_status` (`user_id`, `status`),
    INDEX `idx_space_id_status` (`space_id`, `status`)
) ENGINE=InnoDB CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'Note Table';
```

### 1.2 MySQL 规则检查

✅ **遵循的规则**：
- [x] 所有字段设置 `NOT NULL`（`text` 和 `datetime` 除外）
- [x] 单表索引数量：2个（未超过6个）
- [x] 表存储引擎：`InnoDB`
- [x] 表字符集：`utf8mb4`，Collation：`utf8mb4_unicode_ci`
- [x] 表添加了 `COMMENT` 注释
- [x] 每个字段都有 `COMMENT` 注释
- [x] 设置了主键 `id`
- [x] `NOT NULL` 字段都设置了默认值
- [x] `created_at` 和 `updated_at` 使用 `bigint unsigned`
- [x] 索引命名：`idx_` 开头

### 1.3 应用迁移

```bash
# 进入项目根目录

# 应用数据库迁移
make db_migrate_apply
```

---

## Step 2: 生成 GORM 代码

### 2.1 创建目录结构

```bash
mkdir -p backend/domain/note/internal/dal/{model,query}
mkdir -p backend/domain/note/entity
mkdir -p backend/domain/note/service
mkdir -p backend/domain/note/repository
mkdir -p backend/application/note
```

### 2.2 创建 GORM Gen 生成脚本

创建文件：`backend/domain/note/internal/dal/query/gen.go`

```go
package query

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
	"gorm.io/gen"
	"gorm.io/plugin/dbresolver"
)

var (
	Q    = new(Query)
	Note *note
)

func SetDefault(db *gorm.DB, opts ...gen.DOOption) {
	*Q = *Use(db, opts...)
	Note = &Q.Note
}

func Use(db *gorm.DB, opts ...gen.DOOption) *Query {
	return &Query{
		db:   db,
		Note: newNote(db, opts...),
	}
}

type Query struct {
	db   *gorm.DB
	Note note
}

func (q *Query) Available() bool { return q.db != nil }

func (q *Query) clone(db *gorm.DB) *Query {
	return &Query{
		db:   db,
		Note: q.Note.clone(db),
	}
}

func (q *Query) ReadDB() *Query {
	return q.ReplaceDB(q.db.Clauses(dbresolver.Read))
}

func (q *Query) WriteDB() *Query {
	return q.ReplaceDB(q.db.Clauses(dbresolver.Write))
}

func (q *Query) ReplaceDB(db *gorm.DB) *Query {
	return &Query{
		db:   db,
		Note: q.Note.replaceDB(db),
	}
}

type queryCtx struct {
	Note INoteDo
}

func (q *Query) WithContext(ctx context.Context) *queryCtx {
	return &queryCtx{
		Note: q.Note.WithContext(ctx),
	}
}

func (q *Query) Transaction(fc func(tx *Query) error, opts ...*sql.TxOptions) error {
	return q.db.Transaction(func(tx *gorm.DB) error { return fc(q.clone(tx)) }, opts...)
}

func (q *Query) Begin(opts ...*sql.TxOptions) *QueryTx {
	tx := q.db.Begin(opts...)
	return &QueryTx{Query: q.clone(tx), Error: tx.Error}
}

type QueryTx struct {
	*Query
	Error error
}

func (q *QueryTx) Commit() error {
	return q.db.Commit().Error
}

func (q *QueryTx) Rollback() error {
	return q.db.Rollback().Error
}

func (q *QueryTx) SavePoint(name string) error {
	return q.db.SavePoint(name).Error
}

func (q *QueryTx) RollbackTo(name string) error {
	return q.db.RollbackTo(name).Error
}
```

### 2.3 运行 GORM Gen

```bash
cd backend/domain/note/internal/dal

# 使用 GORM Gen CLI 生成代码
go run gorm.io/gen/tools/gentool@latest -dsn "user:password@tcp(localhost:3306)/opencoze?charset=utf8mb4&parseTime=True&loc=Local" -tables note -outPath ./query -outFile query_gen.go -modelPkgPath ./model
```

⚠️ **注意**：实际项目中，GORM Gen 配置通常在统一的位置，这里仅作演示。

生成后你将得到：
- `backend/domain/note/internal/dal/model/note.gen.go` - Model 定义
- `backend/domain/note/internal/dal/query/note.gen.go` - 查询代码

---

## Step 3: 实现 Repository

### 3.1 定义 Entity

创建文件：`backend/domain/note/entity/note.go`

```go
package entity

const (
	NoteStatusNormal  = 1
	NoteStatusDeleted = 2
)

// Note 笔记领域实体
type Note struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	SpaceID   int64  `json:"space_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Status    int    `json:"status"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}
```

### 3.2 定义 Repository 接口

创建文件：`backend/domain/note/repository/repository.go`

```go
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/note/entity"
	"github.com/coze-dev/coze-studio/backend/domain/note/internal/dal"
)

// NoteRepository 笔记仓储接口
type NoteRepository interface {
	Create(ctx context.Context, note *entity.Note) error
	GetByID(ctx context.Context, noteID int64) (*entity.Note, bool, error)
	Update(ctx context.Context, note *entity.Note) error
	Delete(ctx context.Context, noteID int64) error
	ListByUserID(ctx context.Context, userID int64, offset, limit int) ([]*entity.Note, error)
}

// NewNoteRepo 创建 NoteRepository 实例
func NewNoteRepo(db *gorm.DB) NoteRepository {
	return dal.NewNoteDAO(db)
}
```

### 3.3 实现 Repository

创建文件：`backend/domain/note/internal/dal/note.go`

```go
package dal

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/note/entity"
	"github.com/coze-dev/coze-studio/backend/domain/note/internal/dal/model"
	"github.com/coze-dev/coze-studio/backend/domain/note/internal/dal/query"
)

func NewNoteDAO(db *gorm.DB) *NoteDAO {
	return &NoteDAO{
		query: query.Use(db),
	}
}

type NoteDAO struct {
	query *query.Query
}

// Create 创建笔记
func (dao *NoteDAO) Create(ctx context.Context, note *entity.Note) error {
	noteModel := &model.Note{
		ID:        note.ID,
		UserID:    uint64(note.UserID),
		SpaceID:   uint64(note.SpaceID),
		Title:     note.Title,
		Content:   &note.Content,
		Status:    uint8(note.Status),
		CreatedAt: uint64(note.CreatedAt),
		UpdatedAt: uint64(note.UpdatedAt),
	}

	return dao.query.Note.WithContext(ctx).Create(noteModel)
}

// GetByID 根据 ID 获取笔记
func (dao *NoteDAO) GetByID(ctx context.Context, noteID int64) (*entity.Note, bool, error) {
	noteModel, err := dao.query.Note.WithContext(ctx).
		Where(dao.query.Note.ID.Eq(uint64(noteID))).
		Where(dao.query.Note.Status.Eq(entity.NoteStatusNormal)).
		First()

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}

	return dao.modelToEntity(noteModel), true, nil
}

// Update 更新笔记
func (dao *NoteDAO) Update(ctx context.Context, note *entity.Note) error {
	updates := map[string]interface{}{
		"title":      note.Title,
		"content":    note.Content,
		"updated_at": note.UpdatedAt,
	}

	_, err := dao.query.Note.WithContext(ctx).
		Where(dao.query.Note.ID.Eq(uint64(note.ID))).
		Updates(updates)

	return err
}

// Delete 删除笔记（软删除）
func (dao *NoteDAO) Delete(ctx context.Context, noteID int64) error {
	_, err := dao.query.Note.WithContext(ctx).
		Where(dao.query.Note.ID.Eq(uint64(noteID))).
		Update(dao.query.Note.Status, entity.NoteStatusDeleted)

	return err
}

// ListByUserID 获取用户的笔记列表
func (dao *NoteDAO) ListByUserID(ctx context.Context, userID int64, offset, limit int) ([]*entity.Note, error) {
	noteModels, err := dao.query.Note.WithContext(ctx).
		Where(dao.query.Note.UserID.Eq(uint64(userID))).
		Where(dao.query.Note.Status.Eq(entity.NoteStatusNormal)).
		Order(dao.query.Note.CreatedAt.Desc()).
		Offset(offset).
		Limit(limit).
		Find()

	if err != nil {
		return nil, err
	}

	notes := make([]*entity.Note, 0, len(noteModels))
	for _, noteModel := range noteModels {
		notes = append(notes, dao.modelToEntity(noteModel))
	}

	return notes, nil
}

// modelToEntity 将 GORM Model 转换为领域实体
func (dao *NoteDAO) modelToEntity(noteModel *model.Note) *entity.Note {
	content := ""
	if noteModel.Content != nil {
		content = *noteModel.Content
	}

	return &entity.Note{
		ID:        int64(noteModel.ID),
		UserID:    int64(noteModel.UserID),
		SpaceID:   int64(noteModel.SpaceID),
		Title:     noteModel.Title,
		Content:   content,
		Status:    int(noteModel.Status),
		CreatedAt: int64(noteModel.CreatedAt),
		UpdatedAt: int64(noteModel.UpdatedAt),
	}
}
```

### 3.4 关键点说明

✅ **使用 GORM Gen 的类型安全查询**：
```go
dao.query.Note.WithContext(ctx).Where(dao.query.Note.ID.Eq(uint64(noteID)))
```

✅ **Model 与 Entity 分离**：
- `model.Note` - GORM 生成的数据库模型
- `entity.Note` - 领域实体

---

## Step 4: 实现 Domain Service

### 4.1 定义 Service 接口

创建文件：`backend/domain/note/service/note.go`

```go
package service

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/note/entity"
)

// CreateNoteRequest 创建笔记请求
type CreateNoteRequest struct {
	UserID  int64
	SpaceID int64
	Title   string
	Content string
}

// UpdateNoteRequest 更新笔记请求
type UpdateNoteRequest struct {
	NoteID  int64
	UserID  int64
	Title   string
	Content string
}

// ListNotesRequest 获取笔记列表请求
type ListNotesRequest struct {
	UserID int64
	Offset int
	Limit  int
}

// Note 笔记领域服务接口
type Note interface {
	Create(ctx context.Context, req *CreateNoteRequest) (*entity.Note, error)
	GetByID(ctx context.Context, noteID, userID int64) (*entity.Note, error)
	Update(ctx context.Context, req *UpdateNoteRequest) error
	Delete(ctx context.Context, noteID, userID int64) error
	ListByUserID(ctx context.Context, req *ListNotesRequest) ([]*entity.Note, error)
}
```

### 4.2 实现 Domain Service

创建文件：`backend/domain/note/service/impl/note.go`

```go
package impl

import (
	"context"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/note/entity"
	"github.com/coze-dev/coze-studio/backend/domain/note/repository"
	"github.com/coze-dev/coze-studio/backend/domain/note/service"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
	"github.com/coze-dev/coze-studio/backend/pkg/errorx"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

// Components 依赖组件
type Components struct {
	IDGen    idgen.IDGenerator
	NoteRepo repository.NoteRepository
}

func NewNoteDomain(ctx context.Context, components *Components) *NoteDomainService {
	return &NoteDomainService{
		idgen:    components.IDGen,
		noteRepo: components.NoteRepo,
	}
}

type NoteDomainService struct {
	idgen    idgen.IDGenerator
	noteRepo repository.NoteRepository
}

// Create 创建笔记
func (s *NoteDomainService) Create(ctx context.Context, req *service.CreateNoteRequest) (*entity.Note, error) {
	// 参数验证
	if req.Title == "" {
		return nil, errorx.New(errno.ErrUserInvalidParamCode, errorx.KV("msg", "title is required"))
	}

	// 生成 ID
	noteID, err := s.idgen.GenerateID()
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	note := &entity.Note{
		ID:        noteID,
		UserID:    req.UserID,
		SpaceID:   req.SpaceID,
		Title:     req.Title,
		Content:   req.Content,
		Status:    entity.NoteStatusNormal,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.noteRepo.Create(ctx, note); err != nil {
		return nil, err
	}

	return note, nil
}

// GetByID 获取笔记详情
func (s *NoteDomainService) GetByID(ctx context.Context, noteID, userID int64) (*entity.Note, error) {
	note, exist, err := s.noteRepo.GetByID(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, errorx.New(errno.ErrResourceNotFoundCode, errorx.KV("msg", "note not found"))
	}

	// 权限检查：只能查看自己的笔记
	if note.UserID != userID {
		return nil, errorx.New(errno.ErrPermissionDeniedCode, errorx.KV("msg", "permission denied"))
	}

	return note, nil
}

// Update 更新笔记
func (s *NoteDomainService) Update(ctx context.Context, req *service.UpdateNoteRequest) error {
	// 检查笔记是否存在且属于当前用户
	note, err := s.GetByID(ctx, req.NoteID, req.UserID)
	if err != nil {
		return err
	}

	// 更新字段
	note.Title = req.Title
	note.Content = req.Content
	note.UpdatedAt = time.Now().UnixMilli()

	return s.noteRepo.Update(ctx, note)
}

// Delete 删除笔记
func (s *NoteDomainService) Delete(ctx context.Context, noteID, userID int64) error {
	// 检查笔记是否存在且属于当前用户
	_, err := s.GetByID(ctx, noteID, userID)
	if err != nil {
		return err
	}

	return s.noteRepo.Delete(ctx, noteID)
}

// ListByUserID 获取用户的笔记列表
func (s *NoteDomainService) ListByUserID(ctx context.Context, req *service.ListNotesRequest) ([]*entity.Note, error) {
	return s.noteRepo.ListByUserID(ctx, req.UserID, req.Offset, req.Limit)
}
```

### 4.3 关键点说明

✅ **业务逻辑在领域服务中**：
- 参数验证
- 权限检查
- ID 生成
- 时间戳设置

✅ **依赖注入模式**：
```go
type Components struct {
	IDGen    idgen.IDGenerator
	NoteRepo repository.NoteRepository
}
```

---

## Step 5: 实现 Application Service

### 5.1 创建全局变量

创建文件：`backend/application/note/note.go`

```go
package note

import (
	"github.com/coze-dev/coze-studio/backend/domain/note/service"
)

// NoteApplicationSVC 笔记应用服务全局变量（单例）
var NoteApplicationSVC = &NoteApplicationService{}

// NoteApplicationService 笔记应用服务
type NoteApplicationService struct {
	DomainSVC service.Note // 领域服务
}
```

### 5.2 初始化服务

创建文件：`backend/application/note/init.go`

```go
package note

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/note/repository"
	"github.com/coze-dev/coze-studio/backend/domain/note/service"
	serviceImpl "github.com/coze-dev/coze-studio/backend/domain/note/service/impl"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
)

// InitService 初始化笔记应用服务
func InitService(ctx context.Context, db *gorm.DB, idgen idgen.IDGenerator) *NoteApplicationService {
	// 初始化领域服务（填充全局变量）
	NoteApplicationSVC.DomainSVC = serviceImpl.NewNoteDomain(ctx, &serviceImpl.Components{
		IDGen:    idgen,
		NoteRepo: repository.NewNoteRepo(db),
	})

	return NoteApplicationSVC
}
```

### 5.3 实现应用服务方法

在 `backend/application/note/note.go` 中添加：

```go
import (
	"context"

	"github.com/coze-dev/coze-studio/backend/api/model/note"
	"github.com/coze-dev/coze-studio/backend/application/base/ctxutil"
	"github.com/coze-dev/coze-studio/backend/domain/note/entity"
	noteService "github.com/coze-dev/coze-studio/backend/domain/note/service"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
)

// CreateNote 创建笔记
func (s *NoteApplicationService) CreateNote(ctx context.Context, req *note.CreateNoteRequest) (*note.CreateNoteResponse, error) {
	userID := ctxutil.MustGetUIDFromCtx(ctx)
	spaceID := req.GetSpaceId()

	noteEntity, err := s.DomainSVC.Create(ctx, &noteService.CreateNoteRequest{
		UserID:  userID,
		SpaceID: spaceID,
		Title:   req.GetTitle(),
		Content: req.GetContent(),
	})
	if err != nil {
		return nil, err
	}

	return &note.CreateNoteResponse{
		Data: entityToDTO(noteEntity),
		Code: 0,
	}, nil
}

// GetNoteDetail 获取笔记详情
func (s *NoteApplicationService) GetNoteDetail(ctx context.Context, req *note.GetNoteDetailRequest) (*note.GetNoteDetailResponse, error) {
	userID := ctxutil.MustGetUIDFromCtx(ctx)

	noteEntity, err := s.DomainSVC.GetByID(ctx, req.GetNoteId(), userID)
	if err != nil {
		return nil, err
	}

	return &note.GetNoteDetailResponse{
		Data: entityToDTO(noteEntity),
		Code: 0,
	}, nil
}

// UpdateNote 更新笔记
func (s *NoteApplicationService) UpdateNote(ctx context.Context, req *note.UpdateNoteRequest) (*note.UpdateNoteResponse, error) {
	userID := ctxutil.MustGetUIDFromCtx(ctx)

	err := s.DomainSVC.Update(ctx, &noteService.UpdateNoteRequest{
		NoteID:  req.GetNoteId(),
		UserID:  userID,
		Title:   req.GetTitle(),
		Content: req.GetContent(),
	})
	if err != nil {
		return nil, err
	}

	return &note.UpdateNoteResponse{
		Code: 0,
	}, nil
}

// DeleteNote 删除笔记
func (s *NoteApplicationService) DeleteNote(ctx context.Context, req *note.DeleteNoteRequest) (*note.DeleteNoteResponse, error) {
	userID := ctxutil.MustGetUIDFromCtx(ctx)

	err := s.DomainSVC.Delete(ctx, req.GetNoteId(), userID)
	if err != nil {
		return nil, err
	}

	return &note.DeleteNoteResponse{
		Code: 0,
	}, nil
}

// ListUserNotes 获取用户笔记列表
func (s *NoteApplicationService) ListUserNotes(ctx context.Context, req *note.ListUserNotesRequest) (*note.ListUserNotesResponse, error) {
	userID := ctxutil.MustGetUIDFromCtx(ctx)

	offset := int(req.GetOffset())
	limit := int(req.GetLimit())
	if limit == 0 {
		limit = 20 // 默认值
	}

	noteEntities, err := s.DomainSVC.ListByUserID(ctx, &noteService.ListNotesRequest{
		UserID: userID,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	notes := make([]*note.NoteInfo, 0, len(noteEntities))
	for _, noteEntity := range noteEntities {
		notes = append(notes, entityToDTO(noteEntity))
	}

	return &note.ListUserNotesResponse{
		Data: &note.NoteList{
			Notes:   notes,
			Total:   ptr.Of(int32(len(notes))),
			HasMore: ptr.Of(len(notes) >= limit),
		},
		Code: 0,
	}, nil
}

// entityToDTO 将领域实体转换为 API 模型
func entityToDTO(noteEntity *entity.Note) *note.NoteInfo {
	return &note.NoteInfo{
		NoteId:    noteEntity.ID,
		UserId:    noteEntity.UserID,
		SpaceId:   noteEntity.SpaceID,
		Title:     noteEntity.Title,
		Content:   noteEntity.Content,
		CreatedAt: noteEntity.CreatedAt,
		UpdatedAt: noteEntity.UpdatedAt,
	}
}
```

### 5.4 注册到应用初始化

在 `backend/application/application.go` 中添加：

```go
import (
	"github.com/coze-dev/coze-studio/backend/application/note"
)

func Init(ctx context.Context) (err error) {
	// ... 其他初始化 ...

	// 初始化笔记服务
	note.InitService(ctx, infra.Db(), infra.IDGen())

	return nil
}
```

---

## Step 6: 定义 IDL

### 6.1 创建 Thrift IDL

创建文件：`idl/note/note.thrift`

```thrift
namespace go note

// 笔记信息
struct NoteInfo {
    1: required i64 note_id (agw.js_conv="str", api.js_conv="true")
    2: required i64 user_id (agw.js_conv="str", api.js_conv="true")
    3: required i64 space_id (agw.js_conv="str", api.js_conv="true")
    4: required string title
    5: required string content
    6: required i64 created_at
    7: required i64 updated_at
}

// 创建笔记请求
struct CreateNoteRequest {
    1: required i64 space_id (agw.js_conv="str", api.js_conv="true")
    2: required string title
    3: required string content
}

// 创建笔记响应
struct CreateNoteResponse {
    1: required NoteInfo data
    253: required i32 code
    254: required string msg
}

// 获取笔记详情请求
struct GetNoteDetailRequest {
    1: required i64 note_id (agw.js_conv="str", api.js_conv="true", api.path="note_id")
}

// 获取笔记详情响应
struct GetNoteDetailResponse {
    1: required NoteInfo data
    253: required i32 code
    254: required string msg
}

// 更新笔记请求
struct UpdateNoteRequest {
    1: required i64 note_id (agw.js_conv="str", api.js_conv="true")
    2: required string title
    3: required string content
}

// 更新笔记响应
struct UpdateNoteResponse {
    253: required i32 code
    254: required string msg
}

// 删除笔记请求
struct DeleteNoteRequest {
    1: required i64 note_id (agw.js_conv="str", api.js_conv="true", api.path="note_id")
}

// 删除笔记响应
struct DeleteNoteResponse {
    253: required i32 code
    254: required string msg
}

// 获取笔记列表请求
struct ListUserNotesRequest {
    1: optional i32 offset
    2: optional i32 limit
}

// 笔记列表
struct NoteList {
    1: required list<NoteInfo> notes
    2: optional i32 total
    3: optional bool has_more
}

// 获取笔记列表响应
struct ListUserNotesResponse {
    1: required NoteList data
    253: required i32 code
    254: required string msg
}

// 笔记服务
service NoteService {
    // 创建笔记
    CreateNoteResponse CreateNote(1: CreateNoteRequest req) (api.post="/api/note/create")

    // 获取笔记详情
    GetNoteDetailResponse GetNoteDetail(1: GetNoteDetailRequest req) (api.get="/api/note/:note_id")

    // 更新笔记
    UpdateNoteResponse UpdateNote(1: UpdateNoteRequest req) (api.post="/api/note/update")

    // 删除笔记
    DeleteNoteResponse DeleteNote(1: DeleteNoteRequest req) (api.delete="/api/note/:note_id")

    // 获取用户笔记列表
    ListUserNotesResponse ListUserNotes(1: ListUserNotesRequest req) (api.get="/api/note/list")
}
```

### 6.2 IDL 关键点说明

✅ **字段标注**：
- `agw.js_conv="str"` - 将 int64 转换为字符串（避免 JS 精度问题）
- `api.js_conv="true"` - API 网关转换标志
- `api.path="note_id"` - 路径参数

✅ **响应格式统一**：
```thrift
253: required i32 code
254: required string msg
```

---

## Step 7: 生成 API Handler

### 7.1 生成代码

```bash
# 进入项目根目录

# 生成 API 代码（假设项目有相应的生成脚本）
make gen_api
```

生成后将得到：
- `backend/api/model/note/*.go` - API 模型
- `backend/api/handler/coze/note_service.go` - API Handler

### 7.2 实现 Handler

编辑生成的文件：`backend/api/handler/coze/note_service.go`

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

// GetNoteDetail .
// @router /note/:note_id [GET]
func GetNoteDetail(ctx context.Context, c *app.RequestContext) {
	var req note.GetNoteDetailRequest
	err := c.BindAndValidate(&req)
	if err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}

	resp, err := noteApp.NoteApplicationSVC.GetNoteDetail(ctx, &req)
	if err != nil {
		internalServerErrorResponse(ctx, c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateNote .
// @router /note/update [POST]
func UpdateNote(ctx context.Context, c *app.RequestContext) {
	var req note.UpdateNoteRequest
	err := c.BindAndValidate(&req)
	if err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}

	resp, err := noteApp.NoteApplicationSVC.UpdateNote(ctx, &req)
	if err != nil {
		internalServerErrorResponse(ctx, c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteNote .
// @router /note/:note_id [DELETE]
func DeleteNote(ctx context.Context, c *app.RequestContext) {
	var req note.DeleteNoteRequest
	err := c.BindAndValidate(&req)
	if err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}

	resp, err := noteApp.NoteApplicationSVC.DeleteNote(ctx, &req)
	if err != nil {
		internalServerErrorResponse(ctx, c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListUserNotes .
// @router /note/list [GET]
func ListUserNotes(ctx context.Context, c *app.RequestContext) {
	var req note.ListUserNotesRequest
	err := c.BindAndValidate(&req)
	if err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}

	resp, err := noteApp.NoteApplicationSVC.ListUserNotes(ctx, &req)
	if err != nil {
		internalServerErrorResponse(ctx, c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
```

### 7.3 关键点说明

✅ **Handler 职责**：
1. 绑定和验证请求参数
2. 调用应用服务
3. 返回响应

✅ **错误处理**：
- `invalidParamRequestResponse` - 参数错误
- `internalServerErrorResponse` - 内部错误

---

## Step 8: 测试

### 8.1 启动服务

```bash
# 进入项目根目录

# 启动后端服务
make run_backend
```

### 8.2 API 测试

#### 1. 创建笔记

```bash
curl -X POST http://localhost:8080/api/note/create \
  -H "Content-Type: application/json" \
  -H "Cookie: session_key=YOUR_SESSION_KEY" \
  -d '{
    "space_id": "1",
    "title": "我的第一篇笔记",
    "content": "这是笔记内容"
  }'
```

预期响应：
```json
{
  "code": 0,
  "msg": "",
  "data": {
    "note_id": "7563957783431741441",
    "user_id": "1",
    "space_id": "1",
    "title": "我的第一篇笔记",
    "content": "这是笔记内容",
    "created_at": 1703123456789,
    "updated_at": 1703123456789
  }
}
```

#### 2. 获取笔记详情

```bash
curl -X GET http://localhost:8080/api/note/7563957783431741441 \
  -H "Cookie: session_key=YOUR_SESSION_KEY"
```

#### 3. 更新笔记

```bash
curl -X POST http://localhost:8080/api/note/update \
  -H "Content-Type: application/json" \
  -H "Cookie: session_key=YOUR_SESSION_KEY" \
  -d '{
    "note_id": "7563957783431741441",
    "title": "更新后的标题",
    "content": "更新后的内容"
  }'
```

#### 4. 获取笔记列表

```bash
curl -X GET "http://localhost:8080/api/note/list?offset=0&limit=20" \
  -H "Cookie: session_key=YOUR_SESSION_KEY"
```

#### 5. 删除笔记

```bash
curl -X DELETE http://localhost:8080/api/note/7563957783431741441 \
  -H "Cookie: session_key=YOUR_SESSION_KEY"
```

### 8.3 单元测试

创建文件：`backend/domain/note/service/impl/note_test.go`

```go
package impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/coze-dev/coze-studio/backend/domain/note/entity"
	"github.com/coze-dev/coze-studio/backend/domain/note/service"
)

// MockNoteRepository mock 仓储
type MockNoteRepository struct {
	mock.Mock
}

func (m *MockNoteRepository) Create(ctx context.Context, note *entity.Note) error {
	args := m.Called(ctx, note)
	return args.Error(0)
}

func (m *MockNoteRepository) GetByID(ctx context.Context, noteID int64) (*entity.Note, bool, error) {
	args := m.Called(ctx, noteID)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*entity.Note), args.Bool(1), args.Error(2)
}

// MockIDGenerator mock ID 生成器
type MockIDGenerator struct {
	mock.Mock
}

func (m *MockIDGenerator) GenerateID() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func TestNoteDomainService_Create(t *testing.T) {
	ctx := context.Background()

	// 创建 mock 对象
	mockRepo := new(MockNoteRepository)
	mockIDGen := new(MockIDGenerator)

	// 设置期望
	mockIDGen.On("GenerateID").Return(int64(123456), nil)
	mockRepo.On("Create", ctx, mock.Anything).Return(nil)

	// 创建服务
	svc := NewNoteDomain(ctx, &Components{
		IDGen:    mockIDGen,
		NoteRepo: mockRepo,
	})

	// 执行测试
	req := &service.CreateNoteRequest{
		UserID:  1,
		SpaceID: 1,
		Title:   "测试笔记",
		Content: "这是内容",
	}
	note, err := svc.Create(ctx, req)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, note)
	assert.Equal(t, int64(123456), note.ID)
	assert.Equal(t, "测试笔记", note.Title)

	// 验证 mock 调用
	mockIDGen.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestNoteDomainService_GetByID(t *testing.T) {
	ctx := context.Background()

	// 创建 mock 对象
	mockRepo := new(MockNoteRepository)
	mockIDGen := new(MockIDGenerator)

	// 准备测试数据
	expectedNote := &entity.Note{
		ID:      123456,
		UserID:  1,
		Title:   "测试笔记",
		Content: "这是内容",
	}

	// 设置期望
	mockRepo.On("GetByID", ctx, int64(123456)).Return(expectedNote, true, nil)

	// 创建服务
	svc := NewNoteDomain(ctx, &Components{
		IDGen:    mockIDGen,
		NoteRepo: mockRepo,
	})

	// 执行测试
	note, err := svc.GetByID(ctx, 123456, 1)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, note)
	assert.Equal(t, expectedNote.ID, note.ID)
	assert.Equal(t, expectedNote.Title, note.Title)

	// 验证 mock 调用
	mockRepo.AssertExpectations(t)
}
```

运行测试：

```bash
cd backend/domain/note/service/impl
go test -v
```

---

## 常见问题

### Q1: GORM Gen 生成的代码在哪里？

**A**: 生成的代码位于：
- `backend/domain/note/internal/dal/model/` - 数据库模型
- `backend/domain/note/internal/dal/query/` - 查询代码

### Q2: 为什么使用全局变量 `NoteApplicationSVC`？

**A**: 这是项目的统一模式，有以下优点：
- 单例模式，避免重复创建
- 全局访问，方便在 Handler 中调用
- 在 `Init` 函数中统一初始化

参考 `backend/application/user/user.go`：
```go
var UserApplicationSVC = &UserApplicationService{}
```

### Q3: IDL 如何生成 Go 代码？

**A**: 项目使用 Thrift IDL 自动生成：
1. 定义 `.thrift` 文件
2. 运行生成命令（通常是 `make gen_api`）
3. 自动生成 API 模型和路由注册代码

### Q4: 如何调试 API？

**A**: 推荐使用以下工具：
- **curl** - 命令行测试
- **Postman** - 图形化测试
- **查看日志** - `logs/` 目录

### Q5: 数据库迁移如何回滚？

**A**: 
```bash
# 查看迁移历史
make db_migrate_status

# 回滚到指定版本
atlas migrate down --url "mysql://user:pass@localhost:3306/opencoze" --to VERSION
```

### Q6: 如何添加权限控制？

**A**: 在 Domain Service 中检查：
```go
func (s *NoteDomainService) GetByID(ctx context.Context, noteID, userID int64) (*entity.Note, error) {
    note, exist, err := s.noteRepo.GetByID(ctx, noteID)
    // ...
    
    // 权限检查
    if note.UserID != userID {
        return nil, errorx.New(errno.ErrPermissionDeniedCode)
    }
    
    return note, nil
}
```

---

## 🎉 总结

通过这个实战练习，你已经学会了：

1. ✅ **数据库设计** - 遵循项目的 MySQL 规则
2. ✅ **GORM Gen** - 自动生成类型安全的查询代码
3. ✅ **Repository 模式** - 抽象数据访问层
4. ✅ **Domain Service** - 实现业务逻辑
5. ✅ **Application Service** - 协调领域服务和外部调用
6. ✅ **IDL 定义** - 使用 Thrift 定义 API
7. ✅ **API Handler** - 处理 HTTP 请求
8. ✅ **完整测试** - API 测试和单元测试

**下一步建议**：
- 📖 阅读其他领域模块的代码（如 `workflow`、`knowledge`）
- 🔧 为笔记功能添加更多特性（标签、搜索、分享）
- 🧪 编写更全面的单元测试和集成测试
- 📊 添加性能监控和日志

**参考文档**：
- `BACKEND_ERRATA.md` - 了解项目实际模式
- `BACKEND_GORM_GEN_GUIDE.md` - 深入理解 GORM Gen
- `BACKEND_LEARNING_GUIDE.md` - 系统学习架构

---

<div align="center">
  <strong>🚀 Happy Coding! 🚀</strong>
</div>

