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

// Package entity 定义了提示词(Prompt)领域的核心实体
//
// 本包包含提示词相关的领域实体：
// - PromptResource: 提示词资源，表示一个可复用的提示词模板
//
// 提示词是 AI Agent 的核心配置，用于指导模型的行为和输出格式。
package entity

// PromptResource 提示词资源实体
//
// 表示一个可复用的提示词模板，可以在工作空间中创建和管理。
// 提示词可以包含变量槽位(InputSlot)，在使用时进行填充。
type PromptResource struct {
	// ID 提示词唯一标识
	ID int64
	// SpaceID 所属工作空间ID
	SpaceID int64
	// Name 提示词名称
	Name string
	// Description 提示词描述
	Description string
	// PromptText 提示词内容文本
	PromptText string
	// Status 状态（1:有效）
	Status int32
	// CreatorID 创建者用户ID
	CreatorID int64
	// CreatedAt 创建时间（毫秒时间戳）
	CreatedAt int64
	// UpdatedAt 更新时间（毫秒时间戳）
	UpdatedAt int64
}
