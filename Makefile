# Go 参数
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
WIRE_CMD=/Users/leng/go/bin/wire
SWAG_CMD=/Users/leng/go/bin/swag

# 二进制文件
SERVER_BINARY=./bin/server
MIGRATE_BINARY=./bin/migrate

# 配置文件
CONFIG_DEV=configs/config.dev.yaml
CONFIG_PROD=configs/config.prod.yaml

# 环境变量（可覆盖）
PORT ?= 8080

.PHONY: all build clean test deps wire docs run migrate help init dev prod

# 默认目标
all: deps wire build

# ===============================
# 核心命令
# ===============================

# 运行服务（开发模式）
run:
	@echo "启动开发服务器..."
	@$(GOCMD) run cmd/server/main.go

# 开发模式（包含代码生成）
dev: wire docs
	@echo "开发模式启动..."
	@$(GOCMD) run cmd/server/main.go

# 构建
build: wire docs
	@echo "构建项目..."
	@mkdir -p bin
	$(GOBUILD) -o $(SERVER_BINARY) cmd/server/main.go
	$(GOBUILD) -o $(MIGRATE_BINARY) ./cmd/migrate
	@echo "✅ 构建完成"

# 生产构建
prod: clean deps wire docs
	@echo "生产环境构建..."
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -o $(SERVER_BINARY) cmd/server/main.go
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -o $(MIGRATE_BINARY) ./cmd/migrate
	@echo "✅ 生产构建完成"

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

# ===============================
# 依赖管理
# ===============================

# 安装依赖
deps:
	@echo "安装依赖..."
	$(GOMOD) download
	$(GOMOD) tidy

# 生成 Wire 代码
wire:
	@echo "生成依赖注入代码..."
	$(WIRE_CMD) ./...

# 生成 API 文档
docs:
	@echo "生成 API 文档..."
	$(SWAG_CMD) init -g cmd/server/main.go --output docs --parseDependency --parseInternal

# ===============================
# 数据库管理
# ===============================

# 数据库迁移 (原有GORM方式)
migrate:
	@echo "运行数据库迁移..."
	$(GOCMD) run ./cmd/migrate -config=$(CONFIG_DEV)

# 版本化迁移管理
migrate-up:
	@echo "执行版本化迁移..."
	$(GOCMD) run ./cmd/migrate/*.go -config=$(CONFIG_DEV) -action=migrate-up

migrate-down:
	@echo "回滚最后一个迁移批次..."
	$(GOCMD) run ./cmd/migrate/*.go -config=$(CONFIG_DEV) -action=migrate-down

migrate-status:
	@echo "查看迁移状态..."
	$(GOCMD) run ./cmd/migrate/*.go -config=$(CONFIG_DEV) -action=migrate-status

create-migration:
	@if [ -z "$(name)" ]; then \
		echo "Error: migration name is required"; \
		echo "Usage: make create-migration name=migration_name"; \
		exit 1; \
	fi
	@echo "创建迁移文件: $(name)"
	@$(GOCMD) run ./cmd/migrate/*.go -config=$(CONFIG_DEV) -action=create-migration -migration-name=$(name)

# 初始化项目（首次使用）
init: deps wire docs migrate
	@echo "✅ 项目初始化完成"
	@echo "提示：使用 'make create-admin' 创建管理员账号"

# 创建管理员
create-admin:
	@echo "创建管理员用户..."
	@echo "请输入以下信息："
	@read -p "Email: " email; \
	read -p "Password: " password; \
	read -p "Name: " name; \
	$(GOCMD) run ./cmd/migrate -config=$(CONFIG_DEV) -action=create-user -email=$$email -password=$$password -name="$$name"

# 列出用户
list-users:
	@$(GOCMD) run ./cmd/migrate -config=$(CONFIG_DEV) -action=list-users

# ===============================
# 帮助信息
# ===============================

help:
	@echo "Shield - Go 微服务框架"
	@echo ""
	@echo "常用命令："
	@echo "  make init         初始化项目（首次使用）"
	@echo "  make dev          开发模式运行"
	@echo "  make test         运行测试"
	@echo "  make build        构建项目"
	@echo "  make prod         生产环境构建"
	@echo ""
	@echo "数据库迁移："
	@echo "  make migrate-up           执行版本化迁移"
	@echo "  make migrate-down         回滚最后一个批次"
	@echo "  make migrate-status       查看迁移状态"
	@echo "  make create-migration name=名称  创建新迁移文件"
	@echo ""
	@echo "数据库管理："
	@echo "  make migrate      运行数据库迁移"
	@echo "  make create-admin 创建管理员用户"
	@echo "  make list-users   列出所有用户"
	@echo ""
	@echo "其他命令："
	@echo "  make deps         安装依赖"
	@echo "  make wire         生成依赖注入代码"
	@echo "  make docs         生成 API 文档"
	@echo "  make clean        清理构建文件"
	@echo ""
	@echo "环境变量："
	@echo "  PORT=8080 make dev  指定端口运行"