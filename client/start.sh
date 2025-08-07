#!/bin/bash

# VPS客户端启动脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 检查配置文件
check_config() {
    if [ ! -f "config.yaml" ]; then
        print_error "配置文件 config.yaml 不存在"
        print_info "请先创建配置文件"
        exit 1
    fi
    print_success "配置文件检查通过"
}

# 检查客户端程序
check_client() {
    if [ ! -f "vps-client" ]; then
        print_warning "客户端程序不存在，正在构建..."
        go build -o vps-client .
        if [ $? -ne 0 ]; then
            print_error "构建失败"
            exit 1
        fi
        print_success "客户端构建完成"
    fi
}

# 启动客户端
start_client() {
    print_info "启动VPS客户端..."
    
    # 检查端口是否被占用
    if lsof -i :1080 >/dev/null 2>&1; then
        print_warning "端口1080已被占用"
        read -p "是否停止占用进程并继续？(y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_info "停止占用进程..."
            pkill -f "vps-client"
            sleep 2
        else
            print_error "启动取消"
            exit 1
        fi
    fi
    
    # 启动客户端
    ./vps-client &
    CLIENT_PID=$!
    echo $CLIENT_PID > vps-client.pid
    
    # 等待客户端启动
    sleep 2
    
    # 检查客户端是否真正启动成功
    if kill -0 $CLIENT_PID 2>/dev/null; then
        # 检查端口是否监听
        if lsof -i :1080 >/dev/null 2>&1; then
            print_success "客户端已启动 (PID: $CLIENT_PID)"
            return 0
        else
            print_error "客户端启动失败：端口1080未监听"
            kill $CLIENT_PID 2>/dev/null
            rm -f vps-client.pid
            return 1
        fi
    else
        print_error "客户端启动失败：进程不存在"
        rm -f vps-client.pid
        return 1
    fi
}

# 停止客户端
stop_client() {
    if [ -f "vps-client.pid" ]; then
        PID=$(cat vps-client.pid)
        if kill -0 $PID 2>/dev/null; then
            print_info "停止客户端 (PID: $PID)..."
            kill $PID
            rm -f vps-client.pid
            print_success "客户端已停止"
        else
            print_warning "客户端进程不存在"
            rm -f vps-client.pid
        fi
    else
        print_warning "未找到PID文件"
    fi
}

# 检查状态
check_status() {
    if [ -f "vps-client.pid" ]; then
        PID=$(cat vps-client.pid)
        if kill -0 $PID 2>/dev/null; then
            print_success "客户端正在运行 (PID: $PID)"
        else
            print_warning "客户端进程不存在"
            rm -f vps-client.pid
        fi
    else
        print_warning "客户端未运行"
    fi
}

# 测试连接
test_connection() {
    print_info "测试SOCKS5代理连接..."
    if curl --socks5 127.0.0.1:1080 --connect-timeout 10 -s http://httpbin.org/ip >/dev/null 2>&1; then
        print_success "SOCKS5代理连接正常"
    else
        print_error "SOCKS5代理连接失败"
    fi
}

# 主函数
main() {
    case "${1:-start}" in
        "start")
            check_config
            check_client
            if start_client; then
                sleep 2
                test_connection
            else
                print_error "客户端启动失败，跳过连接测试"
                exit 1
            fi
            ;;
        "stop")
            stop_client
            ;;
        "restart")
            stop_client
            sleep 1
            check_config
            check_client
            if start_client; then
                sleep 2
                test_connection
            else
                print_error "客户端启动失败，跳过连接测试"
                exit 1
            fi
            ;;
        "status")
            check_status
            ;;
        "test")
            test_connection
            ;;
        *)
            echo "用法: $0 {start|stop|restart|status|test}"
            echo ""
            echo "命令:"
            echo "  start   - 启动客户端"
            echo "  stop    - 停止客户端"
            echo "  restart - 重启客户端"
            echo "  status  - 检查状态"
            echo "  test    - 测试连接"
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"