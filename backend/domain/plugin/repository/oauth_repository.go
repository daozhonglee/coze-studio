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

// oauth_repository.go OAuth 仓储接口定义
//
// 本文件定义了 OAuth 认证数据访问的仓储接口 OAuthRepository。
// 提供授权码的管理和令牌刷新操作。

package repository

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/plugin/dto"
)

// OAuthRepository OAuth 认证仓储接口
//
// 提供 OAuth 认证数据的持久化操作：
//   - 授权码的获取、更新、删除
//   - 令牌刷新调度
//   - 过期和非活跃令牌清理
//
//go:generate mockgen -source=./oauth_repository.go -package=mock_plugin_oauth -destination=./mock/mock_oauth_repository.go
type OAuthRepository interface {
	GetAuthorizationCode(ctx context.Context, meta *dto.AuthorizationCodeMeta) (info *dto.AuthorizationCodeInfo, exist bool, err error)
	UpsertAuthorizationCode(ctx context.Context, info *dto.AuthorizationCodeInfo) (err error)
	UpdateAuthorizationCodeLastActiveAt(ctx context.Context, meta *dto.AuthorizationCodeMeta, lastActiveAtMs int64) (err error)
	BatchDeleteAuthorizationCodeByIDs(ctx context.Context, ids []int64) (err error)
	DeleteAuthorizationCode(ctx context.Context, meta *dto.AuthorizationCodeMeta) (err error)
	GetAuthorizationCodeRefreshTokens(ctx context.Context, nextRefreshAt int64, limit int) (infos []*dto.AuthorizationCodeInfo, err error)
	DeleteExpiredAuthorizationCodeTokens(ctx context.Context, expireAt int64, limit int) (err error)
	DeleteInactiveAuthorizationCodeTokens(ctx context.Context, lastActiveAt int64, limit int) (err error)
}
