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
 * @file 工作流变量模块导出
 * @description 提供工作流变量管理的核心功能、组件和服务
 */

/* eslint-disable @coze-arch/no-batch-import-or-export */
/** 流节点变量数据 */
export { FlowNodeVariableData } from '@flowgram-adapter/free-layout-editor';
/** 变量相关 Hooks */
export * from './hooks';
// 旧版变量引擎代码，待替换...
/** 旧版兼容代码 */
export * from './legacy';
/** 变量类型定义 */
export * from './typings';
/** 变量核心逻辑 */
export * from './core';
/** 变量相关组件 */
export * from './components';
/** 变量数据模型 */
export * from './datas';
/** 变量表单扩展 */
export * from './form-extensions';
/** 变量常量 */
export * from './constants';
/** 变量服务 */
export * from './services';
/** 生成输入 JSON Schema */
export { generateInputJsonSchema } from './utils/generate-input-json-schema';
/** 创建工作流变量插件 */
export { createWorkflowVariablePlugins } from './create-workflow-variable-plugin';
