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

// tool.go 工具实体定义
//
// 本文件定义了工具领域的实体类型别名。
// ToolInfo 直接使用跨域模型定义。

package entity

import (
	"github.com/coze-dev/coze-studio/backend/crossdomain/plugin/model"
)

// ToolInfo 工具信息实体
// 直接使用 crossdomain/plugin/model.ToolInfo 的类型别名
type ToolInfo = model.ToolInfo
