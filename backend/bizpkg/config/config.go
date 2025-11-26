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

// Package config 提供业务配置管理
//
// 本包是配置管理的入口，聚合了：
// - 基础配置（base）
// - 知识库配置（knowledge）
// - 模型配置（modelmgr）
package config

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/api/model/admin/config"
	"github.com/coze-dev/coze-studio/backend/bizpkg/config/base"
	"github.com/coze-dev/coze-studio/backend/bizpkg/config/knowledge"
	"github.com/coze-dev/coze-studio/backend/bizpkg/config/modelmgr"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
)

// BasicConfiguration 基础配置类型别名
type BasicConfiguration = config.BasicConfiguration

// ModelStatus 模型状态类型别名
type ModelStatus = config.ModelStatus

// 模型状态常量
const (
	ModelStatus_StatusDefault ModelStatus = 0 // 默认状态，等同于 StatusInUse
	ModelStatus_StatusInUse   ModelStatus = 1 // 在用状态，可创建新实例
	ModelStatus_StatusDeleted ModelStatus = 2 // 已删除，不可用
)

// EmbeddingType 嵌入模型类型别名
type EmbeddingType = config.EmbeddingType

// 嵌入模型类型常量
const (
	EmbeddingType_Ark    EmbeddingType = 0 // 火山引擎 Ark
	EmbeddingType_OpenAI EmbeddingType = 1 // OpenAI
	EmbeddingType_Ollama EmbeddingType = 2 // Ollama
	EmbeddingType_Gemini EmbeddingType = 3 // Google Gemini
	EmbeddingType_HTTP   EmbeddingType = 4 // HTTP 接口
)

// Config 配置管理器聚合
type Config struct {
	base      *base.BaseConfig
	knowledge *knowledge.KnowledgeConfig
	model     *modelmgr.ModelConfig
}

// shardConfig 全局配置单例
var shardConfig *Config

// Init 初始化配置管理器
func Init(ctx context.Context, db *gorm.DB, oss storage.Storage) error {
	shardConfig = &Config{
		base:      base.NewBaseConfig(db),
		knowledge: knowledge.NewKnowledgeConfig(db),
	}

	m, err := modelmgr.Init(ctx, db, oss)
	if err != nil {
		return err
	}

	shardConfig.model = m

	return nil
}

// Base 获取基础配置管理器
func Base() *base.BaseConfig {
	return shardConfig.base
}

// Knowledge 获取知识库配置管理器
func Knowledge() *knowledge.KnowledgeConfig {
	return shardConfig.knowledge
}

// ModelConf 获取模型配置管理器
func ModelConf() *modelmgr.ModelConfig {
	return shardConfig.model
}
