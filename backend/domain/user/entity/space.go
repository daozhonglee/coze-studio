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

// Package entity 定义了用户(User)领域的核心实体（工作空间相关）

package entity

// SpaceType 工作空间类型
type SpaceType int32

const (
	// SpaceTypePersonal 个人空间
	SpaceTypePersonal SpaceType = 1
	// SpaceTypeTeam 团队空间
	SpaceTypeTeam SpaceType = 2
)

// Space 工作空间实体
//
// 工作空间是资源的容器，用于组织和管理应用、工作流等资源。
// 支持个人空间和团队空间两种类型。
type Space struct {
	// ID 空间唯一标识
	ID int64
	// Name 空间名称
	Name string
	// Description 空间描述
	Description string
	// IconURL 空间图标 URL
	IconURL string
	// SpaceType 空间类型（个人/团队）
	SpaceType SpaceType
	// OwnerID 空间所有者用户ID
	OwnerID int64
	// CreatorID 空间创建者用户ID
	CreatorID int64
	// CreatedAt 创建时间（毫秒时间戳）
	CreatedAt int64
	// UpdatedAt 更新时间（毫秒时间戳）
	UpdatedAt int64
}
