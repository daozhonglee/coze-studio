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
 * @file Nodes V2 模块导出
 * @description 新版节点系统，提供更灵活的节点定义和表单渲染
 */

/** V2 节点类型常量和类型集合 */
export { NODES_V2, NODE_V2_TYPES } from './constants';

/** V2 节点工具函数 */
export { isNodeV2, isNodeV2registry, getNodeV2Registry } from './utils';

/** 默认节点元数据 Hook */
export { useDefaultNodeMeta } from './hooks/use-default-node-meta';
/** 节点头部组件 */
export { NodeHeader, type NodeHeaderValue } from './components/node-header';
