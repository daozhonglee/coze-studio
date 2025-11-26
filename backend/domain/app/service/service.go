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

// Package service 定义了应用(APP)领域的服务层接口
//
// 本包提供应用领域的业务服务，封装核心业务逻辑：
// - AppService: 应用服务接口，提供应用的创建、管理、发布等功能
//
// 服务层是应用领域的核心，协调仓储、跨域服务等组件完成业务用例。
package service

import (
	"context"

	connectorModel "github.com/coze-dev/coze-studio/backend/crossdomain/connector/model"
	"github.com/coze-dev/coze-studio/backend/domain/app/entity"
)

// AppService 应用服务接口
//
// 定义应用领域的所有业务操作，包括：
// - 草稿应用的 CRUD 操作
// - 应用发布流程
// - 发布记录查询
type AppService interface {
	// ==================== 草稿应用操作 ====================

	// CreateDraftAPP 创建草稿应用
	CreateDraftAPP(ctx context.Context, req *CreateDraftAPPRequest) (appID int64, err error)
	// GetDraftAPP 获取草稿应用
	GetDraftAPP(ctx context.Context, appID int64) (app *entity.APP, err error)
	// DeleteDraftAPP 删除草稿应用
	DeleteDraftAPP(ctx context.Context, appID int64) (err error)
	// UpdateDraftAPP 更新草稿应用
	UpdateDraftAPP(ctx context.Context, req *UpdateDraftAPPRequest) (err error)
	// GetDraftAPPResources 获取草稿应用关联的所有资源
	GetDraftAPPResources(ctx context.Context, appID int64) (resources []*entity.Resource, err error)

	// ==================== 发布操作 ====================

	// PublishAPP 发布应用
	PublishAPP(ctx context.Context, req *PublishAPPRequest) (resp *PublishAPPResponse, err error)

	// ==================== 发布记录查询 ====================

	// GetAPPPublishRecord 获取应用发布记录
	GetAPPPublishRecord(ctx context.Context, req *GetAPPPublishRecordRequest) (record *entity.PublishRecord, published bool, err error)
	// GetAPPAllPublishRecords 获取应用的所有发布记录
	GetAPPAllPublishRecords(ctx context.Context, appID int64) (records []*entity.PublishRecord, err error)
	// GetPublishConnectorList 获取可发布的连接器列表
	GetPublishConnectorList(ctx context.Context, req *GetPublishConnectorListRequest) (resp *GetPublishConnectorListResponse, err error)
}

// CreateDraftAPPRequest 创建草稿应用请求
type CreateDraftAPPRequest struct {
	// SpaceID 工作空间ID
	SpaceID int64
	// OwnerID 所有者用户ID
	OwnerID int64
	// Name 应用名称
	Name string
	// Desc 应用描述
	Desc string
	// IconURI 图标URI
	IconURI string
}

// UpdateDraftAPPRequest 更新草稿应用请求
type UpdateDraftAPPRequest struct {
	// APPID 应用ID
	APPID int64
	// Name 新的应用名称（可选）
	Name *string
	// Desc 新的应用描述（可选）
	Desc *string
	// IconURI 新的图标URI（可选）
	IconURI *string
}

// DuplicateDraftAPPRequest 复制草稿应用请求
type DuplicateDraftAPPRequest struct {
	// APPID 要复制的应用ID
	APPID int64
	// Name 新应用名称
	Name string
	// Desc 新应用描述
	Desc string
	// IconURI 新应用图标URI
	IconURI string
}

// PublishAPPRequest 发布应用请求
type PublishAPPRequest struct {
	// APPID 要发布的应用ID
	APPID int64
	// Version 发布版本号
	Version string
	// VersionDesc 版本描述
	VersionDesc string
	// ConnectorPublishConfigs 各连接器的发布配置，key 为连接器ID
	ConnectorPublishConfigs map[int64]entity.PublishConfig
}

// PublishAPPResponse 发布应用响应
type PublishAPPResponse struct {
	// PublishRecordID 发布记录ID
	PublishRecordID int64
	// Success 是否发布成功
	Success bool
}

// GetAPPPublishRecordRequest 获取应用发布记录请求
type GetAPPPublishRecordRequest struct {
	// APPID 应用ID
	APPID int64
	// RecordID 发布记录ID（可选）
	RecordID *int64
	// Oldest 是否获取最早的记录
	// 当 Oldest 为 true 且 RecordID 为 nil 时，获取最早的记录
	// 否则获取最新的记录
	Oldest bool
}

// GetPublishConnectorListRequest 获取发布连接器列表请求
type GetPublishConnectorListRequest struct {
}

// GetPublishConnectorListResponse 获取发布连接器列表响应
type GetPublishConnectorListResponse struct {
	// Connectors 可用的连接器列表
	Connectors []*connectorModel.Connector
}
