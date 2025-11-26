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

// Package ocr 提供 OCR 光学字符识别接口
//
// 本包定义 OCR 服务的接口，支持从图片中提取文字：
// - 支持 Base64 编码图片
// - 支持 URL 远程图片
//
// 实现层在 impl/ 目录下，支持 PaddleOCR 和火山引擎 OCR
package ocr

import "context"

// OCR 光学字符识别接口
type OCR interface {
	FromBase64(ctx context.Context, b64 string) (texts []string, err error)
	FromURL(ctx context.Context, url string) (texts []string, err error)
}
