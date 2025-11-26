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

// node_builder.go 节点构建器
//
// 本文件负责将 NodeSchema 转换为可执行的 Node 对象。
// 主要功能：
//   - 从 NodeSchema 实例化具体的节点类型
//   - 处理节点的输入源解析
//   - 支持自定义 NodeBuilder 接口
//   - 特殊处理 Lambda 和 SubWorkflow 类型节点

package compose

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/cloudwego/eino/compose"

	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/execute"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/subworkflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/schema"
	"github.com/coze-dev/coze-studio/backend/pkg/errorx"
	"github.com/coze-dev/coze-studio/backend/pkg/safego"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

// Node 节点包装结构
// 包含 Eino 框架的 Lambda 对象，代表一个可执行的节点
type Node struct {
	Lambda *compose.Lambda
}

// New 从 NodeSchema 实例化节点
//
// 这是节点创建的核心函数，负责将节点的声明式定义转换为可执行对象。
// 支持以下节点类型：
//   - 实现 NodeBuilder 接口的节点：调用 Build 方法构建
//   - Lambda 节点：直接使用预定义的 Lambda
//   - SubWorkflow 节点：构建嵌套的子工作流
//
// 参数：
//   - ctx: 上下文
//   - s: 节点 Schema
//   - inner: 复合节点的内部工作流
//   - sc: 所属工作流 Schema
//   - deps: 预计算的节点依赖信息
//   - requireCheckpoint: 是否需要检查点支持
func New(ctx context.Context, s *schema.NodeSchema,
	inner compose.Runnable[map[string]any, map[string]any], // inner workflow for composite node
	sc *schema.WorkflowSchema, // the workflow this NodeSchema is in
	deps *dependencyInfo, // the dependency for this node pre-calculated by workflow engine
	requireCheckpoint bool,
) (_ *Node, err error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = safego.NewPanicErr(panicErr, debug.Stack())
		}

		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrCreateNodeFail, err, errorx.KV("node_name", s.Name), errorx.KV("cause", err.Error()))
		}
	}()

	var fullSources map[string]*schema.SourceInfo
	if m := entity.NodeMetaByNodeType(s.Type); m != nil && m.InputSourceAware {
		if fullSources, err = GetFullSources(s, sc, deps); err != nil {
			return nil, err
		}
		s.FullSources = fullSources
	}

	// if NodeSchema's Configs implements NodeBuilder, will use it to build the node
	nb, ok := s.Configs.(schema.NodeBuilder)
	if ok {
		opts := []schema.BuildOption{
			schema.WithWorkflowSchema(sc),
			schema.WithInnerWorkflow(inner),
		}

		// build the actual InvokableNode, etc.
		n, err := nb.Build(ctx, s, opts...)
		if err != nil {
			return nil, err
		}

		// wrap InvokableNode, etc. within NodeRunner, converting to eino's Lambda
		return toNode(s, n), nil
	}

	switch s.Type {
	case entity.NodeTypeLambda:
		if s.Lambda == nil {
			return nil, fmt.Errorf("lambda is not defined for NodeTypeLambda")
		}

		return &Node{Lambda: s.Lambda}, nil
	case entity.NodeTypeSubWorkflow:
		subWorkflow, err := buildSubWorkflow(ctx, s, requireCheckpoint)
		if err != nil {
			return nil, err
		}

		return toNode(s, subWorkflow), nil
	default:
		panic(fmt.Sprintf("node schema's Configs does not implement NodeBuilder. type: %v", s.Type))
	}
}

func buildSubWorkflow(ctx context.Context, s *schema.NodeSchema, requireCheckpoint bool) (*subworkflow.SubWorkflow, error) {
	var opts []WorkflowOption
	opts = append(opts, WithIDAsName(s.Configs.(*subworkflow.Config).WorkflowID))
	if requireCheckpoint {
		opts = append(opts, WithParentRequireCheckpoint())
	}
	if s := execute.GetStaticConfig(); s != nil && s.MaxNodeCountPerWorkflow > 0 {
		opts = append(opts, WithMaxNodeCount(s.MaxNodeCountPerWorkflow))
	}
	wf, err := NewWorkflow(ctx, s.SubWorkflowSchema, opts...)
	if err != nil {
		return nil, err
	}

	return &subworkflow.SubWorkflow{
		Runner: wf.Runner,
	}, nil
}
