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
 * @file 工作流渲染模块导出
 * @description 提供工作流画布的渲染相关组件和服务
 */

import 'reflect-metadata';
/** 渲染提供者 */
export * from './workflow-render-provider';
/** 渲染贡献者 */
export * from './workflow-render-contribution';
/** 端口渲染组件 */
export * from './components/workflow-port-render';
/** 快捷键贡献者 */
export * from './workflow-shorcuts-contribution';
/** 渲染注册表 */
export {
  FlowRendererKey,
  FlowRendererRegistry,
  FlowRendererContribution,
} from '@flowgram-adapter/free-layout-editor';
export * from './constants/lines';
