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

// EventType 中断事件类型枚举
//
// 表示 Agent 运行过程中需要暂停等待外部输入的事件类型。
type EventType int64

// 中断事件类型常量
const (
	// EventType_LocalPlugin 本地插件调用
	EventType_LocalPlugin EventType = 1
	// EventType_Question 需要用户回答问题
	EventType_Question EventType = 2
	// EventType_RequireInfos 需要补充信息
	EventType_RequireInfos EventType = 3
	// EventType_SceneChat 场景对话
	EventType_SceneChat EventType = 4
	// EventType_InputNode 工作流输入节点
	EventType_InputNode EventType = 5
	// EventType_WorkflowLocalPlugin 工作流本地插件
	EventType_WorkflowLocalPlugin EventType = 6
	// EventType_OauthPlugin OAuth 插件授权
	EventType_OauthPlugin EventType = 7
	// EventType_WorkflowLLM 工作流大模型调用
	EventType_WorkflowLLM EventType = 100
)
