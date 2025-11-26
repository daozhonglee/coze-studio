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
 * @file PDF.js Shadow 模块导出
 * @description 提供 PDF.js 的封装和类型导出
 */

/** PDF 文档操作 */
export {
  getDocument,
  type PDFDocumentProxy,
  type PDFPageProxy,
  type PageViewport,
} from 'pdfjs-dist';

/** PDF 文本内容类型 */
export { type TextContent } from 'pdfjs-dist/types/src/display/text_layer';
export { type TextItem } from 'pdfjs-dist/types/src/display/api';
export { generatePdfAssetsUrl } from './generate-assets';
export { initPdfJsWorker } from './init-pdfjs-dist';
