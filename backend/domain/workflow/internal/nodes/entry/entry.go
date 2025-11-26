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

// Package entry 实现工作流入口节点
//
// 入口节点是每个工作流的起点，负责接收和处理工作流的输入参数。
// 核心功能：
//
// 1. 输入接收
//   - 接收外部传入的参数
//   - 定义参数类型和默认值
//
// 2. 默认值处理
//   - 当输入为空或未提供时，使用预设的默认值
//   - 支持字符串、数组、对象类型的空值检测
//
// 3. 特殊标识
//   - 入口节点 ID 固定为 "entry"
//   - 每个工作流只能有一个入口节点
//   - 入口节点不能有父节点（不能嵌套在其他节点内部）
package entry

import (
	"context"
	"fmt"

	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/canvas/convert"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/schema"
)

// Config 入口节点配置
type Config struct {
	// DefaultValues 默认值映射，key 为参数名，value 为默认值
	DefaultValues map[string]any
}

func (c *Config) Adapt(_ context.Context, n *vo.Node, _ ...nodes.AdaptOption) (*schema.NodeSchema, error) {
	if n.Parent() != nil {
		return nil, fmt.Errorf("entry node cannot have parent: %s", n.Parent().ID)
	}

	if n.ID != entity.EntryNodeKey {
		return nil, fmt.Errorf("entry node id must be %s, got %s", entity.EntryNodeKey, n.ID)
	}

	ns := &schema.NodeSchema{
		Key:  vo.NodeKey(n.ID),
		Name: n.Data.Meta.Title,
		Type: entity.NodeTypeEntry,
	}

	defaultValues := make(map[string]any, len(n.Data.Outputs))
	for _, v := range n.Data.Outputs {
		variable, err := vo.ParseVariable(v)
		if err != nil {
			return nil, err
		}
		if variable.DefaultValue != nil {
			defaultValues[variable.Name] = variable.DefaultValue
		}
	}

	c.DefaultValues = defaultValues
	ns.Configs = c

	if err := convert.SetOutputTypesForNodeSchema(n, ns); err != nil {
		return nil, err
	}

	return ns, nil
}

func (c *Config) Build(ctx context.Context, ns *schema.NodeSchema, _ ...schema.BuildOption) (any, error) {
	defaultValues, _, err := nodes.ConvertInputs(ctx, c.DefaultValues, ns.OutputTypes, nodes.FailFast(), nodes.SkipRequireCheck())
	if err != nil {
		return nil, err
	}

	return &Entry{
		defaultValues: defaultValues,
		outputTypes:   ns.OutputTypes,
	}, nil
}

type Entry struct {
	defaultValues map[string]any
	outputTypes   map[string]*vo.TypeInfo
}

func (e *Entry) Invoke(_ context.Context, input map[string]any) (out map[string]any, err error) {
	in := make(map[string]any, len(input))
	for k, v := range input {
		in[k] = v
	}
	for k, v := range e.defaultValues {
		if val, ok := in[k]; ok {
			tInfo := e.outputTypes[k]
			switch tInfo.Type {
			case vo.DataTypeString:
				if len(val.(string)) == 0 {
					in[k] = v
				}
			case vo.DataTypeArray:
				if len(val.([]any)) == 0 {
					in[k] = v
				}
			case vo.DataTypeObject:
				if len(val.(map[string]any)) == 0 {
					in[k] = v
				}
			}
		} else {
			in[k] = v
		}
	}

	return in, err
}
