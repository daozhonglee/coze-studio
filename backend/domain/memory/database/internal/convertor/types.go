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

package convertor

import (
	"fmt"
	"strconv"
	"time"

	"github.com/coze-dev/coze-studio/backend/api/model/data/database/table"
	database "github.com/coze-dev/coze-studio/backend/crossdomain/database/model"
	"github.com/coze-dev/coze-studio/backend/infra/rdb/entity"
)

const (
	// TimeFormat 标准时间格式，用于日期类型字段的格式化
	TimeFormat = "2006-01-02 15:04:05"
)

// SwitchToDataType 将业务字段类型转换为数据库数据类型
//
// 用于创建物理表时确定列的数据类型。
//
// 类型映射：
//   - Text -> TEXT（长文本）
//   - Number -> BIGINT（64位整数）
//   - Date -> TIMESTAMP（时间戳）
//   - Float -> DOUBLE（双精度浮点）
//   - Boolean -> BOOLEAN（布尔值）
//   - 其他 -> VARCHAR（变长字符串）
func SwitchToDataType(itemType table.FieldItemType) entity.DataType {
	switch itemType {
	case table.FieldItemType_Text:
		return entity.TypeText
	case table.FieldItemType_Number:
		return entity.TypeBigInt
	case table.FieldItemType_Date:
		return entity.TypeTimestamp
	case table.FieldItemType_Float:
		return entity.TypeDouble
	case table.FieldItemType_Boolean:
		return entity.TypeBoolean
	default:
		// VARCHAR is used by default
		return entity.TypeVarchar
	}
}

// ConvertValueByType 将字符串值转换为指定的字段类型
//
// 用于数据插入/更新时将用户输入的字符串转换为对应的数据库类型。
// 空字符串返回 nil，表示数据库中的 NULL 值。
//
// 参数：
//   - value: 待转换的字符串值
//   - fieldType: 目标字段类型
//
// 返回值：
//   - 转换后的值（类型与 fieldType 对应）
//   - error: 转换失败时的错误信息
func ConvertValueByType(value string, fieldType table.FieldItemType) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	switch fieldType {
	case table.FieldItemType_Number:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert %s to number", value)
		}

		return intVal, nil

	case table.FieldItemType_Float:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal, nil
		}

		return 0.0, fmt.Errorf("cannot convert %s to float", value)

	case table.FieldItemType_Boolean:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal, nil
		}

		// if err, try 0/1
		if value == "0" {
			return false, nil
		}
		if value == "1" {
			return true, nil
		}

		return false, fmt.Errorf("cannot convert %s to boolean", value)

	case table.FieldItemType_Date:
		t, err := time.Parse(TimeFormat, value) // database use this format
		if err != nil {
			return "", fmt.Errorf("cannot convert %s to date", value)
		}

		return t, nil

	case table.FieldItemType_Text:
		return value, nil

	default:
		return value, nil
	}
}

// ConvertDBValueToString 将数据库值转换为字符串
//
// 用于查询结果返回时将数据库中的各种类型值转换为字符串格式。
// 处理了不同数据库驱动返回的不同类型（如 []uint8 vs string）。
//
// 参数：
//   - value: 数据库返回的原始值
//   - fieldType: 字段类型（用于确定格式化方式）
//
// 返回值：
//   - 格式化后的字符串
func ConvertDBValueToString(value interface{}, fieldType table.FieldItemType) string {
	switch fieldType {
	case table.FieldItemType_Text:
		if byteArray, ok := value.([]uint8); ok {
			return string(byteArray)
		}

	case table.FieldItemType_Number:
		switch v := value.(type) {
		case int64:
			return strconv.FormatInt(v, 10)
		case []uint8:
			return string(v)
		}

	case table.FieldItemType_Float:
		switch v := value.(type) {
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		case []uint8:
			return string(v)
		}

	case table.FieldItemType_Boolean:
		switch v := value.(type) {
		case bool:
			return strconv.FormatBool(v)
		case int64:
			return strconv.FormatBool(v != 0)
		case []uint8:
			boolStr := string(v)
			if boolStr == "1" || boolStr == "true" {
				return "true"
			}
			return "false"
		}

	case table.FieldItemType_Date:
		switch v := value.(type) {
		case time.Time:
			return v.Format(TimeFormat)
		case []uint8:
			return string(v)
		}
	}

	return fmt.Sprintf("%v", value)
}

// ConvertSystemFieldToString 将系统字段值转换为字符串
//
// 系统字段包括：_id（主键）、_uid（用户ID）、_cid（连接器ID）、_create_time（创建时间）。
// 这些字段由系统自动维护，需要特殊的格式化处理。
//
// 参数：
//   - fieldName: 系统字段名
//   - value: 数据库返回的原始值
//
// 返回值：
//   - 格式化后的字符串
func ConvertSystemFieldToString(fieldName string, value interface{}) string {
	switch fieldName {
	case database.DefaultIDColName:
		if intVal, ok := value.(int64); ok {
			return strconv.FormatInt(intVal, 10)
		}
	case database.DefaultUidColName, database.DefaultCidColName:
		if byteArray, ok := value.([]uint8); ok {
			return string(byteArray)
		}
	case database.DefaultCreateTimeColName:
		switch v := value.(type) {
		case time.Time:
			return v.Format(TimeFormat)
		case []uint8:
			// Attempt to parse the time represented by a string
			return string(v)
		}
	}

	return fmt.Sprintf("%v", value)
}

// ConvertLogicOperator 将业务逻辑操作符转换为数据库逻辑操作符
//
// 用于 WHERE 条件中多个条件的组合（AND/OR）。
func ConvertLogicOperator(logic database.Logic) entity.LogicalOperator {
	switch logic {
	case database.Logic_And:
		return entity.AND
	case database.Logic_Or:
		return entity.OR
	default:
		return entity.AND // Default use AND
	}
}

// ConvertOperator 将业务比较操作符转换为数据库操作符
//
// 支持的操作符包括：等于、不等于、大于、小于、IN、LIKE、IS NULL 等。
func ConvertOperator(op database.Operation) entity.Operator {
	switch op {
	case database.Operation_EQUAL:
		return entity.OperatorEqual
	case database.Operation_NOT_EQUAL:
		return entity.OperatorNotEqual
	case database.Operation_GREATER_THAN:
		return entity.OperatorGreater
	case database.Operation_GREATER_EQUAL:
		return entity.OperatorGreaterEqual
	case database.Operation_LESS_THAN:
		return entity.OperatorLess
	case database.Operation_LESS_EQUAL:
		return entity.OperatorLessEqual
	case database.Operation_IN:
		return entity.OperatorIn
	case database.Operation_NOT_IN:
		return entity.OperatorNotIn
	case database.Operation_LIKE:
		return entity.OperatorLike
	case database.Operation_NOT_LIKE:
		return entity.OperatorNotLike
	case database.Operation_IS_NULL:
		return entity.OperatorIsNull
	case database.Operation_IS_NOT_NULL:
		return entity.OperatorIsNotNull
	default:
		return entity.OperatorEqual
	}
}
