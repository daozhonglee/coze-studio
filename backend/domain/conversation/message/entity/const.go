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

// Package entity 定义了消息领域的核心实体和常量
//
// 本包包含消息管理相关的实体定义：
// - Message: 消息实体
// - MessageStatus: 消息状态
// - ScrollPageDirection: 滚动分页方向
//
// 设计说明：
// 消息是对话中的基本单元，每条消息属于一个运行记录。
// 支持多种消息类型和状态管理。
package entity

// ScrollPageDirection 滚动分页方向枚举
type ScrollPageDirection string

// 滚动分页方向常量
const (
	// ScrollPageDirectionPrev 向上滚动（加载更早的消息）
	ScrollPageDirectionPrev ScrollPageDirection = "up"
	// ScrollPageDirectionNext 向下滚动（加载更新的消息）
	ScrollPageDirectionNext ScrollPageDirection = "down"
)

// MessageStatus 消息状态枚举
type MessageStatus int32

// 消息状态常量
const (
	// MessageStatusAvailable 可用
	MessageStatusAvailable MessageStatus = 1
	// MessageStatusDeleted 已删除
	MessageStatusDeleted MessageStatus = 2
	// MessageStatusBroken 中断（被打断的消息）
	MessageStatusBroken MessageStatus = 4
)
