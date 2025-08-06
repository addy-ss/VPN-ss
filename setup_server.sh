#!/bin/bash

# 服务器SSH配置和代码拉取脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 打印函数
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

# 检查是否为root用户
check_root() {
    if [[ $EUID -eq 0 ]]; then
        print_warning "Running as root. Consider using a regular user for security."
    fi
}

# 安装必要的软件包
install_packages() {
    print_info "Installing necessary packages..."
    
    if command -v apt-get &> /dev/null; then
        # Ubuntu/Debian
        sudo apt-get update
        sudo apt-get install -y git curl wget
    elif command -v yum &> /dev/null; then
        # CentOS/RHEL
        sudo yum update -y
        sudo yum install -y git curl wget
    elif command -v dnf &> /dev/null; then
        # Fedora
        sudo dnf update -y
        sudo dnf install -y git curl wget
    else
        print_error "Unsupported package manager"
        exit 1
    fi
    
    print_success "Packages installed successfully"
}

# 检查SSH密钥
check_ssh_keys() {
    print_info "Checking SSH keys..."
    
    if [[ -f ~/.ssh/id_rsa && -f ~/.ssh/id_rsa.pub ]]; then
        print_success "SSH keys already exist"
        return 0
    else
        print_warning "SSH keys not found"
        return 1
    fi
}

# 生成SSH密钥
generate_ssh_keys() {
    print_info "Generating SSH keys..."
    
    # 创建.ssh目录
    mkdir -p ~/.ssh
    chmod 700 ~/.ssh
    
    # 生成SSH密钥
    ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa -N "" -C "server@$(hostname)"
    
    print_success "SSH keys generated successfully"
}

# 启动SSH代理
setup_ssh_agent() {
    print_info "Setting up SSH agent..."
    
    # 启动SSH代理
    eval "$(ssh-agent -s)"
    
    # 添加SSH密钥
    ssh-add ~/.ssh/id_rsa
    
    print_success "SSH agent configured"
}

# 显示公钥
show_public_key() {
    print_info "Your SSH public key:"
    echo "=========================================="
    cat ~/.ssh/id_rsa.pub
    echo "=========================================="
    echo ""
    print_warning "Please add this public key to your GitHub account:"
    print_info "1. Go to https://github.com/settings/keys"
    print_info "2. Click 'New SSH key'"
    print_info "3. Paste the key above"
    print_info "4. Click 'Add SSH key'"
    echo ""
    read -p "Press Enter after adding the key to GitHub..."
}

# 测试GitHub连接
test_github_connection() {
    print_info "Testing GitHub SSH connection..."
    
    if ssh -T git@github.com 2>&1 | grep -q "successfully authenticated"; then
        print_success "GitHub SSH connection successful"
        return 0
    else
        print_error "GitHub SSH connection failed"
        return 1
    fi
}

# 配置Git
setup_git() {
    print_info "Configuring Git..."
    
    # 设置Git用户信息
    read -p "Enter your Git username: " git_username
    read -p "Enter your Git email: " git_email
    
    git config --global user.name "$git_username"
    git config --global user.email "$git_email"
    
    print_success "Git configured successfully"
}

# 克隆代码
clone_repository() {
    print_info "Cloning VPN repository..."
    
    # 创建项目目录
    sudo mkdir -p /opt
    sudo chown $USER:$USER /opt
    
    # 进入目录
    cd /opt
    
    # 克隆仓库
    if [[ -d "VPN-ss" ]]; then
        print_warning "Repository already exists, updating..."
        cd VPN-ss
        git pull origin main
    else
        git clone git@github.com:addy-ss/VPN-ss.git
        cd VPN-ss
    fi
    
    print_success "Repository cloned/updated successfully"
}

# 检查Go环境
check_go_environment() {
    print_info "Checking Go environment..."
    
    if command -v go &> /dev/null; then
        go_version=$(go version | awk '{print $3}')
        print_success "Go installed: $go_version"
        
        # 检查Go版本
        if [[ "$go_version" < "go1.21" ]]; then
            print_warning "Go version is older than 1.21, consider updating"
        fi
    else
        print_error "Go is not installed"
        print_info "Installing Go..."
        install_go
    fi
}

# 安装Go
install_go() {
    print_info "Installing Go..."
    
    # 下载最新版本的Go
    GO_VERSION="1.21.5"
    GO_ARCH="linux-amd64"
    
    cd /tmp
    wget https://golang.org/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz
    sudo tar -C /usr/local -xzf go${GO_VERSION}.${GO_ARCH}.tar.gz
    
    # 设置环境变量
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    source ~/.bashrc
    
    print_success "Go installed successfully"
}

# 构建项目
build_project() {
    print_info "Building VPN project..."
    
    cd /opt/VPN-ss
    
    # 下载依赖
    go mod tidy
    
    # 构建项目
    go build -o vps cmd/main.go
    
    if [[ -f "vps" ]]; then
        print_success "Project built successfully"
    else
        print_error "Build failed"
        exit 1
    fi
}

# 运行测试
run_tests() {
    print_info "Running tests..."
    
    cd /opt/VPN-ss
    
    if go test ./...; then
        print_success "All tests passed"
    else
        print_warning "Some tests failed"
    fi
}

# 显示项目信息
show_project_info() {
    print_info "Project information:"
    echo "=========================================="
    echo "Repository: https://github.com/addy-ss/VPN-ss"
    echo "Location: /opt/VPN-ss"
    echo "Version: v1.0.0"
    echo ""
    echo "Available commands:"
    echo "  cd /opt/VPN-ss"
    echo "  ./vps                    # Run the service"
    echo "  ./scripts/start.sh start # Start with script"
    echo "  make run                 # Run with Makefile"
    echo "=========================================="
}

# 主函数
main() {
    print_info "Starting server setup for VPN project..."
    echo ""
    
    # 检查root权限
    check_root
    
    # 安装软件包
    install_packages
    
    # 检查/生成SSH密钥
    if ! check_ssh_keys; then
        generate_ssh_keys
    fi
    
    # 设置SSH代理
    setup_ssh_agent
    
    # 显示公钥
    show_public_key
    
    # 测试GitHub连接
    if ! test_github_connection; then
        print_error "GitHub SSH connection failed. Please check your SSH key configuration."
        exit 1
    fi
    
    # 配置Git
    setup_git
    
    # 检查Go环境
    check_go_environment
    
    # 克隆代码
    clone_repository
    
    # 构建项目
    build_project
    
    # 运行测试
    run_tests
    
    # 显示项目信息
    show_project_info
    
    print_success "Server setup completed successfully!"
    echo ""
    print_info "Next steps:"
    print_info "1. Configure the service: cp config.example.yaml config.yaml"
    print_info "2. Edit configuration: nano config.yaml"
    print_info "3. Start the service: ./vps"
}

# 显示帮助
show_help() {
    echo "Server Setup Script for VPN Project"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -h, --help    Show this help message"
    echo "  --skip-ssh    Skip SSH key generation (if already configured)"
    echo "  --skip-go     Skip Go installation (if already installed)"
    echo ""
    echo "This script will:"
    echo "1. Install necessary packages"
    echo "2. Generate SSH keys (if needed)"
    echo "3. Configure GitHub access"
    echo "4. Clone the VPN repository"
    echo "5. Build the project"
    echo "6. Run tests"
}

# 解析参数
case "${1:-}" in
    -h|--help)
        show_help
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac 