# 🚀 VPS VPN Service 快速启动指南

## 方法一：直接运行（推荐）

### 1. 安装依赖
```bash
go mod tidy
```

### 2. 构建项目
```bash
go build -o vps cmd/main.go
```

### 3. 配置安全设置
```bash
# 复制示例配置文件
cp config.example.yaml config.yaml

# 编辑配置文件，设置安全密码
nano config.yaml
```

### 4. 启动服务
```bash
./vps
```

## 方法二：使用启动脚本

### 1. 给脚本执行权限
```bash
chmod +x scripts/start.sh
```

### 2. 启动服务
```bash
./scripts/start.sh start
```

### 3. 查看状态
```bash
./scripts/start.sh status
```

### 4. 停止服务
```bash
./scripts/start.sh stop
```

## 方法三：Docker部署

### 1. 构建镜像
```bash
docker build -t vps-vpn .
```

### 2. 运行容器
```bash
docker run -d \
  --name vps-vpn \
  -p 8080:8080 \
  -p 8388:8388 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  vps-vpn
```

### 3. 使用Docker Compose
```bash
docker-compose up -d
```

## 🔧 配置说明

### 1. 基础配置 (config.yaml)
```yaml
server:
  port: 8080
  host: "0.0.0.0"
  mode: "debug"

shadowsocks:
  enabled: true
  method: "aes-256-gcm"
  password: "your-secure-password-here"  # 请修改为强密码
  port: 8388
  timeout: 300

log:
  level: "info"
```

### 2. 安全配置 (config/security.yaml)
```yaml
security:
  auth:
    enabled: true
    jwt_secret: ""  # 留空将自动生成
    max_login_attempts: 5
    
  encryption:
    default_method: "aes-256-gcm"
    min_password_length: 12
    
  audit:
    enabled: true
    retention_days: 90
```

## 🧪 测试服务

### 1. 健康检查
```bash
curl http://localhost:8080/api/v1/health
```

### 2. 获取支持的加密方法
```bash
curl http://localhost:8080/api/v1/vpn/methods
```

### 3. 启动VPN服务
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

### 4. 生成配置
```bash
curl -X POST http://localhost:8080/api/v1/vpn/config/generate \
  -H "Content-Type: application/json" \
  -d '{
    "port": 8388,
    "method": "aes-256-gcm",
    "password": "your-password"
  }'
```

## 🐍 Python测试客户端

### 1. 安装依赖
```bash
pip install requests
```

### 2. 运行测试
```bash
python3 scripts/test_client.py
```

## 📊 监控和日志

### 1. 查看日志
```bash
# 实时查看日志
tail -f vps.log

# 查看系统日志
sudo journalctl -u vps-vpn -f
```

### 2. 检查端口
```bash
# 检查服务是否运行
netstat -tlnp | grep :8080
netstat -tlnp | grep :8388

# 或者使用ss命令
ss -tlnp | grep :8080
```

### 3. 进程管理
```bash
# 查看进程
ps aux | grep vps

# 杀死进程
pkill -f vps
```

## 🔒 安全设置

### 1. 防火墙配置
```bash
# Ubuntu/Debian
sudo ufw allow 8080/tcp
sudo ufw allow 8388/tcp

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=8388/tcp
sudo firewall-cmd --reload
```

### 2. 系统服务
```bash
# 创建服务文件
sudo nano /etc/systemd/system/vps-vpn.service
```

服务文件内容：
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

## 🚨 故障排除

### 1. 端口被占用
```bash
# 检查端口占用
sudo lsof -i :8080
sudo lsof -i :8388

# 杀死占用进程
sudo kill -9 <PID>
```

### 2. 权限问题
```bash
# 给执行权限
chmod +x vps
chmod +x scripts/start.sh

# 检查文件权限
ls -la vps
```

### 3. 配置文件问题
```bash
# 验证配置文件
go run cmd/main.go --config=config.yaml

# 重新生成配置
cp config.example.yaml config.yaml
```

### 4. 依赖问题
```bash
# 清理并重新安装依赖
go clean -modcache
go mod tidy
go mod download
```

## 📈 性能优化

### 1. 系统参数优化
```bash
# 增加文件描述符限制
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# 优化网络参数
echo "net.core.rmem_max = 16777216" >> /etc/sysctl.conf
echo "net.core.wmem_max = 16777216" >> /etc/sysctl.conf
sysctl -p
```

### 2. 监控脚本
```bash
#!/bin/bash
# 监控脚本
while true; do
    if ! curl -s http://localhost:8080/api/v1/health > /dev/null; then
        echo "$(date): VPS VPN Service is down!"
        # 重启服务
        ./scripts/start.sh restart
    fi
    sleep 30
done
```

## 🎯 快速验证

运行以下命令验证服务是否正常：

```bash
# 1. 启动服务
./vps &

# 2. 等待几秒
sleep 3

# 3. 测试健康检查
curl http://localhost:8080/api/v1/health

# 4. 测试VPN方法
curl http://localhost:8080/api/v1/vpn/methods

# 5. 停止服务
pkill -f vps
```

如果所有测试都通过，说明服务运行正常！🎉

## 📞 获取帮助

- 查看详细文档：`README.md`
- 安全分析：`SECURITY_ANALYSIS.md`
- 项目演示：`DEMO.md`
- 使用Makefile：`make help` 