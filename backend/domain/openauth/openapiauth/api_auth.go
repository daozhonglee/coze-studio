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

// Package openapiauth 定义了 OpenAPI 认证领域的服务层
//
// 本包提供 API 密钥的管理和认证服务，包括：
// - API 密钥的 CRUD 操作
// - 权限校验
package openapiauth

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/openauth/openapiauth/entity"
)

// APIAuth API 认证服务接口
//
// 定义 API 密钥的管理和认证操作
type APIAuth interface {
	// Create 创建 API 密钥
	Create(ctx context.Context, req *entity.CreateApiKey) (*entity.ApiKey, error)
	// Delete 删除 API 密钥
	Delete(ctx context.Context, req *entity.DeleteApiKey) error
	// Get 获取 API 密钥
	Get(ctx context.Context, req *entity.GetApiKey) (*entity.ApiKey, error)
	// List 列出用户的 API 密钥
	List(ctx context.Context, req *entity.ListApiKey) (*entity.ListApiKeyResp, error)
	// Save 更新 API 密钥元数据
	Save(ctx context.Context, req *entity.SaveMeta) error

	// CheckPermission 验证 API 密钥权限
	// 返回有效的 ApiKey 实体，如果密钥无效则返回 nil
	CheckPermission(ctx context.Context, req *entity.CheckPermission) (*entity.ApiKey, error)
}
