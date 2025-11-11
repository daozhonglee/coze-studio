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

// Package ternary 提供三元运算符功能
//
// Go语言没有内置的三元运算符 (?:)，这个包提供了泛型版本的三元运算符函数。
// 用于简化条件赋值语句，提高代码的可读性。
//
// 使用示例：
//
//	result := ternary.IFElse(age >= 18, "adult", "minor")
//	mode := ternary.IFElse(isChatFlow, workflow.ChatFlow, workflow.Workflow)
package ternary

// IFElse 泛型三元运算符函数
// 根据条件返回对应的值，相当于其他语言中的 condition ? trueValue : falseValue
//
// 参数：
//   - ok: 布尔条件
//   - trueValue: 当条件为true时返回的值
//   - falseValue: 当条件为false时返回的值
//
// 返回值：
//
//	根据条件返回trueValue或falseValue，类型与输入参数一致
//
// 类型参数：
//   - T: 可以是任意类型，支持Go的泛型
func IFElse[T any](ok bool, trueValue, falseValue T) T {
	if ok {
		return trueValue
	}
	return falseValue
}
