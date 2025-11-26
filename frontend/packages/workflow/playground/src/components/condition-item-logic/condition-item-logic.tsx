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
 * @file 条件逻辑选择组件
 * @description 提供 AND/OR 逻辑选择功能，用于条件节点
 */

import { type FC } from 'react';

import { ConditionLogic } from '@coze-workflow/base';
import { I18n } from '@coze-arch/i18n';
import { Select } from '@coze-arch/coze-design';

import { logicTextMap } from './constants';

import styles from './condition-item-logic.module.less';

/**
 * 条件逻辑选择组件属性
 */
export interface ConditionItemLogicProps {
  /** 逻辑类型（AND 或 OR） */
  logic?: ConditionLogic;
  /** 逻辑变更回调 */
  onChange: (logic: ConditionLogic) => void;
  /** 是否显示连接线 */
  showStroke?: boolean;
  className?: string;
  /** 是否只读 */
  readonly?: boolean;
  testId?: string;
}

/**
 * 条件逻辑选择组件
 *
 * 用于在条件节点中选择多个条件之间的逻辑关系（AND/OR）
 */
export const ConditionItemLogic: FC<ConditionItemLogicProps> = props => {
  const {
    logic,
    onChange,
    showStroke = false,
    readonly = false,
    testId,
  } = props;

  return (
    <div className="flex flex-col pt-[16px] pb-[16px] w-[50px]">
      <div className="flex-1 relative">
        {showStroke ? (
          <div className="absolute left-1/2 right-0 top-2.5 bottom-0 rounded-tl-lg border-solid border-0 border-t border-l coz-stroke-plus" />
        ) : null}
      </div>
      <Select
        className={styles['condition-logic-select']}
        placeholder={I18n.t('workflow_detail_condition_pleaseselect')}
        style={{ marginRight: 4 }}
        value={logic}
        disabled={readonly}
        size="small"
        optionList={[
          {
            label: logicTextMap.get(ConditionLogic.AND),
            value: ConditionLogic.AND,
          },
          {
            label: logicTextMap.get(ConditionLogic.OR),
            value: ConditionLogic.OR,
          },
        ]}
        onChange={val => onChange(val as ConditionLogic)}
        data-testid={testId}
      />
      <div className="flex-1 relative">
        {showStroke ? (
          <div className="absolute left-1/2 right-0 top-0 bottom-2.5 rounded-bl-lg border-solid border-0 border-b border-l coz-stroke-plus" />
        ) : null}
      </div>
    </div>
  );
};
