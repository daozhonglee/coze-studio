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

// Package permission 定义了权限(Permission)领域的服务层
//
// 本包提供权限验证服务，支持：
// - 资源操作权限检查
// - Agent 操作权限检查
// - 工作空间访问权限检查
package permission

import (
	"context"
)

type ResourceIdentifier struct {
	Type   ResourceType
	ID     []int64
	Action Action
}

// ActionAndResource 操作和资源组合
type ActionAndResource struct {
	Action             Action
	ResourceIdentifier ResourceIdentifier
}

type CheckAuthzData struct {
	ResourceIdentifier []*ResourceIdentifier
	OperatorID         int64
	IsDraft            *bool
}
type CheckAuthzResult struct {
	Decision Decision
}

// Permission 权限服务接口
//
// 定义各类权限检查操作
type Permission interface {
	CheckAuthz(ctx context.Context, req *CheckAuthzData) (*CheckAuthzResult, error)
}
