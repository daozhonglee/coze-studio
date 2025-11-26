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

// Package repository 定义了数据库记忆领域的仓储接口
//
// 本包提供数据库表元数据的持久化操作接口：
// - DraftDAO: 草稿表元数据访问
// - OnlineDAO: 上线表元数据访问
// - AgentToDatabaseDAO: Agent 与数据库绑定关系管理
//
// 设计说明：
// 仓储层仅负责数据库表的元数据（表名、字段定义等）持久化，
// 实际的数据记录存储在动态创建的物理表中，通过 RDB 接口操作。
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/api/model/data/database/table"
	"github.com/coze-dev/coze-studio/backend/crossdomain/database/model"
	"github.com/coze-dev/coze-studio/backend/domain/memory/database/entity"
	"github.com/coze-dev/coze-studio/backend/domain/memory/database/internal/dal"
	"github.com/coze-dev/coze-studio/backend/domain/memory/database/internal/dal/query"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
)

// NewAgentToDatabaseDAO 创建 Agent-数据库绑定关系 DAO 实例
func NewAgentToDatabaseDAO(db *gorm.DB, idGen idgen.IDGenerator) AgentToDatabaseDAO {
	return dal.NewAgentToDatabaseDAO(db, idGen)
}

// AgentToDatabaseDAO Agent 与数据库表绑定关系的数据访问接口
//
// 管理 Agent 可以访问哪些数据库表的关联关系。
type AgentToDatabaseDAO interface {
	// BatchCreate 批量创建绑定关系
	BatchCreate(ctx context.Context, relations []*model.AgentToDatabase) ([]int64, error)

	// BatchDelete 批量删除绑定关系
	BatchDelete(ctx context.Context, basicRelations []*model.AgentToDatabaseBasic) error

	// ListByAgentID 查询 Agent 绑定的所有数据库表
	ListByAgentID(ctx context.Context, agentID int64, tableType table.TableType) ([]*model.AgentToDatabase, error)
}

// NewDraftDatabaseDAO 创建草稿表 DAO 实例
func NewDraftDatabaseDAO(db *gorm.DB, idGen idgen.IDGenerator) DraftDAO {
	return dal.NewDraftDatabaseDAO(db, idGen)
}

// DraftDAO 草稿数据库表的数据访问接口
//
// 草稿表用于 Agent 开发阶段的数据存储，
// 在 Agent 发布时会同步到对应的上线表。
type DraftDAO interface {
	// Get 根据 ID 获取草稿表信息
	Get(ctx context.Context, id int64) (*entity.Database, error)

	// List 分页查询草稿表列表
	List(ctx context.Context, filter *entity.DatabaseFilter, page *entity.Pagination, orderBy []*model.OrderBy) ([]*entity.Database, int64, error)

	// MGet 批量获取草稿表信息
	MGet(ctx context.Context, ids []int64) ([]*entity.Database, error)

	// CreateWithTX 在事务中创建草稿表记录
	CreateWithTX(ctx context.Context, tx *query.QueryTx, database *entity.Database, draftID, onlineID int64, physicalTableName string) (*entity.Database, error)

	// UpdateWithTX 在事务中更新草稿表记录
	UpdateWithTX(ctx context.Context, tx *query.QueryTx, database *entity.Database) (*entity.Database, error)

	// DeleteWithTX 在事务中删除草稿表记录
	DeleteWithTX(ctx context.Context, tx *query.QueryTx, id int64) error

	// BatchDeleteWithTX 在事务中批量删除草稿表记录
	BatchDeleteWithTX(ctx context.Context, tx *query.QueryTx, ids []int64) error
}

// NewOnlineDatabaseDAO 创建上线表 DAO 实例
func NewOnlineDatabaseDAO(db *gorm.DB, idGen idgen.IDGenerator) OnlineDAO {
	return dal.NewOnlineDatabaseDAO(db, idGen)
}

// OnlineDAO 上线数据库表的数据访问接口
//
// 上线表用于 Agent 发布后的生产环境数据存储，
// 与草稿表结构一致但数据独立。
type OnlineDAO interface {
	// Get 根据 ID 获取上线表信息
	Get(ctx context.Context, id int64) (*entity.Database, error)

	// MGet 批量获取上线表信息
	MGet(ctx context.Context, ids []int64) ([]*entity.Database, error)

	// List 分页查询上线表列表
	List(ctx context.Context, filter *entity.DatabaseFilter, page *entity.Pagination, orderBy []*model.OrderBy) ([]*entity.Database, int64, error)

	// UpdateWithTX 在事务中更新上线表记录
	UpdateWithTX(ctx context.Context, tx *query.QueryTx, database *entity.Database) (*entity.Database, error)

	// CreateWithTX 在事务中创建上线表记录
	CreateWithTX(ctx context.Context, tx *query.QueryTx, database *entity.Database, draftID, onlineID int64, physicalTableName string) (*entity.Database, error)

	// DeleteWithTX 在事务中删除上线表记录
	DeleteWithTX(ctx context.Context, tx *query.QueryTx, id int64) error

	// BatchDeleteWithTX 在事务中批量删除上线表记录
	BatchDeleteWithTX(ctx context.Context, tx *query.QueryTx, ids []int64) error
}
