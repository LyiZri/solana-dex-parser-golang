#!/bin/bash

# 紧急停止脚本 - 停止所有Go处理进程
# 使用方法: ./stop_all_processes.sh

echo "🛑 正在停止所有Go处理进程..."

# 查找所有相关进程
GO_PROCESSES=$(pgrep -f "go run ./src/main.go")

if [ -z "$GO_PROCESSES" ]; then
    echo "✅ 没有发现运行中的Go进程"
    exit 0
fi

echo "🔍 发现 $(echo "$GO_PROCESSES" | wc -l) 个运行中的进程"

# 优雅停止
echo "⏳ 尝试优雅停止进程..."
pkill -TERM -f "go run ./src/main.go"

# 等待5秒
sleep 5

# 检查是否还有进程
REMAINING=$(pgrep -f "go run ./src/main.go")
if [ -z "$REMAINING" ]; then
    echo "✅ 所有进程已成功停止"
else
    echo "⚠️  还有 $(echo "$REMAINING" | wc -l) 个进程未停止，强制终止..."
    pkill -KILL -f "go run ./src/main.go"
    sleep 2
    
    # 最终检查
    FINAL_CHECK=$(pgrep -f "go run ./src/main.go")
    if [ -z "$FINAL_CHECK" ]; then
        echo "✅ 所有进程已强制停止"
    else
        echo "❌ 仍有进程无法停止，请手动检查"
        echo "剩余进程:"
        ps aux | grep "go run ./src/main.go" | grep -v grep
    fi
fi

# 显示当前系统状态
echo ""
echo "📊 当前系统状态:"
LOAD_AVG=$(uptime | awk -F'load average:' '{ print $2 }' | awk '{ print $1 }' | sed 's/,//')
echo "CPU负载: $LOAD_AVG"
echo "Go进程数: $(pgrep -f "go run" | wc -l)"

echo ""
echo "🎯 停止操作完成!" 