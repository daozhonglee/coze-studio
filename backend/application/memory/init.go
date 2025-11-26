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

// Package memory 定义了记忆(Memory)应用层服务
//
// 本包提供记忆相关的应用层业务逻辑，包括：
// - 变量管理（Agent 运行时变量）
// - 数据库管理（Agent 关联的结构化数据）
// - RDB 服务（关系数据库操作）
//
// 记忆是 AI Agent 保持上下文和状态的重要机制。
package memory

import (
	"gorm.io/gorm"

	database "github.com/coze-dev/coze-studio/backend/domain/memory/database/service"
	"github.com/coze-dev/coze-studio/backend/domain/memory/variables/repository"
	variables "github.com/coze-dev/coze-studio/backend/domain/memory/variables/service"
	search "github.com/coze-dev/coze-studio/backend/domain/search/service"
	"github.com/coze-dev/coze-studio/backend/infra/cache"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
	"github.com/coze-dev/coze-studio/backend/infra/rdb"
	rdbService "github.com/coze-dev/coze-studio/backend/infra/rdb/impl/rdb"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
)

// MemoryApplicationServices 记忆应用服务集合
//
// 包含变量和数据库领域服务
type MemoryApplicationServices struct {
	// VariablesDomainSVC 变量领域服务
	VariablesDomainSVC variables.Variables
	// DatabaseDomainSVC 数据库领域服务
	DatabaseDomainSVC database.Database
	// RDBDomainSVC RDB 领域服务
	RDBDomainSVC rdb.RDB
}

// ServiceComponents 记忆应用服务依赖组件
type ServiceComponents struct {
	// IDGen ID 生成器
	IDGen idgen.IDGenerator
	// DB 数据库连接
	DB *gorm.DB
	// EventBus 资源事件总线
	EventBus search.ResourceEventBus
	// TosClient 对象存储客户端
	TosClient storage.Storage
	// ResourceDomainNotifier 资源领域通知器
	ResourceDomainNotifier search.ResourceEventBus
	// CacheCli 缓存客户端
	CacheCli cache.Cmdable
}

// InitService 初始化记忆应用服务
//
// 创建变量和数据库领域服务
func InitService(c *ServiceComponents) *MemoryApplicationServices {
	repo := repository.NewVariableRepo(c.DB, c.IDGen)
	variablesDomainSVC := variables.NewService(repo)
	rdbSVC := rdbService.NewService(c.DB, c.IDGen)
	databaseDomainSVC := database.NewService(rdbSVC, c.DB, c.IDGen, c.TosClient, c.CacheCli)

	VariableApplicationSVC.DomainSVC = variablesDomainSVC
	DatabaseApplicationSVC.DomainSVC = databaseDomainSVC
	DatabaseApplicationSVC.eventbus = c.ResourceDomainNotifier

	return &MemoryApplicationServices{
		VariablesDomainSVC: variablesDomainSVC,
		DatabaseDomainSVC:  databaseDomainSVC,
		RDBDomainSVC:       rdbSVC,
	}
}
