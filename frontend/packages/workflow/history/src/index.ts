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
 * @file 工作流历史记录模块导出
 * @description 提供工作流编辑历史记录和撤销/重做功能
 */

/** 历史记录容器模块 */
export { WorkflowHistoryContainerModule } from './workflow-history-container-module';
/** 清除历史记录 Hook */
export { useClearHistory } from './hooks/use-clear-history';
/** 历史记录配置 */
export { WorkflowHistoryConfig } from './workflow-history-config';
/** 操作上报插件 */
export { createOperationReportPlugin } from './create-operation-report-plugin';

/** 历史服务和插件 */
export {
  HistoryService,
  createFreeHistoryPlugin,
} from '@flowgram-adapter/free-layout-editor';
