# 🚀 VPS VPN Service

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](docker-compose.yml)
[![Security](https://img.shields.io/badge/Security-Audited-green.svg)](SECURITY_ANALYSIS.md)

A high-performance VPN service built with Go, featuring Shadowsocks protocol support, RESTful API management, comprehensive security features, and **multi-server proxy forwarding** for enhanced privacy and security.

## ✨ Features

- 🔒 **Shadowsocks Protocol** - Full compatibility with Shadowsocks clients
- 🌐 **RESTful API** - Easy management through HTTP endpoints
- 🛡️ **Security First** - Audit logging, threat detection, and encryption
- 🐳 **Docker Ready** - Containerized deployment
- 📊 **Monitoring** - Comprehensive logging and health checks
- ⚡ **High Performance** - Optimized for high-throughput connections
- 🔐 **Multiple Encryption** - AES-256-GCM, ChaCha20-Poly1305 support
- 📱 **Client Support** - Works with all major Shadowsocks clients
- 🔄 **Multi-Server Proxy** - Forward requests through multiple servers for enhanced security
- 🛡️ **Layered Security** - Multiple encryption layers and traffic obfuscation

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- Docker (optional)
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/your-username/vps-vpn-service.git
   cd vps-vpn-service
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Configure the service**
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your settings
   ```

4. **Run the service**
   ```bash
   go run cmd/main.go
   ```

### Docker Deployment

```bash
# Build and run with Docker Compose
docker-compose up -d

# Or build manually
docker build -t vps-vpn .
docker run -d -p 8080:8080 -p 8388:8388 vps-vpn
```

## 📖 Documentation

- [Quick Start Guide](QUICK_START.md) - Get up and running quickly
- [Deployment Guide](DEPLOYMENT_GUIDE.md) - Detailed deployment instructions
- [Security Analysis](SECURITY_ANALYSIS.md) - Security features and analysis
- [Project Summary](PROJECT_SUMMARY.md) - Technical overview
- [Demo Guide](DEMO.md) - Usage examples and demonstrations
- [Multi-Server Proxy Guide](MULTI_SERVER_GUIDE.md) - Multi-level proxy configuration and usage

## 🔧 API Reference

### Health Check
```bash
curl http://localhost:8080/api/v1/health
```

### VPN Management
```bash
# Get VPN status
curl http://localhost:8080/api/v1/vpn/status

# Start VPN service (single server)
curl -X POST http://localhost:8080/api/v1/vpn/start \
  -H "Content-Type: application/json" \
  -d '{
    "port": 8388,
    "method": "aes-256-gcm",
    "password": "your-password",
    "timeout": 300
  }'

# Start VPN service with multi-server proxy
curl -X POST http://localhost:8080/api/v1/vpn/start \
  -H "Content-Type: application/json" \
  -d '{
    "port": 8388,
    "method": "aes-256-gcm",
    "password": "your-password",
    "timeout": 300,
    "second_server_enabled": true,
    "second_server_host": "192.168.1.100",
    "second_server_port": 8389,
    "second_server_method": "aes-256-gcm",
    "second_server_password": "second-server-password",
    "second_server_timeout": 300
  }'

# Generate client configuration
curl -X POST http://localhost:8080/api/v1/vpn/config/generate \
  -H "Content-Type: application/json" \
  -d '{
    "port": 8388,
    "method": "aes-256-gcm",
    "password": "your-password"
  }'
```

## 🛡️ Security Features

- **Audit Logging** - Complete operation tracking
- **Threat Detection** - Suspicious activity monitoring
- **Access Control** - IP whitelisting and rate limiting
- **Data Encryption** - End-to-end encryption
- **Security Analysis** - Comprehensive security review

## 📊 Monitoring

### Health Checks
```bash
# API health
curl http://localhost:8080/api/v1/health

# VPN status
curl http://localhost:8080/api/v1/vpn/status
```

### Logs
```bash
# View application logs
tail -f vps.log

# Docker logs
docker-compose logs -f
```

## 🔧 Configuration

### Basic Configuration (config.yaml)
```yaml
server:
  port: 8080
  host: "0.0.0.0"
  mode: "debug"

# Multi-server proxy configuration (optional)
second_server:
  enabled: false   # Enable second server forwarding
  host: "192.168.1.100"  # Second server address
  port: 8389       # Second server port
  method: "aes-256-gcm"  # Encryption method
  password: "second-server-password"  # Second server password
  timeout: 300     # Timeout in seconds

shadowsocks:
  enabled: true
  method: "aes-256-gcm"
  password: "your-secure-password"
  port: 8388
  timeout: 300

log:
  level: "info"
```

## 🧪 Testing

### Run Tests
```bash
go test ./...
```

### Test Client
```bash
python3 scripts/test_client.py
```

## 📦 Deployment Options

### 1. Direct Deployment
```bash
go build -o vps cmd/main.go
./vps
```

### 2. Docker Deployment
```bash
docker-compose up -d
```

### 3. System Service
```bash
sudo make install
sudo systemctl enable vps-vpn
sudo systemctl start vps-vpn
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup
```bash
# Fork and clone
git clone https://github.com/your-username/vps-vpn-service.git
cd vps-vpn-service

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build
go build -o vps cmd/main.go
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📖 [Documentation](QUICK_START.md)
- 🐛 [Report Issues](https://github.com/your-username/vps-vpn-service/issues)
- 💬 [Discussions](https://github.com/your-username/vps-vpn-service/discussions)
- 📧 [Contact](mailto:your-email@example.com)

## 🙏 Acknowledgments

- [Gin Framework](https://github.com/gin-gonic/gin) - HTTP web framework
- [Shadowsocks Protocol](https://shadowsocks.org/) - VPN protocol
- [Go Community](https://golang.org/) - Go programming language

---

⭐ **Star this repository if you find it useful!** 