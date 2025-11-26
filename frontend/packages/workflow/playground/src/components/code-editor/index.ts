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

/**
 * @file 代码编辑器组件导出
 * @description 提供代码编辑器相关组件，包括代码编辑器和文本编辑器
 */

/** 编辑器上下文提供者 */
export { EditorProvider } from '@coze-editor/editor/react';
/** 代码编辑器组件 - 支持语法高亮和代码补全 */
export { CodeEditor } from './code-editor';
/** 文本编辑器组件 - 基础文本编辑功能 */
export { TextEditor } from './text-editor';
