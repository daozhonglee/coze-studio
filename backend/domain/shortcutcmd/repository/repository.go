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

// Package repository 定义了快捷指令(ShortcutCmd)领域的仓储接口
//
// 本包提供快捷指令数据的持久化操作抽象
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/shortcutcmd/entity"
	"github.com/coze-dev/coze-studio/backend/domain/shortcutcmd/internal/dal"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
)

// NewShortCutCmdRepo 创建快捷指令仓储实例
func NewShortCutCmdRepo(db *gorm.DB, idGen idgen.IDGenerator) ShortCutCmdRepo {
	return dal.NewShortCutCmdDAO(db, idGen)
}

// ShortCutCmdRepo 快捷指令仓储接口
//
// 定义快捷指令的数据访问方法
type ShortCutCmdRepo interface {
	// List 列出快捷指令
	List(ctx context.Context, lm *entity.ListMeta) ([]*entity.ShortcutCmd, error)
	// Create 创建快捷指令
	Create(ctx context.Context, shortcut *entity.ShortcutCmd) (*entity.ShortcutCmd, error)
	// Update 更新快捷指令
	Update(ctx context.Context, shortcut *entity.ShortcutCmd) (*entity.ShortcutCmd, error)
	// GetByCmdID 根据快捷指令ID和状态获取
	GetByCmdID(ctx context.Context, cmdID int64, isOnline int32) (*entity.ShortcutCmd, error)
	// PublishCMDs 发布指定的快捷指令
	PublishCMDs(ctx context.Context, objID int64, cmdIDs []int64) error
}
