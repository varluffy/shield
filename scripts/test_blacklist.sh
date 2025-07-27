#!/bin/bash

# 黑名单API测试脚本
# 用于测试HMAC鉴权和黑名单查询功能

set -e

# 配置
BASE_URL="http://localhost:8080/api/v1"
API_KEY="test_api_key_123456"
API_SECRET="test_secret_abcdef123456"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印函数
print_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 生成HMAC签名
generate_hmac_signature() {
    local api_key="$1"
    local timestamp="$2"
    local nonce="$3"
    local body="$4"
    local secret="$5"
    
    # 构建签名字符串：api_key + timestamp + nonce + body
    local message="${api_key}${timestamp}${nonce}${body}"
    
    # 生成HMAC-SHA256签名
    echo -n "${message}" | openssl dgst -sha256 -hmac "${secret}" | awk '{print $2}'
}

# 测试黑名单查询API
test_blacklist_check() {
    print_info "测试黑名单查询API..."
    
    local phone_md5="5d41402abc4b2a76b9719d911017c592"
    local timestamp=$(date +%s)
    local nonce=$(openssl rand -hex 16)
    local body="{\"phone_md5\":\"${phone_md5}\"}"
    
    # 生成签名
    local signature=$(generate_hmac_signature "${API_KEY}" "${timestamp}" "${nonce}" "${body}" "${API_SECRET}")
    
    print_info "API Key: ${API_KEY}"
    print_info "Timestamp: ${timestamp}"
    print_info "Nonce: ${nonce}"
    print_info "Signature: ${signature}"
    print_info "Body: ${body}"
    
    # 发送请求
    local response=$(curl -s -w "\n%{http_code}" \
        -X POST "${BASE_URL}/blacklist/check" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: ${API_KEY}" \
        -H "X-Timestamp: ${timestamp}" \
        -H "X-Nonce: ${nonce}" \
        -H "X-Signature: ${signature}" \
        -d "${body}")
    
    local http_code=$(echo "$response" | tail -n1)
    local body_response=$(echo "$response" | head -n -1)
    
    echo "HTTP Status: ${http_code}"
    echo "Response: ${body_response}"
    
    if [ "$http_code" = "200" ]; then
        print_success "黑名单查询API测试成功"
    else
        print_error "黑名单查询API测试失败，HTTP状态码: ${http_code}"
    fi
}

# 测试无效签名
test_invalid_signature() {
    print_info "测试无效签名..."
    
    local phone_md5="5d41402abc4b2a76b9719d911017c592"
    local timestamp=$(date +%s)
    local nonce=$(openssl rand -hex 16)
    local body="{\"phone_md5\":\"${phone_md5}\"}"
    local invalid_signature="invalid_signature"
    
    # 发送请求
    local response=$(curl -s -w "\n%{http_code}" \
        -X POST "${BASE_URL}/blacklist/check" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: ${API_KEY}" \
        -H "X-Timestamp: ${timestamp}" \
        -H "X-Nonce: ${nonce}" \
        -H "X-Signature: ${invalid_signature}" \
        -d "${body}")
    
    local http_code=$(echo "$response" | tail -n1)
    local body_response=$(echo "$response" | head -n -1)
    
    echo "HTTP Status: ${http_code}"
    echo "Response: ${body_response}"
    
    if [ "$http_code" = "401" ]; then
        print_success "无效签名测试成功 - 正确拒绝了无效签名"
    else
        print_error "无效签名测试失败 - 应该返回401状态码"
    fi
}

# 测试过期时间戳
test_expired_timestamp() {
    print_info "测试过期时间戳..."
    
    local phone_md5="5d41402abc4b2a76b9719d911017c592"
    local timestamp=$(($(date +%s) - 400))  # 400秒前的时间戳
    local nonce=$(openssl rand -hex 16)
    local body="{\"phone_md5\":\"${phone_md5}\"}"
    
    # 生成签名
    local signature=$(generate_hmac_signature "${API_KEY}" "${timestamp}" "${nonce}" "${body}" "${API_SECRET}")
    
    # 发送请求
    local response=$(curl -s -w "\n%{http_code}" \
        -X POST "${BASE_URL}/blacklist/check" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: ${API_KEY}" \
        -H "X-Timestamp: ${timestamp}" \
        -H "X-Nonce: ${nonce}" \
        -H "X-Signature: ${signature}" \
        -d "${body}")
    
    local http_code=$(echo "$response" | tail -n1)
    local body_response=$(echo "$response" | head -n -1)
    
    echo "HTTP Status: ${http_code}"
    echo "Response: ${body_response}"
    
    if [ "$http_code" = "401" ]; then
        print_success "过期时间戳测试成功 - 正确拒绝了过期请求"
    else
        print_error "过期时间戳测试失败 - 应该返回401状态码"
    fi
}

# 主函数
main() {
    print_info "开始黑名单API测试..."
    print_info "请确保服务器正在运行: make run"
    echo
    
    # 等待用户确认
    read -p "服务器是否已启动？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_error "请先启动服务器: make run"
        exit 1
    fi
    
    echo "======================================"
    test_blacklist_check
    echo
    
    echo "======================================"
    test_invalid_signature
    echo
    
    echo "======================================"
    test_expired_timestamp
    echo
    
    print_success "所有测试完成！"
}

# 运行主函数
main "$@"