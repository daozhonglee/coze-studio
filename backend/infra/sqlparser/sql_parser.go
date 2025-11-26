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

// Package sqlparser 提供 SQL 解析器接口
//
// 本包定义 SQL 解析和修改的接口，用于：
// - SQL 语句解析
// - 表名/列名替换
// - SQL 操作类型识别
// - SQL 过滤条件追加
//
// 实现层在 impl/sqlparser/ 目录下
package sqlparser

// TableColumn 表和列名映射
type TableColumn struct {
	NewTableName *string           // if nil, not replace table name
	ColumnMap    map[string]string // Column name mapping: key is original column name, value is new column name
}

// ColumnValue 列值结构
type ColumnValue struct {
	ColName string
	Value   interface{}
}

// PrimaryKeyValue 主键值结构
type PrimaryKeyValue struct {
	ColName string
	Values  []interface{}
}

// OperationType SQL 操作类型
type OperationType string

// SQL 操作类型常量
const (
	OperationTypeSelect   OperationType = "SELECT"
	OperationTypeInsert   OperationType = "INSERT"
	OperationTypeUpdate   OperationType = "UPDATE"
	OperationTypeDelete   OperationType = "DELETE"
	OperationTypeCreate   OperationType = "CREATE"
	OperationTypeAlter    OperationType = "ALTER"
	OperationTypeDrop     OperationType = "DROP"
	OperationTypeTruncate OperationType = "TRUNCATE"
	OperationTypeUnknown  OperationType = "UNKNOWN"
)

// SQLFilterOp SQL 过滤操作符
type SQLFilterOp string

// SQL 过滤操作符常量
const (
	SQLFilterOpAnd SQLFilterOp = "AND"
	SQLFilterOpOr  SQLFilterOp = "OR"
)

// SQLParser SQL 解析器接口
type SQLParser interface {
	// ParseAndModifySQL parses SQL and replaces table/column names according to the provided message
	ParseAndModifySQL(sql string, tableColumns map[string]TableColumn) (string, error) // tableColumns Original table name -> new TableInfo

	// GetSQLOperation identifies the operation type in the SQL statement
	GetSQLOperation(sql string) (OperationType, error)

	// AddColumnsToInsertSQL adds columns to the INSERT SQL statement.
	AddColumnsToInsertSQL(origSQL string, addCols []ColumnValue, colVals *PrimaryKeyValue, isParam bool) (string, map[string]bool, error)

	// GetTableName extracts the table name from a SQL statement. Only supports single-table select/insert/update/delete. If it has multiple tables, return first table name.
	GetTableName(sql string) (string, error)

	// GetInsertDataNums extracts the number of rows to be inserted from a SQL statement. Only supports single-table insert.
	GetInsertDataNums(sql string) (int, error)

	// AppendSQLFilter appends a filter condition to the SQL statement.
	AppendSQLFilter(sql string, op SQLFilterOp, filter string) (string, error)

	// AddSelectFieldsToSelectSQL add select fields to select sql
	AddSelectFieldsToSelectSQL(origSQL string, cols []string) (string, error)
}

// New SQL 解析器工厂函数
var New func() SQLParser
