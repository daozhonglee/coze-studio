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

// interrupt_event.go 中断事件实体
//
// 本文件定义了工作流执行中断相关的实体：
//   - InterruptEvent: 中断事件（问答、输入等）
//   - ResumeRequest: 恢复执行请求
//   - ToolInterruptEvent: 工具中断事件

package entity

import (
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/compose"

	"github.com/coze-dev/coze-studio/backend/api/model/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
)

// InterruptEvent 中断事件
// 表示工作流执行过程中需要暂停等待用户输入的事件
type InterruptEvent struct {
	ID            int64              `json:"id"`
	NodeKey       vo.NodeKey         `json:"node_key"`
	InterruptData string             `json:"interrupt_data,omitempty"`
	NodeType      NodeType           `json:"node_type"`
	NodeTitle     string             `json:"node_title,omitempty"`
	NodeIcon      string             `json:"node_icon,omitempty"`
	EventType     InterruptEventType `json:"event_type"`
	NodePath      []string           `json:"node_path,omitempty"`
	Popped        bool               `json:"popped,omitempty"`

	// index within composite node -> interrupt info for that index
	// TODO: separate the following fields with InterruptEvent
	NestedInterruptInfo      map[int]*compose.InterruptInfo `json:"nested_interrupt_info,omitempty"`
	SubWorkflowInterruptInfo *compose.InterruptInfo         `json:"sub_workflow_interrupt_info,omitempty"`
	ToolInterruptEvent       *ToolInterruptEvent            `json:"tool_interrupt_event,omitempty"`
}

type InterruptEventType = workflow.EventType

const (
	InterruptEventQuestion = workflow.EventType_Question
	InterruptEventInput    = workflow.EventType_InputNode
	InterruptEventLLM      = 100 // interrupt events emitted by LLM node, which are emitted by nodes within workflow tools
)

func (i *InterruptEvent) String() string {
	s, _ := sonic.MarshalIndent(i, "", "  ")
	return string(s)
}

// ResumeRequest 恢复执行请求
// 包含恢复执行所需的执行 ID、事件 ID 和用户提供的数据
type ResumeRequest struct {
	ExecuteID  int64
	EventID    int64
	ResumeData string
	Resumed    bool
}

func (r *ResumeRequest) GetResumeID() string {
	return fmt.Sprintf("%d_%d", r.ExecuteID, r.EventID)
}

// ToolInterruptEvent 工具中断事件
// LLM 调用工作流工具时产生的中断事件
type ToolInterruptEvent struct {
	ToolCallID string
	ToolName   string
	ExecuteID  int64
	*InterruptEvent
}

// ConvRelatedInfo 会话关联信息
// 用于关联中断事件和会话上下文
type ConvRelatedInfo struct {
	EventID  int64
	ExecID   int64
	NodeType NodeType
}
