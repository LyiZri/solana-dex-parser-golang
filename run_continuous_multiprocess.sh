#!/bin/bash

# 大规模连续处理脚本 - 基于起始和结束区块 (倒序处理)
# 使用方法: ./run_continuous_multiprocess.sh
# 特点: 从最新区块开始处理，逐步回到旧区块

echo "🚀 大规模倒序连续处理模式"

# 生产配置 - 使用具体的区块范围（CPU友好）
START_SLOT=263922023     # 起始区块号
END_SLOT=351923058       # 结束区块号 (示例：处理3000个区块)
CYCLE_SIZE=100          # 每次循环1100个区块  
PROCESS_COUNT=50          # 降低到5个进程，减少CPU压力
BATCH_SIZE=10            # 每次HTTP请求10个区块
PORTS_PER_PROCESS=1      # 每个进程分配1个端口

# 计算总区块数
TOTAL_BLOCKS=$((END_SLOT - START_SLOT))

echo "📊 区块范围: $START_SLOT - $END_SLOT"
echo "📊 总区块数: $TOTAL_BLOCKS 个区块 ($(echo "scale=1; $TOTAL_BLOCKS / 1000000" | bc -l)M)"
echo "🔄 每次循环: $CYCLE_SIZE 个区块"
echo "🔢 进程数: $PROCESS_COUNT (CPU友好)"
echo "📦 批量大小: $BATCH_SIZE"
echo "🌐 端口分配: 8000-8004 (每进程1个端口)"

# 计算统计信息
BLOCKS_PER_PROCESS=$((TOTAL_BLOCKS / PROCESS_COUNT))
CYCLES_PER_PROCESS=$(((BLOCKS_PER_PROCESS + CYCLE_SIZE - 1) / CYCLE_SIZE))
TOTAL_CYCLES=$((CYCLES_PER_PROCESS * PROCESS_COUNT))

echo "📋 每进程处理: $BLOCKS_PER_PROCESS 个区块"
echo "🔄 每进程循环数: $CYCLES_PER_PROCESS"
echo "🎯 总循环数: $TOTAL_CYCLES"

# 预估时间（基于测试数据）
ESTIMATED_BLOCKS_PER_SECOND=25  # 降低预期速度，更加保守
ESTIMATED_TOTAL_HOURS=$(echo "scale=1; $TOTAL_BLOCKS / $ESTIMATED_BLOCKS_PER_SECOND / 3600" | bc -l)
echo "⏰ 预估总耗时: ~${ESTIMATED_TOTAL_HOURS} 小时 (CPU友好模式)"

echo ""
echo "⚠️  请确认区块范围："
echo "   起始: $START_SLOT"
echo "   结束: $END_SLOT" 
echo "   总计: $TOTAL_BLOCKS 个区块"
echo ""
read -p "确认开始处理? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ 已取消"
    exit 1
fi

START_TIME=$(date +%s)

echo "⚡ 启动连续处理进程..."

for i in $(seq 0 $((PROCESS_COUNT - 1))); do
    # 计算端口：进程i使用端口 8000 + i
    PORT_START=$((8000 + i * PORTS_PER_PROCESS))
    
    echo "🎬 启动进程 $i: 端口 $PORT_START"
    
    # 直接在后台运行，不重定向到日志文件
    go run ./src/main.go $i $START_SLOT $END_SLOT $CYCLE_SIZE $BATCH_SIZE $PORT_START $PROCESS_COUNT &
    
    # 📊 检查CPU负载，如果过高则暂停
    LOAD_AVG=$(uptime | awk -F'load average:' '{ print $2 }' | awk '{ print $1 }' | sed 's/,//')
    LOAD_THRESHOLD=10.0
    if (( $(echo "$LOAD_AVG > $LOAD_THRESHOLD" | bc -l) )); then
        echo "⚠️  CPU负载过高 ($LOAD_AVG)，等待1秒..."
        sleep 0.1
    fi
done

echo ""
echo "🔄 所有进程已启动，开始连续处理..."
echo "⏹️  停止处理: ./stop_all_processes.sh"

# 简化监控 - 不依赖日志文件
echo ""
echo "📈 实时监控 (每60秒更新一次):"
while true; do
    # 检查是否还有进程在运行
    if ! pgrep -f "go run ./src/main.go" > /dev/null; then
        echo "🎉 所有进程已完成!"
        break
    fi
    
    # 显示进度摘要
    CURRENT_TIME=$(date +%s)
    ELAPSED_TIME=$((CURRENT_TIME - START_TIME))
    ELAPSED_HOURS=$(echo "scale=2; $ELAPSED_TIME / 3600" | bc -l)
    
    # 📊 获取系统状态
    LOAD_AVG=$(uptime | awk -F'load average:' '{ print $2 }' | awk '{ print $1 }' | sed 's/,//')
    RUNNING_PROCESSES=$(pgrep -f "go run ./src/main.go" | wc -l)
    
    echo "$(date '+%H:%M:%S') - 运行: ${ELAPSED_HOURS}h | 活跃进程: $RUNNING_PROCESSES/$PROCESS_COUNT | CPU负载: $LOAD_AVG"
    
    sleep 60
done

END_TIME=$(date +%s)
TOTAL_DURATION=$((END_TIME - START_TIME))
TOTAL_HOURS=$(echo "scale=2; $TOTAL_DURATION / 3600" | bc -l)

echo ""
echo "🎉 大规模处理完成!"
echo "⏱️  总耗时: ${TOTAL_HOURS} 小时"
echo "📊 处理速度: $(echo "scale=1; $TOTAL_BLOCKS / $TOTAL_DURATION" | bc -l) blocks/second"
echo "📊 区块范围: $START_SLOT - $END_SLOT ($TOTAL_BLOCKS 个区块)"

echo ""
echo "🎯 所有进程已完成处理！" 