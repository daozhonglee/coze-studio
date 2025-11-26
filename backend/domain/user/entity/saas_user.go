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

// Package entity 定义了用户(User)领域的核心实体（SaaS 权益相关）

package entity

// BenefitType 权益类型
//
// 定义 SaaS 平台支持的各种权益类型
type BenefitType string

const (
	// BenefitTypeCallToolLimit 工具调用次数限制
	BenefitTypeCallToolLimit BenefitType = "call_tool_limit"
	// BenefitTypeAPIRunQPS API 运行 QPS 限制
	BenefitTypeAPIRunQPS BenefitType = "api_run_qps"
)

// UserLevel 用户等级（字符串类型，与 API 匹配）
type UserLevel string

const (
	// UserLevelUnknown 未知等级
	UserLevelUnknown UserLevel = "unknown"
	// UserLevelBasic 基础版
	UserLevelBasic UserLevel = "basic"
	// UserLevelPro 专业版
	UserLevelPro UserLevel = "v1_pro_instance"
	// UserLevelEnterprise 企业版
	UserLevelEnterprise UserLevel = "enterprise"
)

// EntityBenefitStatus 权益实体状态（字符串类型，与 API 匹配）
type EntityBenefitStatus string

const (
	// EntityBenefitStatusUnknown 未知状态
	EntityBenefitStatusUnknown EntityBenefitStatus = "unknown"
	// EntityBenefitStatusValid 有效
	EntityBenefitStatusValid EntityBenefitStatus = "valid"
	// EntityBenefitStatusExpired 已过期
	EntityBenefitStatusExpired EntityBenefitStatus = "expired"
)

// ResourceUsageStrategy 资源使用策略（字符串类型，与 API 匹配）
type ResourceUsageStrategy string

const (
	// ResourceUsageStrategyUnknown 未知策略
	ResourceUsageStrategyUnknown ResourceUsageStrategy = "unknown"
	// ResourceUsageStrategyByQuota 按配额使用
	ResourceUsageStrategyByQuota ResourceUsageStrategy = "quota"
	// ResourceUsageStrategyUnlimit 无限制使用
	ResourceUsageStrategyUnlimit ResourceUsageStrategy = "unlimit"
)

// GetEnterpriseBenefitRequest 获取企业权益请求
type GetEnterpriseBenefitRequest struct {
	// BenefitType 权益类型
	BenefitType *string `json:"benefit_type,omitempty" form:"benefit_type"`
	// ResourceID 资源ID
	ResourceID *string `json:"resource_id,omitempty" form:"resource_id"`
}

// GetEnterpriseBenefitResponse 获取企业权益响应
type GetEnterpriseBenefitResponse struct {
	// Code 响应码
	Code int32 `json:"code"`
	// Message 响应消息
	Message string `json:"message"`
	// Data 权益数据
	Data *BenefitData `json:"data,omitempty"`
}

// BenefitData 权益数据
type BenefitData struct {
	// BasicInfo 基本信息（包含用户等级）
	BasicInfo *BasicInfo `json:"basic_info,omitempty"`
	// BenefitInfo 权益信息列表
	BenefitInfo []*BenefitInfo `json:"benefit_info,omitempty"`
}

// BasicInfo 基本信息
type BasicInfo struct {
	// UserLevel 用户等级
	UserLevel UserLevel `json:"user_level,omitempty"`
}

// BenefitInfo 权益信息
type BenefitInfo struct {
	// BenefitType 权益类型
	BenefitType BenefitType `json:"benefit_type,omitempty"`
	// Basic 基础权益配置
	Basic *BenefitTypeInfoItem `json:"basic,omitempty"`
	// Effective 生效的权益配置
	Effective *BenefitTypeInfoItem `json:"effective,omitempty"`
	// ResourceID 资源ID
	ResourceID string `json:"resource_id,omitempty"`
}

// BenefitTypeInfoItem 权益类型信息项
type BenefitTypeInfoItem struct {
	// ItemID 项目ID
	ItemID string `json:"item_id,omitempty"`
	// ItemInfo 项目计数器信息
	ItemInfo *CommonCounter `json:"item_info,omitempty"`
	// Status 状态
	Status EntityBenefitStatus `json:"status,omitempty"`
	// BenefitID 权益ID
	BenefitID string `json:"benefit_id,omitempty"`
}

// CommonCounter 通用计数器
//
// 用于记录资源使用情况
type CommonCounter struct {
	// Used 已使用量（仅在 Strategy == ByQuota 时有效）
	Used float64 `json:"used,omitempty"`
	// Total 总配额（仅在 Strategy == ByQuota 时有效）
	Total float64 `json:"total,omitempty"`
	// Strategy 资源使用策略
	Strategy ResourceUsageStrategy `json:"strategy,omitempty"`
	// StartAt 开始时间（秒时间戳）
	StartAt int64 `json:"start_at,omitempty"`
	// EndAt 结束时间（秒时间戳）
	EndAt int64 `json:"end_at,omitempty"`
}

// String 返回权益类型的字符串表示
func (bt BenefitType) String() string {
	return string(bt)
}

// String 返回用户等级的字符串表示
func (ul UserLevel) String() string {
	return string(ul)
}

// String 返回权益状态的字符串表示
func (ebs EntityBenefitStatus) String() string {
	return string(ebs)
}

// String 返回资源使用策略的字符串表示
func (rus ResourceUsageStrategy) String() string {
	return string(rus)
}

// IsValid 验证用户等级是否有效
func (ul UserLevel) IsValid() bool {
	switch ul {
	case UserLevelUnknown, UserLevelBasic, UserLevelPro, UserLevelEnterprise:
		return true
	default:
		return false
	}
}

// IsValid 验证权益状态是否有效
func (ebs EntityBenefitStatus) IsValid() bool {
	switch ebs {
	case EntityBenefitStatusUnknown, EntityBenefitStatusValid, EntityBenefitStatusExpired:
		return true
	default:
		return false
	}
}

// IsValid 验证资源使用策略是否有效
func (rus ResourceUsageStrategy) IsValid() bool {
	switch rus {
	case ResourceUsageStrategyUnknown, ResourceUsageStrategyByQuota, ResourceUsageStrategyUnlimit:
		return true
	default:
		return false
	}
}
