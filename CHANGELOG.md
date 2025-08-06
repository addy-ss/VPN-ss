# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of VPS VPN Service
- Shadowsocks protocol support
- RESTful API for VPN management
- Docker support
- Comprehensive documentation
- Security audit and analysis
- Multiple encryption methods support
- Health check endpoints
- Configuration generation API

### Changed
- Improved error handling
- Enhanced logging system
- Better security implementation

### Fixed
- Shadowsocks protocol compatibility issues
- Address type parsing errors
- Connection timeout handling
- Memory leak in long-running connections

## [1.0.0] - 2025-08-06

### Added
- 🚀 Initial release
- 🔒 Shadowsocks VPN server implementation
- 🌐 RESTful API for VPN management
- 🐳 Docker containerization
- 📊 Comprehensive monitoring and logging
- 🛡️ Security features and audit logging
- 🔐 Multiple encryption methods (AES-256-GCM, ChaCha20-Poly1305)
- 📝 Extensive documentation
- 🧪 Testing tools and scripts
- ⚡ High-performance proxy server

### Features
- **VPN Management API**
  - Start/Stop VPN service
  - Get VPN status
  - Generate client configurations
  - List supported encryption methods

- **Security Features**
  - Audit logging
  - Threat detection
  - Access control
  - Data encryption

- **Deployment Options**
  - Docker deployment
  - System service installation
  - Manual deployment
  - Cloud deployment support

- **Monitoring & Logging**
  - Structured JSON logging
  - Health check endpoints
  - Performance monitoring
  - Error tracking

### Technical Details
- **Language**: Go 1.21+
- **Framework**: Gin (HTTP API)
- **Protocol**: Shadowsocks
- **Encryption**: AES-256-GCM, ChaCha20-Poly1305
- **Container**: Docker support
- **Documentation**: Markdown with examples

### Breaking Changes
- None (initial release)

### Deprecated
- None

### Removed
- None

### Fixed
- None (initial release)

### Security
- Implemented secure password handling
- Added audit logging for security events
- Implemented proper encryption/decryption
- Added threat detection mechanisms

## [未发布]

### 修复
- 修复了代理连接中"invalid encrypted data length"错误
  - 将加密数据长度限制从4096字节增加到65535字节
  - 增加了缓冲区大小从4096字节到8192字节
  - 解决了6808字节等大数据包被拒绝的问题

### 技术改进
- 优化了代理服务器的数据处理能力
- 提高了对大型加密数据包的兼容性

---

## Version History

- **v1.0.0** - Initial release with full Shadowsocks support
- **Unreleased** - Future improvements and features

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 