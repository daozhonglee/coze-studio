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

package schema

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	workflowModel "github.com/coze-dev/coze-studio/backend/crossdomain/workflow/model"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
)

// WorkflowSchema 工作流执行模式的数据结构定义
//
// 这是一个核心的数据结构，将前端的可视化画布转换为可执行的工作流模式。
// 它是工作流引擎执行的基础，定义了节点、连接关系、分支逻辑等执行要素。
//
// 主要用途：
// 1. 工作流执行引擎的输入数据结构
// 2. 工作流合法性验证
// 3. 工作流状态比较和版本控制
// 4. 工作流执行计划的生成
//
// 数据流向：
// 前端Canvas → CanvasToWorkflowSchema() → WorkflowSchema → 执行引擎
type WorkflowSchema struct {
	// 可序列化的字段（JSON）

	// Nodes 工作流中的所有节点schema定义
	// 包含节点的配置、输入输出类型、执行逻辑等
	Nodes []*NodeSchema `json:"nodes"`

	// Connections 节点之间的连接关系
	// 定义数据流的方向和端口连接
	Connections []*Connection `json:"connections"`

	// Hierarchy 节点层级关系映射
	// key: 子节点key, value: 父节点key
	// 用于处理复合节点（如循环、批量等）中的嵌套关系
	Hierarchy map[vo.NodeKey]vo.NodeKey `json:"hierarchy,omitempty"`

	// Branches 分支逻辑定义
	// key: 分支节点key, value: 分支schema
	// 用于条件分支、选择器等节点的执行逻辑
	Branches map[vo.NodeKey]*BranchSchema `json:"branches,omitempty"`

	// GeneratedNodes 批量模式下生成的节点列表
	// 记录因批量处理而动态生成的节点
	GeneratedNodes []vo.NodeKey `json:"generated_nodes,omitempty"`

	// 不可序列化的字段（运行时使用）

	// nodeMap 节点快速查找映射（运行时构建）
	// 用于根据节点key快速定位节点schema
	nodeMap map[vo.NodeKey]*NodeSchema

	// compositeNodes 复合节点列表
	// 包含所有复合节点及其子节点信息
	compositeNodes []*CompositeNode

	// requireCheckPoint 是否需要检查点
	// 某些节点（如长时间运行的任务）需要检查点来支持恢复
	requireCheckPoint bool

	// requireStreaming 是否需要流式输出
	// 用于支持实时流式响应的工作流
	requireStreaming bool

	// historyRounds 历史对话轮数
	// 聊天工作流中需要保持的历史对话轮数
	historyRounds int64

	// once 确保初始化只执行一次
	once sync.Once
}

// Connection 定义工作流节点之间的连接关系
// 表示数据从一个节点流向另一个节点的路径
type Connection struct {
	// FromNode 数据来源节点的key
	FromNode vo.NodeKey `json:"from_node"`

	// ToNode 数据目标节点的key
	ToNode vo.NodeKey `json:"to_node"`

	// FromPort 数据来源的端口标识（可选）
	// 用于区分同一个节点的不同输出端口
	FromPort *string `json:"from_port,omitempty"`
}

func (c *Connection) ID() string {
	if c.FromPort != nil {
		return fmt.Sprintf("%s:%s:%v", c.FromNode, c.ToNode, *c.FromPort)
	}
	return fmt.Sprintf("%v:%v", c.FromNode, c.ToNode)
}

type CompositeNode struct {
	Parent   *NodeSchema
	Children []*NodeSchema
}

func (w *WorkflowSchema) Init() {
	w.once.Do(func() {
		w.nodeMap = make(map[vo.NodeKey]*NodeSchema)
		for _, node := range w.Nodes {
			w.nodeMap[node.Key] = node
		}

		w.doGetCompositeNodes()

		historyRounds := int64(0)
		for _, node := range w.Nodes {
			if node.Type == entity.NodeTypeSubWorkflow {
				node.SubWorkflowSchema.Init()
				historyRounds = max(historyRounds, node.SubWorkflowSchema.HistoryRounds())
				if node.SubWorkflowSchema.requireCheckPoint {
					w.requireCheckPoint = true
					break
				}
			}

			chatHistoryAware, ok := node.Configs.(ChatHistoryAware)
			if ok && chatHistoryAware.ChatHistoryEnabled() {
				historyRounds = max(historyRounds, chatHistoryAware.ChatHistoryRounds())
			}

			if rc, ok := node.Configs.(RequireCheckpoint); ok {
				if rc.RequireCheckpoint() {
					w.requireCheckPoint = true
					break
				}
			}
		}

		w.historyRounds = historyRounds
		w.requireStreaming = w.doRequireStreaming()
	})
}

func (w *WorkflowSchema) GetNode(key vo.NodeKey) *NodeSchema {
	return w.nodeMap[key]
}

func (w *WorkflowSchema) GetAllNodes() map[vo.NodeKey]*NodeSchema {
	return w.nodeMap // TODO: needs to calculate node count separately, considering batch mode nodes
}

func (w *WorkflowSchema) GetCompositeNodes() []*CompositeNode {
	if w.compositeNodes == nil {
		w.compositeNodes = w.doGetCompositeNodes()
	}

	return w.compositeNodes
}

func (w *WorkflowSchema) GetBranch(key vo.NodeKey) *BranchSchema {
	if w.Branches == nil {
		return nil
	}

	return w.Branches[key]
}

func (w *WorkflowSchema) RequireCheckpoint() bool {
	return w.requireCheckPoint
}

func (w *WorkflowSchema) RequireStreaming() bool {
	return w.requireStreaming
}

func (w *WorkflowSchema) HistoryRounds() int64 { return w.historyRounds }

func (w *WorkflowSchema) SetHistoryRounds(historyRounds int64) {
	w.historyRounds = historyRounds
}

func (w *WorkflowSchema) doGetCompositeNodes() (cNodes []*CompositeNode) {
	if w.Hierarchy == nil {
		return nil
	}

	// Build parent to children mapping
	parentToChildren := make(map[vo.NodeKey][]*NodeSchema)
	for childKey, parentKey := range w.Hierarchy {
		if parentSchema := w.nodeMap[parentKey]; parentSchema != nil {
			if childSchema := w.nodeMap[childKey]; childSchema != nil {
				parentToChildren[parentKey] = append(parentToChildren[parentKey], childSchema)
			}
		}
	}

	// Create composite nodes
	for parentKey, children := range parentToChildren {
		if parentSchema := w.nodeMap[parentKey]; parentSchema != nil {
			cNodes = append(cNodes, &CompositeNode{
				Parent:   parentSchema,
				Children: children,
			})
		}
	}

	return cNodes
}

func IsInSameWorkflow(n map[vo.NodeKey]vo.NodeKey, nodeKey, otherNodeKey vo.NodeKey) bool {
	if n == nil {
		return true
	}

	myParents, myParentExists := n[nodeKey]
	theirParents, theirParentExists := n[otherNodeKey]

	if !myParentExists && !theirParentExists {
		return true
	}

	if !myParentExists || !theirParentExists {
		return false
	}

	return myParents == theirParents
}

func IsBelowOneLevel(n map[vo.NodeKey]vo.NodeKey, nodeKey, otherNodeKey vo.NodeKey) bool {
	if n == nil {
		return false
	}
	_, myParentExists := n[nodeKey]
	_, theirParentExists := n[otherNodeKey]

	return myParentExists && !theirParentExists
}

func IsParentOf(n map[vo.NodeKey]vo.NodeKey, nodeKey, otherNodeKey vo.NodeKey) bool {
	if n == nil {
		return false
	}
	theirParent, theirParentExists := n[otherNodeKey]

	return theirParentExists && theirParent == nodeKey
}

func (w *WorkflowSchema) IsEqual(other *WorkflowSchema) bool {
	otherConnectionsMap := make(map[string]bool, len(other.Connections))
	for _, connection := range other.Connections {
		otherConnectionsMap[connection.ID()] = true
	}
	connectionsMap := make(map[string]bool, len(other.Connections))
	for _, connection := range w.Connections {
		connectionsMap[connection.ID()] = true
	}
	if !maps.Equal(otherConnectionsMap, connectionsMap) {
		return false
	}
	otherNodeMap := make(map[vo.NodeKey]*NodeSchema, len(other.Nodes))
	for _, node := range other.Nodes {
		otherNodeMap[node.Key] = node
	}
	nodeMap := make(map[vo.NodeKey]*NodeSchema, len(w.Nodes))

	for _, node := range w.Nodes {
		nodeMap[node.Key] = node
	}

	if !maps.EqualFunc(otherNodeMap, nodeMap, func(node *NodeSchema, other *NodeSchema) bool {
		if node.Name != other.Name {
			return false
		}
		if !reflect.DeepEqual(node.Configs, other.Configs) {
			return false
		}
		if !reflect.DeepEqual(node.InputTypes, other.InputTypes) {
			return false
		}
		if !reflect.DeepEqual(node.InputSources, other.InputSources) {
			return false
		}

		if !reflect.DeepEqual(node.OutputTypes, other.OutputTypes) {
			return false
		}
		if !reflect.DeepEqual(node.OutputSources, other.OutputSources) {
			return false
		}
		if !reflect.DeepEqual(node.ExceptionConfigs, other.ExceptionConfigs) {
			return false
		}
		if !reflect.DeepEqual(node.SubWorkflowBasic, other.SubWorkflowBasic) {
			return false
		}
		return true

	}) {
		return false
	}

	return true

}

func (w *WorkflowSchema) NodeCount() int32 {
	return int32(len(w.Nodes) - len(w.GeneratedNodes))
}

func (w *WorkflowSchema) doRequireStreaming() bool {
	producers := make(map[vo.NodeKey]bool)
	consumers := make(map[vo.NodeKey]bool)

	for _, node := range w.Nodes {
		if node.StreamConfigs != nil && node.StreamConfigs.CanGeneratesStream {
			producers[node.Key] = true
		}

		if node.StreamConfigs != nil && node.StreamConfigs.RequireStreamingInput {
			consumers[node.Key] = true
		}

	}

	if len(producers) == 0 || len(consumers) == 0 {
		return false
	}

	// Build data-flow graph from InputSources
	adj := make(map[vo.NodeKey]map[vo.NodeKey]struct{})
	for _, node := range w.Nodes {
		for _, source := range node.InputSources {
			if source.Source.Ref != nil && len(source.Source.Ref.FromNodeKey) > 0 {
				if _, ok := adj[source.Source.Ref.FromNodeKey]; !ok {
					adj[source.Source.Ref.FromNodeKey] = make(map[vo.NodeKey]struct{})
				}
				adj[source.Source.Ref.FromNodeKey][node.Key] = struct{}{}
			}
		}
	}

	// For each producer, traverse the graph to see if it can reach a consumer
	for p := range producers {
		q := []vo.NodeKey{p}
		visited := make(map[vo.NodeKey]bool)
		visited[p] = true

		for len(q) > 0 {
			curr := q[0]
			q = q[1:]

			if consumers[curr] {
				return true
			}

			for neighbor := range adj[curr] {
				if !visited[neighbor] {
					visited[neighbor] = true
					q = append(q, neighbor)
				}
			}
		}
	}

	return false
}

func (w *WorkflowSchema) GetAllNodesInputFileFields(ctx context.Context) []*workflowModel.FileInfo {

	adaptorURL := func(s string) (string, error) {
		u, err := url.Parse(s)
		if err != nil {
			return "", err
		}
		query := u.Query()
		query.Del("x-wf-file_name")
		u.RawQuery = query.Encode()
		return u.String(), nil
	}

	result := make([]*workflowModel.FileInfo, 0)
	for _, node := range w.Nodes {
		for _, source := range node.InputSources {
			if source.Source.Val != nil && source.Source.FileExtra != nil {
				fileExtra := source.Source.FileExtra
				if fileExtra.FileName != nil {
					fileURL, err := adaptorURL(source.Source.Val.(string))
					if err != nil {
						logs.CtxWarnf(ctx, "failed to parse adaptorURL for node %v: %v", node.Key, err)
						continue
					}
					result = append(result, &workflowModel.FileInfo{
						FileName:      *fileExtra.FileName,
						FileURL:       fileURL,
						FileExtension: filepath.Ext(strings.TrimSpace(*fileExtra.FileName)),
					})
					source.Source.Val = fileURL

				}
				if fileExtra.FileNames != nil {
					vals := source.Source.Val.([]any)
					for idx, fileName := range fileExtra.FileNames {
						fileURL := vals[idx].(string)
						fileURL, err := adaptorURL(fileURL)
						if err != nil {
							logs.CtxWarnf(ctx, "failed to parse adaptorURL for node %v: %v", node.Key, err)
							continue
						}
						result = append(result, &workflowModel.FileInfo{
							FileName:      fileName,
							FileURL:       fileURL,
							FileExtension: filepath.Ext(strings.TrimSpace(fileName)),
						})
						vals[idx] = fileURL
					}
					source.Source.Val = vals
				}

			}
		}
		if node.SubWorkflowSchema != nil {
			result = append(result, node.SubWorkflowSchema.GetAllNodesInputFileFields(ctx)...)
		}

	}

	return result
}
