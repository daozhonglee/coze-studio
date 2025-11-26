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

// message.go 工作流执行消息实体
//
// 本文件定义了工作流执行过程中产生的消息实体：
//   - Message: 统一消息容器
//   - StateMessage: 状态变更消息
//   - DataMessage: 数据消息（用于对话历史）
//   - FunctionCallInfo/ToolResponseInfo: 函数调用相关信息

package entity

import (
	"github.com/cloudwego/eino/schema"

	"github.com/coze-dev/coze-studio/backend/api/model/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
)

// Message 工作流执行消息
// 包含状态消息和数据消息两种类型
type Message struct {
	*StateMessage
	*DataMessage
}

// StateMessage 状态消息
// 表示工作流执行状态的变更，包含执行 ID、状态、Token 使用量、错误信息等
type StateMessage struct {
	ExecuteID      int64
	EventID        int64 // the resuming event ID for current execution
	SpaceID        int64
	Status         WorkflowExecuteStatus
	Usage          *TokenUsage
	LastError      vo.WorkflowError
	InterruptEvent *InterruptEvent
}

// DataMessage 数据消息
// 表示执行过程中产生的完整或分块消息，用于记录到对话历史
type DataMessage struct {
	ExecuteID    int64 // the root execute ID for current execution
	Role         schema.RoleType
	Type         MessageType
	Content      string
	NodeID       string
	NodeTitle    string
	NodeType     NodeType
	Last         bool
	Usage        *TokenUsage
	FunctionCall *FunctionCallInfo
	ToolResponse *ToolResponseInfo
}

type MessageType string

const (
	Answer       MessageType = "answer"
	FunctionCall MessageType = "function_call"
	ToolResponse MessageType = "tool_response"
)

type FunctionInfo struct {
	Name string   `json:"name"`
	Type ToolType `json:"plugin_type"`

	PluginID   int64  `json:"plugin_id,omitempty"`
	PluginName string `json:"plugin_name,omitempty"`
	APIID      int64  `json:"api_id,omitempty"`
	APIName    string `json:"api_name,omitempty"`

	WorkflowName          string           `json:"workflow_name,omitempty"`
	WorkflowTerminatePlan vo.TerminatePlan `json:"terminate_plan,omitempty"`
}

type FunctionCallInfo struct {
	FunctionInfo
	CallID    string         `json:"-"`
	Arguments map[string]any `json:"arguments"`
}

type ToolResponseInfo struct {
	FunctionInfo
	CallID   string
	Response string
}

type ToolType = workflow.PluginType

const (
	PluginTool   = workflow.PluginType_PLUGIN
	WorkflowTool = workflow.PluginType_WORKFLOW
)
