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

package repo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	workflowModel "github.com/coze-dev/coze-studio/backend/crossdomain/workflow/model"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/repo/dal/model"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/repo/dal/query"
	"github.com/coze-dev/coze-studio/backend/infra/cache"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/slices"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ternary"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
	"github.com/coze-dev/coze-studio/backend/pkg/sonic"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

// executeHistoryStoreImpl 执行历史存储的实现
//
// 该结构体负责工作流和节点执行记录的持久化存储。
// 主要功能包括：
//   - 工作流执行记录的 CRUD 操作
//   - 节点执行记录的 CRUD 操作
//   - 测试运行和节点调试的最新执行 ID 缓存
//   - 流式输出的临时存储（Redis）
type executeHistoryStoreImpl struct {
	query *query.Query  // GORM gen 查询对象
	redis cache.Cmdable // Redis 客户端，用于缓存和流式输出存储
}

// CreateWorkflowExecution 创建工作流执行记录
//
// 创建新的工作流执行记录。如果是子工作流执行，还会更新父节点的子执行 ID。
//
// 参数：
//   - ctx: 上下文
//   - execution: 工作流执行实体
//
// 返回值：
//   - error: 创建失败时返回错误
func (e *executeHistoryStoreImpl) CreateWorkflowExecution(ctx context.Context, execution *entity.WorkflowExecution) (err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	var mode int32
	if execution.Mode == workflowModel.ExecuteModeDebug {
		mode = 1
	} else if execution.Mode == workflowModel.ExecuteModeRelease {
		mode = 2
	} else if execution.Mode == workflowModel.ExecuteModeNodeDebug {
		mode = 3
	}

	var syncPattern int32
	switch execution.SyncPattern {
	case workflowModel.SyncPatternSync:
		syncPattern = 1
	case workflowModel.SyncPatternAsync:
		syncPattern = 2
	case workflowModel.SyncPatternStream:
		syncPattern = 3
	default:
	}

	wfExec := &model.WorkflowExecution{
		ID:              execution.ID,
		WorkflowID:      execution.WorkflowID,
		Version:         execution.Version,
		SpaceID:         execution.SpaceID,
		Mode:            mode,
		OperatorID:      execution.Operator,
		Status:          int32(entity.WorkflowRunning),
		Input:           ptr.FromOrDefault(execution.Input, ""),
		RootExecutionID: execution.RootExecutionID,
		ParentNodeID:    ptr.FromOrDefault(execution.ParentNodeID, ""),
		AppID:           ptr.FromOrDefault(execution.AppID, 0),
		AgentID:         ptr.FromOrDefault(execution.AgentID, 0),
		ConnectorID:     execution.ConnectorID,
		ConnectorUID:    execution.ConnectorUID,
		NodeCount:       execution.NodeCount,
		SyncPattern:     syncPattern,
		CommitID:        execution.CommitID,
		LogID:           execution.LogID,
	}

	if execution.ParentNodeID == nil {
		return e.query.WorkflowExecution.WithContext(ctx).Create(wfExec)
	}

	return e.query.Transaction(func(tx *query.Query) error {
		if err := e.query.WorkflowExecution.WithContext(ctx).Create(wfExec); err != nil {
			return err
		}

		// update the parent node execution's sub execute id
		if _, err := e.query.NodeExecution.WithContext(ctx).Where(e.query.NodeExecution.ID.Eq(*execution.ParentNodeExecuteID)).
			UpdateColumn(e.query.NodeExecution.SubExecuteID, wfExec.ID); err != nil {
			return err
		}

		return nil
	})
}

// UpdateWorkflowExecution 更新工作流执行记录
//
// 只有当前状态在允许的状态列表中时才会更新。
// 使用乐观锁机制避免并发更新问题。
//
// 参数：
//   - ctx: 上下文
//   - execution: 工作流执行实体（包含要更新的字段）
//   - allowedStatus: 允许的当前状态列表
//
// 返回值：
//   - int64: 影响的行数
//   - entity.WorkflowExecuteStatus: 当前状态（如果更新失败）
//   - error: 更新失败时返回错误
func (e *executeHistoryStoreImpl) UpdateWorkflowExecution(ctx context.Context, execution *entity.WorkflowExecution,
	allowedStatus []entity.WorkflowExecuteStatus) (_ int64, _ entity.WorkflowExecuteStatus, err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	// Use map[string]any to explicitly specify fields for update
	updateMap := map[string]any{
		"status":          int32(execution.Status),
		"output":          ptr.FromOrDefault(execution.Output, ""),
		"duration":        execution.Duration.Milliseconds(),
		"error_code":      ptr.FromOrDefault(execution.ErrorCode, ""),
		"fail_reason":     ptr.FromOrDefault(execution.FailReason, ""),
		"resume_event_id": ptr.FromOrDefault(execution.CurrentResumingEventID, 0),
	}

	if execution.TokenInfo != nil {
		updateMap["input_tokens"] = execution.TokenInfo.InputTokens
		updateMap["output_tokens"] = execution.TokenInfo.OutputTokens
	}

	statuses := slices.Transform(allowedStatus, func(e entity.WorkflowExecuteStatus) int32 {
		return int32(e)
	})

	info, err := e.query.WorkflowExecution.WithContext(ctx).Where(e.query.WorkflowExecution.ID.Eq(execution.ID),
		e.query.WorkflowExecution.Status.In(statuses...)).Updates(updateMap)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to update workflow execution: %w", err)
	}

	if info.RowsAffected == 0 {
		wfExe, found, err := e.GetWorkflowExecution(ctx, execution.ID)
		if err != nil {
			return 0, 0, err
		}

		if !found {
			return 0, 0, fmt.Errorf("workflow execution not found for ID %d", execution.ID)
		}

		return 0, wfExe.Status, nil
	}

	return info.RowsAffected, execution.Status, nil
}

// TryLockWorkflowExecution 尝试锁定工作流执行以进行恢复操作
//
// 使用数据库行级锁实现分布式锁。只有当工作流处于中断状态且没有其他恢复操作时才能锁定。
//
// 参数：
//   - ctx: 上下文
//   - wfExeID: 工作流执行 ID
//   - resumingEventID: 恢复事件 ID
//
// 返回值：
//   - bool: 是否成功获取锁
//   - entity.WorkflowExecuteStatus: 当前状态（如果获取锁失败）
//   - error: 操作失败时返回错误
func (e *executeHistoryStoreImpl) TryLockWorkflowExecution(ctx context.Context, wfExeID, resumingEventID int64) (
	_ bool, _ entity.WorkflowExecuteStatus, err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	// Update WorkflowExecution set current_resuming_event_id = resumingEventID, status = 1
	// where id = wfExeID and current_resuming_event_id = 0 and status = 5
	result, err := e.query.WorkflowExecution.WithContext(ctx).
		Where(e.query.WorkflowExecution.ID.Eq(wfExeID)).
		Where(e.query.WorkflowExecution.ResumeEventID.Eq(0)).
		Where(e.query.WorkflowExecution.Status.Eq(int32(entity.WorkflowInterrupted))).
		Updates(map[string]interface{}{
			"resume_event_id": resumingEventID,
			"status":          int32(entity.WorkflowRunning),
		})

	if err != nil {
		return false, 0, fmt.Errorf("update workflow execution lock failed: %w", err)
	}

	// If no rows were updated, the lock attempt failed
	if result.RowsAffected == 0 {
		wfExe, found, err := e.GetWorkflowExecution(ctx, wfExeID)
		if err != nil {
			return false, 0, err
		}
		if !found {
			return false, 0, fmt.Errorf("workflow execution not found for ID %d", wfExeID)
		}

		return false, wfExe.Status, nil
	}

	return true, entity.WorkflowInterrupted, nil
}

// GetWorkflowExecution 获取工作流执行记录
//
// 参数：
//   - ctx: 上下文
//   - id: 工作流执行 ID
//
// 返回值：
//   - *entity.WorkflowExecution: 工作流执行实体
//   - bool: 是否存在
//   - error: 获取失败时返回错误
func (e *executeHistoryStoreImpl) GetWorkflowExecution(ctx context.Context, id int64) (*entity.WorkflowExecution, bool, error) {
	rootExes, err := e.query.WorkflowExecution.WithContext(ctx).
		Where(e.query.WorkflowExecution.ID.Eq(id)).
		Find()
	if err != nil {
		return nil, false, vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("failed to find workflow execution: %v", err))
	}

	if len(rootExes) == 0 {
		return nil, false, nil
	}

	rootExe := rootExes[0]
	var exeMode workflowModel.ExecuteMode
	if rootExe.Mode == 1 {
		exeMode = workflowModel.ExecuteModeDebug
	} else if rootExe.Mode == 2 {
		exeMode = workflowModel.ExecuteModeRelease
	} else {
		exeMode = workflowModel.ExecuteModeNodeDebug
	}

	var syncPattern workflowModel.SyncPattern
	switch rootExe.SyncPattern {
	case 1:
		syncPattern = workflowModel.SyncPatternSync
	case 2:
		syncPattern = workflowModel.SyncPatternAsync
	case 3:
		syncPattern = workflowModel.SyncPatternStream
	default:
	}

	exe := &entity.WorkflowExecution{
		ID:         rootExe.ID,
		WorkflowID: rootExe.WorkflowID,
		Version:    rootExe.Version,
		SpaceID:    rootExe.SpaceID,
		ExecuteConfig: workflowModel.ExecuteConfig{
			Operator:     rootExe.OperatorID,
			Mode:         exeMode,
			AppID:        ternary.IFElse(rootExe.AppID > 0, ptr.Of(rootExe.AppID), nil),
			AgentID:      ternary.IFElse(rootExe.AgentID > 0, ptr.Of(rootExe.AgentID), nil),
			ConnectorID:  rootExe.ConnectorID,
			ConnectorUID: rootExe.ConnectorUID,
			SyncPattern:  syncPattern,
		},
		CreatedAt:  time.UnixMilli(rootExe.CreatedAt),
		LogID:      rootExe.LogID,
		NodeCount:  rootExe.NodeCount,
		Status:     entity.WorkflowExecuteStatus(rootExe.Status),
		Duration:   time.Duration(rootExe.Duration) * time.Millisecond,
		Input:      &rootExe.Input,
		Output:     &rootExe.Output,
		ErrorCode:  &rootExe.ErrorCode,
		FailReason: &rootExe.FailReason,
		TokenInfo: &entity.TokenUsage{
			InputTokens:  rootExe.InputTokens,
			OutputTokens: rootExe.OutputTokens,
		},
		UpdatedAt:              ternary.IFElse(rootExe.UpdatedAt > 0, ptr.Of(time.UnixMilli(rootExe.UpdatedAt)), nil),
		ParentNodeID:           ptr.Of(rootExe.ParentNodeID),
		ParentNodeExecuteID:    nil, // keep it nil here, query parent node execution separately
		NodeExecutions:         nil, // keep it nil here, query node executions separately
		RootExecutionID:        rootExe.RootExecutionID,
		CurrentResumingEventID: ternary.IFElse(rootExe.ResumeEventID == 0, nil, ptr.Of(rootExe.ResumeEventID)),
		CommitID:               rootExe.CommitID,
	}

	return exe, true, nil
}

// CreateNodeExecution 创建节点执行记录
//
// 参数：
//   - ctx: 上下文
//   - execution: 节点执行实体
//
// 返回值：
//   - error: 创建失败时返回错误
func (e *executeHistoryStoreImpl) CreateNodeExecution(ctx context.Context, execution *entity.NodeExecution) error {
	nodeExec := &model.NodeExecution{
		ID:                 execution.ID,
		ExecuteID:          execution.ExecuteID,
		NodeID:             execution.NodeID,
		NodeName:           execution.NodeName,
		NodeType:           string(execution.NodeType),
		Status:             int32(entity.NodeRunning),
		Input:              ptr.FromOrDefault(execution.Input, ""),
		CompositeNodeIndex: int64(execution.Index),
		CompositeNodeItems: ptr.FromOrDefault(execution.Items, ""),
		ParentNodeID:       ptr.FromOrDefault(execution.ParentNodeID, ""),
	}

	if execution.Extra != nil {
		m, err := sonic.MarshalString(execution.Extra)
		if err != nil {
			return vo.WrapError(errno.ErrSerializationDeserializationFail,
				fmt.Errorf("failed to marshal extra: %w", err))
		}
		nodeExec.Extra = m
	}

	return e.query.NodeExecution.WithContext(ctx).Create(nodeExec)
}

// UpdateNodeExecutionStreaming 更新节点流式输出
//
// 将节点的流式输出临时存储在 Redis 中，供前端实时获取。
// 数据有效期 24 小时。
//
// 参数：
//   - ctx: 上下文
//   - execution: 节点执行实体（包含输出）
//
// 返回值：
//   - error: 更新失败时返回错误
func (e *executeHistoryStoreImpl) UpdateNodeExecutionStreaming(ctx context.Context, execution *entity.NodeExecution) error {
	if execution.Output == nil {
		return nil
	}

	key := fmt.Sprintf(nodeExecOutputKey, execution.ID)

	if err := e.redis.Set(ctx, key, *execution.Output, nodeExecDataExpiry).Err(); err != nil {
		return vo.WrapError(errno.ErrRedisError, err)
	}

	return nil
}

// UpdateNodeExecution 更新节点执行记录
//
// 参数：
//   - ctx: 上下文
//   - execution: 节点执行实体（包含要更新的字段）
//
// 返回值：
//   - error: 更新失败时返回错误
func (e *executeHistoryStoreImpl) UpdateNodeExecution(ctx context.Context, execution *entity.NodeExecution) (err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	nodeExec := &model.NodeExecution{
		Status:     int32(execution.Status),
		Input:      ptr.FromOrDefault(execution.Input, ""),
		Output:     ptr.FromOrDefault(execution.Output, ""),
		RawOutput:  ptr.FromOrDefault(execution.RawOutput, ""),
		Duration:   execution.Duration.Milliseconds(),
		ErrorInfo:  ptr.FromOrDefault(execution.ErrorInfo, ""),
		ErrorLevel: ptr.FromOrDefault(execution.ErrorLevel, ""),
	}

	if execution.TokenInfo != nil {
		nodeExec.InputTokens = execution.TokenInfo.InputTokens
		nodeExec.OutputTokens = execution.TokenInfo.OutputTokens
	}

	if execution.Extra != nil {
		m, err := sonic.MarshalString(execution.Extra)
		if err != nil {
			return fmt.Errorf("failed to marshal extra: %w", err)
		}
		nodeExec.Extra = m
	}

	if execution.SubWorkflowExecution != nil {
		nodeExec.SubExecuteID = execution.SubWorkflowExecution.ID
	}

	_, err = e.query.NodeExecution.WithContext(ctx).Where(e.query.NodeExecution.ID.Eq(execution.ID)).Updates(nodeExec)
	if err != nil {
		return fmt.Errorf("failed to update node execution: %w", err)
	}

	return nil
}

// CancelAllRunningNodes 取消所有正在运行的节点
//
// 将指定工作流执行下所有运行中的节点标记为失败，同时取消所有子工作流执行。
//
// 参数：
//   - ctx: 上下文
//   - wfExeID: 工作流执行 ID
//
// 返回值：
//   - error: 取消失败时返回错误
func (e *executeHistoryStoreImpl) CancelAllRunningNodes(ctx context.Context, wfExeID int64) (err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	_, err = e.query.NodeExecution.WithContext(ctx).
		Where(e.query.NodeExecution.ExecuteID.Eq(wfExeID),
			e.query.NodeExecution.Status.In(int32(entity.NodeRunning))).
		Updates(map[string]interface{}{
			"error_info":  "workflow cancel by user",
			"error_level": vo.LevelCancel,
			"status":      int32(entity.NodeFailed),
		})
	if err != nil {
		return fmt.Errorf("failed to cancel running nodes: %w", err)
	}

	_, err = e.query.WorkflowExecution.WithContext(ctx).
		Where(e.query.WorkflowExecution.RootExecutionID.Eq(wfExeID)).
		Updates(map[string]interface{}{
			"status":      int32(entity.WorkflowCancel),
			"fail_reason": "workflow cancel by user",
			"error_code":  strconv.Itoa(errno.ErrWorkflowCanceledByUser),
		})
	if err != nil {
		return fmt.Errorf("failed to cancel workflow execution: %w", err)
	}
	return nil
}

// convertNodeExecution 将数据库模型转换为领域实体
//
// 参数：
//   - nodeExec: 数据库节点执行模型
//
// 返回值：
//   - *entity.NodeExecution: 领域层节点执行实体
func convertNodeExecution(nodeExec *model.NodeExecution) *entity.NodeExecution {
	nodeExeEntity := &entity.NodeExecution{
		ID:                   nodeExec.ID,
		ExecuteID:            nodeExec.ExecuteID,
		NodeID:               nodeExec.NodeID,
		NodeName:             nodeExec.NodeName,
		NodeType:             entity.NodeType(nodeExec.NodeType),
		CreatedAt:            time.UnixMilli(nodeExec.CreatedAt),
		Status:               entity.NodeExecuteStatus(nodeExec.Status),
		Duration:             time.Duration(nodeExec.Duration) * time.Millisecond,
		Input:                &nodeExec.Input,
		Output:               &nodeExec.Output,
		RawOutput:            &nodeExec.RawOutput,
		ErrorInfo:            &nodeExec.ErrorInfo,
		ErrorLevel:           &nodeExec.ErrorLevel,
		TokenInfo:            &entity.TokenUsage{InputTokens: nodeExec.InputTokens, OutputTokens: nodeExec.OutputTokens},
		ParentNodeID:         ternary.IFElse(nodeExec.ParentNodeID != "", ptr.Of(nodeExec.ParentNodeID), nil),
		Index:                int(nodeExec.CompositeNodeIndex),
		Items:                ternary.IFElse(nodeExec.CompositeNodeItems != "", ptr.Of(nodeExec.CompositeNodeItems), nil),
		SubWorkflowExecution: ternary.IFElse(nodeExec.SubExecuteID > 0, &entity.WorkflowExecution{ID: nodeExec.SubExecuteID}, nil),
	}

	if nodeExec.UpdatedAt > 0 {
		nodeExeEntity.UpdatedAt = ptr.Of(time.UnixMilli(nodeExec.UpdatedAt))
	}

	if nodeExec.SubExecuteID > 0 {
		nodeExeEntity.SubWorkflowExecution = &entity.WorkflowExecution{
			ID: nodeExec.SubExecuteID,
		}
	}

	if len(nodeExec.Extra) > 0 {
		var extra entity.NodeExtra
		if err := sonic.UnmarshalString(nodeExec.Extra, &extra); err != nil {
			logs.Errorf("failed to unmarshal extra: %v", err)
		} else {
			nodeExeEntity.Extra = &extra
		}
	}

	return nodeExeEntity
}

// GetNodeExecutionsByWfExeID 根据工作流执行 ID 获取所有节点执行记录
//
// 对于支持流式输出且正在运行的节点，会从 Redis 获取最新的输出内容。
//
// 参数：
//   - ctx: 上下文
//   - wfExeID: 工作流执行 ID
//
// 返回值：
//   - []*entity.NodeExecution: 节点执行记录列表
//   - error: 获取失败时返回错误
func (e *executeHistoryStoreImpl) GetNodeExecutionsByWfExeID(ctx context.Context, wfExeID int64) (result []*entity.NodeExecution, err error) {
	nodeExecs, err := e.query.NodeExecution.WithContext(ctx).
		Where(e.query.NodeExecution.ExecuteID.Eq(wfExeID)).
		Find()
	if err != nil {
		return nil, vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("failed to find node executions: %v", err))
	}

	for _, nodeExec := range nodeExecs {
		nodeExeEntity := convertNodeExecution(nodeExec)
		// For nodes that are currently running and support streaming, their complete information needs to be retrieved from Redis.
		if nodeExeEntity.Status == entity.NodeRunning {
			meta := entity.NodeMetaByNodeType(nodeExeEntity.NodeType)
			if meta.ExecutableMeta.IncrementalOutput {
				if err := e.loadNodeExecutionFromRedis(ctx, nodeExeEntity); err != nil {
					logs.CtxErrorf(ctx, "failed to load node execution from redis: %v", err)
				}
			}
		}
		result = append(result, nodeExeEntity)
	}
	return result, nil
}

// loadNodeExecutionFromRedis 从 Redis 加载节点的流式输出
//
// 参数：
//   - ctx: 上下文
//   - nodeExeEntity: 节点执行实体
//
// 返回值：
//   - error: 加载失败时返回错误
func (e *executeHistoryStoreImpl) loadNodeExecutionFromRedis(ctx context.Context, nodeExeEntity *entity.NodeExecution) error {
	key := fmt.Sprintf(nodeExecOutputKey, nodeExeEntity.ID)

	result, err := e.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, cache.Nil) {
			return nil
		}
		return vo.WrapError(errno.ErrRedisError, err)
	}

	if result != "" {
		nodeExeEntity.Output = &result
	}

	return nil
}

// GetNodeExecution 获取指定节点的执行记录
//
// 参数：
//   - ctx: 上下文
//   - wfExeID: 工作流执行 ID
//   - nodeID: 节点 ID
//
// 返回值：
//   - *entity.NodeExecution: 节点执行记录
//   - bool: 是否存在
//   - error: 获取失败时返回错误
func (e *executeHistoryStoreImpl) GetNodeExecution(ctx context.Context, wfExeID int64, nodeID string) (*entity.NodeExecution, bool, error) {
	nodeExec, err := e.query.NodeExecution.WithContext(ctx).
		Where(e.query.NodeExecution.ExecuteID.Eq(wfExeID), e.query.NodeExecution.NodeID.Eq(nodeID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("failed to find node executions: %w", err))
	}

	nodeExeEntity := convertNodeExecution(nodeExec)

	return nodeExeEntity, true, nil
}

// GetNodeExecutionByParent 根据父节点 ID 获取子节点执行记录
//
// 用于获取复合节点（如循环节点、批处理节点）内部的子节点执行记录。
//
// 参数：
//   - ctx: 上下文
//   - wfExeID: 工作流执行 ID
//   - parentNodeID: 父节点 ID
//
// 返回值：
//   - []*entity.NodeExecution: 子节点执行记录列表
//   - error: 获取失败时返回错误
func (e *executeHistoryStoreImpl) GetNodeExecutionByParent(ctx context.Context, wfExeID int64, parentNodeID string) (
	[]*entity.NodeExecution, error) {
	nodeExecs, err := e.query.NodeExecution.WithContext(ctx).
		Where(e.query.NodeExecution.ExecuteID.Eq(wfExeID), e.query.NodeExecution.ParentNodeID.Eq(parentNodeID)).
		Find()
	if err != nil {
		return nil, vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("failed to find node executions: %w", err))
	}
	var result []*entity.NodeExecution
	for _, nodeExec := range nodeExecs {
		nodeExeEntity := convertNodeExecution(nodeExec)
		result = append(result, nodeExeEntity)
	}
	return result, nil
}

const (
	// testRunLastExeKey 测试运行最新执行 ID 的 Redis 键模式
	// 格式：test_run_last_exe_id:{工作流ID}:{用户ID}
	testRunLastExeKey = "test_run_last_exe_id:%d:%d"

	// nodeDebugLastExeKey 节点调试最新执行 ID 的 Redis 键模式
	// 格式：node_debug_last_exe_id:{工作流ID}:{节点ID}:{用户ID}
	nodeDebugLastExeKey = "node_debug_last_exe_id:%d:%s:%d"

	// nodeExecDataExpiry 节点执行数据的过期时间（24 小时）
	nodeExecDataExpiry = 24 * time.Hour

	// nodeExecOutputKey 节点执行输出的 Redis 键模式
	// 格式：wf:node_exec:output:{节点执行ID}
	nodeExecOutputKey = "wf:node_exec:output:%d"
)

// SetTestRunLatestExeID 设置测试运行的最新执行 ID
//
// 用于前端获取最近一次测试运行的结果。
//
// 参数：
//   - ctx: 上下文
//   - wfID: 工作流 ID
//   - uID: 用户 ID
//   - exeID: 执行 ID
//
// 返回值：
//   - error: 设置失败时返回错误
func (e *executeHistoryStoreImpl) SetTestRunLatestExeID(ctx context.Context, wfID int64, uID int64, exeID int64) error {
	key := fmt.Sprintf(testRunLastExeKey, wfID, uID)
	err := e.redis.Set(ctx, key, exeID, 7*24*time.Hour).Err()
	if err != nil {
		return vo.WrapError(errno.ErrRedisError, err)
	}

	return nil
}

// GetTestRunLatestExeID 获取测试运行的最新执行 ID
//
// 参数：
//   - ctx: 上下文
//   - wfID: 工作流 ID
//   - uID: 用户 ID
//
// 返回值：
//   - int64: 执行 ID（不存在时返回 0）
//   - error: 获取失败时返回错误
func (e *executeHistoryStoreImpl) GetTestRunLatestExeID(ctx context.Context, wfID int64, uID int64) (int64, error) {
	key := fmt.Sprintf(testRunLastExeKey, wfID, uID)
	exeIDStr, err := e.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, cache.Nil) {
			return 0, nil
		}
		return 0, vo.WrapError(errno.ErrRedisError, err)
	}
	exeID, err := strconv.ParseInt(exeIDStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return exeID, nil
}

// SetNodeDebugLatestExeID 设置节点调试的最新执行 ID
//
// 参数：
//   - ctx: 上下文
//   - wfID: 工作流 ID
//   - nodeID: 节点 ID
//   - uID: 用户 ID
//   - exeID: 执行 ID
//
// 返回值：
//   - error: 设置失败时返回错误
func (e *executeHistoryStoreImpl) SetNodeDebugLatestExeID(ctx context.Context, wfID int64, nodeID string, uID int64, exeID int64) error {
	key := fmt.Sprintf(nodeDebugLastExeKey, wfID, nodeID, uID)
	err := e.redis.Set(ctx, key, exeID, 7*24*time.Hour).Err()
	if err != nil {
		return vo.WrapError(errno.ErrRedisError, err)
	}
	return nil
}

// GetNodeDebugLatestExeID 获取节点调试的最新执行 ID
//
// 参数：
//   - ctx: 上下文
//   - wfID: 工作流 ID
//   - nodeID: 节点 ID
//   - uID: 用户 ID
//
// 返回值：
//   - int64: 执行 ID（不存在时返回 0）
//   - error: 获取失败时返回错误
func (e *executeHistoryStoreImpl) GetNodeDebugLatestExeID(ctx context.Context, wfID int64, nodeID string, uID int64) (int64, error) {
	key := fmt.Sprintf(nodeDebugLastExeKey, wfID, nodeID, uID)
	exeIDStr, err := e.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, cache.Nil) {
			return 0, nil
		}
		return 0, vo.WrapError(errno.ErrRedisError, err)
	}
	exeID, err := strconv.ParseInt(exeIDStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return exeID, nil
}
