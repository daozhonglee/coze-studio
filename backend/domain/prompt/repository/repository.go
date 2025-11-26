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

// Package repository 定义了提示词(Prompt)领域的仓储接口
//
// 本包提供提示词数据的持久化操作抽象
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/prompt/entity"
	"github.com/coze-dev/coze-studio/backend/domain/prompt/internal/dal"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
)

// NewPromptRepo 创建提示词仓储实例
func NewPromptRepo(db *gorm.DB, generator idgen.IDGenerator) PromptRepository {
	return dal.NewPromptDAO(db, generator)
}

// PromptRepository 提示词仓储接口
//
// 定义提示词资源的数据访问方法
type PromptRepository interface {
	// CreatePromptResource 创建提示词资源
	CreatePromptResource(ctx context.Context, do *entity.PromptResource) (int64, error)
	// GetPromptResource 获取提示词资源
	GetPromptResource(ctx context.Context, promptID int64) (*entity.PromptResource, error)
	// UpdatePromptResource 更新提示词资源
	UpdatePromptResource(ctx context.Context, promptID int64, name, description, promptText *string) error
	// DeletePromptResource 删除提示词资源
	DeletePromptResource(ctx context.Context, ID int64) error
}
