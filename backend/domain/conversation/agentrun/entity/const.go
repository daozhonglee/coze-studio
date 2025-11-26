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

// Package entity 定义了 Agent 运行领域的核心实体和常量
//
// 本包包含 Agent 对话运行相关的定义：
// - RunStatus: 运行状态枚举
// - RunEvent: 运行事件类型
// - ReplyType: 回复类型
// - RoleType: 消息角色类型
//
// 设计说明：
// Agent 运行是对话的核心概念，表示一次用户与 Agent 的交互过程。
// 运行状态用于追踪交互进度，事件用于流式输出通知。
package entity

// ConversationTurnsDefault 对话默认轮次限制
const ConversationTurnsDefault int32 = 100

// RunStatus 运行状态枚举
type RunStatus string

// 运行状态常量
const (
	// RunStatusCreated 已创建，等待处理
	RunStatusCreated RunStatus = "created"
	// RunStatusInProgress 处理中
	RunStatusInProgress RunStatus = "in_progress"
	// RunStatusCompleted 已完成
	RunStatusCompleted RunStatus = "completed"
	// RunStatusFailed 失败
	RunStatusFailed RunStatus = "failed"
	// RunStatusExpired 已过期
	RunStatusExpired RunStatus = "expired"
	// RunStatusCancelled 已取消
	RunStatusCancelled RunStatus = "cancelled"
	// RunStatusRequiredAction 需要用户操作（如插件授权）
	RunStatusRequiredAction RunStatus = "required_action"
	// RunStatusDeleted 已删除
	RunStatusDeleted RunStatus = "deleted"
)

// RunEvent 运行事件类型枚举
//
// 用于流式输出时通知客户端当前的处理状态和消息内容。
type RunEvent string

// 运行事件常量
const (
	// RunEventCreated 对话创建事件
	RunEventCreated RunEvent = "conversation.chat.created"
	// RunEventInProgress 对话处理中事件
	RunEventInProgress RunEvent = "conversation.chat.in_progress"
	// RunEventCompleted 对话完成事件
	RunEventCompleted RunEvent = "conversation.chat.completed"
	// RunEventFailed 对话失败事件
	RunEventFailed RunEvent = "conversation.chat.failed"
	// RunEventExpired 对话过期事件
	RunEventExpired RunEvent = "conversation.chat.expired"
	// RunEventCancelled 对话取消事件
	RunEventCancelled RunEvent = "conversation.chat.cancelled"
	// RunEventRequiredAction 需要用户操作事件
	RunEventRequiredAction RunEvent = "conversation.chat.required_action"

	// RunEventMessageDelta 消息增量更新事件（流式输出）
	RunEventMessageDelta RunEvent = "conversation.message.delta"
	// RunEventMessageCompleted 消息完成事件
	RunEventMessageCompleted RunEvent = "conversation.message.completed"

	// RunEventAck 确认事件
	RunEventAck = "conversation.ack"
	// RunEventError 错误事件
	RunEventError RunEvent = "conversation.error"
	// RunEventStreamDone 流式输出结束事件
	RunEventStreamDone RunEvent = "conversation.stream.done"
)

// ReplyType 回复类型枚举
type ReplyType int64

// 回复类型常量
const (
	// ReplyTypeAnswer 最终回答
	ReplyTypeAnswer ReplyType = 1
	// ReplyTypeSuggest 建议问题
	ReplyTypeSuggest ReplyType = 2
	// ReplyTypeLLMOutput 大模型输出
	ReplyTypeLLMOutput ReplyType = 3
	// ReplyTypeToolOutput 工具输出
	ReplyTypeToolOutput ReplyType = 4
	// ReplyTypeVerbose 详细信息（调试用）
	ReplyTypeVerbose ReplyType = 100
	// ReplyTypePlaceHolder 占位符
	ReplyTypePlaceHolder ReplyType = 101
)

// MetaType 元信息类型枚举
type MetaType int64

// 元信息类型常量
const (
	// MetaTypeKnowledgeCard 知识库卡片
	MetaTypeKnowledgeCard MetaType = 4
)

// RoleType 消息角色类型枚举
type RoleType string

// 角色类型常量
const (
	// RoleTypeSystem 系统角色（设置 Agent 人设）
	RoleTypeSystem RoleType = "system"
	// RoleTypeUser 用户角色
	RoleTypeUser RoleType = "user"
	// RoleTypeAssistant 助手角色（Agent 回复）
	RoleTypeAssistant RoleType = "assistant"
	// RoleTypeTool 工具角色（插件/工具调用）
	RoleTypeTool RoleType = "tool"
)

// MessageSubType 消息子类型枚举
type MessageSubType string

// 消息子类型常量
const (
	// MessageSubTypeKnowledgeCall 知识库召回
	MessageSubTypeKnowledgeCall MessageSubType = "knowledge_recall"
	// MessageSubTypeGenerateFinish 生成完成
	MessageSubTypeGenerateFinish MessageSubType = "generate_answer_finish"
	// MessageSubTypeInterrupt 中断
	MessageSubTypeInterrupt MessageSubType = "interrupt"
)
