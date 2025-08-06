#!/bin/bash

# GitHub Release Script for VPS VPN Service

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Print functions
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

# Check if git is installed
check_git() {
    if ! command -v git &> /dev/null; then
        print_error "Git is not installed"
        exit 1
    fi
}

# Check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_info "Initializing git repository..."
        git init
        git add .
        git commit -m "Initial commit: VPS VPN Service v1.0.0"
    fi
}

# Clean build artifacts
clean_build() {
    print_info "Cleaning build artifacts..."
    rm -f vps
    rm -f *.log
    rm -f *.pid
    rm -rf build/
    rm -rf dist/
    print_success "Build artifacts cleaned"
}

# Run tests
run_tests() {
    print_info "Running tests..."
    if go test ./...; then
        print_success "All tests passed"
    else
        print_error "Tests failed"
        exit 1
    fi
}

# Build project
build_project() {
    print_info "Building project..."
    if go build -o vps cmd/main.go; then
        print_success "Project built successfully"
    else
        print_error "Build failed"
        exit 1
    fi
}

# Check for sensitive data
check_sensitive_data() {
    print_info "Checking for sensitive data..."
    
    # Check for hardcoded passwords
    if grep -r "password.*=" . --exclude-dir=.git --exclude=*.log; then
        print_warning "Found potential hardcoded passwords"
    fi
    
    # Check for API keys
    if grep -r "api_key\|secret_key\|token" . --exclude-dir=.git --exclude=*.log; then
        print_warning "Found potential API keys or secrets"
    fi
    
    print_success "Sensitive data check completed"
}

# Create release tag
create_tag() {
    local version=$1
    print_info "Creating tag v$version..."
    
    if git tag -l | grep -q "v$version"; then
        print_warning "Tag v$version already exists"
        read -p "Do you want to delete and recreate it? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git tag -d "v$version"
        else
            print_error "Tag creation cancelled"
            exit 1
        fi
    fi
    
    git tag -a "v$version" -m "Release v$version"
    print_success "Tag v$version created"
}

# Push to GitHub
push_to_github() {
    local remote_url=$1
    local version=$2
    
    print_info "Pushing to GitHub..."
    
    # Add remote if not exists
    if ! git remote get-url origin > /dev/null 2>&1; then
        git remote add origin "$remote_url"
    fi
    
    # Push code
    git push -u origin main
    git push origin "v$version"
    
    print_success "Code pushed to GitHub"
}

# Create release notes
create_release_notes() {
    local version=$1
    local notes_file="RELEASE_v${version}.md"
    
    print_info "Creating release notes..."
    
    cat > "$notes_file" << EOF
# Release v$version

## 🚀 What's New

- Initial release of VPS VPN Service
- Shadowsocks protocol support
- RESTful API for VPN management
- Docker containerization
- Comprehensive security features
- Multiple encryption methods support

## 📦 Installation

\`\`\`bash
git clone https://github.com/your-username/vps-vpn-service.git
cd vps-vpn-service
go mod tidy
go run cmd/main.go
\`\`\`

## 🔧 Quick Start

1. Copy configuration:
   \`\`\`bash
   cp config.example.yaml config.yaml
   \`\`\`

2. Edit configuration:
   \`\`\`yaml
   shadowsocks:
     password: "your-secure-password"
   \`\`\`

3. Run the service:
   \`\`\`bash
   go run cmd/main.go
   \`\`\`

## 📖 Documentation

- [Quick Start Guide](QUICK_START.md)
- [Deployment Guide](DEPLOYMENT_GUIDE.md)
- [Security Analysis](SECURITY_ANALYSIS.md)

## 🔒 Security

This release includes comprehensive security features:
- Audit logging
- Threat detection
- Access control
- Data encryption

## 🐳 Docker

\`\`\`bash
docker-compose up -d
\`\`\`

## 📊 API

Health check:
\`\`\`bash
curl http://localhost:8080/api/v1/health
\`\`\`

VPN status:
\`\`\`bash
curl http://localhost:8080/api/v1/vpn/status
\`\`\`

## 🔧 Configuration

See [config.example.yaml](config.example.yaml) for all available options.

## 🧪 Testing

\`\`\`bash
go test ./...
python3 scripts/test_client.py
\`\`\`

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

**Download**: [v$version](https://github.com/your-username/vps-vpn-service/releases/tag/v$version)
EOF

    print_success "Release notes created: $notes_file"
}

# Main function
main() {
    local version=${1:-"1.0.0"}
    local github_url=${2:-""}
    
    print_info "Starting GitHub release process for v$version"
    print_info "GitHub URL: $github_url"
    
    # Pre-release checks
    check_git
    check_git_repo
    clean_build
    run_tests
    build_project
    check_sensitive_data
    
    # Create release
    create_tag "$version"
    create_release_notes "$version"
    
    if [[ -n "$github_url" ]]; then
        push_to_github "$github_url" "$version"
    else
        print_warning "GitHub URL not provided, skipping push"
        print_info "To push to GitHub, run:"
        print_info "git remote add origin <your-github-url>"
        print_info "git push -u origin main"
        print_info "git push origin v$version"
    fi
    
    print_success "Release v$version prepared successfully!"
    print_info "Next steps:"
    print_info "1. Review the release notes: RELEASE_v$version.md"
    print_info "2. Push to GitHub: git push origin v$version"
    print_info "3. Create release on GitHub with the notes"
}

# Show help
show_help() {
    echo "GitHub Release Script for VPS VPN Service"
    echo ""
    echo "Usage: $0 [version] [github-url]"
    echo ""
    echo "Arguments:"
    echo "  version      Version number (default: 1.0.0)"
    echo "  github-url   GitHub repository URL"
    echo ""
    echo "Examples:"
    echo "  $0                    # Release v1.0.0"
    echo "  $0 1.1.0             # Release v1.1.0"
    echo "  $0 1.0.0 https://github.com/user/repo.git"
    echo ""
    echo "Options:"
    echo "  -h, --help    Show this help message"
}

# Parse arguments
case "${1:-}" in
    -h|--help)
        show_help
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac 