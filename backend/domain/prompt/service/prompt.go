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

// Package prompt 定义了提示词(Prompt)领域的服务层
//
// 本包提供提示词领域的业务服务，包括：
// - 自定义提示词的 CRUD 操作
// - 官方提示词模板的查询
package prompt

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/prompt/entity"
)

// Prompt 提示词服务接口
//
// 定义提示词领域的业务操作
type Prompt interface {
	// CreatePromptResource 创建提示词资源
	CreatePromptResource(ctx context.Context, p *entity.PromptResource) (int64, error)
	// GetPromptResource 获取提示词资源
	GetPromptResource(ctx context.Context, promptID int64) (*entity.PromptResource, error)
	// UpdatePromptResource 更新提示词资源
	UpdatePromptResource(ctx context.Context, promptID int64, name, description, promptText *string) error
	// DeletePromptResource 删除提示词资源
	DeletePromptResource(ctx context.Context, promptID int64) error

	// ListOfficialPromptResource 列出官方提示词模板
	// 支持按关键词搜索（匹配名称或内容）
	ListOfficialPromptResource(ctx context.Context, keyword string) ([]*entity.PromptResource, error)
}
