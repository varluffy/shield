# 手机号MD5黑名单查询系统

## 📖 系统概述

本系统是基于现有shield项目架构实现的高性能手机号MD5黑名单查询系统，支持10000-20000并发查询，采用Redis缓存 + MySQL持久化的存储方案，实现HMAC签名鉴权。

## 🎯 核心特性

### 高性能架构
- **并发能力**: 支持10000-20000 QPS
- **响应时间**: P99 < 10ms
- **缓存策略**: Redis SET存储，O(1)查询复杂度
- **连接池优化**: Redis连接池100个连接

### 双鉴权体系
- **查询接口**: HMAC签名鉴权（~20μs延迟）
- **管理接口**: JWT Token鉴权（复用现有系统）

### 智能日志
- **采样率**: 查询成功1%采样，错误100%记录
- **异步处理**: 不阻塞主请求流程
- **慢查询告警**: >50ms请求100%记录

### 实时监控
- **分钟级统计**: QPS、命中率、平均延迟
- **Redis临时存储**: 48小时TTL
- **异步持久化**: 定时写入MySQL

## 🏗️ 系统架构

### 数据模型
```
phone_blacklists              # 黑名单主表
├── id (PK)
├── tenant_id (租户隔离)
├── phone_md5 (32位MD5)
├── source (来源：manual/import/api)
├── reason (原因)
├── operator_id (操作人)
└── is_active (是否有效)

blacklist_api_credentials     # API密钥表
├── id (PK)
├── tenant_id
├── api_key (API密钥)
├── api_secret (密钥)
├── rate_limit (速率限制/秒)
├── status (状态)
└── expires_at (过期时间)

blacklist_query_logs         # 查询日志表（可选）
├── tenant_id
├── api_key
├── phone_md5
├── is_hit (是否命中)
├── response_time (响应时间ms)
└── client_ip
```

### Redis存储结构
```
blacklist:tenant:{tenant_id}     # SET存储MD5列表
stats:query:{api_key}:{hour}     # HASH存储小时统计
rate_limit:{api_key}             # ZSET滑动窗口计数
nonce:{api_key}:{nonce}          # STRING防重放Nonce
```

## 🚀 API接口

### 查询接口 (HMAC鉴权)

**POST** `/api/v1/blacklist/check`

**请求头:**
```http
Content-Type: application/json
X-API-Key: {api_key}
X-Timestamp: {unix_timestamp}
X-Nonce: {random_string}
X-Signature: {hmac_sha256_signature}
```

**签名算法:**
```
message = api_key + timestamp + nonce + request_body
signature = HMAC-SHA256(message, api_secret)
```

**请求体:**
```json
{
  "phone_md5": "5d41402abc4b2a76b9719d911017c592"
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "is_blacklist": true,
    "phone_md5": "5d41402abc4b2a76b9719d911017c592"
  },
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 管理接口 (JWT鉴权)

**创建黑名单**
```http
POST /api/v1/admin/blacklist
Authorization: Bearer {jwt_token}

{
  "phone_md5": "5d41402abc4b2a76b9719d911017c592",
  "source": "manual",
  "reason": "用户投诉"
}
```

**批量导入**
```http
POST /api/v1/admin/blacklist/import
Authorization: Bearer {jwt_token}

{
  "phone_md5_list": ["md5_1", "md5_2", ...],
  "source": "import",
  "reason": "批量导入"
}
```

**查询列表**
```http
GET /api/v1/admin/blacklist?page=1&page_size=20
Authorization: Bearer {jwt_token}
```

**查询统计**
```http
GET /api/v1/admin/blacklist/stats?hours=24
Authorization: Bearer {jwt_token}
```

## 🔧 部署配置

### Redis配置优化
```yaml
redis:
  addrs: ["localhost:6379"]
  password: "your_password"
  db: 0
  pool_size: 100          # 高并发连接池
  min_idle_conns: 10
  max_idle_conns: 50
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s
  idle_timeout: 300s
  key_prefix: "shield:"
  enable_tracing: true
```

### 日志配置
```yaml
log:
  level: "info"
  format: "json"
  output: "stdout"
```

## 🧪 测试使用

### 1. 启动服务
```bash
# 启动开发服务器
make run

# 或者编译后启动
make build
./bin/server
```

### 2. 创建API密钥
使用管理接口创建API密钥（需要JWT Token）：
```bash
curl -X POST "http://localhost:8080/api/v1/admin/api-credentials" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试密钥",
    "rate_limit": 1000,
    "description": "用于测试的API密钥"
  }'
```

### 3. 测试黑名单查询
```bash
# 使用提供的测试脚本
./scripts/test_blacklist.sh

# 或手动测试
API_KEY="your_api_key"
API_SECRET="your_api_secret"
TIMESTAMP=$(date +%s)
NONCE=$(openssl rand -hex 16)
BODY='{"phone_md5":"5d41402abc4b2a76b9719d911017c592"}'

# 生成签名
MESSAGE="${API_KEY}${TIMESTAMP}${NONCE}${BODY}"
SIGNATURE=$(echo -n "${MESSAGE}" | openssl dgst -sha256 -hmac "${API_SECRET}" | awk '{print $2}')

curl -X POST "http://localhost:8080/api/v1/blacklist/check" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -H "X-Timestamp: ${TIMESTAMP}" \
  -H "X-Nonce: ${NONCE}" \
  -H "X-Signature: ${SIGNATURE}" \
  -d "${BODY}"
```

## 📊 性能指标

### 预期性能
- **并发能力**: 20000+ QPS
- **查询延迟**: P95 < 5ms, P99 < 10ms
- **鉴权开销**: HMAC ~20μs (比JWT快5倍)
- **内存占用**: 2000万数据约640MB Redis内存
- **可用性**: 99.9%+

### 监控指标
- **QPS**: 每秒查询次数
- **命中率**: 黑名单命中百分比
- **响应时间**: P50/P95/P99延迟
- **错误率**: 4xx/5xx错误百分比
- **连接池状态**: 活跃/空闲连接数

## 🛡️ 安全机制

### HMAC鉴权安全
- **时间窗口**: ±300秒防重放
- **Nonce机制**: 随机数防重复请求
- **签名验证**: HMAC-SHA256防篡改
- **密钥管理**: 支持密钥轮换和过期

### 速率限制
- **滑动窗口**: 基于Redis ZSET实现
- **租户隔离**: 每个API Key独立限制
- **弹性配置**: 支持动态调整限制

### 多租户隔离
- **数据隔离**: 所有数据按tenant_id隔离
- **权限控制**: 基于现有permission系统
- **资源隔离**: Redis key包含租户前缀

## 🔄 运维管理

### 数据同步
```bash
# 同步租户黑名单到Redis
curl -X POST "http://localhost:8080/api/v1/admin/blacklist/sync" \
  -H "Authorization: Bearer ${JWT_TOKEN}"
```

### 统计查询
```bash
# 查看查询统计
curl "http://localhost:8080/api/v1/admin/blacklist/stats?hours=24" \
  -H "Authorization: Bearer ${JWT_TOKEN}"
```

### 健康检查
```bash
# 系统健康检查
curl "http://localhost:8080/health"
```

## 📈 扩容方案

### 水平扩容
- **应用层**: 多实例部署，负载均衡
- **Redis层**: Redis Cluster分片存储
- **MySQL层**: 读写分离，分库分表

### 缓存优化
- **预热策略**: 启动时异步加载热点数据
- **缓存穿透**: 布隆过滤器预过滤
- **缓存雪崩**: TTL随机化，多级缓存

---

**系统特点**: 基于现有shield架构，完全复用基础设施，零重复代码，高性能HMAC鉴权，智能日志采样，实时监控统计。