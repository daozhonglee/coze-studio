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

// Package idgen 提供 ID 生成器接口
//
// 本包定义分布式 ID 生成器的接口，用于生成全局唯一 ID：
// - 支持单个 ID 生成
// - 支持批量 ID 生成
//
// 实现层在 impl/idgen/ 目录下，使用雪花算法
package idgen

import (
	"context"
)

// IDGenerator ID 生成器接口
//
//go:generate mockgen -destination ../../internal/mock/infra/idgen/idgen_mock.go --package mock -source idgen.go
type IDGenerator interface {
	GenID(ctx context.Context) (int64, error)
	GenMultiIDs(ctx context.Context, counts int) ([]int64, error) // suggest batch size <= 200
}
