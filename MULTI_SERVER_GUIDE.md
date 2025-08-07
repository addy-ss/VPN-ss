# 多级代理使用指南

## 概述

本项目现在支持多级代理功能，可以将请求通过多个服务端进行转发，从而提高安全性和隐私保护。

## 架构说明

### 单级代理（原有功能）
```
客户端 -> 服务端 -> 目标网站
```

### 多级代理（新功能）
```
客户端 -> 第一服务端 -> 第二服务端 -> 目标网站
```

## 配置说明

### 服务端配置

在 `config.yaml` 中配置第二个服务端：

```yaml
# 第二个服务端配置
second_server:
  enabled: true   # 启用第二个服务端
  host: "192.168.1.100"  # 第二个服务端地址
  port: 8389       # 第二个服务端端口
  method: "aes-256-gcm"  # 加密方法
  password: "your_second_server_password"  # 第二个服务端密码
  timeout: 300     # 超时时间(秒)
```

### 客户端配置

在 `client/config.yaml` 中配置第二个服务端信息：

```yaml
# 第二个服务端配置
second_server:
  enabled: true   # 启用第二个服务端
  host: "192.168.1.100"  # 第二个服务端地址
  port: 8389       # 第二个服务端端口
  method: "aes-256-gcm"  # 加密方法
  password: "your_second_server_password"  # 第二个服务端密码
  timeout: 300     # 超时时间(秒)
```

## 部署步骤

### 1. 部署第二个服务端

在第二台服务器上部署相同的服务端程序：

```bash
# 在第二台服务器上
./vps-server -config config.yaml
```

### 2. 配置第一个服务端

修改第一个服务端的配置文件，启用第二个服务端：

```yaml
second_server:
  enabled: true
  host: "第二个服务器的IP地址"
  port: 8389
  method: "aes-256-gcm"
  password: "第二个服务器的密码"
  timeout: 300
```

### 3. 配置客户端

修改客户端配置文件，添加第二个服务端信息：

```yaml
second_server:
  enabled: true
  host: "第二个服务器的IP地址"
  port: 8389
  method: "aes-256-gcm"
  password: "第二个服务器的密码"
  timeout: 300
```

## API 使用

### 启动多级代理服务

```bash
curl -X POST http://localhost:8080/api/vpn/start \
  -H "Content-Type: application/json" \
  -d '{
    "port": 8388,
    "method": "aes-256-gcm",
    "password": "your_password",
    "timeout": 300,
    "second_server_enabled": true,
    "second_server_host": "192.168.1.100",
    "second_server_port": 8389,
    "second_server_method": "aes-256-gcm",
    "second_server_password": "your_second_server_password",
    "second_server_timeout": 300
  }'
```

## 安全优势

### 1. 多层加密
- 客户端到第一服务端：第一层加密
- 第一服务端到第二服务端：第二层加密
- 第二服务端到目标：第三层加密

### 2. 流量分散
- 流量通过多个服务器，增加追踪难度
- 每个服务器只看到部分流量信息

### 3. 故障隔离
- 如果某个服务端出现问题，可以快速切换到其他服务端
- 支持负载均衡和故障转移

## 注意事项

### 1. 性能影响
- 多级代理会增加延迟
- 建议选择地理位置相近的服务器

### 2. 配置一致性
- 确保所有服务端的加密方法和密码配置一致
- 检查网络连接和防火墙设置

### 3. 监控和日志
- 启用详细的日志记录
- 监控各个服务端的连接状态

## 故障排除

### 1. 连接失败
- 检查第二个服务端是否正常运行
- 验证网络连接和防火墙设置
- 确认配置参数是否正确

### 2. 性能问题
- 检查服务器资源使用情况
- 考虑调整超时时间
- 优化网络路径

### 3. 加密错误
- 确认所有服务端使用相同的加密方法
- 检查密码配置是否正确
- 验证密钥生成过程

## 高级配置

### 负载均衡
可以配置多个第二个服务端，实现负载均衡：

```yaml
# 支持多个第二个服务端（需要代码扩展）
second_servers:
  - host: "server1.example.com"
    port: 8389
    weight: 1
  - host: "server2.example.com"
    port: 8389
    weight: 1
```

### 故障转移
当第二个服务端不可用时，可以自动切换到备用服务端或直接连接目标。

## 总结

多级代理功能为您的VPN服务提供了更高的安全性和隐私保护。通过合理配置和部署，可以构建一个安全、可靠的多级代理网络。 