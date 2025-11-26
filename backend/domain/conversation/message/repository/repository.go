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

// Package repository 定义了消息领域的仓储接口
//
// 本包提供消息数据的持久化操作接口。
package repository

import (
	"context"

	"gorm.io/gorm"

	message "github.com/coze-dev/coze-studio/backend/crossdomain/message/model"
	"github.com/coze-dev/coze-studio/backend/domain/conversation/message/entity"
	"github.com/coze-dev/coze-studio/backend/domain/conversation/message/internal/dal"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
)

// NewMessageRepo 创建消息仓储实例
func NewMessageRepo(db *gorm.DB, idGen idgen.IDGenerator) MessageRepo {
	return dal.NewMessageDAO(db, idGen)
}

// MessageRepo 消息仓储接口
//
// 提供消息数据的 CRUD 操作。
type MessageRepo interface {
	// PreCreate 预创建消息（分配 ID）
	PreCreate(ctx context.Context, msg *entity.Message) (*entity.Message, error)

	// Create 创建消息
	Create(ctx context.Context, msg *entity.Message) (*entity.Message, error)

	// BatchCreate 批量创建消息
	BatchCreate(ctx context.Context, msg []*entity.Message) ([]*entity.Message, error)

	// List 查询消息列表
	List(ctx context.Context, listMeta *entity.ListMeta) ([]*entity.Message, bool, error)

	// GetByRunIDs 根据运行记录 ID 查询消息
	GetByRunIDs(ctx context.Context, runIDs []int64, orderBy string) ([]*entity.Message, error)

	// Edit 编辑消息
	Edit(ctx context.Context, msgID int64, message *message.Message) (int64, error)

	// GetByID 根据 ID 获取消息
	GetByID(ctx context.Context, msgID int64) (*entity.Message, error)

	// Delete 删除消息
	Delete(ctx context.Context, delMeta *entity.DeleteMeta) error
}
