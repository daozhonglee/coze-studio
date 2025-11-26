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
 * @file 检索横幅组件
 * @description 显示多分支合并时的检索提示
 */

import { I18n } from '@coze-arch/i18n';
import { Banner } from '@coze-arch/coze-design';
import { Typography } from '@coze-arch/bot-semi';

import { useRetrieve } from './use-retrieve';

const { Text } = Typography;

/**
 * 检索横幅组件
 *
 * 当存在多分支合并情况时，显示提示横幅
 * 允许用户检索查看其他分支的修改
 */
const RetrieveBanner = () => {
  const { showRetrieve, author, handleRetrieve } = useRetrieve();

  if (!showRetrieve || IS_BOT_OP) {
    return null;
  }

  return (
    <Banner
      type="info"
      icon={null}
      closeIcon={null}
      description={
        <Text>
          {I18n.t('workflow_publish_multibranch_merge_comfirm_desc', {
            user_name: author,
          })}
          <Text link onClick={handleRetrieve} style={{ marginLeft: 8 }}>
            {I18n.t('workflow_publish_multibranch_merge_retrieve')}
          </Text>
        </Text>
      }
    />
  );
};

export default RetrieveBanner;
