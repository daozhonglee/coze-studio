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
 * @file Slardar 适配器模块导出
 * @description 提供 Slardar 监控的适配器实现
 */

/* eslint-disable @typescript-eslint/no-explicit-any */
import slardarInstance from '@coze-studio/default-slardar';

/** JS 错误插件 */
export const jsErrorPlugin = () => ({});

/** 自定义插件 */
export const customPlugin = () => ({});

/** 创建最小化浏览器客户端 */
export const createMinimalBrowserClient: () => any = () => slardarInstance;
