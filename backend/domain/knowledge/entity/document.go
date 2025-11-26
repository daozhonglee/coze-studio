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

// document.go 知识库文档实体定义
//
// 本文件定义了知识库文档相关的实体结构：
//   - Document: 文档实体
//   - TableInfo: 表格信息
//   - TableColumn: 表格列定义

package entity

import (
	"github.com/coze-dev/coze-studio/backend/crossdomain/knowledge/model"
	"github.com/coze-dev/coze-studio/backend/infra/document"
	"github.com/coze-dev/coze-studio/backend/infra/document/parser"
)

// Document 知识库文档实体
// 包含文档的元信息、内容、分片策略等
type Document struct {
	model.Info

	KnowledgeID      int64
	Type             model.DocumentType
	RawContent       string               // User-defined original content
	URI              string               // Document URI
	URL              string               // Document URL
	Size             int64                // Document bytes
	SliceCount       int64                // Number of slices
	CharCount        int64                // Number of document characters
	FileExtension    parser.FileExtension // Document suffix, csv/pdf...
	Status           DocumentStatus       // Document Status
	StatusMsg        string               // Document Status Details
	Hits             int64                // hit count
	Source           DocumentSource       // document source
	ParsingStrategy  *ParsingStrategy     // parsing strategy
	ChunkingStrategy *ChunkingStrategy    // segmentation strategy

	TableInfo TableInfo
	IsAppend  bool // Whether to append to the table

	// LevelURI string//Hierarchical segmentation preview uri
	// PreviewURI string//preview uri
}

// TableInfo 表格信息结构
type TableInfo struct {
	VirtualTableName  string         `json:"virtual_table_name"`
	PhysicalTableName string         `json:"physical_table_name"`
	TableDesc         string         `json:"table_desc"`
	Columns           []*TableColumn `json:"columns"`
}

// TableSheet 表格工作表信息
type TableSheet struct {
	SheetId       int64  // sheet id
	HeaderLineIdx int64  // header row
	StartLineIdx  int64  // Data start row
	SheetName     string // Name of sheet
	TotalRows     int64  // total number of rows
}

// TableColumn 表格列定义
type TableColumn struct {
	ID          int64
	Name        string
	Type        document.TableColumnType
	Description string
	Indexing    bool  // whether to index
	Sequence    int64 // The original serial number in the table
}

// WhereDocumentOpt 文档查询条件选项
type WhereDocumentOpt struct {
	IDs          []int64
	KnowledgeIDs []int64
	StatusIn     []int32
	StatusNotIn  []int32
	CreatorID    int64
	Name         *string
	Limit        int
	Offset       *int
	Cursor       *string
	SelectAll    bool
}
