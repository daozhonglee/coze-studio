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

// Package repository 定义了用户(User)领域的仓储接口
//
// 本包提供用户数据的持久化操作抽象，遵循 DDD 仓储模式：
// - UserRepository: 用户仓储接口
// - SpaceRepository: 工作空间仓储接口
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/user/internal/dal"
	"github.com/coze-dev/coze-studio/backend/domain/user/internal/dal/model"
)

// NewUserRepo 创建用户仓储实例
func NewUserRepo(db *gorm.DB) UserRepository {
	return dal.NewUserDAO(db)
}

// NewSpaceRepo 创建工作空间仓储实例
func NewSpaceRepo(db *gorm.DB) SpaceRepository {
	return dal.NewSpaceDAO(db)
}

// UserRepository 用户仓储接口
//
// 定义用户相关的数据访问方法
type UserRepository interface {
	// GetUsersByEmail 根据邮箱获取用户
	GetUsersByEmail(ctx context.Context, email string) (*model.User, bool, error)
	// UpdateSessionKey 更新用户会话密钥
	UpdateSessionKey(ctx context.Context, userID int64, sessionKey string) error
	// ClearSessionKey 清除用户会话密钥
	ClearSessionKey(ctx context.Context, userID int64) error
	// UpdatePassword 更新用户密码
	UpdatePassword(ctx context.Context, email, password string) error
	// GetUserByID 根据ID获取用户
	GetUserByID(ctx context.Context, userID int64) (*model.User, error)
	// UpdateAvatar 更新用户头像
	UpdateAvatar(ctx context.Context, userID int64, iconURI string) error
	// CheckUniqueNameExist 检查唯一名称是否已存在
	CheckUniqueNameExist(ctx context.Context, uniqueName string) (bool, error)
	// UpdateProfile 更新用户资料
	UpdateProfile(ctx context.Context, userID int64, updates map[string]any) error
	// CheckEmailExist 检查邮箱是否已存在
	CheckEmailExist(ctx context.Context, email string) (bool, error)
	// CreateUser 创建用户
	CreateUser(ctx context.Context, user *model.User) error
	// GetUserBySessionKey 根据会话密钥获取用户
	GetUserBySessionKey(ctx context.Context, sessionKey string) (*model.User, bool, error)
	// GetUsersByIDs 批量获取用户
	GetUsersByIDs(ctx context.Context, userIDs []int64) ([]*model.User, error)
}

// SpaceRepository 工作空间仓储接口
//
// 定义工作空间相关的数据访问方法
type SpaceRepository interface {
	// CreateSpace 创建工作空间
	CreateSpace(ctx context.Context, space *model.Space) error
	// GetSpaceByIDs 批量获取工作空间
	GetSpaceByIDs(ctx context.Context, spaceIDs []int64) ([]*model.Space, error)
	// AddSpaceUser 添加空间成员
	AddSpaceUser(ctx context.Context, spaceUser *model.SpaceUser) error
	// GetSpaceList 获取用户的工作空间列表
	GetSpaceList(ctx context.Context, userID int64) ([]*model.SpaceUser, error)
}
