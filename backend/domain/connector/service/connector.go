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

// Package connector 定义了连接器(Connector)领域的服务层
//
// 本包提供连接器服务，用于管理应用的发布渠道
package connector

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/connector/entity"
)

// Connector 连接器服务接口
//
// 定义连接器的查询操作
type Connector interface {
	// List 列出所有可用连接器
	List(ctx context.Context) ([]*entity.Connector, error)
	// GetByIDs 批量获取连接器
	GetByIDs(ctx context.Context, ids []int64) (map[int64]*entity.Connector, error)
	// GetByID 获取单个连接器
	GetByID(ctx context.Context, id int64) (*entity.Connector, error)
}
