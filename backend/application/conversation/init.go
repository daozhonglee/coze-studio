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

// Package conversation 定义了对话(Conversation)应用层服务
//
// 本包提供对话相关的应用层业务逻辑，包括：
// - 对话会话管理
// - 消息管理（发送、接收、历史记录）
// - Agent 运行记录管理
// - OpenAPI 对话接口
//
// 对话是用户与 AI Agent 交互的核心功能。
package conversation

import (
	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/application/singleagent"
	"github.com/coze-dev/coze-studio/backend/domain/conversation/agentrun/repository"
	agentrun "github.com/coze-dev/coze-studio/backend/domain/conversation/agentrun/service"
	convRepo "github.com/coze-dev/coze-studio/backend/domain/conversation/conversation/repository"
	conversation "github.com/coze-dev/coze-studio/backend/domain/conversation/conversation/service"
	msgRepo "github.com/coze-dev/coze-studio/backend/domain/conversation/message/repository"
	message "github.com/coze-dev/coze-studio/backend/domain/conversation/message/service"
	shortcutRepo "github.com/coze-dev/coze-studio/backend/domain/shortcutcmd/repository"
	"github.com/coze-dev/coze-studio/backend/domain/shortcutcmd/service"
	uploadService "github.com/coze-dev/coze-studio/backend/domain/upload/service"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
	"github.com/coze-dev/coze-studio/backend/infra/imagex"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
)

// ServiceComponents 对话应用服务依赖组件
type ServiceComponents struct {
	// IDGen ID 生成器
	IDGen idgen.IDGenerator
	// DB 数据库连接
	DB *gorm.DB
	// TosClient 对象存储客户端
	TosClient storage.Storage
	// ImageX 图像处理服务
	ImageX imagex.ImageX

	// SingleAgentDomainSVC 单 Agent 领域服务
	SingleAgentDomainSVC singleagent.SingleAgent
}

// InitService 初始化对话应用服务
//
// 创建消息、对话、Agent 运行记录等领域服务
func InitService(s *ServiceComponents) *ConversationApplicationService {
	mDomainComponents := &message.Components{
		MessageRepo: msgRepo.NewMessageRepo(s.DB, s.IDGen),
	}
	messageDomainSVC := message.NewService(mDomainComponents)

	cDomainComponents := &conversation.Components{
		ConversationRepo: convRepo.NewConversationRepo(s.DB, s.IDGen),
	}

	conversationDomainSVC := conversation.NewService(cDomainComponents)

	arDomainComponents := &agentrun.Components{
		RunRecordRepo: repository.NewRunRecordRepo(s.DB, s.IDGen),
		ImagexSVC:     s.ImageX,
	}

	agentRunDomainSVC := agentrun.NewService(arDomainComponents)
	components := &service.Components{
		ShortCutCmdRepo: shortcutRepo.NewShortCutCmdRepo(s.DB, s.IDGen),
	}
	shortcutCmdDomainSVC := service.NewShortcutCommandService(components)

	ConversationSVC.AgentRunDomainSVC = agentRunDomainSVC
	ConversationSVC.MessageDomainSVC = messageDomainSVC
	ConversationSVC.ConversationDomainSVC = conversationDomainSVC
	ConversationSVC.appContext = s
	ConversationSVC.ShortcutDomainSVC = shortcutCmdDomainSVC

	ConversationOpenAPISVC.ShortcutDomainSVC = shortcutCmdDomainSVC
	uploadSVC := uploadService.NewUploadSVC(s.DB, s.IDGen, s.TosClient)
	ConversationOpenAPISVC.UploaodDomainSVC = uploadSVC
	OpenapiMessageSVC.UploaodDomainSVC = uploadSVC

	return ConversationSVC
}
