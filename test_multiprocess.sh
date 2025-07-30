#!/bin/bash

# 多进程循环测试脚本 - 小规模测试倒序循环功能
# 使用方法: ./test_multiprocess.sh
# 特点: 从最新区块开始处理，逐步回到旧区块

echo "🧪 多进程倒序循环测试模式"

# 测试配置 - 使用具体的区块范围进行小规模测试
START_SLOT=347797409     # 起始区块号
END_SLOT=347806409       # 结束区块号 (测试9000个区块)
CYCLE_SIZE=1000          # 每次循环1000个区块  
PROCESS_COUNT=5          # 5个进程测试
BATCH_SIZE=10            # 每次HTTP请求10个区块
PORTS_PER_PROCESS=1      # 每个进程分配1个端口

# 计算总区块数
TOTAL_BLOCKS=$((END_SLOT - START_SLOT))

echo "📊 区块范围: $START_SLOT - $END_SLOT"
echo "📊 总区块数: $TOTAL_BLOCKS 个区块"
echo "🔄 每次循环: $CYCLE_SIZE 个区块"
echo "🔢 进程数: $PROCESS_COUNT"
echo "📦 批量大小: $BATCH_SIZE"
echo "🌐 端口分配: 8000-8004 (每进程1个端口)"

# 计算统计信息
BLOCKS_PER_PROCESS=$((TOTAL_BLOCKS / PROCESS_COUNT))
CYCLES_PER_PROCESS=$(((BLOCKS_PER_PROCESS + CYCLE_SIZE - 1) / CYCLE_SIZE))
TOTAL_CYCLES=$((CYCLES_PER_PROCESS * PROCESS_COUNT))

echo "📋 每进程处理: $BLOCKS_PER_PROCESS 个区块"
echo "🔄 每进程循环数: $CYCLES_PER_PROCESS"
echo "🎯 总循环数: $TOTAL_CYCLES"

mkdir -p test_logs

START_TIME=$(date +%s)

echo "⚡ 启动测试进程..."

for i in $(seq 0 $((PROCESS_COUNT - 1))); do
    # 计算端口：进程i使用端口 8000 + i
    PORT_START=$((8000 + i * PORTS_PER_PROCESS))
    
    echo "🎬 测试进程 $i: 预计处理 $BLOCKS_PER_PROCESS 个区块，$CYCLES_PER_PROCESS 个循环 -> 端口 $PORT_START"
    
    {
        PROCESS_START_TIME=$(date +%s)
        
        # 使用新的参数格式：processId startSlot endSlot cycleSize batchSize portStart processCount
        go run ./src/main.go $i $START_SLOT $END_SLOT $CYCLE_SIZE $BATCH_SIZE $PORT_START $PROCESS_COUNT
        
        PROCESS_END_TIME=$(date +%s)
        PROCESS_DURATION=$((PROCESS_END_TIME - PROCESS_START_TIME))
        if [ $PROCESS_DURATION -gt 0 ]; then
            PROCESS_SPEED=$(echo "scale=2; $BLOCKS_PER_PROCESS / $PROCESS_DURATION" | bc -l)
        else
            PROCESS_SPEED="很快"
        fi
        echo "✅ 测试进程 $i: $BLOCKS_PER_PROCESS 区块, ${PROCESS_DURATION}s, $PROCESS_SPEED blocks/s (端口 $PORT_START)"
    } > test_logs/process_$i.log 2>&1 &
    
    # 错开启动时间
    sleep 1
done

wait

END_TIME=$(date +%s)
TOTAL_DURATION=$((END_TIME - START_TIME))

if [ $TOTAL_DURATION -gt 0 ]; then
    ACTUAL_SPEED=$(echo "scale=2; $TOTAL_BLOCKS / $TOTAL_DURATION" | bc -l)
else
    ACTUAL_SPEED="很快"
fi

echo ""
echo "🧪 循环测试完成!"
echo "⏱️  总耗时: ${TOTAL_DURATION}s"
echo "📈 测试速度: $ACTUAL_SPEED blocks/second"
echo "📊 区块范围: $START_SLOT - $END_SLOT ($TOTAL_BLOCKS 个区块)"

echo ""
echo "📋 各进程测试结果:"
for i in $(seq 0 $((PROCESS_COUNT - 1))); do
    if [ -f test_logs/process_$i.log ]; then
        echo "进程 $i:"
        tail -1 test_logs/process_$i.log
    fi
done

echo ""
echo "✅ 如果循环测试正常，可以运行大规模任务:"
echo "   ./run_continuous_multiprocess.sh"

echo ""
echo "📊 查看详细测试日志:"
echo "   tail -f test_logs/process_*.log"