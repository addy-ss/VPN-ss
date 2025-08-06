#!/bin/bash

# 测试客户端和服务端连接的脚本

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

# 检查服务端状态
check_server() {
    print_info "检查服务端状态..."
    
    if lsof -i :8388 >/dev/null 2>&1; then
        print_success "服务端正在运行 (端口 8388)"
        SERVER_PID=$(lsof -t -i :8388)
        echo "服务端PID: $SERVER_PID"
    else
        print_error "服务端未运行 (端口 8388)"
        return 1
    fi
}

# 检查客户端状态
check_client() {
    print_info "检查客户端状态..."
    
    if lsof -i :1080 >/dev/null 2>&1; then
        print_success "客户端正在运行 (端口 1080)"
        CLIENT_PID=$(lsof -t -i :1080)
        echo "客户端PID: $CLIENT_PID"
    else
        print_error "客户端未运行 (端口 1080)"
        return 1
    fi
}

# 测试网络连接
test_network() {
    print_info "测试网络连接..."
    
    # 测试到服务器的连接
    if nc -z 206.190.238.198 8388 2>/dev/null; then
        print_success "网络连接到服务器正常"
    else
        print_error "无法连接到服务器 206.190.238.198:8388"
        return 1
    fi
}

# 测试SOCKS5代理
test_socks5() {
    print_info "测试SOCKS5代理..."
    
    # 使用curl测试SOCKS5代理
    if curl --socks5 127.0.0.1:1080 --connect-timeout 10 -s http://httpbin.org/ip >/dev/null 2>&1; then
        print_success "SOCKS5代理工作正常"
        return 0
    else
        print_error "SOCKS5代理测试失败"
        return 1
    fi
}

# 显示详细连接信息
show_connection_info() {
    print_info "显示连接信息..."
    
    echo "=== 服务端配置 ==="
    if [ -f config.yaml ]; then
        echo "配置文件: config.yaml"
        grep -A 10 "shadowsocks:" config.yaml | head -10
    else
        echo "未找到配置文件"
    fi
    
    echo ""
    echo "=== 客户端配置 ==="
    if [ -f client/config.yaml ]; then
        echo "客户端配置文件: client/config.yaml"
        cat client/config.yaml
    else
        echo "未找到客户端配置文件"
    fi
    
    echo ""
    echo "=== 网络连接 ==="
    netstat -an | grep -E "(8388|1080)" | head -5
}

# 主函数
main() {
    echo "=== VPS连接测试 ==="
    echo ""
    
    # 检查服务端
    if ! check_server; then
        print_error "服务端检查失败"
        exit 1
    fi
    
    # 检查客户端
    if ! check_client; then
        print_error "客户端检查失败"
        exit 1
    fi
    
    # 测试网络连接
    if ! test_network; then
        print_error "网络连接测试失败"
        exit 1
    fi
    
    # 测试SOCKS5代理
    if ! test_socks5; then
        print_warning "SOCKS5代理测试失败，但服务端和客户端都在运行"
        print_info "可能的问题："
        print_info "1. 密码不匹配"
        print_info "2. 加密方法不匹配"
        print_info "3. 客户端配置问题"
    else
        print_success "所有测试通过！代理工作正常"
    fi
    
    echo ""
    show_connection_info
}

# 运行主函数
main "$@" 