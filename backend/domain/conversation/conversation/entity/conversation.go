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

// Package entity 定义了对话领域的核心实体
//
// 本包包含对话管理相关的实体定义：
// - Conversation: 对话实体
// - CreateMeta: 创建对话的参数
// - ListMeta: 查询对话列表的参数
//
// 设计说明：
// 对话是用户与 Agent 交互的容器，包含多个运行记录和消息。
// 对话可以分段（Section），每个分段代表一次连续的对话上下文。
package entity

import (
	"github.com/coze-dev/coze-studio/backend/api/model/conversation/common"
	"github.com/coze-dev/coze-studio/backend/crossdomain/conversation/model"
)

// Conversation 对话实体，复用 crossdomain 中的定义
type Conversation = model.Conversation

// CreateMeta 创建对话的请求参数
type CreateMeta struct {
	Name        string       `json:"name"`
	AgentID     int64        `json:"agent_id"`
	UserID      *string      `json:"user_id"`
	CreatorID   int64        `json:"creator_id"`
	ConnectorID int64        `json:"connector_id"`
	Scene       common.Scene `json:"scene"`
	Ext         string       `json:"ext"`
}

// NewConversationCtxRequest 创建新对话上下文的请求参数
type NewConversationCtxRequest struct {
	// ID 对话 ID
	ID int64 `json:"id"`
}

// NewConversationCtxResponse 创建新对话上下文的响应
type NewConversationCtxResponse struct {
	// ID 对话 ID
	ID int64 `json:"id"`
	// SectionID 新的分段 ID
	SectionID int64 `json:"section_id"`
}

// GetCurrent 获取当前对话的参数，复用 crossdomain 中的定义
type GetCurrent = model.GetCurrent

// ListMeta 对话列表查询参数
type ListMeta struct {
	CreatorID   int64        `json:"creator_id"`
	UserID      *string      `json:"user_id"`
	ConnectorID int64        `json:"connector_id"`
	Scene       common.Scene `json:"scene"`
	AgentID     int64        `json:"agent_id"`
	Limit       int          `json:"limit"`
	Page        int          `json:"page"`
	OrderBy     *string      `json:"order_by"`
}

// UpdateMeta 更新对话的请求参数
type UpdateMeta struct {
	// ID 对话 ID
	ID int64 `json:"id"`
	// Name 新的对话名称
	Name string `json:"name"`
}
