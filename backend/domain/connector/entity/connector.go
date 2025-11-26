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

// Package entity 定义了连接器(Connector)领域的核心实体
//
// 本包包含连接器相关的领域实体，连接器用于将应用发布到不同渠道，
// 如 Chat SDK、API、Coze 平台等。
package entity

import (
	"github.com/coze-dev/coze-studio/backend/api/model/app/developer_api"
	"github.com/coze-dev/coze-studio/backend/crossdomain/connector/model"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/conv"
)

// Connector 连接器实体
//
// 使用组合模式扩展基础连接器模型，便于添加领域特定行为
type Connector struct {
	*model.Connector
}

// ToVO 转换为视图对象
func (c *Connector) ToVO() *developer_api.ConnectorInfo {
	return &developer_api.ConnectorInfo{
		ID:              conv.Int64ToStr(c.ID),
		Name:            c.Name,
		Icon:            c.URL,
		ConnectorStatus: c.ConnectorStatus,
	}
}
