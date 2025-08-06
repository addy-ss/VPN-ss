#!/bin/bash

# VPS VPN Service Linux 部署脚本
# 适用于 Ubuntu/Debian/CentOS 等 Linux 发行版

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否为root用户
check_root() {
    if [[ $EUID -eq 0 ]]; then
        log_warn "检测到root用户，建议使用普通用户运行此脚本"
        read -p "是否继续？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# 检测操作系统
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$NAME
        VER=$VERSION_ID
    elif type lsb_release >/dev/null 2>&1; then
        OS=$(lsb_release -si)
        VER=$(lsb_release -sr)
    else
        OS=$(uname -s)
        VER=$(uname -r)
    fi
    log_info "检测到操作系统: $OS $VER"
}

# 安装Docker
install_docker() {
    log_info "检查Docker安装状态..."
    if command -v docker &> /dev/null; then
        log_info "Docker已安装"
        return 0
    fi
    
    log_info "安装Docker..."
    
    # 卸载旧版本
    sudo apt-get remove docker docker-engine docker.io containerd runc 2>/dev/null || true
    
    # 安装依赖
    sudo apt-get update
    sudo apt-get install -y \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg \
        lsb-release
    
    # 添加Docker官方GPG密钥
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    
    # 设置稳定版仓库
    echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
        $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    
    # 安装Docker Engine
    sudo apt-get update
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io
    
    # 启动Docker服务
    sudo systemctl start docker
    sudo systemctl enable docker
    
    # 将当前用户添加到docker组
    sudo usermod -aG docker $USER
    
    log_info "Docker安装完成"
}

# 安装Docker Compose
install_docker_compose() {
    log_info "检查Docker Compose安装状态..."
    if command -v docker-compose &> /dev/null; then
        log_info "Docker Compose已安装"
        return 0
    fi
    
    log_info "安装Docker Compose..."
    
    # 下载Docker Compose
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    
    # 设置执行权限
    sudo chmod +x /usr/local/bin/docker-compose
    
    log_info "Docker Compose安装完成"
}

# 配置防火墙
setup_firewall() {
    log_info "配置防火墙..."
    
    # 检查ufw是否安装
    if command -v ufw &> /dev/null; then
        # 允许SSH
        sudo ufw allow ssh
        # 允许HTTP API端口
        sudo ufw allow 8080
        # 允许Shadowsocks端口
        sudo ufw allow 8388
        # 启用防火墙
        sudo ufw --force enable
        log_info "UFW防火墙配置完成"
    else
        log_warn "未检测到UFW，请手动配置防火墙规则"
        log_info "需要开放的端口: 22(SSH), 8080(HTTP API), 8388(Shadowsocks)"
    fi
}

# 创建配置文件
create_config() {
    log_info "创建配置文件..."
    
    if [[ ! -f config.yaml ]]; then
        if [[ -f config.example.yaml ]]; then
            cp config.example.yaml config.yaml
            log_warn "已创建config.yaml，请修改密码和配置"
        else
            log_error "未找到config.example.yaml文件"
            exit 1
        fi
    else
        log_info "config.yaml已存在"
    fi
}

# 创建日志目录
create_logs_dir() {
    log_info "创建日志目录..."
    mkdir -p logs
    chmod 755 logs
}

# 启动服务
start_service() {
    log_info "启动VPS VPN服务..."
    
    # 构建并启动容器
    docker-compose up -d --build
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 10
    
    # 检查服务状态
    if docker-compose ps | grep -q "Up"; then
        log_info "服务启动成功！"
        log_info "HTTP API地址: http://$(hostname -I | awk '{print $1}'):8080"
        log_info "Shadowsocks端口: 8388"
    else
        log_error "服务启动失败，请检查日志"
        docker-compose logs
        exit 1
    fi
}

# 创建系统服务
create_systemd_service() {
    log_info "创建系统服务..."
    
    sudo tee /etc/systemd/system/vps-vpn.service > /dev/null <<EOF
[Unit]
Description=VPS VPN Service
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=$(pwd)
ExecStart=/usr/local/bin/docker-compose up -d
ExecStop=/usr/local/bin/docker-compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

    # 重新加载systemd
    sudo systemctl daemon-reload
    
    # 启用服务
    sudo systemctl enable vps-vpn.service
    
    log_info "系统服务创建完成"
    log_info "使用以下命令管理服务:"
    log_info "  启动: sudo systemctl start vps-vpn"
    log_info "  停止: sudo systemctl stop vps-vpn"
    log_info "  状态: sudo systemctl status vps-vpn"
    log_info "  重启: sudo systemctl restart vps-vpn"
}

# 显示部署信息
show_deployment_info() {
    log_info "=== 部署完成 ==="
    log_info "服务信息:"
    log_info "  - HTTP API: http://$(hostname -I | awk '{print $1}'):8080"
    log_info "  - Shadowsocks端口: 8388"
    log_info "  - 配置文件: $(pwd)/config.yaml"
    log_info "  - 日志目录: $(pwd)/logs"
    
    log_info "管理命令:"
    log_info "  - 查看状态: docker-compose ps"
    log_info "  - 查看日志: docker-compose logs -f"
    log_info "  - 重启服务: docker-compose restart"
    log_info "  - 停止服务: docker-compose down"
    
    log_info "API测试:"
    log_info "  - 健康检查: curl http://localhost:8080/api/v1/health"
    log_info "  - VPN状态: curl http://localhost:8080/api/v1/vpn/status"
}

# 主函数
main() {
    log_info "开始部署VPS VPN服务..."
    
    check_root
    detect_os
    install_docker
    install_docker_compose
    setup_firewall
    create_config
    create_logs_dir
    start_service
    create_systemd_service
    show_deployment_info
    
    log_info "部署完成！"
}

# 运行主函数
main "$@" 