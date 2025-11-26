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

// Package repository 定义了对话领域的仓储接口
//
// 本包提供对话数据的持久化操作接口。
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/conversation/conversation/entity"
	"github.com/coze-dev/coze-studio/backend/domain/conversation/conversation/internal/dal"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
)

// NewConversationRepo 创建对话仓储实例
func NewConversationRepo(db *gorm.DB, idGen idgen.IDGenerator) ConversationRepo {
	return dal.NewConversationDAO(db, idGen)
}

// ConversationRepo 对话仓储接口
//
// 提供对话数据的 CRUD 操作。
type ConversationRepo interface {
	// Create 创建对话
	Create(ctx context.Context, msg *entity.Conversation) (*entity.Conversation, error)

	// GetByID 根据 ID 获取对话
	GetByID(ctx context.Context, id int64) (*entity.Conversation, error)

	// UpdateSection 更新对话分段
	//
	// 返回新的分段 ID。
	UpdateSection(ctx context.Context, id int64) (int64, error)

	// Get 根据条件获取对话
	Get(ctx context.Context, userID int64, agentID int64, scene int32, connectorID int64) (*entity.Conversation, error)

	// Update 更新对话
	Update(ctx context.Context, req *entity.UpdateMeta) (*entity.Conversation, error)

	// Delete 删除对话
	Delete(ctx context.Context, id int64) (int64, error)

	// List 查询对话列表
	List(ctx context.Context, req *entity.ListMeta) ([]*entity.Conversation, bool, error)
}
