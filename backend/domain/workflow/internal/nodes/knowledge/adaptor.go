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

// Package knowledge 实现知识库相关节点
//
// 本包提供了工作流中与知识库交互的节点实现，包括：
// - 知识库检索 (Retrieve)：从知识库中检索相关文档片段
// - 知识库索引 (Indexer)：向知识库中添加新文档
// - 知识库删除 (Deleter)：从知识库中删除文档
//
// 知识库是 Coze 平台的核心能力之一，用于实现 RAG (检索增强生成) 功能，
// 让 AI 能够基于用户上传的文档进行问答。
//
// 本文件 (adaptor.go) 包含类型转换工具函数：
// - 解析模式转换：快速/精准解析
// - 分块类型转换：自定义/默认分块
// - 检索类型转换：语义/混合/全文检索
package knowledge

import (
	"fmt"

	knowledge "github.com/coze-dev/coze-studio/backend/crossdomain/knowledge/model"
)

// convertParsingType 将前端解析类型字符串转换为后端枚举
// fast: 快速解析，速度快但可能丢失部分格式信息
// accurate: 精准解析，保留更多格式信息但耗时较长
func convertParsingType(p string) (knowledge.ParseMode, error) {
	switch p {
	case "fast":
		return knowledge.FastParseMode, nil
	case "accurate":
		return knowledge.AccurateParseMode, nil
	default:
		return "", fmt.Errorf("invalid parsingType: %s", p)
	}
}

func convertChunkType(p string) (knowledge.ChunkType, error) {
	switch p {
	case "custom":
		return knowledge.ChunkTypeCustom, nil
	case "default":
		return knowledge.ChunkTypeDefault, nil
	default:
		return "", fmt.Errorf("invalid ChunkType: %s", p)
	}
}
func convertRetrievalSearchType(s int64) (knowledge.SearchType, error) {
	switch s {
	case 0:
		return knowledge.SearchTypeSemantic, nil
	case 1:
		return knowledge.SearchTypeHybrid, nil
	case 20:
		return knowledge.SearchTypeFullText, nil
	default:
		return 0, fmt.Errorf("invalid RetrievalSearchType %v", s)
	}
}
