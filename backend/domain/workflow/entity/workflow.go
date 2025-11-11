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

// Package entity 定义了工作流领域的核心实体
//
// 工作流实体采用了组合模式，通过嵌入多个VO(Value Object)来构建完整的实体：
// - Meta: 工作流的元数据信息（创建者、空间、类型等）
// - CanvasInfo: 画布相关信息
// - DraftMeta: 草稿版本的元数据
// - VersionMeta: 发布版本的元数据
//
// 这种设计的好处：
// 1. 职责分离：不同VO负责不同层次的数据
// 2. 组合灵活：可以根据需要组合不同的VO
// 3. 复用性：VO可以在不同上下文中复用
// 4. 可维护性：每个VO职责单一，易于维护
package entity

import (
	"github.com/coze-dev/coze-studio/backend/api/model/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
)

type ContentType = workflow.WorkFlowType
type Tag = workflow.Tag
type Mode = workflow.WorkflowMode

type Workflow struct {
	ID       int64
	CommitID string

	*vo.Meta
	*vo.CanvasInfo
	*vo.DraftMeta
	*vo.VersionMeta
}

func (w *Workflow) GetBasic() *WorkflowBasic {
	var version string
	if w.VersionMeta != nil {
		version = w.VersionMeta.Version
	}
	return &WorkflowBasic{
		ID:       w.ID,
		Version:  version,
		SpaceID:  w.SpaceID,
		AppID:    w.AppID,
		CommitID: w.CommitID,
	}
}

func (w *Workflow) GetLatestVersion() string {
	if w.LatestPublishedVersion == nil {
		return ""
	}

	return *w.LatestPublishedVersion
}

func (w *Workflow) GetVersion() string {
	if w.VersionMeta == nil {
		return ""
	}
	return w.VersionMeta.Version
}

type IDVersionPair struct {
	ID      int64
	Version string
}

type WorkflowBasic struct {
	ID       int64
	Version  string
	SpaceID  int64
	AppID    *int64
	CommitID string
}

type CopyWorkflowFromAppToLibraryResult struct {
	WorkflowIDVersionMap map[int64]IDVersionPair
	ValidateIssues       []*vo.ValidateIssue
	CopiedWorkflows      []*Workflow
}
