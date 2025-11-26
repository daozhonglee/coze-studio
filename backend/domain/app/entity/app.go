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

// Package entity 定义了应用(APP)领域的核心实体
//
// 本包包含应用相关的所有领域实体和值对象：
// - APP: 应用聚合根，表示一个完整的应用定义
// - PublishRecord: 发布记录，记录应用的发布历史
// - Resource: 资源，表示应用关联的资源信息
//
// 应用是 Coze Studio 中的核心概念，用于组织和管理 Agent、工作流、插件等资源。
package entity

import (
	"github.com/coze-dev/coze-studio/backend/crossdomain/app/model"
)

type APP = model.APP

type PublishRecord = model.PublishRecord

type PublishRecordExtraInfo = model.PublishRecordExtraInfo

type PackResourceFailedInfo = model.PackResourceFailedInfo

type ResourceCopyResult = model.ResourceCopyResult

type Resource = model.Resource
