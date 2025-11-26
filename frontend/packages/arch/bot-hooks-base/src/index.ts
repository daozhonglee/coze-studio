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
 * @file 基础 Hooks 模块导出
 * @description 提供通用的 React Hooks，包括路由配置、响应式、上传等
 */

/** 路由配置 Hook */
export { useRouteConfig, TRouteConfigGlobal } from './use-route-config';
/** 响应式判断 Hook */
export { useIsResponsiveByRouteConfig } from './use-responsive';
/** 登录状态 Hook */
export { useLoggedIn } from './use-loggedin';
export { useLineClamp } from './use-line-clamp';
export { useInitialValue } from './use-initial-value';
export { useExposure, UseExposureParams } from './use-exposure';
export {
  useComponentState,
  ComponentStateUpdateFunc,
} from './use-component-state';
export {
  useDragAndPasteUpload,
  UseDragAndPasteUploadParam,
} from './use-drag-and-paste-upload';
export { useDefaultExPandCheck } from './use-default-expand-check';
export { useResetLocationState } from './router/use-reset-location-state';
export {
  PlacementEnum,
  useLayoutContext,
  LayoutContext,
} from './editor-layout';
export { usePageState, PageStateUpdateFunc } from './use-page-state';
export { useUserSenderInfo } from './bot/use-user-sender-info';
export { useMessageReportEvent } from './bot/use-message-report-event';
