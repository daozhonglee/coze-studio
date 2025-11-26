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

// Package datacopy 定义了数据复制(DataCopy)领域的服务层
//
// 本包提供数据复制服务，用于在不同空间或应用间复制数据，
// 支持知识库、数据库、变量等类型的数据复制。
package datacopy

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/datacopy/entity"
)

// DataCopy 数据复制服务接口
//
// 定义数据复制任务的管理操作
type DataCopy interface {
	// CheckAndGenCopyTask 检查并生成复制任务
	// 如果任务已存在，返回现有任务状态；否则创建新任务
	CheckAndGenCopyTask(ctx context.Context, req *CheckAndGenCopyTaskReq) (*CheckAndGenCopyTaskResp, error)
	// UpdateCopyTask 更新复制任务状态
	UpdateCopyTask(ctx context.Context, req *UpdateCopyTaskReq) error
	// UpdateCopyTaskWithTX 在事务中更新复制任务状态
	UpdateCopyTaskWithTX(ctx context.Context, req *UpdateCopyTaskReq, tx *gorm.DB) error
}

// CheckAndGenCopyTaskReq 检查并生成复制任务请求
type CheckAndGenCopyTaskReq struct {
	// Task 复制任务信息
	Task *entity.CopyDataTask
}

// CheckAndGenCopyTaskResp 检查并生成复制任务响应
type CheckAndGenCopyTaskResp struct {
	// CopyTaskStatus 任务状态
	CopyTaskStatus entity.DataCopyTaskStatus
	// FailReason 失败原因
	FailReason string
	// TargetID 目标数据ID
	TargetID int64
}

// UpdateCopyTaskReq 更新复制任务请求
type UpdateCopyTaskReq struct {
	// Task 复制任务信息
	Task *entity.CopyDataTask
}
