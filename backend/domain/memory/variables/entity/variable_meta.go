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

// Package entity 定义了变量记忆领域的核心实体
//
// 本包包含 Agent 变量管理相关的实体定义：
// - VariableMeta: 变量元数据（定义）
// - VariablesMeta: 变量集合
// - VariableInstance: 变量实例（运行时值）
// - VariableMetaSchema: 变量 Schema 定义
//
// 设计说明：
// 变量记忆功能允许 Agent 在对话中持久化存储键值对数据。
// 支持用户自定义变量和系统变量，系统变量由系统自动生成（如 sys_uuid）。
// 变量按渠道（Channel）分类，支持飞书、位置、系统等不同来源。
package entity

import (
	"context"
	"fmt"

	"github.com/coze-dev/coze-studio/backend/api/model/data/variable/project_memory"
)

// VariableMeta 变量元数据定义
//
// 表示一个变量的完整定义，包括名称、类型、默认值、来源渠道等。
type VariableMeta struct {
	// Keyword 变量关键字（唯一标识）
	Keyword string
	// DefaultValue 默认值
	DefaultValue string
	// VariableType 变量类型（KV变量等）
	VariableType project_memory.VariableType
	// Channel 变量来源渠道（自定义/系统/飞书/位置等）
	Channel project_memory.VariableChannel
	// Description 变量描述
	Description string
	// Enable 是否启用
	Enable bool
	// EffectiveChannelList 生效渠道列表
	EffectiveChannelList []string
	// Schema 变量结构定义（JSON Schema 格式）
	Schema string
	// IsReadOnly 是否只读
	IsReadOnly bool
	// PromptDisabled 是否禁止在提示词中使用
	PromptDisabled bool
}

// NewVariableMeta 从 API 模型创建变量元数据
func NewVariableMeta(e *project_memory.Variable) *VariableMeta {
	return &VariableMeta{
		Keyword:              e.Keyword,
		DefaultValue:         e.DefaultValue,
		VariableType:         e.VariableType,
		Channel:              e.Channel,
		Description:          e.Description,
		Enable:               e.Enable,
		EffectiveChannelList: e.EffectiveChannelList,
		Schema:               e.Schema,
		IsReadOnly:           e.IsReadOnly,
	}
}

// ToProjectVariable 转换为 API 模型
func (v *VariableMeta) ToProjectVariable() *project_memory.Variable {
	return &project_memory.Variable{
		Keyword:              v.Keyword,
		DefaultValue:         v.DefaultValue,
		VariableType:         v.VariableType,
		Channel:              v.Channel,
		Description:          v.Description,
		Enable:               v.Enable,
		EffectiveChannelList: v.EffectiveChannelList,
		Schema:               v.Schema,
		IsReadOnly:           v.IsReadOnly,
	}
}

// GetSchema 解析并返回变量的 Schema 对象
func (v *VariableMeta) GetSchema(ctx context.Context) (*VariableMetaSchema, error) {
	return NewVariableMetaSchema([]byte(v.Schema))
}

// CheckSchema 验证变量 Schema 的有效性
func (v *VariableMeta) CheckSchema(ctx context.Context) error {
	schema, err := NewVariableMetaSchema([]byte(v.Schema))
	if err != nil {
		return err
	}

	return schema.check(ctx)
}

// stringSchema 默认的字符串类型 Schema 模板
const stringSchema = "{\n    \"type\": \"string\",\n    \"name\": \"%v\",\n    \"required\": false\n}"

// SetupSchema 设置默认 Schema（如果为空）
func (v *VariableMeta) SetupSchema() {
	if v.Schema == "" {
		v.Schema = fmt.Sprintf(stringSchema, v.Keyword)
	}
}

// SetupIsReadOnly 根据渠道设置只读属性
//
// 飞书、位置、系统渠道的变量自动设置为只读
func (v *VariableMeta) SetupIsReadOnly() {
	if v.Channel == project_memory.VariableChannel_Feishu ||
		v.Channel == project_memory.VariableChannel_Location ||
		v.Channel == project_memory.VariableChannel_System {
		v.IsReadOnly = true
	}
}

// IsSystem 判断是否为系统变量
func (v *VariableMeta) IsSystem() bool {
	return v.Channel == project_memory.VariableChannel_System
}
