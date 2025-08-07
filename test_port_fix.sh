#!/bin/bash

echo "=== 测试端口修复 ==="

# 停止现有服务
echo "停止现有服务..."
pkill -f vps-server

# 等待服务停止
sleep 2

# 启动服务
echo "启动服务..."
./vps-server -config config.yaml &
SERVER_PID=$!

# 等待服务启动
sleep 3

echo "服务PID: $SERVER_PID"

# 检查端口监听
echo "检查端口监听..."
netstat -an | grep 838

# 测试API
echo ""
echo "测试API..."
curl -s http://localhost:8080/api/v1/health

echo ""
echo "启动多级代理服务..."
curl -X POST http://localhost:8080/api/vpn/start \
  -H "Content-Type: application/json" \
  -d '{
    "port": 8389,
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
echo "现在客户端应该连接到 127.0.0.1:8389"
echo "流量应该会转发到 206.190.238.198:8388"

# 保持服务运行
echo "按 Ctrl+C 停止服务"
wait $SERVER_PID 