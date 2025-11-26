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

// Package connector 定义了连接器(Connector)应用层服务
//
// 本包提供连接器相关的应用层业务逻辑，包括：
// - 连接器列表查询
// - 连接器配置管理
//
// 连接器是 Agent 发布的目标渠道（如 API、WebSDK 等）
package connector

import (
	connector "github.com/coze-dev/coze-studio/backend/domain/connector/service"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
)

// InitService 初始化连接器应用服务
func InitService(tos storage.Storage) *ConnectorApplicationService {
	connectorDomainSVC := connector.NewService(tos)
	ConnectorApplicationSVC = New(connectorDomainSVC, tos)

	return ConnectorApplicationSVC
}
