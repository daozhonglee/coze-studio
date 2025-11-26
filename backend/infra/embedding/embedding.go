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

// Package embedding 提供文本嵌入（Embedding）接口
//
// 本包定义文本向量化服务的接口，用于将文本转换为向量表示：
// - 支持稠密向量（Dense）嵌入
// - 支持稀疏向量（Sparse）嵌入
// - 支持混合嵌入（Hybrid）
//
// 实现层在 impl/ 目录下，支持多种嵌入模型：
// - Ark（火山引擎）
// - OpenAI
// - Ollama
// - Gemini
package embedding

import (
	"context"

	"github.com/cloudwego/eino/components/embedding"
)

// Embedder 文本嵌入器接口
//
// 扩展 Eino 框架的 Embedder 接口，增加混合嵌入和维度查询能力
type Embedder interface {
	embedding.Embedder
	EmbedStringsHybrid(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, []map[int]float64, error) // hybrid embedding
	Dimensions() int64
	SupportStatus() SupportStatus
}

// SupportStatus 嵌入支持状态
type SupportStatus int

// 嵌入支持状态常量
const (
	// SupportDense 仅支持稠密向量
	SupportDense SupportStatus = 1
	// SupportDenseAndSparse 支持稠密和稀疏向量
	SupportDenseAndSparse SupportStatus = 3
)
