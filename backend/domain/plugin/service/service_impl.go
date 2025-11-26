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

// service_impl.go 插件服务实现
//
// 本文件提供插件服务的主入口和组件初始化。

package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/plugin/internal/dal"
	"github.com/coze-dev/coze-studio/backend/domain/plugin/repository"
	"github.com/coze-dev/coze-studio/backend/infra/cache"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
	"github.com/coze-dev/coze-studio/backend/pkg/safego"
)

// Components 插件服务依赖组件
type Components struct {
	IDGen      idgen.IDGenerator
	DB         *gorm.DB
	OSS        storage.Storage
	CacheCli   cache.Cmdable
	PluginRepo repository.PluginRepository
	ToolRepo   repository.ToolRepository
	OAuthRepo  repository.OAuthRepository
}

// NewService 创建插件服务实例
func NewService(components *Components) PluginService {
	impl := &pluginServiceImpl{
		db:         components.DB,
		oss:        components.OSS,
		pluginRepo: components.PluginRepo,
		toolRepo:   components.ToolRepo,
		oauthRepo:  components.OAuthRepo,
		oauthCache: dal.NewOAuthCache(components.CacheCli),
	}

	initOnce.Do(func() {
		ctx := context.Background()
		safego.Go(ctx, func() {
			impl.processOAuthAccessToken(ctx)
		})
	})

	return impl
}

// pluginServiceImpl 插件服务实现
type pluginServiceImpl struct {
	db         *gorm.DB
	oss        storage.Storage
	pluginRepo repository.PluginRepository
	toolRepo   repository.ToolRepository
	oauthRepo  repository.OAuthRepository
	oauthCache *dal.OAuthCache
}
