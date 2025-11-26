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
 * @file 错误处理模块导出
 * @description 提供自定义错误类型、错误捕获 Hooks 和路由错误处理
 */

/** 自定义错误类和判断函数 */
export { CustomError, isCustomError } from './custom-error';

/** 错误捕获 Hook */
export { useErrorCatch } from './use-error-catch';

/** Chunk 加载错误判断 */
export { isChunkError } from './source-error';

/** 路由错误捕获 Hook */
export { useRouteErrorCatch } from './use-route-error-catch';
