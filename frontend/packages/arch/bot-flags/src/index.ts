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
 * @file 功能开关模块导出
 * @description 提供功能开关 (Feature Flags) 的获取和使用
 */

/** 功能开关类型定义 */
export { type FEATURE_FLAGS, type FetchFeatureGatingFunction } from './types';

/** 获取功能开关 */
export { getFlags } from './get-flags';
/** 功能开关 Hook */
export { useFlags } from './use-flags';
/** 拉取功能开关配置 */
export { pullFeatureFlags } from './pull-feature-flags';
