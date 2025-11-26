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

// Package entity 定义了应用(APP)领域的核心实体（常量定义）

package entity

import "github.com/coze-dev/coze-studio/backend/crossdomain/app/model"

type PublishStatus = model.PublishStatus

const (
	// PublishStatusOfPacking 打包中 - 正在打包应用资源
	PublishStatusOfPacking PublishStatus = 0
	// PublishStatusOfPackFailed 打包失败 - 资源打包过程出错
	PublishStatusOfPackFailed PublishStatus = 1
	// PublishStatusOfAuditing 审核中 - 等待审核通过
	PublishStatusOfAuditing PublishStatus = 2
	// PublishStatusOfAuditNotPass 审核未通过 - 审核被拒绝
	PublishStatusOfAuditNotPass PublishStatus = 3
	// PublishStatusOfConnectorPublishing 连接器发布中 - 正在发布到各连接器
	PublishStatusOfConnectorPublishing PublishStatus = 4
	// PublishStatusOfPublishDone 发布完成 - 发布流程成功完成
	PublishStatusOfPublishDone PublishStatus = 5
)

type ConnectorPublishStatus = model.ConnectorPublishStatus

const (
	// ConnectorPublishStatusOfDefault 默认状态 - 初始状态
	ConnectorPublishStatusOfDefault ConnectorPublishStatus = 0
	// ConnectorPublishStatusOfAuditing 审核中 - 等待连接器审核
	ConnectorPublishStatusOfAuditing ConnectorPublishStatus = 1
	// ConnectorPublishStatusOfSuccess 发布成功 - 已成功发布到连接器
	ConnectorPublishStatusOfSuccess ConnectorPublishStatus = 2
	// ConnectorPublishStatusOfFailed 发布失败 - 发布到连接器失败
	ConnectorPublishStatusOfFailed ConnectorPublishStatus = 3
	// ConnectorPublishStatusOfDisable 已禁用 - 连接器发布已被禁用
	ConnectorPublishStatusOfDisable ConnectorPublishStatus = 4
)

type ResourceType = model.ResourceType

const (
	// ResourceTypeOfPlugin 插件资源
	ResourceTypeOfPlugin ResourceType = "plugin"
	// ResourceTypeOfWorkflow 工作流资源
	ResourceTypeOfWorkflow ResourceType = "workflow"
	// ResourceTypeOfKnowledge 知识库资源
	ResourceTypeOfKnowledge ResourceType = "knowledge"
	// ResourceTypeOfDatabase 数据库资源
	ResourceTypeOfDatabase ResourceType = "database"
)

type ResourceCopyStatus = model.ResourceCopyStatus

const (
	// ResourceCopyStatusOfSuccess 复制成功
	ResourceCopyStatusOfSuccess ResourceCopyStatus = 1
	// ResourceCopyStatusOfProcessing 复制中
	ResourceCopyStatusOfProcessing ResourceCopyStatus = 2
	// ResourceCopyStatusOfFailed 复制失败
	ResourceCopyStatusOfFailed ResourceCopyStatus = 3
)
