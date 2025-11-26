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

// branch_schema.go 分支 Schema 定义
//
// 本文件定义了工作流分支逻辑的数据结构和构建方法。
// 分支用于实现条件分支、异常处理分支等执行路径的选择。
//
// 主要功能：
//   - 定义分支 Schema 数据结构
//   - 从连接关系构建分支映射
//   - 提供异常分支和条件分支的执行逻辑

package schema

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/compose"

	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
)

// 端口类型常量
const (
	PortDefault      = "default"      // 默认端口
	PortBranchError  = "branch_error" // 异常分支端口
	PortBranchFormat = "branch_%d"    // 条件分支端口格式
)

// BranchSchema 分支 Schema 定义
//
// 描述节点的分支逻辑，包括：
//   - 默认分支映射
//   - 异常分支映射
//   - 条件分支映射（按索引）
type BranchSchema struct {
	From             vo.NodeKey                `json:"from_node"`
	DefaultMapping   map[string]bool           `json:"default_mapping,omitempty"`
	ExceptionMapping map[string]bool           `json:"exception_mapping,omitempty"`
	Mappings         map[int64]map[string]bool `json:"mappings,omitempty"`
}

// BuildBranches 从连接关系构建分支 Schema
//
// 解析连接中的端口信息，将其分类为：
//   - 默认分支（PortDefault）
//   - 异常分支（PortBranchError）
//   - 条件分支（branch_N 格式）
//
// 返回以源节点为 key 的分支映射
func BuildBranches(connections []*Connection) (map[vo.NodeKey]*BranchSchema, error) {
	var branchMap map[vo.NodeKey]*BranchSchema

	for _, conn := range connections {
		if conn.FromPort == nil || len(*conn.FromPort) == 0 {
			continue
		}

		port := *conn.FromPort
		sourceNodeKey := conn.FromNode

		if branchMap == nil {
			branchMap = map[vo.NodeKey]*BranchSchema{}
		}

		// Get or create branch schema for source node
		branch, exists := branchMap[sourceNodeKey]
		if !exists {
			branch = &BranchSchema{
				From: sourceNodeKey,
			}
			branchMap[sourceNodeKey] = branch
		}

		// Classify port type and add to appropriate mapping
		switch {
		case port == PortDefault:
			if branch.DefaultMapping == nil {
				branch.DefaultMapping = map[string]bool{}
			}
			branch.DefaultMapping[string(conn.ToNode)] = true
		case port == PortBranchError:
			if branch.ExceptionMapping == nil {
				branch.ExceptionMapping = map[string]bool{}
			}
			branch.ExceptionMapping[string(conn.ToNode)] = true
		default:
			var branchNum int64
			_, err := fmt.Sscanf(port, PortBranchFormat, &branchNum)
			if err != nil || branchNum < 0 {
				return nil, fmt.Errorf("invalid port format '%s' for connection %+v", port, conn)
			}
			if branch.Mappings == nil {
				branch.Mappings = map[int64]map[string]bool{}
			}
			if _, exists := branch.Mappings[branchNum]; !exists {
				branch.Mappings[branchNum] = make(map[string]bool)
			}
			branch.Mappings[branchNum][string(conn.ToNode)] = true
		}
	}

	return branchMap, nil
}

// OnlyException 判断是否仅包含异常分支
// 当没有条件分支，但有异常和默认分支时返回 true
func (bs *BranchSchema) OnlyException() bool {
	return len(bs.Mappings) == 0 && len(bs.ExceptionMapping) > 0 && len(bs.DefaultMapping) > 0
}

// GetExceptionBranch 获取仅包含异常处理的分支
//
// 根据输入中的 isSuccess 字段决定执行路径：
//   - isSuccess=false: 执行异常分支
//   - 否则: 执行默认分支
func (bs *BranchSchema) GetExceptionBranch() *compose.GraphBranch {
	condition := func(ctx context.Context, in map[string]any) (map[string]bool, error) {
		isSuccess, ok := in["isSuccess"]
		if ok && isSuccess != nil && !isSuccess.(bool) {
			return bs.ExceptionMapping, nil
		}

		return bs.DefaultMapping, nil
	}

	// Combine ExceptionMapping and DefaultMapping into a new map
	endNodes := make(map[string]bool)
	for node := range bs.ExceptionMapping {
		endNodes[node] = true
	}
	for node := range bs.DefaultMapping {
		endNodes[node] = true
	}

	return compose.NewGraphMultiBranch(condition, endNodes)
}

// GetFullBranch 获取完整分支逻辑（包含条件分支和异常分支）
//
// 使用 BranchBuilder 提取分支条件，结合异常处理逻辑，
// 构建完整的分支选择器。
//
// 执行优先级：
//  1. 检查异常（isSuccess=false）
//  2. 执行条件提取器获取分支索引
//  3. 根据索引或默认选择目标节点
func (bs *BranchSchema) GetFullBranch(ctx context.Context, bb BranchBuilder) (*compose.GraphBranch, error) {
	extractor, hasBranch := bb.BuildBranch(ctx)
	if !hasBranch {
		return nil, fmt.Errorf("branch expected but BranchBuilder thinks not. BranchSchema: %v", bs)
	}

	if len(bs.ExceptionMapping) == 0 { // no exception, it's a normal branch
		condition := func(ctx context.Context, in map[string]any) (map[string]bool, error) {
			index, isDefault, err := extractor(ctx, in)
			if err != nil {
				return nil, err
			}

			if isDefault {
				return bs.DefaultMapping, nil
			}

			if _, ok := bs.Mappings[index]; !ok {
				return nil, fmt.Errorf("chosen index= %d, out of range", index)
			}

			return bs.Mappings[index], nil
		}

		// Combine DefaultMapping and normal mappings into a new map
		endNodes := make(map[string]bool)
		for node := range bs.DefaultMapping {
			endNodes[node] = true
		}
		for _, ms := range bs.Mappings {
			for node := range ms {
				endNodes[node] = true
			}
		}

		return compose.NewGraphMultiBranch(condition, endNodes), nil
	}

	condition := func(ctx context.Context, in map[string]any) (map[string]bool, error) {
		isSuccess, ok := in["isSuccess"]
		if ok && isSuccess != nil && !isSuccess.(bool) {
			return bs.ExceptionMapping, nil
		}

		index, isDefault, err := extractor(ctx, in)
		if err != nil {
			return nil, err
		}

		if isDefault {
			return bs.DefaultMapping, nil
		}

		return bs.Mappings[index], nil
	}

	// Combine ALL mappings into a new map
	endNodes := make(map[string]bool)
	for node := range bs.ExceptionMapping {
		endNodes[node] = true
	}
	for node := range bs.DefaultMapping {
		endNodes[node] = true
	}
	for _, ms := range bs.Mappings {
		for node := range ms {
			endNodes[node] = true
		}
	}

	return compose.NewGraphMultiBranch(condition, endNodes), nil
}
