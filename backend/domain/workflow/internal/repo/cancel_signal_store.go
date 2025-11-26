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
	"fmt"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/infra/cache"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

// cancelSignalStoreImpl 取消信号存储的实现
//
// 该结构体基于 Redis 实现工作流执行的取消信号存储。
// 当用户发起取消请求时，设置取消标志位；执行中的工作流会轮询此标志位来响应取消。
type cancelSignalStoreImpl struct {
	redis cache.Cmdable // Redis 客户端
}

// workflowExecutionCancelStatusKey Redis 键模式，用于存储工作流执行的取消状态
// 格式：workflow:cancel:status:{工作流执行ID}
const workflowExecutionCancelStatusKey = "workflow:cancel:status:%d"

// SetWorkflowCancelFlag 设置工作流取消标志
//
// 在 Redis 中设置取消状态键，有效期 24 小时。
// 正在执行的工作流会定期检查此标志来响应取消请求。
//
// 参数：
//   - ctx: 上下文
//   - wfExeID: 工作流执行 ID
//
// 返回值：
//   - error: 设置失败时返回错误
func (c *cancelSignalStoreImpl) SetWorkflowCancelFlag(ctx context.Context, wfExeID int64) (err error) {
	statusKey := fmt.Sprintf(workflowExecutionCancelStatusKey, wfExeID)
	// Define a reasonable expiration for the status key, e.g., 24 hours
	expiration := 24 * time.Hour

	// set a kv to redis to indicate cancellation status
	err = c.redis.Set(ctx, statusKey, "cancelled", expiration).Err()
	if err != nil {
		return vo.WrapError(errno.ErrRedisError,
			fmt.Errorf("failed to set workflow cancel status for wfExeID %d after publishing signal: %w", wfExeID, err))
	}

	return nil
}

// GetWorkflowCancelFlag 获取工作流取消标志
//
// 检查 Redis 中是否存在取消状态键。
//
// 参数：
//   - ctx: 上下文
//   - wfExeID: 工作流执行 ID
//
// 返回值：
//   - bool: 是否已被取消
//   - error: 查询失败时返回错误
func (c *cancelSignalStoreImpl) GetWorkflowCancelFlag(ctx context.Context, wfExeID int64) (bool, error) {
	// Construct Redis key for workflow cancellation status
	key := fmt.Sprintf(workflowExecutionCancelStatusKey, wfExeID)

	// Check if the key exists in Redis
	count, err := c.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, vo.WrapError(errno.ErrRedisError, fmt.Errorf("failed to check cancellation status in Redis: %w", err))
	}

	// If key exists (count == 1), return true; otherwise return false
	return count == 1, nil
}
