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

// Package singleagent 定义了单 Agent 应用层服务
//
// 本包提供单 Agent 相关的应用层业务逻辑，包括：
// - Agent 的创建、更新、删除
// - Agent 草稿管理
// - Agent 发布和版本管理
// - Agent 配置（插件、工作流、知识库等）
//
// 单 Agent 是最基础的 AI Agent 类型，由单个智能体完成任务。
package singleagent

import (
	"github.com/cloudwego/eino/compose"
	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/agent/singleagent/entity"
	"github.com/coze-dev/coze-studio/backend/domain/agent/singleagent/repository"
	singleagent "github.com/coze-dev/coze-studio/backend/domain/agent/singleagent/service"
	connector "github.com/coze-dev/coze-studio/backend/domain/connector/service"
	knowledge "github.com/coze-dev/coze-studio/backend/domain/knowledge/service"
	database "github.com/coze-dev/coze-studio/backend/domain/memory/database/service"
	variables "github.com/coze-dev/coze-studio/backend/domain/memory/variables/service"
	"github.com/coze-dev/coze-studio/backend/domain/plugin/service"
	search "github.com/coze-dev/coze-studio/backend/domain/search/service"
	shortcutCmd "github.com/coze-dev/coze-studio/backend/domain/shortcutcmd/service"
	user "github.com/coze-dev/coze-studio/backend/domain/user/service"
	"github.com/coze-dev/coze-studio/backend/domain/workflow"
	"github.com/coze-dev/coze-studio/backend/infra/cache"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
	"github.com/coze-dev/coze-studio/backend/infra/imagex"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
	"github.com/coze-dev/coze-studio/backend/pkg/kvstore"
)

// SingleAgent 类型别名
type (
	SingleAgent = singleagent.SingleAgent
)

// SingleAgentSVC 单 Agent 应用服务单例
var SingleAgentSVC *SingleAgentApplicationService

// ServiceComponents 单 Agent 应用服务依赖组件
type ServiceComponents struct {
	// IDGen ID 生成器
	IDGen idgen.IDGenerator
	// DB 数据库连接
	DB *gorm.DB
	// Cache 缓存客户端
	Cache cache.Cmdable
	// TosClient 对象存储客户端
	TosClient storage.Storage
	// ImageX 图像处理服务
	ImageX imagex.ImageX
	// EventBus 项目事件总线
	EventBus search.ProjectEventBus
	// CounterRepo 计数器仓储
	CounterRepo repository.CounterRepository

	// KnowledgeDomainSVC 知识库领域服务
	KnowledgeDomainSVC knowledge.Knowledge
	// PluginDomainSVC 插件领域服务
	PluginDomainSVC service.PluginService
	// WorkflowDomainSVC 工作流领域服务
	WorkflowDomainSVC workflow.Service
	// UserDomainSVC 用户领域服务
	UserDomainSVC user.User
	// VariablesDomainSVC 变量领域服务
	VariablesDomainSVC variables.Variables
	// ConnectorDomainSVC 连接器领域服务
	ConnectorDomainSVC connector.Connector
	// DatabaseDomainSVC 数据库领域服务
	DatabaseDomainSVC database.Database
	// ShortcutCMDDomainSVC 快捷命令领域服务
	ShortcutCMDDomainSVC shortcutCmd.ShortcutCmd
	// CPStore 检查点存储
	CPStore compose.CheckPointStore
}

// InitService 初始化单 Agent 应用服务
//
// 创建仓储实例和领域服务
func InitService(c *ServiceComponents) (*SingleAgentApplicationService, error) {
	domainComponents := &singleagent.Components{
		AgentDraftRepo:   repository.NewSingleAgentRepo(c.DB, c.IDGen, c.Cache),
		AgentVersionRepo: repository.NewSingleAgentVersionRepo(c.DB, c.IDGen),
		PublishInfoRepo:  kvstore.New[entity.PublishInfo](c.DB),
		CounterRepo:      repository.NewCounterRepo(c.Cache),
		CPStore:          c.CPStore,
	}

	singleAgentDomainSVC := singleagent.NewService(domainComponents)
	SingleAgentSVC = newApplicationService(c, singleAgentDomainSVC)

	return SingleAgentSVC, nil
}
