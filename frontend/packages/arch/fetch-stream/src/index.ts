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
 * @file Fetch Stream 模块导出
 * @description 提供流式请求功能，支持 SSE 和流式响应处理
 */

import { isFetchStreamErrorInfo } from './utils';
import {
  FetchStreamErrorCode,
  type FetchSteamConfig,
  type FetchStreamErrorInfo,
} from './type';
import { fetchStream } from './fetch-stream';

/** 流式请求相关导出 */
export {
  isFetchStreamErrorInfo,
  fetchStream,
  FetchStreamErrorCode,
  type FetchSteamConfig,
  type FetchStreamErrorInfo,
};
