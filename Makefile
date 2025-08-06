.PHONY: build run test clean deps fmt lint help

# 默认目标
.DEFAULT_GOAL := help

# 项目名称
PROJECT_NAME := vps
BINARY_NAME := vps

# Go相关变量
GO := go
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# 构建目录
BUILD_DIR := build
DIST_DIR := dist

# 版本信息
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# 帮助信息
help: ## 显示帮助信息
	@echo "VPS VPN Service - 基于Go Gin的VPN服务"
	@echo ""
	@echo "可用的命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 安装依赖
deps: ## 安装项目依赖
	@echo "安装项目依赖..."
	$(GO) mod download
	$(GO) mod tidy

# 代码格式化
fmt: ## 格式化代码
	@echo "格式化代码..."
	$(GO) fmt ./...

# 代码检查
lint: ## 运行代码检查
	@echo "运行代码检查..."
	$(GO) vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过 lint 检查"; \
	fi

# 运行测试
test: ## 运行测试
	@echo "运行测试..."
	$(GO) test -v ./...

# 构建项目
build: ## 构建项目
	@echo "构建项目..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go

# 构建所有平台
build-all: ## 构建所有平台的二进制文件
	@echo "构建所有平台的二进制文件..."
	@mkdir -p $(DIST_DIR)
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ]; then \
				ext=".exe"; \
			else \
				ext=""; \
			fi; \
			echo "构建 $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-$$os-$$arch$$ext ./cmd/main.go; \
		done; \
	done

# 运行项目
run: ## 运行项目
	@echo "运行项目..."
	$(GO) run ./cmd/main.go

# 运行项目（后台）
run-daemon: ## 后台运行项目
	@echo "后台运行项目..."
	@nohup $(GO) run ./cmd/main.go > vps.log 2>&1 & echo $$! > vps.pid

# 停止后台运行的项目
stop-daemon: ## 停止后台运行的项目
	@if [ -f vps.pid ]; then \
		kill $$(cat vps.pid) 2>/dev/null || true; \
		rm -f vps.pid; \
		echo "项目已停止"; \
	else \
		echo "没有找到运行中的项目"; \
	fi

# 清理构建文件
clean: ## 清理构建文件
	@echo "清理构建文件..."
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f vps.log
	rm -f vps.pid

# 安装到系统
install: build ## 安装到系统
	@echo "安装到系统..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)

# 卸载
uninstall: ## 卸载
	@echo "卸载..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

# 生成配置文件
config: ## 生成配置文件
	@echo "生成配置文件..."
	@if [ ! -f config.yaml ]; then \
		cp config.yaml.example config.yaml 2>/dev/null || \
		echo "请手动创建 config.yaml 文件"; \
	else \
		echo "config.yaml 已存在"; \
	fi

# 开发模式
dev: ## 开发模式（自动重启）
	@echo "开发模式..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air 未安装，使用 go run"; \
		$(GO) run ./cmd/main.go; \
	fi

# 检查依赖
check-deps: ## 检查依赖
	@echo "检查依赖..."
	$(GO) mod verify
	$(GO) list -m all

# 更新依赖
update-deps: ## 更新依赖
	@echo "更新依赖..."
	$(GO) get -u ./...
	$(GO) mod tidy

# 显示项目信息
info: ## 显示项目信息
	@echo "项目信息:"
	@echo "  名称: $(PROJECT_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  操作系统: $(GOOS)"
	@echo "  架构: $(GOARCH)"
	@echo "  Go版本: $(shell go version)" 