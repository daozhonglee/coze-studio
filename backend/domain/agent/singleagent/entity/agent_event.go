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

package entity

import model "github.com/coze-dev/coze-studio/backend/crossdomain/agent/model"

// AgentEvent Agent 事件，复用 crossdomain 中的定义
//
// 用于流式输出时通知客户端 Agent 的处理状态和结果。
type AgentEvent = model.AgentEvent

// InterruptEventType 中断事件类型，复用 crossdomain 中的定义
//
// 表示 Agent 执行过程中需要暂停的事件类型。
type InterruptEventType = model.InterruptEventType
