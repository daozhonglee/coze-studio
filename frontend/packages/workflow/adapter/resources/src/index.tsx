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
 * @file 资源适配器模块导出
 * @description 提供工作流中各类资源的适配组件和 Hooks
 */

/** 音频播放器 Hook */
export { useAudioPlayer } from './audio';

/** 语音选择弹窗 Hook */
export { useSelectVoiceModal } from './voice';

/** 协作者和权限相关 */
export { CollaboratorsBtn, getIsCozePro, useCozeProRightsStore } from './auth';

/** 知识库弹窗 */
export { DouyinKnowledgeListModal } from './knowledge';

export {
  NLPromptButton,
  NLPromptModal,
  NlPromptAction,
  NlPromptShortcut,
  NLPromptProvider,
} from './prompt';

export { PublishWorkflowModal, usePublishWorkflowModal } from './store';

export { useWorkflowPublishEntry } from './market/use-workflow-publish-entry';
