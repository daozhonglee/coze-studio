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

// Package physicaltable 提供物理数据库表的管理功能
//
// 本包负责以下操作：
// - 创建物理表（根据字段定义生成 DDL）
// - 更新物理表结构（添加/修改/删除列）
// - 字段名映射（逻辑名到物理名的转换）
// - 系统字段定义（主键、用户ID、创建时间等）
//
// 设计说明：
// 物理表使用自动生成的列名（如 f_1, f_2）而非用户定义的字段名，
// 这样可以避免 SQL 注入和特殊字符问题，同时支持字段重命名不影响数据。
// 系统字段（_id, _uid, _cid, _create_time）由系统自动维护。
package physicaltable

import (
	"context"
	"fmt"

	database "github.com/coze-dev/coze-studio/backend/crossdomain/database/model"

	"github.com/coze-dev/coze-studio/backend/api/model/data/database/table"
	"github.com/coze-dev/coze-studio/backend/domain/memory/database/internal/convertor"
	"github.com/coze-dev/coze-studio/backend/infra/rdb"
	entity3 "github.com/coze-dev/coze-studio/backend/infra/rdb/entity"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
)

// CreatePhysicalTable 创建物理数据库表
//
// 该函数执行以下操作：
// 1. 根据列定义创建表结构
// 2. 添加主键索引（基于 _id 列）
// 3. 添加用户索引（基于 _uid 和 _cid 列，用于权限控制）
//
// 参数：
//   - ctx: 上下文
//   - db: RDB 数据库接口
//   - columns: 列定义列表（包含用户字段和系统字段）
//
// 返回值：
//   - CreateTableResponse: 包含创建的表信息
//   - error: 创建失败时的错误
func CreatePhysicalTable(ctx context.Context, db rdb.RDB, columns []*entity3.Column) (*rdb.CreateTableResponse, error) {
	table := &entity3.Table{
		Columns: columns,
	}
	// get indexes
	indexes := make([]*entity3.Index, 0)
	indexes = append(indexes, &entity3.Index{
		Name:    "PRIMARY",
		Type:    entity3.PrimaryKey,
		Columns: []string{database.DefaultIDColName},
	})
	indexes = append(indexes, &entity3.Index{
		Name:    "idx_uid",
		Type:    entity3.NormalKey,
		Columns: []string{database.DefaultUidColName, database.DefaultCidColName},
	})
	table.Indexes = indexes

	physicalTableRes, err := db.CreateTable(ctx, &rdb.CreateTableRequest{Table: table})
	if err != nil {
		return nil, err
	}

	return physicalTableRes, nil
}

// CreateFieldInfo 根据字段定义生成物理列信息
//
// 该函数执行以下操作：
// 1. 为每个字段分配自增的 AlterID
// 2. 生成物理列名（f_1, f_2, ...）
// 3. 转换字段类型为数据库类型
// 4. 添加系统默认列（_id, _uid, _cid, _create_time）
//
// 参数：
//   - fieldItems: 用户定义的字段列表
//
// 返回值：
//   - []*database.FieldItem: 更新后的字段列表（包含 AlterID 和 PhysicalName）
//   - []*entity3.Column: 物理列定义列表（用于创建表）
func CreateFieldInfo(fieldItems []*database.FieldItem) ([]*database.FieldItem, []*entity3.Column) {
	columns := make([]*entity3.Column, len(fieldItems))

	fieldID := int64(1)
	for i, field := range fieldItems {
		field.AlterID = fieldID
		field.PhysicalName = GetFieldPhysicsName(fieldID)

		columns[i] = &entity3.Column{
			Name:     GetFieldPhysicsName(fieldID),
			DataType: convertor.SwitchToDataType(field.Type),
			NotNull:  field.MustRequired,
			Comment:  &field.Desc,
		}

		if field.Type == table.FieldItemType_Text && !field.MustRequired {
			columns[i].DefaultValue = ptr.Of("")
		}

		fieldID++ // field is incremented begin from 1
	}

	columns = append(columns, GetDefaultColumns()...)

	return fieldItems, columns
}

// GetDefaultColumns 获取系统默认列定义
//
// 系统默认列包括：
//   - _id: 主键，BIGINT，自增
//   - _uid: 用户ID，VARCHAR，用于权限控制
//   - _cid: 连接器ID，VARCHAR，标识数据来源
//   - _create_time: 创建时间，TIMESTAMP，自动填充
func GetDefaultColumns() []*entity3.Column {
	return getDefaultColumns()
}

// getDefaultColumns 内部实现：返回系统默认列的定义
func getDefaultColumns() []*entity3.Column {
	return []*entity3.Column{
		{
			Name:          database.DefaultIDColName,
			DataType:      entity3.TypeBigInt,
			NotNull:       true,
			AutoIncrement: true,
		},
		{
			Name:     database.DefaultUidColName,
			DataType: entity3.TypeVarchar,
			NotNull:  true,
		},
		{
			Name:     database.DefaultCidColName,
			DataType: entity3.TypeVarchar,
			NotNull:  true,
		},
		{
			Name:         database.DefaultCreateTimeColName,
			DataType:     entity3.TypeTimestamp,
			NotNull:      true,
			DefaultValue: ptr.Of("CURRENT_TIMESTAMP"),
		},
	}
}

// GetTablePhysicsName 根据表 ID 生成物理表名
//
// 格式：table_{tableID}
func GetTablePhysicsName(tableID int64) string {
	return fmt.Sprintf("table_%d", tableID)
}

// GetFieldPhysicsName 根据字段 ID 生成物理列名
//
// 格式：f_{fieldID}
func GetFieldPhysicsName(fieldID int64) string {
	return fmt.Sprintf("f_%d", fieldID)
}

// UpdateFieldInfo 处理字段信息更新
//
// 该函数对比新旧字段列表，生成需要执行的 DDL 操作：
// 1. 如果 AlterID 存在，更新现有字段
// 2. 如果 AlterID 不存在，添加新字段
// 3. 删除在新列表中不存在的字段
//
// 参数：
//   - newFieldItems: 新的字段定义列表
//   - existingFieldItems: 现有的字段定义列表
//
// 返回值：
//   - []*database.FieldItem: 更新后的字段列表
//   - []*entity3.Column: 需要添加/修改的列定义
//   - []string: 需要删除的物理列名列表
//   - error: 处理失败时的错误
func UpdateFieldInfo(newFieldItems []*database.FieldItem, existingFieldItems []*database.FieldItem) ([]*database.FieldItem, []*entity3.Column, []string, error) {
	existingFieldMap := make(map[int64]*database.FieldItem)
	maxAlterID := int64(-1)
	for _, field := range existingFieldItems {
		if field.AlterID > 0 {
			existingFieldMap[field.AlterID] = field
			maxAlterID = max(maxAlterID, field.AlterID)
		}
	}

	newFieldIDs := make(map[int64]bool)

	updatedColumns := make([]*entity3.Column, 0, len(newFieldItems))
	updatedFieldItems := make([]*database.FieldItem, 0, len(newFieldItems))

	for _, field := range newFieldItems {
		if field.AlterID > 0 {
			// update field
			newFieldIDs[field.AlterID] = true
			field.PhysicalName = GetFieldPhysicsName(field.AlterID)
			updatedFieldItems = append(updatedFieldItems, field)

			updatedColumns = append(updatedColumns, &entity3.Column{
				Name:     GetFieldPhysicsName(field.AlterID),
				DataType: convertor.SwitchToDataType(field.Type),
				NotNull:  field.MustRequired,
				Comment:  &field.Desc,
			})
		} else {
			fieldID := maxAlterID + 1 // auto increment begin from existing maxAlterID
			maxAlterID++
			field.AlterID = fieldID
			field.PhysicalName = GetFieldPhysicsName(fieldID)
			updatedFieldItems = append(updatedFieldItems, field)

			updatedColumns = append(updatedColumns, &entity3.Column{
				Name:     GetFieldPhysicsName(fieldID),
				DataType: convertor.SwitchToDataType(field.Type),
				NotNull:  field.MustRequired,
				Comment:  &field.Desc,
			})
		}
	}

	droppedColumns := make([]string, 0, len(existingFieldMap))
	// get dropped columns
	for alterID := range existingFieldMap {
		if !newFieldIDs[alterID] {
			droppedColumns = append(droppedColumns, GetFieldPhysicsName(alterID))
		}
	}

	return updatedFieldItems, updatedColumns, droppedColumns, nil
}

// UpdatePhysicalTableWithDrops 更新物理表结构
//
// 该函数对比新旧表结构，生成并执行 ALTER TABLE 语句：
// 1. 添加新列
// 2. 修改现有列
// 3. 删除指定的列
//
// 参数：
//   - ctx: 上下文
//   - db: RDB 数据库接口
//   - existingTable: 现有表结构
//   - newColumns: 新的列定义列表
//   - droppedColumns: 需要删除的物理列名列表
//   - tableName: 物理表名
//
// 返回值：
//   - error: 更新失败时的错误
func UpdatePhysicalTableWithDrops(ctx context.Context, db rdb.RDB, existingTable *entity3.Table, newColumns []*entity3.Column, droppedColumns []string, tableName string) error {
	// Create a column name-to-column mapping
	existingColumnMap := make(map[string]*entity3.Column)
	for _, col := range existingTable.Columns {
		existingColumnMap[col.Name] = col
	}

	// Collect columns to add and modify
	var columnsToAdd, columnsToModify []*entity3.Column

	// Find columns to add and modify
	for _, newCol := range newColumns {
		if _, exists := existingColumnMap[newCol.Name]; exists {
			columnsToModify = append(columnsToModify, newCol)
		} else {
			columnsToAdd = append(columnsToAdd, newCol)
		}
	}

	// Apply changes to physical tables
	if len(columnsToAdd) > 0 || len(columnsToModify) > 0 || len(droppedColumns) > 0 {
		// build AlterTableRequest
		alterReq := &rdb.AlterTableRequest{
			TableName:  tableName,
			Operations: getOperation(columnsToAdd, columnsToModify, droppedColumns),
		}

		// Perform table structure changes
		_, err := db.AlterTable(ctx, alterReq)
		if err != nil {
			return err
		}
	}

	return nil
}

// getOperation 将列操作转换为 AlterTableOperation 数组
//
// 将添加、修改、删除操作组合成统一的操作列表，
// 用于批量执行 ALTER TABLE 语句。
func getOperation(columnsToAdd, columnsToModify []*entity3.Column, droppedColumns []string) []*rdb.AlterTableOperation {
	operations := make([]*rdb.AlterTableOperation, 0)

	// Handle add column operations
	for _, column := range columnsToAdd {
		operations = append(operations, &rdb.AlterTableOperation{
			Action: entity3.AddColumn,
			Column: column,
		})
	}

	// Handle modify column operations
	for _, column := range columnsToModify {
		operations = append(operations, &rdb.AlterTableOperation{
			Action: entity3.ModifyColumn,
			Column: column,
		})
	}

	// Handle delete column operations
	for _, columnName := range droppedColumns {
		operations = append(operations, &rdb.AlterTableOperation{
			Action: entity3.DropColumn,
			Column: &entity3.Column{Name: columnName},
		})
	}

	return operations
}

// GetTemplateTypeMap 获取字段类型对应的模板默认值
//
// 用于生成 Excel 导入模板时填充示例值。
func GetTemplateTypeMap() map[table.FieldItemType]string {
	return map[table.FieldItemType]string{
		table.FieldItemType_Boolean: "false",
		table.FieldItemType_Number:  "0",
		table.FieldItemType_Date:    "0001-01-01 00:00:00",
		table.FieldItemType_Text:    "",
		table.FieldItemType_Float:   "0",
	}
}

// GetCreateTimeField 获取创建时间系统字段定义（内部使用）
func GetCreateTimeField() *database.FieldItem {
	return &database.FieldItem{
		Name:          database.DefaultCreateTimeColName,
		Desc:          "create time",
		Type:          table.FieldItemType_Date,
		MustRequired:  false,
		IsSystemField: true,
		AlterID:       103,
		PhysicalName:  database.DefaultCreateTimeColName,
	}
}

// GetUidField 获取用户 ID 系统字段定义（内部使用）
func GetUidField() *database.FieldItem {
	return &database.FieldItem{
		Name:          database.DefaultUidColName,
		Desc:          "user id",
		Type:          table.FieldItemType_Text,
		MustRequired:  false,
		IsSystemField: true,
		AlterID:       101,
		PhysicalName:  database.DefaultUidColName,
	}
}

// GetConnectIDField 获取连接器 ID 系统字段定义（内部使用）
func GetConnectIDField() *database.FieldItem {
	return &database.FieldItem{
		Name:          database.DefaultCidColName,
		Desc:          "connector id",
		Type:          table.FieldItemType_Text,
		MustRequired:  false,
		IsSystemField: true,
		AlterID:       104,
		PhysicalName:  database.DefaultCidColName,
	}
}

// GetIDField 获取主键系统字段定义（内部使用）
func GetIDField() *database.FieldItem {
	return &database.FieldItem{
		Name:          database.DefaultIDColName,
		Desc:          "primary_key",
		Type:          table.FieldItemType_Number,
		MustRequired:  false,
		IsSystemField: true,
		AlterID:       102,
		PhysicalName:  database.DefaultIDColName,
	}
}

// GetDisplayCreateTimeField 获取创建时间系统字段定义（对外展示）
//
// 与 GetCreateTimeField 的区别在于使用用户可见的字段名。
func GetDisplayCreateTimeField() *database.FieldItem {
	return &database.FieldItem{
		Name:          database.DefaultCreateTimeDisplayColName,
		Desc:          "create time",
		Type:          table.FieldItemType_Date,
		MustRequired:  false,
		IsSystemField: true,
		AlterID:       103,
		PhysicalName:  database.DefaultCreateTimeDisplayColName,
	}
}

// GetDisplayUidField 获取用户 ID 系统字段定义（对外展示）
func GetDisplayUidField() *database.FieldItem {
	return &database.FieldItem{
		Name:          database.DefaultUidDisplayColName,
		Desc:          "user id",
		Type:          table.FieldItemType_Text,
		MustRequired:  false,
		IsSystemField: true,
		AlterID:       101,
		PhysicalName:  database.DefaultUidDisplayColName,
	}
}

// GetDisplayIDField 获取主键系统字段定义（对外展示）
func GetDisplayIDField() *database.FieldItem {
	return &database.FieldItem{
		Name:          database.DefaultIDDisplayColName,
		Desc:          "primary_key",
		Type:          table.FieldItemType_Number,
		MustRequired:  false,
		IsSystemField: true,
		AlterID:       102,
		PhysicalName:  database.DefaultIDDisplayColName,
	}
}
