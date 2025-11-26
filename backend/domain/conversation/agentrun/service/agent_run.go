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

// Package agentrun 定义了 Agent 运行领域的服务接口
//
// 本包提供 Agent 对话运行的核心能力：
// - 发起 Agent 运行（流式输出）
// - 运行记录管理（创建、查询、删除）
// - 运行控制（取消）
//
// 设计说明：
// Agent 运行是一次用户与 Agent 的完整交互过程，
// 从用户发送消息开始，到 Agent 回复结束。
// 支持流式输出，通过事件机制通知客户端处理进度。
package agentrun

import (
	"context"

	"github.com/cloudwego/eino/schema"

	"github.com/coze-dev/coze-studio/backend/domain/conversation/agentrun/entity"
)

// Run Agent 运行服务接口
//
// 该接口定义了 Agent 运行相关的所有操作。
//
//go:generate mockgen -destination ../../../../internal/mock/domain/conversation/agentrun/agent_run_mock.go --package agentrun -source agent_run.go
type Run interface {
	// AgentRun 发起 Agent 运行
	//
	// 返回流式响应读取器，客户端可以逐个读取响应事件。
	AgentRun(ctx context.Context, req *entity.AgentRunMeta) (*schema.StreamReader[*entity.AgentRunResponse], error)

	// Delete 删除运行记录
	Delete(ctx context.Context, runID []int64) error

	// Create 创建运行记录
	Create(ctx context.Context, runRecord *entity.AgentRunMeta) (*entity.RunRecordMeta, error)

	// List 查询运行记录列表
	List(ctx context.Context, ListMeta *entity.ListRunRecordMeta) ([]*entity.RunRecordMeta, error)

	// GetByID 根据 ID 获取运行记录
	GetByID(ctx context.Context, runID int64) (*entity.RunRecordMeta, error)

	// Cancel 取消运行中的 Agent 任务
	Cancel(ctx context.Context, req *entity.CancelRunMeta) (*entity.RunRecordMeta, error)
}
