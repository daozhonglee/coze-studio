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
 * @file Fabric 画布模块导出
 * @description 提供基于 Fabric.js 的画布编辑和预览功能
 */

/** Fabric 编辑器组件 */
export { FabricEditor } from './components/fabric-editor';
/** Fabric 预览组件 */
export {
  FabricPreview,
  type IFabricPreview,
} from './components/fabric-preview';
/** 字体加载工具 */
export { loadFont } from './utils/font-loader';
/** 字体树形数据 */
export { fontTreeData } from './assert/font';
// eslint-disable-next-line @coze-arch/no-batch-import-or-export
export * from './typings';
