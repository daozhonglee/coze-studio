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
 * @file 聊天历史轮数选择组件
 * @description 用于配置 LLM 节点的历史对话轮数
 */

import React from 'react';

import { I18n } from '@coze-arch/i18n';
import { IconCozInfoCircle } from '@coze-arch/coze-design/icons';
import { Tooltip, CozInputNumber } from '@coze-arch/coze-design';

/**
 * 聊天历史轮数组件属性
 */
export interface ChatHistoryRoundProps {
  value?: number;
  onChange?: (value: number) => void;
  readonly?: boolean;
}

/** 最小轮数 */
const MIN_ROUND = 1;
/** 最大轮数 */
const MAX_ROUND = 30;

/**
 * 聊天历史轮数选择组件
 *
 * 用于配置 LLM 节点引用的历史对话轮数，范围 1-30 轮
 */
export const ChatHistoryRound = ({
  value,
  onChange,
  readonly,
}: ChatHistoryRoundProps) => (
  <div className="absolute right-[0] top-[9px] flex items-center gap-[4px]">
    <span className="text-xs">{I18n.t('wf_history_rounds')}</span>
    <Tooltip content={I18n.t('model_config_history_round_explain')}>
      <IconCozInfoCircle className="coz-fg-dim text-xs" />
    </Tooltip>

    <CozInputNumber
      className="w-[60px]"
      size="small"
      min={MIN_ROUND}
      max={MAX_ROUND}
      disabled={readonly}
      value={value}
      onChange={w => {
        if (isNaN(w as number)) {
          return;
        }
        onChange?.(w as number);
      }}
    />
  </div>
);
