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

// Package agent 定义了跨域 Agent 服务接口
//
// 本包提供 Agent 跨域服务的契约定义，用于：
// - Agent 流式执行
// - Agent 身份获取
// - 快捷命令组件类型映射
//
// 跨域层作为领域层和 API 层之间的桥梁，提供领域无关的服务接口
package agent

import (
	"context"

	"github.com/cloudwego/eino/schema"

	model "github.com/coze-dev/coze-studio/backend/crossdomain/agent/model"

	agentrun "github.com/coze-dev/coze-studio/backend/crossdomain/agentrun/model"

	"github.com/coze-dev/coze-studio/backend/api/model/playground"
)

// SingleAgent 单 Agent 跨域服务接口
//
// 请求和响应不得引用领域实体，只能使用 crossdomain/model 下的模型
type SingleAgent interface {
	StreamExecute(ctx context.Context,
		agentRuntime *AgentRuntime) (*schema.StreamReader[*model.AgentEvent], error)
	ObtainAgentByIdentity(ctx context.Context, identity *model.AgentIdentity) (*model.SingleAgent, error)
	GetSingleAgentDraft(ctx context.Context, agentID int64) (agentInfo *model.SingleAgent, err error)
}

// AgentRuntime Agent 运行时配置
type AgentRuntime struct {
	AgentVersion     string
	UserID           string
	AgentID          int64
	ConversationId   int64
	IsDraft          bool
	SpaceID          int64
	ConnectorID      int64
	PreRetrieveTools []*agentrun.Tool
	CustomVariables  map[string]string

	HistoryMsg []*schema.Message
	Input      *schema.Message
	ResumeInfo *ResumeInfo
}

// ResumeInfo Agent 恢复信息类型别名
type ResumeInfo = model.InterruptInfo

// AgentEvent Agent 事件类型别名
type AgentEvent = model.AgentEvent

// defaultSVC 默认服务实例
var defaultSVC SingleAgent

// DefaultSVC 获取默认服务实例
func DefaultSVC() SingleAgent {
	return defaultSVC
}

// SetDefaultSVC 设置默认服务实例
func SetDefaultSVC(svc SingleAgent) {
	defaultSVC = svc
}

// ShortcutCommandComponentType 快捷命令组件类型
type ShortcutCommandComponentType string

// 快捷命令组件类型常量
const (
	ShortcutCommandComponentTypeText   ShortcutCommandComponentType = "text"
	ShortcutCommandComponentTypeSelect ShortcutCommandComponentType = "select"
	ShortcutCommandComponentTypeFile   ShortcutCommandComponentType = "file"
)

// ShortcutCommandComponentTypeMapping 输入类型到快捷命令组件类型的映射
var ShortcutCommandComponentTypeMapping = map[playground.InputType]ShortcutCommandComponentType{
	playground.InputType_TextInput:   ShortcutCommandComponentTypeText,
	playground.InputType_Select:      ShortcutCommandComponentTypeSelect,
	playground.InputType_MixUpload:   ShortcutCommandComponentTypeFile,
	playground.InputType_UploadImage: ShortcutCommandComponentTypeFile,
	playground.InputType_UploadDoc:   ShortcutCommandComponentTypeFile,
	playground.InputType_UploadTable: ShortcutCommandComponentTypeFile,
	playground.InputType_UploadAudio: ShortcutCommandComponentTypeFile,
	playground.InputType_VIDEO:       ShortcutCommandComponentTypeFile,
	playground.InputType_ARCHIVE:     ShortcutCommandComponentTypeFile,
	playground.InputType_CODE:        ShortcutCommandComponentTypeFile,
	playground.InputType_TXT:         ShortcutCommandComponentTypeFile,
	playground.InputType_PPT:         ShortcutCommandComponentTypeFile,
}

// ShortcutCommandComponentFileType 快捷命令组件文件类型
type ShortcutCommandComponentFileType string

// 快捷命令组件文件类型常量
const (
	ShortcutCommandComponentFileTypeImage ShortcutCommandComponentFileType = "image"
	ShortcutCommandComponentFileTypeDoc   ShortcutCommandComponentFileType = "doc"
	ShortcutCommandComponentFileTypeTable ShortcutCommandComponentFileType = "table"
	ShortcutCommandComponentFileTypeAudio ShortcutCommandComponentFileType = "audio"
	ShortcutCommandComponentFileTypeVideo ShortcutCommandComponentFileType = "video"
	ShortcutCommandComponentFileTypeZip   ShortcutCommandComponentFileType = "zip"
	ShortcutCommandComponentFileTypeCode  ShortcutCommandComponentFileType = "code"
	ShortcutCommandComponentFileTypeTxt   ShortcutCommandComponentFileType = "txt"
	ShortcutCommandComponentFileTypePPT   ShortcutCommandComponentFileType = "ppt"
)

// ShortcutCommandComponentFileTypeMapping 输入类型到文件类型的映射
var ShortcutCommandComponentFileTypeMapping = map[playground.InputType]ShortcutCommandComponentFileType{
	playground.InputType_UploadImage: ShortcutCommandComponentFileTypeImage,
	playground.InputType_UploadDoc:   ShortcutCommandComponentFileTypeDoc,
	playground.InputType_UploadTable: ShortcutCommandComponentFileTypeTable,
	playground.InputType_UploadAudio: ShortcutCommandComponentFileTypeAudio,
	playground.InputType_VIDEO:       ShortcutCommandComponentFileTypeVideo,
	playground.InputType_ARCHIVE:     ShortcutCommandComponentFileTypeZip,
	playground.InputType_CODE:        ShortcutCommandComponentFileTypeCode,
	playground.InputType_TXT:         ShortcutCommandComponentFileTypeTxt,
	playground.InputType_PPT:         ShortcutCommandComponentFileTypePPT,
}

// ShortcutCommandToolType 快捷命令工具类型
type ShortcutCommandToolType string

// 快捷命令工具类型常量
const (
	ShortcutCommandToolTypeWorkflow ShortcutCommandToolType = "workflow"
	ShortcutCommandToolTypePlugin   ShortcutCommandToolType = "plugin"
)
