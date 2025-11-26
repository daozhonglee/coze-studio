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

// Package service 定义了搜索(Search)领域的服务层接口
//
// 本包提供全文搜索服务，支持：
// - 项目搜索
// - 资源搜索
// - 搜索事件发布
package service

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/search/entity"
)

// ProjectEventBus 项目事件总线接口
//
// 用于发布项目相关的领域事件，触发搜索索引更新
type ProjectEventBus interface {
	// PublishProject 发布项目事件
	PublishProject(ctx context.Context, event *entity.ProjectDomainEvent) error
}

// ResourceEventBus 资源事件总线接口
//
// 用于发布资源相关的领域事件，触发搜索索引更新
type ResourceEventBus interface {
	// PublishResources 发布资源事件
	PublishResources(ctx context.Context, event *entity.ResourceDomainEvent) error
}

// Search 搜索服务接口
//
// 定义项目和资源的搜索操作
type Search interface {
	// SearchProjects 搜索项目
	SearchProjects(ctx context.Context, req *entity.SearchProjectsRequest) (resp *entity.SearchProjectsResponse, err error)
	// SearchResources 搜索资源
	SearchResources(ctx context.Context, req *entity.SearchResourcesRequest) (resp *entity.SearchResourcesResponse, err error)
}
