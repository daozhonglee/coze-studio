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

// Package service 提供工作流执行服务的实现
//
// 这个包实现了工作流的各种执行模式，包括：
// - 同步执行（SyncExecute）：立即返回执行结果
// - 异步执行（AsyncExecute）：后台执行，返回执行ID
// - 单节点调试（AsyncExecuteNode）：只执行指定节点
// - 流式执行（StreamExecute）：实时流式返回执行事件
// - 执行恢复（AsyncResume/StreamResume）：从中断处继续执行
// - 执行取消（Cancel）：取消正在运行的执行
//
// 核心功能：
// 1. 工作流画布到可执行schema的转换
// 2. 输入参数验证和转换
// 3. 执行引擎的初始化和运行
// 4. 执行结果的收集和返回
// 5. 错误处理和状态管理
//
// 执行流程：
// 1. 获取工作流实体和画布
// 2. 转换为WorkflowSchema
// 3. 创建执行引擎（workflow）
// 4. 转换输入参数
// 5. 执行并返回结果
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coze-dev/coze-studio/backend/types/consts"

	einoCompose "github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	workflowapimodel "github.com/coze-dev/coze-studio/backend/api/model/workflow"
	crossmessage "github.com/coze-dev/coze-studio/backend/crossdomain/message"
	workflowModel "github.com/coze-dev/coze-studio/backend/crossdomain/workflow/model"
	"github.com/coze-dev/coze-studio/backend/domain/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/canvas/adaptor"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/compose"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/execute"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes"
	"github.com/coze-dev/coze-studio/backend/pkg/errorx"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/slices"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
	"github.com/coze-dev/coze-studio/backend/pkg/sonic"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

// executableImpl 工作流执行服务的具体实现
//
// 这个结构体实现了ExecutableService接口，提供工作流的各种执行功能。
// 它是工作流执行的核心服务实现，负责协调各个组件完成工作流执行任务。
type executableImpl struct {
	// repo 工作流数据仓库接口
	// 用于访问工作流实体、执行记录等持久化数据
	repo workflow.Repository
}

// SyncExecute 同步执行工作流
//
// 这是工作流执行的主要入口方法，以同步方式执行完整的工作流并返回执行结果。
// 调用者会等待整个工作流执行完成，然后获得最终的执行结果。
//
// 执行流程：
// 1. 验证执行配置和权限
// 2. 获取工作流实体和画布
// 3. 转换为WorkflowSchema
// 4. 创建执行引擎
// 5. 转换输入参数
// 6. 执行工作流
// 7. 收集和返回执行结果
//
// 参数：
//   - ctx: 上下文，用于取消和超时控制
//   - config: 执行配置，包含工作流ID、版本、模式等
//   - input: 工作流输入参数
//
// 返回：
//   - *entity.WorkflowExecution: 完整的执行结果
//   - vo.TerminatePlan: 工作流的终止计划（返回变量或直接输出）
//   - error: 执行过程中的错误
//
// 注意：
//   - 这个方法会阻塞直到工作流执行完成
//   - 支持应用工作流的版本验证
//   - 自动处理文件输入参数
//   - 包含完整的错误处理和状态转换
func (i *impl) SyncExecute(ctx context.Context, config workflowModel.ExecuteConfig, input map[string]any) (*entity.WorkflowExecution, vo.TerminatePlan, error) {
	var (
		err      error
		wfEntity *entity.Workflow
	)

	wfEntity, err = i.Get(ctx, &vo.GetPolicy{
		ID:       config.ID,
		QType:    config.From,
		MetaOnly: false,
		Version:  config.Version,
		CommitID: config.CommitID,
	})
	if err != nil {
		return nil, "", err
	}

	config.WorkflowMode = wfEntity.Mode

	isApplicationWorkflow := wfEntity.AppID != nil
	if isApplicationWorkflow && config.Mode == workflowModel.ExecuteModeRelease {
		err = i.checkApplicationWorkflowReleaseVersion(ctx, *wfEntity.AppID, config.ConnectorID, config.ID, config.Version)
		if err != nil {
			return nil, "", err
		}
	}

	c := &vo.Canvas{}
	if err = sonic.UnmarshalString(wfEntity.Canvas, c); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal canvas: %w", err)
	}

	workflowSC, err := adaptor.CanvasToWorkflowSchema(ctx, c)
	if err != nil {
		return nil, "", fmt.Errorf("failed to convert canvas to workflow schema: %w", err)
	}

	config.InputFileFields = slices.ToMap(workflowSC.GetAllNodesInputFileFields(ctx), func(e *workflowModel.FileInfo) (string, *workflowModel.FileInfo) {
		return e.FileURL, e
	})
	var wfOpts []compose.WorkflowOption
	wfOpts = append(wfOpts, compose.WithIDAsName(wfEntity.ID))
	if s := execute.GetStaticConfig(); s != nil && s.MaxNodeCountPerWorkflow > 0 {
		wfOpts = append(wfOpts, compose.WithMaxNodeCount(s.MaxNodeCountPerWorkflow))
	}

	wf, err := compose.NewWorkflow(ctx, workflowSC, wfOpts...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create workflow: %w", err)
	}

	if wfEntity.AppID != nil && config.AppID == nil {
		config.AppID = wfEntity.AppID
	}

	var cOpts []nodes.ConvertOption
	inputFileFields := make(map[string]*workflowModel.FileInfo)
	cOpts = append(cOpts, nodes.WithCollectFileFields(inputFileFields), nodes.WithNotNeedTrimQueryFileName(true))
	if config.InputFailFast {
		cOpts = append(cOpts, nodes.FailFast())
	}

	convertedInput, ws, err := nodes.ConvertInputs(ctx, input, wf.Inputs(), cOpts...)
	if err != nil {
		return nil, "", err
	} else if ws != nil {
		logs.CtxWarnf(ctx, "convert inputs warnings: %v", *ws)
	}

	for k, v := range inputFileFields {
		config.InputFileFields[k] = v
	}

	inStr, err := sonic.MarshalString(input)
	if err != nil {
		return nil, "", err
	}

	cancelCtx, executeID, opts, lastEventChan, err := compose.NewWorkflowRunner(wfEntity.GetBasic(), workflowSC, config,
		compose.WithInput(inStr)).Prepare(ctx)
	if err != nil {
		return nil, "", err
	}

	startTime := time.Now()

	out, err := wf.SyncRun(cancelCtx, convertedInput, opts...)
	if err != nil {
		if _, ok := einoCompose.ExtractInterruptInfo(err); !ok {
			var wfe vo.WorkflowError
			if errors.As(err, &wfe) {
				return nil, "", wfe.AppendDebug(executeID, wfEntity.SpaceID, wfEntity.ID)
			} else {
				return nil, "", vo.WrapWithDebug(errno.ErrWorkflowExecuteFail, err, executeID, wfEntity.SpaceID, wfEntity.ID, errorx.KV("cause", err.Error()))
			}
		}
	}

	lastEvent := <-lastEventChan

	updateTime := time.Now()

	var outStr string
	if wf.TerminatePlan() == vo.ReturnVariables {
		outStr, err = sonic.MarshalString(out)
		if err != nil {
			return nil, "", err
		}
	} else {
		outStr = out["output"].(string)
	}

	var status entity.WorkflowExecuteStatus
	switch lastEvent.Type {
	case execute.WorkflowSuccess:
		status = entity.WorkflowSuccess
	case execute.WorkflowInterrupt:
		status = entity.WorkflowInterrupted
	case execute.WorkflowFailed:
		status = entity.WorkflowFailed
	case execute.WorkflowCancel:
		status = entity.WorkflowCancel
	}

	var failReason *string
	if lastEvent.Err != nil {
		failReason = ptr.Of(lastEvent.Err.Error())
	}

	return &entity.WorkflowExecution{
		ID:            executeID,
		WorkflowID:    wfEntity.ID,
		Version:       wfEntity.GetVersion(),
		SpaceID:       wfEntity.SpaceID,
		ExecuteConfig: config,
		CreatedAt:     startTime,
		NodeCount:     workflowSC.NodeCount(),
		Status:        status,
		Duration:      lastEvent.Duration,
		Input:         ptr.Of(inStr),
		Output:        ptr.Of(outStr),
		ErrorCode:     ptr.Of("-1"),
		FailReason:    failReason,
		TokenInfo: &entity.TokenUsage{
			InputTokens:  lastEvent.GetInputTokens(),
			OutputTokens: lastEvent.GetOutputTokens(),
		},
		UpdatedAt:       ptr.Of(updateTime),
		RootExecutionID: executeID,
		InterruptEvents: lastEvent.InterruptEvents,
	}, wf.TerminatePlan(), nil
}

// AsyncExecute 异步执行工作流
//
// 以异步方式启动工作流执行，立即返回执行ID，工作流在后台执行。
// 调用者需要通过GetExecution方法轮询执行状态来获取结果。
//
// 执行流程：
// 1. 验证执行配置和权限
// 2. 获取工作流实体和画布
// 3. 转换为WorkflowSchema
// 4. 创建执行引擎并启动异步执行
// 5. 返回执行ID供后续查询
//
// 参数：
//   - ctx: 上下文，用于取消和超时控制
//   - config: 执行配置，包含工作流ID、版本、模式等
//   - input: 工作流输入参数
//
// 返回：
//   - int64: 工作流执行ID，用于后续查询执行状态
//   - error: 启动执行过程中的错误
//
// 注意：
//   - 方法立即返回，不等待执行完成
//   - 支持调试模式下的测试运行记录
//   - 适用于长时间运行的工作流
func (i *impl) AsyncExecute(ctx context.Context, config workflowModel.ExecuteConfig, input map[string]any) (int64, error) {
	var (
		err      error
		wfEntity *entity.Workflow
	)

	wfEntity, err = i.Get(ctx, &vo.GetPolicy{
		ID:       config.ID,
		QType:    config.From,
		MetaOnly: false,
		Version:  config.Version,
		CommitID: config.CommitID,
	})
	if err != nil {
		return 0, err
	}

	config.WorkflowMode = wfEntity.Mode

	isApplicationWorkflow := wfEntity.AppID != nil
	if isApplicationWorkflow && config.Mode == workflowModel.ExecuteModeRelease {
		err = i.checkApplicationWorkflowReleaseVersion(ctx, *wfEntity.AppID, config.ConnectorID, config.ID, config.Version)
		if err != nil {
			return 0, err
		}
	}

	c := &vo.Canvas{}
	if err = sonic.UnmarshalString(wfEntity.Canvas, c); err != nil {
		return 0, fmt.Errorf("failed to unmarshal canvas: %w", err)
	}

	workflowSC, err := adaptor.CanvasToWorkflowSchema(ctx, c)
	if err != nil {
		return 0, fmt.Errorf("failed to convert canvas to workflow schema: %w", err)
	}

	config.InputFileFields = slices.ToMap(workflowSC.GetAllNodesInputFileFields(ctx), func(e *workflowModel.FileInfo) (string, *workflowModel.FileInfo) {
		return e.FileURL, e
	})

	var wfOpts []compose.WorkflowOption
	wfOpts = append(wfOpts, compose.WithIDAsName(wfEntity.ID))
	if s := execute.GetStaticConfig(); s != nil && s.MaxNodeCountPerWorkflow > 0 {
		wfOpts = append(wfOpts, compose.WithMaxNodeCount(s.MaxNodeCountPerWorkflow))
	}

	wf, err := compose.NewWorkflow(ctx, workflowSC, wfOpts...)
	if err != nil {
		return 0, fmt.Errorf("failed to create workflow: %w", err)
	}

	if wfEntity.AppID != nil && config.AppID == nil {
		config.AppID = wfEntity.AppID
	}

	config.CommitID = wfEntity.CommitID

	var cOpts []nodes.ConvertOption
	inputFileFields := make(map[string]*workflowModel.FileInfo)
	cOpts = append(cOpts, nodes.WithCollectFileFields(inputFileFields), nodes.WithNotNeedTrimQueryFileName(true))
	if config.InputFailFast {
		cOpts = append(cOpts, nodes.FailFast())
	}

	convertedInput, ws, err := nodes.ConvertInputs(ctx, input, wf.Inputs(), cOpts...)
	if err != nil {
		return 0, err
	} else if ws != nil {
		logs.CtxWarnf(ctx, "convert inputs warnings: %v", *ws)
	}

	for k, v := range inputFileFields {
		config.InputFileFields[k] = v
	}

	inStr, err := sonic.MarshalString(input)
	if err != nil {
		return 0, err
	}

	cancelCtx, executeID, opts, _, err := compose.NewWorkflowRunner(wfEntity.GetBasic(), workflowSC, config,
		compose.WithInput(inStr)).Prepare(ctx)
	if err != nil {
		return 0, err
	}

	if config.Mode == workflowModel.ExecuteModeDebug {
		if err = i.repo.SetTestRunLatestExeID(ctx, wfEntity.ID, config.Operator, executeID); err != nil {
			logs.CtxErrorf(ctx, "failed to set test run latest exe id: %v", err)
		}
	}

	wf.AsyncRun(cancelCtx, convertedInput, opts...)

	return executeID, nil
}

// handleHistory 处理聊天历史
//
// 为聊天工作流准备历史对话数据，支持按名称查找或直接使用对话ID。
//
// 处理逻辑：
// 1. 检查是否需要历史数据（historyRounds > 0）
// 2. 如果需要按名称查找对话：
//    - 从输入中提取对话名称
//    - 创建或获取对话
//    - 设置对话ID和分段ID
// 3. 预取聊天历史消息
// 4. 设置配置中的历史数据
//
// 参数：
//   - ctx: 上下文
//   - config: 执行配置，会被修改以包含历史数据
//   - input: 输入参数，包含对话名称等
//   - historyRounds: 需要的历史对话轮数
//   - shouldFetchConversationByName: 是否需要按名称查找对话
//
// 返回：
//   - error: 处理过程中的错误
//
// 注意：
//   - 支持应用ID和代理ID两种业务场景
//   - 会自动过滤无效的对话信息
func (i *impl) handleHistory(ctx context.Context, config *workflowModel.ExecuteConfig, input map[string]any, historyRounds int64, shouldFetchConversationByName bool) error {
	if historyRounds <= 0 {
		return nil
	}

	if shouldFetchConversationByName {
		var cID, sID, bizID int64
		var err error
		if config.AppID != nil {
			bizID = *config.AppID
		} else if config.AgentID != nil {
			bizID = *config.AgentID
		}
		for k, v := range input {
			if k == vo.ConversationNameKey {
				cName, ok := v.(string)
				if !ok {
					return errors.New("CONVERSATION_NAME must be string")
				}
				cID, sID, err = i.GetOrCreateConversation(ctx, vo.Draft, bizID, consts.CozeConnectorID, config.Operator, cName)
				if err != nil {
					return err
				}
				config.ConversationID = ptr.Of(cID)
				config.SectionID = ptr.Of(sID)
			}
		}
	}

	messages, scMessages, err := i.prefetchChatHistory(ctx, *config, historyRounds)
	if err != nil {
		logs.CtxErrorf(ctx, "failed to prefetch chat history: %v", err)
	}

	if len(messages) > 0 {
		config.ConversationHistory = messages
	}

	if len(scMessages) > 0 {
		config.ConversationHistorySchemaMessages = scMessages
	}
	return nil
}

// AsyncExecuteNode 异步执行单个节点
//
// 用于工作流调试功能，只执行指定的单个节点而不是整个工作流。
// 这对于开发和调试阶段非常有用，可以快速验证单个节点的逻辑。
//
// 执行流程：
// 1. 验证执行配置和权限
// 2. 获取工作流实体和画布
// 3. 创建以指定节点为起点的子工作流schema
// 4. 处理聊天历史（如果是聊天流）
// 5. 创建节点执行引擎并启动异步执行
// 6. 记录节点调试的最新执行ID
//
// 参数：
//   - ctx: 上下文，用于取消和超时控制
//   - nodeID: 要执行的节点ID
//   - config: 执行配置，包含工作流ID、版本、模式等
//   - input: 节点输入参数
//
// 返回：
//   - int64: 节点执行ID，用于后续查询执行状态
//   - error: 启动执行过程中的错误
//
// 注意：
//   - 只执行从指定节点开始的子工作流
//   - 支持聊天流的对话历史处理
//   - 在节点调试模式下会记录最新的执行ID
func (i *impl) AsyncExecuteNode(ctx context.Context, nodeID string, config workflowModel.ExecuteConfig, input map[string]any) (int64, error) {
	var (
		err      error
		wfEntity *entity.Workflow
	)

	wfEntity, err = i.Get(ctx, &vo.GetPolicy{
		ID:       config.ID,
		QType:    config.From,
		MetaOnly: false,
		Version:  config.Version,
	})
	if err != nil {
		return 0, err
	}

	config.WorkflowMode = wfEntity.Mode

	isApplicationWorkflow := wfEntity.AppID != nil
	if isApplicationWorkflow && config.Mode == workflowModel.ExecuteModeRelease {
		err = i.checkApplicationWorkflowReleaseVersion(ctx, *wfEntity.AppID, config.ConnectorID, config.ID, config.Version)
		if err != nil {
			return 0, err
		}
	}

	c := &vo.Canvas{}
	if err = sonic.UnmarshalString(wfEntity.Canvas, c); err != nil {
		return 0, fmt.Errorf("failed to unmarshal canvas: %w", err)
	}

	workflowSC, err := adaptor.WorkflowSchemaFromNode(ctx, c, nodeID)
	if err != nil {
		return 0, fmt.Errorf("failed to convert canvas to workflow schema: %w", err)
	}

	historyRounds := int64(0)
	if config.WorkflowMode == workflowapimodel.WorkflowMode_ChatFlow {
		historyRounds = workflowSC.HistoryRounds()
	}
	if historyRounds > 0 {
		if err = i.handleHistory(ctx, &config, input, historyRounds, true); err != nil {
			return 0, err
		}
	}
	config.InputFileFields = slices.ToMap(workflowSC.GetAllNodesInputFileFields(ctx), func(e *workflowModel.FileInfo) (string, *workflowModel.FileInfo) {
		return e.FileURL, e
	})

	wf, err := compose.NewWorkflowFromNode(ctx, workflowSC, vo.NodeKey(nodeID), einoCompose.WithGraphName(fmt.Sprintf("%d", wfEntity.ID)))
	if err != nil {
		return 0, fmt.Errorf("failed to create workflow: %w", err)
	}

	var cOpts []nodes.ConvertOption
	inputFileFields := make(map[string]*workflowModel.FileInfo)
	cOpts = append(cOpts, nodes.WithCollectFileFields(inputFileFields), nodes.WithNotNeedTrimQueryFileName(true))
	if config.InputFailFast {
		cOpts = append(cOpts, nodes.FailFast())
	}

	convertedInput, ws, err := nodes.ConvertInputs(ctx, input, wf.Inputs(), cOpts...)
	if err != nil {
		return 0, err
	} else if ws != nil {
		logs.CtxWarnf(ctx, "convert inputs warnings: %v", *ws)
	}
	for k, v := range inputFileFields {
		config.InputFileFields[k] = v
	}

	if wfEntity.AppID != nil && config.AppID == nil {
		config.AppID = wfEntity.AppID
	}

	config.CommitID = wfEntity.CommitID

	inStr, err := sonic.MarshalString(input)
	if err != nil {
		return 0, err
	}

	cancelCtx, executeID, opts, _, err := compose.NewWorkflowRunner(wfEntity.GetBasic(), workflowSC, config,
		compose.WithInput(inStr)).Prepare(ctx)
	if err != nil {
		return 0, err
	}

	if config.Mode == workflowModel.ExecuteModeNodeDebug {
		if err = i.repo.SetNodeDebugLatestExeID(ctx, wfEntity.ID, nodeID, config.Operator, executeID); err != nil {
			logs.CtxErrorf(ctx, "failed to set node debug latest exe id: %v", err)
		}
	}

	wf.AsyncRun(cancelCtx, convertedInput, opts...)

	return executeID, nil
}

// StreamExecute 流式执行工作流
//
// 执行工作流并返回实时执行事件的流式读取器。
// 调用者需要立即开始从返回的流中读取数据，以避免阻塞。
//
// 执行流程：
// 1. 验证执行配置和权限
// 2. 获取工作流实体和画布
// 3. 转换为WorkflowSchema
// 4. 处理聊天历史（如果是聊天流）
// 5. 创建带有流式写入器的执行引擎
// 6. 启动异步执行并返回流式读取器
//
// 参数：
//   - ctx: 上下文，用于取消和超时控制
//   - config: 执行配置，包含工作流ID、版本、模式等
//   - input: 工作流输入参数
//
// 返回：
//   - *schema.StreamReader[*entity.Message]: 流式读取器，用于读取执行事件
//   - error: 启动执行过程中的错误
//
// 注意：
//   - 调用者必须立即开始读取流式数据
//   - 支持实时反馈和渐进式输出
//   - 适用于需要实时展示执行过程的场景
func (i *impl) StreamExecute(ctx context.Context, config workflowModel.ExecuteConfig, input map[string]any) (*schema.StreamReader[*entity.Message], error) {
	var (
		err      error
		wfEntity *entity.Workflow
		ws       *nodes.ConversionWarnings
	)

	wfEntity, err = i.Get(ctx, &vo.GetPolicy{
		ID:       config.ID,
		QType:    config.From,
		MetaOnly: false,
		Version:  config.Version,
		CommitID: config.CommitID,
	})
	if err != nil {
		return nil, err
	}

	config.WorkflowMode = wfEntity.Mode

	isApplicationWorkflow := wfEntity.AppID != nil
	if isApplicationWorkflow && config.Mode == workflowModel.ExecuteModeRelease {
		err = i.checkApplicationWorkflowReleaseVersion(ctx, *wfEntity.AppID, config.ConnectorID, config.ID, config.Version)
		if err != nil {
			return nil, err
		}
	}

	c := &vo.Canvas{}
	if err = sonic.UnmarshalString(wfEntity.Canvas, c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal canvas: %w", err)
	}

	workflowSC, err := adaptor.CanvasToWorkflowSchema(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("failed to convert canvas to workflow schema: %w", err)
	}

	historyRounds := int64(0)
	if config.WorkflowMode == workflowapimodel.WorkflowMode_ChatFlow {
		historyRounds = workflowSC.HistoryRounds()
	}

	if historyRounds > 0 {
		if err = i.handleHistory(ctx, &config, input, historyRounds, false); err != nil {
			return nil, err
		}
	}

	config.InputFileFields = slices.ToMap(workflowSC.GetAllNodesInputFileFields(ctx), func(e *workflowModel.FileInfo) (string, *workflowModel.FileInfo) {
		return e.FileURL, e
	})

	var wfOpts []compose.WorkflowOption

	wfOpts = append(wfOpts, compose.WithIDAsName(wfEntity.ID))
	if s := execute.GetStaticConfig(); s != nil && s.MaxNodeCountPerWorkflow > 0 {
		wfOpts = append(wfOpts, compose.WithMaxNodeCount(s.MaxNodeCountPerWorkflow))
	}

	wf, err := compose.NewWorkflow(ctx, workflowSC, wfOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	if wfEntity.AppID != nil && config.AppID == nil {
		config.AppID = wfEntity.AppID
	}

	config.CommitID = wfEntity.CommitID

	var cOpts []nodes.ConvertOption
	inputFileFields := make(map[string]*workflowModel.FileInfo)
	cOpts = append(cOpts, nodes.WithCollectFileFields(inputFileFields), nodes.WithNotNeedTrimQueryFileName(true))
	if config.InputFailFast {
		cOpts = append(cOpts, nodes.FailFast())
	}

	input, ws, err = nodes.ConvertInputs(ctx, input, wf.Inputs(), cOpts...)
	if err != nil {
		return nil, err
	} else if ws != nil {
		logs.CtxWarnf(ctx, "convert inputs warnings: %v", *ws)
	}
	for k, v := range inputFileFields {
		config.InputFileFields[k] = v
	}

	inStr, err := sonic.MarshalString(input)
	if err != nil {
		return nil, err
	}

	sr, sw := schema.Pipe[*entity.Message](10)

	cancelCtx, executeID, opts, _, err := compose.NewWorkflowRunner(wfEntity.GetBasic(), workflowSC, config,
		compose.WithInput(inStr), compose.WithStreamWriter(sw)).Prepare(ctx)
	if err != nil {
		return nil, err
	}

	_ = executeID

	wf.AsyncRun(cancelCtx, input, opts...)

	return sr, nil
}

// GetExecution 获取工作流执行结果
//
// 根据执行ID查询工作流的执行状态和结果。
// 支持获取完整的执行信息，包括节点执行详情。
//
// 查询逻辑：
// 1. 从数据库获取工作流执行记录
// 2. 如果执行仍在运行，返回基本状态信息
// 3. 如果执行已完成，获取中断事件信息
// 4. 根据includeNodes参数决定是否获取节点执行详情
// 5. 合并复合节点的执行结果
//
// 参数：
//   - ctx: 上下文
//   - wfExe: 包含执行ID的工作流执行对象
//   - includeNodes: 是否包含节点执行详情
//
// 返回：
//   - *entity.WorkflowExecution: 完整的执行结果
//   - error: 查询过程中的错误
//
// 注意：
//   - 如果执行仍在运行，只返回基本状态信息
//   - 中断事件会根据当前状态进行过滤
//   - 复合节点（如批量节点）的执行结果会被合并
func (i *impl) GetExecution(ctx context.Context, wfExe *entity.WorkflowExecution, includeNodes bool) (*entity.WorkflowExecution, error) {
	wfExeID := wfExe.ID
	wfID := wfExe.WorkflowID
	version := wfExe.Version
	rootExeID := wfExe.RootExecutionID

	wfExeEntity, found, err := i.repo.GetWorkflowExecution(ctx, wfExeID)
	if err != nil {
		return nil, err
	}

	if !found {
		return &entity.WorkflowExecution{
			ID:              wfExeID,
			WorkflowID:      wfID,
			Version:         version,
			RootExecutionID: rootExeID,
			Status:          entity.WorkflowRunning,
		}, nil
	}

	interruptEvent, found, err := i.repo.GetFirstInterruptEvent(ctx, wfExeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find interrupt events: %v", err)
	}

	if found {
		// if we are currently interrupted, return this interrupt event,
		// otherwise only return this event if it's the current resuming event
		if wfExeEntity.Status == entity.WorkflowInterrupted ||
			(wfExeEntity.CurrentResumingEventID != nil && *wfExeEntity.CurrentResumingEventID == interruptEvent.ID) {
			wfExeEntity.InterruptEvents = []*entity.InterruptEvent{interruptEvent}
		}
	}

	if !includeNodes {
		return wfExeEntity, nil
	}

	// query the node executions for the root execution
	nodeExecs, err := i.repo.GetNodeExecutionsByWfExeID(ctx, wfExeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find node executions: %v", err)
	}

	nodeGroups := make(map[string]map[int]*entity.NodeExecution)
	nodeGroupMaxIndex := make(map[string]int)
	var nodeIDSet map[string]struct{}
	for i := range nodeExecs {
		nodeExec := nodeExecs[i]
		if nodeExec.ParentNodeID != nil {
			if nodeIDSet == nil {
				nodeIDSet = slices.ToMap(nodeExecs, func(e *entity.NodeExecution) (string, struct{}) {
					return e.NodeID, struct{}{}
				})
			}

			if _, ok := nodeIDSet[*nodeExec.ParentNodeID]; ok {
				if _, ok := nodeGroups[nodeExec.NodeID]; !ok {
					nodeGroups[nodeExec.NodeID] = make(map[int]*entity.NodeExecution)
				}
				nodeGroups[nodeExec.NodeID][nodeExec.Index] = nodeExecs[i]
				if nodeExec.Index > nodeGroupMaxIndex[nodeExec.NodeID] {
					nodeGroupMaxIndex[nodeExec.NodeID] = nodeExec.Index
				}

				continue
			}
		}

		wfExeEntity.NodeExecutions = append(wfExeEntity.NodeExecutions, nodeExec)
	}

	for nodeID, nodeExes := range nodeGroups {
		groupNodeExe := mergeCompositeInnerNodes(nodeExes, nodeGroupMaxIndex[nodeID])
		wfExeEntity.NodeExecutions = append(wfExeEntity.NodeExecutions, groupNodeExe)
	}

	return wfExeEntity, nil
}

func (i *impl) GetNodeExecution(ctx context.Context, exeID int64, nodeID string) (*entity.NodeExecution, *entity.NodeExecution, error) {
	nodeExe, found, err := i.repo.GetNodeExecution(ctx, exeID, nodeID)
	if err != nil {
		return nil, nil, err
	}

	if !found {
		return nil, nil, fmt.Errorf("try getting node exe for exeID : %d, nodeID : %s, but not found", exeID, nodeID)
	}

	if nodeExe.NodeType != entity.NodeTypeBatch {
		return nodeExe, nil, nil
	}

	wfExe, found, err := i.repo.GetWorkflowExecution(ctx, exeID)
	if err != nil {
		return nil, nil, err
	}

	if !found {
		return nil, nil, fmt.Errorf("try getting workflow exe for exeID : %d, but not found", exeID)
	}

	if wfExe.Mode != workflowModel.ExecuteModeNodeDebug {
		return nodeExe, nil, nil
	}

	// when node debugging a node with batch mode, we need to query the inner node executions and return it together
	innerNodeExecs, err := i.repo.GetNodeExecutionByParent(ctx, exeID, nodeExe.NodeID)
	if err != nil {
		return nil, nil, err
	}

	for i := range innerNodeExecs {
		innerNodeID := innerNodeExecs[i].NodeID
		if !vo.IsGeneratedNodeForBatchMode(innerNodeID, nodeExe.NodeID) {
			// inner node is not generated, means this is normal batch, not node in batch mode
			return nodeExe, nil, nil
		}
	}

	var (
		maxIndex  int
		index2Exe = make(map[int]*entity.NodeExecution)
	)

	for i := range innerNodeExecs {
		index2Exe[innerNodeExecs[i].Index] = innerNodeExecs[i]
		if innerNodeExecs[i].Index > maxIndex {
			maxIndex = innerNodeExecs[i].Index
		}
	}

	return nodeExe, mergeCompositeInnerNodes(index2Exe, maxIndex), nil
}

func (i *impl) GetLatestTestRunInput(ctx context.Context, wfID int64, userID int64) (*entity.NodeExecution, bool, error) {
	exeID, err := i.repo.GetTestRunLatestExeID(ctx, wfID, userID)
	if err != nil {
		logs.CtxErrorf(ctx, "[GetLatestTestRunInput] failed to get node execution from redis, wfID: %d, err: %v", wfID, err)
		return nil, false, nil
	}

	if exeID == 0 {
		return nil, false, nil
	}

	nodeExe, _, err := i.GetNodeExecution(ctx, exeID, entity.EntryNodeKey)
	if err != nil {
		logs.CtxErrorf(ctx, "[GetLatestTestRunInput] failed to get node execution, exeID: %d, err: %v", exeID, err)
		return nil, false, nil
	}

	return nodeExe, true, nil
}

func (i *impl) GetLatestNodeDebugInput(ctx context.Context, wfID int64, nodeID string, userID int64) (
	*entity.NodeExecution, *entity.NodeExecution, bool, error) {
	exeID, err := i.repo.GetNodeDebugLatestExeID(ctx, wfID, nodeID, userID)
	if err != nil {
		logs.CtxErrorf(ctx, "[GetLatestNodeDebugInput] failed to get node execution from redis, wfID: %d, nodeID: %s, err: %v",
			wfID, nodeID, err)
		return nil, nil, false, nil
	}

	if exeID == 0 {
		return nil, nil, false, nil
	}

	nodeExe, innerExe, err := i.GetNodeExecution(ctx, exeID, nodeID)
	if err != nil {
		logs.CtxErrorf(ctx, "[GetLatestNodeDebugInput] failed to get node execution, exeID: %d, nodeID: %s, err: %v",
			exeID, nodeID, err)
		return nil, nil, false, nil
	}

	return nodeExe, innerExe, true, nil
}

// mergeCompositeInnerNodes 合并复合节点的内部执行结果
//
// 将批量节点或其他复合节点的多个内部执行结果合并为单个节点执行记录。
// 用于简化前端展示和统计分析，避免显示过多的内部节点详情。
//
// 合并策略：
// 1. 使用第一个执行记录作为基础模板
// 2. 累加所有执行的Token消耗
// 3. 取最长执行时间作为总时间
// 4. 如果任一执行失败，整个节点标记为失败
// 5. 创建索引化执行数组，保持执行顺序
//
// 参数：
//   - nodeExes: 按索引组织的节点执行结果map
//   - maxIndex: 最大索引值，用于确定数组大小
//
// 返回：
//   - *entity.NodeExecution: 合并后的节点执行记录
//
// 注意：
//   - 合并后的记录保持了所有关键信息
//   - IndexedExecutions数组按索引顺序排列
//   - 状态和Token信息经过适当聚合
func mergeCompositeInnerNodes(nodeExes map[int]*entity.NodeExecution, maxIndex int) *entity.NodeExecution {
	var groupNodeExe *entity.NodeExecution
	for _, v := range nodeExes {
		groupNodeExe = &entity.NodeExecution{
			ID:           v.ID,
			ExecuteID:    v.ExecuteID,
			NodeID:       v.NodeID,
			NodeName:     v.NodeName,
			NodeType:     v.NodeType,
			ParentNodeID: v.ParentNodeID,
		}
		break
	}

	var (
		duration  time.Duration
		tokenInfo *entity.TokenUsage
		status    = entity.NodeSuccess
	)

	groupNodeExe.IndexedExecutions = make([]*entity.NodeExecution, maxIndex+1)

	for index, ne := range nodeExes {
		duration = max(duration, ne.Duration)
		if ne.TokenInfo != nil {
			if tokenInfo == nil {
				tokenInfo = &entity.TokenUsage{}
			}
			tokenInfo.InputTokens += ne.TokenInfo.InputTokens
			tokenInfo.OutputTokens += ne.TokenInfo.OutputTokens
		}
		if ne.Status == entity.NodeFailed {
			status = entity.NodeFailed
		} else if ne.Status == entity.NodeRunning {
			status = entity.NodeRunning
		}

		groupNodeExe.IndexedExecutions[index] = nodeExes[index]
	}

	groupNodeExe.Duration = duration
	groupNodeExe.TokenInfo = tokenInfo
	groupNodeExe.Status = status

	return groupNodeExe
}

// AsyncResume 异步恢复工作流执行
//
// 从中断状态恢复工作流执行，使用传入的执行ID和事件ID。
// 恢复执行的中间结果不会实时发出，调用者需要轮询GetExecution方法获取状态。
//
// 恢复流程：
// 1. 验证执行状态（必须是中断状态）
// 2. 验证只能恢复根执行
// 3. 获取原始工作流实体和画布
// 4. 根据执行模式选择恢复策略（完整恢复或节点调试恢复）
// 5. 创建恢复执行引擎并启动异步执行
//
// 参数：
//   - ctx: 上下文，用于取消和超时控制
//   - req: 恢复请求，包含执行ID和恢复事件信息
//   - config: 执行配置，包含恢复相关的参数
//
// 返回：
//   - error: 恢复执行过程中的错误
//
// 注意：
//   - 只能恢复处于中断状态的执行
//   - 只能恢复根执行，不能恢复子执行
//   - 支持节点调试模式下的恢复
func (i *impl) AsyncResume(ctx context.Context, req *entity.ResumeRequest, config workflowModel.ExecuteConfig) error {
	wfExe, found, err := i.repo.GetWorkflowExecution(ctx, req.ExecuteID)
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("workflow execution does not exist, id: %d", req.ExecuteID)
	}

	if wfExe.RootExecutionID != wfExe.ID {
		return fmt.Errorf("only root workflow can be resumed")
	}

	if wfExe.Status != entity.WorkflowInterrupted {
		return fmt.Errorf("workflow execution %d is not interrupted, status is %v, cannot resume", req.ExecuteID, wfExe.Status)
	}

	var from workflowModel.Locator
	if wfExe.Version == "" {
		from = workflowModel.FromDraft
	} else {
		from = workflowModel.FromSpecificVersion
	}

	wfEntity, err := i.Get(ctx, &vo.GetPolicy{
		ID:       wfExe.WorkflowID,
		QType:    from,
		Version:  wfExe.Version,
		CommitID: wfExe.CommitID,
	})
	if err != nil {
		return err
	}

	var canvas vo.Canvas
	err = sonic.UnmarshalString(wfEntity.Canvas, &canvas)
	if err != nil {
		return err
	}

	config.From = from
	config.Version = wfExe.Version
	config.AppID = wfExe.AppID
	config.AgentID = wfExe.AgentID
	config.CommitID = wfExe.CommitID
	config.WorkflowMode = wfEntity.Mode

	if config.ConnectorID == 0 {
		config.ConnectorID = wfExe.ConnectorID
	}

	if wfExe.Mode == workflowModel.ExecuteModeNodeDebug {
		nodeExes, err := i.repo.GetNodeExecutionsByWfExeID(ctx, wfExe.ID)
		if err != nil {
			return err
		}

		if len(nodeExes) == 0 {
			return fmt.Errorf("during node debug resume, no node execution found for workflow execution %d", wfExe.ID)
		}

		var nodeID string
		for _, ne := range nodeExes {
			if ne.ParentNodeID == nil {
				nodeID = ne.NodeID
				break
			}
		}

		workflowSC, err := adaptor.WorkflowSchemaFromNode(ctx, &canvas, nodeID)
		if err != nil {
			return fmt.Errorf("failed to convert canvas to workflow schema: %w", err)
		}

		wf, err := compose.NewWorkflowFromNode(ctx, workflowSC, vo.NodeKey(nodeID),
			einoCompose.WithGraphName(fmt.Sprintf("%d", wfExe.WorkflowID)))
		if err != nil {
			return fmt.Errorf("failed to create workflow: %w", err)
		}

		config.Mode = workflowModel.ExecuteModeNodeDebug

		cancelCtx, _, opts, _, err := compose.NewWorkflowRunner(
			wfEntity.GetBasic(), workflowSC, config, compose.WithResumeReq(req)).Prepare(ctx)
		if err != nil {
			return err
		}

		wf.AsyncRun(cancelCtx, nil, opts...)
		return nil
	}

	workflowSC, err := adaptor.CanvasToWorkflowSchema(ctx, &canvas)
	if err != nil {
		return fmt.Errorf("failed to convert canvas to workflow schema: %w", err)
	}

	var wfOpts []compose.WorkflowOption
	wfOpts = append(wfOpts, compose.WithIDAsName(wfExe.WorkflowID))
	if s := execute.GetStaticConfig(); s != nil && s.MaxNodeCountPerWorkflow > 0 {
		wfOpts = append(wfOpts, compose.WithMaxNodeCount(s.MaxNodeCountPerWorkflow))
	}

	wf, err := compose.NewWorkflow(ctx, workflowSC, wfOpts...)
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	cancelCtx, _, opts, _, err := compose.NewWorkflowRunner(
		wfEntity.GetBasic(), workflowSC, config, compose.WithResumeReq(req)).Prepare(ctx)
	if err != nil {
		return err
	}

	wf.AsyncRun(cancelCtx, nil, opts...)

	return nil
}

// StreamResume 流式恢复工作流执行
//
// 从中断状态恢复工作流执行，使用传入的执行ID和事件ID。
// 恢复执行的中间结果会通过返回的StreamReader实时发出。
// 调用者需要轮询GetExecution方法获取最终执行状态。
//
// 恢复流程：
// 1. 验证执行状态（必须是中断状态）
// 2. 验证只能恢复根执行
// 3. 获取原始工作流实体和画布
// 4. 创建带有流式写入器的恢复执行引擎
// 5. 启动异步恢复执行并返回流式读取器
//
// 参数：
//   - ctx: 上下文，用于取消和超时控制
//   - req: 恢复请求，包含执行ID和恢复事件信息
//   - config: 执行配置，包含恢复相关的参数
//
// 返回：
//   - *schema.StreamReader[*entity.Message]: 流式读取器，用于读取恢复执行事件
//   - error: 恢复执行过程中的错误
//
// 注意：
//   - 只能恢复处于中断状态的执行
//   - 只能恢复根执行，不能恢复子执行
//   - 恢复过程的中间结果会实时流式输出
func (i *impl) StreamResume(ctx context.Context, req *entity.ResumeRequest, config workflowModel.ExecuteConfig) (
	*schema.StreamReader[*entity.Message], error) {
	// 必须获取中断事件
	// 生成状态修改器
	wfExe, found, err := i.repo.GetWorkflowExecution(ctx, req.ExecuteID)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("workflow execution does not exist, id: %d", req.ExecuteID)
	}

	if wfExe.RootExecutionID != wfExe.ID {
		return nil, fmt.Errorf("only root workflow can be resumed")
	}

	if wfExe.Status != entity.WorkflowInterrupted {
		return nil, fmt.Errorf("workflow execution %d is not interrupted, status is %v, cannot resume", req.ExecuteID, wfExe.Status)
	}

	var from workflowModel.Locator
	if wfExe.Version == "" {
		from = workflowModel.FromDraft
	} else {
		from = workflowModel.FromSpecificVersion
	}

	wfEntity, err := i.Get(ctx, &vo.GetPolicy{
		ID:       wfExe.WorkflowID,
		QType:    from,
		Version:  wfExe.Version,
		CommitID: wfExe.CommitID,
	})
	if err != nil {
		return nil, err
	}

	var canvas vo.Canvas
	err = sonic.UnmarshalString(wfEntity.Canvas, &canvas)
	if err != nil {
		return nil, err
	}

	workflowSC, err := adaptor.CanvasToWorkflowSchema(ctx, &canvas)
	if err != nil {
		return nil, fmt.Errorf("failed to convert canvas to workflow schema: %w", err)
	}

	var wfOpts []compose.WorkflowOption
	wfOpts = append(wfOpts, compose.WithIDAsName(wfExe.WorkflowID))
	if s := execute.GetStaticConfig(); s != nil && s.MaxNodeCountPerWorkflow > 0 {
		wfOpts = append(wfOpts, compose.WithMaxNodeCount(s.MaxNodeCountPerWorkflow))
	}

	wf, err := compose.NewWorkflow(ctx, workflowSC, wfOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	config.From = from
	config.Version = wfExe.Version
	config.AppID = wfExe.AppID
	config.AgentID = wfExe.AgentID
	config.CommitID = wfExe.CommitID
	config.WorkflowMode = wfEntity.Mode

	if config.ConnectorID == 0 {
		config.ConnectorID = wfExe.ConnectorID
	}

	sr, sw := schema.Pipe[*entity.Message](10)

	cancelCtx, _, opts, _, err := compose.NewWorkflowRunner(wfEntity.GetBasic(), workflowSC, config,
		compose.WithResumeReq(req), compose.WithStreamWriter(sw)).Prepare(ctx)
	if err != nil {
		return nil, err
	}

	wf.AsyncRun(cancelCtx, nil, opts...)

	return sr, nil
}

// Cancel 取消工作流执行
//
// 取消正在运行的工作流执行，将状态设置为取消。
// 只有正在运行或已中断的执行才能被取消。
//
// 取消逻辑：
// 1. 验证执行存在性和权限
// 2. 检查执行状态（只能取消运行中或中断的执行）
// 3. 验证只能取消根执行
// 4. 更新执行状态为取消
// 5. 取消所有正在运行的节点
//
// 参数：
//   - ctx: 上下文
//   - wfExeID: 要取消的执行ID
//   - wfID: 工作流ID，用于权限验证
//   - spaceID: 空间ID，用于权限验证
//
// 返回：
//   - error: 取消过程中的错误
//
// 注意：
//   - 已达到终止状态的执行不需要取消
//   - 从中断状态取消时需要同时取消所有中断节点
//   - 会发送取消信号确保执行引擎响应
func (i *impl) Cancel(ctx context.Context, wfExeID int64, wfID, spaceID int64) error {
	wfExe, found, err := i.repo.GetWorkflowExecution(ctx, wfExeID)
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("workflow execution does not exist, wfExeID: %d", wfExeID)
	}

	if wfExe.WorkflowID != wfID || wfExe.SpaceID != spaceID {
		return fmt.Errorf("workflow execution id mismatch, wfExeID: %d, wfID: %d, spaceID: %d", wfExeID, wfID, spaceID)
	}

	if wfExe.Status != entity.WorkflowRunning && wfExe.Status != entity.WorkflowInterrupted {
		// already reached terminal state, no need to cancel
		return nil
	}

	if wfExe.ID != wfExe.RootExecutionID {
		return fmt.Errorf("can only cancel root execute ID")
	}

	wfExec := &entity.WorkflowExecution{
		ID:     wfExe.ID,
		Status: entity.WorkflowCancel,
	}

	var (
		updatedRows   int64
		currentStatus entity.WorkflowExecuteStatus
	)
	if updatedRows, currentStatus, err = i.repo.UpdateWorkflowExecution(ctx, wfExec, []entity.WorkflowExecuteStatus{entity.WorkflowInterrupted}); err != nil {
		return fmt.Errorf("failed to save workflow execution to canceled while interrupted: %v", err)
	} else if updatedRows == 0 {
		if currentStatus != entity.WorkflowRunning {
			// already terminal state, try cancel all nodes just in case
			return i.repo.CancelAllRunningNodes(ctx, wfExe.ID)
		} else {
			// current running, let the execution time event handle do the actual updating status to cancel
		}
	} else if err = i.repo.CancelAllRunningNodes(ctx, wfExe.ID); err != nil { // we updated the workflow from interrupted to cancel, so we need to cancel all interrupting nodes
		return fmt.Errorf("failed to update all running nodes to cancel: %v", err)
	}

	// emit cancel signal just in case the execution is running
	return i.repo.SetWorkflowCancelFlag(ctx, wfExeID)
}

func (i *impl) checkApplicationWorkflowReleaseVersion(ctx context.Context, appID, connectorID, workflowID int64, version string) error {
	ok, err := i.repo.IsApplicationConnectorWorkflowVersion(ctx, connectorID, workflowID, version)
	if err != nil {
		return err
	}
	if !ok {
		return vo.WrapError(errno.ErrWorkflowSpecifiedVersionNotFound, fmt.Errorf("applcaition id %v, workflow id %v,connector id %v not have version %v", appID, workflowID, connectorID, version))
	}

	return nil
}

// prefetchChatHistory 预取聊天历史
//
// 从消息服务获取指定轮数的聊天历史记录，用于LLM上下文记忆。
// 支持应用和代理两种业务场景的消息历史获取。
//
// 获取流程：
// 1. 验证必要的对话和分段信息
// 2. 调用消息服务获取最近的运行ID列表
// 3. 过滤掉当前运行，获取历史运行ID
// 4. 批量获取历史消息内容
//
// 参数：
//   - ctx: 上下文
//   - config: 执行配置，包含对话ID、业务ID等
//   - historyRounds: 需要的历史对话轮数
//
// 返回：
//   - []*crossmessage.WfMessage: 工作流消息格式的历史消息
//   - []*schema.Message: Schema格式的历史消息
//   - error: 获取过程中的错误
//
// 注意：
//   - 如果SectionID为空，会跳过历史获取
//   - 会自动过滤无效的对话信息
//   - 返回的消息按时间倒序排列
func (i *impl) prefetchChatHistory(ctx context.Context, config workflowModel.ExecuteConfig, historyRounds int64) ([]*crossmessage.WfMessage, []*schema.Message, error) {
	convID := config.ConversationID
	agentID := config.AgentID
	appID := config.AppID
	userID := config.Operator
	sectionID := config.SectionID
	if sectionID == nil {
		logs.CtxWarnf(ctx, "SectionID is nil, skipping chat history")
		return nil, nil, nil
	}

	if convID == nil || *convID == 0 {
		logs.CtxWarnf(ctx, "ConversationID is 0 or nil, skipping chat history")
		return nil, nil, nil
	}

	var bizID int64
	if appID != nil {
		bizID = *appID
	} else if agentID != nil {
		bizID = *agentID
	} else {
		logs.CtxWarnf(ctx, "AppID and AgentID are both nil, skipping chat history")
		return nil, nil, nil
	}

	runIdsReq := &crossmessage.GetLatestRunIDsRequest{
		ConversationID: *convID,
		BizID:          bizID,
		UserID:         userID,
		Rounds:         historyRounds + 1,
		SectionID:      *sectionID,
	}

	runIds, err := crossmessage.DefaultSVC().GetLatestRunIDs(ctx, runIdsReq)
	if err != nil {
		logs.CtxErrorf(ctx, "failed to get latest run ids: %v", err)
		return nil, nil, err
	}
	if len(runIds) <= 1 {
		return []*crossmessage.WfMessage{}, []*schema.Message{}, nil
	}
	runIds = runIds[1:]

	response, err := crossmessage.DefaultSVC().GetMessagesByRunIDs(ctx, &crossmessage.GetMessagesByRunIDsRequest{
		ConversationID: *convID,
		RunIDs:         runIds,
	})
	if err != nil {
		logs.CtxErrorf(ctx, "failed to get messages by run ids: %v", err)
		return nil, nil, err
	}

	return response.Messages, response.SchemaMessages, nil
}
