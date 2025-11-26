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
 * @file 日志模块导出
 * @description 提供日志记录、错误上报和错误边界功能
 */

/** 上报器 - 向 Slardar 上报日志 */
export { reporter, Reporter } from './reporter';

// Reporter 上报到 Slardar 的类型导出
export type {
  LoggerCommonProperties,
  CustomEvent,
  CustomErrorLog,
  CustomLog,
  ErrorEvent,
} from './reporter';
// 控制台打印
/** 日志打印器 */
export { logger, LoggerContext, Logger } from './logger';

// ErrorBoundary 相关方法
/** 错误边界组件和 Hooks */
export {
  ErrorBoundary,
  useErrorBoundary,
  useErrorHandler,
  type ErrorBoundaryProps,
  type FallbackProps,
} from './error-boundary';

/** Slardar 上报客户端 */
export { SlardarReportClient, type SlardarInstance } from './slardar';

/** 日志级别枚举 */
export { LogLevel } from './types';

/** Slardar 运行时工具 */
export { getSlardarInstance, setUserInfoContext } from './slardar/runtime';
