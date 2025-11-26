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

// Package execute 提供工作流执行过程中的上下文管理功能
//
// 本包负责管理工作流执行的完整上下文信息，包括：
//
// 1. 执行上下文层级
//   - RootCtx: 根工作流执行上下文
//   - SubWorkflowCtx: 子工作流执行上下文
//   - NodeCtx: 节点执行上下文
//   - BatchInfo: 批量/循环执行上下文
//
// 2. 上下文传递机制
//   - 通过 context.Value 传递执行上下文
//   - 支持上下文的嵌套和继承
//   - 提供检查点 ID 用于中断恢复
//
// 3. Token 收集
//   - 追踪 LLM 调用的 Token 使用量
//   - 支持层级聚合（节点 → 子工作流 → 根工作流）
//
// 4. 应用级变量存储
//   - 在整个工作流执行过程中共享的变量
//   - 线程安全的读写操作
package execute

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/eino/compose"

	workflowModel "github.com/coze-dev/coze-studio/backend/crossdomain/workflow/model"
	"github.com/coze-dev/coze-studio/backend/domain/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
)

// Context 工作流执行上下文
// 包含执行过程中需要的所有上下文信息，支持嵌套执行和中断恢复
type Context struct {
	// RootCtx 根工作流上下文（嵌入）
	RootCtx
	// SubWorkflowCtx 子工作流上下文（可选）
	*SubWorkflowCtx
	// NodeCtx 当前执行节点上下文（可选）
	*NodeCtx
	// BatchInfo 批量/循环执行信息（可选）
	*BatchInfo
	// TokenCollector Token 使用量收集器
	TokenCollector *TokenCollector
	// StartTime 执行开始时间（UnixMilli 时间戳）
	StartTime int64
	// CheckPointID 检查点 ID，用于中断恢复
	CheckPointID string
	// AppVarStore 应用级变量存储，在整个执行过程中共享
	AppVarStore *AppVariables
	// executed 已执行节点计数（原子操作）
	executed *atomic.Int64
}

// RootCtx 根工作流执行上下文
// 保存最顶层工作流的执行信息，在整个执行树中共享
type RootCtx struct {
	// RootWorkflowBasic 根工作流基本信息
	RootWorkflowBasic *entity.WorkflowBasic
	// RootExecuteID 根执行 ID，唯一标识此次执行
	RootExecuteID int64
	// ResumeEvent 恢复事件，用于中断后恢复执行
	ResumeEvent *entity.InterruptEvent
	// ExeCfg 执行配置，包含运行模式、操作者等信息
	ExeCfg workflowModel.ExecuteConfig
}

// SubWorkflowCtx 子工作流执行上下文
// 当执行进入子工作流节点时创建
type SubWorkflowCtx struct {
	// SubWorkflowBasic 子工作流基本信息
	SubWorkflowBasic *entity.WorkflowBasic
	// SubExecuteID 子工作流执行 ID
	SubExecuteID int64
}

// NodeCtx 节点执行上下文
// 每个节点执行时都会创建独立的节点上下文
type NodeCtx struct {
	// NodeKey 节点唯一标识
	NodeKey vo.NodeKey
	// NodeExecuteID 节点执行 ID
	NodeExecuteID int64
	// NodeName 节点名称（用于显示）
	NodeName string
	// NodeType 节点类型
	NodeType entity.NodeType
	// NodePath 节点路径，记录从根到当前节点的完整路径
	NodePath []string
	// TerminatePlan 终止计划，用于出口节点
	TerminatePlan *vo.TerminatePlan
	// ResumingEvent 正在恢复的中断事件
	ResumingEvent *entity.InterruptEvent
	// SubWorkflowExeID 如果此节点是子工作流节点，记录子工作流的执行 ID
	SubWorkflowExeID int64
	// CurrentRetryCount 当前重试次数
	CurrentRetryCount int
}

// BatchInfo 批量/循环执行信息
// 在 Batch 或 Loop 节点执行时创建，记录当前迭代的信息
type BatchInfo struct {
	// Index 当前迭代索引
	Index int
	// Items 当前迭代的数据项
	Items map[string]any
	// CompositeNodeKey 复合节点（Batch/Loop）的 Key
	CompositeNodeKey vo.NodeKey
}

type contextKey struct{}

func restoreWorkflowCtx(ctx context.Context, h *WorkflowHandler) (context.Context, error) {
	var storedCtx *Context
	err := compose.ProcessState[ExeContextStore](ctx, func(ctx context.Context, state ExeContextStore) error {
		if state == nil {
			return errors.New("state is nil")
		}

		var e error
		storedCtx, _, e = state.GetWorkflowCtx()
		if e != nil {
			return e
		}

		return nil
	})

	if err != nil {
		return ctx, err
	}

	if storedCtx == nil {
		return ctx, errors.New("stored workflow context is nil")
	}

	storedCtx.ResumeEvent = h.resumeEvent
	currentC := GetExeCtx(ctx)
	if currentC != nil {
		// restore the parent-child relationship between token collectors
		if storedCtx.TokenCollector != nil && storedCtx.TokenCollector.Parent != nil {
			currentTokenCollector := currentC.TokenCollector
			storedCtx.TokenCollector.Parent = currentTokenCollector
		}

		storedCtx.AppVarStore = currentC.AppVarStore
		storedCtx.executed = currentC.executed
	}

	return context.WithValue(ctx, contextKey{}, storedCtx), nil
}

func restoreNodeCtx(ctx context.Context, nodeKey vo.NodeKey, resumeEvent *entity.InterruptEvent,
	exactlyResuming bool) (context.Context, error) {
	var storedCtx *Context
	err := compose.ProcessState[ExeContextStore](ctx, func(ctx context.Context, state ExeContextStore) error {
		if state == nil {
			return errors.New("state is nil")
		}
		var e error
		storedCtx, _, e = state.GetNodeCtx(nodeKey)
		if e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		return ctx, err
	}

	if storedCtx == nil {
		return ctx, errors.New("stored node context is nil")
	}

	if exactlyResuming {
		storedCtx.NodeCtx.ResumingEvent = resumeEvent
	} else {
		storedCtx.NodeCtx.ResumingEvent = nil
	}

	existingC := GetExeCtx(ctx)
	if existingC != nil {
		storedCtx.RootCtx.ResumeEvent = existingC.RootCtx.ResumeEvent
	}

	currentC := GetExeCtx(ctx)

	if currentC != nil {
		// restore the parent-child relationship between token collectors
		if storedCtx.TokenCollector != nil && storedCtx.TokenCollector.Parent != nil {
			currentTokenCollector := currentC.TokenCollector
			storedCtx.TokenCollector.Parent = currentTokenCollector
		}

		storedCtx.AppVarStore = currentC.AppVarStore
		storedCtx.executed = currentC.executed
	}

	storedCtx.NodeCtx.CurrentRetryCount = 0

	return context.WithValue(ctx, contextKey{}, storedCtx), nil
}

func tryRestoreNodeCtx(ctx context.Context, nodeKey vo.NodeKey) (context.Context, bool) {
	var storedCtx *Context
	err := compose.ProcessState[ExeContextStore](ctx, func(ctx context.Context, state ExeContextStore) error {
		if state == nil {
			return errors.New("state is nil")
		}
		var e error
		storedCtx, _, e = state.GetNodeCtx(nodeKey)
		if e != nil {
			return e
		}
		return nil
	})
	if err != nil || storedCtx == nil {
		return ctx, false
	}

	storedCtx.NodeCtx.ResumingEvent = nil

	existingC := GetExeCtx(ctx)
	if existingC != nil {
		storedCtx.RootCtx.ResumeEvent = existingC.RootCtx.ResumeEvent
		storedCtx.AppVarStore = existingC.AppVarStore
	}

	// restore the parent-child relationship between token collectors
	if storedCtx.TokenCollector != nil && storedCtx.TokenCollector.Parent != nil && existingC != nil {
		currentTokenCollector := existingC.TokenCollector
		storedCtx.TokenCollector.Parent = currentTokenCollector

		storedCtx.AppVarStore = existingC.AppVarStore
		storedCtx.executed = existingC.executed
	}

	storedCtx.NodeCtx.CurrentRetryCount = 0

	return context.WithValue(ctx, contextKey{}, storedCtx), true
}

func PrepareRootExeCtx(ctx context.Context, h *WorkflowHandler) (context.Context, error) {
	var parentTokenCollector *TokenCollector
	if currentC := GetExeCtx(ctx); currentC != nil {
		parentTokenCollector = currentC.TokenCollector
	}

	rootExeCtx := &Context{
		RootCtx: RootCtx{
			RootWorkflowBasic: h.rootWorkflowBasic,
			RootExecuteID:     h.rootExecuteID,
			ResumeEvent:       h.resumeEvent,
			ExeCfg:            h.exeCfg,
		},

		TokenCollector: newTokenCollector(fmt.Sprintf("wf_%d", h.rootWorkflowBasic.ID), parentTokenCollector),
		StartTime:      time.Now().UnixMilli(),
		AppVarStore:    NewAppVariables(),
		executed:       ptr.Of(atomic.Int64{}),
	}

	if h.requireCheckpoint {
		rootExeCtx.CheckPointID = strconv.FormatInt(h.rootExecuteID, 10)
		err := compose.ProcessState[ExeContextStore](ctx, func(ctx context.Context, state ExeContextStore) error {
			if state == nil {
				return errors.New("state is nil")
			}
			return state.SetWorkflowCtx(rootExeCtx)
		})
		if err != nil {
			logs.Errorf("PrepareRootExeCtx error ProcessState: %v", err)
			return ctx, err
		}
	}

	return context.WithValue(ctx, contextKey{}, rootExeCtx), nil
}

func GetExeCtx(ctx context.Context) *Context {
	c := ctx.Value(contextKey{})
	if c == nil {
		return nil
	}
	return c.(*Context)
}

func PrepareSubExeCtx(ctx context.Context, wb *entity.WorkflowBasic, requireCheckpoint bool) (context.Context, error) {
	c := GetExeCtx(ctx)
	if c == nil {
		return ctx, nil
	}

	subExecuteID, err := workflow.GetRepository().GenID(ctx)
	if err != nil {
		logs.Errorf("PrepareSubExeCtx error GenID: %v", err)
		return nil, err
	}

	var newCheckpointID string
	if len(c.CheckPointID) > 0 {
		newCheckpointID = c.CheckPointID + "_" + strconv.FormatInt(subExecuteID, 10)
	}

	newC := &Context{
		RootCtx: c.RootCtx,
		SubWorkflowCtx: &SubWorkflowCtx{
			SubWorkflowBasic: wb,
			SubExecuteID:     subExecuteID,
		},
		NodeCtx:        c.NodeCtx,
		BatchInfo:      c.BatchInfo,
		TokenCollector: newTokenCollector(fmt.Sprintf("sub_wf_%d", wb.ID), c.TokenCollector),
		CheckPointID:   newCheckpointID,
		StartTime:      time.Now().UnixMilli(),
		AppVarStore:    c.AppVarStore,
		executed:       c.executed,
	}

	if requireCheckpoint {
		err := compose.ProcessState[ExeContextStore](ctx, func(ctx context.Context, state ExeContextStore) error {
			if state == nil {
				return errors.New("state is nil")
			}
			return state.SetWorkflowCtx(newC)
		})
		if err != nil {
			logs.Errorf("PrepareSubExeCtx error ProcessState: %v", err)
			return ctx, err
		}
	}

	newC.NodeCtx.SubWorkflowExeID = subExecuteID

	return context.WithValue(ctx, contextKey{}, newC), nil
}

func PrepareNodeExeCtx(ctx context.Context, nodeKey vo.NodeKey, nodeName string, nodeType entity.NodeType, plan *vo.TerminatePlan) (context.Context, error) {
	c := GetExeCtx(ctx)
	if c == nil {
		return ctx, nil
	}
	nodeExecuteID, err := workflow.GetRepository().GenID(ctx)
	if err != nil {
		return nil, err
	}

	newC := &Context{
		RootCtx:        c.RootCtx,
		SubWorkflowCtx: c.SubWorkflowCtx,
		NodeCtx: &NodeCtx{
			NodeKey:       nodeKey,
			NodeExecuteID: nodeExecuteID,
			NodeName:      nodeName,
			NodeType:      nodeType,
			TerminatePlan: plan,
		},
		BatchInfo:    c.BatchInfo,
		StartTime:    time.Now().UnixMilli(),
		CheckPointID: c.CheckPointID,
		AppVarStore:  c.AppVarStore,
		executed:     c.executed,
	}

	if c.NodeCtx == nil { // node within top level workflow, also not under composite node
		newC.NodeCtx.NodePath = []string{string(nodeKey)}
	} else {
		if c.BatchInfo != nil && c.BatchInfo.CompositeNodeKey == c.NodeCtx.NodeKey {
			newC.NodeCtx.NodePath = append(c.NodeCtx.NodePath, InterruptEventIndexPrefix+strconv.Itoa(c.BatchInfo.Index), string(nodeKey))
		} else {
			newC.NodeCtx.NodePath = append(c.NodeCtx.NodePath, string(nodeKey))
		}
	}

	tc := c.TokenCollector
	if entity.NodeMetaByNodeType(nodeType).MayUseChatModel {
		tc = newTokenCollector(strings.Join(append([]string{string(newC.NodeType)}, newC.NodeCtx.NodePath...), "."), c.TokenCollector)
	}
	newC.TokenCollector = tc

	return context.WithValue(ctx, contextKey{}, newC), nil
}

func InheritExeCtxWithBatchInfo(ctx context.Context, index int, items map[string]any) (context.Context, string) {
	c := GetExeCtx(ctx)
	if c == nil {
		return ctx, ""
	}
	var newCheckpointID string
	if len(c.CheckPointID) > 0 {
		newCheckpointID = c.CheckPointID
		if c.SubWorkflowCtx != nil {
			newCheckpointID += "_" + strconv.Itoa(int(c.SubWorkflowCtx.SubExecuteID))
		}
		newCheckpointID += "_" + string(c.NodeCtx.NodeKey)
		newCheckpointID += "_" + strconv.Itoa(index)
	}
	return context.WithValue(ctx, contextKey{}, &Context{
		RootCtx:        c.RootCtx,
		SubWorkflowCtx: c.SubWorkflowCtx,
		NodeCtx:        c.NodeCtx,
		TokenCollector: c.TokenCollector,
		BatchInfo: &BatchInfo{
			Index:            index,
			Items:            items,
			CompositeNodeKey: c.NodeCtx.NodeKey,
		},
		CheckPointID: newCheckpointID,
		AppVarStore:  c.AppVarStore,
		executed:     c.executed,
	}), newCheckpointID
}

type ExeContextStore interface {
	GetNodeCtx(key vo.NodeKey) (*Context, bool, error)
	SetNodeCtx(key vo.NodeKey, value *Context) error
	GetWorkflowCtx() (*Context, bool, error)
	SetWorkflowCtx(value *Context) error
}

type AppVariables struct {
	Vars map[string]any
	mu   sync.RWMutex
}

func NewAppVariables() *AppVariables {
	return &AppVariables{
		Vars: make(map[string]any),
	}
}

func (av *AppVariables) Set(key string, value any) {
	av.mu.Lock()
	av.Vars[key] = value
	av.mu.Unlock()
}

func (av *AppVariables) Get(key string) (any, bool) {
	av.mu.RLock()
	defer av.mu.RUnlock()

	if value, ok := av.Vars[key]; ok {
		return value, ok
	}
	return nil, false
}

func GetAppVarStore(ctx context.Context) *AppVariables {
	c := ctx.Value(contextKey{})
	if c == nil {
		return nil
	}
	return c.(*Context).AppVarStore
}
