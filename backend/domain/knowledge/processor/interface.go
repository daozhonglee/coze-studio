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

// interface.go 文档处理器接口定义
//
// 本文件定义了文档处理器接口 DocProcessor。
// 处理器负责文档的创建、存储和索引流程。

package processor

import "github.com/coze-dev/coze-studio/backend/domain/knowledge/entity"

// DocProcessor 文档处理器接口
//
// 定义文档处理的完整流程：
//   - BeforeCreate: 准备数据源
//   - BuildDBModel: 构建数据库模型
//   - InsertDBModel: 插入数据库记录
//   - Indexing: 发起索引任务
//   - GetResp: 获取处理结果
type DocProcessor interface {
	BeforeCreate() error         // 准备数据源
	BuildDBModel() error         // 构建数据库记录
	InsertDBModel() error        // 插入数据库记录
	Indexing() error             // 发起索引任务
	GetResp() []*entity.Document // 返回处理后的文档信息
}
