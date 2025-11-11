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

// CompositeNode 复合节点结构定义
//
// 表示工作流中的复合节点（如循环、批量等），包含父节点和其所有子节点。
// 用于处理具有层级关系的复杂节点结构。
type CompositeNode struct {
	// Parent 父节点schema
	// 复合节点的主体部分，包含主要的配置和逻辑
	Parent *NodeSchema

	// Children 子节点列表
	// 复合节点内部的所有子节点，如循环体内的节点
	Children []*NodeSchema
}

// Init 初始化WorkflowSchema的内部状态和缓存
//
// 这个方法是WorkflowSchema的核心初始化函数，确保schema在执行前
// 具备所有必要的内部数据结构和计算结果。
//
// 初始化过程是线程安全的，只会执行一次（once模式）。
//
// 初始化内容：
// 1. 构建节点快速查找映射
// 2. 计算复合节点结构
// 3. 计算历史对话轮数
// 4. 确定检查点需求
// 5. 确定流式输出需求
//
// 为什么需要初始化：
// - 优化运行时性能：预先构建查找表
// - 缓存计算结果：避免重复计算
// - 确定执行特性：为执行引擎提供必要信息
//
// 注意：
//   - 该方法是幂等的，多次调用不会产生副作用
//   - 递归初始化子工作流
//   - 按需启用检查点和流式输出
func (w *WorkflowSchema) Init() {
	// 确保初始化只执行一次（线程安全）
	w.once.Do(func() {
		// 第一步：构建节点快速查找映射
		// 优化：将线性查找O(n)优化为哈希查找O(1)
		w.nodeMap = make(map[vo.NodeKey]*NodeSchema)
		for _, node := range w.Nodes {
			w.nodeMap[node.Key] = node
		}

		// 第二步：计算复合节点结构
		// 分析节点层级关系，构建父子节点映射
		w.doGetCompositeNodes()

		// 第三步：计算历史对话轮数和工作流特性
		// 遍历所有节点，聚合计算各项指标
		historyRounds := int64(0)
		for _, node := range w.Nodes {
			// 处理子工作流节点
			// 递归初始化子工作流，并继承其历史轮数和检查点需求
			if node.Type == entity.NodeTypeSubWorkflow {
				// 递归初始化子工作流schema
				node.SubWorkflowSchema.Init()

				// 取最大历史轮数（支持嵌套子工作流）
				historyRounds = max(historyRounds, node.SubWorkflowSchema.HistoryRounds())

				// 继承子工作流的检查点需求
				if node.SubWorkflowSchema.requireCheckPoint {
					w.requireCheckPoint = true
					break // 找到一个需要检查点的即可，无需继续检查
				}
			}

			// 处理支持聊天历史的节点
			// 检查节点配置是否实现了ChatHistoryAware接口
			chatHistoryAware, ok := node.Configs.(ChatHistoryAware)
			if ok && chatHistoryAware.ChatHistoryEnabled() {
				// 累加历史对话轮数
				historyRounds = max(historyRounds, chatHistoryAware.ChatHistoryRounds())
			}

			// 处理需要检查点的节点
			// 检查节点配置是否实现了RequireCheckpoint接口
			if rc, ok := node.Configs.(RequireCheckpoint); ok {
				if rc.RequireCheckpoint() {
					w.requireCheckPoint = true
					break // 找到一个需要检查点的即可，无需继续检查
				}
			}
		}

		// 设置计算结果
		w.historyRounds = historyRounds

		// 第四步：确定流式输出需求
		// 分析节点间的流式数据流，判断是否需要启用流式执行
		w.requireStreaming = w.doRequireStreaming()
	})
}

// GetNode 根据节点key获取节点schema
//
// 使用预构建的nodeMap进行O(1)查找，避免线性遍历。
// 如果节点不存在返回nil。
//
// 参数：
//   - key: 节点唯一标识
//
// 返回：
//   - *NodeSchema: 节点schema，如果不存在返回nil
func (w *WorkflowSchema) GetNode(key vo.NodeKey) *NodeSchema {
	return w.nodeMap[key]
}

// GetAllNodes 获取所有节点的映射
//
// 返回包含所有节点的map，key为节点key，value为节点schema。
// 注意：这个方法目前没有考虑批量模式下的生成节点计数。
//
// 返回：
//   - map[vo.NodeKey]*NodeSchema: 所有节点的映射
//
// TODO: 需要单独计算节点数量，考虑批量模式节点
func (w *WorkflowSchema) GetAllNodes() map[vo.NodeKey]*NodeSchema {
	return w.nodeMap // TODO: needs to calculate node count separately, considering batch mode nodes
}

// GetCompositeNodes 获取所有复合节点
//
// 返回工作流中的所有复合节点列表，包含父子关系。
// 如果compositeNodes还未初始化，会自动调用doGetCompositeNodes()进行计算。
//
// 返回：
//   - []*CompositeNode: 复合节点列表
func (w *WorkflowSchema) GetCompositeNodes() []*CompositeNode {
	if w.compositeNodes == nil {
		w.compositeNodes = w.doGetCompositeNodes()
	}

	return w.compositeNodes
}

// GetBranch 根据分支节点key获取分支schema
//
// 用于获取条件分支、选择器等节点的分支执行逻辑。
//
// 参数：
//   - key: 分支节点key
//
// 返回：
//   - *BranchSchema: 分支schema，如果不存在返回nil
func (w *WorkflowSchema) GetBranch(key vo.NodeKey) *BranchSchema {
	if w.Branches == nil {
		return nil
	}

	return w.Branches[key]
}

// RequireCheckpoint 检查是否需要检查点
//
// 返回工作流是否需要启用检查点机制。
// 检查点用于支持长时间运行任务的中断恢复。
//
// 返回：
//   - bool: true表示需要检查点
func (w *WorkflowSchema) RequireCheckpoint() bool {
	return w.requireCheckPoint
}

// RequireStreaming 检查是否需要流式输出
//
// 返回工作流是否需要启用流式执行模式。
// 流式输出用于支持实时响应和渐进式结果展示。
//
// 返回：
//   - bool: true表示需要流式输出
func (w *WorkflowSchema) RequireStreaming() bool {
	return w.requireStreaming
}

// HistoryRounds 获取历史对话轮数
//
// 返回聊天工作流需要保持的历史对话轮数。
// 用于LLM节点等需要上下文记忆的组件。
//
// 返回：
//   - int64: 历史对话轮数
func (w *WorkflowSchema) HistoryRounds() int64 { return w.historyRounds }

// SetHistoryRounds 设置历史对话轮数
//
// 用于动态调整工作流的历史对话轮数设置。
//
// 参数：
//   - historyRounds: 新的历史对话轮数
func (w *WorkflowSchema) SetHistoryRounds(historyRounds int64) {
	w.historyRounds = historyRounds
}

// doGetCompositeNodes 计算并返回所有复合节点
//
// 基于Hierarchy映射构建复合节点结构，将父子关系转换为CompositeNode对象。
// 这个方法是GetCompositeNodes()的内部实现，用于延迟初始化复合节点列表。
//
// 处理逻辑：
// 1. 检查是否存在层级关系，如果没有直接返回nil
// 2. 构建父节点到子节点列表的映射
// 3. 为每个有子节点的父节点创建CompositeNode对象
//
// 返回：
//   - []*CompositeNode: 复合节点列表，每个包含父节点和其所有子节点
//
// 注意：
//   - 只处理在nodeMap中存在的有效节点
//   - 返回的CompositeNode包含完整的父子关系信息
func (w *WorkflowSchema) doGetCompositeNodes() (cNodes []*CompositeNode) {
	if w.Hierarchy == nil {
		return nil
	}

	// 第一步：构建父节点到子节点列表的映射
	// 从 child->parent 的Hierarchy反向构建 parent->children 的关系
	parentToChildren := make(map[vo.NodeKey][]*NodeSchema)
	for childKey, parentKey := range w.Hierarchy {
		// 确保父节点和子节点都存在于nodeMap中
		if parentSchema := w.nodeMap[parentKey]; parentSchema != nil {
			if childSchema := w.nodeMap[childKey]; childSchema != nil {
				parentToChildren[parentKey] = append(parentToChildren[parentKey], childSchema)
			}
		}
	}

	// 第二步：为每个有子节点的父节点创建CompositeNode对象
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

// IsInSameWorkflow 检查两个节点是否在同一个工作流中
//
// 基于节点层级关系判断两个节点是否属于同一复合节点的作用域。
// 用于确定节点间的关系和执行边界。
//
// 参数：
//   - n: 节点层级映射 (child -> parent)
//   - nodeKey: 第一个节点key
//   - otherNodeKey: 第二个节点key
//
// 返回：
//   - bool: true表示两个节点在同一个工作流作用域内
//
// 判断逻辑：
//   - 如果没有层级关系（n==nil），所有节点都在同一作用域
//   - 如果两个节点都没有父节点，在同一作用域
//   - 如果一个有父节点一个没有，在不同作用域
//   - 如果都有父节点，比较父节点是否相同
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

// IsBelowOneLevel 检查节点是否比另一个节点低一级
//
// 判断第一个节点是否直接位于第二个节点的下一层级。
// 用于确定节点间的直接父子关系。
//
// 参数：
//   - n: 节点层级映射 (child -> parent)
//   - nodeKey: 待检查的节点key
//   - otherNodeKey: 参考节点key
//
// 返回：
//   - bool: true表示nodeKey直接位于otherNodeKey的下一级
//
// 判断逻辑：
//   - 如果没有层级关系，返回false
//   - nodeKey有父节点且otherNodeKey没有父节点时，返回true
func IsBelowOneLevel(n map[vo.NodeKey]vo.NodeKey, nodeKey, otherNodeKey vo.NodeKey) bool {
	if n == nil {
		return false
	}
	_, myParentExists := n[nodeKey]
	_, theirParentExists := n[otherNodeKey]

	return myParentExists && !theirParentExists
}

// IsParentOf 检查节点是否是另一个节点的父节点
//
// 直接检查节点间的父子关系，确定第一个节点是否包含第二个节点。
//
// 参数：
//   - n: 节点层级映射 (child -> parent)
//   - nodeKey: 可能的父节点key
//   - otherNodeKey: 可能的子节点key
//
// 返回：
//   - bool: true表示nodeKey是otherNodeKey的直接父节点
//
// 判断逻辑：
//   - 如果没有层级关系，返回false
//   - 检查otherNodeKey的父节点是否等于nodeKey
func IsParentOf(n map[vo.NodeKey]vo.NodeKey, nodeKey, otherNodeKey vo.NodeKey) bool {
	if n == nil {
		return false
	}
	theirParent, theirParentExists := n[otherNodeKey]

	return theirParentExists && theirParent == nodeKey
}

// IsEqual 比较两个WorkflowSchema是否相等
//
// 深度比较两个工作流schema的执行逻辑是否完全相同。
// 用于工作流版本控制和测试运行状态继承判断。
//
// 比较内容：
// 1. 连接关系：比较所有节点间的连接是否相同
// 2. 节点配置：比较每个节点的完整配置和类型信息
//
// 参数：
//   - other: 要比较的另一个WorkflowSchema
//
// 返回：
//   - bool: true表示两个schema执行逻辑完全相同
//
// 比较策略：
//   - 连接比较：使用Connection.ID()进行标准化比较
//   - 节点比较：深度比较所有关键字段（名称、配置、类型等）
//   - 子工作流：比较SubWorkflowBasic信息
//
// 注意：
//   - 只比较执行逻辑相关的字段，不比较运行时状态
//   - 忽略可视化相关的元数据（如位置信息）
//   - 用于优化测试运行，避免重复执行相同的逻辑
func (w *WorkflowSchema) IsEqual(other *WorkflowSchema) bool {
	// 第一步：比较连接关系
	// 使用Connection.ID()标准化连接标识，避免顺序依赖
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

	// 第二步：比较节点配置
	// 构建节点映射，便于按key比较
	otherNodeMap := make(map[vo.NodeKey]*NodeSchema, len(other.Nodes))
	for _, node := range other.Nodes {
		otherNodeMap[node.Key] = node
	}
	nodeMap := make(map[vo.NodeKey]*NodeSchema, len(w.Nodes))
	for _, node := range w.Nodes {
		nodeMap[node.Key] = node
	}

	// 使用自定义比较函数深度比较节点
	if !maps.EqualFunc(otherNodeMap, nodeMap, func(node *NodeSchema, other *NodeSchema) bool {
		// 比较基本信息
		if node.Name != other.Name {
			return false
		}

		// 比较配置（使用深度比较）
		if !reflect.DeepEqual(node.Configs, other.Configs) {
			return false
		}

		// 比较输入输出类型定义
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

		// 比较异常处理配置
		if !reflect.DeepEqual(node.ExceptionConfigs, other.ExceptionConfigs) {
			return false
		}

		// 比较子工作流基本信息
		if !reflect.DeepEqual(node.SubWorkflowBasic, other.SubWorkflowBasic) {
			return false
		}

		return true
	}) {
		return false
	}

	return true
}

// NodeCount 计算工作流的实际节点数量
//
// 返回工作流中用户定义的节点数量，不包括批量模式下动态生成的节点。
// 用于统计和展示工作流的复杂度指标。
//
// 计算公式：
// 总节点数 - 生成节点数 = 用户定义节点数
//
// 返回：
//   - int32: 实际节点数量（不包括生成的节点）
//
// 注意：
//   - GeneratedNodes记录了批量模式下自动创建的内部节点
//   - 这个计数用于前端展示和统计分析
func (w *WorkflowSchema) NodeCount() int32 {
	return int32(len(w.Nodes) - len(w.GeneratedNodes))
}

// doRequireStreaming 分析工作流是否需要流式执行
//
// 通过图论算法分析节点间的流式数据流，判断工作流是否需要启用流式模式。
// 流式执行允许渐进式输出，提升用户体验。
//
// 分析逻辑：
// 1. 识别流式生产者和消费者节点
// 2. 构建数据流图（邻接表）
// 3. 对每个生产者执行可达性分析，检查是否能到达消费者
//
// 参数：无（基于当前schema的节点配置）
//
// 返回：
//   - bool: true表示工作流需要流式执行
//
// 关键概念：
//   - 生产者：CanGeneratesStream=true的节点
//   - 消费者：RequireStreamingInput=true的节点
//   - 可达性：通过BFS算法检查生产者到消费者的路径
//
// 优化策略：
//   - 只有当同时存在生产者和消费者时才需要分析
//   - 使用BFS进行图遍历，检测流式数据链路
//   - 一旦找到一条路径即可确认需要流式执行
func (w *WorkflowSchema) doRequireStreaming() bool {
	// 第一步：识别流式生产者和消费者节点
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

	// 如果没有生产者或消费者，不需要流式执行
	if len(producers) == 0 || len(consumers) == 0 {
		return false
	}

	// 第二步：构建数据流图（邻接表）
	// 基于InputSources建立节点间的依赖关系
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

	// 第三步：对每个生产者执行可达性分析
	// 使用BFS检查是否有路径到达消费者节点
	for p := range producers {
		q := []vo.NodeKey{p}
		visited := make(map[vo.NodeKey]bool)
		visited[p] = true

		for len(q) > 0 {
			curr := q[0]
			q = q[1:]

			// 如果当前节点是消费者，说明找到了流式路径
			if consumers[curr] {
				return true
			}

			// 继续遍历邻居节点
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

// GetAllNodesInputFileFields 收集工作流中所有节点的输入文件信息
//
// 遍历所有节点，提取输入参数中的文件信息，用于文件管理和URL适配。
// 支持递归处理子工作流中的文件信息。
//
// 处理流程：
// 1. 遍历所有节点的输入源
// 2. 提取文件相关的输入参数
// 3. 适配文件URL（移除查询参数）
// 4. 递归处理子工作流
//
// 参数：
//   - ctx: 上下文，用于日志记录
//
// 返回：
//   - []*workflowModel.FileInfo: 所有输入文件的详细信息列表
//
// 文件信息包含：
//   - 文件名
//   - 文件URL（已适配）
//   - 文件扩展名
//
// 注意：
//   - 会修改原始的source.Source.Val，将URL适配为标准格式
//   - 支持单个文件和批量文件数组
//   - 递归处理嵌套的子工作流
func (w *WorkflowSchema) GetAllNodesInputFileFields(ctx context.Context) []*workflowModel.FileInfo {

	// adaptorURL URL适配函数
	// 移除工作流特定的查询参数，转换为标准文件URL格式
	// 主要清理x-wf-file_name参数，该参数用于工作流内部标识
	adaptorURL := func(s string) (string, error) {
		u, err := url.Parse(s)
		if err != nil {
			return "", err
		}
		query := u.Query()
		query.Del("x-wf-file_name") // 移除工作流文件标识参数
		u.RawQuery = query.Encode()
		return u.String(), nil
	}

	// 初始化结果列表
	result := make([]*workflowModel.FileInfo, 0)

	// 遍历所有节点，处理输入源中的文件信息
	for _, node := range w.Nodes {
		for _, source := range node.InputSources {
			// 检查是否包含文件信息
			if source.Source.Val != nil && source.Source.FileExtra != nil {
				fileExtra := source.Source.FileExtra

				// 处理单个文件
				if fileExtra.FileName != nil {
					fileURL, err := adaptorURL(source.Source.Val.(string))
					if err != nil {
						logs.CtxWarnf(ctx, "failed to parse adaptorURL for node %v: %v", node.Key, err)
						continue
					}

					// 收集文件信息
					result = append(result, &workflowModel.FileInfo{
						FileName:      *fileExtra.FileName,
						FileURL:       fileURL,
						FileExtension: filepath.Ext(strings.TrimSpace(*fileExtra.FileName)),
					})

					// 更新源值为适配后的URL
					source.Source.Val = fileURL
				}

				// 处理批量文件数组
				if fileExtra.FileNames != nil {
					vals := source.Source.Val.([]any)

					// 逐个处理批量文件
					for idx, fileName := range fileExtra.FileNames {
						fileURL := vals[idx].(string)
						fileURL, err := adaptorURL(fileURL)
						if err != nil {
							logs.CtxWarnf(ctx, "failed to parse adaptorURL for node %v: %v", node.Key, err)
							continue
						}

						// 收集文件信息
						result = append(result, &workflowModel.FileInfo{
							FileName:      fileName,
							FileURL:       fileURL,
							FileExtension: filepath.Ext(strings.TrimSpace(fileName)),
						})

						// 更新数组中的URL
						vals[idx] = fileURL
					}

					// 更新源值为更新后的数组
					source.Source.Val = vals
				}
			}
		}

		// 递归处理子工作流中的文件信息
		if node.SubWorkflowSchema != nil {
			result = append(result, node.SubWorkflowSchema.GetAllNodesInputFileFields(ctx)...)
		}
	}

	return result
}
