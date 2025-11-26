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

// Package workflow 定义了工作流(Workflow)应用层服务
//
// 本包提供工作流应用层的初始化和配置：
// - 工作流服务的初始化
// - 领域服务的依赖注入
// - 全局配置的加载
//
// 应用层负责协调领域服务完成业务用例，是 API 层和领域层之间的桥梁。
package workflow

import (
	"context"
	"path/filepath"

	"os"

	"gopkg.in/yaml.v3"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/bizpkg/llm/modelbuilder"
	knowledge "github.com/coze-dev/coze-studio/backend/domain/knowledge/service"
	dbservice "github.com/coze-dev/coze-studio/backend/domain/memory/database/service"
	variables "github.com/coze-dev/coze-studio/backend/domain/memory/variables/service"
	plugin "github.com/coze-dev/coze-studio/backend/domain/plugin/service"
	search "github.com/coze-dev/coze-studio/backend/domain/search/service"
	"github.com/coze-dev/coze-studio/backend/domain/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/config"
	wrapPlugin "github.com/coze-dev/coze-studio/backend/domain/workflow/plugin"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/service"
	"github.com/coze-dev/coze-studio/backend/infra/cache"
	"github.com/coze-dev/coze-studio/backend/infra/coderunner"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
	"github.com/coze-dev/coze-studio/backend/infra/imagex"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
)

// ServiceComponents 工作流应用服务依赖组件
//
// 包含初始化工作流服务所需的所有依赖
type ServiceComponents struct {
	// IDGen ID 生成器
	IDGen idgen.IDGenerator
	// DB 数据库连接
	DB *gorm.DB
	// Cache 缓存客户端
	Cache cache.Cmdable
	// DatabaseDomainSVC 数据库领域服务
	DatabaseDomainSVC dbservice.Database
	// VariablesDomainSVC 变量领域服务
	VariablesDomainSVC variables.Variables
	// PluginDomainSVC 插件领域服务
	PluginDomainSVC plugin.PluginService
	// KnowledgeDomainSVC 知识库领域服务
	KnowledgeDomainSVC knowledge.Knowledge
	// DomainNotifier 领域事件通知器
	DomainNotifier search.ResourceEventBus
	// Tos 对象存储客户端
	Tos storage.Storage
	// ImageX 图像处理服务
	ImageX imagex.ImageX
	// CPStore 检查点存储
	CPStore compose.CheckPointStore
	// CodeRunner 代码执行器
	CodeRunner coderunner.Runner
	// WorkflowBuildInChatModel 内置聊天模型
	WorkflowBuildInChatModel modelbuilder.BaseChatModel
}

// initWorkflowConfig 加载工作流配置
//
// 从 resources/conf/workflow/config.yaml 读取工作流配置
func initWorkflowConfig() (workflow.WorkflowConfig, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	configBs, err := os.ReadFile(filepath.Join(wd, "resources/conf/workflow/config.yaml"))
	if err != nil {
		return nil, err
	}
	var cfg *config.WorkflowConfig
	err = yaml.Unmarshal(configBs, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// InitService 初始化工作流应用服务
//
// 完成以下初始化工作：
// - 注册所有节点适配器
// - 创建工作流仓储
// - 创建工作流领域服务
// - 设置事件总线
// - 初始化节点图标缓存
func InitService(_ context.Context, components *ServiceComponents) (*ApplicationService, error) {
	service.RegisterAllNodeAdaptors()

	cfg, err := initWorkflowConfig()
	if err != nil {
		return nil, err
	}

	workflowRepo, err := service.NewWorkflowRepository(components.IDGen, components.DB, components.Cache,
		components.Tos, components.CPStore, components.WorkflowBuildInChatModel, cfg)
	if err != nil {
		return nil, err
	}

	workflow.SetRepository(workflowRepo)

	workflowDomainSVC := service.NewWorkflowService(workflowRepo)
	wrapPlugin.SetOSS(components.Tos)

	coderunner.SetCodeRunner(components.CodeRunner)
	callbacks.AppendGlobalHandlers(service.GetTokenCallbackHandler())

	setEventBus(components.DomainNotifier)

	SVC.DomainSVC = workflowDomainSVC
	SVC.ImageX = components.ImageX
	SVC.TosClient = components.Tos
	SVC.IDGenerator = components.IDGen

	err = SVC.InitNodeIconURLCache(context.Background())
	if err != nil {
		return nil, err
	}

	return SVC, nil
}
