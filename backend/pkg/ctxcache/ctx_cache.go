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

// Package ctxcache 提供上下文缓存工具
//
// 本包提供基于 context 的缓存能力，用于在请求生命周期内缓存数据
package ctxcache

import (
	"context"
	"sync"
)

// ctxCacheKey 上下文缓存键
type ctxCacheKey struct{}

// Init 初始化上下文缓存
func Init(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxCacheKey{}, new(sync.Map))
}

// Get 从上下文缓存获取值
func Get[T any](ctx context.Context, key any) (value T, ok bool) {
	var zero T

	cacheMap, valid := ctx.Value(ctxCacheKey{}).(*sync.Map)
	if !valid {
		return zero, false
	}

	loadedValue, exists := cacheMap.Load(key)
	if !exists {
		return zero, false
	}

	if v, match := loadedValue.(T); match {
		return v, true
	}

	return zero, false
}

// Store 向上下文缓存存储值
func Store(ctx context.Context, key any, obj any) {
	if cacheMap, ok := ctx.Value(ctxCacheKey{}).(*sync.Map); ok {
		cacheMap.Store(key, obj)
	}
}

// HasKey 检查上下文缓存中是否存在键
func HasKey(ctx context.Context, key any) bool {
	if cacheMap, ok := ctx.Value(ctxCacheKey{}).(*sync.Map); ok {
		_, ok := cacheMap.Load(key)
		return ok
	}

	return false
}
