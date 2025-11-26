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

// Package conversation 定义了对话领域的服务接口
//
// 本包提供对话管理的核心能力：
// - 创建、查询、更新、删除对话
// - 对话上下文管理（分段）
//
// 设计说明：
// 对话是用户与 Agent 交互的顶层容器。
// 每个对话可以有多个分段，分段用于隔离不同的对话上下文。
package conversation

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/conversation/conversation/entity"
)

// Conversation 对话服务接口
//
// 该接口定义了对话管理的所有操作。
//
//go:generate mockgen -destination ../../../../internal/mock/domain/conversation/conversation/conversation_mock.go --package conversation -source conversation.go
type Conversation interface {
	// Create 创建对话
	Create(ctx context.Context, req *entity.CreateMeta) (*entity.Conversation, error)

	// GetByID 根据 ID 获取对话
	GetByID(ctx context.Context, id int64) (*entity.Conversation, error)

	// NewConversationCtx 创建新的对话上下文（分段）
	NewConversationCtx(ctx context.Context, req *entity.NewConversationCtxRequest) (*entity.NewConversationCtxResponse, error)

	// GetCurrentConversation 获取当前对话
	GetCurrentConversation(ctx context.Context, req *entity.GetCurrent) (*entity.Conversation, error)

	// Delete 删除对话
	Delete(ctx context.Context, id int64) error

	// List 查询对话列表
	//
	// 返回对话列表和是否有更多数据的标志。
	List(ctx context.Context, req *entity.ListMeta) ([]*entity.Conversation, bool, error)

	// Update 更新对话
	Update(ctx context.Context, req *entity.UpdateMeta) (*entity.Conversation, error)
}
