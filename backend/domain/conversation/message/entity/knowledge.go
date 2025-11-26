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

package entity

// VerboseInfo 详细信息（知识库召回等调试信息）
type VerboseInfo struct {
	// MessageType 消息类型
	MessageType string `json:"msg_type"`
	// Data 数据内容（JSON 格式）
	Data string `json:"data"`
}

// VerboseData 详细数据内容
type VerboseData struct {
	// Chunks 召回的文档片段列表
	Chunks []RecallDataInfo `json:"chunks"`
	// OriReq 原始请求
	OriReq string `json:"ori_req"`
	// StatusCode 状态码
	StatusCode int `json:"status_code"`
}

// RecallDataInfo 知识库召回数据信息
type RecallDataInfo struct {
	// Slice 文档片段内容
	Slice string `json:"slice"`
	// Score 相关性得分
	Score float64 `json:"score"`
	// Meta 元信息
	Meta MetaInfo `json:"meta"`
}

// MetaInfo 召回数据元信息
type MetaInfo struct {
	// Dataset 知识库信息
	Dataset DatasetInfo `json:"dataset"`
	// Document 文档信息
	Document DocumentInfo `json:"document"`
	// Link 链接信息
	Link LinkInfo `json:"link"`
	// Card 卡片信息
	Card CardInfo `json:"card"`
}

// DatasetInfo 知识库信息
type DatasetInfo struct {
	// ID 知识库 ID
	ID string `json:"id"`
	// Name 知识库名称
	Name string `json:"name"`
}

// DocumentInfo 文档信息
type DocumentInfo struct {
	// ID 文档 ID
	ID string `json:"id"`
	// Name 文档名称
	Name string `json:"name"`
	// FormatType 格式类型
	FormatType int32 `json:"format_type"`
	// SourceType 来源类型
	SourceType int32 `json:"source_type"`
}

// LinkInfo 链接信息
type LinkInfo struct {
	// Title 链接标题
	Title string `json:"title"`
	// URL 链接地址
	URL string `json:"url"`
}

// CardInfo 卡片信息
type CardInfo struct {
	// Title 卡片标题
	Title string `json:"title"`
	// Con 卡片内容
	Con string `json:"con"`
	// Index 索引
	Index string `json:"index"`
}
