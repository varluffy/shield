#!/bin/bash

# 黑名单API测试脚本

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# API基础URL
BASE_URL="http://localhost:8080/api/v1"

# 测试数据
TEST_EMAIL="admin@example.com"
TEST_PASSWORD="admin123"
TEST_TENANT_ID="1"
TEST_PHONE_MD5="5d41402abc4b2a76b9719d911017c592"

echo -e "${YELLOW}=== 黑名单API测试开始 ===${NC}"

# 1. 登录获取JWT Token
echo -e "\n${GREEN}1. 登录获取JWT Token...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/test-login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${TEST_EMAIL}\",\"password\":\"${TEST_PASSWORD}\",\"tenant_id\":\"${TEST_TENANT_ID}\"}")

JWT_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.access_token')

if [ "$JWT_TOKEN" == "null" ] || [ -z "$JWT_TOKEN" ]; then
    echo -e "${RED}登录失败！${NC}"
    echo "响应: $LOGIN_RESPONSE"
    exit 1
fi

echo -e "${GREEN}登录成功！Token获取成功${NC}"

# 2. 创建API密钥
echo -e "\n${GREEN}2. 创建API密钥...${NC}"
CREATE_API_KEY_RESPONSE=$(curl -s -X POST "${BASE_URL}/admin/blacklist/api-credentials" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试密钥",
    "description": "用于测试的API密钥",
    "rate_limit": 1000
  }')

API_KEY=$(echo $CREATE_API_KEY_RESPONSE | jq -r '.data.api_key')
API_SECRET=$(echo $CREATE_API_KEY_RESPONSE | jq -r '.data.api_secret')

if [ "$API_KEY" == "null" ] || [ -z "$API_KEY" ]; then
    echo -e "${YELLOW}API密钥可能已存在，使用预设密钥进行测试${NC}"
    # 使用预设的测试密钥（需要在数据库中预先创建）
    API_KEY="test_api_key"
    API_SECRET="test_api_secret"
else
    echo -e "${GREEN}API密钥创建成功！${NC}"
    echo "API Key: $API_KEY"
    echo "API Secret: $API_SECRET"
fi

# 3. 创建黑名单记录
echo -e "\n${GREEN}3. 创建黑名单记录...${NC}"
curl -s -X POST "${BASE_URL}/admin/blacklist" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{
    \"phone_md5\": \"${TEST_PHONE_MD5}\",
    \"source\": \"manual\",
    \"reason\": \"测试黑名单\"
  }" | jq

# 4. 测试HMAC签名查询
echo -e "\n${GREEN}4. 测试HMAC签名查询...${NC}"

# 生成签名参数
TIMESTAMP=$(date +%s)
NONCE=$(openssl rand -hex 16)
BODY="{\"phone_md5\":\"${TEST_PHONE_MD5}\"}"

# 计算HMAC签名
SIGN_STRING="${API_KEY}${TIMESTAMP}${NONCE}${BODY}"
SIGNATURE=$(echo -n "${SIGN_STRING}" | openssl dgst -sha256 -hmac "${API_SECRET}" | cut -d' ' -f2)

echo "Timestamp: $TIMESTAMP"
echo "Nonce: $NONCE"
echo "Signature: $SIGNATURE"

# 发送查询请求
QUERY_RESPONSE=$(curl -s -X POST "${BASE_URL}/blacklist/check" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Key: ${API_KEY}" \
  -H "X-Timestamp: ${TIMESTAMP}" \
  -H "X-Nonce: ${NONCE}" \
  -H "X-Signature: ${SIGNATURE}" \
  -d "${BODY}")

echo -e "\n查询响应:"
echo $QUERY_RESPONSE | jq

# 5. 获取统计信息
echo -e "\n${GREEN}5. 获取查询统计信息...${NC}"
curl -s -X GET "${BASE_URL}/admin/blacklist/stats?hours=1" \
  -H "Authorization: Bearer ${JWT_TOKEN}" | jq

# 6. 获取分钟级统计
echo -e "\n${GREEN}6. 获取分钟级统计信息...${NC}"
curl -s -X GET "${BASE_URL}/admin/blacklist/stats/minutes?minutes=5" \
  -H "Authorization: Bearer ${JWT_TOKEN}" | jq

# 7. 性能测试（发送多个请求）
echo -e "\n${GREEN}7. 性能测试（发送10个并发请求）...${NC}"

for i in {1..10}; do
    # 为每个请求生成新的时间戳和Nonce
    TIMESTAMP=$(date +%s)
    NONCE=$(openssl rand -hex 16)
    SIGN_STRING="${API_KEY}${TIMESTAMP}${NONCE}${BODY}"
    SIGNATURE=$(echo -n "${SIGN_STRING}" | openssl dgst -sha256 -hmac "${API_SECRET}" | cut -d' ' -f2)
    
    # 异步发送请求
    curl -s -X POST "${BASE_URL}/blacklist/check" \
      -H "Content-Type: application/json" \
      -H "X-Tenant-Key: ${API_KEY}" \
      -H "X-Timestamp: ${TIMESTAMP}" \
      -H "X-Nonce: ${NONCE}" \
      -H "X-Signature: ${SIGNATURE}" \
      -d "${BODY}" > /dev/null &
done

# 等待所有请求完成
wait
echo -e "${GREEN}性能测试完成！${NC}"

# 8. 再次获取统计信息查看变化
echo -e "\n${GREEN}8. 再次获取分钟级统计查看变化...${NC}"
sleep 2  # 等待统计更新
curl -s -X GET "${BASE_URL}/admin/blacklist/stats/minutes?minutes=1" \
  -H "Authorization: Bearer ${JWT_TOKEN}" | jq

echo -e "\n${YELLOW}=== 黑名单API测试完成 ===${NC}"