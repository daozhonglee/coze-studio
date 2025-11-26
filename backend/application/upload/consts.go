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

// Package upload 定义了上传(Upload)应用层服务
//
// 本包提供文件上传相关的应用层业务逻辑，包括：
// - 文件大小限制
// - 默认图标资源路径
package upload

// maxFileSize 文件上传最大限制（200MB）
const maxFileSize = 200 * 1024 * 1024

// 默认图标资源路径常量
const (
	// TextKnowledgeDefaultIcon 文本知识库默认图标
	TextKnowledgeDefaultIcon = "default_icon/text_kn_default_icon.png"
	// TableKnowledgeDefaultIcon 表格知识库默认图标
	TableKnowledgeDefaultIcon = "default_icon/table_kn_default_icon.png"
	// ImageKnowledgeDefaultIcon 图片知识库默认图标
	ImageKnowledgeDefaultIcon = "default_icon/image_kn_default_icon.png"
	// DatabaseDefaultIcon 数据库默认图标
	DatabaseDefaultIcon = "default_icon/default_database_icon.png"
)
