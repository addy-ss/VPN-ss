 #!/bin/bash

# VPS客户端快速启动脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    print_info "检查依赖..."
    
    if ! command -v go &> /dev/null; then
        print_error "未找到Go编译器，请先安装Go"
        echo "安装Go: https://golang.org/doc/install"
        exit 1
    fi
    
    print_success "Go已安装: $(go version)"
}

# 安装依赖
install_dependencies() {
    print_info "安装Go依赖..."
    go mod tidy
    print_success "依赖安装完成"
}

# 构建客户端
build_client() {
    print_info "构建客户端..."
    go build -o vps-client main.go
    if [ $? -eq 0 ]; then
        print_success "客户端构建完成"
    else
        print_error "客户端构建失败"
        exit 1
    fi
}

# 配置客户端
configure_client() {
    print_info "配置客户端..."
    
    if [ ! -f config.yaml ]; then
        print_warning "未找到配置文件，创建默认配置..."
        cat > config.yaml << EOF
# VPS客户端配置文件

# 服务器配置
server:
  host: "206.190.238.198"  # 服务器地址
  port: 8388          # 服务器端口

# 本地配置
local:
  port: 1080          # 本地监听端口

# 安全配置
security:
  password: "13687401432Fan!"  # 密码
  method: "aes-256-gcm"      # 加密方法
  timeout: 300               # 超时时间(秒)

# 日志配置
log:
  level: "info"              # 日志级别
  file: ""                   # 日志文件路径
EOF
        print_success "默认配置文件已创建"
    else
        print_info "使用现有配置文件"
    fi
}

# 启动客户端
start_client() {
    print_info "启动客户端..."
    
    # 检查端口是否被占用
    if lsof -Pi :1080 -sTCP:LISTEN -t >/dev/null ; then
        print_warning "端口 1080 已被占用，尝试停止现有进程..."
        pkill -f vps-client || true
        sleep 2
    fi
    
    # 启动客户端
    ./vps-client &
    CLIENT_PID=$!
    
    # 等待启动
    sleep 2
    
    # 检查是否启动成功
    if kill -0 $CLIENT_PID 2>/dev/null; then
        print_success "客户端启动成功 (PID: $CLIENT_PID)"
        echo "客户端正在监听端口 1080"
        echo "按 Ctrl+C 停止客户端"
        
        # 等待用户中断
        trap "stop_client" INT TERM
        wait $CLIENT_PID
    else
        print_error "客户端启动失败"
        exit 1
    fi
}

# 停止客户端
stop_client() {
    print_info "停止客户端..."
    pkill -f vps-client || true
    print_success "客户端已停止"
}

# 显示帮助
show_help() {
    echo "VPS客户端快速启动脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --build-only    仅构建客户端，不启动"
    echo "  --configure     仅配置客户端，不构建和启动"
    echo "  --help          显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0              # 完整流程：检查依赖 -> 构建 -> 配置 -> 启动"
    echo "  $0 --build-only # 仅构建客户端"
    echo "  $0 --configure  # 仅配置客户端"
}

# 主函数
main() {
    case "${1:-}" in
        --help)
            show_help
            exit 0
            ;;
        --build-only)
            check_dependencies
            install_dependencies
            build_client
            print_success "构建完成"
            exit 0
            ;;
        --configure)
            configure_client
            print_success "配置完成"
            exit 0
            ;;
        "")
            # 完整流程
            check_dependencies
            install_dependencies
            build_client
            configure_client
            start_client
            ;;
        *)
            print_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"