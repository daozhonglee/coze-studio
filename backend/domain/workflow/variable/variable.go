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

// Package variable 提供工作流变量管理系统
//
// 这个包是工作流系统中变量管理的核心组件，负责管理三种类型的全局变量：
// - 用户变量：用户级别的持久化数据存储
// - 应用变量：应用级别的配置和状态数据
// - 系统变量：系统级别的内置配置数据
//
// 主要功能：
// 1. 统一的变量存储和访问接口
// 2. 三种变量类型的路由和分发
// 3. 变量元数据信息的获取
// 4. 支持单元测试的Mock机制
//
// 架构设计：
// - Handler：统一的变量操作入口
// - Store：抽象的存储接口，支持不同的后端实现
// - StoreConfig：存储配置和上下文信息
//
// 使用场景：
// - 工作流节点间的状态共享
// - 用户数据的持久化存储
// - 应用配置的动态管理
// - 系统参数的统一访问
package variable

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/compose"

	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
)

// 全局变量处理器单例（暂未使用，预留扩展）
var variableHandlerSingleton *Handler

// GetVariableHandler 获取变量处理器实例
//
// 返回一个新的变量处理器实例，包含用户、应用、系统三种变量存储。
// 每次调用都会创建新的实例，确保线程安全。
//
// 返回：
//   - *Handler: 变量处理器实例
func GetVariableHandler() *Handler {
	return NewVariableHandler()
}

// Handler 变量处理器
//
// 统一管理三种类型的变量存储，提供统一的变量操作接口。
// 根据变量类型自动路由到相应的存储后端进行操作。
type Handler struct {
	// UserVarStore 用户变量存储器
	// 处理用户级别的持久化变量存储和访问
	UserVarStore Store

	// SystemVarStore 系统变量存储器
	// 处理系统级别的内置变量存储和访问
	SystemVarStore Store

	// AppVarStore 应用变量存储器
	// 处理应用级别的配置变量存储和访问
	AppVarStore Store
}

// Get 获取变量值
//
// 根据变量类型和路径从相应的存储器中获取变量值。
// 支持用户变量、系统变量和应用变量的三种类型路由。
//
// 参数：
//   - ctx: 上下文，用于取消和超时控制
//   - t: 变量类型（用户/系统/应用）
//   - path: 变量路径，支持嵌套访问
//   - opts: 可选的存储配置选项
//
// 返回：
//   - any: 变量值
//   - error: 获取失败的错误信息
//
// 注意：
//   - 路径使用compose.FieldPath格式，如["user", "profile", "name"]
//   - 不同的变量类型会路由到不同的存储后端
func (v *Handler) Get(ctx context.Context, t vo.GlobalVarType, path compose.FieldPath, opts ...OptionFn) (any, error) {
	switch t {
	case vo.GlobalUser:
		return v.UserVarStore.Get(ctx, path, opts...)
	case vo.GlobalSystem:
		return v.SystemVarStore.Get(ctx, path, opts...)
	case vo.GlobalAPP:
		return v.AppVarStore.Get(ctx, path, opts...)
	default:
		return nil, fmt.Errorf("unknown variable type: %v", t)
	}
}

// Set 设置变量值
//
// 根据变量类型和路径将值存储到相应的存储器中。
// 支持用户变量、系统变量和应用变量的三种类型路由。
//
// 参数：
//   - ctx: 上下文，用于取消和超时控制
//   - t: 变量类型（用户/系统/应用）
//   - path: 变量路径，支持嵌套设置
//   - value: 要设置的变量值
//   - opts: 可选的存储配置选项
//
// 返回：
//   - error: 设置失败的错误信息
//
// 注意：
//   - 路径使用compose.FieldPath格式，如["user", "profile", "name"]
//   - 不同的变量类型会路由到不同的存储后端
func (v *Handler) Set(ctx context.Context, t vo.GlobalVarType, path compose.FieldPath, value any, opts ...OptionFn) error {
	switch t {
	case vo.GlobalUser:
		return v.UserVarStore.Set(ctx, path, value, opts...)
	case vo.GlobalSystem:
		return v.SystemVarStore.Set(ctx, path, value, opts...)
	case vo.GlobalAPP:
		return v.AppVarStore.Set(ctx, path, value, opts...)
	default:
		return fmt.Errorf("unknown variable type: %v", t)
	}
}

// Init 初始化所有变量存储器
//
// 对所有配置的变量存储器执行初始化操作。
// 通常在应用启动或第一次使用时调用。
//
// 参数：
//   - ctx: 上下文，用于初始化操作
//
// 返回：
//   - context.Context: 返回传入的上下文（保持接口兼容性）
//
// 注意：
//   - 会依次初始化用户、系统、应用变量存储器
//   - 如果某个存储器为nil，会跳过初始化
func (v *Handler) Init(ctx context.Context) context.Context {
	if v.UserVarStore != nil {
		v.UserVarStore.Init(ctx)
	}

	if v.SystemVarStore != nil {
		v.SystemVarStore.Init(ctx)
	}

	if v.AppVarStore != nil {
		v.AppVarStore.Init(ctx)
	}

	return ctx
}

// StoreInfo 存储器信息
//
// 包含变量存储操作所需的上下文信息，用于确定数据隔离和权限控制。
// 这些信息标识了变量所属的业务实体和连接器。
type StoreInfo struct {
	// AppID 应用ID
	// 可选，用于标识变量所属的应用
	AppID *int64

	// AgentID 代理ID
	// 可选，用于标识变量所属的代理
	AgentID *int64

	// ConnectorID 连接器ID
	// 必填，标识使用的连接器实例
	ConnectorID int64

	// ConnectorUID 连接器唯一标识
	// 连接器的全局唯一标识符
	ConnectorUID string
}

// StoreConfig 存储器配置
//
// 包含变量存储操作的所有配置信息，通过Option模式灵活设置。
type StoreConfig struct {
	// StoreInfo 存储器基本信息
	StoreInfo StoreInfo
}

// OptionFn 配置选项函数类型
//
// 用于灵活配置StoreConfig的函数类型，采用Option模式设计。
type OptionFn func(*StoreConfig)

// WithStoreInfo 设置存储器信息选项
//
// 创建一个配置选项函数，用于设置StoreConfig中的StoreInfo。
//
// 参数：
//   - info: 存储器信息
//
// 返回：
//   - OptionFn: 配置选项函数
func WithStoreInfo(info StoreInfo) OptionFn {
	return func(option *StoreConfig) {
		option.StoreInfo = info
	}
}

//go:generate mockgen -destination varmock/var_mock.go --package mockvar -source variable.go

// Store 变量存储器接口
//
// 定义变量存储器的标准接口，支持不同后端实现的变量存储。
// 通过这个接口抽象，可以灵活地切换不同的存储实现（如内存、Redis、数据库等）。
type Store interface {
	// Init 初始化存储器
	// 在首次使用前调用，进行必要的初始化工作
	Init(ctx context.Context)

	// Get 获取变量值
	// 根据路径获取存储的变量值
	Get(ctx context.Context, path compose.FieldPath, opts ...OptionFn) (any, error)

	// Set 设置变量值
	// 根据路径存储变量值
	Set(ctx context.Context, path compose.FieldPath, value any, opts ...OptionFn) error
}

// 全局变量元数据获取器实例
var variablesMetaGetterImpl VariablesMetaGetter

// GetVariablesMetaGetter 获取变量元数据获取器
//
// 返回变量元数据获取器的实例，用于获取应用和代理的变量类型信息。
//
// 返回：
//   - VariablesMetaGetter: 变量元数据获取器实例
func GetVariablesMetaGetter() VariablesMetaGetter {
	return NewVariablesMetaGetter()
}

// VariablesMetaGetter 变量元数据获取器接口
//
// 提供获取应用和代理变量元数据的功能，用于前端展示和类型验证。
// 元数据包含变量的类型信息、结构定义等。
type VariablesMetaGetter interface {
	// GetAppVariablesMeta 获取应用变量元数据
	// 获取指定应用和版本的所有变量类型信息
	GetAppVariablesMeta(ctx context.Context, id, version string) (m map[string]*vo.TypeInfo, err error)

	// GetAgentVariablesMeta 获取代理变量元数据
	// 获取指定代理和版本的所有变量类型信息
	GetAgentVariablesMeta(ctx context.Context, id int64, version string) (m map[string]*vo.TypeInfo, err error)
}
