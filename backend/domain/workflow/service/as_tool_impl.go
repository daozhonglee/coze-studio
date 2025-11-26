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

// as_tool_impl.go 工作流作为工具的服务实现
//
// 本文件实现了工作流作为 LLM 工具使用时的相关功能：
//   - 消息管道创建
//   - 执行配置封装
//   - 工具恢复选项
//   - 工作流转工具接口

package service

import (
	"context"

	einoCompose "github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	workflowModel "github.com/coze-dev/coze-studio/backend/crossdomain/workflow/model"
	"github.com/coze-dev/coze-studio/backend/domain/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/execute"
)

// asToolImpl 工作流作为工具的服务实现
type asToolImpl struct {
	repo workflow.Repository
}

// WithMessagePipe 创建消息管道
// 返回执行选项、消息流读取器和清理函数
func (a *asToolImpl) WithMessagePipe() (einoCompose.Option, *schema.StreamReader[*entity.Message], func()) {
	return execute.WithMessagePipe()
}

// WithExecuteConfig 创建带执行配置的选项
func (a *asToolImpl) WithExecuteConfig(cfg workflowModel.ExecuteConfig) einoCompose.Option {
	return einoCompose.WithToolsNodeOption(einoCompose.WithToolOption(execute.WithExecuteConfig(cfg)))
}

// WithResumeToolWorkflow 创建工具恢复选项
// 用于从中断点恢复工作流工具的执行
func (a *asToolImpl) WithResumeToolWorkflow(resumingEvent *entity.ToolInterruptEvent, resumeData string,
	allInterruptEvents map[string]*entity.ToolInterruptEvent) einoCompose.Option {
	toolCallID2ExeID := make(map[string]int64, len(allInterruptEvents))
	for callID, event := range allInterruptEvents {
		toolCallID2ExeID[callID] = event.ExecuteID
	}
	return einoCompose.WithToolsNodeOption(
		einoCompose.WithToolOption(
			execute.WithResume(&entity.ResumeRequest{
				ExecuteID:  resumingEvent.ExecuteID,
				EventID:    resumingEvent.ID,
				ResumeData: resumeData,
			}, toolCallID2ExeID)))
}

// WorkflowAsModelTool 将工作流转换为 LLM 可调用的工具
// 根据查询策略批量获取工作流并转换为工具接口
func (a *asToolImpl) WorkflowAsModelTool(ctx context.Context, policies []*vo.GetPolicy) (tools []workflow.ToolFromWorkflow, err error) {
	for _, id := range policies {
		t, err := a.repo.WorkflowAsTool(ctx, *id, vo.WorkflowToolConfig{})
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}

	return tools, nil
}
