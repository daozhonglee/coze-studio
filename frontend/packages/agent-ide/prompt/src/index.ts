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
 * @file 提示词模块导出
 * @description 提供 Agent 提示词编辑和提示词库功能
 */

/** 提示词视图组件 */
export { type PromptViewProps, PromptView } from './components/prompt-view';
/** 提示词库操作组件 */
export {
  PromptLibrary,
  ImportToLibrary,
} from './components/prompt-view/components/actions';
/** 获取提示词库数据 Hook */
export { useGetLibrarysData } from './hooks/use-prompt/use-get-library-data';
/** 添加到提示词库 Hook */
export { useAddLibrary } from './hooks/use-prompt/use-add-library';
