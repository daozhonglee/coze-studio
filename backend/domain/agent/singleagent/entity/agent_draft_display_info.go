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

package entity

import "github.com/coze-dev/coze-studio/backend/api/model/app/developer_api"

// AgentDraftDisplayInfo Agent 草稿展示信息
//
// 存储 Agent 草稿在 IDE 中的展示配置，如面板展开状态等。
type AgentDraftDisplayInfo struct {
	// AgentID Agent ID
	AgentID int64
	// DisplayInfo 展示配置数据
	DisplayInfo *developer_api.DraftBotDisplayInfoData
	// SpaceID 空间 ID
	SpaceID *string
}
