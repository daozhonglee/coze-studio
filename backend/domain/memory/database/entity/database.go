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

// Package entity 定义了数据库记忆领域的核心实体
//
// 本包包含 Agent 数据库表管理相关的实体定义：
// - Database: 数据库表实体（复用 crossdomain 定义）
// - DatabaseFilter: 数据库查询过滤条件
// - TableSheet: Excel 工作表信息
// - TableReaderMeta: 表格读取元数据
//
// 设计说明：
// 数据库记忆功能允许 Agent 在对话中读写结构化数据表，
// 支持从 Excel/CSV 文件导入数据，并提供 SQL 查询能力。
package entity

import (
	"github.com/xuri/excelize/v2"

	"github.com/coze-dev/coze-studio/backend/api/model/data/knowledge"
	model "github.com/coze-dev/coze-studio/backend/crossdomain/database/model"
)

// Database 数据库表实体，复用 crossdomain 中的定义
// 表示 Agent 可操作的结构化数据表
type Database = model.Database

// DatabaseFilter 数据库查询过滤条件
//
// 用于列表查询时筛选数据库记录，支持按创建者、空间、表名等维度过滤。
type DatabaseFilter struct {
	// CreatorID 创建者用户 ID
	CreatorID *int64
	// SpaceID 所属空间 ID
	SpaceID *int64
	// TableName 表名（支持模糊匹配）
	TableName *string
	// AppID 所属应用 ID
	AppID *int64
}

// Pagination 分页参数
//
// 用于列表查询的分页控制。
type Pagination struct {
	// Total 记录总数（由查询结果填充）
	Total int64
	// Limit 每页记录数
	Limit int
	// Offset 偏移量（从 0 开始）
	Offset int
}

// TableSheet Excel 工作表定位信息
//
// 用于从 Excel 文件中定位特定工作表及数据起始位置。
type TableSheet struct {
	// SheetID 工作表索引（从 0 开始）
	SheetID int64
	// HeaderLineIdx 表头行索引
	HeaderLineIdx int64
	// StartLineIdx 数据起始行索引
	StartLineIdx int64
}

// TableReaderMeta 表格读取元数据
//
// 包含从文件读取表格数据所需的全部配置信息，
// 支持按头部、全量或分页方式读取。
type TableReaderMeta struct {
	// TosMaxLine 最大行数限制（防止内存溢出）
	TosMaxLine int64
	// SheetId 目标工作表索引
	SheetId int64
	// HeaderLineIdx 表头行索引
	HeaderLineIdx int64
	// StartLineIdx 数据起始行索引
	StartLineIdx int64
	// ReaderMethod 读取方式（头部/全量/分页）
	ReaderMethod model.TableReadDataMethod
	// ReadLineCnt 读取行数（分页模式使用）
	ReadLineCnt int64
	// Schema 列结构定义
	Schema []*knowledge.DocTableColumn
}

// TableReaderSheetData 工作表读取结果
//
// 包含解析后的列定义和样本数据。
type TableReaderSheetData struct {
	// Columns 列定义列表
	Columns []*knowledge.DocTableColumn
	// SampleData 样本数据（二维字符串数组）
	SampleData [][]string
}

// ExcelExtraInfo Excel 文件附加信息
//
// 包含 Excel 文件的元数据，如工作表列表、文件大小等。
type ExcelExtraInfo struct {
	// Sheets 工作表列表
	Sheets []*knowledge.DocTableSheet
	// ExtensionName 文件扩展名（xlsx/xls/csv）
	ExtensionName string
	// FileSize 文件大小（字节）
	FileSize int64
	// SourceFileID 源文件 ID
	SourceFileID int64
	// TosURI 对象存储 URI
	TosURI string
}

// LocalTableMeta 本地表格元数据
//
// 解析本地文件后的中间结构，支持 xlsx/xls/csv 格式。
type LocalTableMeta struct {
	// ExcelFile xlsx 格式的 Excel 文件对象
	ExcelFile *excelize.File
	// RawLines csv/xls 格式的原始行数据
	RawLines [][]string
	// SheetsNameList 工作表名称列表
	SheetsNameList []string
	// SheetsRowCount 每个工作表的行数
	SheetsRowCount []int
	// ExtensionName 文件扩展名
	ExtensionName string
	// FileSize 文件大小（字节）
	FileSize int64
}

// ColumnInfo 列类型信息
//
// 用于类型推断结果的表示。
type ColumnInfo struct {
	// ColumnType 推断出的列类型
	ColumnType knowledge.ColumnType
	// ContainsEmptyValue 是否包含空值
	ContainsEmptyValue bool
}

// SelectFieldList SQL 查询字段列表
//
// 用于指定 SELECT 查询要返回的字段。
type SelectFieldList struct {
	// FieldID 字段 ID 列表
	FieldID []string
	// IsDistinct 是否去重
	IsDistinct bool
}
