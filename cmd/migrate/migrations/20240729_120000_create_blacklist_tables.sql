-- Description: Create blacklist related tables and permissions
-- Created: 20240729_120000

-- +migrate Up
-- 手机号黑名单表
CREATE TABLE IF NOT EXISTS `phone_blacklists` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `uuid` char(36) NOT NULL,
    `tenant_id` bigint unsigned NOT NULL COMMENT '租户ID',
    `phone_md5` char(32) NOT NULL COMMENT '手机号MD5',
    `source` varchar(50) NOT NULL COMMENT '来源：manual, import, api',
    `reason` varchar(200) DEFAULT NULL COMMENT '加入黑名单原因',
    `operator_id` bigint unsigned DEFAULT NULL COMMENT '操作人ID',
    `is_active` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否有效',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at` datetime(3) DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_tenant_phone_md5` (`tenant_id`,`phone_md5`),
    KEY `idx_tenant_id` (`tenant_id`),
    KEY `idx_phone_md5` (`phone_md5`),
    KEY `idx_operator_id` (`operator_id`),
    KEY `idx_created_at` (`created_at`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='手机号黑名单表';

-- 黑名单API密钥表
CREATE TABLE IF NOT EXISTS `blacklist_api_credentials` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `uuid` char(36) NOT NULL,
    `tenant_id` bigint unsigned NOT NULL COMMENT '租户ID',
    `api_key` varchar(64) NOT NULL COMMENT 'API Key',
    `api_secret` varchar(128) NOT NULL COMMENT 'API Secret',
    `name` varchar(100) NOT NULL COMMENT '密钥名称',
    `description` text COMMENT '描述',
    `ip_whitelist` TEXT DEFAULT NULL COMMENT 'IP白名单，逗号分隔，支持CIDR',
    `rate_limit` int NOT NULL DEFAULT '1000' COMMENT '每秒请求限制',
    `status` varchar(20) NOT NULL DEFAULT 'active' COMMENT '状态：active, inactive, suspended',
    `last_used_at` datetime(3) DEFAULT NULL COMMENT '最后使用时间',
    `expires_at` datetime(3) DEFAULT NULL COMMENT '过期时间',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at` datetime(3) DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_api_key` (`api_key`),
    KEY `idx_uuid` (`uuid`),
    KEY `idx_tenant_id` (`tenant_id`),
    KEY `idx_status` (`status`),
    KEY `idx_expires_at` (`expires_at`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='黑名单API密钥表';

-- 黑名单查询日志表（可选，用于详细统计分析）
CREATE TABLE IF NOT EXISTS `blacklist_query_logs` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `tenant_id` bigint unsigned NOT NULL COMMENT '租户ID',
    `api_key` varchar(64) NOT NULL COMMENT 'API Key',
    `phone_md5` char(32) NOT NULL COMMENT '查询的手机号MD5',
    `is_hit` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否命中黑名单',
    `response_time` int NOT NULL COMMENT '响应时间(毫秒)',
    `client_ip` varchar(45) DEFAULT NULL COMMENT '客户端IP',
    `user_agent` text COMMENT '用户代理',
    `request_id` varchar(64) DEFAULT NULL COMMENT '请求ID',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    KEY `idx_tenant_id` (`tenant_id`),
    KEY `idx_api_key` (`api_key`),
    KEY `idx_created_at` (`created_at`),
    KEY `idx_tenant_created` (`tenant_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='黑名单查询日志表';

-- 插入默认的黑名单权限
INSERT INTO `permissions` (`uuid`, `code`, `name`, `description`, `type`, `scope`, `is_builtin`, `module`) VALUES
-- 黑名单查询权限（面向外部API）
(UUID(), 'blacklist_check_api', '黑名单查询', '查询手机号是否在黑名单中', 'api', 'api', 1, 'blacklist'),
-- 黑名单管理权限（面向管理后台）
(UUID(), 'blacklist_create_api', '创建黑名单', '创建黑名单记录', 'api', 'tenant', 1, 'blacklist'),
(UUID(), 'blacklist_import_api', '批量导入黑名单', '批量导入黑名单数据', 'api', 'tenant', 1, 'blacklist'),
(UUID(), 'blacklist_list_api', '查看黑名单列表', '查看黑名单列表', 'api', 'tenant', 1, 'blacklist'),
(UUID(), 'blacklist_delete_api', '删除黑名单', '删除黑名单记录', 'api', 'tenant', 1, 'blacklist'),
(UUID(), 'blacklist_stats_api', '查看黑名单统计', '查看黑名单查询统计', 'api', 'tenant', 1, 'blacklist');

-- 为管理员角色分配黑名单管理权限
INSERT INTO `role_permissions` (`role_id`, `permission_id`)
SELECT r.id, p.id
FROM `roles` r
CROSS JOIN `permissions` p
WHERE r.code = 'admin' 
  AND p.code IN ('blacklist_create_api', 'blacklist_import_api', 'blacklist_list_api', 'blacklist_delete_api', 'blacklist_stats_api')
  AND NOT EXISTS (
    SELECT 1 FROM `role_permissions` rp 
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- +migrate Down
-- 删除角色权限关联
DELETE rp FROM `role_permissions` rp
INNER JOIN `permissions` p ON rp.permission_id = p.id
WHERE p.module = 'blacklist';

-- 删除黑名单相关权限
DELETE FROM `permissions` WHERE `module` = 'blacklist';

-- 删除黑名单相关表
DROP TABLE IF EXISTS `blacklist_query_logs`;
DROP TABLE IF EXISTS `blacklist_api_credentials`;
DROP TABLE IF EXISTS `phone_blacklists`;