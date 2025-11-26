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

// Package checkpoint 提供工作流检查点存储功能
//
// 本包提供检查点（Checkpoint）的存储和读取功能，用于：
// - 工作流执行状态的持久化
// - 支持工作流中断后恢复执行
// - 支持 Redis 和内存两种存储后端
package checkpoint

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cloudwego/eino/compose"

	"github.com/coze-dev/coze-studio/backend/infra/cache"
)

// redisStore Redis 检查点存储实现
type redisStore struct {
	client cache.Cmdable
}

// 检查点存储配置常量
const (
	// checkpointKeyTpl 检查点 Redis 键模板
	checkpointKeyTpl = "checkpoint_key:%s"
	// checkpointExpire 检查点过期时间（7天）
	checkpointExpire = 24 * 7 * 3600 * time.Second
)

// Get 获取检查点数据
func (r *redisStore) Get(ctx context.Context, checkPointID string) ([]byte, bool, error) {
	v, err := r.client.Get(ctx, fmt.Sprintf(checkpointKeyTpl, checkPointID)).Bytes()
	if err != nil {
		if errors.Is(err, cache.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return v, true, nil
}

// Set 保存检查点数据
func (r *redisStore) Set(ctx context.Context, checkPointID string, checkPoint []byte) error {
	return r.client.Set(ctx, fmt.Sprintf(checkpointKeyTpl, checkPointID), checkPoint, checkpointExpire).Err()
}

// NewRedisStore 创建 Redis 检查点存储实例
func NewRedisStore(client cache.Cmdable) compose.CheckPointStore {
	return &redisStore{client: client}
}
