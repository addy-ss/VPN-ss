 # VPS客户端 Makefile

.PHONY: build clean run stop help

# 默认目标
all: build

# 构建客户端
build:
	@echo "构建VPS客户端..."
	go mod tidy
	go build -o vps-client main.go
	@echo "构建完成: vps-client"

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -f vps-client
	@echo "清理完成"

# 运行客户端（使用默认配置）
run: build
	@echo "启动VPS客户端..."
	./vps-client

# 停止客户端
stop:
	@echo "停止VPS客户端..."
	@if [ -f stop.sh ]; then \
		./stop.sh; \
	else \
		pkill -f vps-client || echo "未找到运行中的客户端"; \
	fi

# 安装依赖
deps:
	@echo "安装依赖..."
	go mod download

# 测试
test:
	@echo "运行测试..."
	go test ./...

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
lint:
	@echo "代码检查..."
	golangci-lint run

# 显示帮助
help:
	@echo "VPS客户端 Makefile 命令:"
	@echo "  build   - 构建客户端"
	@echo "  clean   - 清理构建文件"
	@echo "  run     - 构建并运行客户端"
	@echo "  stop    - 停止客户端"
	@echo "  deps    - 安装依赖"
	@echo "  test    - 运行测试"
	@echo "  fmt     - 格式化代码"
	@echo "  lint    - 代码检查"
	@echo "  help    - 显示此帮助信息"