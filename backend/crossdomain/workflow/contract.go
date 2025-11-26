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

package workflow

import (
	"context"

	"github.com/cloudwego/eino/compose"
	einoCompose "github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	workflowModel "github.com/coze-dev/coze-studio/backend/crossdomain/workflow/model"
	"github.com/coze-dev/coze-studio/backend/domain/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity"
	workflowEntity "github.com/coze-dev/coze-studio/backend/domain/workflow/entity"

	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
)

// Workflow 工作流跨域服务接口
type Workflow interface {
	WorkflowAsModelTool(ctx context.Context, policies []*vo.GetPolicy) ([]workflow.ToolFromWorkflow, error)
	WithResumeToolWorkflow(resumingEvent *workflowEntity.ToolInterruptEvent, resumeData string,
		allInterruptEvents map[string]*workflowEntity.ToolInterruptEvent) einoCompose.Option
	ReleaseApplicationWorkflows(ctx context.Context, appID int64, config *ReleaseWorkflowConfig) ([]*vo.ValidateIssue, error)
	GetWorkflowIDsByAppID(ctx context.Context, appID int64) ([]int64, error)

	SyncExecuteWorkflow(ctx context.Context, config workflowModel.ExecuteConfig, input map[string]any) (*workflowEntity.WorkflowExecution, vo.TerminatePlan, error)
	StreamExecute(ctx context.Context, config workflowModel.ExecuteConfig, input map[string]any) (*schema.StreamReader[*workflowEntity.Message], error)
	WithExecuteConfig(cfg workflowModel.ExecuteConfig) einoCompose.Option
	WithMessagePipe() (compose.Option, *schema.StreamReader[*entity.Message], func())
	StreamResume(ctx context.Context, req *entity.ResumeRequest, config workflowModel.ExecuteConfig) (*schema.StreamReader[*entity.Message], error)
	InitApplicationDefaultConversationTemplate(ctx context.Context, spaceID int64, appID int64, userID int64) error
	MGet(ctx context.Context, policy *vo.MGetPolicy) ([]*entity.Workflow, int64, error)
}

// 类型别名定义
type (
	// ExecuteConfig 执行配置
	ExecuteConfig = workflowModel.ExecuteConfig
	// ExecuteMode 执行模式
	ExecuteMode = workflowModel.ExecuteMode
	// WorkflowMessage 工作流消息
	WorkflowMessage = workflowEntity.Message
	// StateMessage 状态消息
	StateMessage = workflowEntity.StateMessage
	// NodeType 节点类型
	NodeType = entity.NodeType
	// MessageType 消息类型
	MessageType = entity.MessageType
	// InterruptEvent 中断事件
	InterruptEvent = workflowEntity.InterruptEvent
	// EventType 事件类型
	EventType = workflowEntity.InterruptEventType
	// ResumeRequest 恢复请求
	ResumeRequest = entity.ResumeRequest
	// WorkflowExecuteStatus 工作流执行状态
	WorkflowExecuteStatus = entity.WorkflowExecuteStatus
)

// 工作流执行状态常量
const (
	WorkflowRunning     = WorkflowExecuteStatus(entity.WorkflowRunning)
	WorkflowSuccess     = WorkflowExecuteStatus(entity.WorkflowSuccess)
	WorkflowFailed      = WorkflowExecuteStatus(entity.WorkflowFailed)
	WorkflowCancel      = WorkflowExecuteStatus(entity.WorkflowCancel)
	WorkflowInterrupted = WorkflowExecuteStatus(entity.WorkflowInterrupted)
)

const (
	Answer       MessageType = "answer"
	FunctionCall MessageType = "function_call"
	ToolResponse MessageType = "tool_response"
)

// 节点类型常量
const (
	NodeTypeOutputEmitter NodeType = "OutputEmitter"
	NodeTypeInputReceiver NodeType = "InputReceiver"
	NodeTypeQuestion      NodeType = "QuestionAnswer"
)

// 执行模式常量
const (
	ExecuteModeDebug     ExecuteMode = "debug"
	ExecuteModeRelease   ExecuteMode = "release"
	ExecuteModeNodeDebug ExecuteMode = "node_debug"
)

// TaskType 任务类型
type TaskType = workflowModel.TaskType

// SyncPattern 同步模式
type SyncPattern = workflowModel.SyncPattern

// 同步模式常量
const (
	SyncPatternSync   SyncPattern = "sync"
	SyncPatternAsync  SyncPattern = "async"
	SyncPatternStream SyncPattern = "stream"
)

// 任务类型常量
const (
	TaskTypeForeground TaskType = "foreground"
	TaskTypeBackground TaskType = "background"
)

// BizType 业务类型
type BizType = workflowModel.BizType

// 业务类型常量
const (
	BizTypeAgent    BizType = "agent"
	BizTypeWorkflow BizType = "workflow"
)

// Locator 定位器（版本定位）
type Locator = workflowModel.Locator

// 版本定位常量
const (
	FromDraft Locator = iota
	FromSpecificVersion
	FromLatestVersion
)

// ReleaseWorkflowConfig 发布工作流配置
type ReleaseWorkflowConfig = vo.ReleaseWorkflowConfig

// ToolInterruptEvent 工具中断事件
type ToolInterruptEvent = workflowEntity.ToolInterruptEvent

// defaultSVC 默认服务实例
var defaultSVC Workflow

// DefaultSVC 获取默认服务实例
func DefaultSVC() Workflow {
	return defaultSVC
}

// SetDefaultSVC 设置默认服务实例
func SetDefaultSVC(svc Workflow) {
	defaultSVC = svc
}
