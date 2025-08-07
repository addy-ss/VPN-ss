# VPS VPN 服务启动指南

## 概述

VPS服务现在支持同时启动HTTP API服务和Shadowsocks VPN服务。当使用 `./scripts/start.sh start` 启动服务时，两个服务会一起启动。

## 服务架构

### 1. HTTP API服务
- **端口**: 8080
- **功能**: 提供RESTful API接口
- **健康检查**: `http://localhost:8080/api/v1/health`

### 2. Shadowsocks VPN服务
- **端口**: 8388
- **功能**: 提供SOCKS5代理服务
- **加密方法**: AES-256-GCM
- **密码**: 13687401432Fan!

## 启动方式

### 使用启动脚本（推荐）
```bash
# 启动服务（包含HTTP API和VPN）
./scripts/start.sh start

# 查看服务状态
./scripts/start.sh status

# 停止服务
./scripts/start.sh stop

# 重启服务
./scripts/start.sh restart
```

### 直接启动
```bash
# 构建项目
go build -o vps cmd/main.go

# 启动服务
./vps
```

## 配置说明

### 配置文件: `config.yaml`
```yaml
server:
  port: 8080
  host: "0.0.0.0"
  mode: "debug"

shadowsocks:
  enabled: true          # 启用VPN服务
  method: "aes-256-gcm"  # 加密方法
  password: "13687401432Fan!"  # 密码
  port: 8388             # VPN端口
  timeout: 300           # 超时时间

log:
  level: "info"
  file: ""
```

### 关键配置项
- `shadowsocks.enabled`: 控制是否启动VPN服务
- `shadowsocks.port`: VPN服务端口
- `shadowsocks.password`: VPN连接密码
- `shadowsocks.method`: 加密方法

## 启动流程

### 1. 服务启动顺序
1. 加载配置文件
2. 初始化HTTP API服务
3. 启动Shadowsocks VPN服务（如果启用）
4. 启动HTTP服务器
5. 等待中断信号

### 2. 启动检查
启动脚本会检查：
- ✅ Go环境
- ✅ 项目依赖
- ✅ 配置文件
- ✅ HTTP API服务响应
- ✅ VPN端口监听状态

### 3. 状态监控
```bash
# 检查服务状态
./scripts/start.sh status

# 输出示例:
[SUCCESS] 服务正在运行 (PID: 71657)
[SUCCESS] API服务正常
[SUCCESS] Shadowsocks端口正常 (8388)
[SUCCESS] HTTP端口正常 (8080)
```

## 服务验证

### 1. HTTP API测试
```bash
# 健康检查
curl http://localhost:8080/api/v1/health

# 获取VPN状态
curl http://localhost:8080/api/v1/vpn/status

# 获取支持的加密方法
curl http://localhost:8080/api/v1/vpn/methods
```

### 2. VPN服务测试
```bash
# 检查端口监听
lsof -i :8388

# 使用客户端测试
cd client
./start.sh test
```

### 3. 完整连接测试
```bash
# 运行完整测试脚本
./test_connection.sh
```

## 日志信息

### 启动日志示例
```
[INFO] 启动VPS VPN服务...
[SUCCESS] 服务已启动，PID: 71657
[INFO] HTTP API: http://localhost:8080
[INFO] Shadowsocks端口: 8388
[SUCCESS] HTTP API服务启动成功！
[SUCCESS] Shadowsocks VPN服务启动成功！
[INFO] VPN配置信息:
[INFO]   端口: 8388
[INFO]   加密方法: aes-256-gcm
[INFO]   密码: 13687401432Fan!
[INFO] 服务启动完成！
```

### 服务端日志
- HTTP API请求日志
- VPN连接日志
- 错误和警告信息

## 故障排除

### 1. 服务启动失败
```bash
# 检查配置文件
cat config.yaml

# 检查端口占用
lsof -i :8080
lsof -i :8388

# 查看详细日志
./vps
```

### 2. VPN服务未启动
```bash
# 检查配置
grep -A 5 "shadowsocks:" config.yaml

# 确保 enabled: true
```

### 3. 端口冲突
```bash
# 停止占用进程
sudo lsof -ti:8080 | xargs kill -9
sudo lsof -ti:8388 | xargs kill -9
```

## 安全注意事项

1. **密码安全**: 修改默认密码
2. **端口安全**: 考虑修改默认端口
3. **防火墙**: 配置适当的防火墙规则
4. **日志安全**: 避免在日志中暴露敏感信息

## 性能优化

1. **连接池**: 服务自动管理连接复用
2. **缓冲区**: 使用8192字节缓冲区
3. **超时设置**: 可调整连接超时时间
4. **日志级别**: 生产环境建议使用info级别

## 总结

现在使用 `./scripts/start.sh start` 启动服务时，会同时启动：
- ✅ HTTP API服务 (端口8080)
- ✅ Shadowsocks VPN服务 (端口8388)
- ✅ 自动状态检查
- ✅ 详细启动信息
- ✅ 完整的服务监控

所有服务都经过测试验证，可以安全使用！ 