# Shadowsocks 使用指南

## 概述

VPS服务现在支持两种Shadowsocks协议：

1. **标准Shadowsocks协议** (端口8388) - 兼容所有标准Shadowsocks客户端
2. **自定义协议** (端口8389) - 用于我们的自定义客户端

## 服务器配置

### 标准Shadowsocks服务器
- **端口**: 8388
- **加密方法**: aes-256-gcm
- **密码**: 13687401432Fan!
- **协议**: 标准Shadowsocks协议

### 自定义协议服务器
- **端口**: 8389
- **加密方法**: aes-256-gcm
- **密码**: 13687401432Fan!
- **协议**: 自定义协议

## 客户端配置

### 1. 标准Shadowsocks客户端配置

#### 支持的客户端
- Shadowsocks Windows
- Shadowsocks Android
- Shadowsocks iOS
- V2Ray (Shadowsocks协议)
- Clash (Shadowsocks协议)
- 其他标准Shadowsocks客户端

#### 配置参数
```json
{
  "server": "127.0.0.1",
  "server_port": 8388,
  "password": "13687401432Fan!",
  "method": "aes-256-gcm",
  "timeout": 300
}
```

#### 配置步骤
1. 打开你的Shadowsocks客户端
2. 添加新服务器
3. 填写以下信息：
   - 服务器地址: `127.0.0.1`
   - 端口: `8388`
   - 密码: `13687401432Fan!`
   - 加密方法: `aes-256-gcm`
4. 保存并连接

### 2. 自定义客户端配置

#### 使用我们的客户端
```bash
cd client
./start.sh start
```

#### 配置文件 (client/config.yaml)
```yaml
server:
  host: "127.0.0.1"
  port: 8389  # 注意：使用8389端口

local:
  port: 1080

security:
  password: "13687401432Fan!"
  method: "aes-256-gcm"
  timeout: 300
```

## 测试连接

### 1. 测试标准Shadowsocks客户端
1. 配置标准Shadowsocks客户端连接到端口8388
2. 启动客户端
3. 访问网站测试连接

### 2. 测试自定义客户端
```bash
# 启动自定义客户端
cd client
./start.sh start

# 测试连接
curl --socks5 127.0.0.1:1080 --connect-timeout 10 -s http://httpbin.org/ip
```

### 3. 检查服务状态
```bash
# 检查服务端状态
./scripts/start.sh status

# 检查端口监听
lsof -i :8388
lsof -i :8389
```

## 故障排除

### 1. 标准客户端连接失败

**可能原因**：
- 密码错误
- 加密方法不匹配
- 端口被占用

**解决方法**：
```bash
# 检查服务状态
./scripts/start.sh status

# 检查端口监听
lsof -i :8388

# 重启服务
./scripts/start.sh restart
```

### 2. 自定义客户端连接失败

**可能原因**：
- 配置文件错误
- 端口冲突
- 服务端未启动

**解决方法**：
```bash
# 检查客户端状态
cd client
./start.sh status

# 检查配置文件
cat config.yaml

# 重启客户端
./start.sh restart
```

### 3. 浏览器无法访问

**可能原因**：
- 客户端未正确配置
- 代理设置错误
- 网络问题

**解决方法**：
1. 确认客户端已连接
2. 检查浏览器代理设置
3. 尝试访问其他网站

## 端口说明

| 端口 | 用途 | 协议 |
|------|------|------|
| 8080 | HTTP API服务 | HTTP |
| 8388 | 标准Shadowsocks | 标准协议 |
| 8389 | 自定义Shadowsocks | 自定义协议 |
| 1080 | 自定义客户端本地代理 | SOCKS5 |

## 安全注意事项

1. **密码安全**: 在生产环境中修改默认密码
2. **端口安全**: 考虑修改默认端口
3. **防火墙**: 配置适当的防火墙规则
4. **日志安全**: 避免在日志中暴露敏感信息

## 性能优化

1. **选择合适的加密方法**:
   - `aes-256-gcm`: 安全性最高，性能较好
   - `chacha20-poly1305`: 性能最好，安全性高

2. **网络优化**:
   - 选择合适的服务器位置
   - 使用稳定的网络连接
   - 调整超时时间

## 总结

现在你有两种选择：

1. **使用标准Shadowsocks客户端** (推荐)
   - 连接到端口8388
   - 兼容所有标准客户端
   - 配置简单

2. **使用自定义客户端**
   - 连接到端口8389
   - 使用我们的客户端
   - 功能更丰富

两种方式都经过测试验证，可以安全使用！ 