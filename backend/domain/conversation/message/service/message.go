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

// Package message 定义了消息领域的服务接口
//
// 本包提供消息管理的核心能力：
// - 创建、查询、编辑、删除消息
// - 消息状态管理（中断标记等）
//
// 设计说明：
// 消息是对话中的基本交互单元，每条消息关联到一个运行记录。
// 支持流式创建（PreCreate + Create）和批量创建。
package message

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/conversation/message/entity"
)

// Message 消息服务接口
//
// 该接口定义了消息管理的所有操作。
type Message interface {
	// List 查询消息列表
	//
	// 返回带游标的分页结果。
	List(ctx context.Context, req *entity.ListMeta) (*entity.ListResult, error)

	// ListWithoutPair 查询消息列表（不配对）
	//
	// 与 List 类似，但不进行消息配对处理。
	ListWithoutPair(ctx context.Context, req *entity.ListMeta) (*entity.ListResult, error)

	// PreCreate 预创建消息
	//
	// 在流式输出开始前创建消息占位，返回分配的消息 ID。
	PreCreate(ctx context.Context, req *entity.Message) (*entity.Message, error)

	// Create 创建消息
	Create(ctx context.Context, req *entity.Message) (*entity.Message, error)

	// BatchCreate 批量创建消息
	BatchCreate(ctx context.Context, req []*entity.Message) ([]*entity.Message, error)

	// GetByRunIDs 根据运行记录 ID 查询消息
	GetByRunIDs(ctx context.Context, conversationID int64, runIDs []int64) ([]*entity.Message, error)

	// GetByID 根据 ID 获取消息
	GetByID(ctx context.Context, id int64) (*entity.Message, error)

	// Edit 编辑消息
	Edit(ctx context.Context, req *entity.Message) (*entity.Message, error)

	// Delete 删除消息
	Delete(ctx context.Context, req *entity.DeleteMeta) error

	// Broken 标记消息为中断状态
	Broken(ctx context.Context, req *entity.BrokenMeta) error
}
