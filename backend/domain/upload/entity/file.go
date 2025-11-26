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

// Package entity 定义了上传(Upload)领域的核心实体
//
// 本包包含文件上传相关的领域实体：
// - File: 文件实体，表示上传的文件信息
//
// 文件上传服务用于管理用户上传的各类文件，如图片、文档等。
package entity

// File 文件实体
//
// 表示一个上传的文件，包含文件的元数据和存储信息
type File struct {
	// ID 文件唯一标识
	ID int64 `json:"id"`
	// Name 文件名称
	Name string `json:"name"`
	// FileSize 文件大小（字节）
	FileSize int64 `json:"file_size"`
	// TosURI 对象存储路径
	TosURI string `json:"tos_uri"`
	// Status 文件状态
	Status FileStatus `json:"status"`
	// Comment 文件备注
	Comment string `json:"comment"`
	// Source 文件来源
	Source FileSource `json:"source"`
	// CreatorID 上传者ID
	CreatorID string `json:"creator_id"`
	// CozeAccountID 关联的 Coze 账户ID
	CozeAccountID int64 `json:"coze_account_id"`
	// ContentType 文件 MIME 类型
	ContentType string `json:"content_type"`
	// CreatedAt 创建时间（毫秒时间戳）
	CreatedAt int64 `json:"created_at"`
	// UpdatedAt 更新时间（毫秒时间戳）
	UpdatedAt int64 `json:"updated_at"`
	// Url 文件访问 URL（运行时填充）
	Url string `json:"url"`
}

// FileStatus 文件状态
type FileStatus int32

const (
	// FileStatusInvalid 无效文件
	FileStatusInvalid FileStatus = 0
	// FileStatusValid 有效文件
	FileStatusValid FileStatus = 1
)

// FileSource 文件来源
type FileSource int32

const (
	// FileSourceAPI 通过 API 上传
	FileSourceAPI FileSource = 1
)
