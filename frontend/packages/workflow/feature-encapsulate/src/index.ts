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
 * @file 工作流封装功能模块导出
 * @description 提供工作流节点封装为子流程的功能
 */

/** 工作流封装插件创建函数 */
export { createWorkflowEncapsulatePlugin } from './create-workflow-encapsulate-plugin';
/** 封装服务 */
export { EncapsulateService } from './encapsulate';
/** 封装面板组件 */
export { EncapsulatePanel } from './render';
/** 封装快捷键 */
export { ENCAPSULATE_SHORTCUTS } from './render';
