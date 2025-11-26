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

// conversation_repository.go 会话仓储扩展
//
// 本文件包含 RepositoryImpl 的会话相关方法实现，提供：
//   - 会话模板管理（草稿/在线环境）
//   - 静态会话管理（基于模板的固定会话）
//   - 动态会话管理（用户自定义会话）
//
// 会话是 ChatFlow 工作流的核心概念，用于维护多轮对话的上下文。

package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gen"
	"gorm.io/gorm"

	crossconversation "github.com/coze-dev/coze-studio/backend/crossdomain/conversation"
	"github.com/coze-dev/coze-studio/backend/domain/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/repo/dal/model"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/slices"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

// batchSize 批量操作的批次大小
const batchSize = 10

// CreateDraftConversationTemplate 创建草稿会话模板
//
// 会话模板定义了 ChatFlow 中的会话类型，如"默认会话"、"客服会话"等。
//
// 参数：
//   - ctx: 上下文
//   - template: 模板创建信息
//
// 返回值：
//   - int64: 创建的模板 ID
//   - error: 创建失败时返回错误
func (r *RepositoryImpl) CreateDraftConversationTemplate(ctx context.Context, template *vo.CreateConversationTemplateMeta) (int64, error) {
	id, err := r.GenID(ctx)
	if err != nil {
		return 0, vo.WrapError(errno.ErrIDGenError, err)
	}
	m := &model.AppConversationTemplateDraft{
		ID:         id,
		AppID:      template.AppID,
		SpaceID:    template.SpaceID,
		Name:       template.Name,
		CreatorID:  template.UserID,
		TemplateID: id,
	}
	err = r.query.AppConversationTemplateDraft.WithContext(ctx).Create(m)
	if err != nil {
		return 0, vo.WrapError(errno.ErrDatabaseError, err)
	}

	return id, nil
}

// GetConversationTemplate 获取会话模板
//
// 支持从草稿环境或在线环境获取。
//
// 参数：
//   - ctx: 上下文
//   - env: 环境（Draft/Online）
//   - policy: 查询策略
//
// 返回值：
//   - *entity.ConversationTemplate: 会话模板
//   - bool: 是否存在
//   - error: 获取失败时返回错误
func (r *RepositoryImpl) GetConversationTemplate(ctx context.Context, env vo.Env, policy vo.GetConversationTemplatePolicy) (*entity.ConversationTemplate, bool, error) {
	var (
		appID      = policy.AppID
		name       = policy.Name
		version    = policy.Version
		templateID = policy.TemplateID
	)

	conditions := make([]gen.Condition, 0)
	if env == vo.Draft {
		if appID != nil {
			conditions = append(conditions, r.query.AppConversationTemplateDraft.AppID.Eq(*appID))
		}
		if name != nil {
			conditions = append(conditions, r.query.AppConversationTemplateDraft.Name.Eq(*name))
		}
		if templateID != nil {
			conditions = append(conditions, r.query.AppConversationTemplateDraft.TemplateID.Eq(*templateID))
		}

		template, err := r.query.AppConversationTemplateDraft.WithContext(ctx).Where(conditions...).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, nil
			}
			return nil, false, vo.WrapError(errno.ErrDatabaseError, err)
		}
		return &entity.ConversationTemplate{
			AppID:      template.AppID,
			Name:       template.Name,
			TemplateID: template.TemplateID,
		}, true, nil

	} else if env == vo.Online {
		if policy.Version != nil {
			conditions = append(conditions, r.query.AppConversationTemplateOnline.Version.Eq(*version))
		}
		if appID != nil {
			conditions = append(conditions, r.query.AppConversationTemplateOnline.AppID.Eq(*appID))
		}
		if name != nil {
			conditions = append(conditions, r.query.AppConversationTemplateOnline.Name.Eq(*name))
		}

		if templateID != nil {
			conditions = append(conditions, r.query.AppConversationTemplateOnline.TemplateID.Eq(*templateID))
		}

		template, err := r.query.AppConversationTemplateOnline.WithContext(ctx).Where(conditions...).Order(r.query.AppConversationTemplateOnline.CreatedAt.Desc()).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, nil
			}
			return nil, false, err
		}
		return &entity.ConversationTemplate{
			AppID:      template.AppID,
			Name:       template.Name,
			TemplateID: template.TemplateID,
		}, true, nil
	}

	return nil, false, fmt.Errorf("unknown env %v", env)

}

// UpdateDraftConversationTemplateName 更新草稿会话模板名称
//
// 参数：
//   - ctx: 上下文
//   - templateID: 模板 ID
//   - name: 新名称
//
// 返回值：
//   - error: 更新失败时返回错误
func (r *RepositoryImpl) UpdateDraftConversationTemplateName(ctx context.Context, templateID int64, name string) error {
	_, err := r.query.AppConversationTemplateDraft.WithContext(ctx).Where(
		r.query.AppConversationTemplateDraft.TemplateID.Eq(templateID),
	).UpdateColumnSimple(r.query.AppConversationTemplateDraft.Name.Value(name))

	if err != nil {
		return vo.WrapError(errno.ErrDatabaseError, err)
	}
	return nil

}

// DeleteDraftConversationTemplate 删除草稿会话模板
//
// 参数：
//   - ctx: 上下文
//   - templateID: 模板 ID
//
// 返回值：
//   - int64: 影响的行数
//   - error: 删除失败时返回错误
func (r *RepositoryImpl) DeleteDraftConversationTemplate(ctx context.Context, templateID int64) (int64, error) {
	resultInfo, err := r.query.AppConversationTemplateDraft.WithContext(ctx).Where(
		r.query.AppConversationTemplateDraft.TemplateID.Eq(templateID),
	).Delete()

	if err != nil {
		return 0, vo.WrapError(errno.ErrDatabaseError, err)
	}
	return resultInfo.RowsAffected, nil

}

// DeleteDynamicConversation 删除动态会话
//
// 参数：
//   - ctx: 上下文
//   - env: 环境（Draft/Online）
//   - id: 会话 ID
//
// 返回值：
//   - int64: 影响的行数
//   - error: 删除失败时返回错误
func (r *RepositoryImpl) DeleteDynamicConversation(ctx context.Context, env vo.Env, id int64) (int64, error) {
	if env == vo.Draft {
		info, err := r.query.AppDynamicConversationDraft.WithContext(ctx).Where(r.query.AppDynamicConversationDraft.ID.Eq(id)).Delete()
		if err != nil {
			return 0, vo.WrapError(errno.ErrDatabaseError, err)
		}
		return info.RowsAffected, nil
	} else if env == vo.Online {
		info, err := r.query.AppDynamicConversationOnline.WithContext(ctx).Where(r.query.AppDynamicConversationOnline.ID.Eq(id)).Delete()
		if err != nil {
			return 0, vo.WrapError(errno.ErrDatabaseError, err)
		}
		return info.RowsAffected, nil
	} else {
		return 0, fmt.Errorf("unknown env %v", env)
	}
}

// ListConversationTemplate 列出会话模板
//
// 参数：
//   - ctx: 上下文
//   - env: 环境（Draft/Online）
//   - policy: 查询策略，包含分页和过滤条件
//
// 返回值：
//   - []*entity.ConversationTemplate: 模板列表
//   - error: 查询失败时返回错误
func (r *RepositoryImpl) ListConversationTemplate(ctx context.Context, env vo.Env, policy *vo.ListConversationTemplatePolicy) ([]*entity.ConversationTemplate, error) {
	if env == vo.Draft {
		return r.listDraftConversationTemplate(ctx, policy)
	} else if env == vo.Online {
		return r.listOnlineConversationTemplate(ctx, policy)
	} else {
		return nil, fmt.Errorf("unknown env %v", env)
	}
}

// listDraftConversationTemplate 列出草稿环境的会话模板
func (r *RepositoryImpl) listDraftConversationTemplate(ctx context.Context, policy *vo.ListConversationTemplatePolicy) ([]*entity.ConversationTemplate, error) {
	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, r.query.AppConversationTemplateDraft.AppID.Eq(policy.AppID))

	if policy.NameLike != nil {
		conditions = append(conditions, r.query.AppConversationTemplateDraft.Name.Like("%%"+*policy.NameLike+"%%"))
	}
	appConversationTemplateDraftDao := r.query.AppConversationTemplateDraft.WithContext(ctx)
	var (
		templates []*model.AppConversationTemplateDraft
		err       error
	)

	if policy.Page != nil {
		templates, err = appConversationTemplateDraftDao.Where(conditions...).Offset(policy.Page.Offset()).Limit(policy.Page.Limit()).Find()
	} else {
		templates, err = appConversationTemplateDraftDao.Where(conditions...).Find()

	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []*entity.ConversationTemplate{}, nil
		}
		return nil, vo.WrapError(errno.ErrDatabaseError, err)
	}

	return slices.Transform(templates, func(a *model.AppConversationTemplateDraft) *entity.ConversationTemplate {
		return &entity.ConversationTemplate{
			SpaceID:    a.SpaceID,
			AppID:      a.AppID,
			Name:       a.Name,
			TemplateID: a.TemplateID,
		}
	}), nil

}

// listOnlineConversationTemplate 列出在线环境的会话模板
func (r *RepositoryImpl) listOnlineConversationTemplate(ctx context.Context, policy *vo.ListConversationTemplatePolicy) ([]*entity.ConversationTemplate, error) {
	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, r.query.AppConversationTemplateOnline.AppID.Eq(policy.AppID))
	if policy.Version == nil {
		return nil, fmt.Errorf("list online template fail, version is required")
	}
	conditions = append(conditions, r.query.AppConversationTemplateOnline.Version.Eq(*policy.Version))

	if policy.NameLike != nil {
		conditions = append(conditions, r.query.AppConversationTemplateOnline.Name.Like("%%"+*policy.NameLike+"%%"))
	}
	appConversationTemplateOnlineDao := r.query.AppConversationTemplateOnline.WithContext(ctx)
	var (
		templates []*model.AppConversationTemplateOnline
		err       error
	)
	if policy.Page != nil {
		templates, err = appConversationTemplateOnlineDao.Where(conditions...).Offset(policy.Page.Offset()).Limit(policy.Page.Limit()).Find()

	} else {
		templates, err = appConversationTemplateOnlineDao.Where(conditions...).Find()

	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []*entity.ConversationTemplate{}, nil
		}
		return nil, vo.WrapError(errno.ErrDatabaseError, err)
	}

	return slices.Transform(templates, func(a *model.AppConversationTemplateOnline) *entity.ConversationTemplate {
		return &entity.ConversationTemplate{
			SpaceID:    a.SpaceID,
			AppID:      a.AppID,
			Name:       a.Name,
			TemplateID: a.TemplateID,
		}
	}), nil

}

// MGetStaticConversation 批量获取静态会话
//
// 静态会话是基于模板创建的固定会话，每个用户在每个连接器下对应一个。
//
// 参数：
//   - ctx: 上下文
//   - env: 环境（Draft/Online）
//   - userID: 用户 ID
//   - connectorID: 连接器 ID
//   - templateIDs: 模板 ID 列表
//
// 返回值：
//   - []*entity.StaticConversation: 静态会话列表
//   - error: 查询失败时返回错误
func (r *RepositoryImpl) MGetStaticConversation(ctx context.Context, env vo.Env, userID, connectorID int64, templateIDs []int64) ([]*entity.StaticConversation, error) {
	if env == vo.Draft {
		return r.mGetDraftStaticConversation(ctx, userID, connectorID, templateIDs)
	} else if env == vo.Online {
		return r.mGetOnlineStaticConversation(ctx, userID, connectorID, templateIDs)
	} else {
		return nil, fmt.Errorf("unknown env %v", env)
	}
}

// mGetDraftStaticConversation 批量获取草稿环境的静态会话
func (r *RepositoryImpl) mGetDraftStaticConversation(ctx context.Context, userID, connectorID int64, templateIDs []int64) ([]*entity.StaticConversation, error) {
	conditions := make([]gen.Condition, 0, 3)
	conditions = append(conditions, r.query.AppStaticConversationDraft.UserID.Eq(userID))
	conditions = append(conditions, r.query.AppStaticConversationDraft.ConnectorID.Eq(connectorID))
	if len(templateIDs) == 1 {
		conditions = append(conditions, r.query.AppStaticConversationDraft.TemplateID.Eq(templateIDs[0]))
	} else {
		conditions = append(conditions, r.query.AppStaticConversationDraft.TemplateID.In(templateIDs...))
	}

	cs, err := r.query.AppStaticConversationDraft.WithContext(ctx).Where(conditions...).Find()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []*entity.StaticConversation{}, nil
		}
		return nil, vo.WrapError(errno.ErrDatabaseError, err)
	}

	return slices.Transform(cs, func(a *model.AppStaticConversationDraft) *entity.StaticConversation {
		return &entity.StaticConversation{
			TemplateID:     a.TemplateID,
			ConversationID: a.ConversationID,
			UserID:         a.UserID,
			ConnectorID:    a.ConnectorID,
		}
	}), nil
}

// mGetOnlineStaticConversation 批量获取在线环境的静态会话
func (r *RepositoryImpl) mGetOnlineStaticConversation(ctx context.Context, userID, connectorID int64, templateIDs []int64) ([]*entity.StaticConversation, error) {
	conditions := make([]gen.Condition, 0, 3)
	conditions = append(conditions, r.query.AppStaticConversationOnline.UserID.Eq(userID))
	conditions = append(conditions, r.query.AppStaticConversationOnline.ConnectorID.Eq(connectorID))
	if len(templateIDs) == 1 {
		conditions = append(conditions, r.query.AppStaticConversationOnline.TemplateID.Eq(templateIDs[0]))
	} else {
		conditions = append(conditions, r.query.AppStaticConversationOnline.TemplateID.In(templateIDs...))
	}

	cs, err := r.query.AppStaticConversationOnline.WithContext(ctx).Where(conditions...).Find()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []*entity.StaticConversation{}, nil
		}
		return nil, vo.WrapError(errno.ErrDatabaseError, err)
	}

	return slices.Transform(cs, func(a *model.AppStaticConversationOnline) *entity.StaticConversation {
		return &entity.StaticConversation{
			TemplateID:     a.TemplateID,
			ConversationID: a.ConversationID,
		}
	}), nil
}

// ListDynamicConversation 列出动态会话
//
// 动态会话是用户自定义创建的会话，可以自由命名。
//
// 参数：
//   - ctx: 上下文
//   - env: 环境（Draft/Online）
//   - policy: 查询策略
//
// 返回值：
//   - []*entity.DynamicConversation: 动态会话列表
//   - error: 查询失败时返回错误
func (r *RepositoryImpl) ListDynamicConversation(ctx context.Context, env vo.Env, policy *vo.ListConversationPolicy) ([]*entity.DynamicConversation, error) {
	if env == vo.Draft {
		return r.listDraftDynamicConversation(ctx, policy)
	} else if env == vo.Online {
		return r.listOnlineDynamicConversation(ctx, policy)
	} else {
		return nil, fmt.Errorf("unknown env %v", env)
	}

}

// listDraftDynamicConversation 列出草稿环境的动态会话
func (r *RepositoryImpl) listDraftDynamicConversation(ctx context.Context, policy *vo.ListConversationPolicy) ([]*entity.DynamicConversation, error) {
	var (
		appID       = policy.APPID
		userID      = policy.UserID
		connectorID = policy.ConnectorID
	)

	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, r.query.AppDynamicConversationDraft.AppID.Eq(appID))
	conditions = append(conditions, r.query.AppDynamicConversationDraft.UserID.Eq(userID))
	conditions = append(conditions, r.query.AppDynamicConversationDraft.ConnectorID.Eq(connectorID))
	if policy.NameLike != nil {
		conditions = append(conditions, r.query.AppDynamicConversationDraft.Name.Like("%%"+*policy.NameLike+"%%"))
	}

	appDynamicConversationDraftDao := r.query.AppDynamicConversationDraft.WithContext(ctx).Where(conditions...)
	var (
		dynamicConversations = make([]*model.AppDynamicConversationDraft, 0)
		err                  error
	)

	if policy.Page != nil {
		dynamicConversations, err = appDynamicConversationDraftDao.Offset(policy.Page.Offset()).Limit(policy.Page.Limit()).Find()

	} else {
		dynamicConversations, err = appDynamicConversationDraftDao.Where(conditions...).Find()
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []*entity.DynamicConversation{}, nil
		}
		return nil, vo.WrapError(errno.ErrDatabaseError, err)
	}
	return slices.Transform(dynamicConversations, func(a *model.AppDynamicConversationDraft) *entity.DynamicConversation {
		return &entity.DynamicConversation{
			ID:             a.ID,
			Name:           a.Name,
			UserID:         a.UserID,
			ConnectorID:    a.ConnectorID,
			ConversationID: a.ConversationID,
		}
	}), nil
}

// listOnlineDynamicConversation 列出在线环境的动态会话
func (r *RepositoryImpl) listOnlineDynamicConversation(ctx context.Context, policy *vo.ListConversationPolicy) ([]*entity.DynamicConversation, error) {
	var (
		appID       = policy.APPID
		userID      = policy.UserID
		connectorID = policy.ConnectorID
	)

	conditions := make([]gen.Condition, 0)
	conditions = append(conditions, r.query.AppDynamicConversationOnline.AppID.Eq(appID))
	conditions = append(conditions, r.query.AppDynamicConversationOnline.UserID.Eq(userID))
	conditions = append(conditions, r.query.AppDynamicConversationOnline.AppID.Eq(appID))
	conditions = append(conditions, r.query.AppDynamicConversationOnline.ConnectorID.Eq(connectorID))
	if policy.NameLike != nil {
		conditions = append(conditions, r.query.AppDynamicConversationOnline.Name.Like("%%"+*policy.NameLike+"%%"))
	}

	appDynamicConversationOnlineDao := r.query.AppDynamicConversationOnline.WithContext(ctx).Where(conditions...)
	var (
		dynamicConversations = make([]*model.AppDynamicConversationOnline, 0)
		err                  error
	)
	if policy.Page != nil {
		dynamicConversations, err = appDynamicConversationOnlineDao.Offset(policy.Page.Offset()).Limit(policy.Page.Limit()).Find()
	} else {
		dynamicConversations, err = appDynamicConversationOnlineDao.Where(conditions...).Find()
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []*entity.DynamicConversation{}, nil
		}
		return nil, vo.WrapError(errno.ErrDatabaseError, err)
	}

	return slices.Transform(dynamicConversations, func(a *model.AppDynamicConversationOnline) *entity.DynamicConversation {
		return &entity.DynamicConversation{
			ID:             a.ID,
			Name:           a.Name,
			UserID:         a.UserID,
			ConnectorID:    a.ConnectorID,
			ConversationID: a.ConversationID,
		}
	}), nil
}

// GetOrCreateStaticConversation 获取或创建静态会话
//
// 如果会话已存在则返回现有会话，否则创建新会话。
//
// 参数：
//   - ctx: 上下文
//   - env: 环境（Draft/Online）
//   - idGen: 会话 ID 生成器
//   - meta: 创建信息
//
// 返回值：
//   - int64: 会话 ID
//   - int64: Section ID
//   - bool: 是否为已存在的会话
//   - error: 操作失败时返回错误
func (r *RepositoryImpl) GetOrCreateStaticConversation(ctx context.Context, env vo.Env, idGen workflow.ConversationIDGenerator, meta *vo.CreateStaticConversation) (int64, int64, bool, error) {
	if env == vo.Draft {
		return r.getOrCreateDraftStaticConversation(ctx, idGen, meta)
	} else if env == vo.Online {
		return r.getOrCreateOnlineStaticConversation(ctx, idGen, meta)
	} else {
		return 0, 0, false, fmt.Errorf("unknown env %v", env)
	}

}

// GetOrCreateDynamicConversation 获取或创建动态会话
//
// 如果同名会话已存在则返回现有会话，否则创建新会话。
//
// 参数：
//   - ctx: 上下文
//   - env: 环境（Draft/Online）
//   - idGen: 会话 ID 生成器
//   - meta: 创建信息
//
// 返回值：
//   - int64: 会话 ID
//   - int64: Section ID
//   - bool: 是否为已存在的会话
//   - error: 操作失败时返回错误
func (r *RepositoryImpl) GetOrCreateDynamicConversation(ctx context.Context, env vo.Env, idGen workflow.ConversationIDGenerator, meta *vo.CreateDynamicConversation) (int64, int64, bool, error) {
	if env == vo.Draft {

		appDynamicConversationDraft := r.query.AppDynamicConversationDraft
		ret, err := appDynamicConversationDraft.WithContext(ctx).Where(
			appDynamicConversationDraft.AppID.Eq(meta.BizID),
			appDynamicConversationDraft.ConnectorID.Eq(meta.ConnectorID),
			appDynamicConversationDraft.UserID.Eq(meta.UserID),
			appDynamicConversationDraft.Name.Eq(meta.Name),
		).First()
		if err == nil {
			cInfo, err := crossconversation.DefaultSVC().GetByID(ctx, ret.ConversationID)
			if err != nil {
				return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, err)
			}
			if cInfo == nil {
				return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("conversation not found"))
			}
			return ret.ConversationID, cInfo.SectionID, true, nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, err)
		}

		conv, err := idGen(ctx, meta.BizID, meta.UserID, meta.ConnectorID)
		if err != nil {
			return 0, 0, false, err
		}

		id, err := r.GenID(ctx)
		if err != nil {
			return 0, 0, false, vo.WrapError(errno.ErrIDGenError, err)
		}

		err = r.query.AppDynamicConversationDraft.WithContext(ctx).Create(&model.AppDynamicConversationDraft{
			ID:             id,
			AppID:          meta.BizID,
			Name:           meta.Name,
			UserID:         meta.UserID,
			ConnectorID:    meta.ConnectorID,
			ConversationID: conv.ID,
		})
		if err != nil {
			return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, err)
		}

		return conv.ID, conv.SectionID, false, nil

	} else if env == vo.Online {
		appDynamicConversationOnline := r.query.AppDynamicConversationOnline
		ret, err := appDynamicConversationOnline.WithContext(ctx).Where(
			appDynamicConversationOnline.AppID.Eq(meta.BizID),
			appDynamicConversationOnline.ConnectorID.Eq(meta.ConnectorID),
			appDynamicConversationOnline.UserID.Eq(meta.UserID),
			appDynamicConversationOnline.Name.Eq(meta.Name),
		).First()
		if err == nil {
			cInfo, err := crossconversation.DefaultSVC().GetByID(ctx, ret.ConversationID)
			if err != nil {
				return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, err)
			}
			if cInfo == nil {
				return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("conversation not found"))
			}
			return ret.ConversationID, cInfo.SectionID, true, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, err)
		}

		conv, err := idGen(ctx, meta.BizID, meta.UserID, meta.ConnectorID)
		if err != nil {
			return 0, 0, false, err
		}
		id, err := r.GenID(ctx)
		if err != nil {
			return 0, 0, false, vo.WrapError(errno.ErrIDGenError, err)
		}

		err = r.query.AppDynamicConversationOnline.WithContext(ctx).Create(&model.AppDynamicConversationOnline{
			ID:             id,
			AppID:          meta.BizID,
			Name:           meta.Name,
			UserID:         meta.UserID,
			ConnectorID:    meta.ConnectorID,
			ConversationID: conv.ID,
		})
		if err != nil {
			return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, err)
		}

		return conv.ID, conv.SectionID, false, nil

	} else {
		return 0, 0, false, fmt.Errorf("unknown env %v", env)
	}

}

// GetStaticConversationByTemplateID 根据模板 ID 获取静态会话
//
// 参数：
//   - ctx: 上下文
//   - env: 环境（Draft/Online）
//   - userID: 用户 ID
//   - connectorID: 连接器 ID
//   - templateID: 模板 ID
//
// 返回值：
//   - *entity.StaticConversation: 静态会话
//   - bool: 是否存在
//   - error: 查询失败时返回错误
func (r *RepositoryImpl) GetStaticConversationByTemplateID(ctx context.Context, env vo.Env, userID, connectorID, templateID int64) (*entity.StaticConversation, bool, error) {
	if env == vo.Draft {
		conditions := make([]gen.Condition, 0, 3)
		conditions = append(conditions, r.query.AppStaticConversationDraft.UserID.Eq(userID))
		conditions = append(conditions, r.query.AppStaticConversationDraft.ConnectorID.Eq(connectorID))
		conditions = append(conditions, r.query.AppStaticConversationDraft.TemplateID.Eq(templateID))
		cs, err := r.query.AppStaticConversationDraft.WithContext(ctx).Where(conditions...).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, nil
			}
			return nil, false, vo.WrapError(errno.ErrDatabaseError, err)
		}
		return &entity.StaticConversation{
			UserID:         cs.UserID,
			ConnectorID:    cs.ConnectorID,
			TemplateID:     cs.TemplateID,
			ConversationID: cs.ConversationID,
		}, true, nil
	} else if env == vo.Online {
		conditions := make([]gen.Condition, 0, 3)
		conditions = append(conditions, r.query.AppStaticConversationOnline.UserID.Eq(userID))
		conditions = append(conditions, r.query.AppStaticConversationOnline.ConnectorID.Eq(connectorID))
		conditions = append(conditions, r.query.AppStaticConversationOnline.TemplateID.Eq(templateID))
		cs, err := r.query.AppStaticConversationOnline.WithContext(ctx).Where(conditions...).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, nil
			}
			return nil, false, vo.WrapError(errno.ErrDatabaseError, err)
		}
		return &entity.StaticConversation{
			UserID:         cs.UserID,
			ConnectorID:    cs.ConnectorID,
			TemplateID:     cs.TemplateID,
			ConversationID: cs.ConversationID,
		}, true, nil
	} else {
		return nil, false, fmt.Errorf("unknown env %v", env)
	}
}

// getOrCreateDraftStaticConversation 获取或创建草稿环境的静态会话
func (r *RepositoryImpl) getOrCreateDraftStaticConversation(ctx context.Context, idGen workflow.ConversationIDGenerator, meta *vo.CreateStaticConversation) (int64, int64, bool, error) {
	cs, err := r.mGetDraftStaticConversation(ctx, meta.UserID, meta.ConnectorID, []int64{meta.TemplateID})
	if err != nil {
		return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, err)
	}

	if len(cs) > 0 {
		cInfo, err := crossconversation.DefaultSVC().GetByID(ctx, cs[0].ConversationID)
		if err != nil {
			return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, err)
		}
		if cInfo == nil {
			return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("conversation not found"))
		}
		return cs[0].ConversationID, cInfo.SectionID, true, nil
	}

	conv, err := idGen(ctx, meta.BizID, meta.UserID, meta.ConnectorID)
	if err != nil {
		return 0, 0, false, err
	}

	id, err := r.GenID(ctx)
	if err != nil {
		return 0, 0, false, vo.WrapError(errno.ErrIDGenError, err)
	}
	object := &model.AppStaticConversationDraft{
		ID:             id,
		UserID:         meta.UserID,
		ConnectorID:    meta.ConnectorID,
		TemplateID:     meta.TemplateID,
		ConversationID: conv.ID,
	}
	err = r.query.AppStaticConversationDraft.WithContext(ctx).Create(object)
	if err != nil {
		return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, err)
	}

	return conv.ID, conv.SectionID, false, nil
}

// getOrCreateOnlineStaticConversation 获取或创建在线环境的静态会话
func (r *RepositoryImpl) getOrCreateOnlineStaticConversation(ctx context.Context, idGen workflow.ConversationIDGenerator, meta *vo.CreateStaticConversation) (int64, int64, bool, error) {
	cs, err := r.mGetOnlineStaticConversation(ctx, meta.UserID, meta.ConnectorID, []int64{meta.TemplateID})
	if err != nil {
		return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, err)
	}

	if len(cs) > 0 {
		cInfo, err := crossconversation.DefaultSVC().GetByID(ctx, cs[0].ConversationID)
		if err != nil {
			return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, err)
		}
		if cInfo == nil {
			return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, fmt.Errorf("conversation not found"))
		}
		return cs[0].ConversationID, cInfo.SectionID, true, nil
	}

	conv, err := idGen(ctx, meta.BizID, meta.UserID, meta.ConnectorID)
	if err != nil {
		return 0, 0, false, err
	}

	id, err := r.GenID(ctx)
	if err != nil {
		return 0, 0, false, vo.WrapError(errno.ErrIDGenError, err)
	}
	object := &model.AppStaticConversationOnline{
		ID:             id,
		UserID:         meta.UserID,
		ConnectorID:    meta.ConnectorID,
		TemplateID:     meta.TemplateID,
		ConversationID: conv.ID,
	}
	err = r.query.AppStaticConversationOnline.WithContext(ctx).Create(object)
	if err != nil {
		return 0, 0, false, vo.WrapError(errno.ErrDatabaseError, err)
	}

	return conv.ID, conv.SectionID, false, nil
}

// BatchCreateOnlineConversationTemplate 批量创建在线会话模板
//
// 在发布应用时，将草稿环境的模板复制到在线环境。
//
// 参数：
//   - ctx: 上下文
//   - templates: 模板列表
//   - version: 版本号
//
// 返回值：
//   - error: 创建失败时返回错误
func (r *RepositoryImpl) BatchCreateOnlineConversationTemplate(ctx context.Context, templates []*entity.ConversationTemplate, version string) error {
	ids, err := r.GenMultiIDs(ctx, len(templates))
	if err != nil {
		return vo.WrapError(errno.ErrIDGenError, err)
	}

	objects := make([]*model.AppConversationTemplateOnline, 0, len(templates))
	for idx := range templates {
		template := templates[idx]
		objects = append(objects, &model.AppConversationTemplateOnline{
			ID:         ids[idx],
			SpaceID:    template.SpaceID,
			AppID:      template.AppID,
			TemplateID: template.TemplateID,
			Name:       template.Name,
			Version:    version,
		})
	}

	err = r.query.AppConversationTemplateOnline.WithContext(ctx).CreateInBatches(objects, batchSize)
	if err != nil {
		return vo.WrapError(errno.ErrDatabaseError, err)
	}
	return nil

}

// GetDynamicConversationByName 根据名称获取动态会话
//
// 参数：
//   - ctx: 上下文
//   - env: 环境（Draft/Online）
//   - appID: 应用 ID
//   - connectorID: 连接器 ID
//   - userID: 用户 ID
//   - name: 会话名称
//
// 返回值：
//   - *entity.DynamicConversation: 动态会话
//   - bool: 是否存在
//   - error: 查询失败时返回错误
func (r *RepositoryImpl) GetDynamicConversationByName(ctx context.Context, env vo.Env, appID, connectorID, userID int64, name string) (*entity.DynamicConversation, bool, error) {
	if env == vo.Draft {
		appDynamicConversationDraft := r.query.AppDynamicConversationDraft
		ret, err := appDynamicConversationDraft.WithContext(ctx).Where(
			appDynamicConversationDraft.AppID.Eq(appID),
			appDynamicConversationDraft.ConnectorID.Eq(connectorID),
			appDynamicConversationDraft.UserID.Eq(userID),
			appDynamicConversationDraft.Name.Eq(name)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, nil
			}
			return nil, false, err
		}

		return &entity.DynamicConversation{
			ID:             ret.ID,
			UserID:         ret.UserID,
			ConnectorID:    ret.ConnectorID,
			ConversationID: ret.ConversationID,
			Name:           ret.Name,
		}, true, nil

	} else if env == vo.Online {
		appDynamicConversationOnline := r.query.AppDynamicConversationOnline
		ret, err := appDynamicConversationOnline.WithContext(ctx).Where(
			appDynamicConversationOnline.AppID.Eq(appID),
			appDynamicConversationOnline.ConnectorID.Eq(connectorID),
			appDynamicConversationOnline.UserID.Eq(userID),
			appDynamicConversationOnline.Name.Eq(name)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, nil
			}
			return nil, false, err
		}
		return &entity.DynamicConversation{
			ID:             ret.ID,
			UserID:         ret.UserID,
			ConnectorID:    ret.ConnectorID,
			ConversationID: ret.ConversationID,
			Name:           ret.Name,
		}, true, nil

	} else {
		return nil, false, fmt.Errorf("unknown env %v", env)
	}

}

// UpdateDynamicConversationNameByID 更新动态会话名称
//
// 参数：
//   - ctx: 上下文
//   - env: 环境（Draft/Online）
//   - templateID: 会话 ID
//   - name: 新名称
//
// 返回值：
//   - error: 更新失败时返回错误
func (r *RepositoryImpl) UpdateDynamicConversationNameByID(ctx context.Context, env vo.Env, templateID int64, name string) error {
	if env == vo.Draft {
		appDynamicConversationDraft := r.query.AppDynamicConversationDraft
		_, err := appDynamicConversationDraft.WithContext(ctx).Where(
			appDynamicConversationDraft.ID.Eq(templateID),
		).UpdateColumnSimple(appDynamicConversationDraft.Name.Value(name))
		if err != nil {
			return vo.WrapError(errno.ErrDatabaseError, err)
		}
		return nil
	} else if env == vo.Online {
		appDynamicConversationOnline := r.query.AppDynamicConversationOnline
		_, err := appDynamicConversationOnline.WithContext(ctx).Where(
			appDynamicConversationOnline.ID.Eq(templateID),
		).UpdateColumnSimple(appDynamicConversationOnline.Name.Value(name))
		if err != nil {
			return vo.WrapError(errno.ErrDatabaseError, err)
		}
		return nil

	} else {
		return fmt.Errorf("unknown env %v", env)
	}

}

// UpdateStaticConversation 更新静态会话的会话 ID
//
// 用于重置会话时替换底层会话。
//
// 参数：
//   - ctx: 上下文
//   - env: 环境（Draft/Online）
//   - templateID: 模板 ID
//   - connectorID: 连接器 ID
//   - userID: 用户 ID
//   - newConversationID: 新会话 ID
//
// 返回值：
//   - error: 更新失败时返回错误
func (r *RepositoryImpl) UpdateStaticConversation(ctx context.Context, env vo.Env, templateID int64, connectorID int64, userID int64, newConversationID int64) error {

	if env == vo.Draft {
		appStaticConversationDraft := r.query.AppStaticConversationDraft
		_, err := appStaticConversationDraft.WithContext(ctx).Where(
			appStaticConversationDraft.TemplateID.Eq(templateID),
			appStaticConversationDraft.ConnectorID.Eq(connectorID),
			appStaticConversationDraft.UserID.Eq(userID),
		).UpdateColumn(appStaticConversationDraft.ConversationID, newConversationID)

		if err != nil {
			return err
		}
		return err

	} else if env == vo.Online {
		appStaticConversationOnline := r.query.AppStaticConversationOnline
		_, err := appStaticConversationOnline.WithContext(ctx).Where(
			appStaticConversationOnline.TemplateID.Eq(templateID),
			appStaticConversationOnline.ConnectorID.Eq(connectorID),
			appStaticConversationOnline.UserID.Eq(userID),
		).UpdateColumn(appStaticConversationOnline.ConversationID, newConversationID)
		if err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("unknown env %v", env)
	}

}

// UpdateDynamicConversation 更新动态会话的会话 ID
//
// 用于重置会话时替换底层会话。
//
// 参数：
//   - ctx: 上下文
//   - env: 环境（Draft/Online）
//   - conversationID: 原会话 ID
//   - newConversationID: 新会话 ID
//
// 返回值：
//   - error: 更新失败时返回错误
func (r *RepositoryImpl) UpdateDynamicConversation(ctx context.Context, env vo.Env, conversationID, newConversationID int64) error {
	if env == vo.Draft {
		appDynamicConversationDraft := r.query.AppDynamicConversationDraft
		_, err := appDynamicConversationDraft.WithContext(ctx).Where(appDynamicConversationDraft.ConversationID.Eq(conversationID)).
			UpdateColumn(appDynamicConversationDraft.ConversationID, newConversationID)
		if err != nil {
			return err
		}

		return nil
	} else if env == vo.Online {
		appDynamicConversationOnline := r.query.AppDynamicConversationOnline
		_, err := appDynamicConversationOnline.WithContext(ctx).Where(appDynamicConversationOnline.ConversationID.Eq(conversationID)).
			UpdateColumn(appDynamicConversationOnline.ConversationID, newConversationID)
		if err != nil {
			return err
		}

		return nil
	} else {
		return fmt.Errorf("unknown env %v", env)
	}

}

// CopyTemplateConversationByAppID 复制应用的会话模板
//
// 在复制应用时，将源应用的会话模板复制到目标应用。
// 不会复制名为 "Default" 的默认模板。
//
// 参数：
//   - ctx: 上下文
//   - appID: 源应用 ID
//   - toAppID: 目标应用 ID
//
// 返回值：
//   - error: 复制失败时返回错误
func (r *RepositoryImpl) CopyTemplateConversationByAppID(ctx context.Context, appID int64, toAppID int64) error {
	appConversationTemplateDraft := r.query.AppConversationTemplateDraft
	templates, err := appConversationTemplateDraft.WithContext(ctx).Where(appConversationTemplateDraft.AppID.Eq(appID), appConversationTemplateDraft.Name.Neq("Default")).Find()
	if err != nil {
		return vo.WrapError(errno.ErrDatabaseError, err)
	}

	if len(templates) == 0 {
		return nil
	}
	templateTemplates := make([]*model.AppConversationTemplateDraft, 0, len(templates))
	ids, err := r.GenMultiIDs(ctx, len(templates))
	if err != nil {
		return vo.WrapError(errno.ErrIDGenError, err)
	}
	for i := range templates {
		copiedTemplate := templates[i]
		copiedTemplate.ID = ids[i]
		copiedTemplate.TemplateID = ids[i]
		copiedTemplate.AppID = toAppID
		templateTemplates = append(templateTemplates, copiedTemplate)
	}
	err = appConversationTemplateDraft.WithContext(ctx).CreateInBatches(templateTemplates, batchSize)
	if err != nil {
		return vo.WrapError(errno.ErrDatabaseError, err)
	}
	return nil

}

// GetStaticConversationByID 根据会话 ID 获取静态会话的模板名称
//
// 参数：
//   - ctx: 上下文
//   - env: 环境（Draft/Online）
//   - bizID: 业务 ID（应用 ID）
//   - connectorID: 连接器 ID
//   - conversationID: 会话 ID
//
// 返回值：
//   - string: 模板名称
//   - bool: 是否存在
//   - error: 查询失败时返回错误
func (r *RepositoryImpl) GetStaticConversationByID(ctx context.Context, env vo.Env, bizID, connectorID, conversationID int64) (string, bool, error) {
	if env == vo.Draft {
		appStaticConversationDraft := r.query.AppStaticConversationDraft
		ret, err := appStaticConversationDraft.WithContext(ctx).Where(
			appStaticConversationDraft.ConnectorID.Eq(connectorID),
			appStaticConversationDraft.ConversationID.Eq(conversationID),
		).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", false, nil
			}
			return "", false, err
		}
		appConversationTemplateDraft := r.query.AppConversationTemplateDraft
		template, err := appConversationTemplateDraft.WithContext(ctx).Where(
			appConversationTemplateDraft.TemplateID.Eq(ret.TemplateID),
			appConversationTemplateDraft.AppID.Eq(bizID),
		).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", false, nil
			}
			return "", false, err
		}
		return template.Name, true, nil
	} else if env == vo.Online {
		appStaticConversationOnline := r.query.AppStaticConversationOnline
		ret, err := appStaticConversationOnline.WithContext(ctx).Where(
			appStaticConversationOnline.ConnectorID.Eq(connectorID),
			appStaticConversationOnline.ConversationID.Eq(conversationID),
		).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", false, nil
			}
			return "", false, err
		}
		appConversationTemplateOnline := r.query.AppConversationTemplateOnline
		template, err := appConversationTemplateOnline.WithContext(ctx).Where(
			appConversationTemplateOnline.TemplateID.Eq(ret.TemplateID),
			appConversationTemplateOnline.AppID.Eq(bizID),
		).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", false, nil
			}
			return "", false, err
		}
		return template.Name, true, nil
	}
	return "", false, fmt.Errorf("unknown env %v", env)
}

// GetDynamicConversationByID 根据会话 ID 获取动态会话
//
// 参数：
//   - ctx: 上下文
//   - env: 环境（Draft/Online）
//   - bizID: 业务 ID（应用 ID）
//   - connectorID: 连接器 ID
//   - conversationID: 会话 ID
//
// 返回值：
//   - *entity.DynamicConversation: 动态会话
//   - bool: 是否存在
//   - error: 查询失败时返回错误
func (r *RepositoryImpl) GetDynamicConversationByID(ctx context.Context, env vo.Env, bizID, connectorID, conversationID int64) (*entity.DynamicConversation, bool, error) {
	if env == vo.Draft {
		appDynamicConversationDraft := r.query.AppDynamicConversationDraft
		ret, err := appDynamicConversationDraft.WithContext(ctx).Where(
			appDynamicConversationDraft.AppID.Eq(bizID),
			appDynamicConversationDraft.ConnectorID.Eq(connectorID),
			appDynamicConversationDraft.ConversationID.Eq(conversationID),
		).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, nil
			}
			return nil, false, err
		}
		return &entity.DynamicConversation{
			ID:             ret.ID,
			UserID:         ret.UserID,
			ConnectorID:    ret.ConnectorID,
			ConversationID: ret.ConversationID,
			Name:           ret.Name,
		}, true, nil
	} else if env == vo.Online {
		appDynamicConversationOnline := r.query.AppDynamicConversationOnline
		ret, err := appDynamicConversationOnline.WithContext(ctx).Where(
			appDynamicConversationOnline.AppID.Eq(bizID),
			appDynamicConversationOnline.ConnectorID.Eq(connectorID),
			appDynamicConversationOnline.ConversationID.Eq(conversationID),
		).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, nil
			}
			return nil, false, err
		}
		return &entity.DynamicConversation{
			ID:             ret.ID,
			UserID:         ret.UserID,
			ConnectorID:    ret.ConnectorID,
			ConversationID: ret.ConversationID,
			Name:           ret.Name,
		}, true, nil
	}
	return nil, false, fmt.Errorf("unknown env %v", env)
}
