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

// Package parser 提供文档解析器接口
//
// 本包定义文档解析器的接口和实现，支持多种文档格式：
// - PDF、DOCX、XLSX、CSV
// - Markdown、JSON、纯文本
// - 图片 OCR 解析
//
// 实现层在 impl/ 目录下，支持内置解析器和 Python 解析器
package parser

import "github.com/cloudwego/eino/components/document/parser"

// Parser 文档解析器接口（复用 Eino 框架定义）
type Parser = parser.Parser
