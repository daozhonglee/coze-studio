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

// Package service 定义了快捷指令(ShortcutCmd)领域的服务层接口
//
// 本包提供快捷指令领域的业务服务
package service

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/shortcutcmd/entity"
)

// ShortcutCmd 快捷指令服务接口
//
// 定义快捷指令的业务操作
//
//go:generate mockgen -destination ../../../internal/mock/domain/shortcutcmd/shortcut_cmd_mock.go --package shortcutcmd -source shortcut_cmd.go
type ShortcutCmd interface {
	// ListCMD 列出快捷指令
	ListCMD(ctx context.Context, lm *entity.ListMeta) ([]*entity.ShortcutCmd, error)
	// CreateCMD 创建快捷指令
	CreateCMD(ctx context.Context, shortcut *entity.ShortcutCmd) (*entity.ShortcutCmd, error)
	// UpdateCMD 更新快捷指令
	UpdateCMD(ctx context.Context, shortcut *entity.ShortcutCmd) (*entity.ShortcutCmd, error)
	// GetByCmdID 根据快捷指令ID和状态获取
	GetByCmdID(ctx context.Context, cmdID int64, isOnline int32) (*entity.ShortcutCmd, error)
	// PublishCMDs 发布快捷指令
	// 将指定的快捷指令从草稿状态发布为在线状态
	PublishCMDs(ctx context.Context, objID int64, cmdIDs []int64) error
}
