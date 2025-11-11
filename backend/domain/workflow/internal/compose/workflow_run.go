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

// Package compose 提供工作流执行的组合和运行管理
//
// 这个包是工作流执行引擎的核心组件，负责：
// 1. 工作流的创建和初始化
// 2. 执行环境的准备和配置
// 3. 工作流运行器的管理
// 4. 工作流恢复和中断处理
// 5. 流式执行和事件处理的集成
//
// 主要组件：
// - WorkflowRunner: 工作流运行器，封装执行逻辑
// - workflowRunOptions: 执行选项配置
// - 各种选项函数：用于灵活配置执行参数
//
// 在Coze Studio项目中，这个包是工作流执行的基础，
// 提供了从前端画布到实际执行的桥梁功能。
package compose

import (
	"context"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/coze-dev/coze-studio/backend/types/consts"

	einoCompose "github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	model "github.com/coze-dev/coze-studio/backend/crossdomain/workflow/model"
	wf "github.com/coze-dev/coze-studio/backend/domain/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/execute"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes/qa"
	schema2 "github.com/coze-dev/coze-studio/backend/domain/workflow/internal/schema"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ternary"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
	"github.com/coze-dev/coze-studio/backend/pkg/safego"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

// WorkflowRunner 工作流运行器
//
// 这是工作流执行的核心结构体，封装了工作流执行所需的所有状态和配置。
// 它负责管理工作流的执行过程，包括初始化、运行、恢复和中断处理。
//
// 主要职责：
// 1. 封装工作流的基本信息（basic）和执行配置（config）
// 2. 管理输入参数（input）和恢复请求（resumeReq）
// 3. 持有工作流schema（schema）和流式写入器（sw）
// 4. 处理执行ID、事件通道和中断事件
//
// 使用场景：
// - 同步执行：通过Prepare方法准备执行环境，然后调用eino的SyncRun
// - 异步执行：通过Prepare方法获取选项，然后调用eino的AsyncRun
// - 恢复执行：当工作流被中断后，从中断点重新开始执行
//
// 生命周期：
// 1. 创建实例（NewWorkflowRunner）
// 2. 准备执行环境（Prepare方法）
// 3. 执行工作流（通过外部调用eino的执行方法）
// 4. 处理执行结果和事件
type WorkflowRunner struct {
	// basic 工作流基本信息
	// 包含工作流ID、版本、空间ID等基础元数据
	basic *entity.WorkflowBasic

	// input 工作流输入参数的JSON字符串
	// 存储用户传递给工作流的输入数据
	input string

	// resumeReq 恢复请求
	// 当工作流需要从中断点恢复时使用，包含恢复所需的信息
	resumeReq *entity.ResumeRequest

	// schema 工作流执行schema
	// 定义了工作流的节点、连接关系和执行逻辑
	schema *schema2.WorkflowSchema

	// sw 流式写入器
	// 用于支持流式输出，将执行过程中的消息实时发送给客户端
	sw *schema.StreamWriter[*entity.Message]

	// container 流式容器
	// 管理流式数据的管道和处理逻辑
	container *execute.StreamContainer

	// config 执行配置
	// 包含执行模式、版本、应用ID等执行参数
	config model.ExecuteConfig

	// executeID 执行ID
	// 唯一标识本次工作流执行实例
	executeID int64

	// eventChan 执行事件通道
	// 用于接收工作流执行过程中的各种事件（开始、完成、错误等）
	eventChan chan *execute.Event

	// interruptEvent 中断事件
	// 当工作流执行被中断时，存储中断的具体信息
	interruptEvent *entity.InterruptEvent
}

// workflowRunOptions 工作流运行选项配置
//
// 这个结构体用于配置WorkflowRunner的创建参数，采用选项模式（Option Pattern）
// 使得WorkflowRunner的创建更加灵活和可扩展。
//
// 主要配置项：
// - input: 工作流输入参数
// - resumeReq: 恢复请求（用于恢复中断的工作流）
// - streamWriter: 流式写入器（用于流式输出）
// - rootTokenCollector: 根token收集器（用于统计token使用量）
type workflowRunOptions struct {
	// input 工作流输入参数的JSON字符串
	input string

	// resumeReq 恢复请求
	// 当需要恢复之前中断的工作流时使用
	resumeReq *entity.ResumeRequest

	// streamWriter 流式写入器
	// 用于支持流式输出，将执行结果实时推送给客户端
	streamWriter *schema.StreamWriter[*entity.Message]

	// rootTokenCollector 根token收集器
	// 用于收集和统计工作流执行过程中的token使用情况
	rootTokenCollector *execute.TokenCollector
}

// WorkflowRunnerOption 工作流运行器选项函数类型
//
// 使用函数选项模式（Functional Options Pattern）来配置WorkflowRunner。
// 这种设计模式提供了良好的API扩展性，无需修改构造函数即可添加新的配置项。
type WorkflowRunnerOption func(*workflowRunOptions)

// WithInput 设置工作流输入参数
//
// 用于指定传递给工作流的输入数据。这个输入数据通常是JSON格式的字符串，
// 包含了工作流执行所需的所有参数。
//
// 参数：
//   - input: 工作流输入参数的JSON字符串
//
// 返回：
//   - WorkflowRunnerOption: 配置函数
func WithInput(input string) WorkflowRunnerOption {
	return func(opts *workflowRunOptions) {
		opts.input = input
	}
}

// WithResumeReq 设置恢复请求
//
// 当工作流执行被中断后，需要从中断点恢复时使用这个选项。
// 恢复请求包含了中断事件的ID和恢复所需的数据。
//
// 参数：
//   - resumeReq: 恢复请求对象，包含中断事件信息和恢复数据
//
// 返回：
//   - WorkflowRunnerOption: 配置函数
func WithResumeReq(resumeReq *entity.ResumeRequest) WorkflowRunnerOption {
	return func(opts *workflowRunOptions) {
		opts.resumeReq = resumeReq
	}
}

// WithStreamWriter 设置流式写入器
//
// 用于启用流式输出功能，将工作流执行过程中的结果实时推送给客户端。
// 这对于需要实时反馈的工作流（如聊天机器人）非常重要。
//
// 参数：
//   - sw: 流式写入器，用于管理流式数据的输出
//
// 返回：
//   - WorkflowRunnerOption: 配置函数
func WithStreamWriter(sw *schema.StreamWriter[*entity.Message]) WorkflowRunnerOption {
	return func(opts *workflowRunOptions) {
		opts.streamWriter = sw
	}
}

// NewWorkflowRunner 创建新的工作流运行器
//
// 这是一个工厂函数，用于创建WorkflowRunner实例。使用选项模式（Functional Options Pattern）
// 使得创建过程更加灵活，可以根据需要配置不同的参数。
//
// 创建流程：
// 1. 初始化默认选项
// 2. 应用所有传入的选项函数
// 3. 如果配置了流式写入器，则创建对应的流式容器
// 4. 返回配置完整的WorkflowRunner实例
//
// 参数：
//   - b: 工作流基本信息，包含ID、版本、空间等元数据
//   - sc: 工作流schema，定义了工作流的执行逻辑和结构
//   - config: 执行配置，包含执行模式、版本、权限等信息
//   - opts: 可选的配置选项，用于定制WorkflowRunner的行为
//
// 返回：
//   - *WorkflowRunner: 配置完成的工作流运行器实例
//
// 示例：
//
//	runner := NewWorkflowRunner(basic, schema, config,
//	    WithInput(inputJson),
//	    WithStreamWriter(streamWriter),
//	    WithResumeReq(resumeReq))
func NewWorkflowRunner(b *entity.WorkflowBasic, sc *schema2.WorkflowSchema, config model.ExecuteConfig, opts ...WorkflowRunnerOption) *WorkflowRunner {
	// 初始化默认选项
	options := &workflowRunOptions{}

	// 应用所有选项函数
	for _, opt := range opts {
		opt(options)
	}

	// 如果配置了流式写入器，则创建对应的流式容器
	var container *execute.StreamContainer
	if options.streamWriter != nil {
		container = execute.NewStreamContainer(options.streamWriter)
	}

	// 返回配置完整的工作流运行器
	return &WorkflowRunner{
		basic:     b,
		input:     options.input,
		resumeReq: options.resumeReq,
		schema:    sc,
		sw:        options.streamWriter,
		container: container,
		config:    config,
	}
}

// Prepare 准备工作流执行环境
//
// 这是WorkflowRunner的核心方法，负责在工作流执行前进行所有必要的准备工作。
// 该方法会初始化执行ID、设置中断处理、配置执行选项，并启动事件处理goroutine。
//
// 主要职责：
// 1. 生成或获取执行ID
// 2. 处理恢复请求（如果有的话）
// 3. 设置中断事件的状态修改器
// 4. 创建工作流执行记录
// 5. 配置超时和取消机制
// 6. 启动事件处理goroutine
//
// 执行流程：
// 1. ID生成：新执行生成新ID，恢复执行使用已有ID
// 2. 中断处理：获取中断事件，验证恢复请求，设置状态修改器
// 3. 选项配置：调用designateOptions获取eino执行选项
// 4. 执行记录：为新执行创建WorkflowExecution记录
// 5. 上下文管理：设置超时和取消机制
// 6. 事件处理：启动goroutine处理执行事件
//
// 参数：
//   - ctx: 原始上下文，用于传递请求信息和取消信号
//
// 返回：
//   - context.Context: 可取消的上下文，包含超时设置
//   - int64: 工作流执行ID，用于唯一标识本次执行
//   - []einoCompose.Option: eino框架的执行选项配置
//   - <-chan *execute.Event: 执行事件通道，用于接收执行状态变化
//   - error: 准备过程中的错误
//
// 注意：
//   - 该方法会启动goroutine进行事件处理
//   - 返回的上下文包含超时机制，可用于取消执行
//   - 中断恢复时会锁定执行记录，防止并发冲突
//   - 流式容器会在出错时自动清理
func (r *WorkflowRunner) Prepare(ctx context.Context) (
	context.Context,
	int64,
	[]einoCompose.Option,
	<-chan *execute.Event,
	error,
) {
	// 初始化变量
	var (
		err       error
		executeID int64
		repo      = wf.GetRepository() // 获取工作流仓储实例
		resumeReq = r.resumeReq        // 恢复请求
		wb        = r.basic            // 工作流基本信息
		sc        = r.schema           // 工作流schema
		sw        = r.sw               // 流式写入器
		container = r.container        // 流式容器
		config    = r.config           // 执行配置
	)

	// 生成执行ID：新执行生成新ID，恢复执行使用已有ID
	if r.resumeReq == nil {
		// 新执行：生成唯一的执行ID
		executeID, err = repo.GenID(ctx)
		if err != nil {
			return ctx, 0, nil, nil, vo.WrapError(errno.ErrIDGenError,
				fmt.Errorf("failed to generate workflow execute ID: %w", err))
		}
	} else {
		// 恢复执行：使用恢复请求中的执行ID
		executeID = resumeReq.ExecuteID
	}

	// 创建执行事件通道，用于接收工作流执行过程中的事件
	eventChan := make(chan *execute.Event)

	// 处理中断事件（仅在恢复执行时需要）
	var (
		interruptEvent *entity.InterruptEvent
		found          bool
	)

	if resumeReq != nil {
		// 获取中断事件：恢复执行时需要找到之前的中断点
		interruptEvent, found, err = repo.GetFirstInterruptEvent(ctx, executeID)
		if err != nil {
			return ctx, 0, nil, nil, err
		}

		// 验证中断事件存在性
		if !found {
			return ctx, 0, nil, nil, fmt.Errorf("interrupt event does not exist, id: %d", resumeReq.EventID)
		}

		// 验证中断事件ID匹配
		if interruptEvent.ID != resumeReq.EventID {
			return ctx, 0, nil, nil, fmt.Errorf("interrupt event id mismatch, expect: %d, actual: %d", resumeReq.EventID, interruptEvent.ID)
		}
	}

	// 设置WorkflowRunner的内部状态
	r.executeID = executeID
	r.eventChan = eventChan
	r.interruptEvent = interruptEvent

	// 启动流式容器（如果配置了流式输出）
	if container != nil {
		// 在后台启动流式数据管道
		go container.PipeAll()

		// 错误处理：如果后续步骤出错，确保容器被正确清理
		defer func() {
			if err != nil {
				container.Done()
			}
		}()
	}

	// 获取eino框架的执行选项配置
	composeOpts, err := r.designateOptions(ctx)
	if err != nil {
		return ctx, 0, nil, nil, err
	}

	// 设置中断恢复的状态修改器（仅在恢复执行时需要）
	if interruptEvent != nil {
		var stateOpt einoCompose.Option

		// 生成状态修改器：根据中断事件类型创建相应的状态恢复逻辑
		stateModifier := GenStateModifierByEventType(interruptEvent.EventType,
			interruptEvent.NodeKey, resumeReq.ResumeData, r.config)

		// 根据中断事件在工作流中的位置设置不同的恢复选项
		if len(interruptEvent.NodePath) == 1 {
			// 中断事件在顶级工作流中
			stateOpt = einoCompose.WithStateModifier(stateModifier)
		} else {
			// 中断事件在嵌套结构中（复合节点或子工作流）
			currentI := len(interruptEvent.NodePath) - 2
			path := interruptEvent.NodePath[currentI]

			if strings.HasPrefix(path, execute.InterruptEventIndexPrefix) {
				// 中断事件在复合节点中（如循环、批量节点）
				indexStr := path[len(execute.InterruptEventIndexPrefix):]
				index, err := strconv.Atoi(indexStr)
				if err != nil {
					return ctx, 0, nil, nil, fmt.Errorf("failed to parse index: %w", err)
				}

				// 设置恢复索引和状态修改器
				currentI--
				parentNodeKey := interruptEvent.NodePath[currentI]
				stateOpt = einoCompose.WithLambdaOption(
					nodes.WithResumeIndex(index, stateModifier)).DesignateNode(parentNodeKey)
			} else {
				// 中断事件在子工作流中
				subWorkflowNodeKey := interruptEvent.NodePath[currentI]
				stateOpt = einoCompose.WithLambdaOption(
					nodes.WithResumeIndex(0, stateModifier)).DesignateNode(subWorkflowNodeKey)
			}

			// 处理更深层次的嵌套结构
			for i := currentI - 1; i >= 0; i-- {
				path := interruptEvent.NodePath[i]
				if strings.HasPrefix(path, execute.InterruptEventIndexPrefix) {
					// 复合节点嵌套
					indexStr := path[len(execute.InterruptEventIndexPrefix):]
					index, err := strconv.Atoi(indexStr)
					if err != nil {
						return ctx, 0, nil, nil, fmt.Errorf("failed to parse index: %w", err)
					}

					i--
					parentNodeKey := interruptEvent.NodePath[i]
					stateOpt = WrapOptWithIndex(stateOpt, vo.NodeKey(parentNodeKey), index)
				} else {
					// 子工作流嵌套
					stateOpt = WrapOpt(stateOpt, vo.NodeKey(path))
				}
			}
		}

		// 将状态选项添加到执行选项中
		composeOpts = append(composeOpts, stateOpt)

		// 处理中断数据更新（问题类型中断）
		if interruptEvent.EventType == entity.InterruptEventQuestion {
			// 合并问题回答数据到中断事件中
			modifiedData, err := qa.AppendInterruptData(interruptEvent.InterruptData, resumeReq.ResumeData)
			if err != nil {
				return ctx, 0, nil, nil, fmt.Errorf("failed to append interrupt data: %w", err)
			}
			interruptEvent.InterruptData = modifiedData

			// 更新中断事件到数据库
			if err = repo.UpdateFirstInterruptEvent(ctx, executeID, interruptEvent); err != nil {
				return ctx, 0, nil, nil, fmt.Errorf("failed to update interrupt event: %w", err)
			}
		} else if interruptEvent.EventType == entity.InterruptEventLLM &&
			interruptEvent.ToolInterruptEvent.EventType == entity.InterruptEventQuestion {
			// 处理LLM工具调用的中断数据
			modifiedData, err := qa.AppendInterruptData(interruptEvent.ToolInterruptEvent.InterruptData, resumeReq.ResumeData)
			if err != nil {
				return ctx, 0, nil, nil, fmt.Errorf("failed to append interrupt data for LLM node: %w", err)
			}
			interruptEvent.ToolInterruptEvent.InterruptData = modifiedData

			// 更新中断事件到数据库
			if err = repo.UpdateFirstInterruptEvent(ctx, executeID, interruptEvent); err != nil {
				return ctx, 0, nil, nil, fmt.Errorf("failed to update interrupt event: %w", err)
			}
		}

		// 锁定工作流执行，防止并发恢复操作
		success, currentStatus, err := repo.TryLockWorkflowExecution(ctx, executeID, resumeReq.EventID)
		if err != nil {
			return ctx, 0, nil, nil, fmt.Errorf("try lock workflow execution unexpected err: %w", err)
		}

		// 验证锁定是否成功
		if !success {
			return ctx, 0, nil, nil, fmt.Errorf("workflow execution lock failed, current status is %v, executeID: %d", currentStatus, executeID)
		}

		// 记录恢复操作日志
		logs.CtxInfof(ctx, "resuming with eventID: %d, executeID: %d, nodeKey: %s", interruptEvent.ID,
			executeID, interruptEvent.NodeKey)
	}

	// 创建工作流执行记录（仅在非恢复执行时需要）
	if interruptEvent == nil {
		// 获取日志ID用于追踪
		var logID string
		logID, _ = ctx.Value(consts.CtxLogIDKey).(string)

		// 创建工作流执行实体
		wfExec := &entity.WorkflowExecution{
			ID:                     executeID,              // 执行ID
			WorkflowID:             wb.ID,                  // 工作流ID
			Version:                wb.Version,             // 工作流版本
			SpaceID:                wb.SpaceID,             // 所属空间
			ExecuteConfig:          config,                 // 执行配置
			Status:                 entity.WorkflowRunning, // 初始状态为运行中
			Input:                  ptr.Of(r.input),        // 输入参数
			RootExecutionID:        executeID,              // 根执行ID
			NodeCount:              sc.NodeCount(),         // 节点数量
			CurrentResumingEventID: ptr.Of(int64(0)),       // 当前恢复事件ID（初始为0）
			CommitID:               wb.CommitID,            // 提交ID
			LogID:                  logID,                  // 日志ID
		}

		// 保存执行记录到数据库
		if err = repo.CreateWorkflowExecution(ctx, wfExec); err != nil {
			return ctx, 0, nil, nil, err
		}
	}

	// 设置上下文和超时控制
	cancelCtx, cancelFn := context.WithCancel(ctx) // 创建可取消的上下文
	var timeoutFn context.CancelFunc
	if s := execute.GetStaticConfig(); s != nil {
		// 根据任务类型设置不同的超时时间
		timeout := ternary.IFElse(config.TaskType == model.TaskTypeBackground, s.BackgroundRunTimeout, s.ForegroundRunTimeout)
		if timeout > 0 {
			// 添加超时控制
			cancelCtx, timeoutFn = context.WithTimeout(cancelCtx, timeout)
		}
	}

	// 创建最后事件通道，用于接收工作流执行的最终结果
	lastEventChan := make(chan *execute.Event, 1)

	// 启动事件处理goroutine
	go func() {
		// panic恢复机制：防止事件处理过程中的异常影响整个系统
		defer func() {
			if panicErr := recover(); panicErr != nil {
				logs.CtxErrorf(ctx, "panic when handling execute event: %v", safego.NewPanicErr(panicErr, debug.Stack()))
			}
		}()

		// 确保流式容器被正确清理
		defer func() {
			if container != nil {
				container.Done()
			}
		}()

		// 注意：此goroutine不使用cancelCtx，因为它需要保持活跃来接收工作流取消事件
		// 它监听eventChan中的事件，并调用cancelFn或timeoutFn来控制执行
		lastEventChan <- execute.HandleExecuteEvent(ctx, executeID, eventChan, cancelFn, timeoutFn,
			repo, sw, config)
		close(lastEventChan)
	}()

	// 返回准备完成的所有资源
	return cancelCtx, executeID, composeOpts, lastEventChan, nil
}
