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

// Package entity 定义了 OpenAPI 认证领域的核心实体
//
// 本包包含 API 密钥(ApiKey)相关的领域实体和请求/响应结构，
// 用于管理用户的 API 访问凭证。
package entity

// ApiKey API 密钥实体
//
// 表示用户的 API 访问凭证，用于 OpenAPI 接口认证
type ApiKey struct {
	// ID 密钥唯一标识
	ID int64 `json:"id"`
	// Name 密钥名称
	Name string `json:"name"`
	// ApiKey 密钥值（加密存储）
	ApiKey string `json:"api_key"`
	// ConnectorID 关联的连接器ID
	ConnectorID int64 `json:"connector"`
	// UserID 所属用户ID
	UserID int64 `json:"user_id"`
	// LastUsedAt 最后使用时间（毫秒时间戳）
	LastUsedAt int64 `json:"last_used_at"`
	// ExpiredAt 过期时间（毫秒时间戳）
	ExpiredAt int64 `json:"expired_at"`
	// CreatedAt 创建时间（毫秒时间戳）
	CreatedAt int64 `json:"created_at"`
	// UpdatedAt 更新时间（毫秒时间戳）
	UpdatedAt int64 `json:"updated_at"`
}

// CreateApiKey 创建 API 密钥请求
type CreateApiKey struct {
	// Name 密钥名称
	Name string `json:"name"`
	// Expire 过期时长（秒）
	Expire int64 `json:"expire"`
	// UserID 用户ID
	UserID int64 `json:"user_id"`
	// AkType 密钥类型
	AkType AkType `json:"ak_type"`
}

// DeleteApiKey 删除 API 密钥请求
type DeleteApiKey struct {
	// ID 密钥ID
	ID int64 `json:"id"`
	// UserID 用户ID（用于权限校验）
	UserID int64 `json:"user_id"`
}

// GetApiKey 获取 API 密钥请求
type GetApiKey struct {
	// ID 密钥ID
	ID int64 `json:"id"`
}

// ListApiKey 列出 API 密钥请求
type ListApiKey struct {
	// UserID 用户ID
	UserID int64 `json:"user_id"`
	// Limit 每页数量
	Limit int64 `json:"limit"`
	// Page 页码
	Page int64 `json:"page"`
}

// ListApiKeyResp 列出 API 密钥响应
type ListApiKeyResp struct {
	// ApiKeys 密钥列表
	ApiKeys []*ApiKey `json:"api_keys"`
	// HasMore 是否有更多
	HasMore bool `json:"has_more"`
}

// SaveMeta 更新 API 密钥元数据请求
type SaveMeta struct {
	// ID 密钥ID
	ID int64 `json:"id"`
	// Name 密钥名称（可选更新）
	Name *string `json:"name"`
	// UserID 用户ID（用于权限校验）
	UserID int64 `json:"user_id"`
	// LastUsedAt 最后使用时间（可选更新）
	LastUsedAt *int64 `json:"last_used_at"`
}

// CheckPermission 检查权限请求
type CheckPermission struct {
	// ApiKey 待验证的密钥值
	ApiKey string `json:"api_key"`
	// UserID 用户ID
	UserID int64 `json:"user_id"`
}
