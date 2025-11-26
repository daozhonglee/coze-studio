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

package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/coze-dev/coze-studio/backend/api/model/data/variable/kvmemory"
	"github.com/coze-dev/coze-studio/backend/api/model/data/variable/project_memory"
	"github.com/coze-dev/coze-studio/backend/domain/memory/variables/entity"
	"github.com/coze-dev/coze-studio/backend/domain/memory/variables/repository"
	"github.com/coze-dev/coze-studio/backend/pkg/errorx"
	"github.com/coze-dev/coze-studio/backend/pkg/i18n"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ternary"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

// sysVariableConf 中文系统变量配置
var sysVariableConf []*kvmemory.VariableInfo = []*kvmemory.VariableInfo{
	{
		Key:                  "sys_uuid",
		Description:          "用户唯一ID",
		DefaultValue:         "",
		Example:              "",
		ExtDesc:              "",
		GroupDesc:            "用户请求/授权后系统自动获取的相关数据",
		GroupExtDesc:         "",
		GroupName:            "用户信息",
		Sensitive:            "false",
		CanWrite:             "false",
		MustNotUseInPrompt:   "false",
		EffectiveChannelList: []string{"全渠道"},
	},
}

// sysVariableConfEN 英文系统变量配置
var sysVariableConfEN []*kvmemory.VariableInfo = []*kvmemory.VariableInfo{
	{
		Key:                  "sys_uuid",
		Description:          "User uniq ID",
		DefaultValue:         "",
		Example:              "",
		ExtDesc:              "",
		GroupDesc:            "Data automatically retrieved by the system after user request or authorization.",
		GroupExtDesc:         "",
		GroupName:            "User information",
		Sensitive:            "false",
		CanWrite:             "false",
		MustNotUseInPrompt:   "false",
		EffectiveChannelList: []string{"All publication channels"},
	},
}

// variablesImpl 变量服务实现
type variablesImpl struct {
	// Repo 变量仓储
	Repo repository.VariableRepository
}

// NewService 创建变量服务实例
func NewService(repo repository.VariableRepository) Variables {
	return &variablesImpl{
		Repo: repo,
	}
}

// GetSysVariableConf 获取系统预定义变量配置
//
// 根据请求上下文的语言设置返回对应语言的系统变量配置。
func (v *variablesImpl) GetSysVariableConf(ctx context.Context) entity.SysConfVariables {
	if i18n.GetLocale(ctx) == i18n.LocaleEN {
		return sysVariableConfEN
	}
	return sysVariableConf
}

// GetProjectVariablesMeta 获取 Project 的变量元数据
func (v *variablesImpl) GetProjectVariablesMeta(ctx context.Context, projectID, version string) (*entity.VariablesMeta, error) {
	return v.GetVariableMeta(ctx, projectID, project_memory.VariableConnector_Project, version)
}

// UpsertProjectMeta 创建或更新 Project 的变量元数据
func (v *variablesImpl) UpsertProjectMeta(ctx context.Context, projectID, version string, userID int64, e *entity.VariablesMeta) (int64, error) {
	return v.upsertVariableMeta(ctx, projectID, project_memory.VariableConnector_Project, version, userID, e)
}

// UpsertMeta 创建或更新变量元数据
func (v *variablesImpl) UpsertMeta(ctx context.Context, e *entity.VariablesMeta) (int64, error) {
	return v.upsertVariableMeta(ctx, e.BizID, e.BizType, e.Version, e.CreatorID, e)
}

// UpsertBotMeta 创建或更新 Agent 的变量元数据
func (v *variablesImpl) UpsertBotMeta(ctx context.Context, agentID int64, version string, userID int64, e *entity.VariablesMeta) (int64, error) {
	bizID := fmt.Sprintf("%d", agentID)
	return v.upsertVariableMeta(ctx, bizID, project_memory.VariableConnector_Bot, version, userID, e)
}

// upsertVariableMeta 创建或更新变量元数据的内部实现
//
// 如果元数据不存在则创建，存在则更新。
func (v *variablesImpl) upsertVariableMeta(ctx context.Context, bizID string, bizType project_memory.VariableConnector, version string, userID int64, e *entity.VariablesMeta) (int64, error) {
	meta, err := v.Repo.GetVariableMeta(ctx, bizID, bizType, version)
	if err != nil {
		return 0, err
	}

	do := &entity.VariablesMeta{
		BizID:     bizID,
		Version:   version,
		CreatorID: int64(userID),
		BizType:   bizType,
		Variables: e.Variables,
	}

	if meta == nil {
		return v.Repo.CreateVariableMeta(ctx, do, bizType)
	}

	do.ID = meta.ID
	err = v.Repo.UpdateProjectVariable(ctx, do, bizType)
	if err != nil {
		return 0, err
	}

	return meta.ID, nil
}

// mergeVariableList 合并系统变量和用户变量列表
//
// 用户变量可以覆盖同名的系统变量。
// 返回结果按原始顺序排序，系统变量排在最后。
func (*variablesImpl) mergeVariableList(_ context.Context, sysVarsList, variablesList []*entity.VariableMeta) *entity.VariablesMeta {
	mergedMap := make(map[string]*entity.VariableMeta)
	for _, sysVar := range sysVarsList {
		mergedMap[sysVar.Keyword] = sysVar
	}

	// Can overwrite sysVar
	for _, variable := range variablesList {
		mergedMap[variable.Keyword] = variable
	}

	res := make([]*entity.VariableMeta, 0)
	for _, variable := range mergedMap {
		res = append(res, variable)
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Channel == project_memory.VariableChannel_System && !(res[j].Channel == project_memory.VariableChannel_System) {
			return false
		}
		if !(res[i].Channel == project_memory.VariableChannel_System) && res[j].Channel == project_memory.VariableChannel_System {
			return true
		}
		indexI := -1
		indexJ := -1

		for index, s := range sysVarsList {
			if s.Keyword == res[i].Keyword {
				indexI = index
			}
			if s.Keyword == res[j].Keyword {
				indexJ = index
			}
		}

		for index, s := range variablesList {
			if s.Keyword == res[i].Keyword && indexI < 0 {
				indexI = index
			}
			if s.Keyword == res[j].Keyword && indexJ < 0 {
				indexJ = index
			}
		}
		return indexI < indexJ
	})

	return &entity.VariablesMeta{
		Variables: res,
	}
}

// GetAgentVariableMeta 获取 Agent 的变量元数据
func (v *variablesImpl) GetAgentVariableMeta(ctx context.Context, agentID int64, version string) (*entity.VariablesMeta, error) {
	bizID := fmt.Sprintf("%d", agentID)
	return v.GetVariableMeta(ctx, bizID, project_memory.VariableConnector_Bot, version)
}

// GetVariableMetaByID 根据 ID 获取变量元数据
func (v *variablesImpl) GetVariableMetaByID(ctx context.Context, id int64) (*entity.VariablesMeta, error) {
	do, err := v.Repo.GetVariableMetaByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if do == nil {
		return nil, nil
	}

	return do, nil
}

// GetVariableMeta 获取变量元数据
//
// 合并系统变量和用户定义的变量，设置默认 Schema 和只读属性。
func (v *variablesImpl) GetVariableMeta(ctx context.Context, bizID string, bizType project_memory.VariableConnector, version string) (*entity.VariablesMeta, error) {
	data, err := v.Repo.GetVariableMeta(ctx, bizID, bizType, version)
	if err != nil {
		return nil, err
	}

	sysVarMeta := v.GetSysVariableConf(ctx)
	if bizType == project_memory.VariableConnector_Project {
		sysVarMeta.RemoveLocalChannelVariable()
	}

	sysVarMetaList := sysVarMeta.ToVariables()

	if data == nil {
		return sysVarMetaList, nil
	}

	resVarMetaList := v.mergeVariableList(ctx, sysVarMetaList.Variables, data.Variables)
	resVarMetaList.SetupSchema()
	resVarMetaList.SetupIsReadOnly()

	return resVarMetaList, nil
}

// DeleteAllVariable 删除业务对象下的所有变量数据
func (v *variablesImpl) DeleteAllVariable(ctx context.Context, bizType project_memory.VariableConnector, bizID string) (err error) {
	return v.Repo.DeleteAllVariableData(ctx, bizType, bizID)
}

// DeleteVariableInstance 删除用户的变量值
//
// 系统变量不允许删除，会自动从 keywords 中过滤掉。
func (v *variablesImpl) DeleteVariableInstance(ctx context.Context, e *entity.UserVariableMeta, keywords []string) (err error) {
	// if e.BizType == int32(project_memory.VariableConnector_Project) {
	// 	keywords = v.removeProjectSysVariable(ctx, keywords)
	// } else {
	// 	keywords, err = v.removeAgentSysVariable(ctx, keywords, e.BizID)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	keywords = v.removeSysVariable(ctx, keywords)

	if len(keywords) == 0 {
		return errorx.New(errno.ErrMemoryNoVariableCanBeChangedCode)
	}

	return v.Repo.DeleteVariableInstance(ctx, e, keywords)
}

// removeAgentSysVariable 从关键字列表中移除 Agent 的系统变量
func (v *variablesImpl) removeAgentSysVariable(ctx context.Context, keywords []string, biz_id string) ([]string, error) {
	vars, err := v.GetVariableMeta(ctx, biz_id, project_memory.VariableConnector_Bot, "")
	if err != nil {
		return nil, err
	}

	sysKeywords := make(map[string]bool)
	for _, v := range vars.Variables {
		if v.Channel == project_memory.VariableChannel_System {
			sysKeywords[v.Keyword] = true
		}
	}

	if len(sysKeywords) == 0 {
		return keywords, nil
	}

	filteredKeywords := make([]string, 0)
	for _, v := range keywords {
		if sysKeywords[v] {
			continue
		}
		filteredKeywords = append(filteredKeywords, v)
	}

	return filteredKeywords, nil
}

// removeSysVariable 从关键字列表中移除系统变量
func (v *variablesImpl) removeSysVariable(ctx context.Context, keywords []string) []string {
	sysConf := v.GetSysVariableConf(ctx)
	sysVarsMap := make(map[string]bool)
	for _, v := range sysConf {
		sysVarsMap[v.Key] = true
	}

	filteredKeywords := make([]string, 0)
	for _, v := range keywords {
		if sysVarsMap[v] {
			continue
		}

		filteredKeywords = append(filteredKeywords, v)
	}

	return filteredKeywords
}

// GetVariableInstance 获取用户的变量值
func (v *variablesImpl) GetVariableInstance(ctx context.Context, e *entity.UserVariableMeta, keywords []string) ([]*kvmemory.KVItem, error) {
	return v.GetVariableChannelInstance(ctx, e, keywords, nil)
}

// GetVariableChannelInstance 获取指定渠道的用户变量值
//
// 根据元数据和实例数据合并变量值，系统变量值动态生成。
// 如果变量没有实例值，则使用默认值。
func (v *variablesImpl) GetVariableChannelInstance(ctx context.Context, e *entity.UserVariableMeta, keywords []string, varChannel *project_memory.VariableChannel) ([]*kvmemory.KVItem, error) {
	meta, err := v.GetVariableMeta(ctx, e.BizID, project_memory.VariableConnector(e.BizType), e.Version)
	if err != nil {
		return nil, err
	}

	if varChannel != nil {
		meta.FilterChannelVariable(*varChannel)
	}

	meta.RemoveDisableVariable()

	metaKey2Variable := map[string]*entity.VariableMeta{}
	metaKeys := make([]string, 0, len(meta.Variables))
	for _, variable := range meta.Variables {
		metaKeys = append(metaKeys, variable.Keyword)
		metaKey2Variable[variable.Keyword] = variable
	}

	kvInstances, err := v.Repo.GetVariableInstances(ctx, e, keywords)
	if err != nil {
		return nil, err
	}

	varBothInMetaAndInstance := map[string]*entity.VariableInstance{}
	for _, v := range kvInstances {
		if _, ok := metaKey2Variable[v.Keyword]; ok {
			varBothInMetaAndInstance[v.Keyword] = v
		}
	}

	newKeywords := ternary.IFElse(len(keywords) > 0, keywords, metaKeys)

	resMemory := make([]*kvmemory.KVItem, 0, len(newKeywords))
	for _, v := range newKeywords {
		if vv, ok := varBothInMetaAndInstance[v]; ok {
			meta := metaKey2Variable[v]
			resMemory = append(resMemory, &kvmemory.KVItem{
				Keyword:        vv.Keyword,
				Value:          vv.Content,
				CreateTime:     vv.CreatedAt / 1000,
				UpdateTime:     vv.UpdatedAt / 1000,
				Schema:         meta.Schema,
				IsSystem:       meta.IsSystem(),
				PromptDisabled: meta.PromptDisabled,
			})
		} else if vv, ok := metaKey2Variable[v]; ok { // only in meta
			now := time.Now()
			resMemory = append(resMemory, &kvmemory.KVItem{
				Keyword:        vv.Keyword,
				Value:          vv.DefaultValue,
				CreateTime:     now.Unix(),
				UpdateTime:     now.Unix(),
				Schema:         vv.Schema,
				IsSystem:       vv.IsSystem(),
				PromptDisabled: vv.PromptDisabled,
			})
		}
	}

	sysKVItems, err := v.getSysKVItems(ctx, meta, e)

	res := v.mergeKVItem(resMemory, sysKVItems)
	res = v.sortKVItem(res, meta)

	return res, nil
}

// getAppKVItems 获取 APP 渠道的变量键值对
func (v *variablesImpl) getAppKVItems(_ context.Context, meta *entity.VariablesMeta) ([]*kvmemory.KVItem, error) {
	resMemory := []*kvmemory.KVItem{}

	for _, v := range meta.Variables {
		if v.Channel == project_memory.VariableChannel_APP {
			resMemory = append(resMemory, &kvmemory.KVItem{
				Keyword: v.Keyword,
				Value:   v.DefaultValue,
				Schema:  v.Schema,
			})
		}
	}

	return resMemory, nil
}

// getSysKVItems 获取系统变量的键值对
//
// 系统变量值是动态生成的，如 sys_uuid 会根据用户信息生成唯一标识。
func (v *variablesImpl) getSysKVItems(ctx context.Context, meta *entity.VariablesMeta, e *entity.UserVariableMeta) ([]*kvmemory.KVItem, error) {
	sysKVItems := []*kvmemory.KVItem{}

	for _, variable := range meta.Variables {
		if variable.Channel == project_memory.VariableChannel_System {
			sysKV, err := e.GenSystemKV(ctx, variable.Keyword)
			if err != nil {
				return nil, err
			}

			if sysKV != nil {
				sysKVItems = append(sysKVItems, sysKV)
			}
		}
	}

	return sysKVItems, nil
}

// mergeKVItem 合并用户变量和系统变量的键值对
//
// 系统变量优先级更高，同名的用户变量会被忽略。
func (v *variablesImpl) mergeKVItem(user []*kvmemory.KVItem, sys []*kvmemory.KVItem) []*kvmemory.KVItem {
	res := make([]*kvmemory.KVItem, 0, len(user))
	sysMap := make(map[string]bool)
	for _, v := range sys {
		res = append(res, v)
		sysMap[v.Keyword] = true
	}

	for _, v := range user {
		if sysMap[v.Keyword] {
			continue
		}
		res = append(res, v)
	}

	return res
}

// sortKVItem 对变量键值对排序
//
// 按元数据中定义的顺序排序，系统变量排在最后。
func (v *variablesImpl) sortKVItem(items []*kvmemory.KVItem, meta *entity.VariablesMeta) []*kvmemory.KVItem {
	sort.Slice(items, func(ii, jj int) bool {
		i := items[ii]
		j := items[jj]

		// If they are all system variables, there is no need to change positions here
		if i.IsSystem && !j.IsSystem {
			return false
		}
		if !i.IsSystem && j.IsSystem {
			return true
		}

		indexI := -1
		indexJ := -1

		for index, s := range meta.Variables {
			if s.Keyword == i.Keyword {
				indexI = index
			}
			if s.Keyword == j.Keyword {
				indexJ = index
			}
		}

		return indexI < indexJ
	})

	return items
}

// SetVariableInstance 设置用户的变量值
//
// 系统变量不允许设置。已存在的变量实例执行更新，不存在的执行插入。
// 返回更新的变量关键字列表。
func (v *variablesImpl) SetVariableInstance(ctx context.Context, e *entity.UserVariableMeta, items []*kvmemory.KVItem) ([]string, error) {
	meta, err := v.GetVariableMeta(ctx, e.BizID, project_memory.VariableConnector(e.BizType), e.Version)
	if err != nil {
		return nil, err
	}

	filerItems := v.filterKVItem(items, meta)
	if len(filerItems) == 0 {
		return nil, errorx.New(errno.ErrMemorySetKvMemoryItemInstanceCode)
	}

	keywords := make([]string, 0, len(filerItems))
	key2Item := make(map[string]*kvmemory.KVItem, len(filerItems))
	for _, v := range filerItems {
		keywords = append(keywords, v.Keyword)
		key2Item[v.Keyword] = v
	}

	kvInstances, err := v.Repo.GetVariableInstances(ctx, e, keywords)
	if err != nil {
		return nil, err
	}

	needUpdateKeywords := make([]string, 0, len(kvInstances))
	needUpdateKVs := make([]*entity.VariableInstance, 0, len(kvInstances))
	for _, v := range kvInstances {
		if vv, ok := key2Item[v.Keyword]; ok {
			needUpdateKeywords = append(needUpdateKeywords, v.Keyword)
			v.Content = vv.Value
			needUpdateKVs = append(needUpdateKVs, v)
			delete(key2Item, v.Keyword)
		}
	}

	err = v.Repo.UpdateVariableInstance(ctx, needUpdateKVs)
	if err != nil {
		return nil, err
	}

	needIndexKVs := make([]*entity.VariableInstance, 0, len(key2Item))
	for _, v := range key2Item {
		needIndexKVs = append(needIndexKVs, &entity.VariableInstance{
			BizType:      e.BizType,
			BizID:        e.BizID,
			Version:      e.Version,
			ConnectorID:  e.ConnectorID,
			ConnectorUID: e.ConnectorUID,
			Type:         int32(project_memory.VariableType_KVVariable),
			Keyword:      v.Keyword,
			Content:      v.Value,
		})
	}

	err = v.Repo.InsertVariableInstance(ctx, needIndexKVs)
	if err != nil {
		return nil, err
	}

	return needUpdateKeywords, nil
}

// filterKVItem 过滤变量键值对
//
// 只保留元数据中定义的非系统变量。
func (v *variablesImpl) filterKVItem(items []*kvmemory.KVItem, meta *entity.VariablesMeta) []*kvmemory.KVItem {
	metaKey2Variable := map[string]*entity.VariableMeta{}
	for _, variable := range meta.Variables {
		metaKey2Variable[variable.Keyword] = variable
	}

	res := make([]*kvmemory.KVItem, 0, len(items))
	for _, v := range items {
		vv, ok := metaKey2Variable[v.Keyword]
		if ok && vv.Channel != project_memory.VariableChannel_System {
			res = append(res, v)
		}
	}

	return res
}

// PublishMeta 发布变量元数据到指定版本
//
// 复制元数据并设置新版本号，用于 Agent 发布时的变量同步。
func (v *variablesImpl) PublishMeta(ctx context.Context, variableMetaID int64, version string) (int64, error) {
	e, err := v.Repo.GetVariableMetaByID(ctx, variableMetaID)
	if err != nil {
		return 0, err
	}
	if e == nil {
		return 0, fmt.Errorf("%d,variable meta not found", variableMetaID)
	}

	e.Version = version
	return v.Repo.CreateVariableMeta(ctx, e, project_memory.VariableConnector(e.BizType))
}

// DecryptSysUUIDKey 解密系统用户唯一标识
//
// 从加密的 sys_uuid 值中解析出原始的用户信息。
func (v *variablesImpl) DecryptSysUUIDKey(ctx context.Context, encryptSysUUIDKey string) *entity.VariableInstance {
	meta := &entity.UserVariableMeta{}
	return meta.DecryptSysUUIDKey(ctx, encryptSysUUIDKey)
}
