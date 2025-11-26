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

// global_handler.go 全局回调处理器
//
// 本文件提供工作流执行的全局回调处理器访问接口。

package service

import (
	"github.com/cloudwego/eino/callbacks"

	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/execute"
)

// GetTokenCallbackHandler 获取 Token 统计回调处理器
// 用于收集和统计工作流执行过程中的 Token 使用量
func GetTokenCallbackHandler() callbacks.Handler {
	return execute.GetTokenCallbackHandler()
}
