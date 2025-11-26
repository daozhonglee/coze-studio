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

// Package shortcutcmd 定义了快捷命令(ShortcutCmd)应用层服务
//
// 本包提供快捷命令相关的应用层业务逻辑，包括：
// - 快捷命令的创建、更新、删除
// - Agent 快捷命令配置
//
// 快捷命令是用户与 Agent 交互的快捷入口
package shortcutcmd

import (
	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/shortcutcmd/repository"
	"github.com/coze-dev/coze-studio/backend/domain/shortcutcmd/service"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
)

// ShortcutCmdSVC 快捷命令应用服务单例
var ShortcutCmdSVC *ShortcutCmdApplicationService

// InitService 初始化快捷命令应用服务
func InitService(db *gorm.DB, idGenSVC idgen.IDGenerator) *ShortcutCmdApplicationService {

	components := &service.Components{
		ShortCutCmdRepo: repository.NewShortCutCmdRepo(db, idGenSVC),
	}
	shortcutCmdDomainSVC := service.NewShortcutCommandService(components)

	ShortcutCmdSVC = &ShortcutCmdApplicationService{
		ShortCutDomainSVC: shortcutCmdDomainSVC,
	}
	return ShortcutCmdSVC
}
