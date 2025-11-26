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
 * @file SQL 编辑器设置器导出
 * @description 提供 SQL 语句编辑功能，支持 AI 自动生成
 */

import type { SetterExtension } from '@flowgram-adapter/free-layout-editor';

import { Sql } from './sql';

/**
 * SQL 设置器扩展
 *
 * 用于数据库节点的 SQL 语句编辑，支持：
 * - SQL 语法高亮
 * - AI 自动生成 SQL (NL2SQL)
 */
export const sql: SetterExtension = {
  key: 'sql',
  component: Sql,
};
