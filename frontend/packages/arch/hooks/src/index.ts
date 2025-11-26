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
 * @file 通用 Hooks 导出
 * @description 提供常用的 React Hooks
 */

/** 鼠标悬停状态 Hook */
export { default as useHover } from './use-hover';
/** 持久化回调函数 Hook */
export { default as usePersistCallback } from './use-persist-callback';
/** 更新时执行副作用 Hook */
export { default as useUpdateEffect } from './use-update-effect';
/** 布尔值切换 Hook */
export { default as useToggle } from './use-toggle';
/** URL 参数 Hook */
export { default as useUrlParams } from './use-url-params';
/** 实时状态 Hook */
export { default as useStateRealtime } from './use-state-realtime';
