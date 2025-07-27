# UltraFit 权限系统设计总结

## 📋 项目概述

UltraFit 是一个基于 Go + Gin 的多租户后台管理系统，采用简洁高效的权限管理架构。本文档总结了经过充分讨论后确定的最终设计方案。

## 🎯 设计共识

### 1. 核心原则
- **简单高效**：避免过度设计，专注核心功能
- **多租户隔离**：完全的数据隔离和权限隔离
- **灵活切换**：系统管理员可以无缝切换管理不同租户
- **安全可靠**：完善的认证和权限验证机制

### 2. 技术选型
- **后端框架**：Gin + Wire + GORM
- **数据库**：MySQL（不使用物理外键）
- **认证**：JWT + Redis
- **验证码**：base64Captcha + Redis
- **权限控制**：基于RBAC的自研权限系统（不使用Casbin）
- **多租户**：行级隔离（tenant_id）

## 🏗️ 架构设计

### 1. 多租户架构
```
系统级 (tenant_id = 0)
├── 系统管理员 - 管理所有租户
└── 系统级权限 - 租户管理、系统监控

租户级 (tenant_id > 0)
├── 租户管理员 - 管理本租户资源
└── 租户级权限 - 用户管理、角色管理
```

### 2. 租户切换机制
```
用户登录 → 获取可访问租户列表 → 选择租户 → 通过Header传输租户UUID
```

**优势**：
- 切换租户不需要重新生成token
- 切换速度快，用户体验好
- 实现简单，维护成本低

### 3. 角色权限设计
- **系统管理员**：`tenant_id = 0`，管理整个系统平台
- **租户管理员**：`tenant_id > 0`，管理具体租户的所有资源

**权限分类**：
- 菜单权限：控制菜单显示和访问
- 按钮权限：控制页面按钮的显示  
- API权限：控制后端接口的访问

**权限继承**：
- 通过 `parent_code` 建立父子关系
- 前端处理继承逻辑（选中父权限自动选中子权限）
- 后端直接存储所有选中的权限ID
- 所有权限验证逻辑相同

### 4. 验证码系统设计
- **验证码库**：[base64Captcha](https://github.com/mojocn/base64Captcha)
- **存储方案**：Redis（分布式支持）
- **验证码类型**：数字、字符串、数学运算
- **过期时间**：5分钟
- **应用场景**：登录安全保护

## 🗄️ 数据库设计

### 1. 设计约束
- **不使用物理外键**：通过业务逻辑维护关系
- **枚举使用字符串**：不使用MySQL的enum类型
- **支持软删除**：所有表都有 `deleted_at` 字段
- **审计字段**：记录创建时间、更新时间、创建人

### 2. 核心数据表
- **tenants**：租户表（增加tenant_uuid字段）
- **users**：用户表（增加last_login_at字段）
- **roles**：角色表（简化设计，只有两个核心角色）
- **permissions**：权限表（支持树形结构）
- **user_roles**：用户角色关联表
- **role_permissions**：角色权限关联表
- **login_attempts**：登录安全表
- **refresh_tokens**：刷新令牌表

### 3. 关键设计点

#### 租户表设计
```sql
CREATE TABLE `tenants` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_uuid` char(36) NOT NULL,  -- 新增：用于对外标识
  `tenant_code` varchar(50) NOT NULL,
  `tenant_name` varchar(100) NOT NULL,
  `status` varchar(20) DEFAULT 'active', -- 枚举使用字符串
  `max_users` int DEFAULT 10,
  -- 其他字段...
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_uuid` (`tenant_uuid`, `deleted_at`)
);
```

#### 用户表设计
```sql
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned NOT NULL,
  `username` varchar(100) NOT NULL,
  `email` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `status` varchar(20) DEFAULT 'active', -- 'active', 'inactive', 'locked'
  `last_login_at` timestamp NULL DEFAULT NULL, -- 新增：记录最后登录时间
  -- 其他字段...
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_email` (`tenant_id`, `email`, `deleted_at`)
);
```

#### 权限表设计
```sql
CREATE TABLE `permissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint unsigned NOT NULL,
  `permission_code` varchar(100) NOT NULL,
  `permission_name` varchar(100) NOT NULL,
  `permission_type` varchar(20) NOT NULL, -- 'menu', 'button', 'api'
  `parent_code` varchar(100) DEFAULT NULL, -- 支持树形结构
  `resource_path` varchar(200) DEFAULT NULL,
  `method` varchar(10) DEFAULT NULL,
  -- 其他字段...
  PRIMARY KEY (`id`),
  KEY `idx_parent_code` (`parent_code`)
);
```

## 🔐 认证系统设计

### 1. JWT + Redis 架构
- **访问令牌**：2小时有效期，存储用户基本信息
- **刷新令牌**：30天有效期，存储在数据库
- **Redis缓存**：token信息缓存，支持主动失效

### 2. 登录安全控制
- **图形验证码**：使用base64Captcha生成图形验证码
- **失败限制**：5次失败后锁定账号
- **登录记录**：记录最后登录时间和IP
- **状态管理**：active、inactive、locked 三种状态

### 3. 权限验证流程
```
请求到达 → 租户上下文验证 → 用户身份验证 → API权限匹配 → 权限检查 → 放行/拒绝
```

### 4. 验证码验证流程
```
登录请求 → 验证码校验 → 用户认证 → 权限验证 → 登录成功
```

## 🌐 API 设计

### 1. 接口规范
- **Base URL**: `/api/v1`
- **认证方式**: `Authorization: Bearer <access_token>`
- **租户标识**: `X-Tenant-ID: <tenant_uuid>`

### 2. 核心接口分类
- **验证码接口**：生成验证码
- **认证接口**：登录、刷新、登出、获取用户信息
- **系统管理接口**：租户管理（系统管理员专用）
- **租户管理接口**：用户管理、角色管理、权限管理
- **安全日志接口**：登录历史、权限变更日志

### 3. 自动权限验证
中间件自动识别API路径和方法，匹配对应权限进行验证，无需手动配置每个路由的权限。

### 4. 验证码集成
- **获取验证码**：`GET /captcha/generate`
- **登录验证**：登录接口需要验证码ID和答案
- **存储管理**：Redis存储，5分钟过期

## 📱 前端设计

### 1. 租户切换
- 顶部显示租户下拉选择器
- 系统管理员可以看到所有租户
- 选择租户后通过Header传输租户UUID

### 2. 权限分配界面
- 树形结构展示权限
- 权限类型标签区分（菜单/按钮/API）
- 选中父权限自动选中子权限
- 实时权限预览

### 3. 按钮权限控制
```javascript
<el-button v-if="hasPermission('user_create_btn')" @click="createUser">
  创建用户
</el-button>
```

### 4. 验证码组件
```javascript
// 验证码组件集成
const CaptchaComponent = {
  methods: {
    async refreshCaptcha() {
      const response = await axios.get('/api/v1/captcha/generate')
      this.captchaId = response.data.data.captcha_id
      this.captchaImage = response.data.data.captcha_image
    }
  }
}
```

## 🚀 系统初始化

### 1. 系统启动初始化
1. 创建系统管理员角色（`tenant_id = 0`）
2. 创建系统管理员用户账号
3. 初始化系统级权限（租户管理、系统监控等）

### 2. 租户创建流程
1. 系统管理员创建租户（生成UUID）
2. 自动创建租户管理员角色
3. 初始化租户基础权限（用户管理、角色管理等）
4. 创建租户管理员账号

## 📊 性能优化

### 1. 权限缓存策略
- 用户权限列表缓存到Redis（1小时）
- 登录时预加载权限到缓存
- 权限变更时清除相关缓存

### 2. 验证码缓存策略
- 验证码答案存储到Redis（5分钟）
- 验证后立即删除
- 支持分布式验证

### 3. 数据库优化
- 关键字段建立索引
- 权限查询优化
- 避免N+1查询问题

## 🔒 安全策略

### 1. 数据安全
- 严格的租户数据隔离
- 跨租户访问权限验证
- 敏感信息加密存储

### 2. 认证安全
- JWT令牌验证
- 图形验证码保护
- 登录失败次数限制
- 账号锁定机制

### 3. 权限安全
- API访问权限控制
- 权限变更审计日志
- 最小权限原则

### 4. 验证码安全
- 有效期控制（5分钟）
- 一次性使用
- 防暴力破解
- 适当的图片复杂度

## 🛠️ 开发规范

### 1. 代码规范
- 统一的错误处理
- 清晰的接口定义
- 完善的日志记录

### 2. 测试策略
- 单元测试覆盖核心逻辑
- 集成测试验证权限流程
- 性能测试验证系统承载能力

## 📋 文档体系

### 1. 架构设计文档
- [权限系统架构设计](./architecture/permission-system-design.md)
- [验证码系统架构设计](./architecture/captcha-system-design.md)
- [多租户架构设计](./architecture/multi-tenant-design.md)
- [认证系统设计](./architecture/auth-system-design.md)

### 2. API接口文档
- [权限系统API接口](./api/permission-api.md)
- [认证系统API接口](./api/auth-api.md)

### 3. 数据库设计文档
- [数据库模型设计](./database/schema-design.md)

## 🎉 设计优势

### 1. 简洁高效
- 避免过度设计，专注核心功能
- 两个角色设计满足当前需求
- 权限继承前端处理，后端存储简单

### 2. 安全可靠
- 完善的认证机制（JWT + Redis）
- 图形验证码防恶意登录
- 严格的权限验证
- 详细的审计日志

### 3. 灵活扩展
- 预留关键扩展点
- 支持自定义权限
- 支持权限配置化
- 验证码多类型支持

### 4. 易于维护
- 清晰的架构设计
- 统一的代码规范
- 完善的文档体系
- 分布式友好设计

## 🔄 下一步计划

1. **确认设计方案**：与团队确认最终设计方案
2. **配置依赖包**：添加base64Captcha依赖
3. **开始编码实现**：按照设计文档进行开发
4. **单元测试**：编写核心功能的单元测试
5. **集成测试**：验证整个权限流程
6. **性能测试**：验证系统性能指标

---

## 📝 变更记录

| 版本 | 日期 | 变更内容 | 负责人 |
|------|------|----------|--------|
| 1.0 | 2024-01-01 | 初始设计方案 | 开发团队 |
| 1.1 | 2024-01-01 | 增加验证码系统设计 | 开发团队 |

---

**注意**：本文档是经过充分讨论和评估后确定的最终设计方案，已充分考虑了扩展性、性能和安全性等因素。建议在开发过程中严格按照本设计方案执行。 