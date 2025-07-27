# UltraFit 项目文档

欢迎来到 UltraFit 项目文档中心！这里包含了项目的所有技术文档、架构设计、开发指南和API文档。

## 📚 文档导航

### 🚀 快速开始
- [**快速开始指南**](./guides/getting-started.md) - 新手入门必读
- [开发环境搭建](./guides/development-setup.md) - 开发环境配置
- [开发规则约束](./guides/development-rules.md) - 代码规范和最佳实践

### 🏗️ 系统架构
- [**架构文档总览**](./architecture/) - 完整的系统架构设计
  - [核心架构](./architecture/core/) - Go微服务核心架构和Wire依赖注入
  - [业务架构](./architecture/business/) - 认证、权限、多租户等业务架构

### 🌐 API 文档  
- [**API文档中心**](./api/) - 完整的API文档和规范
  - [Swagger文档](./api/swagger.yaml) - OpenAPI规范
  - [认证API](./api/auth-api.md) - 用户认证相关接口
  - [权限API](./api/permission-api.md) - 权限管理接口

### ⚙️ 技术框架
- [**框架使用指南**](./frameworks/) - 技术栈和框架文档
  - [Go Web框架](./frameworks/go-web-framework.md) - Gin框架使用指南
  - [数据库指南](./frameworks/go-database-guide.md) - GORM和数据库最佳实践
  - [可观测性](./frameworks/go-observability.md) - 日志、监控和追踪

### 📋 业务需求
- [**业务文档**](./business/) - 项目需求和设计文档
  - [需求规格说明](./business/requirements.md) - 详细的功能需求
  - [系统设计总结](./business/system-design-summary.md) - 高层次设计概览

### 📖 使用指南
- [**开发指南**](./guides/) - 开发相关的实用指南
  - [权限快速上手](./guides/permission-quick-start.md) - 权限系统使用指南
  - [系统操作指南](./guides/system-guides.md) - 系统操作和维护

### 📊 优化报告
- [**项目报告**](./reports/) - 优化报告和改进总结
  - [项目优化报告](./reports/project-optimization.md) - 整体优化分析
  - [权限系统优化](./reports/permission-system-optimization.md) - 权限模块优化
  - [测试系统重构](./reports/test-system-refactoring.md) - 测试架构改进

## 🎯 推荐阅读路径

### 新手开发者
1. [快速开始指南](./guides/getting-started.md) ← **从这里开始**
2. [开发环境搭建](./guides/development-setup.md)
3. [开发规则约束](./guides/development-rules.md)
4. [权限快速上手](./guides/permission-quick-start.md)

### 架构师
1. [架构文档总览](./architecture/)
2. [核心架构设计](./architecture/core/)
3. [业务架构设计](./architecture/business/)
4. [系统设计总结](./business/system-design-summary.md)

### 前端开发者
1. [API文档中心](./api/)
2. [认证API文档](./api/auth-api.md)
3. [权限API文档](./api/permission-api.md)
4. [Swagger在线文档](./api/swagger.yaml)

### 运维工程师
1. [可观测性指南](./frameworks/go-observability.md)
2. [数据库运维](./frameworks/go-database-guide.md)
3. [系统操作指南](./guides/system-guides.md)

## 🔧 技术栈

- **Web框架**: Gin
- **依赖注入**: Wire  
- **ORM**: GORM
- **数据库**: MySQL
- **日志**: Zap + OpenTelemetry
- **追踪**: Jaeger
- **配置**: Viper
- **测试**: Testify + Gomock

## 📝 文档维护

本文档使用 Markdown 格式编写，遵循以下原则：
- **及时更新**: 代码变更时同步更新文档
- **清晰简洁**: 使用简洁明了的语言
- **示例丰富**: 提供完整的代码示例
- **结构化**: 保持良好的文档层次结构

## 🤝 贡献指南

如需更新文档，请遵循：
1. 保持与现有文档风格一致
2. 添加必要的示例和说明
3. 更新相关的索引和链接
4. 通过代码审查流程

---

**💡 提示**: 如果您是新加入的团队成员，建议从 [快速开始指南](./guides/getting-started.md) 开始阅读！ 