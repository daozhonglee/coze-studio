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

// Package app 定义了应用(APP/Project)应用层服务
//
// 本包提供应用相关的应用层业务逻辑，包括：
// - 应用的创建、更新、删除、复制
// - 应用发布管理
// - 资源复制和移动
// - 版本管理
//
// 应用是承载多个工作流的容器，支持发布到不同的连接器。
package app

import (
	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/app/repository"
	"github.com/coze-dev/coze-studio/backend/domain/app/service"
	connector "github.com/coze-dev/coze-studio/backend/domain/connector/service"
	variables "github.com/coze-dev/coze-studio/backend/domain/memory/variables/service"
	search "github.com/coze-dev/coze-studio/backend/domain/search/service"
	user "github.com/coze-dev/coze-studio/backend/domain/user/service"
	"github.com/coze-dev/coze-studio/backend/infra/cache"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
)

// ServiceComponents 应用应用服务依赖组件
type ServiceComponents struct {
	// IDGen ID 生成器
	IDGen idgen.IDGenerator
	// DB 数据库连接
	DB *gorm.DB
	// OSS 对象存储服务
	OSS storage.Storage
	// CacheCli 缓存客户端
	CacheCli cache.Cmdable
	// ProjectEventBus 项目事件总线
	ProjectEventBus search.ProjectEventBus

	// UserSVC 用户领域服务
	UserSVC user.User
	// ConnectorSVC 连接器领域服务
	ConnectorSVC connector.Connector
	// VariablesSVC 变量领域服务
	VariablesSVC variables.Variables
}

// InitService 初始化应用应用服务
//
// 创建仓储和领域服务
func InitService(components *ServiceComponents) (*APPApplicationService, error) {
	appRepo := repository.NewAPPRepo(&repository.APPRepoComponents{
		IDGen:    components.IDGen,
		DB:       components.DB,
		CacheCli: components.CacheCli,
	})

	domainComponents := &service.Components{
		IDGen:   components.IDGen,
		DB:      components.DB,
		APPRepo: appRepo,
	}

	domainSVC := service.NewService(domainComponents)

	APPApplicationSVC.DomainSVC = domainSVC
	APPApplicationSVC.appRepo = appRepo

	APPApplicationSVC.oss = components.OSS
	APPApplicationSVC.projectEventBus = components.ProjectEventBus

	APPApplicationSVC.userSVC = components.UserSVC
	APPApplicationSVC.connectorSVC = components.ConnectorSVC
	APPApplicationSVC.variablesSVC = components.VariablesSVC

	return APPApplicationSVC, nil
}
