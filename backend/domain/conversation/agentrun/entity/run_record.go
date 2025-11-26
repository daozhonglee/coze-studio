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

import (
	"github.com/cloudwego/eino/schema"

	"github.com/coze-dev/coze-studio/backend/api/model/conversation/common"
	message2 "github.com/coze-dev/coze-studio/backend/api/model/conversation/message"
	singleagent "github.com/coze-dev/coze-studio/backend/crossdomain/agent/model"
	agentrun "github.com/coze-dev/coze-studio/backend/crossdomain/agentrun/model"
	message "github.com/coze-dev/coze-studio/backend/crossdomain/message/model"
	"github.com/coze-dev/coze-studio/backend/domain/conversation/agentrun/internal/dal/model"
)

// RunRecord 运行记录数据库模型，复用 dal 层定义
type RunRecord = model.RunRecord

// RunRecordMeta 运行记录元数据
//
// 包含一次 Agent 运行的完整信息，包括状态、用量、错误等。
type RunRecordMeta struct {
	// ID 运行记录 ID
	ID int64 `json:"id"`
	// ConversationID 所属对话 ID
	ConversationID int64 `json:"conversation_id"`
	// SectionID 对话分段 ID
	SectionID int64 `json:"section_id"`
	// AgentID Agent ID
	AgentID int64 `json:"agent_id"`
	// Status 运行状态
	Status RunStatus `json:"status"`
	// Error 错误信息（如果失败）
	Error *RunError `json:"error"`
	// Usage Token 用量统计
	Usage *agentrun.Usage `json:"usage"`
	// Ext 扩展字段
	Ext string `json:"ext"`
	// CreatedAt 创建时间戳
	CreatedAt int64 `json:"created_at"`
	// UpdatedAt 更新时间戳
	UpdatedAt int64 `json:"updated_at"`
	// ChatRequest 原始聊天请求
	ChatRequest *string `json:"chat_message"`
	// CompletedAt 完成时间戳
	CompletedAt int64 `json:"completed_at"`
	// FailedAt 失败时间戳
	FailedAt int64 `json:"failed_at"`
	// CreatorID 创建者用户 ID
	CreatorID int64 `json:"creator_id"`
}

// ChunkRunItem 流式输出的运行记录片段
type ChunkRunItem = RunRecordMeta

// ChunkMessageItem 流式输出的消息片段
//
// 用于在流式输出过程中逐步返回消息内容。
type ChunkMessageItem struct {
	// ID 消息 ID
	ID int64 `json:"id"`
	// ConversationID 所属对话 ID
	ConversationID int64 `json:"conversation_id"`
	// SectionID 对话分段 ID
	SectionID int64 `json:"section_id"`
	// RunID 所属运行记录 ID
	RunID int64 `json:"run_id"`
	// AgentID Agent ID
	AgentID int64 `json:"agent_id"`
	// Role 消息角色
	Role RoleType `json:"role"`
	// Type 消息类型
	Type message.MessageType `json:"type"`
	// Content 消息内容
	Content string `json:"content"`
	// ContentType 内容类型
	ContentType message.ContentType `json:"content_type"`
	// MessageType 消息类型（冗余字段）
	MessageType message.MessageType `json:"message_type"`
	// ReplyID 回复的消息 ID
	ReplyID int64 `json:"reply_id"`
	// Ext 扩展字段
	Ext map[string]string `json:"ext"`
	// ReasoningContent 推理内容（思维链）
	ReasoningContent *string `json:"reasoning_content"`
	// Index 消息序号
	Index int64 `json:"index"`
	// RequiredAction 需要用户操作的信息
	RequiredAction *message2.RequiredAction `json:"required_action"`
	// SeqID 序列 ID
	SeqID int64 `json:"seq_id"`
	// CreatedAt 创建时间戳
	CreatedAt int64 `json:"created_at"`
	// UpdatedAt 更新时间戳
	UpdatedAt int64 `json:"updated_at"`
	// IsFinish 是否完成
	IsFinish bool `json:"is_finish"`
}

// RunError 运行错误信息
type RunError struct {
	// Code 错误码
	Code int64 `json:"code"`
	// Msg 错误消息
	Msg string `json:"msg"`
}

// CustomerConfig 用户自定义配置
//
// 允许用户在运行时覆盖 Agent 的默认配置。
type CustomerConfig struct {
	// ModelConfig 模型配置覆盖
	ModelConfig *ModelConfig `json:"model_config"`
	// AgentConfig Agent 配置覆盖
	AgentConfig *AgentConfig `json:"agent_config"`
}

// ModelConfig 模型配置
type ModelConfig struct {
	// ModelId 使用的模型 ID
	ModelId *int64 `json:"model_id,omitempty"`
}

// AgentConfig Agent 配置
type AgentConfig struct {
	// Prompt 自定义提示词
	Prompt *string `json:"prompt"`
}

// Tool 工具定义，复用 crossdomain 中的定义
type Tool = agentrun.Tool

// AnswerFinshContent 回答完成内容
type AnswerFinshContent struct {
	// MsgType 消息子类型
	MsgType MessageSubType `json:"msg_type"`
	// Data 数据内容
	Data string `json:"data"`
	// FromUnit 来源单元
	FromUnit string `json:"from_unit"`
}

// Data 完成数据
type Data struct {
	// FinishReason 完成原因
	FinishReason int32 `json:"finish_reason"`
	// FinData 完成数据
	FinData string `json:"fin_data"`
}

// MetaInfo 元信息
type MetaInfo struct {
	// Type 元信息类型
	Type MetaType `json:"type"`
	// Info 信息内容
	Info string `json:"info"`
}

// AgentRunMeta Agent 运行请求元数据
//
// 包含发起一次 Agent 运行所需的所有信息。
type AgentRunMeta struct {
	// ConversationID 对话 ID
	ConversationID int64 `json:"conversation_id"`
	// ConnectorID 连接器 ID（渠道标识）
	ConnectorID int64 `json:"connector_id"`
	// SpaceID 空间 ID
	SpaceID int64 `json:"space_id"`
	// Scene 使用场景
	Scene common.Scene `json:"scene"`
	// SectionID 对话分段 ID
	SectionID int64 `json:"section_id"`
	// Name 用户名称
	Name string `json:"name"`
	// UserID 用户 ID（外部系统）
	UserID string `json:"user_id"`
	// CozeUID Coze 平台用户 ID
	CozeUID int64 `json:"coze_uid"`
	// AgentID Agent ID
	AgentID int64 `json:"agent_id"`
	// ContentType 输入内容类型
	ContentType message.ContentType `json:"content_type"`
	// Content 输入内容（支持多模态）
	Content []*message.InputMetaData `json:"content"`
	// PreRetrieveTools 预检索工具列表
	PreRetrieveTools []*Tool `json:"tools"`
	// IsDraft 是否使用草稿版本
	IsDraft bool `json:"is_draft"`
	// CustomerConfig 用户自定义配置
	CustomerConfig *CustomerConfig `json:"customer_config"`
	// DisplayContent 展示内容
	DisplayContent string `json:"display_content"`
	// CustomVariables 自定义变量
	CustomVariables map[string]string `json:"custom_variables"`
	// Version Agent 版本
	Version string `json:"version"`
	// Ext 扩展字段
	Ext map[string]string `json:"ext"`
	// AdditionalMessages 附加消息（用于上下文）
	AdditionalMessages []*AdditionalMessage `json:"additional_messages"`
	// ChatflowParameters 对话流参数
	ChatflowParameters map[string]any `json:"chatflow_parameters"`
}

// AdditionalMessage 附加消息
//
// 用于在 Agent 运行时提供额外的上下文消息。
type AdditionalMessage struct {
	// Role 消息角色
	Role schema.RoleType `json:"role"`
	// Type 消息类型
	Type message.MessageType `json:"type"`
	// Content 消息内容
	Content []*message.InputMetaData `json:"content"`
	// ContentType 内容类型
	ContentType message.ContentType `json:"content_type"`
	// Name 发送者名称
	Name *string `json:"name"`
	// Meta 元信息
	Meta map[string]string `json:"meta"`
}

// UpdateMeta 运行记录更新元数据
type UpdateMeta struct {
	// Status 新状态
	Status RunStatus
	// LastError 错误信息
	LastError *RunError
	// Usage Token 用量
	Usage *agentrun.Usage
	// UpdatedAt 更新时间戳
	UpdatedAt int64
	// CompletedAt 完成时间戳
	CompletedAt int64
	// FailedAt 失败时间戳
	FailedAt int64
}

// AgentRunResponse Agent 运行响应
//
// 流式输出的单个响应事件，包含运行状态或消息内容。
type AgentRunResponse struct {
	// Event 事件类型
	Event RunEvent `json:"event"`
	// ChunkRunItem 运行记录片段
	ChunkRunItem *ChunkRunItem `json:"run_record_item"`
	// ChunkMessageItem 消息片段
	ChunkMessageItem *ChunkMessageItem `json:"message_item"`
	// Error 错误信息
	Error *RunError `json:"error"`
}

// AgentRespEvent Agent 响应事件
//
// 内部使用的 Agent 响应事件结构，包含各类型的响应数据。
type AgentRespEvent struct {
	// EventType 事件类型
	EventType message.MessageType `json:"event_type"`

	// ToolMidAnswer 工具中间回答流
	ToolMidAnswer *schema.StreamReader[*schema.Message]
	// ToolAsAnswer 工具作为回答的流
	ToolAsAnswer *schema.StreamReader[*schema.Message]
	// ModelAnswer 模型回答流
	ModelAnswer *schema.StreamReader[*schema.Message]
	// ToolsMessage 工具消息列表
	ToolsMessage []*schema.Message
	// FuncCall 函数调用消息
	FuncCall *schema.Message
	// Suggest 建议问题
	Suggest *schema.Message
	// Knowledge 知识库检索结果
	Knowledge []*schema.Document
	// Interrupt 中断信息
	Interrupt *singleagent.InterruptInfo
	// Err 错误
	Err error
}

// ModelAnswerEvent 模型回答事件
type ModelAnswerEvent struct {
	// Message 消息内容
	Message *schema.Message
	// Err 错误
	Err error
}

// ListRunRecordMeta 运行记录列表查询参数
type ListRunRecordMeta struct {
	// ConversationID 对话 ID
	ConversationID int64 `json:"conversation_id"`
	// AgentID Agent ID
	AgentID int64 `json:"agent_id"`
	// SectionID 分段 ID
	SectionID int64 `json:"section_id"`
	// Limit 返回数量限制
	Limit int32 `json:"limit"`
	// OrderBy 排序方式（desc/asc）
	OrderBy string `json:"order_by"`
	// BeforeID 游标：小于此 ID
	BeforeID int64 `json:"before_id"`
	// AfterID 游标：大于此 ID
	AfterID int64 `json:"after_id"`
}

// CancelRunMeta 取消运行请求参数
type CancelRunMeta struct {
	// ConversationID 对话 ID
	ConversationID int64 `json:"conversation_id"`
	// RunID 运行记录 ID
	RunID int64 `json:"run_id"`
}
