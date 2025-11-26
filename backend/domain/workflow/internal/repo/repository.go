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

// Package repo 实现工作流领域的仓储层
//
// 本包提供工作流实体的持久化和检索功能，是领域驱动设计中仓储模式的具体实现。
// 主要职责包括：
//   - 工作流元数据的 CRUD 操作
//   - 工作流草稿和版本管理
//   - 工作流执行历史记录
//   - 会话模板和会话管理
//   - 中断事件和取消信号的存储
//
// 技术实现：
//   - 使用 GORM + gen 进行数据库操作
//   - 使用 Redis 进行缓存和分布式状态管理
//   - 使用对象存储（TOS）存储大型资源
//
// 设计说明：
// 本包采用组合模式，将多个子存储（如 CheckPointStore、InterruptEventStore 等）
// 组合到 RepositoryImpl 中，实现职责分离和灵活扩展。
package repo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	einoCompose "github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/exp/maps"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	workflow3 "github.com/coze-dev/coze-studio/backend/api/model/workflow"
	"github.com/coze-dev/coze-studio/backend/application/base/ctxutil"
	"github.com/coze-dev/coze-studio/backend/bizpkg/llm/modelbuilder"
	workflowModel "github.com/coze-dev/coze-studio/backend/crossdomain/workflow/model"
	"github.com/coze-dev/coze-studio/backend/domain/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/canvas/adaptor"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/compose"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/execute"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/repo/dal/model"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/repo/dal/query"
	"github.com/coze-dev/coze-studio/backend/infra/cache"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
	"github.com/coze-dev/coze-studio/backend/pkg/errorx"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/slices"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ternary"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
	"github.com/coze-dev/coze-studio/backend/pkg/safego"
	"github.com/coze-dev/coze-studio/backend/pkg/sonic"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

const (
	// batchCreateSize 批量创建时的批次大小
	batchCreateSize = 10
)

// RepositoryImpl 工作流仓储的具体实现
//
// 该结构体实现了 workflow.Repository 接口，提供工作流相关的所有持久化操作。
// 采用组合模式嵌入多个子存储接口，实现功能的模块化：
//   - IDGenerator: ID 生成器，用于生成唯一标识
//   - CheckPointStore: 检查点存储，用于工作流执行的断点续传
//   - InterruptEventStore: 中断事件存储，用于工作流的暂停和恢复
//   - CancelSignalStore: 取消信号存储，用于工作流的取消操作
//   - ExecuteHistoryStore: 执行历史存储，用于记录工作流执行日志
//   - Suggester: 建议生成器，用于生成后续问题建议
type RepositoryImpl struct {
	idgen.IDGenerator                            // ID 生成器
	query                        *query.Query    // GORM gen 查询对象
	redis                        cache.Cmdable   // Redis 客户端
	tos                          storage.Storage // 对象存储客户端
	einoCompose.CheckPointStore                  // Eino 检查点存储
	workflow.InterruptEventStore                 // 中断事件存储
	workflow.CancelSignalStore                   // 取消信号存储
	workflow.ExecuteHistoryStore                 // 执行历史存储
	builtinModel                 modelbuilder.BaseChatModel
	workflow.WorkflowConfig      // 工作流配置
	workflow.Suggester           // 建议生成器
}

// NewRepository 创建工作流仓储实例
//
// 参数：
//   - idgen: ID 生成器，用于生成唯一标识
//   - db: GORM 数据库连接
//   - redis: Redis 客户端，用于缓存和分布式状态管理
//   - tos: 对象存储客户端，用于存储大型资源
//   - cpStore: 检查点存储，用于工作流执行的断点续传
//   - chatModel: 聊天模型，用于生成后续问题建议
//   - workflowConfig: 工作流配置
//
// 返回值：
//   - workflow.Repository: 工作流仓储接口实现
//   - error: 创建失败时返回错误
func NewRepository(idgen idgen.IDGenerator, db *gorm.DB, redis cache.Cmdable, tos storage.Storage,
	cpStore einoCompose.CheckPointStore, chatModel modelbuilder.BaseChatModel, workflowConfig workflow.WorkflowConfig) (workflow.Repository, error) {
	var sg workflow.Suggester
	var err error
	if chatModel != nil {
		sg, err = NewSuggester(chatModel)
		if err != nil {
			return nil, err
		}
	} else {
		logs.Warnf("[NewRepository] Failed to create suggester: %v", err)
	}

	return &RepositoryImpl{
		IDGenerator:     idgen,
		query:           query.Use(db),
		redis:           redis,
		tos:             tos,
		CheckPointStore: cpStore,
		InterruptEventStore: &interruptEventStoreImpl{
			redis: redis,
		},
		CancelSignalStore: &cancelSignalStoreImpl{
			redis: redis,
		},
		ExecuteHistoryStore: &executeHistoryStoreImpl{
			query: query.Use(db),
			redis: redis,
		},

		builtinModel:   chatModel,
		Suggester:      sg,
		WorkflowConfig: workflowConfig,
	}, nil

}

// Suggest 根据用户输入和助手回答生成后续问题建议
//
// 参数：
//   - ctx: 上下文
//   - input: 建议请求信息，包含用户输入和助手回答
//
// 返回值：
//   - []string: 生成的后续问题建议列表
//   - error: 生成失败时返回错误
func (r *RepositoryImpl) Suggest(ctx context.Context, input *vo.SuggestInfo) ([]string, error) {
	if r.Suggester == nil {
		return []string{}, nil
	}
	return r.Suggester.Suggest(ctx, input)
}

// CreateMeta 创建工作流元数据
//
// 参数：
//   - ctx: 上下文
//   - meta: 工作流元数据，包含名称、描述、图标等基本信息
//
// 返回值：
//   - int64: 创建的工作流 ID
//   - error: 创建失败时返回错误
func (r *RepositoryImpl) CreateMeta(ctx context.Context, meta *vo.Meta) (int64, error) {
	id, err := r.GenID(ctx)
	if err != nil {
		return 0, vo.WrapError(errno.ErrIDGenError, err)
	}
	wfMeta := &model.WorkflowMeta{
		ID:          id,
		Name:        meta.Name,
		Description: meta.Desc,
		IconURI:     meta.IconURI,
		ContentType: int32(meta.ContentType),
		Mode:        int32(meta.Mode),
		CreatorID:   meta.CreatorID,
		AuthorID:    meta.AuthorID,
		SpaceID:     meta.SpaceID,
		DeletedAt:   gorm.DeletedAt{Valid: false},
	}

	if meta.Tag != nil {
		wfMeta.Tag = int32(*meta.Tag)
	}

	if meta.SourceID != nil {
		wfMeta.SourceID = *meta.SourceID
	}

	if meta.AppID != nil {
		wfMeta.AppID = *meta.AppID
	}

	if err = r.query.WorkflowMeta.Create(wfMeta); err != nil {
		return 0, vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("create workflow meta: %w", err))
	}

	return id, nil
}

// updateReferences 更新工作流引用关系
//
// 该方法用于维护工作流之间的引用关系（如子工作流引用）。
// 采用增量更新策略：比较当前引用和目标引用，执行新增、禁用、启用操作。
//
// 参数：
//   - ctx: 上下文
//   - id: 引用方工作流 ID
//   - wfRefs: 目标引用关系集合
//
// 返回值：
//   - error: 更新失败时返回错误
func (r *RepositoryImpl) updateReferences(ctx context.Context, id int64, wfRefs map[entity.WorkflowReferenceKey]struct{}) (
	err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	currentRefs, err := r.query.WorkflowReference.WithContext(ctx).Where(
		r.query.WorkflowReference.ReferringID.Eq(id)).Find()
	if err != nil {
		return fmt.Errorf("failed to find workflow reference: %w", err)
	}

	if len(currentRefs) == 0 {
		if len(wfRefs) == 0 {
			return nil
		}

		refsToCreateModel := make([]*model.WorkflowReference, 0, len(wfRefs))
		refIDs, err := r.GenMultiIDs(ctx, len(wfRefs))
		if err != nil {
			return fmt.Errorf("failed to gen id for workflow reference: %w", err)
		}
		var index int
		for key := range wfRefs {
			refsToCreateModel = append(refsToCreateModel, &model.WorkflowReference{
				ID:               refIDs[index],
				ReferredID:       key.ReferredID,
				ReferringID:      key.ReferringID,
				ReferType:        int32(key.ReferType),
				ReferringBizType: int32(key.ReferringBizType),
				Status:           1,
			})
			index++
		}

		return r.query.WorkflowReference.WithContext(ctx).Create(refsToCreateModel...)
	}

	if len(wfRefs) == 0 {
		_, err = r.query.WorkflowReference.WithContext(ctx).
			Where(r.query.WorkflowReference.ID.In(slices.Transform(currentRefs,
				func(reference *model.WorkflowReference) int64 {
					return reference.ID
				})...)).
			UpdateColumnSimple(r.query.WorkflowReference.Status.Value(0))
		return err
	}

	var (
		refsToDisable  []int64
		refsToEnable   []int64
		refsToCreate   = maps.Clone(wfRefs)
		existingRefMap = slices.ToMap(currentRefs, func(reference *model.WorkflowReference) (
			entity.WorkflowReferenceKey, *model.WorkflowReference) {
			return entity.WorkflowReferenceKey{
				ReferredID:       reference.ReferredID,
				ReferringID:      reference.ReferringID,
				ReferType:        vo.ReferType(reference.ReferType),
				ReferringBizType: vo.ReferringBizType(reference.ReferringBizType),
			}, reference
		})
	)
	for key, ref := range existingRefMap {
		if ref.Status == 1 {
			if _, ok := wfRefs[key]; !ok {
				refsToDisable = append(refsToDisable, ref.ID)
			}
		} else {
			if _, ok := wfRefs[key]; ok {
				refsToEnable = append(refsToEnable, ref.ID)
				delete(refsToCreate, key)
			}
		}
	}

	for key := range refsToCreate {
		if _, ok := existingRefMap[key]; ok {
			delete(refsToCreate, key)
		}
	}

	if len(refsToCreate) > 0 {
		refsToCreateModel := make([]*model.WorkflowReference, 0, len(refsToCreate))
		refIDs, err := r.GenMultiIDs(ctx, len(refsToCreate))
		if err != nil {
			return fmt.Errorf("failed to gen id for workflow reference: %w", err)
		}
		var index int
		for key := range refsToCreate {
			refsToCreateModel = append(refsToCreateModel, &model.WorkflowReference{
				ID:               refIDs[index],
				ReferredID:       key.ReferredID,
				ReferringID:      key.ReferringID,
				ReferType:        int32(key.ReferType),
				ReferringBizType: int32(key.ReferringBizType),
				Status:           1,
			})
			index++
		}

		if err = r.query.WorkflowReference.WithContext(ctx).Create(refsToCreateModel...); err != nil {
			return fmt.Errorf("failed to create workflow reference for workflowID %d, childIDs %v: %v",
				id, refsToCreate, err)
		}
	}

	if len(refsToDisable) > 0 {
		_, err = r.query.WorkflowReference.WithContext(ctx).
			Where(r.query.WorkflowReference.ID.In(refsToDisable...)).
			UpdateColumnSimple(r.query.WorkflowReference.Status.Value(0))
		if err != nil {
			return fmt.Errorf("failed to disable workflow reference for workflowID %d, childIDs %v: %v",
				id, refsToDisable, err)
		}
	}

	if len(refsToEnable) > 0 {
		_, err = r.query.WorkflowReference.WithContext(ctx).
			Where(r.query.WorkflowReference.ID.In(refsToEnable...)).
			UpdateColumnSimple(r.query.WorkflowReference.Status.Value(1))
		if err != nil {
			return fmt.Errorf("failed to enable workflow reference for workflowID %d, childIDs %v: %v",
				id, refsToEnable, err)
		}
	}

	return nil
}

// CreateVersion 创建工作流版本（发布操作）
//
// 该方法用于将工作流草稿发布为正式版本。发布时会：
// 1. 更新工作流引用关系
// 2. 创建版本记录
// 3. 更新草稿状态（标记为已测试运行成功、未修改）
// 4. 更新元数据状态（标记为已发布）
//
// 参数：
//   - ctx: 上下文
//   - id: 工作流 ID
//   - info: 版本信息，包含版本号、描述、画布内容等
//   - newRefs: 新的引用关系
//
// 返回值：
//   - error: 创建失败时返回错误
func (r *RepositoryImpl) CreateVersion(ctx context.Context, id int64, info *vo.VersionInfo, newRefs map[entity.WorkflowReferenceKey]struct{}) (err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	if err = r.updateReferences(ctx, id, newRefs); err != nil {
		return err
	}

	if err = r.query.WorkflowVersion.WithContext(ctx).Create(&model.WorkflowVersion{
		// ID: auto_increment
		WorkflowID:         id,
		Version:            info.Version,
		VersionDescription: info.VersionDescription,
		Canvas:             info.Canvas,
		InputParams:        info.InputParamsStr,
		OutputParams:       info.OutputParamsStr,
		CreatorID:          info.VersionCreatorID,
		CommitID:           info.CommitID,
	}); err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}

	var result gen.ResultInfo
	result, err = r.query.WorkflowDraft.WithContext(ctx).
		Where(r.query.WorkflowDraft.ID.Eq(id),
			r.query.WorkflowDraft.CommitID.Eq(info.CommitID)).
		UpdateColumnSimple(
			r.query.WorkflowDraft.Modified.Value(false),
			r.query.WorkflowDraft.TestRunSuccess.Value(true),
		)
	if err != nil {
		return fmt.Errorf("update workflow draft when publish failed: %w", err)
	}

	if result.RowsAffected == 0 {
		logs.CtxWarnf(ctx, "update workflow draft when publish failed: no rows affected. WorkflowID: %d", id)
	}

	_, err = r.query.WorkflowMeta.WithContext(ctx).
		Where(r.query.WorkflowMeta.ID.Eq(id)).
		UpdateColumnSimple(
			r.query.WorkflowMeta.Status.Value(1),
			r.query.WorkflowMeta.LatestVersion.Value(info.Version),
			r.query.WorkflowMeta.LatestVersionTs.Value(time.Now().UnixMilli()),
		)
	if err != nil {
		logs.CtxWarnf(ctx, "update workflow meta when publish failed: %v", err)
	}

	return nil
}

// CreateOrUpdateDraft 创建或更新工作流草稿
//
// 参数：
//   - ctx: 上下文
//   - id: 工作流 ID
//   - draft: 草稿信息，包含画布内容、输入输出参数等
//
// 返回值：
//   - error: 操作失败时返回错误
func (r *RepositoryImpl) CreateOrUpdateDraft(ctx context.Context, id int64, draft *vo.DraftInfo) error {
	d := &model.WorkflowDraft{
		ID:           id,
		Canvas:       draft.Canvas,
		InputParams:  draft.InputParamsStr,
		OutputParams: draft.OutputParamsStr,
		CommitID:     draft.CommitID,
	}

	if draft.DraftMeta != nil {
		d.Modified = draft.DraftMeta.Modified
		d.TestRunSuccess = draft.DraftMeta.TestRunSuccess
	}

	if err := r.query.WorkflowDraft.WithContext(ctx).Save(d); err != nil {
		return vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("save workflow draft: %w", err))
	}

	return nil
}

// UpdateWorkflowDraftTestRunSuccess 更新工作流草稿的测试运行成功状态
//
// 参数：
//   - ctx: 上下文
//   - id: 工作流 ID
//
// 返回值：
//   - error: 更新失败时返回错误
func (r *RepositoryImpl) UpdateWorkflowDraftTestRunSuccess(ctx context.Context, id int64) error {
	if _, err := r.query.WorkflowDraft.WithContext(ctx).Where(r.query.WorkflowDraft.ID.Eq(id)).UpdateColumnSimple(r.query.WorkflowDraft.TestRunSuccess.Value(true)); err != nil {
		return vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("update workflow draft test run success failed: %w", err))
	}

	return nil
}

// Delete 删除工作流及其相关数据
//
// 该方法在事务中执行，会删除：
//   - 工作流元数据
//   - 工作流草稿
//   - 所有版本记录
//   - 引用关系（被引用和引用他人的关系）
//
// 参数：
//   - ctx: 上下文
//   - id: 工作流 ID
//
// 返回值：
//   - error: 删除失败时返回错误
func (r *RepositoryImpl) Delete(ctx context.Context, id int64) (err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	return r.query.Transaction(func(tx *query.Query) error {
		// Delete from workflow_meta
		_, err := tx.WorkflowMeta.WithContext(ctx).Where(tx.WorkflowMeta.ID.Eq(id)).Delete()
		if err != nil {
			return fmt.Errorf("delete workflow meta: %w", err)
		}

		_, err = tx.WorkflowDraft.WithContext(ctx).Where(tx.WorkflowDraft.ID.Eq(id)).Delete()
		if err != nil {
			return fmt.Errorf("delete workflow draft: %w", err)
		}

		_, err = tx.WorkflowVersion.WithContext(ctx).Where(tx.WorkflowVersion.WorkflowID.Eq(id)).Delete()
		if err != nil {
			return fmt.Errorf("delete workflow versions: %w", err)
		}

		_, err = tx.WorkflowReference.WithContext(ctx).Where(tx.WorkflowReference.ReferredID.Eq(id)).Delete()
		if err != nil {
			return fmt.Errorf("delete workflow references: %w", err)
		}

		_, err = tx.WorkflowReference.WithContext(ctx).Where(tx.WorkflowReference.ReferringID.Eq(id)).Delete()
		if err != nil {
			return fmt.Errorf("delete incoming workflow references: %w", err)
		}

		return nil
	})
}

// MDelete 批量删除工作流
//
// 该方法会同步删除元数据，异步删除草稿、版本和引用关系。
//
// 参数：
//   - ctx: 上下文
//   - ids: 要删除的工作流 ID 列表
//
// 返回值：
//   - error: 删除失败时返回错误
func (r *RepositoryImpl) MDelete(ctx context.Context, ids []int64) error {
	_, err := r.query.WorkflowMeta.WithContext(ctx).Where(r.query.WorkflowMeta.ID.In(ids...)).Delete()
	if err != nil {
		return vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("delete workflow meta failed err=%w", err))
	}

	safego.Go(ctx, func() {
		_, err = r.query.WorkflowDraft.WithContext(ctx).Where(r.query.WorkflowDraft.ID.In(ids...)).Delete()
		if err != nil {
			logs.Warnf("delete workflow draft failed err=%v, ids %v", err, ids)
		}

		_, err = r.query.WorkflowVersion.WithContext(ctx).Where(r.query.WorkflowVersion.WorkflowID.In(ids...)).Delete()
		if err != nil {
			logs.Warnf("delete workflow version failed err=%v, ids %v", err, ids)
		}

		_, err = r.query.WorkflowReference.WithContext(ctx).Where(r.query.WorkflowReference.ID.In(ids...)).Delete()
		if err != nil {
			logs.Warnf("delete workflow reference failed err=%v, ids %v", err, ids)

		}
		_, err = r.query.WorkflowReference.WithContext(ctx).Where(r.query.WorkflowReference.ReferringID.In(ids...)).Delete()
		if err != nil {
			logs.Warnf("delete workflow reference refer failed err=%v, ids %v", err, ids)
		}
	})

	return nil
}

// GetMeta 获取工作流元数据
//
// 参数：
//   - ctx: 上下文
//   - id: 工作流 ID
//
// 返回值：
//   - *vo.Meta: 工作流元数据
//   - error: 获取失败时返回错误（包括不存在的情况）
func (r *RepositoryImpl) GetMeta(ctx context.Context, id int64) (_ *vo.Meta, err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	meta, err := r.query.WorkflowMeta.WithContext(ctx).Debug().Where(r.query.WorkflowMeta.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, vo.WrapError(errno.ErrWorkflowNotFound, fmt.Errorf("workflow meta not found for ID %d: %w", id, err),
				errorx.KV("id", strconv.FormatInt(id, 10)))
		}
		return nil, fmt.Errorf("failed to get workflow meta for ID %d: %w", id, err)
	}

	return r.convertMeta(ctx, meta)
}

// convertMeta 将数据库模型转换为领域值对象
//
// 参数：
//   - ctx: 上下文
//   - meta: 数据库工作流元数据模型
//
// 返回值：
//   - *vo.Meta: 领域层元数据值对象
//   - error: 转换失败时返回错误
func (r *RepositoryImpl) convertMeta(ctx context.Context, meta *model.WorkflowMeta) (*vo.Meta, error) {
	url, err := r.tos.GetObjectUrl(ctx, meta.IconURI)
	if err != nil {
		logs.Warnf("failed to get url for workflow meta %v", err)
	}
	// Initialize the result entity
	wfMeta := &vo.Meta{
		Name:        meta.Name,
		Desc:        meta.Description,
		IconURI:     meta.IconURI,
		IconURL:     url,
		ContentType: entity.ContentType(meta.ContentType),
		Mode:        entity.Mode(meta.Mode),
		CreatorID:   meta.CreatorID,
		AuthorID:    meta.AuthorID,
		SpaceID:     meta.SpaceID,
		CreatedAt:   time.UnixMilli(meta.CreatedAt),
	}
	if meta.Tag != 0 {
		tag := entity.Tag(meta.Tag)
		wfMeta.Tag = &tag
	}
	if meta.SourceID != 0 {
		wfMeta.SourceID = &meta.SourceID
	}
	if meta.AppID != 0 {
		wfMeta.AppID = &meta.AppID
	}
	if meta.UpdatedAt > 0 {
		wfMeta.UpdatedAt = ptr.Of(time.UnixMilli(meta.UpdatedAt))
	}
	if meta.Status > 0 {
		wfMeta.HasPublished = true
	}
	if meta.LatestVersion != "" {
		wfMeta.LatestPublishedVersion = ptr.Of(meta.LatestVersion)
	}

	return wfMeta, nil
}

// UpdateMeta 更新工作流元数据
//
// 支持部分更新，只更新传入的非空字段。
//
// 参数：
//   - ctx: 上下文
//   - id: 工作流 ID
//   - metaUpdate: 要更新的字段（非空字段才会更新）
//
// 返回值：
//   - error: 更新失败时返回错误
func (r *RepositoryImpl) UpdateMeta(ctx context.Context, id int64, metaUpdate *vo.MetaUpdate) error {
	var expressions []field.AssignExpr

	if metaUpdate.Name != nil {
		expressions = append(expressions, r.query.WorkflowMeta.Name.Value(*metaUpdate.Name))
	}

	if metaUpdate.Desc != nil {
		expressions = append(expressions, r.query.WorkflowMeta.Description.Value(*metaUpdate.Desc))
	}

	if metaUpdate.IconURI != nil {
		expressions = append(expressions, r.query.WorkflowMeta.IconURI.Value(*metaUpdate.IconURI))
	}

	if metaUpdate.HasPublished != nil {
		if *metaUpdate.HasPublished {
			expressions = append(expressions, r.query.WorkflowMeta.Status.Value(1))
		} else {
			expressions = append(expressions, r.query.WorkflowMeta.Status.Value(0))
		}
	}

	if metaUpdate.LatestPublishedVersion != nil {
		expressions = append(expressions, r.query.WorkflowMeta.LatestVersion.Value(*metaUpdate.LatestPublishedVersion))
	}

	if metaUpdate.WorkflowMode != nil {
		expressions = append(expressions, r.query.WorkflowMeta.Mode.Value(int32(*metaUpdate.WorkflowMode)))
	}

	if len(expressions) == 0 {
		return nil
	}

	_, err := r.query.WorkflowMeta.WithContext(ctx).Where(r.query.WorkflowMeta.ID.Eq(id)).
		UpdateColumnSimple(expressions...)
	if err != nil {
		return vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("update workflow meta: %w", err))
	}

	return nil
}

// GetEntity 获取完整的工作流实体
//
// 根据查询策略获取工作流实体，支持多种查询模式：
//   - FromDraft: 从草稿获取
//   - FromSpecificVersion: 从指定版本获取
//   - FromLatestVersion: 从最新版本获取
//   - MetaOnly: 仅获取元数据
//
// 参数：
//   - ctx: 上下文
//   - policy: 查询策略，包含查询类型、版本号等
//
// 返回值：
//   - *entity.Workflow: 工作流实体
//   - error: 获取失败时返回错误
func (r *RepositoryImpl) GetEntity(ctx context.Context, policy *vo.GetPolicy) (_ *entity.Workflow, err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	meta, err := r.GetMeta(ctx, policy.ID)
	if err != nil {
		return nil, err
	}

	if policy.MetaOnly {
		return &entity.Workflow{
			ID:   policy.ID,
			Meta: meta,
		}, nil
	}

	var (
		canvas, inputParams, outputParams string
		draftMeta                         *vo.DraftMeta
		versionMeta                       *vo.VersionMeta
		commitID                          string
	)
	switch policy.QType {
	case workflowModel.FromDraft:
		draft, err := r.DraftV2(ctx, policy.ID, policy.CommitID)
		if err != nil {
			return nil, err
		}

		canvas = draft.Canvas
		inputParams = draft.InputParamsStr
		outputParams = draft.OutputParamsStr
		draftMeta = draft.DraftMeta
		commitID = draft.CommitID
	case workflowModel.FromSpecificVersion:
		v, existed, err := r.GetVersion(ctx, policy.ID, policy.Version)
		if err != nil {
			return nil, err
		}
		if !existed {
			return nil, vo.WrapError(errno.ErrWorkflowNotFound, fmt.Errorf("workflow version %s not found for ID %d: %w", policy.Version, policy.ID, err), errorx.KV("id", strconv.FormatInt(policy.ID, 10)))
		}
		canvas = v.Canvas
		inputParams = v.InputParamsStr
		outputParams = v.OutputParamsStr
		versionMeta = v.VersionMeta
		commitID = v.CommitID
	case workflowModel.FromLatestVersion:
		v, err := r.GetLatestVersion(ctx, policy.ID)
		if err != nil {
			return nil, err
		}
		canvas = v.Canvas
		inputParams = v.InputParamsStr
		outputParams = v.OutputParamsStr
		versionMeta = v.VersionMeta
		commitID = v.CommitID
	default:
		panic(fmt.Sprintf("invalid query type: %v", policy.QType))
	}

	var inputs, outputs []*vo.NamedTypeInfo
	if inputParams != "" {
		err = sonic.UnmarshalString(inputParams, &inputs)
		if err != nil {
			return nil, vo.WrapError(errno.ErrSerializationDeserializationFail, err)
		}
	}
	if outputParams != "" {
		err = sonic.UnmarshalString(outputParams, &outputs)
		if err != nil {
			return nil, vo.WrapError(errno.ErrSerializationDeserializationFail, err)
		}
	}

	return &entity.Workflow{
		ID:       policy.ID,
		CommitID: commitID,
		Meta:     meta,
		CanvasInfo: &vo.CanvasInfo{
			Canvas:          canvas,
			InputParams:     inputs,
			OutputParams:    outputs,
			InputParamsStr:  inputParams,
			OutputParamsStr: outputParams,
		},
		DraftMeta:   draftMeta,
		VersionMeta: versionMeta,
	}, nil
}

// CreateChatFlowRoleConfig 创建 ChatFlow 角色配置
//
// ChatFlow 是一种特殊的工作流模式，支持多轮对话，角色配置定义了对话助手的外观和行为。
//
// 参数：
//   - ctx: 上下文
//   - chatFlowRole: ChatFlow 角色配置，包含名称、头像、背景等
//
// 返回值：
//   - int64: 创建的配置 ID
//   - error: 创建失败时返回错误
func (r *RepositoryImpl) CreateChatFlowRoleConfig(ctx context.Context, chatFlowRole *entity.ChatFlowRole) (int64, error) {
	id, err := r.GenID(ctx)
	if err != nil {
		return 0, vo.WrapError(errno.ErrIDGenError, err)
	}
	chatFlowRoleConfig := &model.ChatFlowRoleConfig{
		ID:                  id,
		WorkflowID:          chatFlowRole.WorkflowID,
		Name:                chatFlowRole.Name,
		Description:         chatFlowRole.Description,
		Avatar:              chatFlowRole.AvatarUri,
		AudioConfig:         chatFlowRole.AudioConfig,
		BackgroundImageInfo: chatFlowRole.BackgroundImageInfo,
		OnboardingInfo:      chatFlowRole.OnboardingInfo,
		SuggestReplyInfo:    chatFlowRole.SuggestReplyInfo,
		UserInputConfig:     chatFlowRole.UserInputConfig,
		CreatorID:           chatFlowRole.CreatorID,
		Version:             chatFlowRole.Version,
	}

	if err := r.query.ChatFlowRoleConfig.WithContext(ctx).Create(chatFlowRoleConfig); err != nil {
		return 0, vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("create chat flow role: %w", err))
	}

	return id, nil
}

// UpdateChatFlowRoleConfig 更新 ChatFlow 角色配置
//
// 支持部分更新，只更新传入的非空字段。
//
// 参数：
//   - ctx: 上下文
//   - workflowID: 工作流 ID
//   - chatFlowRole: 要更新的配置字段
//
// 返回值：
//   - error: 更新失败时返回错误
func (r *RepositoryImpl) UpdateChatFlowRoleConfig(ctx context.Context, workflowID int64, chatFlowRole *vo.ChatFlowRoleUpdate) error {
	var expressions []field.AssignExpr
	if chatFlowRole.Name != nil {
		expressions = append(expressions, r.query.ChatFlowRoleConfig.Name.Value(*chatFlowRole.Name))
	}
	if chatFlowRole.Description != nil {
		expressions = append(expressions, r.query.ChatFlowRoleConfig.Description.Value(*chatFlowRole.Description))
	}
	if chatFlowRole.AvatarUri != nil {
		expressions = append(expressions, r.query.ChatFlowRoleConfig.Avatar.Value(*chatFlowRole.AvatarUri))
	}
	if chatFlowRole.AudioConfig != nil {
		expressions = append(expressions, r.query.ChatFlowRoleConfig.AudioConfig.Value(*chatFlowRole.AudioConfig))
	}
	if chatFlowRole.BackgroundImageInfo != nil {
		expressions = append(expressions, r.query.ChatFlowRoleConfig.BackgroundImageInfo.Value(*chatFlowRole.BackgroundImageInfo))
	}
	if chatFlowRole.OnboardingInfo != nil {
		expressions = append(expressions, r.query.ChatFlowRoleConfig.OnboardingInfo.Value(*chatFlowRole.OnboardingInfo))
	}
	if chatFlowRole.SuggestReplyInfo != nil {
		expressions = append(expressions, r.query.ChatFlowRoleConfig.SuggestReplyInfo.Value(*chatFlowRole.SuggestReplyInfo))
	}
	if chatFlowRole.UserInputConfig != nil {
		expressions = append(expressions, r.query.ChatFlowRoleConfig.UserInputConfig.Value(*chatFlowRole.UserInputConfig))
	}

	if len(expressions) == 0 {
		return nil
	}

	_, err := r.query.ChatFlowRoleConfig.WithContext(ctx).Where(r.query.ChatFlowRoleConfig.WorkflowID.Eq(workflowID)).
		UpdateColumnSimple(expressions...)
	if err != nil {
		return vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("update chat flow role: %w", err))
	}

	return nil
}

// GetChatFlowRoleConfig 获取 ChatFlow 角色配置
//
// 参数：
//   - ctx: 上下文
//   - workflowID: 工作流 ID
//   - version: 版本号（可选，为空时返回最新配置）
//
// 返回值：
//   - *entity.ChatFlowRole: 角色配置
//   - error: 获取失败时返回错误
//   - bool: 配置是否存在
func (r *RepositoryImpl) GetChatFlowRoleConfig(ctx context.Context, workflowID int64, version string) (_ *entity.ChatFlowRole, err error, isExist bool) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()
	role := &model.ChatFlowRoleConfig{}
	if version != "" {
		role, err = r.query.ChatFlowRoleConfig.WithContext(ctx).Where(r.query.ChatFlowRoleConfig.WorkflowID.Eq(workflowID), r.query.ChatFlowRoleConfig.Version.Eq(version)).First()
	} else {
		role, err = r.query.ChatFlowRoleConfig.WithContext(ctx).Where(r.query.ChatFlowRoleConfig.WorkflowID.Eq(workflowID)).First()
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err, false
		}
		return nil, fmt.Errorf("failed to get chat flow role for chatflowID %d: %w", workflowID, err), true
	}
	res := &entity.ChatFlowRole{
		ID:                  role.ID,
		WorkflowID:          role.WorkflowID,
		Name:                role.Name,
		Description:         role.Description,
		AvatarUri:           role.Avatar,
		AudioConfig:         role.AudioConfig,
		BackgroundImageInfo: role.BackgroundImageInfo,
		OnboardingInfo:      role.OnboardingInfo,
		SuggestReplyInfo:    role.SuggestReplyInfo,
		UserInputConfig:     role.UserInputConfig,
		CreatorID:           role.CreatorID,
		CreatedAt:           time.UnixMilli(role.CreatedAt),
	}
	if role.UpdatedAt > 0 {
		res.UpdatedAt = time.UnixMilli(role.UpdatedAt)
	}
	return res, err, true
}

// DeleteChatFlowRoleConfig 删除 ChatFlow 角色配置
//
// 参数：
//   - ctx: 上下文
//   - id: 配置 ID
//   - workflowID: 工作流 ID
//
// 返回值：
//   - error: 删除失败时返回错误
func (r *RepositoryImpl) DeleteChatFlowRoleConfig(ctx context.Context, id int64, workflowID int64) error {
	_, err := r.query.ChatFlowRoleConfig.WithContext(ctx).Where(r.query.ChatFlowRoleConfig.ID.Eq(id), r.query.ChatFlowRoleConfig.WorkflowID.Eq(workflowID)).Delete()
	return err
}

// GetVersion 获取指定版本的工作流
//
// 参数：
//   - ctx: 上下文
//   - id: 工作流 ID
//   - version: 版本号
//
// 返回值：
//   - *vo.VersionInfo: 版本信息
//   - bool: 版本是否存在
//   - error: 获取失败时返回错误
func (r *RepositoryImpl) GetVersion(ctx context.Context, id int64, version string) (_ *vo.VersionInfo, existed bool, err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	wfVersion, err := r.query.WorkflowVersion.WithContext(ctx).
		Where(r.query.WorkflowVersion.WorkflowID.Eq(id), r.query.WorkflowVersion.Version.Eq(version)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to get workflow version %s for ID %d: %w", version, id, err)
	}

	return &vo.VersionInfo{
		VersionMeta: &vo.VersionMeta{
			Version:            wfVersion.Version,
			VersionDescription: wfVersion.VersionDescription,
			VersionCreatedAt:   time.UnixMilli(wfVersion.CreatedAt),
			VersionCreatorID:   wfVersion.CreatorID,
		},
		CanvasInfo: vo.CanvasInfo{
			Canvas:          wfVersion.Canvas,
			InputParamsStr:  wfVersion.InputParams,
			OutputParamsStr: wfVersion.OutputParams,
		},
		CommitID: wfVersion.CommitID,
	}, true, nil
}

// GetVersionListByConnectorAndWorkflowID 获取连接器关联的工作流版本列表
//
// 参数：
//   - ctx: 上下文
//   - connectorID: 连接器 ID
//   - workflowID: 工作流 ID
//   - limit: 返回数量限制
//
// 返回值：
//   - []string: 版本号列表（按创建时间降序）
//   - error: 获取失败时返回错误
func (r *RepositoryImpl) GetVersionListByConnectorAndWorkflowID(ctx context.Context, connectorID, workflowID int64, limit int) (_ []string, err error) {
	if limit <= 0 {
		return nil, vo.WrapError(errno.ErrInvalidParameter, errors.New("limit must be greater than 0"))
	}

	connectorWorkflowVersion := r.query.ConnectorWorkflowVersion
	vl, err := connectorWorkflowVersion.WithContext(ctx).
		Where(connectorWorkflowVersion.ConnectorID.Eq(connectorID),
			connectorWorkflowVersion.WorkflowID.Eq(workflowID)).
		Order(connectorWorkflowVersion.CreatedAt.Desc()).
		Limit(limit).
		Find()
	if err != nil {
		return nil, vo.WrapError(errno.ErrDatabaseError, err)
	}
	var versionList []string
	for _, v := range vl {
		versionList = append(versionList, v.Version)
	}
	return versionList, nil
}

// IsApplicationConnectorWorkflowVersion 检查是否为应用连接器的工作流版本
//
// 参数：
//   - ctx: 上下文
//   - connectorID: 连接器 ID
//   - workflowID: 工作流 ID
//   - version: 版本号
//
// 返回值：
//   - bool: 是否为应用连接器版本
//   - error: 检查失败时返回错误
func (r *RepositoryImpl) IsApplicationConnectorWorkflowVersion(ctx context.Context, connectorID, workflowID int64, version string) (b bool, err error) {
	connectorWorkflowVersion := r.query.ConnectorWorkflowVersion
	_, err = connectorWorkflowVersion.WithContext(ctx).
		Where(connectorWorkflowVersion.ConnectorID.Eq(connectorID),
			connectorWorkflowVersion.WorkflowID.Eq(workflowID),
			connectorWorkflowVersion.Version.Eq(version)).
		First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, vo.WrapError(errno.ErrDatabaseError, err)
	}
	return true, nil
}

// DraftV2 获取工作流草稿（V2 版本）
//
// 支持通过 commitID 获取特定快照版本的草稿。
// 如果指定的 commitID 在草稿表中不存在，会尝试从快照表中获取。
//
// 参数：
//   - ctx: 上下文
//   - id: 工作流 ID
//   - commitID: 提交 ID（可选，为空时返回最新草稿）
//
// 返回值：
//   - *vo.DraftInfo: 草稿信息
//   - error: 获取失败时返回错误
func (r *RepositoryImpl) DraftV2(ctx context.Context, id int64, commitID string) (_ *vo.DraftInfo, err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	var conds []gen.Condition
	conds = append(conds, r.query.WorkflowDraft.ID.Eq(id))
	if commitID != "" {
		conds = append(conds, r.query.WorkflowDraft.CommitID.Eq(commitID))
	}

	draft, err := r.query.WorkflowDraft.WithContext(ctx).Where(conds...).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if len(commitID) == 0 {
				return nil, vo.WrapError(errno.ErrWorkflowNotFound, fmt.Errorf("workflow draft not found for ID %d: %w", id, err),
					errorx.KV("id", strconv.FormatInt(id, 10)))
			} else {
				snapshot, err := r.query.WorkflowSnapshot.WithContext(ctx).Where(
					r.query.WorkflowSnapshot.WorkflowID.Eq(id),
					r.query.WorkflowSnapshot.CommitID.Eq(commitID),
				).First()
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return nil, vo.WrapError(errno.ErrWorkflowSnapshotNotFound,
							fmt.Errorf("workflow snapshot not found for ID %d, commitID %s: %w",
								id, commitID, err),
							errorx.KV("id", strconv.FormatInt(id, 10)),
							errorx.KV("commit_id", commitID))
					} else {
						return nil, fmt.Errorf("failed to query workflow snapshot for ID %d, commitID %s: %w",
							id, commitID, err)
					}
				}

				return &vo.DraftInfo{
					DraftMeta: &vo.DraftMeta{
						Timestamp:  time.UnixMilli(snapshot.CreatedAt),
						IsSnapshot: true,
					},

					Canvas:          snapshot.Canvas,
					InputParamsStr:  snapshot.InputParams,
					OutputParamsStr: snapshot.OutputParams,
					CommitID:        snapshot.CommitID,
				}, nil
			}
		}
		return nil, fmt.Errorf("failed to get workflow draft for ID %d, commitID %s: %w", id, commitID, err)
	}

	return &vo.DraftInfo{
		DraftMeta: &vo.DraftMeta{
			TestRunSuccess: draft.TestRunSuccess,
			Modified:       draft.Modified,
			Timestamp:      time.UnixMilli(draft.UpdatedAt),
			IsSnapshot:     false,
		},

		Canvas:          draft.Canvas,
		InputParamsStr:  draft.InputParams,
		OutputParamsStr: draft.OutputParams,
		CommitID:        draft.CommitID,
	}, nil
}

// MGetDrafts 批量获取工作流草稿
//
// 支持按 ID 列表、名称、空间 ID、应用 ID 等条件查询，并支持分页。
//
// 参数：
//   - ctx: 上下文
//   - policy: 查询策略，包含查询条件和分页参数
//
// 返回值：
//   - []*entity.Workflow: 工作流实体列表
//   - int64: 总数（需要 NeedTotalNumber=true）
//   - error: 查询失败时返回错误
func (r *RepositoryImpl) MGetDrafts(ctx context.Context, policy *vo.MGetPolicy) (_ []*entity.Workflow, totalCount int64, err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	q := policy.MetaQuery
	if len(q.IDs) == 0 && q.Page == nil && q.Name == nil && q.AppID == nil {
		return nil, 0, vo.WrapError(errno.ErrInternalBadRequest,
			fmt.Errorf("insufficient query parameters for workflow draft: %+v", q),
			errorx.KV("scene", "query workflow drafts"))
	}

	var (
		conditions []gen.Condition
	)
	if len(q.IDs) > 0 {
		conditions = append(conditions, r.query.WorkflowDraft.ID.In(q.IDs...))
	}

	if q.Name != nil {
		conditions = append(conditions, r.query.WorkflowMeta.Name.Like(`%%`+*q.Name+`%%`))
	}

	if q.SpaceID != nil {
		conditions = append(conditions, r.query.WorkflowMeta.SpaceID.Eq(*q.SpaceID))
	}

	if q.PublishStatus != nil {
		if *q.PublishStatus == vo.HasPublished {
			conditions = append(conditions, r.query.WorkflowMeta.Status.Eq(1))
		} else {
			conditions = append(conditions, r.query.WorkflowMeta.Status.Eq(0))
		}
	}

	if q.AppID != nil {
		conditions = append(conditions, r.query.WorkflowMeta.AppID.Eq(*q.AppID))
	}

	if q.LibOnly {
		conditions = append(conditions, r.query.WorkflowMeta.AppID.Eq(0))
	}

	if q.Mode != nil {
		conditions = append(conditions, r.query.WorkflowMeta.Mode.Eq(int32(*q.Mode)))
	}

	type combinedDraft struct {
		model.WorkflowDraft
		Name          string `gorm:"column:name"`
		Description   string `gorm:"column:description"`
		AppID         int64  `gorm:"column:app_id"`
		Status        int32  `gorm:"column:status"`
		SpaceID       int64  `gorm:"column:space_id"`
		IconURI       string `gorm:"column:icon_uri"`
		ContentType   int32  `gorm:"column:content_type"`
		Mode          int32  `gorm:"column:mode"`
		CreatedAt     int64  `gorm:"column:created_at"`
		CreatorID     int64  `gorm:"column:creator_id"`
		Tag           int32  `gorm:"column:tag"`
		LatestVersion string `gorm:"column:latest_version"`
	}

	selectColumns := r.query.WorkflowDraft.Columns(r.query.WorkflowDraft.ALL)
	selectColumns = append(selectColumns, r.query.WorkflowMeta.Name.As("name"),
		r.query.WorkflowMeta.Description.As("description"),
		r.query.WorkflowMeta.AppID.As("app_id"),
		r.query.WorkflowMeta.Status.As("status"),
		r.query.WorkflowMeta.SpaceID.As("space_id"),
		r.query.WorkflowMeta.IconURI.As("icon_uri"),
		r.query.WorkflowMeta.ContentType.As("content_type"),
		r.query.WorkflowMeta.Mode.As("mode"),
		r.query.WorkflowMeta.CreatedAt.As("created_at"),
		r.query.WorkflowMeta.CreatorID.As("creator_id"),
		r.query.WorkflowMeta.Tag.As("tag"),
		r.query.WorkflowMeta.LatestVersion.As("latest_version"),
	)

	d := r.query.WorkflowDraft.Debug().WithContext(ctx).
		Join(r.query.WorkflowMeta, r.query.WorkflowDraft.ID.EqCol(r.query.WorkflowMeta.ID)).
		Select(selectColumns...).
		Where(conditions...)

	if q.NeedTotalNumber {
		totalCount, err = d.Count()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get workflow draft count for policy %+v: %w", policy, err)
		}
	}

	if q.DescByUpdate {
		d = d.Order(r.query.WorkflowDraft.UpdatedAt.Desc())
	} else {
		d = d.Order(r.query.WorkflowMeta.CreatedAt.Desc())
	}

	var combinedDrafts []combinedDraft
	if q.Page != nil {
		_, err = d.ScanByPage(&combinedDrafts, q.Page.Offset(), q.Page.Limit())
	} else {
		err = d.Scan(&combinedDrafts)
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to get workflow draft for policy %+v: %w", policy, err)
	}

	result := make([]*entity.Workflow, len(combinedDrafts))
	for i, draft := range combinedDrafts {
		url, err := r.tos.GetObjectUrl(ctx, draft.IconURI)
		if err != nil {
			logs.Warnf("failed to get url for workflow meta %v", err)
		}

		canvasInfo := &vo.CanvasInfo{
			Canvas:          draft.Canvas,
			InputParamsStr:  draft.InputParams,
			OutputParamsStr: draft.OutputParams,
		}
		if err = canvasInfo.Unmarshal(); err != nil {
			return nil, 0, vo.WrapError(errno.ErrSerializationDeserializationFail, err)
		}

		wf := &entity.Workflow{
			ID:       draft.ID,
			CommitID: draft.CommitID,
			Meta: &vo.Meta{
				SpaceID:     draft.SpaceID,
				CreatorID:   draft.CreatorID,
				CreatedAt:   time.UnixMilli(draft.CreatedAt),
				ContentType: entity.ContentType(draft.ContentType),
				Name:        draft.Name,
				Desc:        draft.Description,
				IconURI:     draft.IconURI,
				IconURL:     url,
				Mode:        entity.Mode(draft.Mode),
			},
			DraftMeta: &vo.DraftMeta{
				TestRunSuccess: draft.TestRunSuccess,
				Modified:       draft.Modified,
				Timestamp:      time.UnixMilli(draft.UpdatedAt),
				IsSnapshot:     false,
			},
			CanvasInfo: canvasInfo,
		}

		if draft.Tag != 0 {
			wf.Meta.Tag = ptr.Of(entity.Tag(draft.Tag))
		}
		if draft.AppID != 0 {
			wf.Meta.AppID = &draft.AppID
		}
		if draft.Status > 0 {
			wf.Meta.HasPublished = true
		}
		if draft.LatestVersion != "" {
			wf.Meta.LatestPublishedVersion = &draft.LatestVersion
		}

		result[i] = wf
	}

	return result, totalCount, nil
}

// MGetLatestVersion 批量获取工作流最新版本
//
// 支持按 ID 列表、名称、空间 ID、应用 ID 等条件查询，并支持分页。
//
// 参数：
//   - ctx: 上下文
//   - policy: 查询策略，包含查询条件和分页参数
//
// 返回值：
//   - []*entity.Workflow: 工作流实体列表
//   - int64: 总数（需要 NeedTotalNumber=true）
//   - error: 查询失败时返回错误
func (r *RepositoryImpl) MGetLatestVersion(ctx context.Context, policy *vo.MGetPolicy) (
	_ []*entity.Workflow, totalCount int64, err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	q := policy.MetaQuery
	if len(q.IDs) == 0 && q.Page == nil && q.Name == nil && q.AppID == nil {
		return nil, 0, vo.WrapError(errno.ErrInternalBadRequest,
			fmt.Errorf("insufficient query parameters for workflow latest versions: %+v", q),
			errorx.KV("scene", "query latest workflow version"))
	}

	var (
		conditions []gen.Condition
	)
	if len(q.IDs) > 0 {
		conditions = append(conditions, r.query.WorkflowVersion.WorkflowID.In(q.IDs...))
	}

	if q.Name != nil {
		conditions = append(conditions, r.query.WorkflowMeta.Name.Like(`%%`+*q.Name+`%%`))
	}

	if q.SpaceID != nil {
		conditions = append(conditions, r.query.WorkflowMeta.SpaceID.Eq(*q.SpaceID))
	}

	if q.PublishStatus != nil {
		if *q.PublishStatus == vo.HasPublished {
			conditions = append(conditions, r.query.WorkflowMeta.Status.Eq(1))
		} else {
			conditions = append(conditions, r.query.WorkflowMeta.Status.Eq(0))
		}
	}

	if q.AppID != nil {
		conditions = append(conditions, r.query.WorkflowMeta.AppID.Eq(*q.AppID))
	}

	if q.LibOnly {
		conditions = append(conditions, r.query.WorkflowMeta.AppID.Eq(0))
	}

	if q.Mode != nil {
		conditions = append(conditions, r.query.WorkflowMeta.Mode.Eq(int32(*q.Mode)))
	}

	type combinedVersion struct {
		model.WorkflowMeta
		Version            string `gorm:"column:version"`             // release version
		VersionDescription string `gorm:"column:version_description"` // version description
		Canvas             string `gorm:"column:canvas"`              // Front-end schema
		InputParams        string `gorm:"column:input_params"`
		OutputParams       string `gorm:"column:output_params"`
		VersionCreatorID   int64  `gorm:"column:version_creator_id"` // Publish user ID
		VersionCreatedAt   int64  `gorm:"column:version_created_at"` // Creation time millisecond timestamp
		CommitID           string `gorm:"column:commit_id"`          // the commit id corresponding to this version
	}

	selectColumns := r.query.WorkflowMeta.Columns(r.query.WorkflowMeta.ALL)
	selectColumns = append(selectColumns, r.query.WorkflowVersion.Version.As("version"),
		r.query.WorkflowVersion.VersionDescription.As("version_description"),
		r.query.WorkflowVersion.Canvas.As("canvas"),
		r.query.WorkflowVersion.InputParams.As("input_params"),
		r.query.WorkflowVersion.OutputParams.As("output_params"),
		r.query.WorkflowVersion.CreatorID.As("version_creator_id"),
		r.query.WorkflowVersion.CreatedAt.As("version_created_at"),
		r.query.WorkflowVersion.CommitID.As("commit_id"),
	)

	d := r.query.WorkflowMeta.Debug().WithContext(ctx).
		Join(r.query.WorkflowVersion, r.query.WorkflowVersion.WorkflowID.EqCol(r.query.WorkflowMeta.ID),
			r.query.WorkflowVersion.Version.EqCol(r.query.WorkflowMeta.LatestVersion)).
		Select(selectColumns...).
		Where(conditions...)

	if q.NeedTotalNumber {
		totalCount, err = d.Count()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get workflow latest versions count for policy %+v: %w", policy, err)
		}
	}

	if q.DescByUpdate {
		d = d.Order(r.query.WorkflowMeta.LatestVersionTs.Desc())
	} else {
		d = d.Order(r.query.WorkflowMeta.LatestVersionTs.Asc())
	}

	var combinedVersions []combinedVersion
	if q.Page != nil {
		_, err = d.ScanByPage(&combinedVersions, q.Page.Offset(), q.Page.Limit())
	} else {
		err = d.Scan(&combinedVersions)
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to get workflow latest versions for policy %+v: %w", policy, err)
	}

	result := make([]*entity.Workflow, len(combinedVersions))
	for i, version := range combinedVersions {
		url, err := r.tos.GetObjectUrl(ctx, version.IconURI)
		if err != nil {
			logs.Warnf("failed to get url for workflow meta %v", err)
		}

		canvasInfo := &vo.CanvasInfo{
			Canvas:          version.Canvas,
			InputParamsStr:  version.InputParams,
			OutputParamsStr: version.OutputParams,
		}
		if err = canvasInfo.Unmarshal(); err != nil {
			return nil, 0, vo.WrapError(errno.ErrSerializationDeserializationFail, err)
		}

		wf := &entity.Workflow{
			ID:       version.ID,
			CommitID: version.CommitID,
			Meta: &vo.Meta{
				SpaceID:     version.SpaceID,
				CreatorID:   version.CreatorID,
				CreatedAt:   time.UnixMilli(version.CreatedAt),
				ContentType: entity.ContentType(version.ContentType),
				Name:        version.Name,
				Desc:        version.Description,
				IconURI:     version.IconURI,
				IconURL:     url,
				Mode:        entity.Mode(version.Mode),
			},
			VersionMeta: &vo.VersionMeta{
				Version:            version.Version,
				VersionDescription: version.VersionDescription,
				VersionCreatedAt:   time.UnixMilli(version.VersionCreatedAt),
				VersionCreatorID:   version.VersionCreatorID,
			},
			CanvasInfo: canvasInfo,
		}

		if version.Tag != 0 {
			wf.Meta.Tag = ptr.Of(entity.Tag(version.Tag))
		}
		if version.AppID != 0 {
			wf.Meta.AppID = &version.AppID
		}
		if version.Status > 0 {
			wf.Meta.HasPublished = true
		}
		if version.LatestVersion != "" {
			wf.Meta.LatestPublishedVersion = &version.LatestVersion
		}

		result[i] = wf
	}

	return result, totalCount, nil
}

// MGetReferences 批量获取工作流引用关系
//
// 参数：
//   - ctx: 上下文
//   - policy: 查询策略，包含被引用 ID、引用方 ID、引用类型等
//
// 返回值：
//   - []*entity.WorkflowReference: 引用关系列表
//   - error: 查询失败时返回错误
func (r *RepositoryImpl) MGetReferences(ctx context.Context, policy *vo.MGetReferencePolicy) (
	_ []*entity.WorkflowReference, err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	if len(policy.ReferredIDs) == 0 {
		return nil, vo.WrapError(errno.ErrInternalBadRequest, errors.New("referred IDs cannot be empty when querying references"))
	}

	var conds []gen.Condition
	if len(policy.ReferredIDs) == 1 {
		conds = append(conds, r.query.WorkflowReference.ReferredID.Eq(policy.ReferredIDs[0]))
	} else {
		conds = append(conds, r.query.WorkflowReference.ReferredID.In(policy.ReferredIDs...))
	}

	if len(policy.ReferringIDs) == 1 {
		conds = append(conds, r.query.WorkflowReference.ReferringID.Eq(policy.ReferringIDs[0]))
	} else if len(policy.ReferringIDs) > 1 {
		conds = append(conds, r.query.WorkflowReference.ReferringID.In(policy.ReferringIDs...))
	}

	if len(policy.ReferType) == 1 {
		conds = append(conds, r.query.WorkflowReference.ReferType.Eq(int32(policy.ReferType[0])))
	} else if len(policy.ReferType) > 1 {
		conds = append(conds, r.query.WorkflowReference.ReferType.In(
			slices.Transform(policy.ReferType, func(r vo.ReferType) int32 {
				return int32(r)
			})...))
	}

	if len(policy.ReferringBizType) == 1 {
		conds = append(conds, r.query.WorkflowReference.ReferringBizType.Eq(int32(policy.ReferringBizType[0])))
	} else if len(policy.ReferringBizType) > 1 {
		conds = append(conds, r.query.WorkflowReference.ReferringBizType.In(
			slices.Transform(policy.ReferringBizType, func(r vo.ReferringBizType) int32 {
				return int32(r)
			})...))
	}

	conds = append(conds, r.query.WorkflowReference.Status.Eq(1))

	refs, err := r.query.WorkflowReference.WithContext(ctx).Where(conds...).Find()
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow references: %w", err)
	}

	result := make([]*entity.WorkflowReference, 0, len(refs))
	for _, ref := range refs {
		result = append(result, &entity.WorkflowReference{
			ID: ref.ID,
			WorkflowReferenceKey: entity.WorkflowReferenceKey{
				ReferredID:       ref.ReferredID,
				ReferringID:      ref.ReferringID,
				ReferType:        vo.ReferType(ref.ReferType),
				ReferringBizType: vo.ReferringBizType(ref.ReferringBizType),
			},
			CreatedAt: time.UnixMilli(ref.CreatedAt),
			Enabled:   ref.Status == 1,
		})
	}

	return result, nil
}

// MGetMetas 批量获取工作流元数据
//
// 参数：
//   - ctx: 上下文
//   - query: 查询条件，包含 ID 列表、名称、空间 ID 等
//
// 返回值：
//   - map[int64]*vo.Meta: ID 到元数据的映射
//   - int64: 总数（需要 NeedTotalNumber=true）
//   - error: 查询失败时返回错误
func (r *RepositoryImpl) MGetMetas(ctx context.Context, query *vo.MetaQuery) (
	_ map[int64]*vo.Meta, _ int64, err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	if len(query.IDs) == 0 && query.Page == nil && query.Name == nil && query.AppID == nil {
		return nil, 0, vo.WrapError(errno.ErrInternalBadRequest,
			fmt.Errorf("insufficient query parameters for workflow meta: %+v", query),
			errorx.KV("scene", "query workflow metas"))
	}

	var conditions []gen.Condition
	if len(query.IDs) > 0 {
		conditions = append(conditions, r.query.WorkflowMeta.ID.In(query.IDs...))
	}

	if query.Name != nil {
		conditions = append(conditions, r.query.WorkflowMeta.Name.Like(`%%`+*query.Name+`%%`))
	}

	if query.SpaceID != nil {
		conditions = append(conditions, r.query.WorkflowMeta.SpaceID.Eq(*query.SpaceID))
	}

	if query.PublishStatus != nil {
		if *query.PublishStatus == vo.HasPublished {
			conditions = append(conditions, r.query.WorkflowMeta.Status.Eq(1))
		} else {
			conditions = append(conditions, r.query.WorkflowMeta.Status.Eq(0))
		}
	}

	if query.AppID != nil {
		conditions = append(conditions, r.query.WorkflowMeta.AppID.Eq(*query.AppID))
	}

	if query.LibOnly { // if AppID not specified, we can only query those within Library
		conditions = append(conditions, r.query.WorkflowMeta.AppID.Eq(0))
	}

	if query.Mode != nil {
		conditions = append(conditions, r.query.WorkflowMeta.Mode.Eq(int32(*query.Mode)))
	}

	var result []*model.WorkflowMeta

	workflowMetaDo := r.query.WorkflowMeta.WithContext(ctx).Debug().Where(conditions...)

	var total int64
	if query.NeedTotalNumber { // this is the total count
		total, err = workflowMetaDo.Count()
		if err != nil {
			return nil, 0, err
		}
	}

	if query.DescByUpdate {
		workflowMetaDo = workflowMetaDo.Order(r.query.WorkflowMeta.UpdatedAt.Desc())
	} else {
		workflowMetaDo = workflowMetaDo.Order(r.query.WorkflowMeta.CreatedAt.Desc())
	}

	if query.Page != nil {
		result, _, err = workflowMetaDo.FindByPage(query.Page.Offset(), query.Page.Limit())
		if err != nil {
			return nil, 0, err
		}
	} else {
		if len(conditions) == 0 {
			return nil, 0, errors.New("no conditions provided")
		}
		result, err = workflowMetaDo.Find()
		if err != nil {
			return nil, 0, err
		}
	}

	wfMap := make(map[int64]*vo.Meta, len(result))
	for _, meta := range result {
		converted, err := r.convertMeta(ctx, meta)
		if err != nil {
			return nil, 0, err
		}
		wfMap[meta.ID] = converted
	}
	return wfMap, total, nil
}

// GetLatestVersion 获取工作流最新版本
//
// 参数：
//   - ctx: 上下文
//   - id: 工作流 ID
//
// 返回值：
//   - *vo.VersionInfo: 版本信息
//   - error: 获取失败时返回错误
func (r *RepositoryImpl) GetLatestVersion(ctx context.Context, id int64) (*vo.VersionInfo, error) {
	version, err := r.query.WorkflowVersion.WithContext(ctx).Where(r.query.WorkflowVersion.WorkflowID.Eq(id)).
		Order(r.query.WorkflowVersion.CreatedAt.Desc()).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, vo.WrapError(errno.ErrWorkflowNotFound,
				fmt.Errorf("workflow version not found for ID %d: %w", id, err),
				errorx.KV("id", strconv.FormatInt(id, 10)))
		}
		return nil, fmt.Errorf("failed to query workflow version for ID %d: %w", id, err)
	}
	return &vo.VersionInfo{
		VersionMeta: &vo.VersionMeta{
			Version:            version.Version,
			VersionDescription: version.VersionDescription,
			VersionCreatedAt:   time.UnixMilli(version.CreatedAt),
			VersionCreatorID:   version.CreatorID,
		},
		CanvasInfo: vo.CanvasInfo{
			Canvas:          version.Canvas,
			InputParamsStr:  version.InputParams,
			OutputParamsStr: version.OutputParams,
		},
	}, nil
}

// CreateSnapshotIfNeeded 按需创建工作流快照
//
// 在执行工作流前调用，确保当前草稿有对应的快照可用于历史查看。
// 如果快照已存在，则不重复创建。
//
// 参数：
//   - ctx: 上下文
//   - id: 工作流 ID
//   - commitID: 提交 ID
//
// 返回值：
//   - error: 创建失败时返回错误
func (r *RepositoryImpl) CreateSnapshotIfNeeded(ctx context.Context, id int64, commitID string) error {
	latestSnapshot, err := r.query.WorkflowSnapshot.WithContext(ctx).Where(
		r.query.WorkflowSnapshot.WorkflowID.Eq(id),
		r.query.WorkflowSnapshot.CommitID.Eq(commitID),
	).First()

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logs.CtxErrorf(ctx, "query workflow snapshot failed err=%v", err)
		}
	} else if latestSnapshot != nil { // already have this snapshot, no need to create it
		return nil
	}

	draft, err := r.query.WorkflowDraft.WithContext(ctx).Where(
		r.query.WorkflowDraft.ID.Eq(id),
		r.query.WorkflowDraft.CommitID.Eq(commitID),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return vo.WrapError(errno.ErrWorkflowNotFound,
				fmt.Errorf("workflow draft not found for ID %d, commitID %s: %w", id, commitID, err),
				errorx.KV("id", strconv.FormatInt(id, 10)))
		}
		return vo.WrapError(errno.ErrDatabaseError,
			fmt.Errorf("failed to query workflow draft for ID %d, commitID %s: %w", id, commitID, err))
	}

	return r.query.WorkflowSnapshot.WithContext(ctx).Save(&model.WorkflowSnapshot{
		// ID: auto_increment
		WorkflowID:   id,
		CommitID:     commitID,
		Canvas:       draft.Canvas,
		InputParams:  draft.InputParams,
		OutputParams: draft.OutputParams,
	})
}

// WorkflowAsTool 将工作流转换为工具
//
// 该方法将工作流编译为可被其他工作流或 Agent 调用的工具。
// 支持配置输入输出参数的禁用和默认值。
//
// 参数：
//   - ctx: 上下文
//   - policy: 工作流获取策略
//   - wfToolConfig: 工具配置，包含输入输出参数配置
//
// 返回值：
//   - workflow.ToolFromWorkflow: 工具接口实现
//   - error: 转换失败时返回错误
func (r *RepositoryImpl) WorkflowAsTool(ctx context.Context, policy vo.GetPolicy, wfToolConfig vo.WorkflowToolConfig) (workflow.ToolFromWorkflow, error) {
	var (
		canvas               vo.Canvas
		inputParamsCfg       = wfToolConfig.InputParametersConfig
		outputParamsCfg      = wfToolConfig.OutputParametersConfig
		inputParamsConfigMap = slices.ToMap(inputParamsCfg, func(w *workflow3.APIParameter) (string, *workflow3.APIParameter) {
			return w.Name, w
		})
	)

	wfEntity, err := r.GetEntity(ctx, &policy)
	if err != nil {
		return nil, err
	}

	if err = sonic.UnmarshalString(wfEntity.Canvas, &canvas); err != nil {
		return nil, vo.WrapError(errno.ErrSerializationDeserializationFail, err)
	}

	name := fmt.Sprintf("ts_%s_%s", wfEntity.Name, wfEntity.Name)
	desc := wfEntity.Desc

	var params map[string]*schema.ParameterInfo

	for _, tInfo := range wfEntity.InputParams {
		if p, ok := inputParamsConfigMap[tInfo.Name]; ok && p.LocalDisable {
			continue
		}
		param, err := tInfo.ToParameterInfo()
		if err != nil {
			return nil, err
		}

		if params == nil {
			params = make(map[string]*schema.ParameterInfo)
		}
		params[tInfo.Name] = param
	}

	toolInfo := &schema.ToolInfo{
		Name:        name,
		Desc:        desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(params),
	}

	workflowSC, err := adaptor.CanvasToWorkflowSchema(ctx, &canvas)
	if err != nil {
		return nil, vo.WrapError(errno.ErrSchemaConversionFail, err)
	}

	var opts []compose.WorkflowOption
	opts = append(opts, compose.WithIDAsName(policy.ID),
		compose.WithParentRequireCheckpoint()) // always assumes the 'parent' may pass a checkpoint ID
	if s := execute.GetStaticConfig(); s != nil && s.MaxNodeCountPerWorkflow > 0 {
		opts = append(opts, compose.WithMaxNodeCount(s.MaxNodeCountPerWorkflow))
	}

	wf, err := compose.NewWorkflow(ctx, workflowSC, opts...)
	if err != nil {
		return nil, vo.WrapError(errno.ErrWorkflowCompileFail, err)
	}

	type streamFunc func(ctx context.Context, in map[string]any, opts ...einoCompose.Option) (*schema.StreamReader[map[string]any], error)

	if wf.StreamRun() {
		convertStream := func(stream streamFunc) streamFunc {
			return func(ctx context.Context, in map[string]any, opts ...einoCompose.Option) (*schema.StreamReader[map[string]any], error) {
				if len(inputParamsConfigMap) == 0 {
					return stream(ctx, in, opts...)
				}
				input := make(map[string]any, len(in))
				for k, v := range in {
					if p, ok := inputParamsConfigMap[k]; ok {
						if p.LocalDisable {
							if p.LocalDefault != nil {
								input[k], err = transformDefaultValue(*p.LocalDefault, p)
								if err != nil {
									return nil, err
								}
							}
						} else {
							input[k] = v
						}

					} else {
						input[k] = v
					}
				}
				return stream(ctx, input, opts...)
			}
		}
		return compose.NewStreamableWorkflow(
			toolInfo,
			convertStream(wf.Runner.Stream),
			wf.TerminatePlan(),
			wfEntity,
			workflowSC,
			r,
		), nil
	}

	type invokeFunc func(ctx context.Context, in map[string]any, opts ...einoCompose.Option) (out map[string]any, err error)
	convertInvoke := func(invoke invokeFunc) invokeFunc {
		return func(ctx context.Context, in map[string]any, opts ...einoCompose.Option) (out map[string]any, err error) {
			if len(inputParamsCfg) == 0 && len(outputParamsCfg) == 0 {
				return invoke(ctx, in, opts...)
			}
			input := make(map[string]any, len(in))
			for k, v := range in {
				if p, ok := inputParamsConfigMap[k]; ok {
					if p.LocalDisable {
						if p.LocalDefault != nil {
							input[k], err = transformDefaultValue(*p.LocalDefault, p)
							if err != nil {
								return nil, fmt.Errorf("failed to transfer default value, default value=%v,value type=%v,err=%w", *p.LocalDefault, p.Type, err)
							}
						}
					} else {
						input[k] = v
					}
				} else {
					input[k] = v
				}
			}

			out, err = invoke(ctx, input, opts...)
			if err != nil {
				return nil, err
			}

			if wf.TerminatePlan() == vo.ReturnVariables && len(outputParamsCfg) > 0 {
				return filterDisabledAPIParameters(outputParamsCfg, out), nil
			}

			return out, nil

		}
	}

	return compose.NewInvokableWorkflow(
		toolInfo,
		convertInvoke(wf.Runner.Invoke),
		wf.TerminatePlan(),
		wfEntity,
		workflowSC,
		r,
	), nil
}

// CopyWorkflow 复制工作流
//
// 创建工作流的副本，支持：
//   - 自动生成副本名称（原名称_序号）
//   - 复制到不同空间或应用
//   - 修改画布内容
//
// 参数：
//   - ctx: 上下文
//   - workflowID: 源工作流 ID
//   - policy: 复制策略，包含目标空间、应用等配置
//
// 返回值：
//   - *entity.Workflow: 复制后的工作流实体
//   - error: 复制失败时返回错误
func (r *RepositoryImpl) CopyWorkflow(ctx context.Context, workflowID int64, policy vo.CopyWorkflowPolicy) (
	_ *entity.Workflow, err error) {
	const (
		copyWorkflowRedisKeyPrefix         = "copy_workflow_redis_key_prefix"
		copyWorkflowRedisKeyExpireInterval = time.Hour * 24 * 7
	)

	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	var (
		copiedID      int64
		workflowMeta  = r.query.WorkflowMeta
		workflowDraft = r.query.WorkflowDraft
	)

	copiedID, err = r.IDGenerator.GenID(ctx)
	if err != nil {
		return nil, vo.WrapError(errno.ErrIDGenError, err)
	}

	var copiedWorkflow *entity.Workflow
	wfMeta, err := workflowMeta.WithContext(ctx).Where(workflowMeta.ID.Eq(workflowID)).First()
	if err != nil {
		return nil, err
	}

	wfDraft, err := workflowDraft.WithContext(ctx).Where(workflowDraft.ID.Eq(workflowID)).First()
	if err != nil {
		return nil, err
	}

	commitID, err := r.IDGenerator.GenID(ctx)
	if err != nil {
		return nil, err
	}

	var copiedWorkflowName string
	if policy.ShouldModifyWorkflowName {
		copiedWorkflowRedisKey := fmt.Sprintf("%s:%d:%d", copyWorkflowRedisKeyPrefix, workflowID, ctxutil.MustGetUIDFromCtx(ctx))
		copiedNameSuffix, err := r.redis.Incr(ctx, copiedWorkflowRedisKey).Result()
		if err != nil {
			return nil, vo.WrapError(errno.ErrRedisError, err)
		}
		err = r.redis.Expire(ctx, copiedWorkflowRedisKey, copyWorkflowRedisKeyExpireInterval).Err()
		if err != nil {
			logs.Warnf("failed to set the rediskey %v expiration time, err=%v", copiedWorkflowRedisKey, err)
		}
		copiedWorkflowName = fmt.Sprintf("%s_%d", wfMeta.Name, copiedNameSuffix)
	} else {
		copiedWorkflowName = wfMeta.Name
	}

	err = r.query.Transaction(func(tx *query.Query) error {
		wfMeta.Name = copiedWorkflowName
		wfMeta.SourceID = workflowID
		wfMeta.Status = 0
		wfMeta.ID = copiedID
		wfMeta.CreatedAt = 0
		wfMeta.UpdatedAt = 0
		wfMeta.LatestVersion = ""
		if policy.TargetSpaceID != nil {
			wfMeta.SpaceID = *policy.TargetSpaceID
		}
		if policy.TargetAppID != nil {
			wfMeta.AppID = *policy.TargetAppID
		}
		wfMeta.CreatorID = ctxutil.MustGetUIDFromCtx(ctx)
		err = workflowMeta.WithContext(ctx).Create(wfMeta)
		if err != nil {
			return err
		}

		wfDraft.ID = copiedID
		// copy workflow are treated as modified and not tested run
		wfDraft.TestRunSuccess = false
		wfDraft.Modified = true
		wfDraft.UpdatedAt = 0
		wfDraft.CommitID = strconv.FormatInt(commitID, 10)
		if policy.ModifiedCanvasSchema != nil {
			wfDraft.Canvas = *policy.ModifiedCanvasSchema
		}
		err = workflowDraft.WithContext(ctx).Create(wfDraft)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err

	}

	copiedWorkflow = &entity.Workflow{
		ID:       copiedID,
		CommitID: wfDraft.CommitID,
		Meta: &vo.Meta{
			SpaceID:   wfMeta.SpaceID,
			Name:      wfMeta.Name,
			CreatorID: wfMeta.CreatorID,
			IconURI:   wfMeta.IconURI,
			Desc:      wfMeta.Description,
			AppID:     ternary.IFElse(wfMeta.AppID == 0, (*int64)(nil), ptr.Of(wfMeta.AppID)),
			Mode:      workflowModel.WorkflowMode(wfMeta.Mode),
		},
		CanvasInfo: &vo.CanvasInfo{
			Canvas:          wfDraft.Canvas,
			InputParamsStr:  wfDraft.InputParams,
			OutputParamsStr: wfDraft.OutputParams,
		},
	}

	return copiedWorkflow, nil
}

// GetDraftWorkflowsByAppID 根据应用 ID 获取所有草稿工作流
//
// 参数：
//   - ctx: 上下文
//   - AppID: 应用 ID
//
// 返回值：
//   - map[int64]*vo.DraftInfo: 工作流 ID 到草稿信息的映射
//   - map[int64]string: 工作流 ID 到名称的映射
//   - error: 查询失败时返回错误
func (r *RepositoryImpl) GetDraftWorkflowsByAppID(ctx context.Context, AppID int64) (
	_ map[int64]*vo.DraftInfo, _ map[int64]string, err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrDatabaseError, err)
		}
	}()

	var (
		workflowMeta  = r.query.WorkflowMeta
		workflowDraft = r.query.WorkflowDraft
	)

	wfMetas, err := workflowMeta.WithContext(ctx).Where(workflowMeta.AppID.Eq(AppID)).Find()
	if err != nil {
		return nil, nil, err
	}
	draftIDs := slices.Transform(wfMetas, func(a *model.WorkflowMeta) int64 {
		return a.ID
	})

	wfDrafts, err := workflowDraft.WithContext(ctx).Where(workflowDraft.ID.In(draftIDs...)).Find()
	if err != nil {
		return nil, nil, err
	}
	result := make(map[int64]*vo.DraftInfo, len(wfDrafts))
	for _, d := range wfDrafts {
		result[d.ID] = &vo.DraftInfo{
			Canvas:          d.Canvas,
			InputParamsStr:  d.InputParams,
			OutputParamsStr: d.OutputParams,
		}
	}

	wid2Named := slices.ToMap(wfMetas, func(e *model.WorkflowMeta) (int64, string) {
		return e.ID, e.Name
	})
	return result, wid2Named, nil
}

// BatchCreateConnectorWorkflowVersion 批量创建连接器工作流版本关联
//
// 参数：
//   - ctx: 上下文
//   - appID: 应用 ID
//   - connectorID: 连接器 ID
//   - workflowIDs: 工作流 ID 列表
//   - version: 版本号
//
// 返回值：
//   - error: 创建失败时返回错误
func (r *RepositoryImpl) BatchCreateConnectorWorkflowVersion(ctx context.Context, appID, connectorID int64, workflowIDs []int64, version string) error {
	objects := make([]*model.ConnectorWorkflowVersion, 0, len(workflowIDs))
	for idx := range workflowIDs {
		workflowID := workflowIDs[idx]
		objects = append(objects, &model.ConnectorWorkflowVersion{
			AppID:       appID,
			ConnectorID: connectorID,
			Version:     version,
			WorkflowID:  workflowID,
		})
	}
	err := r.query.ConnectorWorkflowVersion.WithContext(ctx).CreateInBatches(objects, batchCreateSize)
	if err != nil {
		return vo.WrapError(errno.ErrDatabaseError, err)
	}

	return nil
}

// GetKnowledgeRecallChatModel 获取知识库召回使用的聊天模型
//
// 返回值：
//   - modelbuilder.BaseChatModel: 聊天模型实例
func (r *RepositoryImpl) GetKnowledgeRecallChatModel() modelbuilder.BaseChatModel {
	return r.builtinModel
}

// GetObjectUrl 获取对象存储资源的 URL
//
// 参数：
//   - ctx: 上下文
//   - objectKey: 对象键
//   - opts: 可选配置
//
// 返回值：
//   - string: 资源 URL
//   - error: 获取失败时返回错误
func (r *RepositoryImpl) GetObjectUrl(ctx context.Context, objectKey string, opts ...storage.GetOptFn) (string, error) {
	return r.tos.GetObjectUrl(ctx, objectKey, opts...)
}

// filterDisabledAPIParameters 过滤被禁用的 API 参数
//
// 递归处理嵌套的对象类型参数。
//
// 参数：
//   - parametersCfg: 参数配置列表
//   - m: 原始参数映射
//
// 返回值：
//   - map[string]any: 过滤后的参数映射
func filterDisabledAPIParameters(parametersCfg []*workflow3.APIParameter, m map[string]any) map[string]any {
	result := make(map[string]any, len(m))
	responseParameterMap := slices.ToMap(parametersCfg, func(p *workflow3.APIParameter) (string, *workflow3.APIParameter) {
		return p.Name, p
	})
	for key, value := range m {
		if parameter, ok := responseParameterMap[key]; ok {
			if parameter.LocalDisable {
				continue
			}
			if parameter.Type == workflow3.ParameterType_Object && len(parameter.SubParameters) > 0 {
				val := filterDisabledAPIParameters(parameter.SubParameters, value.(map[string]interface{}))
				result[key] = val
			} else {
				result[key] = value
			}
		} else {
			result[key] = value
		}
	}
	return result
}

// transformDefaultValue 将字符串默认值转换为实际类型
//
// 根据参数类型将字符串值转换为对应的 Go 类型。
//
// 参数：
//   - value: 字符串默认值
//   - p: 参数配置，包含类型信息
//
// 返回值：
//   - any: 转换后的值
//   - error: 转换失败时返回错误
func transformDefaultValue(value string, p *workflow3.APIParameter) (any, error) {
	switch p.Type {
	default:
		return value, nil
	case workflow3.ParameterType_String:
		return value, nil
	case workflow3.ParameterType_Object:
		ret := make(map[string]any)
		err := sonic.UnmarshalString(value, &ret)
		if err != nil {
			return nil, vo.WrapError(errno.ErrSerializationDeserializationFail, err)
		}
		return ret, nil
	case workflow3.ParameterType_Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return nil, vo.WrapError(errno.ErrSerializationDeserializationFail, err)
		}
		return b, nil
	case workflow3.ParameterType_Number:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, vo.WrapError(errno.ErrSerializationDeserializationFail, err)
		}
		return f, nil
	case workflow3.ParameterType_Integer:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, vo.WrapError(errno.ErrSerializationDeserializationFail, err)
		}
		return i, nil
	case workflow3.ParameterType_Array:
		ret := make([]any, 0)
		err := sonic.UnmarshalString(value, &ret)
		if err != nil {
			return nil, vo.WrapError(errno.ErrSerializationDeserializationFail, err)
		}
		return ret, nil
	}
}
