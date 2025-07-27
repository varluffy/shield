# 架构文档

本目录包含系统的所有架构设计文档，分为核心架构和业务架构两个部分。

## 目录结构

### 核心架构 ([core/](./core/))
技术架构和框架相关的文档。

- [go-microservices-core.md](./core/go-microservices-core.md) - Go微服务核心架构设计
- [wire-architecture.md](./core/wire-architecture.md) - Wire依赖注入架构设计

### 业务架构 ([business/](./business/))
业务领域的架构设计文档。

- [auth-system-design.md](./business/auth-system-design.md) - 认证系统架构设计
- [captcha-system-design.md](./business/captcha-system-design.md) - 验证码系统架构设计
- [multi-tenant-design.md](./business/multi-tenant-design.md) - 多租户架构设计
- [permission-system.md](./business/permission-system.md) - 权限系统架构设计
- [schema-design.md](./business/schema-design.md) - 数据库架构设计

## 架构原则

1. **清洁架构**: 遵循依赖倒置原则，保持层次分离
2. **微服务设计**: 服务边界清晰，松耦合高内聚
3. **可观测性**: 完整的日志、监控和追踪体系
4. **安全性**: 多层次安全防护机制

## 相关文档

### 核心文档
- **[系统设计总览](./system-design.md)** - 完整的系统设计文档
- **[开发指南](../development/)** - 开发规范和最佳实践

### 详细文档
- [技术框架文档](../frameworks/) - 框架使用指南
- [业务需求文档](../business/) - 业务需求规格
- [API 文档](../api/) - 接口规范和使用说明

### 入门指南
- **[快速开始](../development/getting-started.md)** - 新手必读
- **[架构规范](../development/architecture.md)** - 架构开发约束 