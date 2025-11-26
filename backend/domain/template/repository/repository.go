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

// Package repository 定义了模板(Template)领域的仓储接口
//
// 本包提供模板数据的持久化操作抽象
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/template/internal/dal"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"

	"github.com/coze-dev/coze-studio/backend/domain/template/entity"

	"github.com/coze-dev/coze-studio/backend/domain/template/internal/dal/model"
)

// NewTemplateDAO 创建模板仓储实例
func NewTemplateDAO(db *gorm.DB, idGen idgen.IDGenerator) TemplateRepository {
	return dal.NewTemplateDAO(db, idGen)
}

// TemplateRepository 模板仓储接口
//
// 定义模板的数据访问方法
type TemplateRepository interface {
	// Create 创建模板
	Create(ctx context.Context, template *model.Template) (int64, error)

	// List 列出模板（带过滤和分页）
	List(ctx context.Context, filter *entity.TemplateFilter, page *entity.Pagination, orderByField string) ([]*model.Template, int64, error)
}
