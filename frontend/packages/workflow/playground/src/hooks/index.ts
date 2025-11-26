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
 * @file 工作流 Hooks 导出
 * @description 提供工作流编辑器的 React Hooks
 */

/** 全局状态 Hook - 管理工作流全局状态 */
export { useGlobalState } from './use-global-state';
/** 执行状态实体 Hook */
export { useExecStateEntity } from './use-exec-state-entity';

/** 获取空间 ID */
export { useSpaceId } from './use-space-id';
/** 获取最新工作流 JSON */
export { useLatestWorkflowJson } from './use-latest-workflow-json';

/** 工作流操作 Hook */
export { useWorkflowOperation } from './use-workflow-operation';
/** 连线服务 Hook */
export { useLineService } from './use-line-service';
/** 工作流运行服务 Hook */
export { useWorkflowRunService } from './use-workflow-run-service';
/** 测试运行报告服务 Hook */
export { useTestRunReporterService } from './use-test-run-reporter-service';

/** 滚动到节点 Hook */
export { useScrollToNode } from './use-scroll-to-node';
/** 滚动到连线 Hook */
export { useScrollToLine } from './use-scroll-to-line';
/** 检查是否有协作者 Hook */
export { useHaveCollaborators } from './use-have-collaborators';
/** 节点渲染数据 Hook */
export { useNodeRenderData } from './use-node-render-data';
/** 撤销/重做 Hook */
export { useRedoUndo } from './use-redo-undo';
/** 输入变量 Hook */
export { useInputVariables } from './use-input-variables';
/** 获取工作流模式 Hook */
export { useGetWorkflowMode } from './use-get-workflow-mode';

/** 角色服务 Hook */
export { useRoleService, useRoleServiceStore } from './use-role-service';
/** 文件上传 Hook */
export { useUpload } from './use-upload';
/** 变量服务 Hook */
export { useVariableService } from './use-variable-service';
/** 节点渲染场景 Hook */
export { useNodeRenderScene } from './use-node-render-scene';
/** 测试表单状态 Hook */
export { useTestFormState } from './use-test-form-state';
/** 更新排序后的端口连线 Hook */
export { useUpdateSortedPortLines } from './use-update-sorted-port-lines';
/** 添加节点 Hook */
export { useAddNode } from './use-add-node';
/** 浮动布局服务 Hook */
export {
  useFloatLayoutService,
  useFloatLayoutSize,
} from './use-float-layout-service';
/** 打开追踪列表面板 Hook */
export { useOpenTraceListPanel } from './use-open-trace-list-panel';
/** 测试运行 Hook */
export { useTestRun } from './use-test-run';
/** 数据集信息 Hook */
export { useDataSetInfos } from './use-dataset-info';
/** 节点版本服务 Hook */
export { useNodeVersionService } from './node-version';
/** 保存服务 Hook */
export { useSaveService } from './use-save-service';
/** 数据库节点服务 Hook */
export { useDatabaseNodeService } from './use-database-node-service';
/** 插件节点服务 Hook */
export {
  usePluginNodeStore,
  usePluginNodeService,
} from './use-plugin-node-service';

/** 新建数据库查询 Hook */
export { useNewDatabaseQuery } from './use-new-database-query';
/** 当前数据库查询 Hook */
export { useCurrentDatabaseQuery } from './use-current-database-query';
/** 当前数据库 ID Hook */
export { useCurrentDatabaseID } from './use-current-database-id';

/** 关联 Bot 服务 Hook */
export { useRelatedBotService } from './use-related-bot-service';

/** 工作流预设 Hook */
export { useWorkflowPreset } from './use-workflow-preset';
/** 工作流模型 Hook */
export { useWorkflowModels } from './use-workflow-models';
/** 依赖服务 Hook */
export {
  useDependencyService,
  useDependencyEntity,
} from './use-dependency-service';
