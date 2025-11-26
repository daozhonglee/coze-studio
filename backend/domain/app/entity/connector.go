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

// Package entity 定义了应用(APP)领域的核心实体（连接器相关）

package entity

import (
	"github.com/coze-dev/coze-studio/backend/crossdomain/app/model"
	"github.com/coze-dev/coze-studio/backend/types/consts"
)

// ConnectorIDWhiteList 允许的连接器ID白名单
//
// 当前支持的连接器：
// - WebSDKConnectorID: Web SDK 连接器，用于网页嵌入
// - APIConnectorID: API 连接器，用于 API 调用
var ConnectorIDWhiteList = []int64{
	consts.WebSDKConnectorID,
	consts.APIConnectorID,
}

type ConnectorPublishRecord = model.ConnectorPublishRecord
type PublishConfig = model.PublishConfig
type SelectedWorkflow = model.SelectedWorkflow
