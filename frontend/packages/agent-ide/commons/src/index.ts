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
 * @file Agent IDE 公共模块导出
 * @description 提供 Agent IDE 的公共组件、Hooks 和工具函数
 */

/** 差异节点渲染组件 */
export { DiffNodeRender } from './components/diff-node-render';

/** 发布服务条款组件 */
export { PublishTermService } from './components/term-service';

/** 发送差异事件 Hook */
export { useSendDiffEvent } from './hooks/use-send-diff-event';

/** 工具函数 */
export { sendTeaEventInBot, transTimestampText } from './utils';

/** 差异表格缩进常量 */
export { DIFF_TABLE_INDENT_BASE, DIFF_TABLE_INDENT_LENGTH } from './constants';
