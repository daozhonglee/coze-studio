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

// Package dynconf 提供动态配置接口
//
// 本包定义动态配置服务的接口，用于：
// - 配置的动态读取和监听
// - 支持多种配置中心（ZooKeeper、etcd、Nacos）
//
// 实现层在 impl/ 目录下
package dynconf

import "context"

// Provider 动态配置提供者接口
//
// 支持 ZooKeeper、etcd、Nacos 等配置中心
type Provider interface {
	Initialize(ctx context.Context, namespace, group string, opts ...Option) (DynamicClient, error)
}

// DynamicClient 动态配置客户端接口
type DynamicClient interface {
	AddListener(key string, callback func(value string, err error)) error
	RemoveListener(key string) error
	Get(ctx context.Context, key string) (string, error)
}

// options 配置选项
type options struct{}

// Option 配置选项函数
type Option struct {
	apply func(opts *options)

	implSpecificOptFn any
}
