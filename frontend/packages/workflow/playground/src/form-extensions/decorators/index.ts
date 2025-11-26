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
 * @file 表单装饰器导出
 * @description 提供表单渲染的装饰器扩展，用于美化和增强表单展示
 */

import { type DecoratorExtension } from '@flowgram-adapter/free-layout-editor';

import { style } from './style';
import { formLayout } from './form-layout';
import { formItemFeedback } from './form-item-feedback';
import { formItem } from './form-item';
import { formCard, formCardAction } from './form-card';
import { columnsTitle } from './columns-title';

/**
 * 表单装饰器集合
 *
 * 包含以下装饰器：
 * - style: 基础样式
 * - formLayout: 表单布局
 * - formCard: 卡片式表单容器
 * - formItem: 表单项
 * - formItemFeedback: 表单项反馈
 * - columnsTitle: 列标题
 */
export const decorators: DecoratorExtension[] = [
  style,
  formLayout,
  formCard,
  formCardAction,
  formItem,
  formItemFeedback,
  columnsTitle,
] as DecoratorExtension[];
