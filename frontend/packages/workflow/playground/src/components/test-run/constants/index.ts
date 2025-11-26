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
 * @file 测试运行常量导出
 * @description 定义测试运行相关的常量和枚举
 */

export {
  TestFormType,
  FieldName,
  TestRunDataSource,
  SETTING_FIELD_TEMPLATE,
  DEFAULT_FIELD_TEMPLATE,
  NODE_FIELD_TEMPLATE,
  BATCH_FIELD_TEMPLATE,
  INPUT_FIELD_TEMPLATE,
  getBotFieldTemplate,
  getConversationTemplate,
  DATASETS_FIELD_TEMPLATE,
  COMMON_FIELD,
  TYPE_FIELD_MAP,
  TESTSET_CHAT_NAME,
  TESTSET_BOT_NAME,
  INPUT_JSON_FIELD_TEMPLATE,
} from './test-form';

/** 测试集连接器 ID（固定字符串） */
export const TESTSET_CONNECTOR_ID = '10000';

/** 起始节点 ID 标记（仅作标识，不用于判断起始节点） */
export const START_NODE_ID = '100001';

/*******************************************************************************
 * 日志相关常量
 */

/**
 * 结束节点输出方案
 */
export enum EndTerminalPlan {
  Variable = 1,
  Text = 2,
}
