# VPS VPN Service 演示

这个演示将展示如何使用VPS VPN Service项目。

## 快速开始

### 1. 启动服务

```bash
# 使用启动脚本
./scripts/start.sh start

# 或者直接运行
go run cmd/main.go
```

### 2. 测试API接口

#### 健康检查
```bash
curl http://localhost:8080/api/v1/health
```

响应：
```json
{
  "status": "healthy",
  "service": "vps-vpn"
}
```

#### 获取支持的加密方法
```bash
curl http://localhost:8080/api/v1/vpn/methods
```

响应：
```json
{
  "methods": [
    "aes-256-gcm",
    "chacha20-poly1305",
    "aes-128-gcm",
    "aes-192-gcm"
  ]
}
```

#### 启动VPN服务
```bash
curl -X POST http://localhost:8080/api/v1/vpn/start \
  -H "Content-Type: application/json" \
  -d '{
    "port": 8388,
    "method": "aes-256-gcm",
    "password": "your-secure-password",
    "timeout": 300
  }'
```

响应：
```json
{
  "message": "VPN server started successfully",
  "port": 8388,
  "method": "aes-256-gcm"
}
```

#### 获取VPN状态
```bash
curl http://localhost:8080/api/v1/vpn/status
```

响应：
```json
{
  "status": "running"
}
```

#### 生成Shadowsocks配置
```bash
curl -X POST http://localhost:8080/api/v1/vpn/config/generate \
  -H "Content-Type: application/json" \
  -d '{
    "port": 8388,
    "method": "aes-256-gcm",
    "password": "your-password"
  }'
```

响应：
```json
{
  "config": "{\n\t\t\"server\": \"0.0.0.0\",\n\t\t\"server_port\": 8388,\n\t\t\"password\": \"base64-encoded-password\",\n\t\t\"method\": \"aes-256-gcm\",\n\t\t\"timeout\": 300\n\t}"
}
```

#### 停止VPN服务
```bash
curl -X POST http://localhost:8080/api/v1/vpn/stop
```

响应：
```json
{
  "message": "VPN server stopped successfully"
}
```

### 3. 使用Python测试客户端

```bash
# 安装依赖
pip install requests

# 运行测试客户端
python3 scripts/test_client.py
```

### 4. 使用Docker

#### 构建镜像
```bash
docker build -t vps-vpn .
```

#### 运行容器
```bash
docker run -d \
  --name vps-vpn \
  -p 8080:8080 \
  -p 8388:8388 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  vps-vpn
```

#### 使用Docker Compose
```bash
docker-compose up -d
```

### 5. 客户端配置示例

#### Shadowsocks客户端配置
```json
{
  "server": "your-server-ip",
  "server_port": 8388,
  "password": "your-secure-password",
  "method": "aes-256-gcm",
  "timeout": 300
}
```

#### 支持的客户端
- Shadowsocks Windows
- Shadowsocks Android
- Shadowsocks iOS
- V2Ray (Shadowsocks协议)
- Clash (Shadowsocks协议)

## 高级用法

### 1. 自定义配置

编辑 `config.yaml` 文件：

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  mode: "release"  # 生产环境使用

shadowsocks:
  enabled: true
  method: "chacha20-poly1305"  # 更快的加密方法
  password: "your-very-secure-password"
  port: 8388
  timeout: 300

log:
  level: "info"
  file: "/var/log/vps.log"  # 日志文件
```

### 2. 使用Makefile

```bash
# 构建项目
make build

# 运行测试
make test

# 代码格式化
make fmt

# 代码检查
make lint

# 构建所有平台
make build-all

# 显示帮助
make help
```

### 3. 系统服务

创建systemd服务文件 `/etc/systemd/system/vps-vpn.service`：

```ini
[Unit]
Description=VPS VPN Service
After=network.target

[Service]
Type=simple
User=vps
WorkingDirectory=/opt/vps
ExecStart=/opt/vps/vps
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启用服务：
```bash
sudo systemctl enable vps-vpn
sudo systemctl start vps-vpn
sudo systemctl status vps-vpn
```

## 故障排除

### 1. 端口被占用
```bash
# 检查端口占用
sudo netstat -tlnp | grep :8080
sudo netstat -tlnp | grep :8388

# 杀死占用进程
sudo kill -9 <PID>
```

### 2. 防火墙配置
```bash
# Ubuntu/Debian
sudo ufw allow 8080/tcp
sudo ufw allow 8388/tcp

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=8388/tcp
sudo firewall-cmd --reload
```

### 3. 查看日志
```bash
# 查看应用日志
tail -f vps.log

# 查看系统日志
sudo journalctl -u vps-vpn -f
```

### 4. 性能优化

#### 调整系统参数
```bash
# 增加文件描述符限制
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# 优化网络参数
echo "net.core.rmem_max = 16777216" >> /etc/sysctl.conf
echo "net.core.wmem_max = 16777216" >> /etc/sysctl.conf
sysctl -p
```

## 安全建议

1. **强密码**: 使用随机生成的强密码
2. **防火墙**: 只开放必要的端口
3. **SSL/TLS**: 生产环境启用HTTPS
4. **日志监控**: 定期检查日志文件
5. **定期更新**: 保持依赖包最新
6. **访问控制**: 限制API访问IP
7. **备份配置**: 定期备份配置文件

## 监控和告警

### 1. 健康检查脚本
```bash
#!/bin/bash
if ! curl -s http://localhost:8080/api/v1/health > /dev/null; then
    echo "VPS VPN Service is down!"
    # 发送告警邮件或通知
fi
```

### 2. 添加到crontab
```bash
# 每分钟检查一次
* * * * * /path/to/health_check.sh
```

## 扩展功能

### 1. 添加用户管理
- 多用户支持
- 用户配额限制
- 使用统计

### 2. 添加Web界面
- 管理面板
- 实时监控
- 配置管理

### 3. 添加更多协议
- V2Ray支持
- Trojan支持
- WireGuard支持

### 4. 添加数据库
- 用户数据持久化
- 使用统计
- 日志存储

这个演示展示了VPS VPN Service的基本用法和高级功能。项目提供了完整的VPN解决方案，支持Shadowsocks协议，并提供了丰富的API接口。 