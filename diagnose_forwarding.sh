#!/bin/bash

echo "=== 多级代理转发诊断 ==="

# 1. 检查配置文件
echo "1. 检查配置文件..."
if [ -f "config.yaml" ]; then
    echo "配置文件存在"
    echo "第二个服务端配置:"
    grep -A 5 "second_server:" config.yaml
else
    echo "配置文件不存在"
    exit 1
fi

echo ""

# 2. 检查服务状态
echo "2. 检查服务状态..."
ps aux | grep vps-server

echo ""

# 3. 检查端口监听
echo "3. 检查端口监听..."
netstat -an | grep 838

echo ""

# 4. 测试API连接
echo "4. 测试API连接..."
curl -s http://localhost:8080/api/v1/health

echo ""

# 5. 启动多级代理服务
echo "5. 启动多级代理服务..."
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

# 6. 检查VPN状态
echo "6. 检查VPN状态..."
curl -s http://localhost:8080/api/vpn/status

echo ""

# 7. 测试到第二服务器的连接
echo "7. 测试到第二服务器的连接..."
nc -zv 206.190.238.198 8388

echo ""

# 8. 测试本地端口连接
echo "8. 测试本地端口连接..."
nc -zv 127.0.0.1 8389

echo ""

# 9. 运行测试客户端
echo "9. 运行测试客户端..."
cd test_tools
./test_client

echo ""
echo "=== 诊断完成 ==="
echo "请检查上述输出，特别注意："
echo "1. 配置文件中的second_server设置"
echo "2. 服务是否正常运行"
echo "3. 端口是否正确监听"
echo "4. 到第二服务器的连接是否成功"
echo "5. 测试客户端的输出" 