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

// node_builder.go 节点构建器接口
//
// 本文件定义了节点构建相关的接口和选项。
// NodeBuilder 和 BranchBuilder 是节点实现的核心接口。

package schema

import (
	"context"

	"github.com/cloudwego/eino/compose"
)

// BuildOptions 构建选项
type BuildOptions struct {
	WS    *WorkflowSchema                                  // 所属工作流 Schema
	Inner compose.Runnable[map[string]any, map[string]any] // 内部工作流（复合节点）
}

// GetBuildOptions 解析构建选项
func GetBuildOptions(opts ...BuildOption) *BuildOptions {
	bo := &BuildOptions{}
	for _, o := range opts {
		o(bo)
	}
	return bo
}

// BuildOption 构建选项函数类型
type BuildOption func(options *BuildOptions)

// WithWorkflowSchema 设置所属工作流 Schema
func WithWorkflowSchema(ws *WorkflowSchema) BuildOption {
	return func(options *BuildOptions) {
		options.WS = ws
	}
}

// WithInnerWorkflow 设置内部工作流（用于复合节点）
func WithInnerWorkflow(inner compose.Runnable[map[string]any, map[string]any]) BuildOption {
	return func(options *BuildOptions) {
		options.Inner = inner
	}
}

// NodeBuilder 节点构建器接口
//
// 负责将 NodeSchema 转换为可执行的节点实例。
// 返回的 executable 必须实现以下接口之一：
//   - nodes.InvokableNode: 同步调用
//   - nodes.StreamableNode: 流式输出
//   - nodes.CollectableNode: 收集流式输入
//   - nodes.TransformableNode: 流式转换
//
// 带 WOpt 后缀的版本支持 NodeOption 参数。
// 注意：节点应实现普通版本或 WOpt 版本，不要混用。
type NodeBuilder interface {
	Build(ctx context.Context, ns *NodeSchema, opts ...BuildOption) (
		executable any, err error)
}

// BranchBuilder 分支构建器接口
//
// 负责构建从节点输出到分支端口的映射函数。
// 返回的 extractor 函数根据节点输出返回分支索引。
type BranchBuilder interface {
	BuildBranch(ctx context.Context) (extractor func(ctx context.Context,
		nodeOutput map[string]any) (int64, bool /*if is default branch*/, error), hasBranch bool)
}
