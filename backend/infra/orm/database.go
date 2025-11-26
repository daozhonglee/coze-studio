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

// Package orm 提供 ORM 数据库接口
//
// 本包定义数据库访问的 ORM 抽象层，基于 GORM 框架
package orm

import (
	"gorm.io/gorm"
)

// DB 数据库连接类型别名（GORM）
type DB = gorm.DB
