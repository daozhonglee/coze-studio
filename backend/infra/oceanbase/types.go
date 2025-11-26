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

// Package oceanbase 提供 OceanBase 向量数据库的客户端封装（类型定义）
package oceanbase

// VectorIndexConfig 向量索引配置
//
// 定义向量索引的参数，支持多种索引类型（HNSW、IVF）和距离度量方式
type VectorIndexConfig struct {
	Distance string
	//Index types: hnsw, hnsw_sq, hnsw_bq, ivf_flat, ivf_sq8, ivf_pq
	Type string
	//Index library type: vsag, ob
	Lib string
	// HNSW Index parameters
	M              *int
	EfConstruction *int
	EfSearch       *int
	// IVF Index parameters
	Nlist *int
	Nbits *int
	IVFM  *int
}

// VectorData 向量数据结构体
//
// 存储文档的向量化数据，包括原始内容、元数据和嵌入向量
type VectorData struct {
	ID             int64                  `json:"id"`
	CollectionName string                 `json:"collection_name"`
	VectorID       string                 `json:"vector_id"`
	Content        string                 `json:"content"`
	Metadata       map[string]interface{} `json:"metadata"`
	Embedding      []float64              `json:"embedding"`
}

// VectorSearchResult 向量搜索结果
type VectorSearchResult struct {
	ID       int64   `json:"id"`
	Content  string  `json:"content"`
	Metadata string  `json:"metadata"`
	Distance float64 `json:"distance"`
}

// VectorMemoryEstimate 向量内存估算
type VectorMemoryEstimate struct {
	MinMemoryMB         int `json:"min_memory_mb"`
	RecommendedMemoryMB int `json:"recommended_memory_mb"`
	EstimatedMemoryMB   int `json:"estimated_memory_mb"`
}

// 向量索引类型常量
const (
	VectorIndexTypeHNSW   = "hnsw"
	VectorIndexTypeHNSWSQ = "hnsw_sq"
	VectorIndexTypeHNSWBQ = "hnsw_bq"
	VectorIndexTypeIVF    = "ivf_flat"
	VectorIndexTypeIVFSQ  = "ivf_sq8"
	VectorIndexTypeIVFPQ  = "ivf_pq"
)

// 向量距离度量类型常量
const (
	VectorDistanceTypeL2           = "l2"
	VectorDistanceTypeCosine       = "cosine"
	VectorDistanceTypeInnerProduct = "inner_product"
)

// 向量索引库类型常量
const (
	VectorLibTypeVSAG = "vsag"
	VectorLibTypeOB   = "ob"
)

// DefaultVectorIndexConfig 获取默认向量索引配置
//
// 默认使用 HNSW 索引，余弦距离，适用于大多数场景
func DefaultVectorIndexConfig() *VectorIndexConfig {
	m := 16
	efConstruction := 200
	efSearch := 64

	return &VectorIndexConfig{
		Distance:       VectorDistanceTypeCosine,
		Type:           VectorIndexTypeHNSW,
		Lib:            VectorLibTypeVSAG,
		M:              &m,
		EfConstruction: &efConstruction,
		EfSearch:       &efSearch,
	}
}

// HNSWVectorIndexConfig 创建 HNSW 向量索引配置
//
// HNSW（Hierarchical Navigable Small World）是一种高效的近似最近邻搜索算法
func HNSWVectorIndexConfig(distance string, m, efConstruction, efSearch int) *VectorIndexConfig {
	return &VectorIndexConfig{
		Distance:       distance,
		Type:           VectorIndexTypeHNSW,
		Lib:            VectorLibTypeVSAG,
		M:              &m,
		EfConstruction: &efConstruction,
		EfSearch:       &efSearch,
	}
}

// IVFVectorIndexConfig 创建 IVF 向量索引配置
//
// IVF（Inverted File）是一种基于聚类的向量索引算法
func IVFVectorIndexConfig(distance string, nlist, nbits, m int) *VectorIndexConfig {
	return &VectorIndexConfig{
		Distance: distance,
		Type:     VectorIndexTypeIVF,
		Lib:      VectorLibTypeOB,
		Nlist:    &nlist,
		Nbits:    &nbits,
		IVFM:     &m,
	}
}
