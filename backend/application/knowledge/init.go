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

// Package knowledge 定义了知识库(Knowledge)应用层服务
//
// 本包提供知识库相关的应用层业务逻辑，包括：
// - 知识库的创建、更新、删除、查询
// - 文档管理（上传、分片、重新分片）
// - 表格数据处理
// - 图片知识库管理
//
// 知识库是 AI Agent 获取外部知识的重要来源，支持多种数据格式。
package knowledge

import (
	"context"
	"fmt"
	"os"

	"github.com/coze-dev/coze-studio/backend/application/search"
	knowledgeImpl "github.com/coze-dev/coze-studio/backend/domain/knowledge/service"
	"github.com/coze-dev/coze-studio/backend/infra/eventbus"
	"github.com/coze-dev/coze-studio/backend/types/consts"
)

// ServiceComponents 知识库应用服务依赖组件（复用领域层配置）
type ServiceComponents = knowledgeImpl.KnowledgeSVCConfig

// InitService 初始化知识库应用服务
//
// 创建领域服务并注册消息队列消费者
func InitService(ctx context.Context, c *ServiceComponents, bus search.ResourceEventBus) (*KnowledgeApplicationService, error) {
	knowledgeDomainSVC, knowledgeEventHandler := knowledgeImpl.NewKnowledgeSVC(c)

	nameServer := os.Getenv(consts.MQServer)
	if err := eventbus.GetDefaultSVC().RegisterConsumer(nameServer, consts.RMQTopicKnowledge, consts.RMQConsumeGroupKnowledge, knowledgeEventHandler); err != nil {
		return nil, fmt.Errorf("register knowledge consumer failed, err=%w", err)
	}

	KnowledgeSVC.DomainSVC = knowledgeDomainSVC
	KnowledgeSVC.eventBus = bus
	KnowledgeSVC.storage = c.Storage
	return KnowledgeSVC, nil
}
