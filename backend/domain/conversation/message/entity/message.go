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

package entity

import (
	model "github.com/coze-dev/coze-studio/backend/crossdomain/message/model"
)

// Message 消息实体，复用 crossdomain 中的定义
type Message = model.Message

// ListMeta 消息列表查询参数
type ListMeta struct {
	// ConversationID 对话 ID
	ConversationID int64 `json:"conversation_id"`
	// RunID 运行记录 ID 列表（可选过滤）
	RunID []*int64 `json:"run_id"`
	// UserID 用户 ID
	UserID string `json:"user_id"`
	// AgentID Agent ID
	AgentID int64 `json:"agent_id"`
	// OrderBy 排序方式
	OrderBy *string `json:"order_by"`
	// Limit 返回数量限制
	Limit int `json:"limit"`
	// Cursor 游标（消息 ID）
	Cursor int64 `json:"cursor"`
	// Direction 滚动方向（up/down）
	Direction ScrollPageDirection `json:"direction"`
	// MessageType 消息类型过滤
	MessageType []*model.MessageType `json:"message_type"`
}

// ListResult 消息列表查询结果
type ListResult struct {
	// Messages 消息列表
	Messages []*Message `json:"messages"`
	// PrevCursor 上一页游标
	PrevCursor int64 `json:"prev_cursor"`
	// NextCursor 下一页游标
	NextCursor int64 `json:"next_cursor"`
	// HasMore 是否有更多数据
	HasMore bool `json:"has_more"`
	// Direction 查询方向
	Direction ScrollPageDirection `json:"direction"`
}

// GetByRunIDsRequest 根据运行记录 ID 查询消息的请求参数
type GetByRunIDsRequest struct {
	// ConversationID 对话 ID
	ConversationID int64 `json:"conversation_id"`
	// RunID 运行记录 ID 列表
	RunID []int64 `json:"run_id"`
}

// DeleteMeta 删除消息的请求参数
type DeleteMeta struct {
	// ConversationID 对话 ID（可选）
	ConversationID *int64 `json:"conversation_id"`
	// MessageIDs 要删除的消息 ID 列表
	MessageIDs []int64 `json:"message_ids"`
	// RunIDs 要删除的运行记录 ID 列表（删除关联消息）
	RunIDs []int64 `json:"run_ids"`
}

// BrokenMeta 标记消息中断的请求参数
type BrokenMeta struct {
	// ID 消息 ID
	ID int64 `json:"id"`
	// Position 中断位置
	Position *int32 `json:"position"`
}
