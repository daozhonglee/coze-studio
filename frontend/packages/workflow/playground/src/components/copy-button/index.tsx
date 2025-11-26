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
 * @file 复制按钮组件
 * @description 提供一键复制功能，点击后显示成功状态
 */

import React, { useState } from 'react';

import { I18n } from '@coze-arch/i18n';
import { IconCozCopy, IconCozCheckMark } from '@coze-arch/coze-design/icons';
import { Tooltip } from '@coze-arch/coze-design';
import { UIIconButton } from '@coze-arch/bot-semi';

/** 成功状态默认显示时长（毫秒） */
const DELAY = 4000;

/**
 * 复制按钮组件
 *
 * 点击后将内容复制到剪贴板，并切换为成功状态
 * 默认 4 秒后恢复初始状态
 *
 * @param value - 要复制的内容
 * @param delayTime - 成功状态显示时长（可选）
 */
export const CopyButton = ({
  value = '',
  delayTime,
}: {
  value: string;
  delayTime?: number;
}) => {
  const [isSuccess, setSuccess] = useState(false);
  const handleOnClick = e => {
    e.stopPropagation();
    navigator.clipboard.writeText(value as string);
    setSuccess(true);
    setTimeout(() => {
      setSuccess(false);
    }, delayTime ?? DELAY);
  };

  return isSuccess ? (
    <Tooltip content={I18n.t('Duplicate_success')}>
      <UIIconButton
        icon={<IconCozCheckMark color={'rgba(107, 109, 117, 1)'} />}
      />
    </Tooltip>
  ) : (
    <Tooltip content={I18n.t('Copy')}>
      <UIIconButton
        onClick={handleOnClick}
        icon={<IconCozCopy color={'rgba(107, 109, 117, 1)'} />}
      />
    </Tooltip>
  );
};
