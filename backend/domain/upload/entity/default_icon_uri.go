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

// Package entity 定义了上传(Upload)领域的核心实体（默认图标）

package entity

// 默认图标 URI 常量
//
// 各类资源的默认图标路径，存储在对象存储中
const (
	// BotIconURI 机器人默认图标
	BotIconURI = "default_icon/user_default_icon.png"
	// UserIconURI 用户默认头像
	UserIconURI = "default_icon/user_default_icon.png"
	// PluginIconURI 插件默认图标
	PluginIconURI = "default_icon/plugin_default_icon.png"
	// DatasetIconURI 数据集默认图标
	DatasetIconURI = "default_icon/plugin_default_icon.png"
	// WorkflowIconURI 工作流默认图标
	WorkflowIconURI = "default_icon/plugin_default_icon.png"
	// ImageflowIconURI 图像流默认图标
	ImageflowIconURI = "default_icon/plugin_default_icon.png"
	// SocietyIconURI 社区默认图标
	SocietyIconURI = "default_icon/plugin_default_icon.png"
	// ConnectorIconURI 连接器默认图标
	ConnectorIconURI = "default_icon/plugin_default_icon.png"
	// ChatFlowIconURI 对话流默认图标
	ChatFlowIconURI = "default_icon/plugin_default_icon.png"
	// VoiceIconURI 语音默认图标
	VoiceIconURI = "default_icon/plugin_default_icon.png"
	// EnterpriseIconURI 企业/团队默认图标
	EnterpriseIconURI = "default_icon/team_default_icon.png"
	// ModelIconURI 模型默认图标
	ModelIconURI = "default_icon/team_default_icon.png"
)

// GetDefaultShortcutIconURI 获取快捷指令可选图标列表
func GetDefaultShortcutIconURI() []string {
	return []string{
		"default_icon/shortcut_1coz_ai.png",
		"default_icon/shortcut_2icon_ai_writing.png",
		"default_icon/shortcut_3coz_imageflow.png",
		"default_icon/shortcut_4icon_aisearch.png",
		"default_icon/shortcut_5icon_summary.png",
		"default_icon/shortcut_6icon_digest.png",
		"default_icon/shortcut_7icon_video.png",
		"default_icon/shortcut_8icon_audio.png",
		"default_icon/shortcut_9coz_variables.png",
		"default_icon/shortcut_10coz_folder.png",
		"default_icon/shortcut_11coz_trans_switch.png",
		"default_icon/shortcut_12coz_location.png",
		"default_icon/shortcut_13coz_link.png",
		"default_icon/shortcut_14coz_clock.png",
	}
}
