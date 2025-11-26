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

// Package repository 定义了应用(APP)领域的仓储接口（查询选项）

package repository

import (
	"github.com/coze-dev/coze-studio/backend/domain/app/internal/dal"
)

// APPSelectedOptions 应用查询选项函数类型
//
// 使用函数选项模式，允许调用者灵活选择需要查询的字段，
// 减少不必要的数据传输和处理开销
type APPSelectedOptions func(*dal.APPSelectedOption)

// WithAPPID 选择查询应用ID
func WithAPPID() APPSelectedOptions {
	return func(opts *dal.APPSelectedOption) {
		opts.APPID = true
	}
}

// WithAPPPublishAtMS 选择查询发布时间
func WithAPPPublishAtMS() APPSelectedOptions {
	return func(opts *dal.APPSelectedOption) {
		opts.PublishAtMS = true
	}
}

// WithPublishVersion 选择查询发布版本
func WithPublishVersion() APPSelectedOptions {
	return func(opts *dal.APPSelectedOption) {
		opts.PublishVersion = true
	}
}

// WithPublishRecordID 选择查询发布记录ID
func WithPublishRecordID() APPSelectedOptions {
	return func(opts *dal.APPSelectedOption) {
		opts.PublishRecordID = true
	}
}

// WithAPPPublishStatus 选择查询发布状态
func WithAPPPublishStatus() APPSelectedOptions {
	return func(opts *dal.APPSelectedOption) {
		opts.PublishStatus = true
	}
}

// WithPublishRecordExtraInfo 选择查询发布记录额外信息
func WithPublishRecordExtraInfo() APPSelectedOptions {
	return func(opts *dal.APPSelectedOption) {
		opts.PublishRecordExtraInfo = true
	}
}
