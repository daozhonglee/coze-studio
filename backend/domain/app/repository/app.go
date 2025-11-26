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

// Package repository 定义了应用(APP)领域的仓储接口
//
// 本包提供应用数据的持久化操作抽象，遵循 DDD 仓储模式：
// - AppRepository: 应用仓储接口，定义所有应用相关的数据访问方法
//
// 仓储接口隔离了领域层与基础设施层，使得领域逻辑不依赖具体的存储实现。
package repository

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/app/entity"
)

// AppRepository 应用仓储接口
//
// 定义应用领域的所有数据访问方法，包括：
// - 草稿应用的 CRUD 操作
// - 发布记录的管理
// - 资源复制任务的管理
type AppRepository interface {
	// ==================== 草稿应用操作 ====================

	// CreateDraftAPP 创建草稿应用
	CreateDraftAPP(ctx context.Context, app *entity.APP) (appID int64, err error)
	// GetDraftAPP 获取草稿应用
	GetDraftAPP(ctx context.Context, appID int64) (app *entity.APP, exist bool, err error)
	// CheckDraftAPPExist 检查草稿应用是否存在
	CheckDraftAPPExist(ctx context.Context, appID int64) (exist bool, err error)
	// DeleteDraftAPP 删除草稿应用
	DeleteDraftAPP(ctx context.Context, appID int64) (err error)
	// UpdateDraftAPP 更新草稿应用
	UpdateDraftAPP(ctx context.Context, app *entity.APP) (err error)

	// ==================== 发布记录操作 ====================

	// GetPublishRecord 获取发布记录
	GetPublishRecord(ctx context.Context, req *GetPublishRecordRequest) (record *entity.PublishRecord, exist bool, err error)
	// CheckAPPVersionExist 检查应用版本是否已存在
	CheckAPPVersionExist(ctx context.Context, appID int64, version string) (exist bool, err error)
	// CreateAPPPublishRecord 创建应用发布记录
	CreateAPPPublishRecord(ctx context.Context, record *entity.PublishRecord) (recordID int64, err error)
	// UpdateAPPPublishStatus 更新应用发布状态
	UpdateAPPPublishStatus(ctx context.Context, req *UpdateAPPPublishStatusRequest) (err error)
	// UpdateConnectorPublishStatus 更新连接器发布状态
	UpdateConnectorPublishStatus(ctx context.Context, recordID int64, status entity.ConnectorPublishStatus) (err error)
	// GetAPPAllPublishRecords 获取应用的所有发布记录
	GetAPPAllPublishRecords(ctx context.Context, appID int64, opts ...APPSelectedOptions) (records []*entity.PublishRecord, err error)

	// ==================== 资源复制任务操作 ====================

	// InitResourceCopyTask 初始化资源复制任务
	InitResourceCopyTask(ctx context.Context, result *entity.ResourceCopyResult) (taskID string, err error)
	// SaveResourceCopyTaskResult 保存资源复制任务结果
	SaveResourceCopyTaskResult(ctx context.Context, taskID string, result *entity.ResourceCopyResult) (err error)
	// GetResourceCopyTaskResult 获取资源复制任务结果
	GetResourceCopyTaskResult(ctx context.Context, taskID string) (result *entity.ResourceCopyResult, exist bool, err error)
}

// GetPublishRecordRequest 获取发布记录请求
type GetPublishRecordRequest struct {
	// APPID 应用ID
	APPID int64
	// RecordID 发布记录ID（可选）
	RecordID *int64
	// OldestSuccess 是否获取最早的成功记录
	// 当 OldestSuccess 为 true 且 RecordID 为 nil 时，获取最早的成功记录
	// 否则获取最新的记录
	OldestSuccess bool
}

// UpdateAPPPublishStatusRequest 更新应用发布状态请求
type UpdateAPPPublishStatusRequest struct {
	// RecordID 发布记录ID
	RecordID int64
	// PublishStatus 新的发布状态
	PublishStatus entity.PublishStatus
	// PublishRecordExtraInfo 发布记录的额外信息（可选）
	PublishRecordExtraInfo *entity.PublishRecordExtraInfo
}
