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

// Package coderunner 提供代码执行器接口
//
// 本包定义代码运行服务的接口，用于：
// - 执行 Python/JavaScript 代码
// - 支持沙箱执行和直接执行两种模式
//
// 实现层在 impl/ 目录下
package coderunner

import "context"

// Language 编程语言类型
type Language string

// 支持的编程语言常量
const (
	Python     Language = "Python"
	JavaScript Language = "JavaScript"
)

// RunRequest 代码执行请求
type RunRequest struct {
	Code     string
	Params   map[string]any
	Language Language
}

// RunResponse 代码执行响应
type RunResponse struct {
	Result map[string]any
}

// Runner 代码执行器接口
//
//go:generate mockgen -destination  ../../internal/mock/domain/workflow/crossdomain/code/code_mock.go  --package code  -source code.go
type Runner interface {
	Run(ctx context.Context, request *RunRequest) (*RunResponse, error)
}

// GetCodeRunner 获取代码执行器实例
func GetCodeRunner() Runner {
	return runnerImpl
}

// SetCodeRunner 设置代码执行器实例
func SetCodeRunner(runner Runner) {
	runnerImpl = runner
}

// runnerImpl 代码执行器单例
var runnerImpl Runner
