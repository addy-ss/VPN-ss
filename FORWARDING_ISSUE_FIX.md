# 多级代理转发问题修复

## 问题描述

用户配置了多级代理：
- 本地服务端端口：8389
- 转发到：206.190.238.198:8388
- 但是流量没有到达第二服务器

## 问题分析

### 1. 配置问题
从日志可以看到：
```
Standard Shadowsocks server started on port 8389
Custom Shadowsocks server started on port 8390
Second server forwarding enabled: 206.190.238.198:8388
```

### 2. 代码问题
在 `forwardToSecondServer` 方法中，代码试图访问：
```go
secondServerAddr := fmt.Sprintf("%s:%d", p.secondConfig.SecondServerHost, p.secondConfig.SecondServerPort)
```

但是 `secondConfig` 是一个 `*Config` 类型，它只有以下字段：
- `Method`
- `Password` 
- `Port`
- `Timeout`

没有 `SecondServerHost` 和 `SecondServerPort` 字段。

## 解决方案

### 1. 修复代码错误
将 `forwardToSecondServer` 方法中的配置访问改为：

```go
// 修复前
secondServerAddr := fmt.Sprintf("%s:%d", p.secondConfig.SecondServerHost, p.secondConfig.SecondServerPort)
secondServerConn, err := net.DialTimeout("tcp", secondServerAddr, time.Duration(p.secondConfig.Timeout)*time.Second)

// 修复后
secondServerAddr := fmt.Sprintf("%s:%d", p.config.SecondServerHost, p.config.SecondServerPort)
secondServerConn, err := net.DialTimeout("tcp", secondServerAddr, time.Duration(p.config.SecondServerTimeout)*time.Second)
```

### 2. 添加调试日志
在 `handleProxy` 方法中添加了调试日志：
```go
p.logger.Infof("Received target: %s, useSecondServer: %v", target, p.useSecondServer)

if p.useSecondServer && p.secondConfig != nil {
    p.logger.Infof("Forwarding to second server: %s:%d", p.config.SecondServerHost, p.config.SecondServerPort)
    return p.forwardToSecondServer(conn, target, decryptor)
}

p.logger.Infof("Connecting directly to target: %s", target)
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

### 3. 测试转发功能
```bash
# 使用测试脚本
./test_forwarding.sh

# 或者使用测试客户端
cd test_tools
./test_client
```

### 4. 检查日志
查看服务端日志，应该能看到：
```
Received target: google.com:80, useSecondServer: true
Forwarding to second server: 206.190.238.198:8388
Connected to second server 206.190.238.198:8388, forwarding target: google.com:80
```

## 配置验证

### 服务端配置 (config.yaml)
```yaml
second_server:
  enabled: true
  host: "206.190.238.198"
  port: 8388
  method: "aes-256-gcm"
  password: "13687401432Fan!"
  timeout: 300
```

### 客户端连接
客户端应该连接到：`127.0.0.1:8389`

## 数据流

修复后的数据流：
```
客户端 -> 127.0.0.1:8389 (本地服务端) -> 206.190.238.198:8388 (第二服务端) -> 目标网站
```

## 验证方法

### 1. 网络连接测试
```bash
# 测试到第二服务器的连接
telnet 206.190.238.198 8388
```

### 2. 日志监控
```bash
# 查看服务端日志
tail -f vps.log
```

### 3. 网络抓包
```bash
# 使用tcpdump监控流量
sudo tcpdump -i any host 206.190.238.198
```

## 常见问题

### 1. 连接超时
- 检查第二服务器是否正常运行
- 验证网络连接和防火墙设置
- 确认端口配置正确

### 2. 加密错误
- 确保所有服务端使用相同的加密方法
- 验证密码配置正确
- 检查密钥生成过程

### 3. 性能问题
- 监控网络延迟
- 检查服务器资源使用情况
- 考虑优化网络路径

## 总结

主要问题是代码中错误地访问了不存在的配置字段。修复后，多级代理转发功能应该能正常工作。

关键修复点：
1. ✅ 修复了 `forwardToSecondServer` 方法中的配置访问
2. ✅ 添加了详细的调试日志
3. ✅ 创建了测试工具验证功能
4. ✅ 提供了完整的测试和验证方法 