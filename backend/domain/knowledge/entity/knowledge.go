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

// knowledge.go 知识库实体定义
//
// 本文件定义了知识库领域的核心实体 Knowledge 及其查询选项。

package entity

import model "github.com/coze-dev/coze-studio/backend/crossdomain/knowledge/model"

// Knowledge 知识库实体
// 封装跨域知识库模型
type Knowledge struct {
	*model.Knowledge
}

// WhereKnowledgeOption 知识库查询条件选项
type WhereKnowledgeOption struct {
	KnowledgeIDs []int64
	AppID        *int64
	SpaceID      *int64
	Name         *string // Exact match
	Status       []int32
	UserID       *int64
	Query        *string // fuzzy match
	Page         *int
	PageSize     *int
	Order        *Order
	OrderType    *OrderType
	FormatType   *int64
}

// OrderType 排序类型
type OrderType int32

// 排序类型常量
const (
	OrderTypeAsc  OrderType = 1 // 升序
	OrderTypeDesc OrderType = 2 // 降序
)

// Order 排序字段
type Order int32

// 排序字段常量
const (
	OrderCreatedAt Order = 1 // 按创建时间排序
	OrderUpdatedAt Order = 2 // 按更新时间排序
)
