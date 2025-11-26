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

// Package plugin 实现插件调用节点
//
// 插件节点用于在工作流中调用外部插件的 API，支持：
// - 调用已安装的插件工具
// - 传递参数和接收返回值
// - 处理授权中断（OAuth 等）
//
// 插件是 Coze 平台的核心扩展机制，允许工作流与外部服务交互，
// 如发送邮件、调用第三方 API、访问数据库等。
package plugin

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudwego/eino/compose"

	"github.com/coze-dev/coze-studio/backend/api/model/app/bot_common"
	workflowModel "github.com/coze-dev/coze-studio/backend/crossdomain/workflow/model"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/canvas/convert"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/execute"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/nodes"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/internal/schema"
	"github.com/coze-dev/coze-studio/backend/pkg/errorx"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/slices"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

// Config 插件节点配置
// 包含调用插件所需的标识信息
type Config struct {
	// PluginID 插件 ID
	PluginID int64
	// ToolID 工具/API ID，一个插件可包含多个工具
	ToolID int64
	// PluginVersion 插件版本号
	PluginVersion string
	// PluginFrom 插件来源，区分市场插件和自定义插件
	PluginFrom *bot_common.PluginFrom
}

func (c *Config) Adapt(ctx context.Context, n *vo.Node, opts ...nodes.AdaptOption) (*schema.NodeSchema, error) {
	ns := &schema.NodeSchema{
		Key:     vo.NodeKey(n.ID),
		Type:    entity.NodeTypePlugin,
		Name:    n.Data.Meta.Title,
		Configs: c,
	}
	inputs := n.Data.Inputs

	apiParams := slices.ToMap(inputs.APIParams, func(e *vo.Param) (string, *vo.Param) {
		return e.Name, e
	})

	ps, ok := apiParams["pluginID"]
	if !ok {
		return nil, fmt.Errorf("plugin id param is not found")
	}

	pID, err := strconv.ParseInt(ps.Input.Value.Content.(string), 10, 64)

	c.PluginID = pID
	c.PluginFrom = inputs.PluginFrom

	ps, ok = apiParams["apiID"]
	if !ok {
		return nil, fmt.Errorf("plugin id param is not found")
	}

	tID, err := strconv.ParseInt(ps.Input.Value.Content.(string), 10, 64)
	if err != nil {
		return nil, err
	}

	c.ToolID = tID

	ps, ok = apiParams["pluginVersion"]
	if !ok {
		return nil, fmt.Errorf("plugin version param is not found")
	}
	version := ps.Input.Value.Content.(string)

	c.PluginVersion = version

	if err := convert.SetInputsForNodeSchema(n, ns); err != nil {
		return nil, err
	}

	if err := convert.SetOutputTypesForNodeSchema(n, ns); err != nil {
		return nil, err
	}

	return ns, nil
}

func (c *Config) Build(_ context.Context, _ *schema.NodeSchema, _ ...schema.BuildOption) (any, error) {
	return &Plugin{
		pluginID:      c.PluginID,
		toolID:        c.ToolID,
		pluginVersion: c.PluginVersion,
		pluginFrom:    c.PluginFrom,
	}, nil
}

type Plugin struct {
	pluginID      int64
	toolID        int64
	pluginVersion string
	pluginFrom    *bot_common.PluginFrom
}

func (p *Plugin) Invoke(ctx context.Context, parameters map[string]any) (ret map[string]any, err error) {
	var exeCfg workflowModel.ExecuteConfig
	if ctxExeCfg := execute.GetExeCtx(ctx); ctxExeCfg != nil {
		exeCfg = ctxExeCfg.ExeCfg
	}
	result, err := ExecutePlugin(ctx, parameters, &vo.PluginEntity{
		PluginID:      p.pluginID,
		PluginVersion: ptr.Of(p.pluginVersion),
		PluginFrom:    p.pluginFrom,
	}, p.toolID, exeCfg)
	if err != nil {
		if extra, ok := compose.IsInterruptRerunError(err); ok {
			// TODO: temporarily replace interrupt with real error, because frontend cannot handle interrupt for now
			interruptData := extra.(*entity.InterruptEvent).InterruptData
			return nil, vo.NewError(errno.ErrAuthorizationRequired, errorx.KV("extra", interruptData))
		}
		return nil, err
	}

	return result, nil

}
