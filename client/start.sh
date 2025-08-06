 #!/bin/bash

# 客户端启动脚本

# 默认配置
SERVER_HOST="206.190.238.198"
SERVER_PORT="8388"
LOCAL_PORT="1080"
PASSWORD="13687401432Fan!"
METHOD="aes-256-gcm"
TIMEOUT="300"

# 显示帮助信息
show_help() {
    echo "VPS客户端代理启动脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -s, --server HOST     服务器地址 (默认: 127.0.0.1)"
    echo "  -p, --port PORT       服务器端口 (默认: 8388)"
    echo "  -l, --local PORT      本地监听端口 (默认: 1080)"
    echo "  -k, --password PASS   密码 (默认: your_password)"
    echo "  -m, --method METHOD   加密方法 (默认: aes-256-gcm)"
    echo "  -t, --timeout SEC     超时时间 (默认: 300)"
    echo "  -h, --help            显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -s 192.168.1.100 -p 8388 -l 1080 -k mypassword"
    echo "  $0 --server 10.0.0.1 --port 8388 --local 1080 --password mypassword"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--server)
            SERVER_HOST="$2"
            shift 2
            ;;
        -p|--port)
            SERVER_PORT="$2"
            shift 2
            ;;
        -l|--local)
            LOCAL_PORT="$2"
            shift 2
            ;;
        -k|--password)
            PASSWORD="$2"
            shift 2
            ;;
        -m|--method)
            METHOD="$2"
            shift 2
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "错误: 未找到Go编译器，请先安装Go"
    exit 1
fi

# 检查端口是否被占用
if lsof -Pi :$LOCAL_PORT -sTCP:LISTEN -t >/dev/null ; then
    echo "错误: 端口 $LOCAL_PORT 已被占用"
    exit 1
fi

echo "启动VPS客户端代理..."
echo "服务器地址: $SERVER_HOST:$SERVER_PORT"
echo "本地端口: $LOCAL_PORT"
echo "加密方法: $METHOD"
echo ""

# 编译并运行
echo "编译客户端..."
go build -o vps-client main.go

if [ $? -ne 0 ]; then
    echo "编译失败"
    exit 1
fi

echo "启动客户端..."
./vps-client \
    -server="$SERVER_HOST" \
    -port="$SERVER_PORT" \
    -local="$LOCAL_PORT" \
    -password="$PASSWORD" \
    -method="$METHOD" \
    -timeout="$TIMEOUT"