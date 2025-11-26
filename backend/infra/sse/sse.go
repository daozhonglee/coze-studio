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

// Package sse 提供 Server-Sent Events 接口
//
// 本包定义 SSE 服务的接口，用于向客户端推送实时事件：
// - 工作流执行进度
// - 对话消息流
// - 实时通知
package sse

import (
	"context"

	"github.com/hertz-contrib/sse"
)

// SSender SSE 发送器接口
type SSender interface {
	Send(ctx context.Context, s *sse.Stream, event *sse.Event) error
}
