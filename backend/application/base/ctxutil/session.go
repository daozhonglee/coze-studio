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

// Package ctxutil 提供上下文工具函数
//
// 本包提供从上下文中获取用户会话、API 认证信息等的工具函数
package ctxutil

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/user/entity"
	"github.com/coze-dev/coze-studio/backend/pkg/ctxcache"
	"github.com/coze-dev/coze-studio/backend/types/consts"
)

// GetUserSessionFromCtx 从上下文获取用户会话信息
func GetUserSessionFromCtx(ctx context.Context) *entity.Session {
	data, ok := ctxcache.Get[*entity.Session](ctx, consts.SessionDataKeyInCtx)
	if !ok {
		return nil
	}

	return data
}

// MustGetUIDFromCtx 从上下文获取用户 ID，如果不存在则 panic
func MustGetUIDFromCtx(ctx context.Context) int64 {
	sessionData := GetUserSessionFromCtx(ctx)
	if sessionData == nil {
		panic("mustGetUIDFromCtx: sessionData is nil")
	}

	return sessionData.UserID
}

// GetUIDFromCtx 从上下文获取用户 ID，如果不存在返回 nil
func GetUIDFromCtx(ctx context.Context) *int64 {
	sessionData := GetUserSessionFromCtx(ctx)
	if sessionData == nil {
		return nil
	}

	return &sessionData.UserID
}
