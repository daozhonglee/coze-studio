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

// Package es 提供 Elasticsearch 客户端接口
//
// 本包定义 Elasticsearch 服务的接口，用于：
// - 索引管理（创建/删除索引）
// - 文档 CRUD 操作
// - 搜索查询
// - 批量索引
//
// 实现层在 impl/es/ 目录下，支持 ES7 和 ES8
package es

import (
	"context"
)

// Client Elasticsearch 客户端接口
type Client interface {
	Create(ctx context.Context, index, id string, document any) error
	Update(ctx context.Context, index, id string, document any) error
	Delete(ctx context.Context, index, id string) error
	Search(ctx context.Context, index string, req *Request) (*Response, error)
	Exists(ctx context.Context, index string) (bool, error)
	CreateIndex(ctx context.Context, index string, properties map[string]any) error
	DeleteIndex(ctx context.Context, index string) error
	Types() Types
	NewBulkIndexer(index string) (BulkIndexer, error)
}

// Types ES 类型定义接口
type Types interface {
	NewLongNumberProperty() any
	NewTextProperty() any
	NewUnsignedLongNumberProperty() any
}

// BulkIndexer 批量索引器接口
type BulkIndexer interface {
	Add(ctx context.Context, item BulkIndexerItem) error
	Close(ctx context.Context) error
}
