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

// Package service 定义了上传(Upload)领域的服务层接口
//
// 本包提供文件上传领域的业务服务
package service

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/upload/entity"
)

// UploadService 上传服务接口
//
// 定义文件上传和查询的业务操作
//
//go:generate mockgen -destination ../../../internal/mock/domain/upload/upload_service_mock.go --package upload -source interface.go
type UploadService interface {
	// UploadFile 上传单个文件
	UploadFile(ctx context.Context, req *UploadFileRequest) (resp *UploadFileResponse, err error)
	// UploadFiles 批量上传文件
	UploadFiles(ctx context.Context, req *UploadFilesRequest) (resp *UploadFilesResponse, err error)
	// GetFiles 批量获取文件
	GetFiles(ctx context.Context, req *GetFilesRequest) (resp *GetFilesResponse, err error)
	// GetFile 获取单个文件
	GetFile(ctx context.Context, req *GetFileRequest) (resp *GetFileResponse, err error)
}

// UploadFileRequest 上传文件请求
type UploadFileRequest struct {
	// File 文件信息
	File *entity.File `json:"file"`
}

// UploadFileResponse 上传文件响应
type UploadFileResponse struct {
	// File 上传成功的文件信息
	File *entity.File `json:"file"`
}

// UploadFilesRequest 批量上传文件请求
type UploadFilesRequest struct {
	// Files 文件列表
	Files []*entity.File `json:"files"`
}

// UploadFilesResponse 批量上传文件响应
type UploadFilesResponse struct {
	// Files 上传成功的文件列表
	Files []*entity.File `json:"files"`
}

// GetFilesRequest 批量获取文件请求
type GetFilesRequest struct {
	// IDs 文件ID列表
	IDs []int64 `json:"ids"`
}

// GetFilesResponse 批量获取文件响应
type GetFilesResponse struct {
	// Files 文件列表
	Files []*entity.File `json:"files"`
}

// GetFileRequest 获取文件请求
type GetFileRequest struct {
	// ID 文件ID
	ID int64 `json:"id"`
}

// GetFileResponse 获取文件响应
type GetFileResponse struct {
	// File 文件信息
	File *entity.File `json:"file"`
}
