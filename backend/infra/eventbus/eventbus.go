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

// Package eventbus 提供消息总线接口
//
// 本包定义消息队列的生产者和消费者接口，用于：
// - 异步事件发布（工作流事件、资源事件等）
// - 消息消费处理
//
// 实现层在 impl/ 目录下，支持多种消息队列：
// - RocketMQ（默认）
// - NSQ
// - Kafka
// - Pulsar
package eventbus

import "context"

// Producer 消息生产者接口
//
//go:generate  mockgen -destination ../../internal/mock/infra/eventbus/eventbus_mock.go -package mock -source eventbus.go Factory
type Producer interface {
	Send(ctx context.Context, body []byte, opts ...SendOpt) error
	BatchSend(ctx context.Context, bodyArr [][]byte, opts ...SendOpt) error
}

// defaultSVC 默认消费者服务实例
var defaultSVC ConsumerService

// SetDefaultSVC 设置默认消费者服务
func SetDefaultSVC(svc ConsumerService) {
	defaultSVC = svc
}

// GetDefaultSVC 获取默认消费者服务
func GetDefaultSVC() ConsumerService {
	return defaultSVC
}

// ConsumerService 消费者服务接口
type ConsumerService interface {
	RegisterConsumer(nameServer, topic, group string, consumerHandler ConsumerHandler, opts ...ConsumerOpt) error
}

// ConsumerHandler 消息处理器接口
type ConsumerHandler interface {
	HandleMessage(ctx context.Context, msg *Message) error
}

// Message 消息结构体
type Message struct {
	Topic string
	Group string
	Body  []byte
}
