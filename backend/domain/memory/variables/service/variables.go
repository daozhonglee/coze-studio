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

// Package service 定义了变量记忆领域的服务接口
//
// 本包提供 Agent/Project 变量的完整管理能力：
// - 变量元数据的增删改查
// - 变量实例（运行时值）的读写
// - 系统变量配置
// - 变量发布
//
// 设计说明：
// 变量分为元数据（定义）和实例（运行时值）两层。
// 元数据定义变量的名称、类型、默认值等，在 Agent 配置时设置。
// 实例是用户在对话过程中设置的具体值，按用户隔离存储。
package service

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/api/model/data/variable/kvmemory"
	"github.com/coze-dev/coze-studio/backend/api/model/data/variable/project_memory"
	"github.com/coze-dev/coze-studio/backend/domain/memory/variables/entity"
)

// Variables 变量记忆服务接口
//
// 该接口定义了变量管理的所有操作，包括：
// - 元数据管理：获取、创建、更新变量定义
// - 实例管理：获取、设置、删除用户的变量值
// - 系统配置：获取系统预定义变量
type Variables interface {
	// GetVariableMeta 获取变量元数据
	GetVariableMeta(ctx context.Context, bizID string, bizType project_memory.VariableConnector, version string) (*entity.VariablesMeta, error)

	// GetVariableMetaByID 根据 ID 获取变量元数据
	GetVariableMetaByID(ctx context.Context, id int64) (*entity.VariablesMeta, error)

	// GetAgentVariableMeta 获取 Agent 的变量元数据
	GetAgentVariableMeta(ctx context.Context, agentID int64, version string) (*entity.VariablesMeta, error)

	// GetProjectVariablesMeta 获取 Project 的变量元数据
	GetProjectVariablesMeta(ctx context.Context, projectID, version string) (*entity.VariablesMeta, error)

	// GetSysVariableConf 获取系统预定义变量配置
	GetSysVariableConf(ctx context.Context) entity.SysConfVariables

	// UpsertMeta 创建或更新变量元数据
	UpsertMeta(ctx context.Context, e *entity.VariablesMeta) (int64, error)

	// UpsertProjectMeta 创建或更新 Project 的变量元数据
	UpsertProjectMeta(ctx context.Context, projectID, version string, userID int64, e *entity.VariablesMeta) (int64, error)

	// UpsertBotMeta 创建或更新 Agent 的变量元数据
	UpsertBotMeta(ctx context.Context, agentID int64, version string, userID int64, e *entity.VariablesMeta) (int64, error)

	// PublishMeta 发布变量元数据到指定版本
	PublishMeta(ctx context.Context, variableMetaID int64, version string) (int64, error)

	// SetVariableInstance 设置用户的变量值
	SetVariableInstance(ctx context.Context, e *entity.UserVariableMeta, items []*kvmemory.KVItem) ([]string, error)

	// GetVariableInstance 获取用户的变量值
	GetVariableInstance(ctx context.Context, e *entity.UserVariableMeta, keywords []string) ([]*kvmemory.KVItem, error)

	// GetVariableChannelInstance 获取指定渠道的用户变量值
	GetVariableChannelInstance(ctx context.Context, e *entity.UserVariableMeta, keywords []string, varChannel *project_memory.VariableChannel) ([]*kvmemory.KVItem, error)

	// DeleteVariableInstance 删除用户的变量值
	DeleteVariableInstance(ctx context.Context, e *entity.UserVariableMeta, keywords []string) error

	// DeleteAllVariable 删除业务对象下的所有变量数据
	DeleteAllVariable(ctx context.Context, bizType project_memory.VariableConnector, bizID string) (err error)

	// DecryptSysUUIDKey 解密系统用户唯一标识
	DecryptSysUUIDKey(ctx context.Context, encryptSysUUIDKey string) *entity.VariableInstance
}
