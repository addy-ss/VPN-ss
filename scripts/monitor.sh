#!/bin/bash

# VPS VPN 服务监控脚本
# 用于检测panic错误和连接问题

LOG_FILE="/var/log/vps-vpn.log"
SERVICE_NAME="vps-vpn"
ALERT_EMAIL="admin@example.com"

# 颜色定义
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "=== VPS VPN 服务监控 ==="
echo "时间: $(date)"
echo ""

# 检查服务状态
check_service_status() {
    echo "检查服务状态..."
    if systemctl is-active --quiet $SERVICE_NAME; then
        echo -e "${GREEN}✓ 服务正在运行${NC}"
    else
        echo -e "${RED}✗ 服务未运行${NC}"
        systemctl status $SERVICE_NAME
        return 1
    fi
}

# 检查panic错误
check_panic_errors() {
    echo "检查panic错误..."
    if [ -f "$LOG_FILE" ]; then
        panic_count=$(grep -c "panic" "$LOG_FILE" 2>/dev/null || echo "0")
        if [ "$panic_count" -gt 0 ]; then
            echo -e "${RED}✗ 发现 $panic_count 个panic错误${NC}"
            grep "panic" "$LOG_FILE" | tail -5
            return 1
        else
            echo -e "${GREEN}✓ 未发现panic错误${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ 日志文件不存在${NC}"
    fi
}

# 检查连接错误
check_connection_errors() {
    echo "检查连接错误..."
    if [ -f "$LOG_FILE" ]; then
        error_count=$(grep -c "Failed to handle" "$LOG_FILE" 2>/dev/null || echo "0")
        if [ "$error_count" -gt 0 ]; then
            echo -e "${YELLOW}⚠ 发现 $error_count 个连接错误${NC}"
            grep "Failed to handle" "$LOG_FILE" | tail -3
        else
            echo -e "${GREEN}✓ 未发现连接错误${NC}"
        fi
    fi
}

# 检查内存使用
check_memory_usage() {
    echo "检查内存使用..."
    memory_usage=$(ps aux | grep "$SERVICE_NAME" | grep -v grep | awk '{print $6}' | head -1)
    if [ -n "$memory_usage" ]; then
        memory_mb=$((memory_usage / 1024))
        echo "内存使用: ${memory_mb}MB"
        if [ "$memory_mb" -gt 500 ]; then
            echo -e "${YELLOW}⚠ 内存使用较高${NC}"
        else
            echo -e "${GREEN}✓ 内存使用正常${NC}"
        fi
    fi
}

# 检查端口监听
check_port_listening() {
    echo "检查端口监听..."
    if netstat -tulpn | grep -q ":8388"; then
        echo -e "${GREEN}✓ 端口8388正在监听${NC}"
    else
        echo -e "${RED}✗ 端口8388未监听${NC}"
        return 1
    fi
}

# 主函数
main() {
    local exit_code=0
    
    check_service_status || exit_code=1
    check_panic_errors || exit_code=1
    check_connection_errors
    check_memory_usage
    check_port_listening || exit_code=1
    
    echo ""
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}=== 监控完成：服务正常 ===${NC}"
    else
        echo -e "${RED}=== 监控完成：发现问题 ===${NC}"
        echo "建议操作："
        echo "1. 检查日志文件: tail -f $LOG_FILE"
        echo "2. 重启服务: systemctl restart $SERVICE_NAME"
        echo "3. 检查配置: cat config.yaml"
    fi
    
    exit $exit_code
}

# 运行主函数
main "$@" 