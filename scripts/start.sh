#!/bin/bash

# VPS VPN Service 启动脚本

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

# 检查Go是否安装
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go未安装，请先安装Go 1.21或更高版本"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "Go版本: $GO_VERSION"
}

# 安装依赖
install_deps() {
    print_info "安装项目依赖..."
    go mod tidy
    print_success "依赖安装完成"
}

# 构建项目
build_project() {
    print_info "构建项目..."
    go build -o vps cmd/main.go
    print_success "项目构建完成"
}

# 检查配置文件
check_config() {
    if [ ! -f "config.yaml" ]; then
        print_warning "配置文件 config.yaml 不存在"
        if [ -f "config.example.yaml" ]; then
            print_info "复制示例配置文件..."
            cp config.example.yaml config.yaml
            print_success "配置文件已创建，请根据需要修改"
        else
            print_error "示例配置文件不存在"
            exit 1
        fi
    else
        print_success "配置文件已存在"
    fi
}

# 启动服务
start_service() {
    print_info "启动VPS VPN服务..."
    
    # 检查是否已经在运行
    if pgrep -f "vps" > /dev/null; then
        print_warning "服务可能已经在运行"
        read -p "是否继续启动？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 0
        fi
    fi
    
    # 启动服务
    ./vps &
    VPS_PID=$!
    echo $VPS_PID > vps.pid
    
    print_success "服务已启动，PID: $VPS_PID"
    print_info "HTTP API: http://localhost:8080"
    print_info "Shadowsocks端口: 8388"
    print_info "使用 Ctrl+C 停止服务"
    
    # 等待服务启动
    sleep 2
    
    # 检查服务是否正常启动
    if curl -s http://localhost:8080/api/v1/health > /dev/null; then
        print_success "服务启动成功！"
    else
        print_warning "服务可能未正常启动，请检查日志"
    fi
}

# 停止服务
stop_service() {
    if [ -f "vps.pid" ]; then
        PID=$(cat vps.pid)
        if kill -0 $PID 2>/dev/null; then
            print_info "停止服务 (PID: $PID)..."
            kill $PID
            rm -f vps.pid
            print_success "服务已停止"
        else
            print_warning "服务未在运行"
            rm -f vps.pid
        fi
    else
        print_warning "PID文件不存在"
    fi
}

# 显示帮助信息
show_help() {
    echo "VPS VPN Service 启动脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  start     启动服务"
    echo "  stop      停止服务"
    echo "  restart   重启服务"
    echo "  build     构建项目"
    echo "  install   安装依赖"
    echo "  status    查看服务状态"
    echo "  help      显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 start    # 启动服务"
    echo "  $0 stop     # 停止服务"
    echo "  $0 status   # 查看状态"
}

# 查看服务状态
check_status() {
    if [ -f "vps.pid" ]; then
        PID=$(cat vps.pid)
        if kill -0 $PID 2>/dev/null; then
            print_success "服务正在运行 (PID: $PID)"
            
            # 检查API是否响应
            if curl -s http://localhost:8080/api/v1/health > /dev/null; then
                print_success "API服务正常"
            else
                print_warning "API服务无响应"
            fi
            
            # 检查端口
            if netstat -an 2>/dev/null | grep -q ":8388 "; then
                print_success "Shadowsocks端口正常"
            else
                print_warning "Shadowsocks端口未监听"
            fi
        else
            print_warning "服务未运行 (PID文件存在但进程不存在)"
            rm -f vps.pid
        fi
    else
        print_warning "服务未运行"
    fi
}

# 主函数
main() {
    case "${1:-start}" in
        start)
            check_go
            install_deps
            build_project
            check_config
            start_service
            ;;
        stop)
            stop_service
            ;;
        restart)
            stop_service
            sleep 2
            check_go
            install_deps
            build_project
            check_config
            start_service
            ;;
        build)
            check_go
            build_project
            ;;
        install)
            check_go
            install_deps
            ;;
        status)
            check_status
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 捕获中断信号
trap 'print_info "收到中断信号，正在停止服务..."; stop_service; exit 0' INT TERM

# 运行主函数
main "$@" 