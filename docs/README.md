# Shield 项目文档中心

欢迎来到 Shield 项目文档中心！这里提供完整的项目文档，包括开发指南、架构设计、API 规范和最佳实践。

## 🚀 快速开始

**新加入的开发者**，请从这里开始：

1. **[开发入门指南](./development/getting-started.md)** ← **从这里开始！**
2. **[架构设计规范](./development/architecture.md)** - 理解项目架构
3. **[API 开发指南](./development/api-guide.md)** - 学习 API 开发标准
4. **[测试指南](./development/testing-guide.md)** - 掌握测试最佳实践

## 📚 核心文档

### 💻 开发指南
**统一的开发文档，消除重复和混乱**

- **[📖 快速开始](./development/getting-started.md)** - 15分钟快速上手项目
- **[🏗️ 架构设计](./development/architecture.md)** - 清洁架构规范和开发约束
- **[🔧 API 开发](./development/api-guide.md)** - RESTful API 设计和实现标准
- **[🧪 测试指南](./development/testing-guide.md)** - 测试策略、工具和最佳实践

### 📖 API 文档
- **[API 规范中心](./api/)** - 完整的 API 文档和规范
  - [认证 API](./api/auth-api.md) - 用户认证和授权接口
  - [权限 API](./api/permission-api.md) - 权限管理系统接口
  - [Swagger 文档](./api/swagger.yaml) - OpenAPI 3.0 规范

### 🏛️ 架构文档
- **[系统架构](./architecture/)** - 深度架构设计文档
  - [核心架构](./architecture/core/) - Go 微服务核心架构
  - [业务架构](./architecture/business/) - 业务域架构设计

## 🎯 按角色导航

### 👨‍💻 后端开发者
1. [开发入门指南](./development/getting-started.md) - 环境搭建和快速开始
2. [架构设计规范](./development/architecture.md) - 分层架构和开发约束
3. [API 开发指南](./development/api-guide.md) - API 设计和实现
4. [测试指南](./development/testing-guide.md) - 测试策略和工具

### 🏗️ 架构师
1. [架构文档总览](./architecture/) - 完整架构设计
2. [核心架构](./architecture/core/) - 技术架构和模式
3. [业务架构](./architecture/business/) - 业务域设计
4. [系统设计总结](./business/system-design-summary.md) - 高层次设计

### 🌐 前端开发者
1. [API 规范中心](./api/) - 接口文档和规范
2. [认证 API](./api/auth-api.md) - 登录和权限接口
3. [权限 API](./api/permission-api.md) - 权限管理接口
4. [Swagger 文档](./api/swagger.yaml) - 交互式 API 文档

### 🔧 运维工程师
1. [快速开始指南](./development/getting-started.md) - 部署和配置
2. [可观测性指南](./frameworks/go-observability.md) - 监控和日志
3. [数据库指南](./frameworks/go-database-guide.md) - 数据库运维
4. [系统操作指南](./guides/system-guides.md) - 系统维护

## 🔗 其他资源

### 📋 业务文档
- [业务需求文档](./business/) - 功能需求和设计规格
- [系统设计总结](./business/system-design-summary.md) - 业务系统设计

### ⚙️ 技术框架
- [技术框架指南](./frameworks/) - 深度技术文档
- [Go Web 框架](./frameworks/go-web-framework.md) - Gin 框架使用
- [数据库指南](./frameworks/go-database-guide.md) - GORM 最佳实践
- [可观测性](./frameworks/go-observability.md) - 监控和追踪

### 📊 项目报告
- [优化报告](./reports/) - 性能优化和重构报告
- [权限系统优化](./reports/permission-system-optimization.md)
- [测试系统重构](./reports/test-system-refactoring.md)

## 🔧 技术栈概览

- **Web 框架**: Gin - 高性能 HTTP 路由框架
- **依赖注入**: Wire - Google 官方编译时 DI 工具
- **ORM**: GORM - Go 语言功能丰富的 ORM 库
- **数据库**: MySQL - 主数据库存储
- **缓存**: Redis - 分布式缓存（验证码、会话等）
- **日志**: Zap + OpenTelemetry - 结构化日志和分布式追踪
- **配置**: Viper - 多格式配置管理
- **测试**: Testify + Gomock - 单元测试和 Mock 框架

## 📝 文档组织原则

本文档遵循以下组织原则：

### ✨ 新的文档结构 (2024)
- **单一入口**: 统一的文档门户和导航
- **角色导向**: 按开发者角色组织学习路径
- **消除重复**: 合并和精简重复内容
- **AI 友好**: 优化 AI 辅助开发体验

### 📁 目录说明
```
docs/
├── development/        # 🆕 统一开发指南
│   ├── getting-started.md    # 快速开始（整合版）
│   ├── architecture.md       # 架构规范（整合版）
│   ├── api-guide.md         # API 开发（整合版）
│   └── testing-guide.md     # 测试指南（整合版）
├── api/               # API 文档和规范
├── architecture/      # 深度架构文档
├── business/          # 业务需求文档
├── frameworks/        # 技术框架指南
└── reports/          # 项目报告和优化
```

## 🤝 文档维护指南

### 更新原则
1. **保持同步**: 代码变更时及时更新文档
2. **单一来源**: 避免信息重复，维护单一权威来源
3. **实用导向**: 专注于实际开发需要的信息
4. **示例丰富**: 提供完整可运行的代码示例

### 贡献流程
1. 遵循现有文档风格和结构
2. 更新相关的交叉引用和链接
3. 确保新内容与整体文档架构一致
4. 通过代码审查和文档审查流程

---

**💡 开始您的开发之旅**: 新团队成员请从 **[开发入门指南](./development/getting-started.md)** 开始！

**🤖 AI 开发者**: 参考 [CLAUDE.md](../CLAUDE.md) 获取 AI 辅助开发的专门指导。 