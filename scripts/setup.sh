#!/bin/bash

# UltraFit 项目设置脚本

set -e

echo "🚀 正在设置 UltraFit 开发环境..."

# 检查 Go 版本
echo "📋 检查 Go 版本..."
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
GO_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
GO_MINOR=$(echo $GO_VERSION | cut -d. -f2)

if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 21 ]); then
    echo "❌ 需要 Go 1.21 或更高版本，当前版本: $GO_VERSION" 
    exit 1
fi

echo "✅ Go 版本: $GO_VERSION"

# 安装开发工具
echo "🔧 安装开发工具..."

tools=(
    "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    "golang.org/x/tools/cmd/goimports@latest" 
    "github.com/google/wire/cmd/wire@latest"
    "go.uber.org/mock/mockgen@latest"
)

for tool in "${tools[@]}"; do
    echo "安装 $tool..."
    go install $tool
done

# 下载项目依赖
echo "📦 下载项目依赖..."
go mod tidy

# 生成 Wire 代码
echo "⚡ 生成 Wire 代码..."
go generate ./...

# 检查 Docker
echo "🐳 检查 Docker..."
if command -v docker &> /dev/null; then
    echo "✅ Docker 已安装"
    
    # 启动依赖服务
    echo "🚀 启动依赖服务..."
    
    # 启动 MySQL
    echo "启动 MySQL..."
    docker run -d --name ultrafit-mysql \
        -e MYSQL_ROOT_PASSWORD=123456 \
        -e MYSQL_DATABASE=ultrafit_dev \
        -p 3306:3306 \
        mysql:8.0 2>/dev/null || echo "MySQL 容器可能已存在"
    
    # 启动 Jaeger
    echo "启动 Jaeger..."
    docker run -d --name jaeger \
        -p 16686:16686 \
        -p 14268:14268 \
        jaegertracing/all-in-one:latest 2>/dev/null || echo "Jaeger 容器可能已存在"
    
    echo "⏳ 等待服务启动..."
    sleep 10
    
else
    echo "⚠️  Docker 未安装，请手动启动 MySQL 和 Jaeger"
fi

# 运行数据库迁移
echo "🗄️  运行数据库迁移..."
go run cmd/migrate/main.go -config=configs/config.dev.yaml

echo "🎉 设置完成！"
echo ""
echo "📋 下一步："
echo "1. 启动应用: make dev"
echo "2. 访问健康检查: curl http://localhost:8080/health"
echo "3. 查看 Jaeger: http://localhost:16686"
echo ""
echo "🔧 常用命令:"
echo "- make help        # 查看所有命令"
echo "- make dev         # 开发模式运行"
echo "- make test        # 运行测试"
echo "- make migrate     # 数据库迁移" 