package processor

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/go-solana-parse/src/config"
	"github.com/go-solana-parse/src/model"
	rpccall "github.com/go-solana-parse/src/rpc_call"
	"github.com/go-solana-parse/src/solana"
)

func ScanBlockData() {
	startTime := time.Now()

	var failedSlots []uint64

	// startSlot := uint64(263922023)
	startSlot := uint64(247806009)

	endSlot := uint64(347587512)

	cycleSize := 100

	currentBatchArr := [][]uint64{}

	runtime.GOMAXPROCS(30)

	for slot := endSlot; slot > startSlot; slot -= uint64(cycleSize) {
		currentBatch := []uint64{}

		for i := 0; i < cycleSize; i++ {
			currentBatch = append(currentBatch, slot-uint64(i))
		}

		currentBatchArr = append(currentBatchArr, currentBatch)
	}

	batchSize := 10

	// 工具函数：将二维数组切分为三维数组，每组最多30个batch
	batchesOf30 := splitToChunks(currentBatchArr, 20)
	for batchIndex, batchGroup := range batchesOf30 { // 外层串行
		startTime := time.Now()
		var wg sync.WaitGroup
		for _, currentBatch := range batchGroup { // 内层并发
			wg.Add(1)
			go func(batch []uint64) {
				defer wg.Done()
				fmt.Println("start get block data", batch[0])
				results := solana.GetMultipleBlocksData(batch, "3ed35a0b-35f6-4adb-8caa-5c72cd36b023", batchSize)

				fullBlockData := []model.ParseBlockDataDenoReq{}

				for _, slot := range batch {
					block, exists := results[slot]
					if !exists || block == nil {
						failedSlots = append(failedSlots, slot)
						continue
					}
					if len(block.Transactions) == 0 {
						continue
					}
					transactions := []model.TransactionInfo{}
					for _, transaction := range block.Transactions {
						for _, account := range transaction.Transaction.Message.AccountKeys {
							if account == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
								transactions = append(transactions, transaction)
								break
							}
						}
					}
					block.Transactions = transactions
					fullBlockData = append(fullBlockData, model.ParseBlockDataDenoReq{
						BlockNum:  strconv.Itoa(int(slot)),
						BlockData: *block,
					})
				}

				// wg2 := &sync.WaitGroup{}
				// for _, block := range fullBlockData {
				// 	wg2.Add(1)
				// 	go func(b model.ParseBlockDataDenoReq) {
				// 		defer wg2.Done()
				rpccall.SendMultipleParseDataToDeno(fullBlockData)
				// 	}(block)
				// }
				// wg2.Wait()
				fmt.Println("get block data done", batch[0])
			}(currentBatch)
		}
		wg.Wait() // 等待这一组30个全部完成
		elapsed := time.Since(startTime)
		fmt.Printf("🚀 多核处理完成 开始slot: %d 结束slot: %d 总耗时: %.1fs\n", batchGroup[batchIndex][0], batchGroup[batchIndex][len(batchGroup[batchIndex])-1], elapsed.Seconds())
	}

	fmt.Println("failedSlots ", failedSlots)

	elapsed := time.Since(startTime)
	fmt.Printf("🚀 多核处理完成: 总耗时: %.1fs\n", elapsed.Seconds())
}

func getData() {
	config.LoadSvcConfig()

	// 🚀 启用多核CPU支持
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	fmt.Printf("🔥 多核CPU模式启用: 使用 %d 个CPU核心\n", numCPU)

	// 支持命令行参数：go run main.go [processId] [startSlot] [endSlot] [cycleSize] [batchSize] [portStart] [processCount]
	// var processId, startSlot, endSlot, cycleSize, batchSize, portStart, processCount int

	timeStart := time.Now()

	startSlot := 337199528
	endSlot := startSlot + 500
	cycleSize := 100
	batchSize := 10

	processContinuousBlocksMultiCore(startSlot, endSlot, cycleSize, batchSize, numCPU)

	timeEnd := time.Now()
	fmt.Printf("🎉 多核处理完成: %d 区块, 总耗时: %.1fs\n",
		endSlot-startSlot, timeEnd.Sub(timeStart).Seconds())
}

// 多核CPU版本：持续循环处理区块
func processContinuousBlocksMultiCore(startSlot, endSlot, cycleSize, batchSize, numCPU int) {

	// 计算此进程的实际slot范围
	myStartSlot := startSlot
	myEndSlot := endSlot
	myTotalBlocks := myEndSlot - myStartSlot

	totalCycles := (myTotalBlocks + cycleSize - 1) / cycleSize

	fmt.Printf("🎬 负责区块范围: %d - %d (%d 个区块) [多核并行]\n",
		myStartSlot, myEndSlot-1, myTotalBlocks)

	overallStartTime := time.Now()

	// 🚀 多核并行处理：创建多个goroutine并行处理不同的cycle
	// 根据CPU核心数和cycle数量决定并发goroutine数量
	maxConcurrentCycles := numCPU
	if maxConcurrentCycles > totalCycles {
		maxConcurrentCycles = totalCycles
	}

	fmt.Printf("🔥 启动 %d 个并发goroutine处理 %d 个cycle\n", maxConcurrentCycles, totalCycles)

	// 创建用于传递cycle任务的channel
	type CycleTask struct {
		cycleIndex     int
		cycleStartSlot int
		cycleEndSlot   int
		actualBlocks   int
	}

	type CycleResult struct {
		cycleIndex     int
		processedCount int
		elapsed        time.Duration
		failedSlots    []uint64
		err            error
	}

	cycleTasks := make(chan CycleTask, totalCycles)
	cycleResults := make(chan CycleResult, totalCycles)

	// 启动worker goroutines
	for i := 0; i < maxConcurrentCycles; i++ {
		go func(workerID int) {
			fmt.Printf("🔧 Worker %d 启动\n", workerID)
			for task := range cycleTasks {
				cycleStartTime := time.Now()

				// 使用新的分批处理逻辑，包含失败记录
				processedCount, failedSlots := processSingleRangeHighSpeedMultiCoreWithFailureTracking(
					uint64(task.cycleStartSlot),
					uint64(task.cycleEndSlot),
					batchSize,
				)

				cycleElapsed := time.Since(cycleStartTime)

				cycleResults <- CycleResult{
					cycleIndex:     task.cycleIndex,
					processedCount: processedCount,
					elapsed:        cycleElapsed,
					failedSlots:    failedSlots,
					err:            nil,
				}

				fmt.Printf("✅ Worker %d 完成 cycle %d: %d 区块, %.1fs\n",
					workerID, task.cycleIndex+1, task.actualBlocks, cycleElapsed.Seconds())
			}
		}(i)
	}

	// 🔄 生成倒序cycle任务
	go func() {
		defer close(cycleTasks)
		for cycle := 0; cycle < totalCycles; cycle++ {
			// 倒序计算：从最后的区块开始
			cycleEndSlot := myEndSlot - cycle*cycleSize
			cycleStartSlot := cycleEndSlot - cycleSize

			// 确保不超过此进程的范围
			if cycleStartSlot < myStartSlot {
				cycleStartSlot = myStartSlot
			}

			actualBlocks := cycleEndSlot - cycleStartSlot

			// 如果没有区块需要处理，跳过
			if actualBlocks <= 0 {
				continue
			}

			cycleTasks <- CycleTask{
				cycleIndex:     cycle,
				cycleStartSlot: cycleStartSlot,
				cycleEndSlot:   cycleEndSlot,
				actualBlocks:   actualBlocks,
			}
		}
	}()

	// 收集结果
	totalProcessedBlocks := 0
	completedCycles := 0
	var allFailedSlots []uint64

	for result := range cycleResults {
		totalProcessedBlocks += result.processedCount
		allFailedSlots = append(allFailedSlots, result.failedSlots...)
		completedCycles++

		// 🧠 内存监控（每5个cycle）
		if completedCycles%5 == 0 {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			fmt.Printf("💾 内存状态: %.1f MB (Heap: %.1f MB)\n",
				float64(memStats.Sys)/1024/1024, float64(memStats.HeapAlloc)/1024/1024)
		}

		// 🧹 垃圾回收（每10个cycle）
		if completedCycles%10 == 0 {
			runtime.GC()
			fmt.Printf("🧹 内存清理完成\n")
		}

		// 计算进度
		progress := float64(completedCycles) / float64(totalCycles) * 100
		overallElapsed := time.Since(overallStartTime)

		if completedCycles == totalCycles {
			break
		}

		fmt.Printf("📈 多核进度: %.1f%% (%d/%d), 已用时: %.1fm\n",
			progress, completedCycles, totalCycles, overallElapsed.Minutes())
	}

	// 保存失败的区块到文件
	if len(allFailedSlots) > 0 {
		saveFailedSlotsToFile(allFailedSlots, uint64(myStartSlot), uint64(myEndSlot))
	}

	overallElapsed := time.Since(overallStartTime)
	fmt.Printf("🎉 多核处理完成: %d 区块, 失败: %d 区块, 总耗时: %.1fm\n",
		totalProcessedBlocks, len(allFailedSlots), overallElapsed.Minutes())
}

// 多核优化版本的处理函数（带失败跟踪）
func processSingleRangeHighSpeedMultiCoreWithFailureTracking(startSlot, endSlot uint64, batchSize int) (int, []uint64) {
	// 创建失败记录
	var failedSlots []uint64
	totalProcessedBlocks := 0
	totalFilteredTxs := 0

	// 🎯 每50个区块为一个小批次（避免内存过大）
	const smallCycleSize = 50
	totalSlots := endSlot - startSlot

	// 🔄 创建倒序的 slot 数组
	var reversedSlots []uint64
	for i := uint64(0); i < totalSlots; i++ {
		slot := endSlot - 1 - i
		reversedSlots = append(reversedSlots, slot)
	}

	// 分小批次处理
	for i := 0; i < len(reversedSlots); i += smallCycleSize {
		batchEnd := i + smallCycleSize
		if batchEnd > len(reversedSlots) {
			batchEnd = len(reversedSlots)
		}

		currentBatch := reversedSlots[i:batchEnd]

		fmt.Printf("🚀 多核处理: %d - %d, batchSize: %d\n", currentBatch[0], currentBatch[len(currentBatch)-1], batchSize)

		// 获取这一小批的区块数据
		results := solana.GetMultipleBlocksData(currentBatch, "3ed35a0b-35f6-4adb-8caa-5c72cd36b023", batchSize)

		// 处理结果
		var fullBlockData []model.ParseBlockDataDenoReq
		batchProcessedBlocks := 0
		batchFilteredTxs := 0

		for _, slot := range currentBatch {
			block, exists := results[slot]
			if !exists || block == nil {
				// 记录获取失败的区块
				failedSlots = append(failedSlots, slot)
				continue
			}

			// 检查区块是否为空
			if len(block.Transactions) == 0 {
				batchProcessedBlocks++
				continue
			}

			// 过滤包含特定token的交易
			transactions := []model.TransactionInfo{}
			for _, transaction := range block.Transactions {
				for _, account := range transaction.Transaction.Message.AccountKeys {
					if account == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
						transactions = append(transactions, transaction)
						break
					}
				}
			}

			if len(transactions) > 0 {
				batchFilteredTxs += len(transactions)
			}

			block.Transactions = transactions

			fullBlockData = append(fullBlockData, model.ParseBlockDataDenoReq{
				BlockNum:  strconv.Itoa(int(slot)),
				BlockData: *block,
			})

			batchProcessedBlocks++
		}

		// 发送这一小批数据到RPC并等待返回
		if len(fullBlockData) > 0 {
			err := rpccall.SendMultipleParseDataToDeno(fullBlockData)
			if err != nil {
				// 将这一小批的所有区块都标记为失败
				for _, data := range fullBlockData {
					if slotInt, parseErr := strconv.ParseUint(data.BlockNum, 10, 64); parseErr == nil {
						failedSlots = append(failedSlots, slotInt)
					}
				}
			}
		}

		totalProcessedBlocks += batchProcessedBlocks
		totalFilteredTxs += batchFilteredTxs

		// 清理内存
		results = nil
		fullBlockData = nil
	}

	return totalProcessedBlocks, failedSlots
}

// saveFailedSlotsToFile 保存失败的区块到文件
func saveFailedSlotsToFile(failedSlots []uint64, startSlot, endSlot uint64) {
	if len(failedSlots) == 0 {
		return
	}

	// 生成文件名，包含时间戳和区块范围
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("failed_slots_%d_%d_%s.txt", startSlot, endSlot, timestamp)

	// 创建文件
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("❌ 创建失败记录文件失败: %v\n", err)
		return
	}
	defer file.Close()

	// 写入文件头信息
	file.WriteString(fmt.Sprintf("# 失败区块记录\n"))
	file.WriteString(fmt.Sprintf("# 处理时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	file.WriteString(fmt.Sprintf("# 区块范围: %d - %d\n", startSlot, endSlot))
	file.WriteString(fmt.Sprintf("# 失败区块数: %d\n", len(failedSlots)))
	file.WriteString("# ==========================================\n")

	// 写入失败的区块号
	for _, slot := range failedSlots {
		file.WriteString(fmt.Sprintf("%d\n", slot))
	}

	fmt.Printf("📝 失败区块已保存到文件: %s (%d 个区块)\n", filename, len(failedSlots))
}

// 工具函数：将二维数组切分为三维数组，每组最多30个batch
func splitToChunks(arr [][]uint64, chunkSize int) [][][]uint64 {
	var chunks [][][]uint64
	for i := 0; i < len(arr); i += chunkSize {
		end := i + chunkSize
		if end > len(arr) {
			end = len(arr)
		}
		chunks = append(chunks, arr[i:end])
	}
	return chunks
}
