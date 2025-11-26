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
 * @file 工作流 Playground 模块主入口
 * @description 提供工作流编辑器的核心组件、Hooks、服务和类型定义
 */

/** 工作流全局状态 */
export { WorkflowGlobalState } from './entities';
/** 工作流全局状态实体类型 */
export { WorkflowGlobalStateEntity } from './typing';
/** 工作流 Playground 主组件 */
export { WorkflowPlayground } from './workflow-playground';
export {
  useGlobalState,
  useSpaceId,
  useLatestWorkflowJson,
  useGetWorkflowMode,
  useAddNode,
} from './hooks';

export { useAddNodeVisibleStore } from './hooks/use-add-node-visible';

export {
  useInnerSideSheetStoreShallow,
  useSingletonInnerSideSheet,
} from './components/workflow-inner-side-sheet';
export { TestFormDefaultValue } from './components/test-run/types';
export { DND_ACCEPT_KEY } from './constants';
export { WorkflowCustomDragService, WorkflowEditService } from './services';
export {
  AddNodeRef,
  HandleAddNode,
  WorkflowInfo,
  WorkflowPlaygroundProps,
  WorkflowPlaygroundRef,
  NodeTemplate,
  type ProjectApi,
} from './typing';
export { WorkflowPlaygroundContext } from './workflow-playground-context';
import AddOperation from './ui-components/add-operation';
export { AddOperation };
export { TooltipContent as ResourceRefTooltip } from './components/workflow-header/components/reference-modal/tooltip-content';
export { useWorkflowPlayground } from './use-workflow-playground';
export {
  navigateResource,
  LinkNode,
} from './components/workflow-header/components';
export { usePluginDetail } from './node-registries/plugin';
