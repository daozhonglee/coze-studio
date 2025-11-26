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

// option.go 仓储查询选项
//
// 本文件定义了插件和工具查询时的字段选择选项。
// 使用函数式选项模式，支持按需选择查询字段以优化性能。

package repository

import (
	"github.com/coze-dev/coze-studio/backend/domain/plugin/internal/dal"
)

// PluginSelectedOptions 插件查询字段选择选项
type PluginSelectedOptions func(*dal.PluginSelectedOption)

func WithPluginID() PluginSelectedOptions {
	return func(opts *dal.PluginSelectedOption) {
		opts.PluginID = true
	}
}

func WithPluginOpenapiDoc() PluginSelectedOptions {
	return func(opts *dal.PluginSelectedOption) {
		opts.OpenapiDoc = true
	}
}

func WithPluginManifest() PluginSelectedOptions {
	return func(opts *dal.PluginSelectedOption) {
		opts.Manifest = true
	}
}

func WithPluginIconURI() PluginSelectedOptions {
	return func(opts *dal.PluginSelectedOption) {
		opts.IconURI = true
	}
}

func WithPluginVersion() PluginSelectedOptions {
	return func(opts *dal.PluginSelectedOption) {
		opts.Version = true
	}
}

// ToolSelectedOptions 工具查询字段选择选项
type ToolSelectedOptions func(option *dal.ToolSelectedOption)

func WithToolID() ToolSelectedOptions {
	return func(opts *dal.ToolSelectedOption) {
		opts.ToolID = true
	}
}

func WithToolMethod() ToolSelectedOptions {
	return func(opts *dal.ToolSelectedOption) {
		opts.ToolMethod = true
	}
}

func WithToolSubURL() ToolSelectedOptions {
	return func(opts *dal.ToolSelectedOption) {
		opts.ToolSubURL = true
	}
}

func WithToolActivatedStatus() ToolSelectedOptions {
	return func(opts *dal.ToolSelectedOption) {
		opts.ActivatedStatus = true
	}
}

func WithToolDebugStatus() ToolSelectedOptions {
	return func(opts *dal.ToolSelectedOption) {
		opts.DebugStatus = true
	}
}
