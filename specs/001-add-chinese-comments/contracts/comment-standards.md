# Comment Standards Contract: 中文注释规范

**Feature**: 001-add-chinese-comments  
**Date**: 2025-11-25  
**Version**: 1.0.0

## 概述

本文档定义了 Coze Studio 项目中文注释的标准格式和规范，作为所有注释添加工作的契约。

---

## Go 代码注释规范

### 1. 包注释 (Package Comment)

```go
// Package [包名] [一句话描述包的核心功能]
//
// [详细描述，2-5行，说明：]
// - 包的主要职责
// - 包含的核心概念/类型
// - 与其他包的关系（如有）
//
// 设计说明：
// [可选：解释设计决策、架构选择等]
//
// 使用示例：
//
//	[示例代码，如有必要]
package packagename
```

**示例**:

```go
// Package entity 定义了工作流领域的核心实体
//
// 本包包含工作流相关的所有领域实体和值对象：
// - Workflow: 工作流聚合根
// - Node: 工作流节点
// - Edge: 节点间的连接关系
//
// 设计说明：
// 采用组合模式，通过嵌入多个VO来构建完整的实体，实现职责分离和灵活组合。
package entity
```

### 2. 结构体注释 (Struct Comment)

```go
// [结构体名] [一句话描述结构体的职责]
//
// [详细描述，说明：]
// - 主要用途
// - 关键字段说明
// - 使用场景
type StructName struct {
    // [字段名] [字段说明]
    FieldName Type
}
```

**示例**:

```go
// Workflow 工作流聚合根，表示一个完整的工作流定义
//
// 工作流包含元数据、画布信息和版本信息，是工作流领域的核心实体。
// 通过组合多个值对象来实现完整的工作流表示。
type Workflow struct {
    // ID 工作流唯一标识
    ID int64
    // CommitID 当前提交版本号
    CommitID string
    // Meta 工作流元数据，包含创建者、空间等基本信息
    *vo.Meta
}
```

### 3. 接口注释 (Interface Comment)

```go
// [接口名] [一句话描述接口的用途]
//
// 该接口定义了[功能领域]的标准行为，实现者包括：
// - [实现类1]: [简述]
// - [实现类2]: [简述]
type InterfaceName interface {
    // [方法名] [方法功能描述]
    //
    // 参数：
    //   - ctx: 上下文，用于传递请求级数据和取消信号
    //   - [param]: [参数说明]
    //
    // 返回值：
    //   - [返回值说明]
    //   - error: [可能的错误情况]
    MethodName(ctx context.Context, param Type) (Result, error)
}
```

### 4. 函数注释 (Function Comment)

```go
// [函数名] [一句话描述函数功能]
//
// [详细描述，如有复杂逻辑]
//
// 参数：
//   - [param1]: [参数说明和有效值范围]
//   - [param2]: [参数说明]
//
// 返回值：
//   - [返回值1]: [说明]
//   - error: [可能的错误类型和含义]
//
// 注意事项：
//   - [并发安全性、副作用等特殊说明]
func FunctionName(param1, param2 Type) (Result, error)
```

---

## TypeScript/React 代码注释规范

### 1. 文件头注释

```typescript
/**
 * @fileoverview [文件功能一句话描述]
 * 
 * [详细描述，说明：]
 * - 主要功能
 * - 导出的关键内容
 * - 使用场景
 * 
 * @module [模块路径]
 * @author [作者（可选）]
 */
```

### 2. 函数注释

```typescript
/**
 * [函数功能一句话描述]
 * 
 * [详细描述（可选，用于复杂逻辑）]
 * 
 * @param paramName - [参数说明]
 * @param options - [配置项说明]
 * @param options.key - [具体配置说明]
 * @returns [返回值说明]
 * @throws {ErrorType} [可能抛出的错误]
 * 
 * @example
 * ```typescript
 * const result = functionName(arg1, { key: value });
 * ```
 */
function functionName(paramName: Type, options: Options): ReturnType
```

### 3. React 组件注释

```typescript
/**
 * [组件功能一句话描述]
 * 
 * [详细描述，说明：]
 * - 组件用途
 * - 主要交互行为
 * - 使用场景
 * 
 * @param props - 组件属性
 * @param props.propName - [属性说明]
 * 
 * @example
 * ```tsx
 * <ComponentName propName="value" onEvent={handler} />
 * ```
 */
const ComponentName: React.FC<Props> = (props) => {
  // 实现
};
```

### 4. React Hook 注释

```typescript
/**
 * [Hook 功能一句话描述]
 * 
 * [详细描述，说明：]
 * - Hook 解决的问题
 * - 使用场景
 * - 依赖项说明
 * 
 * @param config - [配置参数说明]
 * @returns [返回值说明，包括返回对象的各个字段]
 * 
 * @example
 * ```tsx
 * const { data, loading, error } = useHookName({ key: value });
 * ```
 */
function useHookName(config: Config): HookResult
```

### 5. 接口/类型注释

```typescript
/**
 * [类型用途一句话描述]
 * 
 * [详细说明使用场景]
 */
interface InterfaceName {
  /** [字段说明] */
  fieldName: Type;
  
  /**
   * [方法说明]
   * @param arg - [参数说明]
   */
  methodName(arg: Type): ReturnType;
}
```

### 6. 行内注释

```typescript
// 处理边界情况：当用户未选择任何选项时，默认使用第一个
if (!selectedItem) {
  selectedItem = items[0];
}

// TODO: 需要优化性能，当前实现在大数据量下可能较慢
// FIXME: 修复在 Safari 浏览器下的显示问题
// NOTE: 这里的逻辑依赖于 API 返回的特定格式
```

---

## 配置文件注释规范

### YAML 文件

```yaml
# [文件用途说明]

# [配置块说明]
configKey:
  # [字段说明]
  field: value
```

### Makefile

```makefile
# [目标说明]
# 
# 用法: make [target]
# 依赖: [前置条件]
target-name:
	@echo "执行命令"
```

### Docker Compose

```yaml
services:
  # [服务名] - [服务用途]
  # 端口: [端口说明]
  # 依赖: [依赖的其他服务]
  service-name:
    image: image:tag
```

---

## 注释质量标准

### 必须遵守

1. **准确性**: 注释必须准确反映代码功能，禁止误导性描述
2. **简洁性**: 避免冗余，不解释显而易见的代码
3. **一致性**: 遵循本文档定义的格式规范
4. **中文为主**: 所有新增注释使用中文，保留原有英文注释

### 应该包含

1. 所有公共/导出的函数、类、组件
2. 复杂的业务逻辑（超过 20 行或包含条件分支）
3. 非显而易见的算法或数据转换
4. 与外部系统的集成点

### 不应该包含

1. 显而易见的 getter/setter
2. 简单的一行函数
3. 自动生成的代码
4. 测试文件中的辅助函数

---

## 验收标准

注释通过验收需满足：

- [ ] 格式符合本文档规范
- [ ] 内容准确描述代码功能
- [ ] 没有语法错误或错别字
- [ ] 对新开发者理解代码有帮助

