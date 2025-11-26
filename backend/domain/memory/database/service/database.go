/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package service 定义了数据库记忆领域的服务接口和请求/响应类型
//
// 本包提供 Agent 数据库表的完整生命周期管理能力：
// - 表的 CRUD 操作（创建、读取、更新、删除）
// - 记录级别的增删改查
// - Excel/CSV 文件导入
// - SQL 查询执行
// - Agent 与数据库的绑定关系管理
//
// 设计说明：
// 采用草稿/上线双表机制，支持 Agent 发布时数据同步。
// 每个逻辑表对应两个物理表：draft（草稿）和 online（上线）。
package service

import (
	"context"

	database "github.com/coze-dev/coze-studio/backend/crossdomain/database/model"

	"github.com/coze-dev/coze-studio/backend/api/model/data/database/table"
	"github.com/coze-dev/coze-studio/backend/api/model/data/knowledge"
	"github.com/coze-dev/coze-studio/backend/domain/memory/database/entity"
)

// Database 数据库记忆服务接口
//
// 该接口定义了 Agent 数据库表管理的所有操作，包括：
// - 表管理：创建、更新、删除、查询数据库表
// - 记录操作：增删改查表中的数据记录
// - 文件导入：从 Excel/CSV 导入数据
// - SQL 执行：执行自定义 SQL 查询
// - 绑定管理：Agent 与数据库表的关联关系
//
//go:generate mockgen -destination  ../../../../internal/mock/domain/memory/database/database_mock.go  --package database  -source database.go
type Database interface {
	// CreateDatabase 创建数据库表
	// 同时创建 draft 和 online 两个物理表
	CreateDatabase(ctx context.Context, req *CreateDatabaseRequest) (*CreateDatabaseResponse, error)

	// UpdateDatabase 更新数据库表结构
	// 支持添加、修改、删除字段
	UpdateDatabase(ctx context.Context, req *UpdateDatabaseRequest) (*UpdateDatabaseResponse, error)

	// DeleteDatabase 删除数据库表
	// 同时删除 draft 和 online 物理表
	DeleteDatabase(ctx context.Context, req *DeleteDatabaseRequest) error

	// MGetDatabase 批量获取数据库表信息
	MGetDatabase(ctx context.Context, req *MGetDatabaseRequest) (*MGetDatabaseResponse, error)

	// ListDatabase 分页查询数据库表列表
	ListDatabase(ctx context.Context, req *ListDatabaseRequest) (*ListDatabaseResponse, error)

	// GetDraftDatabaseByOnlineID 根据上线表 ID 获取草稿表
	GetDraftDatabaseByOnlineID(ctx context.Context, req *GetDraftDatabaseByOnlineIDRequest) (*GetDraftDatabaseByOnlineIDResponse, error)

	// DeleteDatabaseByAppID 删除应用下的所有数据库表
	DeleteDatabaseByAppID(ctx context.Context, req *DeleteDatabaseByAppIDRequest) (*DeleteDatabaseByAppIDResponse, error)

	// GetAllDatabaseByAppID 获取应用下的所有数据库表
	GetAllDatabaseByAppID(ctx context.Context, req *GetAllDatabaseByAppIDRequest) (*GetAllDatabaseByAppIDResponse, error)

	// GetDatabaseTemplate 生成数据库导入模板文件
	// 根据字段定义生成 Excel 模板供用户下载填写
	GetDatabaseTemplate(ctx context.Context, req *GetDatabaseTemplateRequest) (*GetDatabaseTemplateResponse, error)

	// GetDatabaseTableSchema 解析上传文件的表结构
	// 从 Excel/CSV 文件中提取列定义和预览数据
	GetDatabaseTableSchema(ctx context.Context, req *GetDatabaseTableSchemaRequest) (*GetDatabaseTableSchemaResponse, error)

	// ValidateDatabaseTableSchema 验证上传文件与表结构是否匹配
	ValidateDatabaseTableSchema(ctx context.Context, req *ValidateDatabaseTableSchemaRequest) (*ValidateDatabaseTableSchemaResponse, error)

	// SubmitDatabaseInsertTask 提交异步数据导入任务
	// 后台批量导入 Excel/CSV 数据
	SubmitDatabaseInsertTask(ctx context.Context, req *SubmitDatabaseInsertTaskRequest) error

	// GetDatabaseFileProgressData 获取文件导入进度
	GetDatabaseFileProgressData(ctx context.Context, req *GetDatabaseFileProgressDataRequest) (*GetDatabaseFileProgressDataResponse, error)

	// AddDatabaseRecord 添加数据记录
	AddDatabaseRecord(ctx context.Context, req *AddDatabaseRecordRequest) error

	// UpdateDatabaseRecord 更新数据记录
	UpdateDatabaseRecord(ctx context.Context, req *UpdateDatabaseRecordRequest) error

	// DeleteDatabaseRecord 删除数据记录
	DeleteDatabaseRecord(ctx context.Context, req *DeleteDatabaseRecordRequest) error

	// ListDatabaseRecord 分页查询数据记录
	ListDatabaseRecord(ctx context.Context, req *ListDatabaseRecordRequest) (*ListDatabaseRecordResponse, error)

	// ExecuteSQL 执行 SQL 语句
	// 支持 SELECT/INSERT/UPDATE/DELETE 操作
	ExecuteSQL(ctx context.Context, req *ExecuteSQLRequest) (*ExecuteSQLResponse, error)

	// BindDatabase 绑定数据库表到 Agent
	BindDatabase(ctx context.Context, req *BindDatabaseToAgentRequest) error

	// UnBindDatabase 解除数据库表与 Agent 的绑定
	UnBindDatabase(ctx context.Context, req *UnBindDatabaseToAgentRequest) error

	// MGetDatabaseByAgentID 获取 Agent 绑定的所有数据库表
	MGetDatabaseByAgentID(ctx context.Context, req *MGetDatabaseByAgentIDRequest) (*MGetDatabaseByAgentIDResponse, error)

	// MGetRelationsByAgentID 获取 Agent 与数据库表的绑定关系
	MGetRelationsByAgentID(ctx context.Context, req *MGetRelationsByAgentIDRequest) (*MGetRelationsByAgentIDResponse, error)

	// PublishDatabase 发布数据库表
	// 将草稿表数据同步到上线表
	PublishDatabase(ctx context.Context, req *PublishDatabaseRequest) (*PublishDatabaseResponse, error)
}

type CreateDatabaseRequest struct {
	Database *entity.Database
}

type CreateDatabaseResponse struct {
	Database *entity.Database
}

type UpdateDatabaseRequest struct {
	Database *entity.Database
}

type UpdateDatabaseResponse struct {
	Database *entity.Database
}

type MGetDatabaseRequest = database.MGetDatabaseRequest
type MGetDatabaseResponse = database.MGetDatabaseResponse
type GetDatabaseTemplateRequest struct {
	UserID     int64
	TableName  string
	FieldItems []*table.FieldItem
}

type GetDatabaseTemplateResponse struct {
	Url string
}
type ListDatabaseRequest struct {
	CreatorID   *int64
	SpaceID     *int64
	ConnectorID *int64
	TableName   *string
	AppID       int64
	TableType   table.TableType
	OrderBy     []*database.OrderBy

	Limit  int
	Offset int
}

type ListDatabaseResponse struct {
	Databases []*entity.Database

	HasMore    bool
	TotalCount int64
}

type AddDatabaseRecordRequest struct {
	DatabaseID  int64
	TableType   table.TableType
	ConnectorID *int64
	UserID      int64
	Records     []map[string]string
}

type UpdateDatabaseRecordRequest struct {
	DatabaseID  int64
	TableType   table.TableType
	ConnectorID *int64
	UserID      int64
	Records     []map[string]string
}

type DeleteDatabaseRecordRequest struct {
	DatabaseID  int64
	TableType   table.TableType
	ConnectorID *int64
	UserID      int64
	Records     []map[string]string
}

type ListDatabaseRecordRequest struct {
	DatabaseID  int64
	ConnectorID *int64
	TableType   table.TableType
	UserID      int64

	Limit  int
	Offset int
}

type ListDatabaseRecordResponse struct {
	Records   []map[string]string
	FieldList []*database.FieldItem

	HasMore    bool
	TotalCount int64
}

type ExecuteSQLRequest = database.ExecuteSQLRequest

type ExecuteSQLResponse = database.ExecuteSQLResponse

type BindDatabaseToAgentRequest = database.BindDatabaseToAgentRequest

type UnBindDatabaseToAgentRequest = database.UnBindDatabaseToAgentRequest

type MGetDatabaseByAgentIDRequest struct {
	AgentID       int64
	TableType     table.TableType
	NeedSysFields bool
}

type MGetDatabaseByAgentIDResponse struct {
	Databases []*entity.Database
}

type PublishDatabaseRequest = database.PublishDatabaseRequest

type PublishDatabaseResponse = database.PublishDatabaseResponse

type DeleteDatabaseRequest = database.DeleteDatabaseRequest

type MGetRelationsByAgentIDRequest struct {
	AgentID       int64
	TableType     table.TableType
	NeedSysFields bool
}

type MGetRelationsByAgentIDResponse struct {
	Relations []*database.AgentToDatabase
}
type GetDatabaseTableSchemaRequest struct {
	TableSheet    entity.TableSheet
	TableDataType table.TableDataType
	DatabaseID    int64
	TosURL        string
	UserID        int64
}

type GetDatabaseTableSchemaResponse struct {
	SheetList   []*knowledge.DocTableSheet
	TableMeta   []*knowledge.DocTableColumn
	PreviewData []map[int64]string
}

type ValidateDatabaseTableSchemaRequest struct {
	TableSheet    entity.TableSheet
	TableDataType table.TableDataType
	DatabaseID    int64
	TosURL        string
	UserID        int64
	Fields        []*database.FieldItem
}

type ValidateDatabaseTableSchemaResponse struct {
	Valid      bool
	InvalidMsg *string // if valid is false, it will be set
}

func (r *ValidateDatabaseTableSchemaResponse) GetInvalidMsg() string {
	if r.Valid || r.InvalidMsg == nil {
		return ""
	}

	return *r.InvalidMsg
}

type SubmitDatabaseInsertTaskRequest struct {
	DatabaseID  int64
	FileURI     string
	TableType   table.TableType
	TableSheet  entity.TableSheet
	ConnectorID *int64
	UserID      int64
}

type GetDatabaseFileProgressDataRequest struct {
	DatabaseID int64
	TableType  table.TableType
	UserID     int64
}

type GetDatabaseFileProgressDataResponse struct {
	FileName       string
	Progress       int32
	StatusDescript *string
}

type GetDraftDatabaseByOnlineIDRequest struct {
	OnlineID int64
}

type GetDraftDatabaseByOnlineIDResponse struct {
	Database *entity.Database
}

type DeleteDatabaseByAppIDRequest struct {
	AppID int64
}

type DeleteDatabaseByAppIDResponse struct {
	DeletedDatabaseIDs []int64 //online database ids
}

type GetAllDatabaseByAppIDRequest = database.GetAllDatabaseByAppIDRequest

type GetAllDatabaseByAppIDResponse = database.GetAllDatabaseByAppIDResponse
