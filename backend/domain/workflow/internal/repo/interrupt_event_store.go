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
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/infra/cache"
	"github.com/coze-dev/coze-studio/backend/pkg/sonic"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

// interruptEventStoreImpl 中断事件存储的实现
//
// 该结构体基于 Redis 实现工作流中断事件的存储和管理。
// 工作流在执行过程中遇到需要用户交互的节点（如问答节点）时，
// 会产生中断事件并暂停执行，等待用户响应后恢复。
//
// 中断事件以列表形式存储在 Redis 中，支持：
//   - 按顺序处理多个中断事件
//   - 优先处理上一次恢复的节点产生的新事件
type interruptEventStoreImpl struct {
	redis cache.Cmdable // Redis 客户端
}

const (
	// interruptEventListKeyPattern 中断事件列表的 Redis 键模式
	// 格式：interrupt_event_list:{工作流执行ID}
	interruptEventListKeyPattern = "interrupt_event_list:%d"

	// interruptEventTTL 中断事件的过期时间（24 小时）
	interruptEventTTL = 24 * time.Hour

	// previousResumedEventKeyPattern 上一次恢复的事件键模式
	// 用于优先处理同一节点产生的新中断事件
	previousResumedEventKeyPattern = "previous_resumed_event:%d"

	// ConvToEventExecFormat 会话与执行关联信息的键模式
	ConvToEventExecFormat = "conv_relate_info:%d"
)

// SaveInterruptEvents 保存中断事件到 Redis 列表
//
// 将多个中断事件追加到 Redis 列表末尾。如果事件已存在则跳过。
// 如果有之前恢复过的事件，会优先将同一节点的新事件放到列表头部。
//
// 参数：
//   - ctx: 上下文
//   - wfExeID: 工作流执行 ID
//   - events: 要保存的中断事件列表
//
// 返回值：
//   - error: 保存失败时返回错误
func (i *interruptEventStoreImpl) SaveInterruptEvents(ctx context.Context, wfExeID int64, events []*entity.InterruptEvent) (err error) {
	if len(events) == 0 {
		return nil
	}

	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrRedisError, err)
		}
	}()

	listKey := fmt.Sprintf(interruptEventListKeyPattern, wfExeID)
	previousResumedEventKey := fmt.Sprintf(previousResumedEventKeyPattern, wfExeID)

	currentEvents, err := i.ListInterruptEvents(ctx, wfExeID)
	if err != nil {
		return err
	}

	for _, currentE := range currentEvents {
		if len(events) == 0 {
			break
		}
		j := len(events)
		for i := 0; i < j; i++ {
			if events[i].ID == currentE.ID {
				events = append(events[:i], events[i+1:]...)
				i--
				j--
			}
		}
	}

	if len(events) == 0 {
		return nil
	}

	previousEventStr, err := i.redis.Get(ctx, previousResumedEventKey).Result()
	if err != nil {
		if !errors.Is(err, cache.Nil) {
			return fmt.Errorf("failed to get previous resumed event for wfExeID %d: %w", wfExeID, err)
		}
	}

	var previousEvent *entity.InterruptEvent
	if previousEventStr != "" {
		err = sonic.UnmarshalString(previousEventStr, &previousEvent)
		if err != nil {
			return vo.WrapError(errno.ErrSerializationDeserializationFail,
				fmt.Errorf("failed to unmarshal previous resumed event (wfExeID %d) from JSON: %w", wfExeID, err))
		}
	}

	var topPriorityEvent *entity.InterruptEvent
	if previousEvent != nil {
		for i := range events {
			if previousEvent.NodeKey == events[i].NodeKey {
				topPriorityEvent = events[i]
				events = append(events[:i], events[i+1:]...)
				break
			}
		}
	}

	pipe := i.redis.Pipeline()
	eventJSONs := make([]interface{}, 0, len(events))

	for _, event := range events {
		eventJSON, err := sonic.MarshalString(event)
		if err != nil {
			return vo.WrapError(errno.ErrSerializationDeserializationFail,
				fmt.Errorf("failed to marshal interrupt event %d to JSON: %w", event.ID, err))
		}
		eventJSONs = append(eventJSONs, eventJSON)
	}

	if topPriorityEvent != nil {
		topPriorityEventJSON, err := sonic.MarshalString(topPriorityEvent)
		if err != nil {
			return vo.WrapError(errno.ErrSerializationDeserializationFail,
				fmt.Errorf("failed to marshal top priority interrupt event %d to JSON: %w", topPriorityEvent.ID, err))
		}
		pipe.LPush(ctx, listKey, topPriorityEventJSON)
	}

	if len(eventJSONs) > 0 {
		pipe.RPush(ctx, listKey, eventJSONs...)
	}

	pipe.Expire(ctx, listKey, interruptEventTTL)

	_, err = pipe.Exec(ctx) // ignore_security_alert SQL_INJECTION
	if err != nil {
		return fmt.Errorf("failed to save interrupt events to Redis list: %w", err)
	}

	return nil
}

// GetFirstInterruptEvent 获取列表中的第一个中断事件（不删除）
//
// 参数：
//   - ctx: 上下文
//   - wfExeID: 工作流执行 ID
//
// 返回值：
//   - *entity.InterruptEvent: 中断事件
//   - bool: 是否存在事件
//   - error: 获取失败时返回错误
func (i *interruptEventStoreImpl) GetFirstInterruptEvent(ctx context.Context, wfExeID int64) (
	_ *entity.InterruptEvent, _ bool, err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrRedisError, err)
		}
	}()

	listKey := fmt.Sprintf(interruptEventListKeyPattern, wfExeID)

	eventJSON, err := i.redis.LIndex(ctx, listKey, 0).Result()
	if err != nil {
		if errors.Is(err, cache.Nil) {
			return nil, false, nil // List is empty or key does not exist
		}
		return nil, false, fmt.Errorf("failed to get first interrupt event from Redis list for wfExeID %d: %w", wfExeID, err)
	}

	var event entity.InterruptEvent
	err = sonic.UnmarshalString(eventJSON, &event)
	if err != nil {
		return nil, false, vo.WrapError(errno.ErrSerializationDeserializationFail,
			fmt.Errorf("failed to unmarshal first interrupt event (wfExeID %d) from JSON: %w", wfExeID, err))
	}

	return &event, true, nil
}

// UpdateFirstInterruptEvent 更新列表中的第一个中断事件
//
// 同时会记录此事件为"上一次恢复的事件"，用于后续优先处理。
//
// 参数：
//   - ctx: 上下文
//   - wfExeID: 工作流执行 ID
//   - event: 更新后的事件
//
// 返回值：
//   - error: 更新失败时返回错误
func (i *interruptEventStoreImpl) UpdateFirstInterruptEvent(ctx context.Context, wfExeID int64, event *entity.InterruptEvent) (err error) {
	defer func() {
		if err != nil {
			err = vo.WrapIfNeeded(errno.ErrRedisError, err)
		}
	}()

	listKey := fmt.Sprintf(interruptEventListKeyPattern, wfExeID)
	eventJSON, err := sonic.MarshalString(event)
	if err != nil {
		return vo.WrapError(errno.ErrSerializationDeserializationFail,
			fmt.Errorf("failed to marshal interrupt event %d to JSON: %w", event.ID, err))
	}
	err = i.redis.LSet(ctx, listKey, 0, eventJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to update first interrupt event in Redis list for wfExeID %d: %w", wfExeID, err)
	}

	previousResumedEventKey := fmt.Sprintf(previousResumedEventKeyPattern, wfExeID)
	err = i.redis.Set(ctx, previousResumedEventKey, eventJSON, interruptEventTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set previous resumed event for wfExeID %d: %w", wfExeID, err)
	}

	return nil
}

// PopFirstInterruptEvent 获取并删除列表中的第一个中断事件
//
// 参数：
//   - ctx: 上下文
//   - wfExeID: 工作流执行 ID
//
// 返回值：
//   - *entity.InterruptEvent: 中断事件
//   - bool: 是否存在事件
//   - error: 操作失败时返回错误
func (i *interruptEventStoreImpl) PopFirstInterruptEvent(ctx context.Context, wfExeID int64) (*entity.InterruptEvent, bool, error) {
	listKey := fmt.Sprintf(interruptEventListKeyPattern, wfExeID)

	eventJSON, err := i.redis.LPop(ctx, listKey).Result()
	if err != nil {
		if errors.Is(err, cache.Nil) {
			return nil, false, nil // List is empty or key does not exist
		}
		return nil, false, vo.WrapError(errno.ErrRedisError,
			fmt.Errorf("failed to pop first interrupt event from Redis list for wfExeID %d: %w", wfExeID, err))
	}

	var event entity.InterruptEvent
	err = sonic.UnmarshalString(eventJSON, &event)
	if err != nil {
		// If unmarshalling fails, the event is already popped.
		// Consider if you need to re-queue or handle this scenario.
		return nil, true, vo.WrapError(errno.ErrSerializationDeserializationFail,
			fmt.Errorf("failed to unmarshal popped interrupt event (wfExeID %d) from JSON: %w", wfExeID, err))
	}

	return &event, true, nil
}

// ListInterruptEvents 列出所有中断事件
//
// 参数：
//   - ctx: 上下文
//   - wfExeID: 工作流执行 ID
//
// 返回值：
//   - []*entity.InterruptEvent: 中断事件列表
//   - error: 获取失败时返回错误
func (i *interruptEventStoreImpl) ListInterruptEvents(ctx context.Context, wfExeID int64) ([]*entity.InterruptEvent, error) {
	listKey := fmt.Sprintf(interruptEventListKeyPattern, wfExeID)

	eventJSONs, err := i.redis.LRange(ctx, listKey, 0, -1).Result()
	if err != nil {
		if errors.Is(err, cache.Nil) {
			return nil, nil // List is empty or key does not exist
		}
		return nil, vo.WrapError(errno.ErrRedisError,
			fmt.Errorf("failed to get all interrupt events from Redis list for wfExeID %d: %w", wfExeID, err))
	}

	var events []*entity.InterruptEvent
	for _, s := range eventJSONs {
		var event entity.InterruptEvent
		err = sonic.UnmarshalString(s, &event)
		if err != nil {
			return nil, vo.WrapError(errno.ErrSerializationDeserializationFail,
				fmt.Errorf("failed to unmarshal first interrupt event (wfExeID %d) from JSON: %w", wfExeID, err))
		}
		events = append(events, &event)
	}

	return events, nil
}

// BindConvRelatedInfo 绑定会话关联信息
//
// 将会话 ID 与工作流执行信息关联，用于通过会话查找对应的执行。
//
// 参数：
//   - ctx: 上下文
//   - convID: 会话 ID
//   - info: 关联信息
//
// 返回值：
//   - error: 绑定失败时返回错误
func (i *interruptEventStoreImpl) BindConvRelatedInfo(ctx context.Context, convID int64, info entity.ConvRelatedInfo) error {
	data, err := sonic.Marshal(info)
	if err != nil {
		return err
	}
	result := i.redis.Set(ctx, fmt.Sprintf(ConvToEventExecFormat, convID), data, interruptEventTTL)
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

// GetConvRelatedInfo 获取会话关联信息
//
// 参数：
//   - ctx: 上下文
//   - convID: 会话 ID
//
// 返回值：
//   - *entity.ConvRelatedInfo: 关联信息
//   - bool: 是否存在
//   - func() error: 删除函数，用于清理关联信息
//   - error: 获取失败时返回错误
func (i *interruptEventStoreImpl) GetConvRelatedInfo(ctx context.Context, convID int64) (*entity.ConvRelatedInfo, bool, func() error, error) {
	data, err := i.redis.Get(ctx, fmt.Sprintf(ConvToEventExecFormat, convID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil, nil
		}
		return nil, false, nil, err
	}
	rInfo := &entity.ConvRelatedInfo{}
	err = sonic.UnmarshalString(string(data), rInfo)
	if err != nil {
		return nil, false, nil, err
	}
	return rInfo, true, func() error {
		return i.redis.Del(ctx, fmt.Sprintf(ConvToEventExecFormat, convID)).Err()
	}, nil
}
