#!/bin/bash

# VPS客户端测试脚本

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

# 测试编译
test_build() {
    print_info "测试编译..."
    if go build -o vps-client-test main.go; then
        print_success "编译成功"
        rm -f vps-client-test
    else
        print_error "编译失败"
        exit 1
    fi
}

# 测试参数解析
test_args() {
    print_info "测试参数解析..."
    if ./vps-client --help &>/dev/null; then
        print_success "参数解析正常"
    else
        print_error "参数解析失败"
        exit 1
    fi
}

# 测试端口监听
test_port() {
    print_info "测试端口监听..."
    
    # 检查默认端口是否可用
    if lsof -Pi :1080 -sTCP:LISTEN -t >/dev/null ; then
        print_warning "端口 1080 已被占用，跳过端口测试"
        return
    fi
    
    # 启动客户端（后台运行）
    ./vps-client -server=127.0.0.1 -port=8388 -local=1080 -password=test &
    CLIENT_PID=$!
    
    # 等待启动
    sleep 2
    
    # 检查是否启动成功
    if kill -0 $CLIENT_PID 2>/dev/null; then
        print_success "客户端启动成功"
        
        # 检查端口是否监听
        if lsof -Pi :1080 -sTCP:LISTEN -t >/dev/null ; then
            print_success "端口监听正常"
        else
            print_error "端口监听失败"
        fi
        
        # 停止客户端
        kill $CLIENT_PID
        sleep 1
    else
        print_error "客户端启动失败"
        exit 1
    fi
}

# 测试配置文件
test_config() {
    print_info "测试配置文件..."
    
    if [ -f config.yaml ]; then
        print_success "配置文件存在"
        
        # 检查配置文件格式
        if python3 -c "import yaml; yaml.safe_load(open('config.yaml'))" 2>/dev/null; then
            print_success "配置文件格式正确"
        else
            print_warning "配置文件格式可能有问题"
        fi
    else
        print_warning "配置文件不存在"
    fi
}

# 测试脚本
test_scripts() {
    print_info "测试脚本..."
    
    # 测试启动脚本
    if [ -f start.sh ] && [ -x start.sh ]; then
        print_success "启动脚本存在且可执行"
    else
        print_warning "启动脚本不存在或不可执行"
    fi
    
    # 测试停止脚本
    if [ -f stop.sh ] && [ -x stop.sh ]; then
        print_success "停止脚本存在且可执行"
    else
        print_warning "停止脚本不存在或不可执行"
    fi
    
    # 测试快速启动脚本
    if [ -f quick_start.sh ] && [ -x quick_start.sh ]; then
        print_success "快速启动脚本存在且可执行"
    else
        print_warning "快速启动脚本不存在或不可执行"
    fi
}

# 主测试函数
main() {
    echo "开始测试VPS客户端..."
    echo ""
    
    test_build
    echo ""
    
    test_args
    echo ""
    
    test_config
    echo ""
    
    test_scripts
    echo ""
    
    test_port
    echo ""
    
    print_success "所有测试完成！"
    echo ""
    echo "客户端已准备就绪，可以使用以下命令启动："
    echo "  ./vps-client"
    echo "  ./start.sh"
    echo "  ./quick_start.sh"
}

# 运行测试
main "$@"