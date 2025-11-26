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

// Package prompt 定义了提示词(Prompt)领域的服务层实现

package prompt

import (
	"context"
	"strings"

	"github.com/coze-dev/coze-studio/backend/domain/prompt/entity"
	"github.com/coze-dev/coze-studio/backend/domain/prompt/internal/official"
	"github.com/coze-dev/coze-studio/backend/domain/prompt/repository"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/slices"
)

// promptService 提示词服务实现
type promptService struct {
	Repo repository.PromptRepository
}

// NewService 创建提示词服务实例
func NewService(repo repository.PromptRepository) Prompt {
	return &promptService{
		Repo: repo,
	}
}

// CreatePromptResource 创建提示词资源
func (s *promptService) CreatePromptResource(ctx context.Context, p *entity.PromptResource) (int64, error) {
	return s.Repo.CreatePromptResource(ctx, p)
}

// UpdatePromptResource 更新提示词资源
func (s *promptService) UpdatePromptResource(ctx context.Context, promptID int64, name, description, promptText *string) error {
	return s.Repo.UpdatePromptResource(ctx, promptID, name, description, promptText)
}

// GetPromptResource 获取提示词资源
func (s *promptService) GetPromptResource(ctx context.Context, promptID int64) (*entity.PromptResource, error) {
	return s.Repo.GetPromptResource(ctx, promptID)
}

// DeletePromptResource 删除提示词资源
func (s *promptService) DeletePromptResource(ctx context.Context, promptID int64) error {
	err := s.Repo.DeletePromptResource(ctx, promptID)
	if err != nil {
		return err
	}

	return nil
}

// ListOfficialPromptResource 列出官方提示词模板
//
// 从预定义的官方提示词列表中查询，支持按关键词过滤
func (s *promptService) ListOfficialPromptResource(ctx context.Context, keyword string) ([]*entity.PromptResource, error) {
	promptList := official.GetPromptList()

	promptList = searchPromptResourceList(ctx, promptList, keyword)
	return deepCopyPromptResource(promptList), nil
}

// deepCopyPromptResource 深拷贝提示词资源列表
func deepCopyPromptResource(pl []*entity.PromptResource) []*entity.PromptResource {
	return slices.Transform(pl, func(p *entity.PromptResource) *entity.PromptResource {
		return &entity.PromptResource{
			ID:          p.ID,
			SpaceID:     p.SpaceID,
			Name:        p.Name,
			Description: p.Description,
			PromptText:  p.PromptText,
			Status:      1,
		}
	})
}

// searchPromptResourceList 按关键词搜索提示词列表
//
// 支持在名称和内容中进行模糊匹配（忽略大小写）
func searchPromptResourceList(ctx context.Context, resource []*entity.PromptResource, keyword string) []*entity.PromptResource {
	if len(keyword) == 0 {
		return resource
	}

	retVal := make([]*entity.PromptResource, 0, len(resource))
	for _, promptResource := range resource {
		if promptResource == nil {
			continue
		}
		// name match
		if strings.Contains(strings.ToLower(promptResource.Name), strings.ToLower(keyword)) {
			retVal = append(retVal, promptResource)
			continue
		}
		// Body Match
		if strings.Contains(strings.ToLower(promptResource.PromptText), strings.ToLower(keyword)) {
			retVal = append(retVal, promptResource)
		}
	}
	return retVal
}
