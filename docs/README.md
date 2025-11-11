# Coze Studio 技术文档

本目录包含 Coze Studio 项目的技术文档，按模块组织。

## 📂 文档目录结构

### 🔧 后端开发文档 (`backend/`)
后端架构、开发指南和最佳实践文档。

- **[BACKEND_README.md](backend/BACKEND_README.md)** - 后端项目概览
- **[BACKEND_QUICKSTART.md](backend/BACKEND_QUICKSTART.md)** - 后端快速入门指南
- **[BACKEND_LEARNING_GUIDE.md](backend/BACKEND_LEARNING_GUIDE.md)** - 后端学习路径
- **[BACKEND_PRACTICE.md](backend/BACKEND_PRACTICE.md)** - 后端开发最佳实践
- **[BACKEND_GORM_GEN_GUIDE.md](backend/BACKEND_GORM_GEN_GUIDE.md)** - GORM Gen 使用指南
- **[BACKEND_GORM_GEN_IMPLEMENTATION.md](backend/BACKEND_GORM_GEN_IMPLEMENTATION.md)** - GORM Gen 实现细节
- **[BACKEND_IDL_GENERATION.md](backend/BACKEND_IDL_GENERATION.md)** - IDL 接口定义和代码生成
- **[BACKEND_ERRATA.md](backend/BACKEND_ERRATA.md)** - 后端文档勘误
- **[qa.md](backend/qa.md)** - 常见问题解答
- **[DOCUMENTATION_FIX_SUMMARY.md](backend/DOCUMENTATION_FIX_SUMMARY.md)** - 文档修复汇总
- **[DOCUMENTATION_FIXES.md](backend/DOCUMENTATION_FIXES.md)** - 文档修复详情

### 🔄 工作流API文档 (`workflow-api/`)
工作流相关API的调用链路和实现分析。

- **[WORKFLOW_CREATE_API_WORKFLOW.md](workflow-api/WORKFLOW_CREATE_API_WORKFLOW.md)** - 创建工作流API执行路径
- **[WORKFLOW_TEST_RUN_CALL_CHAIN.md](workflow-api/WORKFLOW_TEST_RUN_CALL_CHAIN.md)** - 测试运行API调用链路

## 📚 文档分类

### 入门指南
适合新手快速了解项目架构和开发流程：
- [后端快速入门](backend/BACKEND_QUICKSTART.md)
- [后端学习路径](backend/BACKEND_LEARNING_GUIDE.md)

### 开发指南
日常开发中的技术细节和工具使用：
- [GORM Gen 使用指南](backend/BACKEND_GORM_GEN_GUIDE.md)
- [IDL 接口定义](backend/BACKEND_IDL_GENERATION.md)
- [后端开发最佳实践](backend/BACKEND_PRACTICE.md)

### API分析
核心API的调用链路和实现分析：
- [创建工作流API](workflow-api/WORKFLOW_CREATE_API_WORKFLOW.md)
- [测试运行API](workflow-api/WORKFLOW_TEST_RUN_CALL_CHAIN.md)

### 参考文档
技术细节和问题排查：
- [常见问题解答](backend/qa.md)
- [文档勘误](backend/BACKEND_ERRATA.md)

## 🔍 文档索引

### 按主题查找

#### 工作流相关
- 创建工作流流程 → [WORKFLOW_CREATE_API_WORKFLOW.md](workflow-api/WORKFLOW_CREATE_API_WORKFLOW.md)
- 测试运行流程 → [WORKFLOW_TEST_RUN_CALL_CHAIN.md](workflow-api/WORKFLOW_TEST_RUN_CALL_CHAIN.md)

#### 数据库操作
- GORM使用 → [BACKEND_GORM_GEN_GUIDE.md](backend/BACKEND_GORM_GEN_GUIDE.md)
- 代码生成 → [BACKEND_GORM_GEN_IMPLEMENTATION.md](backend/BACKEND_GORM_GEN_IMPLEMENTATION.md)

#### API开发
- IDL定义 → [BACKEND_IDL_GENERATION.md](backend/BACKEND_IDL_GENERATION.md)
- API调用链 → [workflow-api/](workflow-api/)

#### 项目架构
- 后端架构 → [BACKEND_README.md](backend/BACKEND_README.md)
- 最佳实践 → [BACKEND_PRACTICE.md](backend/BACKEND_PRACTICE.md)

## 📝 文档维护

- **创建时间**: 2024年10月-11月
- **维护者**: Coze Studio开发团队
- **更新频率**: 随项目演进持续更新

## 🤝 贡献指南

如需更新文档：
1. 保持文档结构清晰
2. 使用中文编写
3. 包含代码示例和图表
4. 及时更新目录索引

