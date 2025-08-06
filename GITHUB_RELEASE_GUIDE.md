# 🚀 GitHub Release Guide

This guide will help you publish your VPS VPN Service project to GitHub.

## 📋 Prerequisites

1. **GitHub Account** - Create one at [github.com](https://github.com)
2. **Git** - Install Git on your system
3. **Go** - Version 1.21 or higher
4. **Docker** (optional) - For containerized deployment

## 🔧 Preparation Steps

### 1. Clean Up Project

Before publishing, ensure your project is clean:

```bash
# Remove sensitive files
rm -f config.yaml
rm -f *.log
rm -f *.pid
rm -f vps

# Check .gitignore
cat .gitignore
```

### 2. Test Everything

```bash
# Run tests
go test ./...

# Build project
go build -o vps cmd/main.go

# Test the build
./vps --help
```

### 3. Update Documentation

- ✅ README.md - Updated with badges and clear instructions
- ✅ LICENSE - MIT License added
- ✅ CONTRIBUTING.md - Contribution guidelines
- ✅ CHANGELOG.md - Release history
- ✅ .gitignore - Proper exclusions

## 🚀 Publishing to GitHub

### Method 1: Using the Release Script

```bash
# Make script executable
chmod +x scripts/github_release.sh

# Run release script
./scripts/github_release.sh 1.0.0 https://github.com/your-username/vps-vpn-service.git
```

### Method 2: Manual Steps

#### Step 1: Initialize Git Repository

```bash
# Initialize git (if not already done)
git init

# Add all files
git add .

# Initial commit
git commit -m "Initial commit: VPS VPN Service v1.0.0"
```

#### Step 2: Create GitHub Repository

1. Go to [github.com](https://github.com)
2. Click "New repository"
3. Name it: `vps-vpn-service`
4. Make it **Public** or **Private**
5. **Don't** initialize with README (we already have one)
6. Click "Create repository"

#### Step 3: Connect and Push

```bash
# Add remote origin
git remote add origin https://github.com/your-username/vps-vpn-service.git

# Push to GitHub
git push -u origin main

# Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

## 🏷️ Creating a Release

### 1. Go to GitHub Repository

Visit: `https://github.com/your-username/vps-vpn-service`

### 2. Create Release

1. Click "Releases" in the right sidebar
2. Click "Create a new release"
3. Select tag: `v1.0.0`
4. Title: `Release v1.0.0`
5. Description: Copy from `RELEASE_v1.0.0.md`

### 3. Release Notes Template

```markdown
# Release v1.0.0

## 🚀 What's New

- Initial release of VPS VPN Service
- Shadowsocks protocol support
- RESTful API for VPN management
- Docker containerization
- Comprehensive security features
- Multiple encryption methods support

## 📦 Installation

```bash
git clone https://github.com/your-username/vps-vpn-service.git
cd vps-vpn-service
go mod tidy
go run cmd/main.go
```

## 🔧 Quick Start

1. Copy configuration:
   ```bash
   cp config.example.yaml config.yaml
   ```

2. Edit configuration:
   ```yaml
   shadowsocks:
     password: "your-secure-password"
   ```

3. Run the service:
   ```bash
   go run cmd/main.go
   ```

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

```bash
docker-compose up -d
```

## 📊 API

Health check:
```bash
curl http://localhost:8080/api/v1/health
```

VPN status:
```bash
curl http://localhost:8080/api/v1/vpn/status
```

## 🔧 Configuration

See [config.example.yaml](config.example.yaml) for all available options.

## 🧪 Testing

```bash
go test ./...
python3 scripts/test_client.py
```

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.
```

## 📊 Repository Features

### 1. Repository Settings

Enable these features in your GitHub repository:

- ✅ **Issues** - For bug reports and feature requests
- ✅ **Discussions** - For community discussions
- ✅ **Wiki** - For additional documentation
- ✅ **Actions** - For CI/CD (optional)

### 2. Repository Topics

Add these topics to your repository:

```
vpn
shadowsocks
go
golang
proxy
security
docker
api
rest
vps
```

### 3. Repository Description

```
A high-performance VPN service built with Go, featuring Shadowsocks protocol support, RESTful API management, and comprehensive security features.
```

## 🔧 GitHub Actions (Optional)

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Test
      run: go test -v ./...
    
    - name: Build
      run: go build -o vps cmd/main.go
```

## 📈 Promoting Your Project

### 1. Social Media

Share your project on:
- Twitter/X
- Reddit (r/golang, r/vpn)
- Hacker News
- Dev.to

### 2. GitHub Community

- Add to GitHub Topics
- Create Discussions
- Respond to Issues
- Update Documentation

### 3. Documentation Sites

- Add to Awesome Go list
- Submit to Go.dev
- Share on Go forums

## 🔒 Security Considerations

### Before Publishing

- ✅ No hardcoded passwords
- ✅ No API keys in code
- ✅ No sensitive data in commits
- ✅ Proper .gitignore
- ✅ Security analysis completed

### After Publishing

- Monitor for security issues
- Respond to vulnerability reports
- Keep dependencies updated
- Regular security audits

## 📝 Maintenance

### Regular Tasks

1. **Update Dependencies**
   ```bash
   go get -u ./...
   go mod tidy
   ```

2. **Run Tests**
   ```bash
   go test ./...
   ```

3. **Update Documentation**
   - Keep README current
   - Update CHANGELOG
   - Review security docs

4. **Monitor Issues**
   - Respond to bug reports
   - Review feature requests
   - Address security concerns

## 🎉 Success Metrics

Track these metrics for your project:

- ⭐ Stars
- 🔄 Forks
- 📥 Downloads
- 🐛 Issues resolved
- 📈 Contributors
- 📊 Traffic analytics

## 🆘 Troubleshooting

### Common Issues

1. **Build Fails**
   ```bash
   go clean -modcache
   go mod tidy
   go build -o vps cmd/main.go
   ```

2. **Tests Fail**
   ```bash
   go test -v ./...
   ```

3. **Git Issues**
   ```bash
   git status
   git add .
   git commit -m "Fix: description"
   ```

4. **GitHub Issues**
   - Check repository permissions
   - Verify remote URL
   - Ensure SSH keys are set up

## 📞 Support

If you encounter issues:

1. Check existing issues on GitHub
2. Create a new issue with details
3. Join discussions
4. Review documentation

---

**Happy Publishing! 🚀**

Your VPS VPN Service is now ready for the world to see! 