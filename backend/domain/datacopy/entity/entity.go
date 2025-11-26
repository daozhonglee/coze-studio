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

// Package entity 定义了数据复制(DataCopy)领域的核心实体
//
// 本包包含数据复制任务相关的领域实体，
// 用于跟踪和管理跨空间/应用的数据复制操作。
package entity

// CopyDataTask 数据复制任务实体
//
// 表示一个数据复制任务，记录源数据和目标数据的关联关系
type CopyDataTask struct {
	// TaskUniqKey 任务唯一标识键
	TaskUniqKey string
	// OriginDataID 源数据ID
	OriginDataID int64
	// TargetDataID 目标数据ID
	TargetDataID int64
	// OriginSpaceID 源工作空间ID
	OriginSpaceID int64
	// TargetSpaceID 目标工作空间ID
	TargetSpaceID int64
	// OriginUserID 源用户ID
	OriginUserID int64
	// TargetUserID 目标用户ID
	TargetUserID int64
	// OriginAppID 源应用ID
	OriginAppID int64
	// TargetAppID 目标应用ID
	TargetAppID int64
	// Status 任务状态
	Status DataCopyTaskStatus
	// DataType 数据类型
	DataType DataType
	// StartTime 任务开始时间（毫秒时间戳）
	StartTime int64
	// FinishTime 任务完成时间（毫秒时间戳）
	FinishTime int64
	// ExtInfo 扩展信息
	ExtInfo string
	// ErrorMsg 复制失败时的错误信息
	ErrorMsg string
}

// DataCopyTaskStatus 数据复制任务状态
type DataCopyTaskStatus int

const (
	// DataCopyTaskStatusCreate 已创建
	DataCopyTaskStatusCreate DataCopyTaskStatus = 1
	// DataCopyTaskStatusInProgress 进行中
	DataCopyTaskStatusInProgress DataCopyTaskStatus = 2
	// DataCopyTaskStatusSuccess 成功
	DataCopyTaskStatusSuccess DataCopyTaskStatus = 3
	// DataCopyTaskStatusFail 失败
	DataCopyTaskStatusFail DataCopyTaskStatus = 4
)

// DataType 数据类型
type DataType int

const (
	// DataTypeKnowledge 知识库
	DataTypeKnowledge DataType = 1
	// DataTypeDatabase 数据库
	DataTypeDatabase DataType = 2
	// DataTypeVariable 变量
	DataTypeVariable DataType = 3
)
