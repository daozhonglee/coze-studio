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

// Package adaptor 提供画布到工作流schema的转换功能
//
// 这个包是工作流系统的核心转换器，负责将前端的可视化画布
// 转换为后端可执行的工作流schema，为执行引擎提供运行时配置。
//
// 主要功能：
// 1. Canvas到WorkflowSchema的转换
// 2. 节点适配器的管理和注册
// 3. 批量模式的特殊处理
// 4. 孤立节点的剪枝优化
// 5. 端口标准化和连接关系建立
//
// 转换流程：
// 前端Canvas → 适配器转换 → 执行Schema → 工作流引擎
//
// 关键概念：
// - Node: 前端可视化节点
// - NodeSchema: 执行引擎可理解的节点定义
// - Connection: 节点间的执行依赖关系
// - Branch: 条件分支逻辑
package adaptor

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

	einoCompose "github.com/cloudwego/eino/compose"

	model "github.com/coze-dev/coze-studio/backend/crossdomain/workflow/model"
	"github.com/coze-dev/coze-studio/backend/domain/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/canvas/convert"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/batch"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/code"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/conversation"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/database"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/emitter"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/entry"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/exit"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/httprequester"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/intentdetector"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/json"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/knowledge"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/llm"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/loop"
	_break "github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/loop/break"
	_continue "github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/loop/continue"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/plugin"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/qa"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/receiver"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/selector"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/subworkflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/textprocessor"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/variableaggregator"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/variableassigner"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/schema"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/slices"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ternary"
	"github.com/coze-dev/coze-studio/backend/pkg/safego"
	"github.com/coze-dev/coze-studio/backend/pkg/sonic"
)

// CanvasToWorkflowSchema 将前端画布转换为工作流执行schema
//
// 这是整个转换流程的核心函数，负责将用户在前端创建的可视化工作流
// 转换为后端执行引擎能够理解和执行的数据结构。
//
// 转换步骤：
// 1. 清理孤立节点和无效连接
// 2. 处理批量模式节点
// 3. 转换节点为NodeSchema
// 4. 建立节点层级关系
// 5. 转换边为连接关系
// 6. 标准化端口信息
// 7. 构建分支逻辑
// 8. 初始化schema
//
// 参数：
//   - ctx: 上下文，用于传递请求信息和取消信号
//   - s: 前端画布对象，包含节点和连接信息
//
// 返回：
//   - sc: 转换后的工作流schema，可被执行引擎使用
//   - err: 转换过程中的错误
//
// 注意：
//   - 函数包含panic恢复机制，确保转换过程中的异常不会影响整个系统
//   - 不支持嵌套的内部工作流
//   - 会自动清理不可达的孤立节点
func CanvasToWorkflowSchema(ctx context.Context, s *vo.Canvas) (sc *schema.WorkflowSchema, err error) {
	// panic恢复：防止转换过程中的异常影响整个系统
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = safego.NewPanicErr(panicErr, debug.Stack())
		}
	}()

	// 第一步：清理孤立节点，只保留连接的节点和边
	connectedNodes, connectedEdges := PruneIsolatedNodes(s.Nodes, s.Edges, nil)
	s = &vo.Canvas{
		Nodes: connectedNodes,
		Edges: connectedEdges,
	}

	// 初始化schema对象
	sc = &schema.WorkflowSchema{}

	// 构建节点映射，用于快速查找
	nodeMap := make(map[string]*vo.Node)

	// 遍历所有节点进行转换
	for i, node := range s.Nodes {
		nodeMap[node.ID] = s.Nodes[i]

		// 处理复合节点的子节点
		for j, subNode := range node.Blocks {
			nodeMap[subNode.ID] = node.Blocks[j]
			subNode.SetParent(node)

			// 不支持嵌套的内部工作流
			if len(subNode.Blocks) > 0 {
				return nil, fmt.Errorf("nested inner-workflow is not supported")
			}

			// 子节点不应有边信息
			if len(subNode.Edges) > 0 {
				return nil, fmt.Errorf("nodes in inner-workflow should not have edges info")
			}

			// 处理break和continue节点，连接到父节点
			if subNode.Type == entity.NodeTypeBreak.IDStr() || subNode.Type == entity.NodeTypeContinue.IDStr() {
				sc.Connections = append(sc.Connections, &schema.Connection{
					FromNode: vo.NodeKey(subNode.ID),
					ToNode:   vo.NodeKey(subNode.Parent().ID),
				})
			}
		}

		// 处理批量模式节点转换
		newNode, enableBatch, err := parseBatchMode(node)
		if err != nil {
			return nil, err
		}

		if enableBatch {
			node = newNode
			// 记录生成的内部节点ID
			sc.GeneratedNodes = append(sc.GeneratedNodes, vo.NodeKey(node.Blocks[0].ID))
		}

		// 将节点转换为NodeSchema
		nsList, hierarchy, err := NodeToNodeSchema(ctx, node, s)
		if err != nil {
			return nil, err
		}

		// 添加转换后的节点
		sc.Nodes = append(sc.Nodes, nsList...)

		// 建立层级关系映射
		if len(hierarchy) > 0 {
			if sc.Hierarchy == nil {
				sc.Hierarchy = make(map[vo.NodeKey]vo.NodeKey)
			}

			for k, v := range hierarchy {
				sc.Hierarchy[k] = v
			}
		}

		// 处理节点内部的边（复合节点的情况）
		for _, edge := range node.Edges {
			sc.Connections = append(sc.Connections, EdgeToConnection(edge))
		}
	}

	// 处理画布级别的边
	for _, edge := range s.Edges {
		sc.Connections = append(sc.Connections, EdgeToConnection(edge))
	}

	// 标准化端口信息
	newConnections, err := normalizePorts(sc.Connections, nodeMap)
	if err != nil {
		return nil, err
	}
	sc.Connections = newConnections

	// 基于连接关系构建分支逻辑
	branches, err := schema.BuildBranches(newConnections)
	if err != nil {
		return nil, err
	}
	sc.Branches = branches

	// 初始化schema，构建内部索引和缓存
	sc.Init()

	return sc, nil
}

// normalizePorts 标准化连接中的端口信息
//
// 不同类型的节点有不同的端口命名约定，这个函数负责将前端的端口标识
// 转换为执行引擎能够理解的标准化格式。
//
// 主要处理：
// 1. 移除空的端口标识
// 2. 处理循环和批量节点的内部端口（设为nil）
// 3. 标准化选择器节点的条件分支端口
//
// 参数：
//   - connections: 原始连接列表
//   - nodeMap: 节点ID到节点对象的映射
//
// 返回：
//   - normalized: 标准化后的连接列表
//   - err: 处理过程中的错误
//
// 端口标准化规则：
//   - 空端口设为nil
//   - 循环/批量节点的内部端口设为nil（执行引擎不需要）
//   - 选择器节点：
//   - "true" → "branch_0"
//   - "false" → "default"
//   - "true_N" → "branch_N"
func normalizePorts(connections []*schema.Connection, nodeMap map[string]*vo.Node) (normalized []*schema.Connection, err error) {
	for i := range connections {
		conn := connections[i]

		// 处理无端口的情况，直接保留
		if conn.FromPort == nil {
			normalized = append(normalized, conn)
			continue
		}

		// 处理空端口标识
		if len(*conn.FromPort) == 0 {
			conn.FromPort = nil
			normalized = append(normalized, conn)
			continue
		}

		// 处理循环和批量节点的内部端口，这些对执行引擎不可见
		if *conn.FromPort == "loop-function-inline-output" || *conn.FromPort == "loop-output" ||
			*conn.FromPort == "batch-function-inline-output" || *conn.FromPort == "batch-output" {
			conn.FromPort = nil
			normalized = append(normalized, conn)
			continue
		}

		// 根据节点类型进行端口标准化
		node, ok := nodeMap[string(conn.FromNode)]
		if !ok {
			return nil, fmt.Errorf("node %s not found in node map", conn.FromNode)
		}

		var newPort string
		switch node.Type {
		case entity.NodeTypeSelector.IDStr():
			// 选择器节点的条件分支端口标准化
			if *conn.FromPort == "true" {
				newPort = fmt.Sprintf(schema.PortBranchFormat, 0) // "branch_0"
			} else if *conn.FromPort == "false" {
				newPort = schema.PortDefault // "default"
			} else if strings.HasPrefix(*conn.FromPort, "true_") {
				// 处理多分支情况，如 "true_1" -> "branch_1"
				portN := strings.TrimPrefix(*conn.FromPort, "true_")
				n, err := strconv.Atoi(portN)
				if err != nil {
					return nil, fmt.Errorf("invalid port name: %s", *conn.FromPort)
				}
				newPort = fmt.Sprintf(schema.PortBranchFormat, n)
			}
		default:
			// 其他节点类型保持原端口标识
			newPort = *conn.FromPort
		}

		normalized = append(normalized, &schema.Connection{
			FromNode: conn.FromNode,
			ToNode:   conn.ToNode,
			FromPort: &newPort,
		})
	}

	return normalized, nil
}

var blockTypeToSkip = map[entity.NodeType]bool{
	entity.NodeTypeComment: true,
}

// NodeToNodeSchema 将单个前端节点转换为执行schema
//
// 这是节点转换的核心函数，根据节点类型选择合适的适配器进行转换。
// 支持普通节点、子工作流节点和复合节点（包含子节点）的转换。
//
// 转换策略：
// 1. 子工作流节点：直接转换为子工作流schema
// 2. 普通节点：通过节点适配器进行转换
// 3. 复合节点：递归转换子节点，建立层级关系
// 4. 跳过节点：注释等不需要执行的节点类型
//
// 参数：
//   - ctx: 上下文
//   - n: 要转换的节点
//   - c: 完整的画布上下文
//
// 返回：
//   - []*schema.NodeSchema: 转换后的节点schema列表
//   - map[vo.NodeKey]vo.NodeKey: 子节点到父节点的层级映射
//   - error: 转换错误
//
// 注意：
//   - 复合节点会展开为多个schema节点
//   - 返回的层级映射用于建立节点关系
func NodeToNodeSchema(ctx context.Context, n *vo.Node, c *vo.Canvas) ([]*schema.NodeSchema, map[vo.NodeKey]vo.NodeKey, error) {
	et := entity.IDStrToNodeType(n.Type)

	// 特殊处理：子工作流节点
	if et == entity.NodeTypeSubWorkflow {
		ns, err := toSubWorkflowNodeSchema(ctx, n)
		if err != nil {
			return nil, nil, err
		}
		if ns.ExceptionConfigs, err = toExceptionConfig(n, ns.Type); err != nil {
			return nil, nil, err
		}
		return []*schema.NodeSchema{ns}, nil, nil
	}

	// 通过节点适配器进行标准转换
	na, ok := nodes.GetNodeAdaptor(et)
	if ok {
		ns, err := na.Adapt(ctx, n, nodes.WithCanvas(c))
		if err != nil {
			return nil, nil, err
		}

		if ns.ExceptionConfigs, err = toExceptionConfig(n, ns.Type); err != nil {
			return nil, nil, err
		}

		// 处理复合节点：递归转换子节点
		if len(n.Blocks) > 0 {
			var (
				allNS     []*schema.NodeSchema
				hierarchy = make(map[vo.NodeKey]vo.NodeKey)
			)

			for _, childN := range n.Blocks {
				childN.SetParent(n)
				childNS, _, err := NodeToNodeSchema(ctx, childN, c)
				if err != nil {
					return nil, nil, err
				}

				allNS = append(allNS, childNS...)
				hierarchy[vo.NodeKey(childN.ID)] = vo.NodeKey(n.ID)
			}

			allNS = append(allNS, ns)
			return allNS, hierarchy, nil
		}

		return []*schema.NodeSchema{ns}, nil, nil
	}

	// 跳过不需要执行的节点类型（如注释）
	_, ok = blockTypeToSkip[et]
	if ok {
		return nil, nil, nil
	}

	return nil, nil, fmt.Errorf("unsupported block type: %v", n.Type)
}

// EdgeToConnection 将前端边转换为执行连接关系
//
// 前端边定义了可视化的连接关系，这个函数将其转换为执行引擎
// 能够理解的连接对象，包括特殊的循环和批量节点处理。
//
// 特殊处理：
// - 循环节点的内联输入：转换为END标识符
// - 批量节点的内联输入：转换为END标识符
// - 源端口标识：保留用于后续标准化
//
// 参数：
//   - e: 前端边对象
//
// 返回：
//   - *schema.Connection: 执行引擎使用的连接对象
//
// 注意：
//   - END标识符表示流程的结束点
//   - FromPort可能为nil，后续会通过normalizePorts标准化
func EdgeToConnection(e *vo.Edge) *schema.Connection {
	toNode := vo.NodeKey(e.TargetNodeID)

	// 特殊处理：循环和批量节点的内联输入连接到流程结束
	if len(e.SourcePortID) > 0 && (e.TargetPortID == "loop-function-inline-input" || e.TargetPortID == "batch-function-inline-input") {
		toNode = einoCompose.END
	}

	conn := &schema.Connection{
		FromNode: vo.NodeKey(e.SourceNodeID),
		ToNode:   toNode,
	}

	// 设置源端口标识（如果有的话）
	if len(e.SourceNodeID) > 0 {
		conn.FromPort = &e.SourcePortID
	}

	return conn
}

// toExceptionConfig 转换节点的异常处理配置
//
// 将前端节点的错误处理设置转换为执行引擎能够理解的异常配置。
// 包括超时时间、重试次数、错误数据等异常处理策略。
//
// 配置优先级：
// 1. 用户自定义设置（最高优先级）
// 2. 节点默认配置（最低优先级）
//
// 参数：
//   - n: 节点对象
//   - nType: 节点类型
//
// 返回：
//   - *schema.ExceptionConfig: 异常处理配置
//   - error: 配置错误
//
// 验证规则：
//   - 如果选择返回默认数据，必须提供DataOnErr
//   - 如果未设置ProcessType但提供了DataOnErr且开启了开关，自动设置为返回默认数据
func toExceptionConfig(n *vo.Node, nType entity.NodeType) (*schema.ExceptionConfig, error) {
	nodeMeta := entity.NodeMetaByNodeType(nType)

	var settingOnErr *vo.SettingOnError

	// 获取用户自定义的错误处理设置
	if n.Data.Inputs != nil {
		settingOnErr = n.Data.Inputs.SettingOnError
	}

	// 如果没有自定义设置且节点没有默认超时，直接返回nil
	if settingOnErr == nil && nodeMeta.DefaultTimeoutMS == 0 {
		return nil, nil
	}

	// 使用节点默认配置作为基础
	metaConf := &schema.ExceptionConfig{
		TimeoutMS: nodeMeta.DefaultTimeoutMS,
	}

	// 应用用户自定义设置
	if settingOnErr != nil {
		metaConf = &schema.ExceptionConfig{
			TimeoutMS:   settingOnErr.TimeoutMs,
			MaxRetry:    settingOnErr.RetryTimes,
			DataOnErr:   settingOnErr.DataOnErr,
			ProcessType: settingOnErr.ProcessType,
		}

		// 验证：如果选择返回默认数据，必须提供错误数据
		if metaConf.ProcessType != nil && *metaConf.ProcessType == vo.ErrorProcessTypeReturnDefaultData {
			if len(metaConf.DataOnErr) == 0 {
				return nil, errors.New("error process type is returning default value, but dataOnError is not specified")
			}
		}

		// 自动推断处理类型：如果有错误数据且开启了开关，设置为返回默认数据
		if metaConf.ProcessType == nil && len(metaConf.DataOnErr) > 0 && settingOnErr.Switch {
			metaConf.ProcessType = ptr.Of(vo.ErrorProcessTypeReturnDefaultData)
		}
	}

	return metaConf, nil
}

// toSubWorkflowNodeSchema 转换子工作流节点为执行schema
//
// 子工作流节点引用另一个工作流作为子流程，这个函数负责：
// 1. 解析引用的工作流ID和版本
// 2. 加载子工作流的定义
// 3. 递归转换子工作流为schema
// 4. 创建子工作流节点配置
//
// 转换流程：
// 1. 解析工作流ID和版本信息
// 2. 从仓库加载子工作流实体
// 3. 反序列化子工作流的画布
// 4. 递归调用CanvasToWorkflowSchema转换子工作流
// 5. 创建子工作流节点schema
// 6. 设置输入输出参数类型
//
// 参数：
//   - ctx: 上下文
//   - n: 子工作流节点
//
// 返回：
//   - *schema.NodeSchema: 子工作流节点schema
//   - error: 转换错误
//
// 注意：
//   - 子工作流会递归初始化（subWorkflowSC.Init()）
//   - 支持流式输出配置
//   - 版本控制：支持草稿版本和发布版本
func toSubWorkflowNodeSchema(ctx context.Context, n *vo.Node) (*schema.NodeSchema, error) {
	// 解析引用的工作流ID
	idStr := n.Data.Inputs.WorkflowID
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow id: %w", err)
	}

	version := n.Data.Inputs.WorkflowVersion

	// 根据是否有版本信息决定查询类型
	queryType := ternary.IFElse(len(version) == 0, model.FromDraft, model.FromSpecificVersion)

	// 从仓库加载子工作流实体
	subWF, err := workflow.GetRepository().GetEntity(ctx, &vo.GetPolicy{
		ID:      id,
		QType:   queryType,
		Version: version,
	})
	if err != nil {
		return nil, err
	}

	// 反序列化子工作流的画布
	var subCanvas vo.Canvas
	if err = sonic.UnmarshalString(subWF.Canvas, &subCanvas); err != nil {
		return nil, err
	}

	// 递归转换子工作流为schema
	subWorkflowSC, err := CanvasToWorkflowSchema(ctx, &subCanvas)
	if err != nil {
		return nil, err
	}

	// 初始化子工作流schema（构建索引和缓存）
	subWorkflowSC.Init()

	// 创建子工作流配置
	cfg := &subworkflow.Config{}

	// 构建节点schema
	ns := &schema.NodeSchema{
		Key:               vo.NodeKey(n.ID),
		Type:              entity.NodeTypeSubWorkflow,
		Name:              n.Data.Meta.Title,
		SubWorkflowBasic:  subWF.GetBasic(), // 基本信息（ID、版本等）
		SubWorkflowSchema: subWorkflowSC,    // 完整的子工作流schema
		Configs:           cfg,
	}

	// 配置流式输出能力
	ns.StreamConfigs = &schema.StreamConfig{
		CanGeneratesStream: subWorkflowSC.RequireStreaming(),
	}

	// 再次验证并设置工作流ID
	workflowIDStr := n.Data.Inputs.WorkflowID
	if workflowIDStr == "" {
		return nil, fmt.Errorf("sub workflow node's workflowID is empty")
	}
	workflowID, err := strconv.ParseInt(workflowIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("sub workflow node's workflowID is not a number: %s", workflowIDStr)
	}
	cfg.WorkflowID = workflowID
	cfg.WorkflowVersion = n.Data.Inputs.WorkflowVersion

	// 设置输入参数类型
	if err := convert.SetInputsForNodeSchema(n, ns); err != nil {
		return nil, err
	}

	// 设置输出参数类型
	if err := convert.SetOutputTypesForNodeSchema(n, ns); err != nil {
		return nil, err
	}

	return ns, nil
}

// PruneIsolatedNodes 剪枝孤立节点
//
// 清理工作流画布中不可达的孤立节点和无效边，确保只保留
// 能通过连接关系访问到的节点，提高执行效率和准确性。
//
// 算法原理：
// 基于图论的可达性分析，通过依赖计数识别孤立节点。
//
// 处理步骤：
// 1. 递归处理复合节点的子节点
// 2. 计算每个节点的依赖计数
// 3. 标记入口和出口节点为可达
// 4. 根据边关系更新依赖计数
// 5. 识别并移除孤立节点
//
// 特殊处理：
// - 入口和出口节点始终被认为是可达的
// - break/continue节点依赖于父节点
// - 复合节点递归处理
//
// 参数：
//   - nodes: 原始节点列表
//   - edges: 原始边列表
//   - parentNode: 父节点（用于复合节点处理）
//
// 返回：
//   - []*vo.Node: 剪枝后的节点列表
//   - []*vo.Edge: 剪枝后的边列表
//
// 注意：
//   - 这个函数会修改传入的节点结构
//   - 递归处理嵌套的复合节点
func PruneIsolatedNodes(nodes []*vo.Node, edges []*vo.Edge, parentNode *vo.Node) ([]*vo.Node, []*vo.Edge) {
	// 初始化节点依赖计数映射
	nodeDependencyCount := map[string]int{}
	if parentNode != nil {
		nodeDependencyCount[parentNode.ID] = 0
	}

	// 第一遍遍历：递归处理复合节点，初始化依赖计数
	for _, node := range nodes {
		// 递归处理复合节点的子节点
		if len(node.Blocks) > 0 && len(node.Edges) > 0 {
			node.Blocks, node.Edges = PruneIsolatedNodes(node.Blocks, node.Edges, node)
		}

		// 初始化节点的依赖计数
		nodeDependencyCount[node.ID] = 0

		// break/continue节点依赖于父节点
		if node.Type == entity.NodeTypeContinue.IDStr() || node.Type == entity.NodeTypeBreak.IDStr() {
			if parentNode != nil {
				nodeDependencyCount[parentNode.ID]++
			}
		}
	}

	// 标记入口和出口节点为可达（依赖计数设为1）
	nodeDependencyCount[entity.EntryNodeKey] = 1
	nodeDependencyCount[entity.ExitNodeKey] = 1

	// 第二遍遍历：根据边关系更新依赖计数
	for _, edge := range edges {
		if _, ok := nodeDependencyCount[edge.TargetNodeID]; ok {
			nodeDependencyCount[edge.TargetNodeID]++
		} else {
			panic(fmt.Errorf("node id %v not existed, but appears in the edge", edge.TargetNodeID))
		}
	}

	// 识别孤立节点（依赖计数为0的节点）
	isolatedNodeIDs := make(map[string]struct{})
	for nodeId, count := range nodeDependencyCount {
		if count == 0 {
			isolatedNodeIDs[nodeId] = struct{}{}
		}
	}

	// 过滤掉孤立节点
	connectedNodes := make([]*vo.Node, 0)
	for _, node := range nodes {
		if _, ok := isolatedNodeIDs[node.ID]; !ok {
			connectedNodes = append(connectedNodes, node)
		}
	}

	// 过滤掉连接到孤立节点的边
	connectedEdges := make([]*vo.Edge, 0)
	for _, edge := range edges {
		if _, ok := isolatedNodeIDs[edge.SourceNodeID]; !ok {
			connectedEdges = append(connectedEdges, edge)
		}
	}

	return connectedNodes, connectedEdges
}

// parseBatchMode 解析批量模式节点
//
// 将启用了批量处理的节点转换为特殊的批量节点结构。
// 批量模式允许对数组数据进行并行或批处理，提高执行效率。
//
// 批量模式转换原理：
// 原始节点 → 批量父节点 + 内部处理节点
//
// 转换逻辑：
// 1. 验证批量模式配置和输出格式
// 2. 提取输出参数的元素类型定义
// 3. 创建批量父节点（负责数据分批和聚合）
// 4. 创建内部处理节点（实际的业务逻辑）
// 5. 建立父子节点间的连接关系
//
// 参数：
//   - n: 原始节点
//
// 返回：
//   - batchN: 转换后的批量节点（如果启用批量模式）
//   - enabled: 是否启用了批量模式
//   - err: 转换错误
//
// 批量节点结构：
// ```
// 批量父节点 (Batch Node)
// ├── 输入：数组数据 + 批大小 + 并发数
// ├── 输出：处理结果数组
// └── 子节点：内部处理节点
//
//	├── 输入：单个数据项
//	└── 输出：单个处理结果
//
// ```
//
// 注意：
//   - 批量模式的输出必须是数组类型
//   - 数组元素必须是对象类型
//   - 会生成唯一的内部节点ID
func parseBatchMode(n *vo.Node) (
	batchN *vo.Node, // 转换后的批量节点
	enabled bool, // 是否启用了批量模式
	err error) {

	// 检查节点是否有批量模式配置
	if n.Data == nil || n.Data.Inputs == nil {
		return nil, false, nil
	}

	batchInfo := n.Data.Inputs.NodeBatchInfo
	if batchInfo == nil || !batchInfo.BatchEnable {
		return nil, false, nil // 未启用批量模式
	}

	enabled = true

	// 准备变量定义
	var (
		innerOutput []*vo.Variable                           // 内部节点的输出参数
		outerOutput []*vo.Param                              // 批量节点的输出参数
		innerInput  = n.Data.Inputs.InputParameters          // 内部节点的输入（来自原始节点）
		outerInput  = n.Data.Inputs.NodeBatchInfo.InputLists // 批量节点的输入列表
	)

	// 验证输出格式：批量模式必须有且只有一个输出，且为数组类型
	if len(n.Data.Outputs) != 1 {
		return nil, false, fmt.Errorf("node batch mode output should be one list, actual count: %d", len(n.Data.Outputs))
	}

	out := n.Data.Outputs[0] // 获取输出定义

	// 解析输出变量
	v, err := vo.ParseVariable(out)
	if err != nil {
		return nil, false, err
	}

	// 验证输出类型必须是数组
	if v.Type != vo.VariableTypeList {
		return nil, false, fmt.Errorf("node batch mode output should be list, actual type: %s", v.Type)
	}

	// 解析数组元素类型
	objV, err := vo.ParseVariable(v.Schema)
	if err != nil {
		return nil, false, fmt.Errorf("node batch mode output schema should be variable, parse err: %w", err)
	}

	// 验证元素类型必须是对象
	if objV.Type != vo.VariableTypeObject {
		return nil, false, fmt.Errorf("node batch mode output element should be object, actual type: %s", objV.Type)
	}

	// 提取对象的字段定义作为内部输出
	objFieldStr, err := sonic.MarshalString(objV.Schema)
	if err != nil {
		return nil, false, err
	}

	err = sonic.UnmarshalString(objFieldStr, &innerOutput)
	if err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal obj schema into variable list: %w", err)
	}

	// 创建批量节点的输出参数（引用内部节点的输出）
	outerOutputP := &vo.Param{
		Name: v.Name,
		Input: &vo.BlockInput{
			Type:   vo.VariableTypeList,
			Schema: objV,
			Value: &vo.BlockInputValue{
				Type: vo.BlockInputValueTypeRef,
				Content: &vo.BlockInputReference{
					Source:  vo.RefSourceTypeBlockOutput,
					BlockID: vo.GenerateNodeIDForBatchMode(n.ID), // 生成批量模式专用ID
					Name:    "",                                  // 空名称表示映射所有输出
				},
			},
		},
	}
	outerOutput = append(outerOutput, outerOutputP)

	// 创建批量父节点
	parentN := &vo.Node{
		ID:   n.ID,
		Type: entity.NodeTypeBatch.IDStr(),
		Data: &vo.Data{
			Meta: &vo.NodeMetaFE{
				Title: n.Data.Meta.Title,
			},
			Inputs: &vo.Inputs{
				InputParameters: outerInput,
				Batch: &vo.Batch{
					BatchSize: &vo.BlockInput{ // 批处理大小
						Type: vo.VariableTypeInteger,
						Value: &vo.BlockInputValue{
							Type:    vo.BlockInputValueTypeLiteral,
							Content: strconv.FormatInt(batchInfo.BatchSize, 10),
						},
					},
					ConcurrentSize: &vo.BlockInput{ // 并发处理数量
						Type: vo.VariableTypeInteger,
						Value: &vo.BlockInputValue{
							Type:    vo.BlockInputValueTypeLiteral,
							Content: strconv.FormatInt(batchInfo.ConcurrentSize, 10),
						},
					},
				},
			},
			Outputs: slices.Transform(outerOutput, func(a *vo.Param) any {
				return a
			}),
		},
	}

	// 创建内部处理节点（实际执行业务逻辑）
	innerN := &vo.Node{
		ID:   n.ID + "_inner", // 内部节点ID
		Type: n.Type,          // 保持原始节点类型
		Data: &vo.Data{
			Meta: &vo.NodeMetaFE{
				Title: n.Data.Meta.Title + "_inner",
			},
			Inputs: &vo.Inputs{
				InputParameters: innerInput,
				LLMParam:        n.Data.Inputs.LLMParam,       // LLM相关配置
				LLM:             n.Data.Inputs.LLM,            // LLM配置
				SettingOnError:  n.Data.Inputs.SettingOnError, // 错误处理
				SubWorkflow:     n.Data.Inputs.SubWorkflow,    // 子工作流
				PluginAPIParam:  n.Data.Inputs.PluginAPIParam, // 插件API
			},
			Outputs: slices.Transform(innerOutput, func(a *vo.Variable) any {
				return a
			}),
		},
	}

	// 建立父子节点关系
	parentN.Blocks = []*vo.Node{innerN}

	// 创建内部连接关系
	parentN.Edges = []*vo.Edge{
		{ // 父节点到子节点的输出连接
			SourceNodeID: parentN.ID,
			TargetNodeID: innerN.ID,
			SourcePortID: "batch-function-inline-output",
		},
		{ // 子节点到父节点的输入连接
			SourceNodeID: innerN.ID,
			TargetNodeID: parentN.ID,
			TargetPortID: "batch-function-inline-input",
		},
	}

	// 设置父子关系
	innerN.SetParent(parentN)

	return parentN, true, nil
}

// RegisterAllNodeAdaptors 注册所有节点类型的适配器
//
// 这个函数在系统启动时被调用，注册所有支持的节点类型及其对应的适配器。
// 每个适配器负责将前端节点配置转换为执行引擎能够理解的schema格式。
//
// 注册机制：
// - 工厂函数模式：每个节点类型对应一个适配器工厂函数
// - 延迟初始化：每次需要时才创建新的配置实例
// - 类型安全：通过泛型确保适配器类型正确
//
// 支持的节点类型包括：
// - 基础节点：Entry, Exit, Variable操作
// - 逻辑节点：Selector, Loop, Batch, Break, Continue
// - 功能节点：LLM, HTTP, Database, Plugin
// - 复合节点：SubWorkflow, Knowledge, IntentDetector
//
// 注意：
//   - 这个函数应该在应用启动时调用一次
//   - 注册顺序不影响功能，但影响错误消息的显示顺序
func RegisterAllNodeAdaptors() {
	// register a generator function so that each time a NodeAdaptor is needed,
	// we can provide a brand new Config instance.
	nodes.RegisterNodeAdaptor(entity.NodeTypeEntry, func() nodes.NodeAdaptor {
		return &entry.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeSelector, func() nodes.NodeAdaptor {
		return &selector.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeBatch, func() nodes.NodeAdaptor {
		return &batch.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeBreak, func() nodes.NodeAdaptor {
		return &_break.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeContinue, func() nodes.NodeAdaptor {
		return &_continue.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeInputReceiver, func() nodes.NodeAdaptor {
		return &receiver.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeJsonSerialization, func() nodes.NodeAdaptor {
		return &json.SerializationConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeJsonDeserialization, func() nodes.NodeAdaptor {
		return &json.DeserializationConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeVariableAssigner, func() nodes.NodeAdaptor {
		return &variableassigner.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeVariableAssignerWithinLoop, func() nodes.NodeAdaptor {
		return &variableassigner.InLoopConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypePlugin, func() nodes.NodeAdaptor {
		return &plugin.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeCodeRunner, func() nodes.NodeAdaptor {
		return &code.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeOutputEmitter, func() nodes.NodeAdaptor {
		return &emitter.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeExit, func() nodes.NodeAdaptor {
		return &exit.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeVariableAggregator, func() nodes.NodeAdaptor {
		return &variableaggregator.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeTextProcessor, func() nodes.NodeAdaptor {
		return &textprocessor.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeIntentDetector, func() nodes.NodeAdaptor {
		return &intentdetector.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeQuestionAnswer, func() nodes.NodeAdaptor {
		return &qa.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeHTTPRequester, func() nodes.NodeAdaptor {
		return &httprequester.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeLoop, func() nodes.NodeAdaptor {
		return &loop.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeKnowledgeIndexer, func() nodes.NodeAdaptor {
		return &knowledge.IndexerConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeKnowledgeRetriever, func() nodes.NodeAdaptor {
		return &knowledge.RetrieveConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeKnowledgeDeleter, func() nodes.NodeAdaptor {
		return &knowledge.DeleterConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeDatabaseInsert, func() nodes.NodeAdaptor {
		return &database.InsertConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeDatabaseUpdate, func() nodes.NodeAdaptor {
		return &database.UpdateConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeDatabaseQuery, func() nodes.NodeAdaptor {
		return &database.QueryConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeDatabaseDelete, func() nodes.NodeAdaptor {
		return &database.DeleteConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeDatabaseCustomSQL, func() nodes.NodeAdaptor {
		return &database.CustomSQLConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeLLM, func() nodes.NodeAdaptor {
		return &llm.Config{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeCreateConversation, func() nodes.NodeAdaptor {
		return &conversation.CreateConversationConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeConversationUpdate, func() nodes.NodeAdaptor {
		return &conversation.UpdateConversationConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeConversationDelete, func() nodes.NodeAdaptor {
		return &conversation.DeleteConversationConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeConversationList, func() nodes.NodeAdaptor {
		return &conversation.ConversationListConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeConversationHistory, func() nodes.NodeAdaptor {
		return &conversation.ConversationHistoryConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeClearConversationHistory, func() nodes.NodeAdaptor {
		return &conversation.ClearConversationHistoryConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeMessageList, func() nodes.NodeAdaptor {
		return &conversation.MessageListConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeCreateMessage, func() nodes.NodeAdaptor {
		return &conversation.CreateMessageConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeEditMessage, func() nodes.NodeAdaptor {
		return &conversation.EditMessageConfig{}
	})
	nodes.RegisterNodeAdaptor(entity.NodeTypeDeleteMessage, func() nodes.NodeAdaptor {
		return &conversation.DeleteMessageConfig{}
	})

	// register branch adaptors
	nodes.RegisterBranchAdaptor(entity.NodeTypeSelector, func() nodes.BranchAdaptor {
		return &selector.Config{}
	})
	nodes.RegisterBranchAdaptor(entity.NodeTypeIntentDetector, func() nodes.BranchAdaptor {
		return &intentdetector.Config{}
	})
	nodes.RegisterBranchAdaptor(entity.NodeTypeQuestionAnswer, func() nodes.BranchAdaptor {
		return &qa.Config{}
	})
}
