# VPS VPN Service Linux 部署指南

本文档提供在Linux服务器上部署VPS VPN服务的详细步骤。

## 部署方式

### 方式一：Docker部署（推荐）

#### 1. 自动部署脚本

使用提供的自动部署脚本：

```bash
# 给脚本执行权限
chmod +x deploy.sh

# 运行部署脚本
./deploy.sh
```

#### 2. 手动Docker部署

**步骤1：安装Docker和Docker Compose**

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y docker.io docker-compose

# CentOS/RHEL
sudo yum install -y docker docker-compose

# 启动Docker服务
sudo systemctl start docker
sudo systemctl enable docker

# 将用户添加到docker组
sudo usermod -aG docker $USER
```

**步骤2：配置项目**

```bash
# 复制配置文件
cp config.example.yaml config.yaml

# 编辑配置文件
nano config.yaml
```

**步骤3：启动服务**

```bash
# 构建并启动容器
docker-compose up -d --build

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 方式二：直接编译部署

#### 1. 安装Go环境

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y golang-go

# CentOS/RHEL
sudo yum install -y golang

# 验证安装
go version
```

#### 2. 编译项目

```bash
# 下载依赖
go mod tidy

# 编译项目
go build -o vps cmd/main.go

# 或者使用Makefile
make build
```

#### 3. 配置服务

```bash
# 复制配置文件
cp config.example.yaml config.yaml

# 编辑配置文件
nano config.yaml
```

#### 4. 创建系统服务

创建服务文件 `/etc/systemd/system/vps-vpn.service`：

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

#### 5. 启动服务

```bash
# 创建用户和目录
sudo useradd -r -s /bin/false vps
sudo mkdir -p /opt/vps
sudo cp vps /opt/vps/
sudo cp config.yaml /opt/vps/
sudo chown -R vps:vps /opt/vps

# 启用并启动服务
sudo systemctl daemon-reload
sudo systemctl enable vps-vpn
sudo systemctl start vps-vpn

# 查看状态
sudo systemctl status vps-vpn
```

## 配置文件说明

编辑 `config.yaml` 文件：

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  mode: "release"  # 生产环境使用release模式

shadowsocks:
  enabled: true
  method: "aes-256-gcm"
  password: "your-strong-password-here"  # 修改为强密码
  port: 8388
  timeout: 300

log:
  level: "info"
  file: "/opt/vps/logs/vps.log"  # 生产环境建议使用文件日志
```

## 防火墙配置

### UFW (Ubuntu)

```bash
# 允许SSH
sudo ufw allow ssh

# 允许HTTP API端口
sudo ufw allow 8080

# 允许Shadowsocks端口
sudo ufw allow 8388

# 启用防火墙
sudo ufw enable
```

### firewalld (CentOS/RHEL)

```bash
# 允许SSH
sudo firewall-cmd --permanent --add-service=ssh

# 允许HTTP API端口
sudo firewall-cmd --permanent --add-port=8080/tcp

# 允许Shadowsocks端口
sudo firewall-cmd --permanent --add-port=8388/tcp

# 重新加载防火墙
sudo firewall-cmd --reload
```

### iptables

```bash
# 允许SSH
sudo iptables -A INPUT -p tcp --dport 22 -j ACCEPT

# 允许HTTP API端口
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT

# 允许Shadowsocks端口
sudo iptables -A INPUT -p tcp --dport 8388 -j ACCEPT

# 保存规则
sudo iptables-save > /etc/iptables/rules.v4
```

## 服务管理

### Docker方式

```bash
# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 重启服务
docker-compose restart

# 停止服务
docker-compose down

# 更新服务
docker-compose pull
docker-compose up -d --build
```

### 系统服务方式

```bash
# 启动服务
sudo systemctl start vps-vpn

# 停止服务
sudo systemctl stop vps-vpn

# 重启服务
sudo systemctl restart vps-vpn

# 查看状态
sudo systemctl status vps-vpn

# 查看日志
sudo journalctl -u vps-vpn -f

# 启用开机自启
sudo systemctl enable vps-vpn
```

## 监控和日志

### 查看服务状态

```bash
# API健康检查
curl http://localhost:8080/api/v1/health

# VPN状态
curl http://localhost:8080/api/v1/vpn/status

# 支持的加密方法
curl http://localhost:8080/api/v1/vpn/methods
```

### 日志管理

```bash
# 查看应用日志
tail -f /opt/vps/logs/vps.log

# 查看系统日志
sudo journalctl -u vps-vpn -f

# 查看Docker日志
docker-compose logs -f
```

## 安全建议

1. **强密码**：使用强密码，避免默认密码
2. **防火墙**：确保正确配置防火墙规则
3. **SSL/TLS**：生产环境建议启用HTTPS
4. **日志轮转**：配置日志轮转避免磁盘空间不足
5. **定期更新**：定期更新系统和依赖包
6. **监控**：设置监控和告警

## 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   # 检查端口占用
   sudo netstat -tlnp | grep :8080
   sudo netstat -tlnp | grep :8388
   ```

2. **权限问题**
   ```bash
   # 检查文件权限
   ls -la /opt/vps/
   sudo chown -R vps:vps /opt/vps/
   ```

3. **服务启动失败**
   ```bash
   # 查看详细日志
   sudo journalctl -u vps-vpn -n 50
   ```

4. **网络连接问题**
   ```bash
   # 测试端口连通性
   telnet localhost 8080
   telnet localhost 8388
   ```

## 备份和恢复

### 备份配置

```bash
# 备份配置文件
cp config.yaml config.yaml.backup

# 备份日志
tar -czf logs-backup-$(date +%Y%m%d).tar.gz logs/
```

### 恢复配置

```bash
# 恢复配置文件
cp config.yaml.backup config.yaml

# 重启服务
sudo systemctl restart vps-vpn
# 或
docker-compose restart
```

## 性能优化

1. **系统调优**
   ```bash
   # 增加文件描述符限制
   echo "* soft nofile 65536" >> /etc/security/limits.conf
   echo "* hard nofile 65536" >> /etc/security/limits.conf
   ```

2. **网络优化**
   ```bash
   # 优化TCP参数
   echo 'net.core.rmem_max = 16777216' >> /etc/sysctl.conf
   echo 'net.core.wmem_max = 16777216' >> /etc/sysctl.conf
   sysctl -p
   ```

## 更新部署

### 更新Docker镜像

```bash
# 拉取最新代码
git pull

# 重新构建并启动
docker-compose up -d --build
```

### 更新二进制文件

```bash
# 重新编译
go build -o vps cmd/main.go

# 替换二进制文件
sudo cp vps /opt/vps/

# 重启服务
sudo systemctl restart vps-vpn
```

## 联系支持

如果遇到问题，请：

1. 查看日志文件
2. 检查配置文件
3. 验证网络连接
4. 提交Issue到项目仓库 