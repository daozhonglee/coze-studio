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

// Package middleware 提供 HTTP 中间件（请求检查）
package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

// RequestAuthTypeStr 请求认证类型的上下文键
const RequestAuthTypeStr = "RequestAuthTypeStr"

// RequestAuthType 请求认证类型
type RequestAuthType = int32

// 请求认证类型常量
const (
	RequestAuthTypeWebAPI     RequestAuthType = 0 // Web API（Session 认证）
	RequestAuthTypeOpenAPI    RequestAuthType = 1 // OpenAPI（Token 认证）
	RequestAuthTypeStaticFile RequestAuthType = 2 // 静态文件（无需认证）
)

// RequestInspectorMW 请求检查中间件
//
// 分析请求路径，确定认证类型（Web API / OpenAPI / 静态文件）
func RequestInspectorMW() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		authType := RequestAuthTypeWebAPI // default is web api, session auth

		if isNeedOpenapiAuth(ctx) {
			authType = RequestAuthTypeOpenAPI
		} else if isStaticFile(ctx) {
			authType = RequestAuthTypeStaticFile
		}

		ctx.Set(RequestAuthTypeStr, authType)
		ctx.Next(c)
	}
}

var staticFilePath = map[string]bool{
	"/static":      true,
	"/":            true,
	"/sign":        true,
	"/favicon.png": true,
}

func isStaticFile(ctx *app.RequestContext) bool {
	path := string(ctx.GetRequest().URI().Path())
	if staticFilePath[path] {
		return true
	}

	if strings.HasPrefix(path, "/static/") ||
		strings.HasPrefix(path, "/explore/") ||
		strings.HasPrefix(path, "/admin/") ||
		strings.HasPrefix(path, "/space/") {
		return true
	}

	if path == "/information/auth/success" {
		return true
	}

	return false
}
