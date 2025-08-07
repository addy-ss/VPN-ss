# 标准Shadowsocks服务器转发功能

## 问题描述

用户要求第一层和第二层都使用标准Shadowsocks配置，但是原来的标准Shadowsocks服务器不支持转发到第二服务器。

## 解决方案

### 1. 扩展标准Shadowsocks服务器
修改了 `internal/vpn/standard_shadowsocks.go`，添加了转发功能：

#### 新增字段
```go
type StandardShadowsocksServer struct {
    config          *Config
    listener        net.Listener
    logger          *logrus.Logger
    ctx             context.Context
    cancel          context.CancelFunc
    useSecondServer bool  // 新增：是否启用第二服务器
}
```

#### 新增方法
1. **forwardToSecondServer**: 转发到第二服务器
2. **forwardBetweenServers**: 在两个服务器之间转发数据
3. **writeTargetAddress**: 写入目标地址

### 2. 转发逻辑
```go
// 在handleConnection方法中添加
if s.useSecondServer {
    s.logger.Infof("Forwarding to second server: %s:%d", s.config.SecondServerHost, s.config.SecondServerPort)
    s.forwardToSecondServer(ssconn, tgt.String(), cipher)
    return
}
```

### 3. 配置传递
在 `cmd/main.go` 中，将第二服务器配置传递给标准Shadowsocks服务器：

```go
standardConfig := &vpn.Config{
    Port:     config.AppConfig.Shadowsocks.Port,
    Method:   config.AppConfig.Shadowsocks.Method,
    Password: config.AppConfig.Shadowsocks.Password,
    Timeout:  config.AppConfig.Shadowsocks.Timeout,
    // 第二个服务端配置
    SecondServerEnabled:  config.AppConfig.SecondServer.Enabled,
    SecondServerHost:     config.AppConfig.SecondServer.Host,
    SecondServerPort:     config.AppConfig.SecondServer.Port,
    SecondServerMethod:   config.AppConfig.SecondServer.Method,
    SecondServerPassword: config.AppConfig.SecondServer.Password,
    SecondServerTimeout:  config.AppConfig.SecondServer.Timeout,
}
```

## 数据流

### 标准Shadowsocks多级代理
```
客户端 -> 127.0.0.1:8389 (标准Shadowsocks) -> 206.190.238.198:8388 (标准Shadowsocks) -> 目标网站
```

### 协议兼容性
- **第一层**: 标准Shadowsocks协议
- **第二层**: 标准Shadowsocks协议
- **客户端**: 任何标准Shadowsocks客户端

## 配置示例

### 服务端配置 (config.yaml)
```yaml
shadowsocks:
  enabled: true
  method: "aes-256-gcm"
  password: "13687401432Fan!"
  port: 8389
  timeout: 300

second_server:
  enabled: true
  host: "206.190.238.198"
  port: 8388
  method: "aes-256-gcm"
  password: "13687401432Fan!"
  timeout: 300
```

### 客户端配置
任何标准Shadowsocks客户端都可以使用：
```json
{
  "server": "127.0.0.1",
  "server_port": 8389,
  "password": "13687401432Fan!",
  "method": "aes-256-gcm"
}
```

## 端口分配

### 修复后的端口分配
- **8389端口**: 标准Shadowsocks服务器（支持转发）
- **8390端口**: 自定义协议服务器（支持转发）
- **8080端口**: HTTP API服务器

### 客户端连接
- **标准Shadowsocks客户端**: 连接到 `127.0.0.1:8389`
- **自定义协议客户端**: 连接到 `127.0.0.1:8390`

## 测试步骤

### 1. 重新编译
```bash
go build -o vps-server cmd/main.go
```

### 2. 启动服务
```bash
./scripts/start.sh start
```

### 3. 测试标准Shadowsocks转发
```bash
./test_standard_forwarding.sh
```

### 4. 使用标准客户端测试
可以使用任何标准Shadowsocks客户端连接到 `127.0.0.1:8389`

## 验证方法

### 1. 检查服务日志
启动服务后，应该看到：
```
Standard Shadowsocks server started on port 8389
Custom Shadowsocks server started on port 8390
Second server forwarding enabled: 206.190.238.198:8388
```

### 2. 测试连接
使用标准Shadowsocks客户端连接后，应该看到：
```
Received target: google.com:80, useSecondServer: true
Forwarding to second server: 206.190.238.198:8388
Connected to second server 206.190.238.198:8388, forwarding target: google.com:80
```

### 3. 网络抓包验证
```bash
# 监控到第二服务器的流量
sudo tcpdump -i any host 206.190.238.198
```

## 优势

### 1. 协议兼容性
- 支持所有标准Shadowsocks客户端
- 无需修改客户端配置
- 保持协议标准性

### 2. 安全性
- 两层标准Shadowsocks加密
- 流量分散到多个服务器
- 增加追踪难度

### 3. 灵活性
- 可以选择启用或禁用转发
- 支持不同的加密方法
- 可以配置不同的超时时间

## 常见问题

### 1. 连接失败
- 检查第二服务器是否正常运行
- 验证网络连接和防火墙设置
- 确认密码和加密方法配置正确

### 2. 性能问题
- 监控网络延迟
- 检查服务器资源使用情况
- 考虑优化网络路径

### 3. 协议兼容性
- 确保使用标准Shadowsocks协议
- 验证客户端配置正确
- 检查加密方法支持

## 总结

现在标准Shadowsocks服务器也支持转发功能了！

### 主要改进
1. ✅ **扩展标准服务器**: 添加了转发到第二服务器的功能
2. ✅ **保持协议兼容**: 使用标准Shadowsocks协议
3. ✅ **支持所有客户端**: 任何标准Shadowsocks客户端都可以使用
4. ✅ **双重加密**: 两层标准Shadowsocks加密

### 使用方式
- **客户端**: 连接到 `127.0.0.1:8389`
- **协议**: 标准Shadowsocks
- **转发**: 自动转发到 `206.190.238.198:8388`

现在您的多级代理系统完全使用标准Shadowsocks协议，兼容性更好，安全性更高！ 