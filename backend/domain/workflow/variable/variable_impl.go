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

// Package variable 提供工作流变量系统的具体实现
//
// 这个包实现了变量存储器的具体逻辑，包含：
// - 三种变量存储器的实现（用户、应用、系统）
// - 变量的存储和获取逻辑
// - 与底层存储服务的集成
// - 业务ID映射和权限控制
//
// 存储架构：
// - varStore：统一的存储器实现结构体
// - VariableChannel：区分不同类型的变量存储渠道
// - StoreConfig：存储操作的上下文配置
//
// 技术实现：
// - 通过project_memory服务进行实际的变量存储
// - 支持应用ID、代理ID等业务实体的变量隔离
// - 提供类型安全的数据存取操作
//
// 注意事项：
// - 所有操作都需要提供正确的StoreConfig
// - 变量路径使用FieldPath格式进行嵌套访问
// - 错误处理包含详细的上下文信息
package variable

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/compose"

	"github.com/coze-dev/coze-studio/backend/api/model/data/variable/kvmemory"
	"github.com/coze-dev/coze-studio/backend/api/model/data/variable/project_memory"
	crossvariables "github.com/coze-dev/coze-studio/backend/crossdomain/variables"
	variablesModel "github.com/coze-dev/coze-studio/backend/crossdomain/variables/model"
	"github.com/coze-dev/coze-studio/backend/domain/memory/variables/entity"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ternary"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

// varStore 变量存储器实现
//
// 统一的变量存储器实现，通过不同的variableChannel区分存储渠道。
// 所有变量操作最终都会通过project_memory服务进行实际的存储和获取。
type varStore struct {
	// variableChannel 变量存储渠道
	// 标识变量的类型和存储位置（用户/应用/系统）
	variableChannel project_memory.VariableChannel
}

// NewVariableHandler 创建新的变量处理器
//
// 工厂函数，创建包含三种变量存储器的完整处理器实例。
// 每次调用都会创建新的实例，确保线程安全。
//
// 返回：
//   - *Handler: 配置完整的变量处理器实例
func NewVariableHandler() *Handler {
	return &Handler{
		UserVarStore:   newUserVarStore(),
		AppVarStore:    newAppVarStore(),
		SystemVarStore: newSystemVarStore(),
	}
}

// newUserVarStore 创建用户变量存储器
//
// 创建专门处理用户级别变量的存储器实例。
//
// 返回：
//   - Store: 用户变量存储器
func newUserVarStore() Store {
	return &varStore{
		variableChannel: project_memory.VariableChannel_Custom,
	}
}

// newAppVarStore 创建应用变量存储器
//
// 创建专门处理应用级别变量的存储器实例。
//
// 返回：
//   - Store: 应用变量存储器
func newAppVarStore() Store {
	return &varStore{
		variableChannel: project_memory.VariableChannel_APP,
	}
}

// newSystemVarStore 创建系统变量存储器
//
// 创建专门处理系统级别变量的存储器实例。
//
// 返回：
//   - Store: 系统变量存储器
func newSystemVarStore() Store {
	return &varStore{
		variableChannel: project_memory.VariableChannel_System,
	}
}

// Init 初始化变量存储器
//
// varStore的初始化方法，目前为空实现。
// 预留用于将来可能的初始化逻辑，如连接池建立、缓存初始化等。
//
// 参数：
//   - ctx: 上下文
func (v *varStore) Init(ctx context.Context) {
}

// Get 获取变量值
//
// 从指定的变量存储渠道获取变量值，支持嵌套路径访问。
// 根据变量类型自动进行相应的反序列化和类型转换。
//
// 参数：
//   - ctx: 上下文，用于取消和超时控制
//   - path: 变量路径，支持嵌套访问（如["user", "profile", "name"]）
//   - opts: 存储配置选项，必须包含StoreInfo
//
// 返回：
//   - any: 变量值，根据类型自动转换
//   - error: 获取失败的错误信息
//
// 处理逻辑：
// 1. 解析业务ID（应用ID或代理ID）
// 2. 调用存储服务获取变量数据
// 3. 根据变量类型进行反序列化
// 4. 支持嵌套对象和数组的路径访问
func (v *varStore) Get(ctx context.Context, path compose.FieldPath, opts ...OptionFn) (any, error) {
	// 解析配置选项
	opt := &StoreConfig{}
	for _, o := range opts {
		o(opt)
	}

	// 确定业务ID和类型（应用或代理）
	var (
		bizID   string
		bizType project_memory.VariableConnector
	)

	if opt.StoreInfo.AppID != nil {
		// 应用级别的变量
		bizID = strconv.FormatInt(*opt.StoreInfo.AppID, 10)
		bizType = project_memory.VariableConnector_Project
	} else if opt.StoreInfo.AgentID != nil {
		// 代理级别的变量
		bizID = strconv.FormatInt(*opt.StoreInfo.AgentID, 10)
		bizType = project_memory.VariableConnector_Bot
	} else {
		// 必须提供应用ID或代理ID之一
		return nil, fmt.Errorf("there must be one of the App ID or Agent ID")
	}

	meta := &variablesModel.UserVariableMeta{
		BizType:      bizType,
		BizID:        bizID,
		ConnectorID:  opt.StoreInfo.ConnectorID,
		ConnectorUID: opt.StoreInfo.ConnectorUID,
	}
	if len(path) == 0 {
		return nil, errors.New("field path is required")
	}
	key := path[0]
	kvItems, err := crossvariables.DefaultSVC().GetVariableChannelInstance(ctx, meta, []string{key}, project_memory.VariableChannelPtr(v.variableChannel))
	if err != nil {
		return nil, err
	}

	if len(kvItems) == 0 {
		return nil, fmt.Errorf("variable %s not exists", key)
	}

	value := kvItems[0].GetValue()

	schema := kvItems[0].GetSchema()

	varSchema, err := entity.NewVariableMetaSchema([]byte(schema))
	if err != nil {
		return nil, err
	}

	if varSchema.IsArrayType() {
		if value == "" {
			return nil, nil
		}
		result := make([]interface{}, 0)
		err = sonic.Unmarshal([]byte(value), &result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	if varSchema.IsObjectType() {
		if value == "" {
			return nil, nil
		}
		result := make(map[string]any)
		err = sonic.Unmarshal([]byte(value), &result)
		if err != nil {
			return nil, err
		}
		if len(path) > 1 {
			if val, ok := takeMapValue(result, path[1:]); ok {
				return val, nil
			}
			return nil, nil
		}
		return result, nil
	}

	if varSchema.IsStringType() {
		return value, nil
	}

	if varSchema.IsBooleanType() {
		if value == "" {
			return false, nil
		}
		result, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	if varSchema.IsNumberType() {
		if value == "" {
			return 0, nil
		}
		result, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	if varSchema.IsIntegerType() {
		if value == "" {
			return 0, nil
		}
		result, err := strconv.ParseInt(value, 64, 10)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	return value, nil
}

// Set 设置变量值
//
// 将变量值存储到指定的变量存储渠道中。
// 支持各种数据类型的自动序列化，支持嵌套对象和数组。
//
// 参数：
//   - ctx: 上下文，用于取消和超时控制
//   - path: 变量路径，支持嵌套设置（如["user", "profile", "name"]）
//   - value: 要存储的变量值，支持string、map、slice等类型
//   - opts: 存储配置选项，必须包含StoreInfo
//
// 返回：
//   - error: 设置失败的错误信息
//
// 处理逻辑：
// 1. 解析业务ID（应用ID或代理ID）
// 2. 根据数据类型进行序列化
// 3. 构建存储元数据
// 4. 调用存储服务保存变量
// 5. 处理存储响应和错误
func (v *varStore) Set(ctx context.Context, path compose.FieldPath, value any, opts ...OptionFn) (err error) {
	// 解析配置选项
	opt := &StoreConfig{}
	for _, o := range opts {
		o(opt)
	}

	// 确定业务ID和类型（应用或代理）
	var (
		bizID   string
		bizType project_memory.VariableConnector
	)

	if opt.StoreInfo.AppID != nil {
		// 应用级别的变量
		bizID = strconv.FormatInt(*opt.StoreInfo.AppID, 10)
		bizType = project_memory.VariableConnector_Project
	} else if opt.StoreInfo.AgentID != nil {
		// 代理级别的变量
		bizID = strconv.FormatInt(*opt.StoreInfo.AgentID, 10)
		bizType = project_memory.VariableConnector_Bot
	} else {
		// 必须提供应用ID或代理ID之一
		return fmt.Errorf("there must be one of the App ID or Agent ID")
	}

	meta := &variablesModel.UserVariableMeta{
		BizType:      bizType,
		BizID:        bizID,
		ConnectorID:  opt.StoreInfo.ConnectorID,
		ConnectorUID: opt.StoreInfo.ConnectorUID,
	}

	if len(path) == 0 {
		return errors.New("field path is required")
	}

	key := path[0]
	kvItems := make([]*kvmemory.KVItem, 0, 1)

	valueString := ""
	if _, ok := value.(string); ok {
		valueString = value.(string)
	} else {
		valueString, err = sonic.MarshalString(value)
		if err != nil {
			return err
		}
	}

	isSystem := ternary.IFElse[bool](v.variableChannel == project_memory.VariableChannel_System, true, false)
	kvItems = append(kvItems, &kvmemory.KVItem{
		Keyword:  key,
		Value:    valueString,
		IsSystem: isSystem,
	})

	_, err = crossvariables.DefaultSVC().SetVariableInstance(ctx, meta, kvItems)
	if err != nil {
		return err
	}

	return nil
}

// variablesMetaGetter 变量元数据获取器的实现
//
// 实现VariablesMetaGetter接口，提供应用和代理变量元数据的获取功能。
// 通过调用crossdomain服务获取变量的类型定义信息。
type variablesMetaGetter struct {
}

// NewVariablesMetaGetter 创建变量元数据获取器
//
// 工厂函数，创建用于获取应用和代理变量元数据的实例。
//
// 返回：
//   - VariablesMetaGetter: 变量元数据获取器实例
func NewVariablesMetaGetter() VariablesMetaGetter {
	return &variablesMetaGetter{}
}

// GetAppVariablesMeta 获取应用变量元数据
//
// 获取指定应用和版本的所有变量的类型信息，用于前端展示和类型验证。
//
// 参数：
//   - ctx: 上下文
//   - id: 应用ID
//   - version: 应用版本
//
// 返回：
//   - map[string]*vo.TypeInfo: 变量名到类型信息的映射
//   - error: 获取失败的错误信息
//
// 注意：
//   - 返回的类型信息包含变量的数据类型、结构等
//   - 用于前端变量选择器和类型检查
func (v variablesMetaGetter) GetAppVariablesMeta(ctx context.Context, id, version string) (m map[string]*vo.TypeInfo, err error) {
	var varMetas *entity.VariablesMeta
	varMetas, err = crossvariables.DefaultSVC().GetProjectVariablesMeta(ctx, id, version)
	if err != nil {
		return nil, err
	}

	m = make(map[string]*vo.TypeInfo, len(varMetas.Variables))
	for _, v := range varMetas.Variables {
		varSchema, err := v.GetSchema(ctx)
		if err != nil {
			return nil, vo.WrapIfNeeded(errno.ErrVariablesAPIFail, err)
		}

		t, err := varMeta2TypeInfo(varSchema)
		if err != nil {
			return nil, err
		}

		m[v.Keyword] = t
	}

	return m, nil
}

// GetAgentVariablesMeta 获取代理变量元数据
//
// 获取指定代理和版本的所有变量的类型信息，用于前端展示和类型验证。
//
// 参数：
//   - ctx: 上下文
//   - id: 代理ID
//   - version: 代理版本
//
// 返回：
//   - map[string]*vo.TypeInfo: 变量名到类型信息的映射
//   - error: 获取失败的错误信息
//
// 注意：
//   - 返回的类型信息包含变量的数据类型、结构等
//   - 用于前端变量选择器和类型检查
func (v variablesMetaGetter) GetAgentVariablesMeta(ctx context.Context, id int64, version string) (m map[string]*vo.TypeInfo, err error) {
	var varMetas *entity.VariablesMeta
	varMetas, err = crossvariables.DefaultSVC().GetAgentVariableMeta(ctx, id, version)
	if err != nil {
		return nil, err
	}

	m = make(map[string]*vo.TypeInfo, len(varMetas.Variables))
	for _, v := range varMetas.Variables {
		varSchema, err := v.GetSchema(ctx)
		if err != nil {
			return nil, vo.WrapIfNeeded(errno.ErrVariablesAPIFail, err)
		}

		t, err := varMeta2TypeInfo(varSchema)
		if err != nil {
			return nil, err
		}

		m[v.Keyword] = t
	}

	return m, nil
}

// varMeta2TypeInfo 将变量元数据转换为类型信息
//
// 将内部的VariableMetaSchema转换为前端使用的TypeInfo格式。
// 支持基本类型（字符串、数字、布尔）和复杂类型（对象、数组）。
//
// 参数：
//   - v: 变量元数据schema
//
// 返回：
//   - *vo.TypeInfo: 类型信息对象
//   - error: 转换失败的错误信息
//
// 支持的类型：
//   - 基本类型：string, number, integer, boolean
//   - 复杂类型：object（包含属性定义）, array（包含元素类型）
func varMeta2TypeInfo(v *entity.VariableMetaSchema) (*vo.TypeInfo, error) {
	if v.IsBooleanType() {
		return &vo.TypeInfo{
			Type: vo.DataTypeBoolean,
		}, nil
	}
	if v.IsStringType() {
		return &vo.TypeInfo{
			Type: vo.DataTypeString,
		}, nil
	}
	if v.IsNumberType() {
		return &vo.TypeInfo{
			Type: vo.DataTypeNumber,
		}, nil
	}
	if v.IsIntegerType() {
		return &vo.TypeInfo{
			Type: vo.DataTypeInteger,
		}, nil
	}
	if v.IsArrayType() {
		if len(v.Schema) == 0 {
			return nil, vo.WrapError(errno.ErrVariablesAPIFail, fmt.Errorf("array type should contain element type info"))
		}

		elemType, err := entity.NewVariableMetaSchema(v.Schema)
		if err != nil {
			return nil, vo.WrapIfNeeded(errno.ErrVariablesAPIFail, err)
		}

		et, err := varMeta2TypeInfo(elemType)
		if err != nil {
			return nil, err
		}

		return &vo.TypeInfo{
			Type:         vo.DataTypeArray,
			ElemTypeInfo: et,
		}, nil
	}
	if v.IsObjectType() {
		ps, err := v.GetObjectProperties(v.Schema)
		if err != nil {
			return nil, vo.WrapIfNeeded(errno.ErrVariablesAPIFail, err)
		}

		properties := make(map[string]*vo.TypeInfo, len(ps))
		for k, p := range ps {
			pt, err := varMeta2TypeInfo(p)
			if err != nil {
				return nil, err
			}
			properties[k] = pt
		}

		return &vo.TypeInfo{
			Type:       vo.DataTypeObject,
			Properties: properties,
		}, nil
	}
	return nil, vo.WrapError(errno.ErrVariablesAPIFail, fmt.Errorf("invalid variable type"))
}

// takeMapValue 从嵌套map中提取值
//
// 根据路径从嵌套的map结构中提取值，支持多层嵌套访问。
// 用于处理对象类型变量的路径访问，如["user", "profile", "name"]。
//
// 参数：
//   - m: 源map对象
//   - path: 访问路径，每个元素为一个键名
//
// 返回：
//   - any: 提取到的值，如果路径不存在返回nil
//   - bool: 是否成功提取到值
//
// 访问逻辑：
//   - 逐层深入嵌套结构
//   - 如果任一层级不存在或类型不匹配，返回false
//   - 支持任意深度的嵌套访问
func takeMapValue(m map[string]any, path []string) (any, bool) {
	if m == nil {
		return nil, false
	}

	container := m
	for _, p := range path[:len(path)-1] {
		if _, ok := container[p]; !ok {
			return nil, false
		}
		container = container[p].(map[string]any)
	}

	if v, ok := container[path[len(path)-1]]; ok {
		return v, true
	}

	return nil, false
}
