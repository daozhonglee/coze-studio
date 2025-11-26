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

// Package taskgroup 提供并发任务组工具
//
// 本包提供并发任务执行能力：
// - 可中断任务组（一个失败则全部停止）
// - 不可中断任务组（一个失败继续执行其他）
// - 并发数量限制
// - panic 恢复
package taskgroup

import (
	"context"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"github.com/coze-dev/coze-studio/backend/pkg/logs"
)

// TaskGroup 任务组接口
type TaskGroup interface {
	Go(f func() error)
	Wait() error
}

// taskGroup 任务组实现
type taskGroup struct {
	errGroup    *errgroup.Group
	ctx         context.Context
	execAllTask atomic.Bool
}

// NewTaskGroup 创建可中断任务组
//
// 一个任务失败则其他任务停止执行
func NewTaskGroup(ctx context.Context, concurrentCount int) TaskGroup {
	t := &taskGroup{}
	t.errGroup, t.ctx = errgroup.WithContext(ctx)
	t.errGroup.SetLimit(concurrentCount)
	t.execAllTask.Store(false)

	return t
}

// NewUninterruptibleTaskGroup 创建不可中断任务组
//
// 一个任务失败其他任务继续执行
func NewUninterruptibleTaskGroup(ctx context.Context, concurrentCount int) TaskGroup {
	t := &taskGroup{}
	t.errGroup, t.ctx = errgroup.WithContext(ctx)
	t.errGroup.SetLimit(concurrentCount)
	t.execAllTask.Store(true)

	return t
}

// Go 添加任务到任务组
func (t *taskGroup) Go(f func() error) {
	t.errGroup.Go(func() error {
		defer func() {
			if err := recover(); err != nil {
				logs.CtxErrorf(t.ctx, "[TaskGroup] exec panic recover:%+v", err)
			}
		}()

		if !t.execAllTask.Load() {
			select {
			case <-t.ctx.Done():
				return t.ctx.Err()
			default:
			}
		}

		return f()
	})
}

// Wait 等待所有任务完成
func (t *taskGroup) Wait() error {
	return t.errGroup.Wait()
}
