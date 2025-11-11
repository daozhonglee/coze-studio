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

// Package oceanbase 提供了 OceanBase 向量数据库的客户端封装
//
// 在 Coze Studio 项目中，主要用于文档向量检索功能：
// 1. 作为向量搜索存储的后端实现
// 2. 支持文档向量化存储和相似度检索
// 3. 通过 embedding 服务将文档转换为向量后存储
// 4. 提供基于向量相似度的文档检索能力
//
// 使用场景：
// - 文档存储：将文档内容转换为向量并存储到 OceanBase
// - 语义搜索：根据查询文本的向量相似度检索相关文档
// - 知识库问答：为 RAG（Retrieval-Augmented Generation）提供向量检索能力
//
// 架构层次：
// OceanBaseClient -> OceanBaseOfficialClient -> GORM -> OceanBase Database
package oceanbase

import (
	"context"

	"gorm.io/gorm"
)

// OceanBaseClient OceanBase 向量数据库客户端
// 提供向量数据的 CRUD 操作，是向量搜索存储的核心组件
// 在项目中被 searchstore/oceanbase 包使用，作为向量存储的底层实现
type OceanBaseClient struct {
	official *OceanBaseOfficialClient // 官方客户端实现
}

// NewOceanBaseClient 创建新的 OceanBase 客户端实例
// 参数 dsn 是数据库连接字符串，格式为 "user:password@tcp(host:port)/database"
// 在项目中使用时，通常由 Factory 创建，通过环境变量或配置文件获取 DSN
func NewOceanBaseClient(dsn string) (*OceanBaseClient, error) {
	official, err := NewOceanBaseOfficialClient(dsn)
	if err != nil {
		return nil, err
	}

	return &OceanBaseClient{official: official}, nil
}

// CreateCollection 创建向量集合（表）
// 在项目中使用：为每个文档集合创建对应的向量存储表
// 参数 dimension 是向量的维度，通常与 embedding 模型的输出维度一致
func (c *OceanBaseClient) CreateCollection(ctx context.Context, collectionName string, dimension int) error {
	return c.official.CreateCollection(ctx, collectionName, dimension)
}

// InsertVectors 批量插入向量数据
// 在项目中使用：存储文档的向量化和元数据信息
// 每个 VectorResult 包含文档ID、内容、元数据和对应的向量嵌入
func (c *OceanBaseClient) InsertVectors(ctx context.Context, collectionName string, vectors []VectorResult) error {
	return c.official.InsertVectors(ctx, collectionName, vectors)
}

// SearchVectors 基于向量相似度搜索
// 在项目中使用：实现文档检索功能，根据查询向量找到最相似的文档
// 返回按相似度排序的结果列表，用于 RAG 问答系统
func (c *OceanBaseClient) SearchVectors(ctx context.Context, collectionName string, queryVector []float64, topK int, threshold float64) ([]VectorResult, error) {
	return c.official.SearchVectors(ctx, collectionName, queryVector, topK, threshold)
}

// SearchVectorsWithStrategy 带搜索策略的向量搜索
// 在项目中使用：支持不同的搜索策略（如不同的相似度阈值）
// 目前实现与 SearchVectors 相同，将来可扩展不同的搜索算法
func (c *OceanBaseClient) SearchVectorsWithStrategy(ctx context.Context, collectionName string, queryVector []float64, topK int, threshold float64, strategy SearchStrategy) ([]VectorResult, error) {
	return c.official.SearchVectors(ctx, collectionName, queryVector, topK, threshold)
}

// GetDB 获取底层的 GORM 数据库实例
// 在项目中使用：执行自定义 SQL 查询或直接操作数据库
// 主要用于需要复杂查询或数据库管理操作的场景
func (c *OceanBaseClient) GetDB() *gorm.DB {
	return c.official.GetDB()
}

// DebugCollectionData 调试集合数据
// 在项目中使用：开发和调试阶段查看集合中的数据内容
// 输出集合的基本信息和数据统计，用于排查问题
func (c *OceanBaseClient) DebugCollectionData(ctx context.Context, collectionName string) error {
	return c.official.DebugCollectionData(ctx, collectionName)
}

// BatchInsertVectors 批量插入向量数据（与 InsertVectors 功能相同）
// 在项目中使用：处理大量文档批量向量化存储的场景
// 提供更好的性能和错误处理
func (c *OceanBaseClient) BatchInsertVectors(ctx context.Context, collectionName string, vectors []VectorResult) error {
	return c.official.InsertVectors(ctx, collectionName, vectors)
}

// DeleteVector 删除指定的向量数据
// 在项目中使用：删除不再需要的文档向量
// 支持按 vectorID 删除单个向量记录
func (c *OceanBaseClient) DeleteVector(ctx context.Context, collectionName string, vectorID string) error {
	return c.official.GetDB().WithContext(ctx).Table(collectionName).Where("vector_id = ?", vectorID).Delete(nil).Error
}

// InitDatabase 初始化数据库连接测试
// 在项目中使用：验证数据库连接是否正常
// 执行简单的 SELECT 1 查询来测试连接可用性
func (c *OceanBaseClient) InitDatabase(ctx context.Context) error {
	var result int
	return c.official.GetDB().WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
}

// DropCollection 删除整个向量集合（表）
// 在项目中使用：清理不再使用的文档集合
// 直接删除表结构和所有数据，操作不可逆
func (c *OceanBaseClient) DropCollection(ctx context.Context, collectionName string) error {
	return c.official.GetDB().WithContext(ctx).Migrator().DropTable(collectionName)
}

// SearchStrategy 搜索策略接口
// 在项目中使用：定义不同的向量搜索策略，如相似度阈值等
// 未来可扩展为不同的搜索算法或参数配置
type SearchStrategy interface {
	GetThreshold() float64 // 获取相似度阈值，过滤低相似度的结果
}

// DefaultSearchStrategy 默认搜索策略实现
// 在项目中使用：提供最基本的搜索策略，阈值为 0.0（不进行相似度过滤）
type DefaultSearchStrategy struct{}

func NewDefaultSearchStrategy() *DefaultSearchStrategy {
	return &DefaultSearchStrategy{}
}

func (s *DefaultSearchStrategy) GetThreshold() float64 {
	return 0.0
}
