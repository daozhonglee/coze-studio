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

// chatflow_role.go ChatFlow 角色实体
//
// 本文件定义了 ChatFlow 工作流的角色配置实体。
// ChatFlowRole 用于定义对话机器人的外观和行为设置。

package entity

import "time"

// ChatFlowRole ChatFlow 角色配置
// 定义对话机器人的名称、头像、背景、开场白等个性化设置
type ChatFlowRole struct {
	ID                  int64
	WorkflowID          int64
	ConnectorID         int64
	Name                string
	Description         string
	Version             string
	AvatarUri           string
	BackgroundImageInfo string
	OnboardingInfo      string
	SuggestReplyInfo    string
	AudioConfig         string
	UserInputConfig     string
	CreatorID           int64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
