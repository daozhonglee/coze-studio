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

// stream.go 流式字段类型定义
//
// 本文件定义了工作流中字段流式类型的相关结构。
// 用于标识和追踪节点输入字段是否为流式数据源。

package schema

import (
	"github.com/cloudwego/eino/compose"

	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
)

// FieldStreamType 字段流式类型
type FieldStreamType string

const (
	FieldIsStream    FieldStreamType = "yes"     // 确定是流式数据
	FieldNotStream   FieldStreamType = "no"      // 确定不是流式数据
	FieldMaybeStream FieldStreamType = "maybe"   // 可能是流式，需要运行时解析
	FieldSkipped     FieldStreamType = "skipped" // 字段来源节点被跳过
)

// SourceInfo 字段来源信息
//
// 描述节点输入字段的来源和流式特性：
//   - IsIntermediate: 是否为中间容器（包含子字段）
//   - FieldType: 流式类型
//   - FromNodeKey: 来源节点 key
//   - FromPath: 来源字段路径
//   - SubSources: 子字段的来源信息
type SourceInfo struct {
	// IsIntermediate means this field is itself not a field source, but a map containing one or more field sources.
	IsIntermediate bool
	// FieldType the stream type of the field. May require request-time resolution in addition to compile-time.
	FieldType FieldStreamType
	// FromNodeKey is the node key that produces this field source. empty if the field is a static value or variable.
	FromNodeKey vo.NodeKey
	// FromPath is the path of this field source within the source node. empty if the field is a static value or variable.
	FromPath compose.FieldPath
	TypeInfo *vo.TypeInfo
	// SubSources are SourceInfo for keys within this intermediate Map(Object) field.
	SubSources map[string]*SourceInfo
}

// Skipped 判断字段来源是否被跳过
// 对于中间容器，递归检查所有子字段
func (s *SourceInfo) Skipped() bool {
	if !s.IsIntermediate {
		return s.FieldType == FieldSkipped
	}

	for _, sub := range s.SubSources {
		if !sub.Skipped() {
			return false
		}
	}

	return true
}
