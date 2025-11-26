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
 * @file 可调整大小的侧边面板组件
 * @description 节点侧边窗口，支持宽度调整
 */

import { type FC } from 'react';

import { type ResizableProps, Resizable } from 're-resizable';

import { useResizable } from './use-resizable';

/**
 * 可调整大小的侧边面板属性
 */
interface ResizableSidePanelProps extends ResizableProps {
  bypass?: boolean;
}

/**
 * 可调整大小的侧边面板组件
 *
 * 包装 Resizable 组件，提供节点配置面板的宽度调整功能
 */
export const ResizableSidePanel: FC<ResizableSidePanelProps> = ({
  children,
  bypass,
  ...props
}) => {
  const resizable = useResizable();

  if (bypass) {
    return <>{children}</>;
  }

  return <Resizable {...{ ...resizable, ...props }}>{children}</Resizable>;
};
