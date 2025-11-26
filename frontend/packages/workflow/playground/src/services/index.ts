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
 * @file 工作流服务层导出
 * @description 提供工作流编辑器的各种服务，包括运行、保存、验证等
 */

/** 工作流运行服务 - 管理测试运行和执行 */
export { WorkflowRunService } from './workflow-run-service';
/** 测试运行报告服务 - 收集和展示测试结果 */
export { TestRunReporterService } from './test-run-reporter-service';
/** 工作流编辑服务 - 处理编辑操作 */
export { WorkflowEditService } from './workflow-edit-service';

/** 工作流保存服务 - 处理保存和发布 */
export { WorkflowSaveService } from './workflow-save-service';
/** 角色服务 - 管理工作流角色配置 */
export { RoleService } from './role-service';
/** 关联用例数据服务 */
export { RelatedCaseDataService } from './related-case-data-service';

/** Chatflow 服务 - 对话流程管理 */
export { ChatflowService } from './chatflow-service';
/** 节点版本服务 - 管理节点版本 */
export { NodeVersionService } from './node-version-service';
/** 工作流拖拽服务 - 处理节点拖拽 */
export { WorkflowCustomDragService } from './workflow-drag-service';
/** 工作流操作服务 - 复制、粘贴、删除等操作 */
export { WorkflowOperationService } from './workflow-operation-service';
/** 工作流验证服务 - 校验工作流配置 */
export { WorkflowValidationService } from './workflow-validation-service';
/** 工作流模型服务 - 管理可用的 AI 模型 */
export { WorkflowModelsService } from './workflow-models-service';
/** 浮动布局服务 - 管理侧边面板布局 */
export { WorkflowFloatLayoutService } from './workflow-float-layout-service';
/** 值表达式服务接口 */
export { ValueExpressionService } from './value-expression-service';
/** 值表达式服务实现 */
export { ValueExpressionServiceImpl } from './value-expression-service-impl';
/** 数据库节点服务接口 */
export { DatabaseNodeService } from './database-node-service';
/** 数据库节点服务实现 */
export { DatabaseNodeServiceImpl } from './database-node-service-impl';
/** 触发器服务 - 管理工作流触发方式 */
export { TriggerService } from './trigger-service';
/** 插件节点服务 - 管理插件调用 */
export { PluginNodeService, type PluginNodeStore } from './plugin-node-service';

/** 子工作流节点服务 */
export { SubWorkflowNodeService } from '@/node-registries/sub-workflow/services';
/** 工作流依赖服务 - 管理节点间依赖关系 */
export { WorkflowDependencyService } from './workflow-dependency-service';
