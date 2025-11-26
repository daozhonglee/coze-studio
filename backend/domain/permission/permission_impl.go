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

// Package permission 定义了权限(Permission)领域的服务层实现
//
// 注意：当前实现为开放模式，默认允许所有操作。
// 在生产环境中应根据实际需求实现权限校验逻辑。
package permission

import (
	"context"
)

// permissionImpl 权限服务实现
type permissionImpl struct{}

// NewService 创建权限服务实例
func NewService() Permission {
	return &permissionImpl{}
}

func DefaultSVC() Permission {
	return NewService()
}

func (p *permissionImpl) CheckAuthz(ctx context.Context, req *CheckAuthzData) (*CheckAuthzResult, error) {

	authzChecker := NewAuthzChecker()

	for _, resourceIdentifier := range req.ResourceIdentifier {
		allowed, err := authzChecker.CheckResourcePermission(ctx, &ResourcePermissionRequest{
			ResourceType: resourceIdentifier.Type,
			ResourceIDs:  resourceIdentifier.ID,
			Action:       resourceIdentifier.Action,
			OperatorID:   req.OperatorID,
			IsDraft:      req.IsDraft,
		})
		if err != nil {
			return nil, err
		}

		if !allowed {
			return &CheckAuthzResult{Decision: Deny}, nil
		}
	}

	return &CheckAuthzResult{Decision: Allow}, nil
}
