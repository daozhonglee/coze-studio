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

// Package entity 定义了单 Agent 领域的核心实体
//
// 本包包含 Agent 管理相关的实体定义：
// - SingleAgent: Agent 实体
// - AgentIdentity: Agent 身份标识
// - ExecuteRequest: Agent 执行请求
// - InterruptInfo: 中断信息
//
// 设计说明：
// Agent 是 Coze 平台的核心概念，代表一个可对话的智能体。
// Agent 有草稿版本（用于编辑）和发布版本（用于上线）两种状态。
package entity

import (
	model "github.com/coze-dev/coze-studio/backend/crossdomain/agent/model"
)

// SingleAgent Agent 实体
//
// 使用组合而非别名，便于在领域层扩展行为。
// 包含 Agent 的完整配置：人设、技能、知识库、工作流等。
type SingleAgent struct {
	*model.SingleAgent
}

// AgentIdentity Agent 身份标识，复用 crossdomain 中的定义
//
// 用于唯一标识一个 Agent 实例，包含 AgentID 和 ConnectorID。
type AgentIdentity = model.AgentIdentity

// ExecuteRequest Agent 执行请求，复用 crossdomain 中的定义
//
// 包含执行 Agent 所需的所有参数：输入内容、上下文、配置等。
type ExecuteRequest = model.ExecuteRequest

// InterruptInfo 中断信息，复用 crossdomain 中的定义
//
// 当 Agent 执行需要暂停等待外部输入时使用。
type InterruptInfo = model.InterruptInfo

// DuplicateInfo Agent 复制信息
//
// 用于复制 Agent 时传递必要的参数。
type DuplicateInfo struct {
	// UserID 目标用户 ID
	UserID int64
	// SpaceID 目标空间 ID
	SpaceID int64
	// NewAgentID 新 Agent ID
	NewAgentID int64
	// DraftAgent 要复制的草稿 Agent
	DraftAgent *SingleAgent
}
