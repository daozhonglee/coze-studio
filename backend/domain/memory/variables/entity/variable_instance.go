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

package entity

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/coze-dev/coze-studio/backend/api/model/data/variable/kvmemory"
	"github.com/coze-dev/coze-studio/backend/api/model/data/variable/project_memory"
	variables "github.com/coze-dev/coze-studio/backend/crossdomain/variables/model"
	"github.com/coze-dev/coze-studio/backend/pkg/errorx"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

// UserVariableMeta 用户变量元数据
//
// 包装 crossdomain 中的用户变量元数据，添加领域行为。
// 用于标识变量实例所属的用户和连接器信息。
type UserVariableMeta struct {
	*variables.UserVariableMeta
}

// NewUserVariableMeta 创建用户变量元数据
func NewUserVariableMeta(v *variables.UserVariableMeta) *UserVariableMeta {
	return &UserVariableMeta{
		UserVariableMeta: v,
	}
}

// VariableInstance 变量实例（运行时值）
//
// 表示用户在运行时设置的变量值，与变量元数据对应。
// 通过 BizType+BizID+ConnectorUID+ConnectorID+Keyword 唯一标识。
type VariableInstance struct {
	// ID 实例 ID
	ID int64
	// BizType 业务类型（Bot/Project）
	BizType project_memory.VariableConnector
	// BizID 业务 ID
	BizID string
	// Version 版本号
	Version string
	// Keyword 变量关键字
	Keyword string
	// Type 变量类型
	Type int32
	// Content 变量值
	Content string
	// ConnectorUID 连接器用户唯一标识
	ConnectorUID string
	// ConnectorID 连接器 ID
	ConnectorID int64
	// CreatedAt 创建时间戳（毫秒）
	CreatedAt int64
	// UpdatedAt 更新时间戳（毫秒）
	UpdatedAt int64
}

const (
	// sysUUIDKey 系统用户唯一标识变量的关键字
	sysUUIDKey string = "sys_uuid"
)

// GenSystemKV 生成系统变量的键值对
//
// 目前仅支持 sys_uuid 系统变量，用于生成用户唯一标识。
func (v *UserVariableMeta) GenSystemKV(ctx context.Context, keyword string) (*kvmemory.KVItem, error) {
	if keyword != sysUUIDKey { // The outfield only supports this one variable for the time being
		return nil, nil
	}

	return v.genUUID(ctx)
}

// genUUID 生成用户唯一标识的键值对
//
// 将 BizType、BizID、ConnectorUID、ConnectorID 组合并 Base64 编码，
// 作为用户在该 Agent/Project 下的唯一标识。
func (v *UserVariableMeta) genUUID(ctx context.Context) (*kvmemory.KVItem, error) {
	if v.BizID == "" {
		return nil, errorx.New(errno.ErrMemoryGetSysUUIDInstanceCode, errorx.KV("msg", "biz_id is empty"))
	}

	if v.ConnectorUID == "" {
		return nil, errorx.New(errno.ErrMemoryGetSysUUIDInstanceCode, errorx.KV("msg", "connector_uid is empty"))
	}

	if v.ConnectorID == 0 {
		return nil, errorx.New(errno.ErrMemoryGetSysUUIDInstanceCode, errorx.KV("msg", "connector_id is empty"))
	}

	encryptSysUUIDKey := v.encryptSysUUIDKey(ctx)
	now := time.Now().Unix()

	return &kvmemory.KVItem{
		Keyword:    sysUUIDKey,
		Value:      encryptSysUUIDKey,
		Schema:     stringSchema,
		CreateTime: now,
		UpdateTime: now,
		IsSystem:   true,
	}, nil
}

// encryptSysUUIDKey 加密系统用户标识
//
// 使用 Base64 编码将用户信息组合成唯一标识字符串。
func (v *UserVariableMeta) encryptSysUUIDKey(ctx context.Context) string {
	// Combine four fields with a special delimiter (e.g. |)
	plain := fmt.Sprintf("%d|%s|%s|%d", v.BizType, v.BizID, v.ConnectorUID, v.ConnectorID)
	return base64.StdEncoding.EncodeToString([]byte(plain))
}

// DecryptSysUUIDKey 解密系统用户标识
//
// 从加密的 sys_uuid 值中解析出原始的用户信息。
// 用于数据库表的权限控制（限制用户只能访问自己的数据）。
func (v *UserVariableMeta) DecryptSysUUIDKey(ctx context.Context, encryptSysUUIDKey string) *VariableInstance {
	data, err := base64.StdEncoding.DecodeString(encryptSysUUIDKey)
	if err != nil {
		return nil
	}

	parts := strings.Split(string(data), "|")
	if len(parts) != 4 {
		return nil
	}

	bizType64, _ := strconv.ParseInt(parts[0], 10, 32)
	connectorID, _ := strconv.ParseInt(parts[3], 10, 64)
	return &VariableInstance{
		BizType:      project_memory.VariableConnector(bizType64),
		BizID:        parts[1],
		ConnectorUID: parts[2],
		ConnectorID:  connectorID,
	}
}
