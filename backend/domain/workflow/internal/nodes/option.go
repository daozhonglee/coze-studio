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

// option.go 定义了节点执行时的配置选项
//
// 本文件提供了节点选项的统一抽象，主要用于：
// - 嵌套工作流的状态管理和恢复
// - 节点特定选项的传递
// - 批量/循环节点中索引相关的配置
//
// 选项模式采用函数式选项模式（Functional Options Pattern），
// 允许灵活组合和扩展节点行为。
package nodes

import (
	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/compose"

	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
)

// NodeOptions 节点选项集合
// 包含通用选项和嵌套工作流相关的特殊选项
type NodeOptions struct {
	// Nested 嵌套工作流选项，用于子工作流、循环等场景
	Nested *NestedWorkflowOptions
}

// NestedWorkflowOptions 嵌套工作流选项
// 用于管理子工作流、批量处理、循环等嵌套执行场景的状态
type NestedWorkflowOptions struct {
	// optsForNested 传递给嵌套工作流的 compose 选项
	optsForNested []compose.Option
	// toResumeIndexes 需要恢复执行的索引及其状态修改器
	// 用于中断后恢复执行的场景
	toResumeIndexes map[int]compose.StateModifier
	// optsForIndexed 按索引分组的选项，用于批量处理场景
	optsForIndexed map[int][]compose.Option
}

// NodeOption 节点选项函数类型
// 采用函数式选项模式，支持链式配置
type NodeOption struct {
	// apply 应用到 NodeOptions 的函数
	apply func(opts *NodeOptions)
	// implSpecificOptFn 节点实现特定的选项函数
	// 允许各节点类型定义自己的选项
	implSpecificOptFn any
}

type NestedWorkflowOption func(*NestedWorkflowOptions)

func WithOptsForNested(opts ...compose.Option) NodeOption {
	return NodeOption{
		apply: func(options *NodeOptions) {
			if options.Nested == nil {
				options.Nested = &NestedWorkflowOptions{}
			}
			options.Nested.optsForNested = append(options.Nested.optsForNested, opts...)
		},
	}
}

func (c *NodeOptions) GetOptsForNested() []compose.Option {
	if c.Nested == nil {
		return nil
	}
	return c.Nested.optsForNested
}

func WithResumeIndex(i int, m compose.StateModifier) NodeOption {
	return NodeOption{
		apply: func(options *NodeOptions) {
			if options.Nested == nil {
				options.Nested = &NestedWorkflowOptions{}
			}
			if options.Nested.toResumeIndexes == nil {
				options.Nested.toResumeIndexes = map[int]compose.StateModifier{}
			}

			options.Nested.toResumeIndexes[i] = m
		},
	}
}

func (c *NodeOptions) GetResumeIndexes() map[int]compose.StateModifier {
	if c.Nested == nil {
		return nil
	}
	return c.Nested.toResumeIndexes
}

func WithOptsForIndexed(index int, opts ...compose.Option) NodeOption {
	return NodeOption{
		apply: func(options *NodeOptions) {
			if options.Nested == nil {
				options.Nested = &NestedWorkflowOptions{}
			}
			if options.Nested.optsForIndexed == nil {
				options.Nested.optsForIndexed = map[int][]compose.Option{}
			}
			options.Nested.optsForIndexed[index] = opts
		},
	}
}

func (c *NodeOptions) GetOptsForIndexed(index int) []compose.Option {
	if c.Nested == nil {
		return nil
	}
	return c.Nested.optsForIndexed[index]
}

func (c *NodeOptions) HasIndexedOpts() bool {
	return c.Nested != nil && len(c.Nested.optsForIndexed) > 0
}

func (c *NodeOptions) HasOptsForIndex(index int) bool {
	if c.Nested == nil || c.Nested.optsForIndexed == nil {
		return false
	}
	_, ok := c.Nested.optsForIndexed[index]
	return ok
}

// WrapImplSpecificOptFn is the option to wrap the implementation specific option function.
func WrapImplSpecificOptFn[T any](optFn func(*T)) NodeOption {
	return NodeOption{
		implSpecificOptFn: optFn,
	}
}

// GetCommonOptions extract model Options from Option list, optionally providing a base Options with default values.
func GetCommonOptions(base *NodeOptions, opts ...NodeOption) *NodeOptions {
	if base == nil {
		base = &NodeOptions{}
	}

	for i := range opts {
		opt := opts[i]
		if opt.apply != nil {
			opt.apply(base)
		}
	}

	return base
}

// GetImplSpecificOptions extract the implementation specific options from Option list, optionally providing a base options with default values.
// e.g.
//
//	myOption := &MyOption{
//		Field1: "default_value",
//	}
//
//	myOption := model.GetImplSpecificOptions(myOption, opts...)
func GetImplSpecificOptions[T any](base *T, opts ...NodeOption) *T {
	if base == nil {
		base = new(T)
	}

	for i := range opts {
		opt := opts[i]
		if opt.implSpecificOptFn != nil {
			optFn, ok := opt.implSpecificOptFn.(func(*T))
			if ok {
				optFn(base)
			}
		}
	}

	return base
}

// NestedWorkflowState 嵌套工作流状态
// 用于持久化和恢复嵌套工作流的执行状态，支持中断恢复场景
type NestedWorkflowState struct {
	// Index2Done 各索引的完成状态，用于批量/循环场景追踪进度
	Index2Done map[int]bool `json:"index_2_done,omitempty"`
	// Index2InterruptInfo 各索引的中断信息，用于恢复中断的执行
	Index2InterruptInfo map[int]*compose.InterruptInfo `json:"index_2_interrupt_info,omitempty"`
	// FullOutput 完整输出结果
	FullOutput map[string]any `json:"full_output,omitempty"`
	// IntermediateVars 中间变量，用于跨迭代传递状态
	IntermediateVars map[string]any `json:"intermediate_vars,omitempty"`
}

// String 返回状态的 JSON 字符串表示，便于调试和日志记录
func (c *NestedWorkflowState) String() string {
	s, _ := sonic.MarshalIndent(c, "", "  ")
	return string(s)
}

// NestedWorkflowAware 嵌套工作流状态感知接口
// 实现此接口的组件可以保存和读取嵌套工作流的执行状态
// 主要用于支持工作流的中断恢复功能
type NestedWorkflowAware interface {
	// SaveNestedWorkflowState 保存嵌套工作流状态
	// 参数 key: 节点唯一标识
	// 参数 state: 要保存的状态
	SaveNestedWorkflowState(key vo.NodeKey, state *NestedWorkflowState) error
	// GetNestedWorkflowState 获取嵌套工作流状态
	// 返回: 状态、是否存在、错误
	GetNestedWorkflowState(key vo.NodeKey) (*NestedWorkflowState, bool, error)
}
