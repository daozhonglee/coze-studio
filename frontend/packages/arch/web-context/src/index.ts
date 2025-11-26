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
 * @file Web 上下文模块导出
 * @description 提供全局变量、事件总线、路由跳转等 Web 上下文功能
 */

/** 路由跳转 */
export { redirect } from './location';
/** 全局事件总线 */
export { GlobalEventBus } from './event-bus';
/** 全局变量 */
export { globalVars } from './global-var';
/** 错误码常量 */
export { COZE_TOKEN_INSUFFICIENT_ERROR_CODE } from './const/custom';
/** 应用枚举 */
export { BaseEnum, SpaceAppEnum } from './const/app';

// 社区 Bot 详情业务场景特定键值
/** 默认会话键值 */
export {
  defaultConversationKey,
  defaultConversationUniqId,
} from './const/community';
