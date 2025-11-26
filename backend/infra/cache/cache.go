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

// Package cache 定义了缓存基础设施层接口
//
// 本包提供缓存操作的抽象接口，包括：
// - 字符串操作（Set/Get/Incr）
// - 哈希操作（HSet/HGetAll）
// - 通用操作（Del/Exists/Expire）
// - 列表操作（LPush/LPop/LRange）
// - 管道操作（Pipeline）
//
// 实现层在 impl/redis/ 目录下，使用 Redis 作为缓存后端
package cache

import (
	"context"
	"time"
)

// Nil 缓存空值错误，用于判断键不存在的情况
var Nil error

// SetDefaultNilError 设置默认的空值错误
func SetDefaultNilError(err error) {
	Nil = err
}

// Cmdable 缓存命令接口，聚合所有缓存操作能力
type Cmdable interface {
	Pipeline() Pipeliner
	StringCmdable
	HashCmdable
	GenericCmdable
	ListCmdable
}

// StringCmdable 字符串操作接口
type StringCmdable interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd
	Get(ctx context.Context, key string) StringCmd
	IncrBy(ctx context.Context, key string, value int64) IntCmd
	Incr(ctx context.Context, key string) IntCmd
}

// HashCmdable 哈希操作接口
type HashCmdable interface {
	HSet(ctx context.Context, key string, values ...interface{}) IntCmd
	HGetAll(ctx context.Context, key string) MapStringStringCmd
}

// GenericCmdable 通用操作接口
type GenericCmdable interface {
	Del(ctx context.Context, keys ...string) IntCmd
	Exists(ctx context.Context, keys ...string) IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) BoolCmd
}

// Pipeliner 管道接口，支持批量执行命令
type Pipeliner interface {
	StatefulCmdable
	Exec(ctx context.Context) ([]Cmder, error)
}

// StatefulCmdable 有状态命令接口
type StatefulCmdable interface {
	Cmdable
}

// ListCmdable 列表操作接口
type ListCmdable interface {
	LIndex(ctx context.Context, key string, index int64) StringCmd
	LPush(ctx context.Context, key string, values ...interface{}) IntCmd
	RPush(ctx context.Context, key string, values ...interface{}) IntCmd
	LSet(ctx context.Context, key string, index int64, value interface{}) StatusCmd
	LPop(ctx context.Context, key string) StringCmd
	LRange(ctx context.Context, key string, start, stop int64) StringSliceCmd
}
type Cmder interface {
	Err() error
}

type baseCmd interface {
	Err() error
}

type IntCmd interface {
	baseCmd
	Result() (int64, error)
}

type MapStringStringCmd interface {
	baseCmd
	Result() (map[string]string, error)
}

type BoolCmd interface {
	baseCmd
	Result() (bool, error)
}

type StatusCmd interface {
	baseCmd
	Result() (string, error)
}

type StringCmd interface {
	baseCmd
	Result() (string, error)
	Val() string
	Int64() (int64, error)
	Bytes() ([]byte, error)
}

type StringSliceCmd interface {
	baseCmd
	Result() ([]string, error)
}
