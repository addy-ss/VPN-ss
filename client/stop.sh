 #!/bin/bash

# 客户端停止脚本

echo "正在停止VPS客户端代理..."

# 查找并杀死vps-client进程
PIDS=$(pgrep -f "vps-client")

if [ -z "$PIDS" ]; then
    echo "未找到运行中的vps-client进程"
    exit 0
fi

echo "找到进程: $PIDS"

# 发送SIGTERM信号
for PID in $PIDS; do
    echo "正在停止进程 $PID..."
    kill -TERM $PID
    
    # 等待进程结束
    for i in {1..10}; do
        if ! kill -0 $PID 2>/dev/null; then
            echo "进程 $PID 已停止"
            break
        fi
        sleep 1
    done
    
    # 如果进程仍然存在，强制杀死
    if kill -0 $PID 2>/dev/null; then
        echo "强制停止进程 $PID..."
        kill -KILL $PID
    fi
done

echo "VPS客户端代理已停止"