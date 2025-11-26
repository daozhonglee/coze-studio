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

// Package convertor 提供数据库记忆领域的数据转换功能
//
// 本包负责以下转换操作：
// - 查询结果集到业务数据的转换
// - 物理字段名到逻辑字段名的映射
// - 数据类型转换（字符串/数值/日期等）
// - SQL 操作符转换
//
// 设计说明：
// 由于物理表使用自动生成的列名（如 f_1, f_2），
// 需要在查询结果返回时转换为用户定义的逻辑字段名。
package convertor

import (
	"fmt"

	"github.com/coze-dev/coze-studio/backend/api/model/data/database/table"
	"github.com/coze-dev/coze-studio/backend/infra/rdb/entity"
)

// ConvertResultSetToString 将查询结果集转换为字符串格式的记录列表
//
// 该函数执行以下转换：
// 1. 将物理列名转换为逻辑字段名
// 2. 根据字段类型将值转换为字符串格式
// 3. 处理空值和系统字段
//
// 参数：
//   - resultSet: 数据库查询结果集
//   - physicalToFieldName: 物理列名到逻辑字段名的映射
//   - physicalToFieldType: 物理列名到字段类型的映射
//
// 返回值：
//   - 字符串格式的记录列表
func ConvertResultSetToString(resultSet *entity.ResultSet, physicalToFieldName map[string]string, physicalToFieldType map[string]table.FieldItemType) []map[string]string {
	records := make([]map[string]string, 0, len(resultSet.Rows))

	for _, row := range resultSet.Rows {
		record := make(map[string]string)

		for physicalName, value := range row {
			if logicalName, exists := physicalToFieldName[physicalName]; exists {
				if value == nil {
					record[logicalName] = ""
				} else {
					fieldType, hasType := physicalToFieldType[physicalName]
					if hasType {
						convertedValue := ConvertDBValueToString(value, fieldType)
						record[logicalName] = convertedValue
					} else {
						record[logicalName] = fmt.Sprintf("%v", value)
					}
				}
			} else {
				if value == nil {
					record[physicalName] = ""
				} else {
					record[physicalName] = ConvertSystemFieldToString(physicalName, value)
				}
			}
		}
		records = append(records, record)
	}

	return records
}

// ConvertResultSet 将查询结果集转换为原始类型的记录列表
//
// 与 ConvertResultSetToString 不同，该函数保留值的原始类型。
// 主要用于 SQL 执行结果的返回。
//
// 参数：
//   - resultSet: 数据库查询结果集
//   - physicalToFieldName: 物理列名到逻辑字段名的映射
//   - physicalToFieldType: 物理列名到字段类型的映射（未使用）
//
// 返回值：
//   - 保留原始类型的记录列表
func ConvertResultSet(resultSet *entity.ResultSet, physicalToFieldName map[string]string, physicalToFieldType map[string]table.FieldItemType) []map[string]any {
	records := make([]map[string]any, 0, len(resultSet.Rows))

	for _, row := range resultSet.Rows {
		record := make(map[string]any)

		for physicalName, value := range row {
			if logicalName, exists := physicalToFieldName[physicalName]; exists {
				record[logicalName] = value
			} else {
				record[physicalName] = value
			}
		}
		records = append(records, record)
	}

	return records
}
