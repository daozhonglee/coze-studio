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

// MessageExtKey 消息扩展字段键名枚举
type MessageExtKey string

// 消息扩展字段键名常量
const (
	// MessageExtKeyInputTokens 输入 Token 数
	MessageExtKeyInputTokens MessageExtKey = "input_tokens"
	// MessageExtKeyOutputTokens 输出 Token 数
	MessageExtKeyOutputTokens MessageExtKey = "output_tokens"
	// MessageExtKeyToken 总 Token 数
	MessageExtKeyToken MessageExtKey = "token"
	// MessageExtKeyPluginStatus 插件执行状态
	MessageExtKeyPluginStatus MessageExtKey = "plugin_status"
	// MessageExtKeyTimeCost 耗时
	MessageExtKeyTimeCost MessageExtKey = "time_cost"
	// MessageExtKeyWorkflowTokens 工作流 Token 消耗
	MessageExtKeyWorkflowTokens MessageExtKey = "workflow_tokens"
	// MessageExtKeyBotState Agent 状态
	MessageExtKeyBotState MessageExtKey = "bot_state"
	// MessageExtKeyPluginRequest 插件请求
	MessageExtKeyPluginRequest MessageExtKey = "plugin_request"
	// MessageExtKeyToolName 工具名称
	MessageExtKeyToolName MessageExtKey = "tool_name"
	// MessageExtKeyPlugin 插件信息
	MessageExtKeyPlugin MessageExtKey = "plugin"
	// MessageExtKeyMockHitInfo Mock 命中信息
	MessageExtKeyMockHitInfo MessageExtKey = "mock_hit_info"
	// MessageExtKeyMessageTitle 消息标题
	MessageExtKeyMessageTitle MessageExtKey = "message_title"
	// MessageExtKeyStreamPluginRunning 流式插件运行中
	MessageExtKeyStreamPluginRunning MessageExtKey = "stream_plugin_running"
	// MessageExtKeyExecuteDisplayName 执行显示名称
	MessageExtKeyExecuteDisplayName MessageExtKey = "execute_display_name"
	// MessageExtKeyTaskType 任务类型
	MessageExtKeyTaskType MessageExtKey = "task_type"
	// MessageExtKeyCallID 调用 ID
	MessageExtKeyCallID MessageExtKey = "call_id"
	// ExtKeyResumeInfo 恢复信息
	ExtKeyResumeInfo MessageExtKey = "resume_info"
	// ExtKeyBreakPoint 断点信息
	ExtKeyBreakPoint MessageExtKey = "break_point"
	// ExtKeyToolCallsIDs 工具调用 ID 列表
	ExtKeyToolCallsIDs MessageExtKey = "tool_calls_ids"
	// ExtKeyRequiresAction 需要用户操作标记
	ExtKeyRequiresAction MessageExtKey = "requires_action"
)

// BotStateExt Agent 状态扩展信息
type BotStateExt struct {
	// BotID Bot ID
	BotID string `json:"bot_id"`
	// AgentName Agent 名称
	AgentName string `json:"agent_name"`
	// AgentID Agent ID
	AgentID string `json:"agent_id"`
	// Awaiting 等待状态
	Awaiting string `json:"awaiting"`
}

// UsageExt Token 用量扩展信息
type UsageExt struct {
	// TotalCount 总 Token 数
	TotalCount int64 `json:"total_count"`
	// InputTokens 输入 Token 数
	InputTokens int64 `json:"input_tokens"`
	// OutputTokens 输出 Token 数
	OutputTokens int64 `json:"output_tokens"`
}
