#!/bin/bash

# 多级代理演示脚本
echo "=== VPS 多级代理演示 ==="

# 检查配置文件
if [ ! -f "config.yaml" ]; then
    echo "创建配置文件..."
    cp config.example.yaml config.yaml
    echo "请编辑 config.yaml 文件，配置第二个服务端信息"
fi

# 启动服务端
echo "启动服务端..."
./vps-server -config config.yaml &

# 等待服务启动
sleep 3

# 测试API
echo "测试多级代理API..."

# 启动多级代理服务
curl -X POST http://localhost:8080/api/vpn/start \
  -H "Content-Type: application/json" \
  -d '{
    "port": 8388,
    "method": "aes-256-gcm",
    "password": "demo_password",
    "timeout": 300,
    "second_server_enabled": true,
    "second_server_host": "127.0.0.1",
    "second_server_port": 8389,
    "second_server_method": "aes-256-gcm",
    "second_server_password": "second_demo_password",
    "second_server_timeout": 300
  }'

echo ""
echo "检查服务状态..."
curl http://localhost:8080/api/vpn/status

echo ""
echo "获取支持的加密方法..."
curl http://localhost:8080/api/vpn/methods

echo ""
echo "=== 演示完成 ==="
echo "服务端运行在 http://localhost:8080"
echo "多级代理端口: 8388"
echo "第二个服务端端口: 8389"
echo ""
echo "要停止服务，请按 Ctrl+C" 