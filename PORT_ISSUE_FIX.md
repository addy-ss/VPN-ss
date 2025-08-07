# 端口配置问题修复

## 问题描述

用户反映流量直接到达目标服务器，没有经过配置的第二服务器。

## 问题分析

### 1. 端口配置混乱
原始配置：
- `shadowsocks.port: 8389` (标准Shadowsocks服务器)
- 自定义协议服务器：`8389 + 1 = 8390`
- 客户端连接：`127.0.0.1:8389`

### 2. 问题根源
- 8389端口运行的是**标准Shadowsocks服务器**，它**不支持**转发到第二服务器
- 8390端口运行的是**自定义协议服务器**，它**支持**转发功能
- 客户端连接到8389端口，所以使用的是标准服务器，不会转发

## 解决方案

### 1. 修改端口分配
```go
// 修改前
standardConfig.Port = config.AppConfig.Shadowsocks.Port        // 8389
customConfig.Port = config.AppConfig.Shadowsocks.Port + 1      // 8390

// 修改后
standardConfig.Port = config.AppConfig.Shadowsocks.Port - 1    // 8388
customConfig.Port = config.AppConfig.Shadowsocks.Port          // 8389
```

### 2. 端口分配说明
- **8388端口**：标准Shadowsocks服务器（兼容标准客户端）
- **8389端口**：自定义协议服务器（支持多级代理转发）
- **8080端口**：HTTP API服务器

## 修复后的配置

### 服务端配置 (config.yaml)
```yaml
shadowsocks:
  enabled: true
  method: "aes-256-gcm"
  password: "13687401432Fan!"
  port: 8389  # 自定义协议服务器端口
  timeout: 300

second_server:
  enabled: true
  host: "206.190.238.198"
  port: 8388
  method: "aes-256-gcm"
  password: "13687401432Fan!"
  timeout: 300
```

### 客户端连接
- **多级代理客户端**：连接到 `127.0.0.1:8389`
- **标准Shadowsocks客户端**：连接到 `127.0.0.1:8388`

## 数据流

### 多级代理数据流
```
客户端 -> 127.0.0.1:8389 (自定义协议) -> 206.190.238.198:8388 -> 目标网站
```

### 标准代理数据流
```
客户端 -> 127.0.0.1:8388 (标准Shadowsocks) -> 目标网站
```

## 测试步骤

### 1. 重新编译
```bash
go build -o vps-server cmd/main.go
```

### 2. 启动服务
```bash
./scripts/start.sh start
```

### 3. 运行诊断
```bash
./diagnose_forwarding.sh
```

### 4. 测试客户端
```bash
cd test_tools
./test_client
```

## 验证方法

### 1. 检查端口监听
```bash
netstat -an | grep 838
```
应该看到：
- `8388` 端口：标准Shadowsocks服务器
- `8389` 端口：自定义协议服务器

### 2. 检查服务日志
启动服务后，应该看到：
```
Standard Shadowsocks server started on port 8388
Custom Shadowsocks server started on port 8389
Second server forwarding enabled: 206.190.238.198:8388
```

### 3. 测试转发功能
使用测试客户端连接到 `127.0.0.1:8389`，应该看到：
```
Received target: google.com:80, useSecondServer: true
Forwarding to second server: 206.190.238.198:8388
Connected to second server 206.190.238.198:8388, forwarding target: google.com:80
```

## 常见问题

### 1. 端口冲突
如果8388或8389端口被占用，可以修改配置文件中的端口：
```yaml
shadowsocks:
  port: 8390  # 改为其他端口
```

### 2. 连接失败
- 检查第二服务器是否正常运行
- 验证网络连接和防火墙设置
- 确认密码和加密方法配置正确

### 3. 性能问题
- 监控网络延迟
- 检查服务器资源使用情况
- 考虑优化网络路径

## 总结

主要问题是端口配置混乱，导致客户端连接到了不支持转发的标准服务器。修复后：

1. ✅ **8388端口**：标准Shadowsocks服务器（单级代理）
2. ✅ **8389端口**：自定义协议服务器（支持多级代理）
3. ✅ **客户端连接**：使用8389端口进行多级代理
4. ✅ **转发功能**：流量会正确转发到第二服务器

现在您的多级代理转发功能应该能正常工作了！ 