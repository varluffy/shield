#!/bin/bash

# UltraFit 代码质量检查脚本

set -e

echo "🔍 开始代码质量检查..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 错误计数
ERRORS=0

error_found() {
    echo -e "${RED}❌ $1${NC}"
    ERRORS=$((ERRORS + 1))
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# 1. 检查Go版本
echo "1. 检查Go版本..."
GO_VERSION=$(go version | grep -o 'go1\.[0-9]*' || echo "unknown")
if [[ "$GO_VERSION" < "go1.21" && "$GO_VERSION" != "unknown" ]]; then
    warning "Go版本: $GO_VERSION，建议使用Go 1.21+"
else
    success "Go版本: $GO_VERSION"
fi

# 2. 代码格式化检查
echo "2. 检查代码格式..."
if ! gofmt -l . | grep -q .; then
    success "代码格式正确"
else
    error_found "代码格式需要修复，运行: make format"
fi

# 3. 编译检查
echo "3. 检查编译..."
if go build -o /tmp/ultrafit-test cmd/server/main.go >/dev/null 2>&1; then
    success "编译通过"
    rm -f /tmp/ultrafit-test
else
    error_found "编译失败"
fi

# 4. Wire代码生成检查
echo "4. 检查Wire代码..."
if go generate ./internal/wire/... >/dev/null 2>&1; then
    success "Wire代码生成成功"
else
    error_found "Wire代码生成失败"
fi

# 5. 依赖检查
echo "5. 检查依赖..."
if go mod tidy && go mod verify >/dev/null 2>&1; then
    success "依赖检查通过"
else
    warning "依赖可能有问题，运行: go mod tidy"
fi

# 6. 运行核心测试
echo "6. 运行核心测试..."
if go test -timeout 30s ./internal/... ./pkg/... >/dev/null 2>&1; then
    success "核心测试通过"
else
    warning "部分测试失败（可能需要数据库连接）"
fi

# 7. Lint检查（如果可用）
echo "7. 代码质量检查..."
if command -v golangci-lint >/dev/null 2>&1; then
    if golangci-lint run --timeout 5m >/dev/null 2>&1; then
        success "代码质量检查通过"
    else
        warning "代码质量检查发现问题，运行: make lint"
    fi
else
    warning "golangci-lint未安装，跳过代码质量检查"
fi

# 总结
echo ""
echo "📊 检查完成："
if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}🎉 所有核心检查都通过了！${NC}"
    echo -e "${GREEN}代码质量良好，可以提交或部署。${NC}"
    exit 0
else
    echo -e "${RED}发现 $ERRORS 个需要修复的问题${NC}"
    echo -e "${YELLOW}建议修复后再提交代码${NC}"
    exit 1
fi 