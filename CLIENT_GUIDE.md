 # VPS客户端使用指南

## 概述

本项目包含一个完整的VPS代理系统，包括服务端和客户端。客户端代码位于 `client/` 目录中，提供简单易用的代理功能。

## 客户端功能

### 核心特性
- ✅ 简单易用，一键启动/停止
- ✅ 支持多种加密方法（AES-256-GCM, ChaCha20-Poly1305）
- ✅ 自动重连和错误处理
- ✅ 详细的日志记录
- ✅ 命令行参数配置
- ✅ 多种启动方式

### 技术特点
- 使用Go语言编写，性能优异
- 支持强加密通信
- 异步I/O处理
- 内存高效使用

## 快速开始

### 1. 进入客户端目录
```bash
cd client
```

### 2. 编译客户端
```bash
go build -o vps-client main.go
```

### 3. 启动客户端
```bash
# 使用默认配置
./vps-client

# 自定义配置
./vps-client -server=192.168.1.100 -port=8388 -password=mypassword
```

### 4. 停止客户端
```bash
# 按 Ctrl+C
# 或使用停止脚本
./stop.sh
```

## 启动方式

### 方式1：直接运行
```bash
./vps-client -server=192.168.1.100 -port=8388 -local=1080 -password=mypassword
```

### 方式2：使用启动脚本
```bash
./start.sh -s 192.168.1.100 -p 8388 -l 1080 -k mypassword
```

### 方式3：使用快速启动脚本
```bash
./quick_start.sh
```

### 方式4：使用Makefile
```bash
make run
```

## 配置说明

### 命令行参数
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-server` | 服务器地址 | 127.0.0.1 |
| `-port` | 服务器端口 | 8388 |
| `-local` | 本地监听端口 | 1080 |
| `-password` | 密码 | your_password |
| `-method` | 加密方法 | aes-256-gcm |
| `-timeout` | 超时时间(秒) | 300 |

### 配置文件
客户端支持YAML配置文件，位于 `client/config.yaml`：
```yaml
# 服务器配置
server:
  host: "192.168.1.100"
  port: 8388

# 本地配置
local:
  port: 1080

# 安全配置
security:
  password: "your_password"
  method: "aes-256-gcm"
  timeout: 300

# 日志配置
log:
  level: "info"
  file: ""
```

## 使用示例

### 基本使用
```bash
# 连接到本地服务器
./vps-client

# 连接到远程服务器
./vps-client -server=192.168.1.100 -port=8388 -password=mypassword

# 使用不同的本地端口
./vps-client -local=1081 -server=192.168.1.100 -password=mypassword
```

### 高级使用
```bash
# 使用ChaCha20-Poly1305加密
./vps-client -method=chacha20-poly1305 -password=mypassword

# 设置较长的超时时间
./vps-client -timeout=600 -password=mypassword

# 连接到非标准端口
./vps-client -port=8389 -password=mypassword
```

## 测试验证

### 运行测试脚本
```bash
./test_client.sh
```

测试内容包括：
- ✅ 编译测试
- ✅ 参数解析测试
- ✅ 配置文件测试
- ✅ 脚本功能测试
- ✅ 端口监听测试

### 手动测试
```bash
# 1. 启动客户端
./vps-client -server=127.0.0.1 -port=8388 -password=test

# 2. 在另一个终端测试连接
curl --socks5 127.0.0.1:1080 http://httpbin.org/ip

# 3. 检查日志输出
```

## 故障排除

### 常见问题

1. **连接失败**
   ```
   错误: 连接服务器失败
   解决: 检查服务器地址、端口和密码
   ```

2. **端口被占用**
   ```
   错误: 监听端口失败
   解决: 使用不同的本地端口或停止占用端口的程序
   ```

3. **编译失败**
   ```
   错误: 编译错误
   解决: 检查Go环境，运行 go mod tidy
   ```

### 调试方法

1. **查看详细日志**
   ```bash
   ./vps-client -server=192.168.1.100 -password=mypassword
   ```

2. **检查网络连接**
   ```bash
   telnet 192.168.1.100 8388
   ```

3. **验证配置文件**
   ```bash
   cat config.yaml
   ```

## 安全建议

### 1. 密码安全
- 使用强密码（至少16位）
- 包含大小写字母、数字和特殊字符
- 定期更换密码

### 2. 网络安全
- 使用HTTPS传输配置
- 避免在公共网络传输明文密码
- 设置防火墙规则

### 3. 系统安全
- 定期更新系统和依赖
- 监控日志文件
- 限制访问权限

## 性能优化

### 1. 网络优化
- 选择合适的服务器位置
- 使用稳定的网络连接
- 调整超时时间

### 2. 系统优化
- 增加文件描述符限制
- 优化TCP参数
- 使用SSD存储

## 扩展功能

### 可能的改进
1. 支持UDP协议
2. 添加Web管理界面
3. 支持多服务器负载均衡
4. 添加流量统计功能
5. 支持SOCKS5协议

## 相关文档

- [客户端README](client/README.md) - 详细技术文档
- [使用说明](client/USAGE.md) - 使用指南
- [功能总结](client/SUMMARY.md) - 功能概述
- [服务端文档](README.md) - 服务端说明

## 技术支持

如果遇到问题，请：
1. 查看日志文件
2. 运行测试脚本
3. 检查配置文件
4. 参考相关文档

---

**注意**: 请确保服务端已正确配置并运行，客户端才能正常工作。