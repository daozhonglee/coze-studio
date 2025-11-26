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

// Package entity 定义了用户(User)领域的核心实体
//
// 本包包含用户相关的所有领域实体和值对象：
// - User: 用户聚合根，表示系统中的用户
// - UserBenefit: 用户权益信息
// - SaasUserData: SaaS 用户数据（来自外部 API）
//
// 用户是系统的核心实体，用于身份认证、权限控制和资源归属。
package entity

// User 用户聚合根，表示系统中的一个用户
//
// 包含用户的基本信息、认证状态和会话信息
type User struct {
	// UserID 用户唯一标识
	UserID int64

	// Name 用户昵称
	Name string
	// UniqueName 用户唯一名称（可用于 @提及）
	UniqueName string
	// Email 用户邮箱（用于登录）
	Email string
	// Description 用户描述/简介
	Description string
	// IconURI 头像的存储路径
	IconURI string
	// IconURL 头像的访问 URL
	IconURL string
	// UserVerified 用户是否已验证
	UserVerified bool
	// Locale 用户语言/地区设置
	Locale string
	// SessionKey 当前会话密钥
	SessionKey string

	// CreatedAt 创建时间（毫秒时间戳）
	CreatedAt int64
	// UpdatedAt 更新时间（毫秒时间戳）
	UpdatedAt int64
}

// UserBenefit 用户权益信息
//
// 记录用户的订阅等级和资源使用情况
type UserBenefit struct {
	// UserID 用户ID
	UserID int64
	// UserLevel 用户等级
	UserLevel UserLevel
	// UsedCount 已使用次数
	UsedCount int32
	// TotalCount 总可用次数
	TotalCount int32
	// IsUnlimited 是否无限制
	IsUnlimited bool
	// ResetDatetime 重置时间（秒时间戳）
	ResetDatetime int64
	// CallQPS 调用 QPS 限制
	CallQPS int32
}

// SaasUserData SaaS 用户数据
//
// 来自外部 SaaS API 的用户信息
type SaasUserData struct {
	// UserID 用户ID
	UserID string `json:"user_id"`
	// UserName 用户名
	UserName string `json:"user_name"`
	// NickName 昵称
	NickName string `json:"nick_name"`
	// AvatarURL 头像 URL
	AvatarURL string `json:"avatar_url"`
}
