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

// Package template 定义了模板(Template)应用层服务
//
// 本包提供模板相关的应用层业务逻辑，包括：
// - 模板的创建、更新、删除
// - 官方推荐模板
// - 模板分类管理
//
// 模板为用户提供预配置的 Agent 和工作流模板
package template

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/template/repository"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
)

// ServiceComponents 模板应用服务依赖组件
type ServiceComponents struct {
	// DB 数据库连接
	DB *gorm.DB
	// IDGen ID 生成器
	IDGen idgen.IDGenerator
	// Storage 对象存储服务
	Storage storage.Storage
}

// InitService 初始化模板应用服务
func InitService(ctx context.Context, components *ServiceComponents) *ApplicationService {

	tRepo := repository.NewTemplateDAO(components.DB, components.IDGen)

	ApplicationSVC.templateRepo = tRepo
	ApplicationSVC.storage = components.Storage

	return ApplicationSVC
}
