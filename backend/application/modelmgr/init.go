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

// Package modelmgr 定义了模型管理(ModelMgr)应用层服务
//
// 本包提供模型管理相关的应用层业务逻辑，包括：
// - 模型列表查询
// - 模型图标获取
// - 模型启用/禁用配置
package modelmgr

import (
	"github.com/coze-dev/coze-studio/backend/infra/storage"
)

// InitService 初始化模型管理应用服务
func InitService(tosClient storage.Storage) *ModelmgrApplicationService {
	ModelmgrApplicationSVC = &ModelmgrApplicationService{tosClient}
	return ModelmgrApplicationSVC
}
