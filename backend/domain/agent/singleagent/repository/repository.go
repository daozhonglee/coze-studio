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

// Package repository 定义了单 Agent 领域的仓储接口
//
// 本包提供 Agent 数据的持久化操作接口：
// - SingleAgentDraftRepo: 草稿 Agent 仓储
// - SingleAgentVersionRepo: 发布版本仓储
// - CounterRepository: 计数器仓储
//
// 设计说明：
// Agent 数据分为草稿表和版本表两部分存储。
// 草稿表保存当前编辑中的配置，版本表保存每次发布的快照。
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/agent/singleagent/entity"
	"github.com/coze-dev/coze-studio/backend/domain/agent/singleagent/internal/dal"
	"github.com/coze-dev/coze-studio/backend/infra/cache"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
)

// NewSingleAgentRepo 创建草稿 Agent 仓储实例
func NewSingleAgentRepo(db *gorm.DB, idGen idgen.IDGenerator, cli cache.Cmdable) SingleAgentDraftRepo {
	return dal.NewSingleAgentDraftDAO(db, idGen, cli)
}

// NewSingleAgentVersionRepo 创建版本 Agent 仓储实例
func NewSingleAgentVersionRepo(db *gorm.DB, idGen idgen.IDGenerator) SingleAgentVersionRepo {
	return dal.NewSingleAgentVersion(db, idGen)
}

// NewCounterRepo 创建计数器仓储实例
func NewCounterRepo(cli cache.Cmdable) CounterRepository {
	return dal.NewCountRepo(cli)
}

// SingleAgentDraftRepo 草稿 Agent 仓储接口
//
// 提供草稿 Agent 的 CRUD 操作。
type SingleAgentDraftRepo interface {
	// Create 创建草稿 Agent
	Create(ctx context.Context, creatorID int64, draft *entity.SingleAgent) (draftID int64, err error)

	// CreateWithID 使用指定 ID 创建草稿 Agent
	CreateWithID(ctx context.Context, creatorID, agentID int64, draft *entity.SingleAgent) (draftID int64, err error)

	// Get 获取草稿 Agent
	Get(ctx context.Context, agentID int64) (*entity.SingleAgent, error)

	// MGet 批量获取草稿 Agent
	MGet(ctx context.Context, agentIDs []int64) ([]*entity.SingleAgent, error)

	// Delete 删除草稿 Agent
	Delete(ctx context.Context, spaceID, agentID int64) (err error)

	// Update 更新草稿 Agent
	Update(ctx context.Context, agentInfo *entity.SingleAgent) (err error)

	// Save 保存草稿 Agent（不存在则创建，存在则更新）
	Save(ctx context.Context, agentInfo *entity.SingleAgent) (err error)

	// GetDisplayInfo 获取草稿展示信息
	GetDisplayInfo(ctx context.Context, userID, agentID int64) (*entity.AgentDraftDisplayInfo, error)

	// UpdateDisplayInfo 更新草稿展示信息
	UpdateDisplayInfo(ctx context.Context, userID int64, e *entity.AgentDraftDisplayInfo) error
}

// SingleAgentVersionRepo 版本 Agent 仓储接口
//
// 提供发布版本 Agent 的存储和查询操作。
type SingleAgentVersionRepo interface {
	// GetLatest 获取最新发布版本
	GetLatest(ctx context.Context, agentID int64) (*entity.SingleAgent, error)

	// Get 获取指定版本的 Agent
	Get(ctx context.Context, agentID int64, version string) (*entity.SingleAgent, error)

	// List 查询发布历史列表
	List(ctx context.Context, agentID int64, pageIndex, pageSize int32) ([]*entity.SingleAgentPublish, error)

	// SavePublishRecord 保存发布记录
	SavePublishRecord(ctx context.Context, p *entity.SingleAgentPublish, e *entity.SingleAgent) (err error)

	// Create 创建新版本
	Create(ctx context.Context, connectorID int64, version string, e *entity.SingleAgent) (int64, error)
}

// CounterRepository 计数器仓储接口
//
// 提供基于 Redis 的计数器操作，用于弹窗展示次数等场景。
type CounterRepository interface {
	// Get 获取计数值
	Get(ctx context.Context, key string) (int64, error)

	// IncrBy 增加计数值
	IncrBy(ctx context.Context, key string, incr int64) error

	// Set 设置计数值
	Set(ctx context.Context, key string, value int64) error

	// Del 删除计数器
	Del(ctx context.Context, key string) error
}
