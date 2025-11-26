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

// workflow_reference.go 工作流引用关系实体
//
// 本文件定义了工作流之间的引用关系实体。
// 用于追踪子工作流、工作流工具等引用关系。

package entity

import (
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
)

// WorkflowReference 工作流引用关系
// 记录工作流之间的引用/被引用关系
type WorkflowReference struct {
	ID int64
	WorkflowReferenceKey
	CreatedAt time.Time
	Enabled   bool
}

// WorkflowReferenceKey 工作流引用关系键
type WorkflowReferenceKey struct {
	ReferredID  int64
	ReferringID int64
	vo.ReferType
	vo.ReferringBizType
}
