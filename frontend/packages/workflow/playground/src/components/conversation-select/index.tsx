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
 * @file 会话选择组件
 * @description 用于选择工作流中的会话
 */

import React from 'react';

import { withField } from '@coze-arch/bot-semi';

import { Conversations } from './conversations';

/**
 * 会话选择组件属性
 */
interface ConversationSelectProps {
  value?: string;
  onChange?: (value: string) => void;
}

/**
 * 会话选择组件
 *
 * 提供会话列表选择功能
 */
export const ConversationSelect: React.FC<ConversationSelectProps> = ({
  value,
  onChange,
  ...props
}) => <Conversations value={value} onChange={onChange} {...props} />;

/** 带表单字段包装的会话选择组件 */
export const ConversationSelectWithField = withField(ConversationSelect, {
  valueKey: 'value',
  onKeyChangeFnName: 'onChange',
});

ConversationSelectWithField.defaultProps = {
  fieldStyle: { overflow: 'visible' },
};
