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

// Package entity 定义了搜索(Search)领域的核心实体
//
// 本包包含搜索相关的领域实体和请求/响应结构，
// 支持项目和资源的全文搜索。
package entity

import (
	"github.com/coze-dev/coze-studio/backend/api/model/app/intelligence/common"
	model "github.com/coze-dev/coze-studio/backend/crossdomain/search/model"
)

// 资源索引字段常量
const (
	// FieldOfResType 资源类型字段
	FieldOfResType = "res_type"
	// FieldOfPublishStatus 发布状态字段
	FieldOfPublishStatus = "publish_status"
	// FieldOfResSubType 资源子类型字段
	FieldOfResSubType = "res_sub_type"
	// FieldOfBizStatus 业务状态字段
	FieldOfBizStatus = "biz_status"
	// FieldOfScores 评分字段
	FieldOfScores = "scores"
)

// SearchProjectsRequest 搜索项目请求
type SearchProjectsRequest struct {
	SpaceID   int64
	ProjectID int64
	OwnerID   int64
	Name      string
	Status    []common.IntelligenceStatus
	Types     []common.IntelligenceType

	IsPublished    bool
	IsFav          bool
	IsRecentlyOpen bool
	OrderFiledName string
	OrderAsc       bool

	Cursor string
	Limit  int32
}

type SearchProjectsResponse struct {
	HasMore    bool
	NextCursor string

	Data []*ProjectDocument
}

type SearchResourcesRequest = model.SearchResourcesRequest

type SearchResourcesResponse = model.SearchResourcesResponse
