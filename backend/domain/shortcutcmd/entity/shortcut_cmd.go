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

// Package entity 定义了快捷指令(ShortcutCmd)领域的核心实体
//
// 本包包含快捷指令相关的领域实体：
// - ShortcutCmd: 快捷指令实体，定义用户自定义的快捷操作
// - ListMeta: 查询元数据，用于过滤快捷指令列表
//
// 快捷指令允许用户定义预设的对话输入，方便快速触发特定功能。
package entity

import "github.com/coze-dev/coze-studio/backend/domain/shortcutcmd/internal/dal/model"

// ShortcutCmd 快捷指令实体
//
// 复用自 DAL 层的 model.ShortcutCommand，表示一个快捷指令定义
type ShortcutCmd = model.ShortcutCommand

// ListMeta 快捷指令查询元数据
//
// 用于指定查询条件，过滤快捷指令列表
type ListMeta struct {
	// ObjectID 关联对象ID（如 Agent ID）
	ObjectID int64 `json:"object_id"`
	// SpaceID 所属工作空间ID
	SpaceID int64 `json:"space_id"`
	// IsOnline 是否在线状态（0: 草稿, 1: 已发布）
	IsOnline int32 `json:"is_online"`
	// CommandIDs 指定查询的快捷指令ID列表
	CommandIDs []int64 `json:"command_ids"`
}
