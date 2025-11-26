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
	"fmt"

	"github.com/coze-dev/coze-studio/backend/api/model/app/bot_common"
	"github.com/coze-dev/coze-studio/backend/api/model/data/variable/project_memory"
)

// VariablesMeta 变量集合实体
//
// 表示一组变量的完整定义，关联到特定的业务对象（如 Agent 或 Project）。
type VariablesMeta struct {
	// ID 变量集合 ID
	ID int64
	// CreatorID 创建者用户 ID
	CreatorID int64
	// BizType 业务类型（Bot/Project）
	BizType project_memory.VariableConnector
	// BizID 业务 ID（Agent ID 或 Project ID）
	BizID string
	// CreatedAt 创建时间戳
	CreatedAt int64
	// UpdatedAt 更新时间戳
	UpdatedAt int64
	// Version 版本号
	Version string
	// Variables 变量列表
	Variables []*VariableMeta
}

// NewVariablesWithAgentVariables 从 Agent 变量列表创建变量集合
func NewVariablesWithAgentVariables(vars []*bot_common.Variable) *VariablesMeta {
	res := make([]*VariableMeta, 0)
	for _, variable := range vars {
		res = append(res, agentVariableMetaToProjectVariableMeta(variable))
	}
	return &VariablesMeta{
		Variables: res,
	}
}

// NewVariables 从 Project 变量列表创建变量集合
func NewVariables(vars []*project_memory.Variable) *VariablesMeta {
	res := make([]*VariableMeta, 0)
	for _, variable := range vars {
		res = append(res, &VariableMeta{
			Keyword:              variable.Keyword,
			DefaultValue:         variable.DefaultValue,
			VariableType:         variable.VariableType,
			Channel:              variable.Channel,
			Description:          variable.Description,
			Enable:               variable.Enable,
			EffectiveChannelList: variable.EffectiveChannelList,
			Schema:               variable.Schema,
			IsReadOnly:           variable.IsReadOnly,
		})
	}
	return &VariablesMeta{
		Variables: res,
	}
}

// ToAgentVariables 转换为 Agent 变量列表格式
func (v *VariablesMeta) ToAgentVariables() []*bot_common.Variable {
	res := make([]*bot_common.Variable, 0, len(v.Variables))
	for idx := range v.Variables {
		v := v.Variables[idx]
		isSystem := v.Channel == project_memory.VariableChannel_System
		isDisabled := !v.Enable
		agentVariable := &bot_common.Variable{
			Key:            &v.Keyword,
			DefaultValue:   &v.DefaultValue,
			Description:    &v.Description,
			IsDisabled:     &isDisabled,
			IsSystem:       &isSystem,
			PromptDisabled: &v.PromptDisabled,
		}

		res = append(res, agentVariable)
	}

	return res
}

// ToProjectVariables 转换为 Project 变量列表格式
func (v *VariablesMeta) ToProjectVariables() []*project_memory.Variable {
	res := make([]*project_memory.Variable, 0, len(v.Variables))
	for _, v := range v.Variables {
		res = append(res, v.ToProjectVariable())
	}
	return res
}

// SetupIsReadOnly 为所有变量设置只读属性
func (v *VariablesMeta) SetupIsReadOnly() {
	for _, variable := range v.Variables {
		variable.SetupIsReadOnly()
	}
}

// SetupSchema 为所有变量设置默认 Schema
func (v *VariablesMeta) SetupSchema() {
	for _, variable := range v.Variables {
		variable.SetupSchema()
	}
}

// agentVariableMetaToProjectVariableMeta Agent 变量转换为通用变量元数据
func agentVariableMetaToProjectVariableMeta(variable *bot_common.Variable) *VariableMeta {
	temp := &VariableMeta{
		Keyword:        variable.GetKey(),
		DefaultValue:   variable.GetDefaultValue(),
		VariableType:   project_memory.VariableType_KVVariable,
		Description:    variable.GetDescription(),
		Enable:         !variable.GetIsDisabled(),
		Schema:         fmt.Sprintf(stringSchema, variable.GetKey()),
		PromptDisabled: variable.GetPromptDisabled(),
	}
	if variable.GetIsSystem() {
		temp.IsReadOnly = true
		temp.Channel = project_memory.VariableChannel_System
	} else {
		temp.Channel = project_memory.VariableChannel_Custom
	}

	return temp
}

// GroupByChannel 按渠道分组变量
func (v *VariablesMeta) GroupByChannel() map[project_memory.VariableChannel][]*project_memory.Variable {
	res := make(map[project_memory.VariableChannel][]*project_memory.Variable)
	for _, variable := range v.Variables {
		ch := variable.Channel
		res[ch] = append(res[ch], variable.ToProjectVariable())
	}

	return res
}

// RemoveDisableVariable 移除禁用的变量
func (v *VariablesMeta) RemoveDisableVariable() {
	var res []*VariableMeta
	for _, vv := range v.Variables {
		if vv.Enable {
			res = append(res, vv)
		}
	}

	v.Variables = res
}

// FilterChannelVariable 按渠道过滤变量
func (v *VariablesMeta) FilterChannelVariable(ch project_memory.VariableChannel) {
	var res []*VariableMeta
	for _, vv := range v.Variables {
		if vv.Channel != ch {
			continue
		}

		res = append(res, vv)
	}

	v.Variables = res
}
