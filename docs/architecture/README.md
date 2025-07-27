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

- [技术框架文档](../frameworks/)
- [业务需求文档](../business/)
- [开发指南](../guides/) 