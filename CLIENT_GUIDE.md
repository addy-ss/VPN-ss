# VPS 客户端使用指南

## 概述

VPS客户端已经成功修改为完全基于配置文件运行，不再支持命令行参数。这样可以确保配置的一致性和安全性。

## 主要改进

### 1. 配置文件驱动
- ✅ 移除了所有命令行参数
- ✅ 所有配置都通过 `client/config.yaml` 文件管理
- ✅ 支持默认值，即使配置文件不完整也能正常工作

### 2. 简化的启动方式
- ✅ 直接运行 `./vps-client` 即可启动
- ✅ 提供了便捷的启动脚本 `./start.sh`
- ✅ 自动检查配置文件和程序文件

### 3. 增强的管理功能
- ✅ 启动脚本支持多种操作：start, stop, restart, status, test
- ✅ 自动PID管理
- ✅ 连接测试功能

## 使用方法

### 1. 基本启动
```bash
cd client
./vps-client
```

### 2. 使用启动脚本（推荐）
```bash
cd client

# 启动客户端
./start.sh start

# 检查状态
./start.sh status

# 测试连接
./start.sh test

# 停止客户端
./start.sh stop

# 重启客户端
./start.sh restart
```

### 3. 配置文件修改
编辑 `client/config.yaml` 文件：

```yaml
# 服务器配置
server:
  host: "127.0.0.1"  # 修改为你的服务器地址
  port: 8388

# 本地配置
local:
  port: 1080

# 安全配置
security:
  password: "13687401432Fan!"  # 修改为你的密码
  method: "aes-256-gcm"
  timeout: 300

# 日志配置
log:
  level: "info"
  file: ""
```

## 测试验证

### 1. 使用测试脚本
```bash
# 在主目录运行
./test_connection.sh
```

### 2. 手动测试
```bash
# 测试SOCKS5代理
curl --socks5 127.0.0.1:1080 --connect-timeout 10 -s http://httpbin.org/ip
```

### 3. 使用客户端启动脚本测试
```bash
cd client
./start.sh test
```

## 故障排除

### 1. 配置文件问题
```bash
# 检查配置文件是否存在
ls -la client/config.yaml

# 检查配置文件语法
cd client
./start.sh start
```

### 2. 端口占用
```bash
# 检查端口占用
lsof -i :1080

# 停止占用进程
pkill -f vps-client
```

### 3. 连接问题
```bash
# 检查服务端是否运行
lsof -i :8388

# 检查客户端是否运行
cd client
./start.sh status
```

## 安全注意事项

1. **密码安全**：确保使用强密码，不要在公共场合暴露
2. **配置文件权限**：确保配置文件只有所有者可读
3. **网络安全**：建议在可信网络中使用

## 日志说明

客户端会输出详细的连接日志，包括：
- 连接建立过程
- SOCKS5握手
- 目标地址解析
- 加密/解密过程
- 错误信息

日志级别可以通过配置文件中的 `log.level` 字段调整。

## 支持的加密方法

- `aes-256-gcm` (推荐，安全性最高)
- `chacha20-poly1305` (性能较好)

## 性能优化

1. **连接复用**：客户端会自动处理连接复用
2. **超时设置**：可以通过 `security.timeout` 调整超时时间
3. **缓冲区优化**：使用8192字节的缓冲区提高传输效率

## 总结

现在客户端已经完全基于配置文件运行，提供了：
- ✅ 简化的启动方式
- ✅ 统一的配置管理
- ✅ 便捷的管理脚本
- ✅ 完整的测试功能
- ✅ 详细的日志记录

所有功能都经过测试验证，可以安全使用。