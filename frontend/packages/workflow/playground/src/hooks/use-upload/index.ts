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
 * @file 文件上传 Hook 导出
 * @description 提供文件上传功能和相关工具函数
 */

/** 文件上传 Hook 和配置类型 */
export { useUpload, UploadConfig } from './use-upload';
/** 文件处理工具函数 */
export {
  getAccept,
  getFileExtension,
  getBase64,
  getImageSize,
  formatBytes,
} from './utils';
/** 可预览的图片类型常量 */
export { PREVIEW_IMAGE_TYPE } from './constant';

/** 文件项类型和状态枚举 */
export { FileItem, FileItemStatus } from './types';
