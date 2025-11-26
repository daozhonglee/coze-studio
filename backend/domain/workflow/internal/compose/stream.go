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

// stream.go 流式数据源解析
//
// 本文件负责计算节点的完整输入源信息，特别是处理流式数据的传递。
// 主要功能：
//   - 解析节点的真实输入源（可能与 Schema 中定义的不同）
//   - 处理复合节点中的跨层引用
//   - 确定输入字段是否为流式类型

package compose

import (
	"fmt"

	"github.com/cloudwego/eino/compose"

	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/schema"
)

// GetFullSources 计算节点的完整输入源
//
// 节点的真实输入源可能与 NodeSchema 中定义的不同，原因包括：
//  1. 复合节点内部的节点可能引用父工作流中的字段，这会被路由到内部 Start 节点
//  2. 复合节点需要将输入源委托给内部工作流
//  3. 某些节点有未在 InputSources 中定义的隐式输入源
//
// 参数：
//   - s: 目标节点 Schema
//   - sc: 所属工作流 Schema
//   - dep: 预计算的依赖信息
//
// 返回值：
//   - map[string]*schema.SourceInfo: 完整的输入源映射
//   - error: 解析错误
//
// It may be different from a NodeSchema's InputSources because of the following reasons:
//  1. a inner node under composite node may refer to a field from a node in its parent workflow,
//     this is instead routed to and sourced from the inner workflow's start node.
//  2. at the same time, the composite node needs to delegate the input source to the inner workflow.
//  3. also, some node may have implicit input sources not defined in its NodeSchema's InputSources.
func GetFullSources(s *schema.NodeSchema, sc *schema.WorkflowSchema, dep *dependencyInfo) (
	map[string]*schema.SourceInfo, error) {
	fullSource := make(map[string]*schema.SourceInfo)
	var fieldInfos []vo.FieldInfo
	for _, s := range dep.staticValues {
		fieldInfos = append(fieldInfos, vo.FieldInfo{
			Path:   s.path,
			Source: vo.FieldSource{Val: s.val},
		})
	}

	for _, v := range dep.variableInfos {
		fieldInfos = append(fieldInfos, vo.FieldInfo{
			Path: v.toPath,
			Source: vo.FieldSource{
				Ref: &vo.Reference{
					VariableType: &v.varType,
					FromPath:     v.fromPath[1:],
				},
			},
		})
	}

	for f := range dep.inputsFull {
		fieldInfos = append(fieldInfos, vo.FieldInfo{
			Path: []string{""},
			Source: vo.FieldSource{Ref: &vo.Reference{
				FromNodeKey: f,
				FromPath:    []string{""},
			}},
		})
	}

	for f, ms := range dep.inputs {
		for _, m := range ms {
			fieldInfos = append(fieldInfos, vo.FieldInfo{
				Path: m.ToPath(),
				Source: vo.FieldSource{Ref: &vo.Reference{
					FromNodeKey: f,
					FromPath:    m.FromPath(),
				}},
			})
		}
	}

	for f := range dep.inputsNoDirectDependencyFull {
		fieldInfos = append(fieldInfos, vo.FieldInfo{
			Path: []string{""},
			Source: vo.FieldSource{Ref: &vo.Reference{
				FromNodeKey: f,
				FromPath:    []string{""},
			}},
		})
	}

	for f, ms := range dep.inputsNoDirectDependency {
		for _, m := range ms {
			fieldInfos = append(fieldInfos, vo.FieldInfo{
				Path: m.ToPath(),
				Source: vo.FieldSource{Ref: &vo.Reference{
					FromNodeKey: f,
					FromPath:    m.FromPath(),
				}},
			})
		}
	}

	for i := range fieldInfos {
		fInfo := fieldInfos[i]
		path := fInfo.Path
		currentSource := fullSource
		var (
			tInfo    *vo.TypeInfo
			lastPath string
		)
		if len(path) > 1 {
			tInfo = s.InputTypes[path[0]]
			for j := 0; j < len(path)-1; j++ {
				if j > 0 {
					tInfo = tInfo.Properties[path[j]]
				}
				if current, ok := currentSource[path[j]]; !ok {
					currentSource[path[j]] = &schema.SourceInfo{
						IsIntermediate: true,
						FieldType:      schema.FieldNotStream,
						TypeInfo:       tInfo,
						SubSources:     make(map[string]*schema.SourceInfo),
					}
				} else if !current.IsIntermediate {
					return nil, fmt.Errorf("existing sourceInfo for path %s is not intermediate, conflict", path[:j+1])
				}

				currentSource = currentSource[path[j]].SubSources
			}

			lastPath = path[len(path)-1]
			tInfo = tInfo.Properties[lastPath]
		} else {
			lastPath = path[0]
			tInfo = s.InputTypes[lastPath]
		}

		// static values or variables
		if fInfo.Source.Ref == nil || fInfo.Source.Ref.FromNodeKey == "" {
			currentSource[lastPath] = &schema.SourceInfo{
				IsIntermediate: false,
				FieldType:      schema.FieldNotStream,
				TypeInfo:       tInfo,
			}
			continue
		}

		fromNodeKey := fInfo.Source.Ref.FromNodeKey
		var (
			streamType schema.FieldStreamType
			err        error
		)
		if len(fromNodeKey) > 0 {
			if fromNodeKey == compose.START {
				streamType = schema.FieldNotStream // TODO: set start node to not stream for now until composite node supports transform
			} else {
				fromNode := sc.GetNode(fromNodeKey)
				if fromNode == nil {
					return nil, fmt.Errorf("node %s not found", fromNodeKey)
				}
				streamType, err = nodes.IsStreamingField(fromNode, fInfo.Source.Ref.FromPath, sc)
				if err != nil {
					return nil, err
				}
			}
		}

		currentSource[lastPath] = &schema.SourceInfo{
			IsIntermediate: false,
			FieldType:      streamType,
			FromNodeKey:    fromNodeKey,
			FromPath:       fInfo.Source.Ref.FromPath,
			TypeInfo:       tInfo,
		}
	}

	return fullSource, nil
}
