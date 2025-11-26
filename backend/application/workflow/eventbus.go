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

// Package workflow 定义了工作流(Workflow)应用层服务（事件总线）

package workflow

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/api/model/resource/common"
	"github.com/coze-dev/coze-studio/backend/domain/search/entity"
	search "github.com/coze-dev/coze-studio/backend/domain/search/entity"
	"github.com/coze-dev/coze-studio/backend/domain/search/service"
)

// eventBus 资源事件总线
var eventBus service.ResourceEventBus

// setEventBus 设置事件总线
func setEventBus(bus service.ResourceEventBus) {
	eventBus = bus
}

// PublishWorkflowResource 发布工作流资源事件
//
// 当工作流创建、更新、删除时，通过事件总线通知搜索服务更新索引
func PublishWorkflowResource(ctx context.Context, workflowID int64, mode *int32, op search.OpType, r *search.ResourceDocument) error {
	if r == nil {
		r = &search.ResourceDocument{}
	}

	r.ResType = common.ResType_Workflow
	r.ResID = workflowID
	r.ResSubType = mode

	event := &entity.ResourceDomainEvent{
		OpType:   entity.OpType(op),
		Resource: r,
	}

	if op == search.Created {
		event.Resource.CreateTimeMS = r.CreateTimeMS
		event.Resource.UpdateTimeMS = r.UpdateTimeMS
	} else if op == search.Updated {
		event.Resource.UpdateTimeMS = r.UpdateTimeMS
	}

	err := eventBus.PublishResources(ctx, event)
	if err != nil {
		return err
	}

	return nil
}
