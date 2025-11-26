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

// Package repository 定义了 Agent 运行领域的仓储接口
//
// 本包提供运行记录的持久化操作接口。
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/conversation/agentrun/entity"
	"github.com/coze-dev/coze-studio/backend/domain/conversation/agentrun/internal/dal"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
)

// NewRunRecordRepo 创建运行记录仓储实例
func NewRunRecordRepo(db *gorm.DB, idGen idgen.IDGenerator) RunRecordRepo {

	return dal.NewRunRecordDAO(db, idGen)
}

// RunRecordRepo 运行记录仓储接口
//
// 提供运行记录的 CRUD 操作。
type RunRecordRepo interface {
	// Create 创建运行记录
	Create(ctx context.Context, runMeta *entity.AgentRunMeta) (*entity.RunRecordMeta, error)

	// GetByID 根据 ID 获取运行记录
	GetByID(ctx context.Context, id int64) (*entity.RunRecordMeta, error)

	// Cancel 取消运行记录
	Cancel(ctx context.Context, req *entity.CancelRunMeta) (*entity.RunRecordMeta, error)

	// Delete 批量删除运行记录
	Delete(ctx context.Context, id []int64) error

	// UpdateByID 更新运行记录
	UpdateByID(ctx context.Context, id int64, update *entity.UpdateMeta) error

	// List 查询运行记录列表
	List(ctx context.Context, meta *entity.ListRunRecordMeta) ([]*entity.RunRecordMeta, error)
}
