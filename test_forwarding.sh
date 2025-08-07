#!/bin/bash

echo "=== 测试多级代理转发功能 ==="

# 检查服务是否运行
echo "检查服务状态..."
curl -s http://localhost:8080/api/v1/health

echo ""
echo "启动多级代理服务..."
curl -X POST http://localhost:8080/api/vpn/start \
  -H "Content-Type: application/json" \
  -d '{
    "port": 8388,
    "method": "aes-256-gcm",
    "password": "13687401432Fan!",
    "timeout": 300,
    "second_server_enabled": true,
    "second_server_host": "206.190.238.198",
    "second_server_port": 8388,
    "second_server_method": "aes-256-gcm",
    "second_server_password": "13687401432Fan!",
    "second_server_timeout": 300
  }'

echo ""
echo "检查VPN状态..."
curl -s http://localhost:8080/api/vpn/status

echo ""
echo "=== 测试完成 ==="
echo "现在可以测试客户端连接到 127.0.0.1:8389"
echo "流量应该会转发到 206.190.238.198:8388" 