# Quickstart: 项目代码中文注释添加

**Feature**: 001-add-chinese-comments  
**Date**: 2025-11-25

## 概述

本指南帮助你快速开始为 Coze Studio 项目添加中文注释。

## 前置条件

- 已克隆 coze-studio 仓库
- 了解项目的基本结构
- 熟悉 TypeScript/React 和 Go 的注释语法

## 快速开始步骤

### 1. 了解项目结构

```text
coze-studio/
├── frontend/              # 前端代码 (~9000 文件)
│   └── packages/
│       ├── arch/          # Level-1: 基础架构
│       ├── common/        # Level-2: 共享组件
│       ├── workflow/      # Level-3: 工作流功能
│       ├── agent-ide/     # Level-3: Agent IDE
│       └── studio/        # Level-3: Studio 功能
├── backend/               # 后端代码 (~1000 文件)
│   ├── domain/            # 领域层
│   ├── application/       # 应用层
│   ├── api/               # API 层
│   └── infra/             # 基础设施层
└── docker/                # Docker 配置
```

### 2. 确定注释范围

**排除以下文件**：
- `**/auto-generated/**` - 自动生成的代码
- `**/node_modules/**` - 第三方依赖
- `**/dist/**`, `**/build/**` - 构建产物
- 包含 "DO NOT EDIT" 标记的文件

### 3. 选择工作批次

按优先级选择要处理的模块：

| 优先级 | 模块 | 路径 |
|--------|------|------|
| P1 | 后端 domain 层 | `backend/domain/` |
| P1 | 后端 application 层 | `backend/application/` |
| P1 | 前端 workflow | `frontend/packages/workflow/` |
| P2 | 后端 api/infra 层 | `backend/api/`, `backend/infra/` |
| P2 | 前端 arch/common | `frontend/packages/arch/`, `frontend/packages/common/` |

### 4. 添加注释

#### Go 代码示例

**原始代码**：
```go
package entity

type Workflow struct {
    ID       int64
    CommitID string
}
```

**添加注释后**：
```go
// Package entity 定义了工作流领域的核心实体
//
// 本包包含工作流聚合根及其关联的值对象，是工作流领域的核心。
package entity

// Workflow 工作流聚合根，表示一个完整的工作流定义
//
// 工作流是工作流领域的核心实体，包含工作流的标识和版本信息。
type Workflow struct {
    // ID 工作流唯一标识
    ID int64
    // CommitID 当前提交版本号，用于版本追踪
    CommitID string
}
```

#### TypeScript/React 示例

**原始代码**：
```typescript
export const createApiNodeInfo = (
  apiParams: Partial<PluginApi> | undefined,
  templateIcon?: string,
): ApiNodeDataDTO => {
  // ...
};
```

**添加注释后**：
```typescript
/**
 * 创建 API 节点信息
 * 
 * 根据插件 API 参数创建工作流中的 API 节点数据传输对象。
 * 用于在工作流画布上添加新的 API 调用节点。
 * 
 * @param apiParams - 插件 API 参数，包含 API ID、名称、插件信息等
 * @param templateIcon - 可选的节点图标 URL
 * @returns API 节点数据传输对象
 * 
 * @example
 * ```typescript
 * const nodeInfo = createApiNodeInfo({
 *   api_id: '123',
 *   name: 'Send Message',
 *   plugin_id: 'plugin-456'
 * });
 * ```
 */
export const createApiNodeInfo = (
  apiParams: Partial<PluginApi> | undefined,
  templateIcon?: string,
): ApiNodeDataDTO => {
  // ...
};
```

### 5. 检查注释质量

#### 自检清单

- [ ] 包/模块有顶级注释说明职责
- [ ] 公共函数/组件有 JSDoc/Go doc
- [ ] 复杂逻辑有行内注释
- [ ] 注释准确反映代码功能
- [ ] 无错别字和语法错误

#### 使用脚本检查覆盖率

```bash
# 统计已添加中文注释的 Go 文件
find backend -name "*.go" -exec grep -l "[\u4e00-\u9fff]" {} \; | wc -l

# 统计已添加中文注释的 TypeScript 文件
find frontend -name "*.ts" -o -name "*.tsx" | xargs grep -l "[\u4e00-\u9fff]" | wc -l
```

### 6. 参考已有注释

项目中已有部分优秀的中文注释示例：

**后端示例**：
- `backend/domain/workflow/entity/workflow.go` - 包和结构体注释
- `backend/domain/workflow/variable/variable.go` - 接口和方法注释

**前端示例**：
- `frontend/packages/workflow/playground/src/hooks/` - Hook 注释

## 常见问题

### Q: 英文注释需要替换吗？

A: 不需要。保留原有英文注释，在其旁边或下方添加中文注释作为补充。

### Q: 配置文件需要添加注释吗？

A: 需要，但优先级较低（P3）。配置文件的注释应说明每个配置项的用途和有效值范围。

### Q: 测试文件需要添加注释吗？

A: 暂时不需要，测试文件将作为单独的后续任务处理。

### Q: 如何判断文件是否为自动生成？

A: 检查以下特征：
1. 路径包含 `auto-generated`、`generated`、`__generated__`
2. 文件头部包含 "DO NOT EDIT"、"auto-generated" 等标记
3. 文件是 `.d.ts` 类型声明文件（通常是生成的）

## 下一步

1. 阅读完整的 [注释规范](./contracts/comment-standards.md)
2. 选择一个批次开始工作
3. 完成后进行自检
4. 提交 PR 进行代码审查

