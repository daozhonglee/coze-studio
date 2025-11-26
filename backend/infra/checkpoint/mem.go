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

// Package checkpoint 提供工作流检查点存储功能（内存实现）
package checkpoint

import (
	"context"
	"sync"

	"github.com/cloudwego/eino/compose"
)

// inMemoryStore 内存检查点存储实现（用于测试和开发）
type inMemoryStore struct {
	m  map[string][]byte
	mu sync.RWMutex
}

// Get 获取检查点数据
func (i *inMemoryStore) Get(_ context.Context, checkPointID string) ([]byte, bool, error) {
	i.mu.RLock()
	v, ok := i.m[checkPointID]
	i.mu.RUnlock()
	return v, ok, nil
}

// Set 保存检查点数据
func (i *inMemoryStore) Set(_ context.Context, checkPointID string, checkPoint []byte) error {
	i.mu.Lock()
	i.m[checkPointID] = checkPoint
	i.mu.Unlock()
	return nil
}

// NewInMemoryStore 创建内存检查点存储实例
func NewInMemoryStore() compose.CheckPointStore {
	return &inMemoryStore{
		m: make(map[string][]byte),
	}
}
