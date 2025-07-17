#!/bin/bash

# 多进程测试脚本 - 小规模测试
# 使用方法: ./test_multiprocess.sh

echo "🧪 多进程测试模式"

# 测试配置 - 较小的数量用于快速测试
# TOTAL_BLOCKS=4000        # 测试100个区块
# START_SLOT=337200528    # 起始slot
# PROCESS_COUNT=25         # 5个进程测试
# BATCH_SIZE=10           # 每批10个
TOTAL_BLOCKS=1000        # 测试100个区块
START_SLOT=347797409    # 起始slot
PROCESS_COUNT=5         # 5个进程测试
BATCH_SIZE=10          # 每批10个

BLOCKS_PER_PROCESS=$((TOTAL_BLOCKS / PROCESS_COUNT))

echo "📊 测试目标: $TOTAL_BLOCKS 个区块"
echo "🔢 进程数: $PROCESS_COUNT"
echo "📦 每进程: $BLOCKS_PER_PROCESS 个区块"
echo "📋 批量大小: $BATCH_SIZE"

mkdir -p test_logs

START_TIME=$(date +%s)

echo "⚡ 启动测试进程..."

for i in $(seq 0 $((PROCESS_COUNT - 1))); do
    PROC_START_SLOT=$((START_SLOT + i * BLOCKS_PER_PROCESS))
    PROC_END_SLOT=$((PROC_START_SLOT + BLOCKS_PER_PROCESS))
    
    if [ $i -eq $((PROCESS_COUNT - 1)) ]; then
        PROC_END_SLOT=$((START_SLOT + TOTAL_BLOCKS))
    fi
    
    ACTUAL_BLOCKS=$((PROC_END_SLOT - PROC_START_SLOT))
    
    echo "🎬 测试进程 $i: slot $PROC_START_SLOT-$((PROC_END_SLOT - 1)) ($ACTUAL_BLOCKS 区块)"
    
    {
        PROCESS_START_TIME=$(date +%s)
        go run ./src/main.go $PROC_START_SLOT $PROC_END_SLOT $BATCH_SIZE
        PROCESS_END_TIME=$(date +%s)
        PROCESS_DURATION=$((PROCESS_END_TIME - PROCESS_START_TIME))
        if [ $PROCESS_DURATION -gt 0 ]; then
            PROCESS_SPEED=$(echo "scale=2; $ACTUAL_BLOCKS / $PROCESS_DURATION" | bc -l)
        else
            PROCESS_SPEED="很快"
        fi
        echo "✅ 测试进程 $i: $ACTUAL_BLOCKS 区块, ${PROCESS_DURATION}s, $PROCESS_SPEED blocks/s"
    } > test_logs/process_$i.log 2>&1 &
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
echo "🧪 测试完成!"
echo "⏱️  总耗时: ${TOTAL_DURATION}s"
echo "📈 测试速度: $ACTUAL_SPEED blocks/second"

echo ""
echo "📋 各进程测试结果:"
for i in $(seq 0 $((PROCESS_COUNT - 1))); do
    if [ -f test_logs/process_$i.log ]; then
        echo "进程 $i:"
        tail -1 test_logs/process_$i.log
    fi
done

echo ""
echo "✅ 如果测试正常，可以运行大规模任务:"
echo "   ./run_multiprocess.sh"