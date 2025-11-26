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

// plugin.go 插件实体定义
//
// 本文件定义了插件领域的核心实体 PluginInfo。
// PluginInfo 封装了跨域插件模型，提供插件信息的访问和修改方法。

package entity

import (
	"github.com/coze-dev/coze-studio/backend/crossdomain/plugin/model"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
)

// PluginInfo 插件信息实体
// 封装 crossdomain/plugin/model.PluginInfo，提供领域层的访问接口
type PluginInfo struct {
	*model.PluginInfo
}

// NewPluginInfo 创建插件信息实体
func NewPluginInfo(info *model.PluginInfo) *PluginInfo {
	return &PluginInfo{
		PluginInfo: info,
	}
}

// SetName 设置插件名称
// 同时更新 Manifest 和 OpenapiDoc 中的名称字段
func (p PluginInfo) SetName(name string) {
	if p.Manifest == nil || p.OpenapiDoc == nil {
		return
	}
	p.Manifest.NameForModel = name
	p.Manifest.NameForHuman = name
	p.OpenapiDoc.Info.Title = name
}

// GetServerURL 获取插件服务器 URL
func (p PluginInfo) GetServerURL() string {
	return ptr.FromOrDefault(p.ServerURL, "")
}

// GetRefProductID 获取关联产品 ID
func (p PluginInfo) GetRefProductID() int64 {
	return ptr.FromOrDefault(p.RefProductID, 0)
}

// GetVersionDesc 获取版本描述
func (p PluginInfo) GetVersionDesc() string {
	return ptr.FromOrDefault(p.VersionDesc, "")
}
