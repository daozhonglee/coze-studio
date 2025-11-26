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

// Package rerank 提供文档重排序接口
//
// 本包定义文档重排序服务的接口，用于知识库检索结果的优化：
// - 根据查询相关性对文档重新排序
// - 支持多种重排序算法（RRF、VikingDB）
package rerank

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// Reranker 文档重排序器接口
type Reranker interface {
	Rerank(ctx context.Context, req *Request) (*Response, error)
}

// Request 重排序请求
type Request struct {
	Query string
	Data  [][]*Data
	TopN  *int64
}

// Response 重排序响应
type Response struct {
	SortedData []*Data // High score
	TokenUsage *int64
}

// Data 文档数据及评分
type Data struct {
	Document *schema.Document
	Score    float64
}
