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

// SingleAgentPublish Agent 发布记录
//
// 记录 Agent 的每次发布历史，包括发布的渠道、版本、状态等。
type SingleAgentPublish struct {
	// ID 发布记录 ID
	ID int64
	// AgentID Agent ID
	AgentID int64
	// PublishID 发布 ID（唯一标识本次发布）
	PublishID string
	// ConnectorIds 发布的渠道 ID 列表
	ConnectorIds []int64
	// Version 发布版本号
	Version string
	// PublishResult 发布结果
	PublishResult *string
	// PublishInfo 发布详情
	PublishInfo *string
	// CreatorID 发布者用户 ID
	CreatorID int64
	// PublishTime 发布时间戳
	PublishTime int64
	// CreatedAt 创建时间戳
	CreatedAt int64
	// UpdatedAt 更新时间戳
	UpdatedAt int64
	// Status 发布状态
	Status int32
	// Extra 扩展字段
	Extra *string
}
