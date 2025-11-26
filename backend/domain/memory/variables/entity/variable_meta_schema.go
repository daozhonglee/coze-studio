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

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/coze-dev/coze-studio/backend/pkg/errorx"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

// 变量 Schema 支持的类型常量
const (
	variableMetaSchemaTypeObject  = "object"  // 对象类型
	variableMetaSchemaTypeArray   = "list"    // 数组类型
	variableMetaSchemaTypeInteger = "integer" // 整数类型
	variableMetaSchemaTypeString  = "string"  // 字符串类型
	variableMetaSchemaTypeBoolean = "boolean" // 布尔类型
	variableMetaSchemaTypeNumber  = "float"   // 浮点数类型
)

// VariableMetaSchema 变量 Schema 定义
//
// 用于定义变量值的结构和类型约束，支持嵌套的对象和数组类型。
// Schema 遵循类似 JSON Schema 的格式，但有所简化。
type VariableMetaSchema struct {
	// Type 数据类型（string/integer/float/boolean/object/list）
	Type string `json:"type,omitempty"`
	// Name 字段名称（用于对象属性）
	Name string `json:"name,omitempty"`
	// Description 字段描述
	Description string `json:"description,omitempty"`
	// Readonly 是否只读
	Readonly bool `json:"readonly,omitempty"`
	// Enable 是否启用
	Enable bool `json:"enable,omitempty"`
	// Schema 嵌套 Schema（用于对象和数组类型）
	Schema json.RawMessage `json:"schema,omitempty"`
}

// NewVariableMetaSchema 从 JSON 字节数组创建 Schema 对象
func NewVariableMetaSchema(schema []byte) (*VariableMetaSchema, error) {
	schemaObj := &VariableMetaSchema{}
	err := json.Unmarshal(schema, schemaObj)
	if err != nil {
		return nil, errorx.New(errno.ErrMemorySchemeInvalidCode, errorx.KVf("msg", "schema json invalid: %s \n json = %s", err.Error(), string(schema)))
	}

	return schemaObj, nil
}

// IsArrayType 判断是否为数组类型
func (v *VariableMetaSchema) IsArrayType() bool {
	return v.Type == variableMetaSchemaTypeArray
}

// GetArrayType 获取数组元素的类型
//
// 从数组类型的 Schema 中提取元素类型，如 {"type":"int"} 返回 "int"
func (v *VariableMetaSchema) GetArrayType(schema []byte) (string, error) {
	schemaObj, err := NewVariableMetaSchema(schema)
	if err != nil {
		return "", errorx.New(errno.ErrMemorySchemeInvalidCode, errorx.KVf("msg", "NewVariableMetaSchema failed, %v", err.Error()))
	}

	if schemaObj.Type == "" {
		return "", errorx.New(errno.ErrMemorySchemeInvalidCode, errorx.KVf("msg", "array type not found in %s", schema))
	}

	return schemaObj.Type, nil
}

// IsStringType 判断是否为字符串类型
func (v *VariableMetaSchema) IsStringType() bool {
	return v.Type == variableMetaSchemaTypeString
}

// IsIntegerType 判断是否为整数类型
func (v *VariableMetaSchema) IsIntegerType() bool {
	return v.Type == variableMetaSchemaTypeInteger
}

// IsBooleanType 判断是否为布尔类型
func (v *VariableMetaSchema) IsBooleanType() bool {
	return v.Type == variableMetaSchemaTypeBoolean
}

// IsNumberType 判断是否为浮点数类型
func (v *VariableMetaSchema) IsNumberType() bool {
	return v.Type == variableMetaSchemaTypeNumber
}

// IsObjectType 判断是否为对象类型
func (v *VariableMetaSchema) IsObjectType() bool {
	return v.Type == variableMetaSchemaTypeObject
}

// GetObjectProperties 获取对象类型的属性定义
//
// 从对象类型的 Schema 中提取属性列表，返回 name -> Schema 的映射。
// 示例输入: [{"name":"field1","type":"string"}]
func (v *VariableMetaSchema) GetObjectProperties(schema []byte) (map[string]*VariableMetaSchema, error) {
	schemas := make([]*VariableMetaSchema, 0)
	err := json.Unmarshal(schema, &schemas)
	if err != nil {
		return nil, errorx.New(errno.ErrMemorySchemeInvalidCode, errorx.KV("msg", "schema array content json invalid"))
	}

	properties := make(map[string]*VariableMetaSchema)
	for _, schemaObj := range schemas {
		properties[schemaObj.Name] = schemaObj
	}

	return properties, nil
}

// check 验证 Schema 的有效性
func (v *VariableMetaSchema) check(ctx context.Context) error {
	return v.checkAppVariableSchema(ctx, v, "")
}

// checkAppVariableSchema 递归验证变量 Schema
//
// 验证 Schema 的名称格式、类型有效性，以及嵌套 Schema 的正确性。
func (v *VariableMetaSchema) checkAppVariableSchema(ctx context.Context, schemaObj *VariableMetaSchema, schema string) (err error) {
	if len(schema) == 0 && schemaObj == nil {
		return errorx.New(errno.ErrMemorySchemeInvalidCode, errorx.KV("msg", "schema is nil"))
	}

	if schemaObj == nil {
		schemaObj, err = NewVariableMetaSchema([]byte(schema))
		if err != nil {
			return errorx.New(errno.ErrMemorySchemeInvalidCode, errorx.KVf("msg", "checkAppVariableSchema failed , %v", err.Error()))
		}
	}

	if !schemaObj.nameValidate() {
		return errorx.New(errno.ErrMemorySchemeInvalidCode, errorx.KVf("msg", "name(%s) is invalid", schemaObj.Name))
	}

	if schemaObj.Type == variableMetaSchemaTypeObject {
		return v.checkSchemaObj(ctx, schemaObj.Schema)
	} else if schemaObj.Type == variableMetaSchemaTypeArray {
		_, err := v.GetArrayType(schemaObj.Schema)
		if err != nil {
			return errorx.New(errno.ErrMemorySchemeInvalidCode, errorx.KVf("msg", "GetArrayType failed : %v", err.Error()))
		}
	}

	return nil
}

// checkSchemaObj 验证对象类型 Schema 的所有属性
func (v *VariableMetaSchema) checkSchemaObj(ctx context.Context, schema []byte) error {
	properties, err := v.GetObjectProperties(schema)
	if err != nil {
		return errorx.New(errno.ErrMemorySchemeInvalidCode, errorx.KVf("msg", "GetObjectProperties failed : %v", err.Error()))
	}

	for _, schemaObj := range properties {
		if err := v.checkAppVariableSchema(ctx, schemaObj, ""); err != nil {
			return err
		}
	}

	return nil
}

// nameValidate 验证变量名称的有效性
//
// 变量名必须符合标识符规则：以字母或下划线开头，只包含字母、数字、下划线和 $。
// 不能使用保留字（如 true、false、and、or、null 等）。
func (v *VariableMetaSchema) nameValidate() bool {
	identifier := v.Name

	reservedWords := map[string]bool{
		"true": true, "false": true, "and": true, "AND": true,
		"or": true, "OR": true, "not": true, "NOT": true,
		"null": true, "nil": true, "If": true, "Switch": true,
	}

	if reservedWords[identifier] {
		return false
	}

	// Check if some of the following regular rules are met
	pattern := `^[a-zA-Z_][a-zA-Z_$0-9]*$`
	match, _ := regexp.MatchString(pattern, identifier)

	return match
}
