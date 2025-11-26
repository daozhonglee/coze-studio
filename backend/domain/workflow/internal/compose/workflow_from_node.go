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

// workflow_from_node.go 单节点工作流
//
// 本文件提供从单个节点创建工作流的能力，用于节点调试场景。
// 与完整工作流不同，这种工作流：
//   - 没有 Entry 和 Exit 节点
//   - 直接以目标节点作为输入输出
//   - 仅支持 Invoke 模式（非流式）

package compose

import (
	"context"

	"github.com/cloudwego/eino/compose"

	workflow2 "github.com/coze-dev/coze-studio/backend/domain/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/schema"
)

// NewWorkflowFromNode 从单个节点创建工作流
//
// 用于节点调试场景，创建一个仅包含指定节点的最小工作流。
// 该工作流没有 Entry 和 Exit 节点，直接以目标节点作为执行入口。
//
// 参数：
//   - ctx: 上下文
//   - sc: 工作流 Schema（包含所有节点定义）
//   - nodeKey: 目标节点的 Key
//   - opts: 编译选项
//
// 返回值：
//   - *Workflow: 构建的工作流对象
//   - error: 构建错误
func NewWorkflowFromNode(ctx context.Context, sc *schema.WorkflowSchema, nodeKey vo.NodeKey, opts ...compose.GraphCompileOption) (
	*Workflow, error) {
	sc.Init()
	ns := sc.GetNode(nodeKey)

	wf := &Workflow{
		workflow:          compose.NewWorkflow[map[string]any, map[string]any](compose.WithGenLocalState(GenState())),
		hierarchy:         sc.Hierarchy,
		connections:       sc.Connections,
		schema:            sc,
		fromNode:          true,
		streamRun:         false, // single node run can only invoke
		requireCheckpoint: sc.RequireCheckpoint(),
		input:             ns.InputTypes,
		output:            ns.OutputTypes,
		terminatePlan:     vo.ReturnVariables,
	}

	compositeNodes := sc.GetCompositeNodes()
	processedNodeKey := make(map[vo.NodeKey]struct{})
	for i := range compositeNodes {
		cNode := compositeNodes[i]
		if err := wf.AddCompositeNode(ctx, cNode); err != nil {
			return nil, err
		}

		processedNodeKey[cNode.Parent.Key] = struct{}{}
		for _, child := range cNode.Children {
			processedNodeKey[child.Key] = struct{}{}
		}
	}

	// add all nodes other than composite nodes and their children
	for _, ns := range sc.Nodes {
		if _, ok := processedNodeKey[ns.Key]; !ok {
			if err := wf.AddNode(ctx, ns); err != nil {
				return nil, err
			}
		}
	}

	wf.End().AddInput(string(nodeKey))

	var compileOpts []compose.GraphCompileOption
	compileOpts = append(compileOpts, opts...)
	if wf.requireCheckpoint {
		compileOpts = append(compileOpts, compose.WithCheckPointStore(workflow2.GetRepository()))
	}

	r, err := wf.Compile(ctx, compileOpts...)
	if err != nil {
		return nil, err
	}
	wf.Runner = r

	return wf, nil
}
