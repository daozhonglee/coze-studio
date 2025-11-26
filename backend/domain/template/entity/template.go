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

// Package entity 定义了模板(Template)领域的核心实体
//
// 本包包含模板相关的领域实体，用于模板的查询和过滤。
package entity

// TemplateFilter 模板过滤条件
//
// 用于指定模板列表查询的过滤条件
type TemplateFilter struct {
	// AgentID 关联的 Agent ID
	AgentID *int64
	// SpaceID 所属工作空间ID
	SpaceID *int64
	// ProductEntityType 产品实体类型
	ProductEntityType *int64
}

// Pagination 分页参数
type Pagination struct {
	// Limit 每页数量
	Limit int
	// Offset 偏移量
	Offset int
}
