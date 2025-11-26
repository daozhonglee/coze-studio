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

// Package search 定义了搜索(Search)应用层服务
//
// 本包提供搜索相关的应用层业务逻辑，包括：
// - 项目/智能体搜索
// - 资源搜索（工作流、知识库、插件、数据库）
// - 搜索索引管理
// - 事件总线消费处理
//
// 搜索功能基于 Elasticsearch 实现全文检索。
package search

import (
	"context"
	"fmt"
	"os"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/application/singleagent"
	app "github.com/coze-dev/coze-studio/backend/domain/app/service"
	connector "github.com/coze-dev/coze-studio/backend/domain/connector/service"
	knowledge "github.com/coze-dev/coze-studio/backend/domain/knowledge/service"
	database "github.com/coze-dev/coze-studio/backend/domain/memory/database/service"
	"github.com/coze-dev/coze-studio/backend/domain/plugin/service"
	prompt "github.com/coze-dev/coze-studio/backend/domain/prompt/service"
	search "github.com/coze-dev/coze-studio/backend/domain/search/service"
	user "github.com/coze-dev/coze-studio/backend/domain/user/service"
	"github.com/coze-dev/coze-studio/backend/domain/workflow"
	"github.com/coze-dev/coze-studio/backend/infra/cache"
	"github.com/coze-dev/coze-studio/backend/infra/es"
	"github.com/coze-dev/coze-studio/backend/infra/eventbus"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
	"github.com/coze-dev/coze-studio/backend/types/consts"
)

// ServiceComponents 搜索应用服务依赖组件
type ServiceComponents struct {
	// DB 数据库连接
	DB *gorm.DB
	// Cache 缓存客户端
	Cache cache.Cmdable
	// TOS 对象存储服务
	TOS storage.Storage
	// ESClient Elasticsearch 客户端
	ESClient es.Client
	// ProjectEventBus 项目事件总线
	ProjectEventBus ProjectEventBus
	// ResourceEventBus 资源事件总线
	ResourceEventBus ResourceEventBus
	// SingleAgentDomainSVC 单 Agent 领域服务
	SingleAgentDomainSVC singleagent.SingleAgent
	// APPDomainSVC 应用领域服务
	APPDomainSVC app.AppService
	// KnowledgeDomainSVC 知识库领域服务
	KnowledgeDomainSVC knowledge.Knowledge
	// PluginDomainSVC 插件领域服务
	PluginDomainSVC service.PluginService
	// WorkflowDomainSVC 工作流领域服务
	WorkflowDomainSVC workflow.Service
	// UserDomainSVC 用户领域服务
	UserDomainSVC user.User
	// ConnectorDomainSVC 连接器领域服务
	ConnectorDomainSVC connector.Connector
	// PromptDomainSVC 提示词领域服务
	PromptDomainSVC prompt.Prompt
	// DatabaseDomainSVC 数据库领域服务
	DatabaseDomainSVC database.Database
}

// InitService 初始化搜索应用服务
//
// 创建领域服务并注册事件消费者
func InitService(ctx context.Context, s *ServiceComponents) (*SearchApplicationService, error) {
	searchDomainSVC := search.NewDomainService(ctx, s.ESClient)

	SearchSVC.DomainSVC = searchDomainSVC
	SearchSVC.ServiceComponents = s

	// setup consumer
	searchConsumer := search.NewProjectHandler(ctx, s.ESClient)

	logs.Infof("start search domain consumer...")
	nameServer := os.Getenv(consts.MQServer)

	err := eventbus.GetDefaultSVC().RegisterConsumer(nameServer, consts.RMQTopicApp, consts.RMQConsumeGroupApp, searchConsumer)
	if err != nil {
		return nil, fmt.Errorf("register search consumer failed, err=%w", err)
	}

	searchResourceConsumer := search.NewResourceHandler(ctx, s.ESClient)

	err = eventbus.GetDefaultSVC().RegisterConsumer(nameServer, consts.RMQTopicResource, consts.RMQConsumeGroupResource, searchResourceConsumer)
	if err != nil {
		return nil, fmt.Errorf("register search consumer failed, err=%w", err)
	}

	return SearchSVC, nil
}

// 事件总线类型别名
type (
	// ResourceEventBus 资源事件总线类型别名
	ResourceEventBus = search.ResourceEventBus
	// ProjectEventBus 项目事件总线类型别名
	ProjectEventBus = search.ProjectEventBus
)

// NewResourceEventBus 创建资源事件总线
func NewResourceEventBus(p eventbus.Producer) search.ResourceEventBus {
	return search.NewResourceEventBus(p)
}

// NewProjectEventBus 创建项目事件总线
func NewProjectEventBus(p eventbus.Producer) search.ProjectEventBus {
	return search.NewProjectEventBus(p)
}
