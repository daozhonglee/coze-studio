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

// Package service 定义了用户(User)领域的服务层接口
//
// 本包提供用户领域的业务服务，封装核心业务逻辑：
// - User: 用户服务接口，提供用户注册、登录、资料管理等功能
// - SaasUserProvider: SaaS 用户信息提供者接口
package service

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/user/entity"
)

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	// UserID 用户ID
	UserID int64
	// Name 新昵称（可选）
	Name *string
	// UniqueName 新唯一名称（可选）
	UniqueName *string
	// Description 新描述（可选）
	Description *string
	// Locale 新语言设置（可选）
	Locale *string
}

// ValidateProfileUpdateRequest 验证资料更新请求
type ValidateProfileUpdateRequest struct {
	// UniqueName 要验证的唯一名称（可选）
	UniqueName *string
	// Email 要验证的邮箱（可选）
	Email *string
}

// ValidateProfileUpdateResult 资料更新验证结果
type ValidateProfileUpdateResult int

const (
	// ValidateSuccess 验证成功
	ValidateSuccess ValidateProfileUpdateResult = 0
	// UniqueNameExist 唯一名称已存在
	UniqueNameExist ValidateProfileUpdateResult = 2
	// UniqueNameTooShortOrTooLong 唯一名称长度不符合要求
	UniqueNameTooShortOrTooLong ValidateProfileUpdateResult = 3
	// EmailExist 邮箱已存在
	EmailExist ValidateProfileUpdateResult = 5
)

// ValidateProfileUpdateResponse 资料更新验证响应
type ValidateProfileUpdateResponse struct {
	// Code 验证结果码
	Code ValidateProfileUpdateResult
	// Msg 验证结果消息
	Msg string
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	// Email 用户邮箱
	Email string
	// Password 用户密码
	Password string
	// Name 用户昵称
	Name string
	// UniqueName 用户唯一名称
	UniqueName string
	// Description 用户描述
	Description string
	// SpaceID 初始工作空间ID（如果为0则自动创建个人空间）
	SpaceID int64
	// Locale 用户语言设置
	Locale string
}

// CreateUserResponse 创建用户响应
type CreateUserResponse struct {
	// UserID 新创建的用户ID
	UserID int64
}

// User 用户服务接口
//
// 定义用户领域的所有业务操作，包括：
// - 用户注册、登录、登出
// - 密码管理
// - 资料管理
// - 会话验证
type User interface {
	SaasUserProvider

	// Create 创建或注册新用户
	Create(ctx context.Context, req *CreateUserRequest) (user *entity.User, err error)
	// Login 用户登录
	Login(ctx context.Context, email, password string) (user *entity.User, err error)
	// Logout 用户登出
	Logout(ctx context.Context, userID int64) (err error)
	// ResetPassword 重置密码
	ResetPassword(ctx context.Context, email, password string) (err error)
	// GetUserInfo 获取用户信息
	GetUserInfo(ctx context.Context, userID int64) (user *entity.User, err error)
	// UpdateAvatar 更新用户头像
	UpdateAvatar(ctx context.Context, userID int64, ext string, imagePayload []byte) (url string, err error)
	// UpdateProfile 更新用户资料
	UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (err error)
	// ValidateProfileUpdate 验证资料更新
	ValidateProfileUpdate(ctx context.Context, req *ValidateProfileUpdateRequest) (resp *ValidateProfileUpdateResponse, err error)
	// GetUserProfiles 获取用户资料
	GetUserProfiles(ctx context.Context, userID int64) (user *entity.User, err error)
	// MGetUserProfiles 批量获取用户资料
	MGetUserProfiles(ctx context.Context, userIDs []int64) (users []*entity.User, err error)
	// ValidateSession 验证会话
	ValidateSession(ctx context.Context, sessionKey string) (session *entity.Session, exist bool, err error)
	// GetUserSpaceList 获取用户的工作空间列表
	GetUserSpaceList(ctx context.Context, userID int64) (spaces []*entity.Space, err error)
	GetUserSpaceBySpaceID(ctx context.Context, spaceID []int64) (space []*entity.Space, err error)
}

// SaasUserProvider SaaS 用户信息提供者接口
//
// 用于从外部 SaaS 平台获取用户信息和权益
type SaasUserProvider interface {
	// GetSaasUserInfo 获取 SaaS 用户信息
	GetSaasUserInfo(ctx context.Context) (user *entity.SaasUserData, err error)
	// GetUserBenefit 获取用户权益信息
	GetUserBenefit(ctx context.Context) (benefit *entity.UserBenefit, err error)
}
