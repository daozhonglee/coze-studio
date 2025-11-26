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

// conversation.go 会话实体定义
//
// 本文件定义了 ChatFlow 工作流的会话相关实体：
//   - ConversationTemplate: 会话模板
//   - StaticConversation: 静态会话（基于模板）
//   - DynamicConversation: 动态会话（用户自定义）

package entity

// ConversationTemplate 会话模板
// 定义会话的基础配置，用于创建静态会话
type ConversationTemplate struct {
	SpaceID    int64
	AppID      int64
	Name       string
	TemplateID int64
}

// StaticConversation 静态会话
// 基于模板创建的固定会话，关联用户和连接器
type StaticConversation struct {
	UserID         int64
	ConnectorID    int64
	TemplateID     int64
	ConversationID int64
}

// DynamicConversation 动态会话
// 用户自定义创建的会话，支持命名
type DynamicConversation struct {
	ID             int64
	UserID         int64
	ConnectorID    int64
	ConversationID int64
	Name           string
}
