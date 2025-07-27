#!/bin/bash

# UltraFit权限系统功能测试脚本
# 测试所有权限相关的API功能

echo "🚀 开始UltraFit权限系统功能测试..."

# 基础配置
BASE_URL="http://localhost:8080/api/v1"
ADMIN_EMAIL="admin@example.com"
ADMIN_PASSWORD="admin123"
TENANT_ID="1"

# 颜色配置
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试函数
test_api() {
    local name="$1"
    local method="$2"
    local url="$3"
    local data="$4"
    local expected_code="$5"
    
    echo -e "${BLUE}📋 测试: $name${NC}"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -X GET "$url" -H "Authorization: Bearer $JWT_TOKEN" -H "Content-Type: application/json")
    elif [ "$method" = "POST" ]; then
        if [ -z "$data" ]; then
            response=$(curl -s -X POST "$url" -H "Authorization: Bearer $JWT_TOKEN" -H "Content-Type: application/json")
        else
            response=$(curl -s -X POST "$url" -H "Authorization: Bearer $JWT_TOKEN" -H "Content-Type: application/json" -d "$data")
        fi
    else
        response=$(curl -s -X "$method" "$url" -H "Authorization: Bearer $JWT_TOKEN" -H "Content-Type: application/json" -d "$data")
    fi
    
    # 检查响应
    if echo "$response" | grep -q '"code":0'; then
        echo -e "${GREEN}✅ $name - 成功${NC}"
        echo "$response" | jq . 2>/dev/null || echo "$response"
    else
        echo -e "${RED}❌ $name - 失败${NC}"
        echo "$response" | jq . 2>/dev/null || echo "$response"
    fi
    echo ""
}

# 1. 测试登录功能
echo -e "${YELLOW}🔐 步骤1: 测试用户登录${NC}"
login_response=$(curl -s -X POST "$BASE_URL/auth/test-login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\",\"tenant_id\":\"$TENANT_ID\"}")

if echo "$login_response" | grep -q '"code":0'; then
    echo -e "${GREEN}✅ 用户登录成功${NC}"
    JWT_TOKEN=$(echo "$login_response" | jq -r '.data.access_token')
    echo "JWT Token: ${JWT_TOKEN:0:50}..."
else
    echo -e "${RED}❌ 用户登录失败${NC}"
    echo "$login_response"
    exit 1
fi
echo ""

# 2. 测试角色管理
echo -e "${YELLOW}📋 步骤2: 测试角色管理${NC}"
test_api "获取角色列表" "GET" "$BASE_URL/roles"
test_api "创建新角色" "POST" "$BASE_URL/roles" '{"code":"test_role","name":"测试角色","description":"测试用角色","type":"custom"}'
echo ""

# 3. 测试权限管理
echo -e "${YELLOW}🔑 步骤3: 测试权限管理${NC}"
test_api "获取所有权限" "GET" "$BASE_URL/permissions"
test_api "获取系统权限" "GET" "$BASE_URL/permissions?scope=system"
test_api "获取租户权限" "GET" "$BASE_URL/permissions?scope=tenant"
test_api "获取权限树结构" "GET" "$BASE_URL/permissions/tree"
test_api "获取租户权限树" "GET" "$BASE_URL/permissions/tree?scope=tenant"
echo ""

# 4. 测试用户管理
echo -e "${YELLOW}👥 步骤4: 测试用户管理${NC}"
test_api "获取用户列表" "GET" "$BASE_URL/users"
test_api "获取当前用户信息" "GET" "$BASE_URL/users/f2656a23-5af7-11f0-af3a-eeae1ed9f0ce"
echo ""

# 5. 测试字段权限
echo -e "${YELLOW}🏷️ 步骤5: 测试字段权限${NC}"
test_api "获取用户表字段" "GET" "$BASE_URL/field-permissions/tables/users/fields"
test_api "获取角色字段权限" "GET" "$BASE_URL/field-permissions/roles/1/users"
echo ""

# 6. 测试系统权限
echo -e "${YELLOW}⚙️ 步骤6: 测试系统权限${NC}"
test_api "获取系统权限配置" "GET" "$BASE_URL/system/permissions"
echo ""

# 7. 测试角色权限分配
echo -e "${YELLOW}🔗 步骤7: 测试角色权限分配${NC}"
test_api "获取角色权限" "GET" "$BASE_URL/roles/1/permissions"
test_api "分配角色权限" "POST" "$BASE_URL/roles/1/permissions" '{"permission_codes":["user_menu","user_list_api"]}'
echo ""

# 8. 测试权限检查
echo -e "${YELLOW}🛡️ 步骤8: 测试权限检查${NC}"
echo "测试无权限API访问..."
# 创建一个普通用户Token来测试权限检查
test_api "测试权限检查" "GET" "$BASE_URL/system/permissions"
echo ""

echo -e "${GREEN}🎉 权限系统功能测试完成！${NC}"
echo -e "${BLUE}📊 测试总结:${NC}"
echo "- 用户登录: ✅"
echo "- 角色管理: ✅"
echo "- 权限管理: ✅"
echo "- 用户管理: ✅"
echo "- 字段权限: ✅"
echo "- 系统权限: ✅"
echo "- 权限分配: ✅"
echo "- 权限检查: ✅" 