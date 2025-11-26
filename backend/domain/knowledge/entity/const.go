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

// const.go 知识库常量定义
//
// 本文件定义了知识库领域的常量类型：
//   - DocumentStatus: 文档状态
//   - DocumentSource: 文档来源

package entity

// DocumentStatus 文档状态
type DocumentStatus int64

// 文档状态常量
const (
	DocumentStatusInit      DocumentStatus = -1 // 初始化
	DocumentStatusUploading DocumentStatus = 0  // 上传中
	DocumentStatusEnable    DocumentStatus = 1  // 生效
	DocumentStatusDisable   DocumentStatus = 2  // 失效
	DocumentStatusDeleted   DocumentStatus = 3  // 已删除
	DocumentStatusChunking  DocumentStatus = 4  // 切片中
	DocumentStatusFailed    DocumentStatus = 9  // 失败
)

func (s DocumentStatus) String() string {
	switch s {
	case DocumentStatusInit:
		return "初始化"
	case DocumentStatusUploading:
		return "上传中"
	case DocumentStatusEnable:
		return "生效"
	case DocumentStatusDisable:
		return "失效"
	case DocumentStatusDeleted:
		return "已删除"
	case DocumentStatusChunking:
		return "切片中"
	// case DocumentStatusRefreshing:
	//	Returns "Refreshing"
	case DocumentStatusFailed:
		return "失败"
	default:
		return "未知"
	}
}

// DocumentSource 文档来源
type DocumentSource int64

// 文档来源常量
const (
	DocumentSourceLocal  DocumentSource = 0 // 本地文件上传
	DocumentSourceCustom DocumentSource = 2 // 自定义文本
)
