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

// Package singleagent 定义了单 Agent 领域的服务接口
//
// 本包提供 Agent 的完整生命周期管理能力：
// - 草稿管理：创建、编辑、删除 Agent 草稿
// - 发布管理：发布 Agent 到各渠道、查询发布历史
// - 执行能力：运行 Agent 处理用户输入
//
// 设计说明：
// Agent 采用草稿/发布双版本机制：
// - 草稿版本用于开发者在 IDE 中编辑和调试
// - 发布版本用于上线到各渠道供用户使用
package singleagent

import (
	"context"

	"github.com/cloudwego/eino/schema"

	"github.com/coze-dev/coze-studio/backend/api/model/playground"
	"github.com/coze-dev/coze-studio/backend/domain/agent/singleagent/entity"
)

// SingleAgent Agent 服务接口
//
// 该接口定义了 Agent 管理的所有操作。
//
//go:generate mockgen -destination ../../../../internal/mock/domain/agent/singleagent/single_agent_mock.go --package singleagent -source single_agent.go
type SingleAgent interface {
	// ==================== 草稿 Agent 管理 ====================

	// CreateSingleAgentDraft 创建 Agent 草稿
	CreateSingleAgentDraft(ctx context.Context, creatorID int64, draft *entity.SingleAgent) (agentID int64, err error)

	// CreateSingleAgentDraftWithID 使用指定 ID 创建 Agent 草稿
	CreateSingleAgentDraftWithID(ctx context.Context, creatorID, agentID int64, draft *entity.SingleAgent) (int64, error)

	// MGetSingleAgentDraft 批量获取 Agent 草稿
	MGetSingleAgentDraft(ctx context.Context, agentIDs []int64) (agents []*entity.SingleAgent, err error)

	// GetSingleAgentDraft 获取 Agent 草稿
	GetSingleAgentDraft(ctx context.Context, agentID int64) (agentInfo *entity.SingleAgent, err error)

	// UpdateSingleAgentDraft 更新 Agent 草稿
	UpdateSingleAgentDraft(ctx context.Context, agentInfo *entity.SingleAgent) (err error)

	// DeleteAgentDraft 删除 Agent 草稿
	DeleteAgentDraft(ctx context.Context, spaceID, agentID int64) (err error)

	// UpdateAgentDraftDisplayInfo 更新 Agent 草稿展示信息
	UpdateAgentDraftDisplayInfo(ctx context.Context, userID int64, e *entity.AgentDraftDisplayInfo) error

	// GetAgentDraftDisplayInfo 获取 Agent 草稿展示信息
	GetAgentDraftDisplayInfo(ctx context.Context, userID, agentID int64) (*entity.AgentDraftDisplayInfo, error)

	// ==================== 上线 Agent 管理 ====================

	// CreateSingleAgent 创建发布版本的 Agent
	CreateSingleAgent(ctx context.Context, connectorID int64, version string, e *entity.SingleAgent) (int64, error)

	// DuplicateInMemory 在内存中复制 Agent
	DuplicateInMemory(ctx context.Context, req *entity.DuplicateInfo) (newAgent *entity.SingleAgent, err error)

	// StreamExecute 流式执行 Agent
	//
	// 返回事件流，客户端可以逐个读取 Agent 的处理结果。
	StreamExecute(ctx context.Context, req *entity.ExecuteRequest) (events *schema.StreamReader[*entity.AgentEvent], err error)

	// GetSingleAgent 获取发布版本的 Agent
	GetSingleAgent(ctx context.Context, agentID int64, version string) (botInfo *entity.SingleAgent, err error)

	// ListAgentPublishHistory 查询 Agent 发布历史
	ListAgentPublishHistory(ctx context.Context, agentID int64, pageIndex, pageSize int32, connectorID *int64) ([]*entity.SingleAgentPublish, error)

	// ObtainAgentByIdentity 根据身份标识获取 Agent
	//
	// 支持通过 connectorID 和 agentID 组合获取 Agent。
	ObtainAgentByIdentity(ctx context.Context, identity *entity.AgentIdentity) (*entity.SingleAgent, error)

	// ==================== 弹窗计数 ====================

	// GetAgentPopupCount 获取 Agent 弹窗展示次数
	GetAgentPopupCount(ctx context.Context, uid, agentID int64, agentPopupType playground.BotPopupType) (int64, error)

	// IncrAgentPopupCount 增加 Agent 弹窗展示次数
	IncrAgentPopupCount(ctx context.Context, uid, agentID int64, agentPopupType playground.BotPopupType) error

	// ==================== 发布管理 ====================

	// GetPublishedTime 获取 Agent 最后发布时间
	GetPublishedTime(ctx context.Context, agentID int64) (int64, error)

	// GetPublishedInfo 获取 Agent 发布信息
	GetPublishedInfo(ctx context.Context, agentID int64) (*entity.PublishInfo, error)

	// SavePublishRecord 保存发布记录
	SavePublishRecord(ctx context.Context, p *entity.SingleAgentPublish, e *entity.SingleAgent) error

	// GetPublishConnectorList 获取可发布的渠道列表
	GetPublishConnectorList(ctx context.Context, agentID int64) (*entity.PublishConnectorData, error)
}
