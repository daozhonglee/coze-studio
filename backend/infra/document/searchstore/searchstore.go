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

// Package searchstore 提供向量搜索存储接口
//
// 本包定义向量搜索存储的接口，用于知识库文档的索引和检索：
// - 文档索引（向量化存储）
// - 向量检索（相似度搜索）
// - 文档删除
//
// 实现层在 impl/ 目录下，支持多种后端：
// - Elasticsearch（全文检索）
// - Milvus（向量检索）
// - OceanBase（向量+全文检索）
// - VikingDB（向量检索）
package searchstore

import (
	"context"

	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/retriever"
)

// SearchStore 搜索存储接口
//
// 聚合 Eino 框架的 Indexer 和 Retriever 接口
type SearchStore interface {
	indexer.Indexer

	retriever.Retriever

	Delete(ctx context.Context, ids []string) error
}
