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
 * @file Studio 国际化资源模块导出
 * @description 提供 Studio 的多语言资源配置
 */

/* eslint-disable */
// 由 dl-i18n 命令自动生成
import localeEn from './locales/en.json';
import localeZhCN from './locales/zh-CN.json';

/** 默认语言配置 */
const defaultConfig = {
  en: { i18n: localeEn },
  'zh-CN': { i18n: localeZhCN },
} as { en: { i18n: typeof localeEn }; 'zh-CN': { i18n: typeof localeZhCN } };

export { localeEn, localeZhCN, defaultConfig };
export type {
  I18nOptionsMap,
  I18nKeysHasOptionsType,
  I18nKeysNoOptionsType,
  LocaleData,
} from './locale-data';
