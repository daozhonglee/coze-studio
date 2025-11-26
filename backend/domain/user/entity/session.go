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

// Package entity 定义了用户(User)领域的核心实体（会话相关）

package entity

import (
	"time"
)

// SessionKey 会话密钥的存储键名
const SessionKey = "session_key"

// Session 用户会话实体
//
// 表示一个经过验证的用户会话，包含用户身份和会话有效期信息
type Session struct {
	// UserID 用户ID
	UserID int64
	// Locale 用户语言/地区设置
	Locale string
	// UserEmail 用户邮箱
	UserEmail string

	// CreatedAt 会话创建时间
	CreatedAt time.Time
	// ExpiresAt 会话过期时间
	ExpiresAt time.Time
}
