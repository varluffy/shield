# Go参数
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
WIRE_CMD=/Users/leng/go/bin/wire
SWAG_CMD=/Users/leng/go/bin/swag

# 二进制文件
BINARY_NAME=ultrafit
BINARY_PATH=./$(BINARY_NAME)
SERVER_BINARY=./bin/server
MIGRATE_BINARY=./bin/migrate  
ADMIN_BINARY=./bin/admin

# 配置文件
CONFIG_DEV=configs/config.dev.yaml
CONFIG_PROD=configs/config.prod.yaml

# 默认端口
DEFAULT_PORT=8080

# 项目信息
PROJECT_NAME=ultrafit
VERSION=1.0.0

.PHONY: all build clean test deps wire docs run migrate admin help setup init create-admin quick-init full-setup \
        check-port kill-port stop-service start-service restart-service safe-run status

# 默认目标
all: clean deps wire build

# ===============================
# 🚀 服务管理命令（重要！）
# ===============================

# 检查端口是否被占用
check-port:
	@echo "检查端口 $(DEFAULT_PORT) 是否被占用..."
	@lsof -i :$(DEFAULT_PORT) && echo "⚠️  端口 $(DEFAULT_PORT) 被占用！" || echo "✅ 端口 $(DEFAULT_PORT) 未被占用"

# 杀死占用端口的进程
kill-port:
	@echo "停止占用端口 $(DEFAULT_PORT) 的进程..."
	@lsof -ti :$(DEFAULT_PORT) | xargs -r kill -9 && echo "✅ 已停止占用端口的进程" || echo "ℹ️  没有进程占用端口"

# 停止所有ultrafit相关服务
stop-service:
	@echo "停止所有ultrafit相关服务..."
	@pkill -f "ultrafit\|go run.*cmd/server" && echo "✅ 已停止ultrafit服务" || echo "ℹ️  没有ultrafit服务在运行"
	@make kill-port

# 安全启动（先检查端口，再启动）
safe-run: check-port
	@echo "开始安全启动ultrafit服务..."
	@if lsof -i :$(DEFAULT_PORT) > /dev/null 2>&1; then \
		echo "❌ 端口 $(DEFAULT_PORT) 被占用，请先运行 'make stop-service'"; \
		exit 1; \
	else \
		echo "✅ 端口可用，正在启动服务..."; \
		$(GOCMD) run cmd/server/main.go; \
	fi

# 重启服务
restart-service: stop-service
	@echo "等待2秒后重启服务..."
	@sleep 2
	@make safe-run

# 检查服务状态
status:
	@echo "=== 服务状态检查 ==="
	@echo "端口状态:"
	@make check-port
	@echo ""
	@echo "ultrafit进程:"
	@ps aux | grep -E "ultrafit|go run.*cmd/server" | grep -v grep || echo "没有ultrafit进程在运行"

# 后台启动服务
start-service: check-port
	@echo "在后台启动ultrafit服务..."
	@if lsof -i :$(DEFAULT_PORT) > /dev/null 2>&1; then \
		echo "❌ 端口 $(DEFAULT_PORT) 被占用，请先运行 'make stop-service'"; \
		exit 1; \
	else \
		echo "✅ 端口可用，正在后台启动服务..."; \
		nohup $(GOCMD) run cmd/server/main.go > app.log 2>&1 & \
		echo "✅ 服务已在后台启动，日志文件: app.log"; \
		echo "使用 'make status' 检查服务状态"; \
	fi

# ===============================
# 🔧 开发命令
# ===============================

# 运行服务（开发模式，带前置检查）
run: stop-service
	@echo "等待1秒让进程完全停止..."
	@sleep 1
	@make wire
	@make docs
	@echo "✅ 端口清理完成，正在启动服务..."
	$(GOCMD) run cmd/server/main.go

# 构建所有二进制文件
build: wire docs
	@echo "构建所有二进制文件..."
	@mkdir -p bin
	$(GOBUILD) -o $(SERVER_BINARY) cmd/server/main.go
	$(GOBUILD) -o $(MIGRATE_BINARY) cmd/migrate/main.go
	$(GOBUILD) -o $(ADMIN_BINARY) cmd/admin/main.go

# 清理
clean:
	@echo "清理构建文件..."
	$(GOCLEAN)
	@rm -rf bin/
	@rm -f app.log

# 测试
test:
	@echo "运行测试..."
	$(GOTEST) -v -cover ./...

# 安装依赖
deps:
	@echo "安装依赖..."
	$(GOMOD) download
	$(GOMOD) tidy

# 生成Wire代码
wire:
	@echo "生成Wire代码..."
	$(WIRE_CMD) ./...

# 生成API文档
docs:
	@echo "生成API文档..."
	$(SWAG_CMD) init -g cmd/server/main.go --output docs --parseDependency --parseInternal

# ===============================
# 🗄️ 数据库命令
# ===============================

# 数据库迁移
migrate:
	@echo "运行数据库迁移..."
	$(GOCMD) run cmd/migrate/main.go -config=$(CONFIG_DEV)

# 管理员工具
admin:
	@echo "运行管理工具..."
	$(GOCMD) run cmd/admin/main.go -config=$(CONFIG_DEV)

# 创建管理员用户
create-admin:
	@echo "创建管理员用户..."
	@make admin

# ===============================
# ⚙️ 初始化命令
# ===============================

# 基本设置
setup:
	@echo "设置项目环境..."
	@make deps
	@make wire
	@make docs
	@echo "✅ 项目设置完成！"

# 项目初始化
init: setup
	@echo "初始化项目..."
	@make migrate
	@echo "✅ 项目初始化完成！"

# 快速初始化（用于开发）
quick-init: stop-service
	@echo "快速初始化项目..."
	@make deps
	@make wire
	@make migrate
	@echo "✅ 快速初始化完成！"

# 完整设置（用于新环境）
full-setup: stop-service
	@echo "完整设置项目..."
	@make setup
	@make migrate
	@make create-admin
	@echo "✅ 完整设置完成！"

# 开发模式运行（安全启动）
dev: stop-service
	@echo "开发模式启动..."
	@make wire
	@make docs
	@make start-service

# 生产构建
build-prod: clean deps wire
	@echo "生产构建..."
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -o $(SERVER_BINARY) cmd/server/main.go
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -o $(MIGRATE_BINARY) cmd/migrate/main.go
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -o $(ADMIN_BINARY) cmd/admin/main.go

# ===============================
# 💡 帮助信息
# ===============================

help:
	@echo "UltraFit - Go微服务开发工具"
	@echo ""
	@echo "🚀 服务管理命令（重要！）："
	@echo "  make check-port      - 检查端口是否被占用"
	@echo "  make kill-port       - 杀死占用端口的进程"
	@echo "  make stop-service    - 停止所有ultrafit服务"
	@echo "  make safe-run        - 安全启动服务（先检查端口）"
	@echo "  make start-service   - 后台启动服务"
	@echo "  make restart-service - 重启服务"
	@echo "  make status          - 检查服务状态"
	@echo ""
	@echo "🔧 开发命令："
	@echo "  make run             - 运行服务（开发模式，带前置检查）"
	@echo "  make build           - 构建所有二进制文件"
	@echo "  make clean           - 清理构建文件"
	@echo "  make test            - 运行测试"
	@echo "  make deps            - 安装依赖"
	@echo "  make wire            - 生成Wire代码"
	@echo "  make docs            - 生成API文档"
	@echo ""
	@echo "🗄️ 数据库命令："
	@echo "  make migrate         - 数据库迁移"
	@echo "  make admin           - 管理员工具"
	@echo "  make create-admin    - 创建管理员用户"
	@echo ""
	@echo "⚙️ 初始化命令："
	@echo "  make init            - 项目初始化"
	@echo "  make setup           - 基本设置"
	@echo "  make quick-init      - 快速初始化"
	@echo "  make full-setup      - 完整设置"
	@echo "  make dev             - 开发模式运行"
	@echo ""
	@echo "💡 推荐流程："
	@echo "  1. make init         - 首次设置"
	@echo "  2. make run          - 开发运行（自动停止旧服务）"
	@echo "  3. make stop-service - 停止服务"
	@echo "  4. make status       - 检查状态"
	@echo ""
	@echo "⚠️  注意事项："
	@echo "  • 所有启动命令都会自动检查端口占用"
	@echo "  • 'make run' 会自动停止旧服务再启动"
	@echo "  • 使用 'make stop-service' 彻底停止服务"
	@echo "  • 使用 'make status' 检查服务状态" 