package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-solana-parse/src/config"
	"github.com/go-solana-parse/src/model"
	"github.com/go-solana-parse/src/parser"
	rpccall "github.com/go-solana-parse/src/rpc_call"
	"github.com/go-solana-parse/src/solana"
)

func main() {
	//
	getData()
	// parsePerBlockData()
}

func getData() {
	config.LoadSvcConfig()
	// aws.S3UploadExample()
	// google.GetBigQueryData()

	// 支持命令行参数：go run main.go [startSlot] [endSlot] [batchSize]
	var startSlot, endSlot, batchSize int
	if len(os.Args) >= 3 {
		startSlot, _ = strconv.Atoi(os.Args[1])
		endSlot, _ = strconv.Atoi(os.Args[2])
		if len(os.Args) >= 4 {
			batchSize, _ = strconv.Atoi(os.Args[3])
		} else {
			batchSize = 10 // 默认每批10个
		}
		fmt.Printf("使用参数: slot %d-%d, 批量大小: %d\n", startSlot, endSlot, batchSize)
	} else {
		// 默认值
		rountine := 500
		startSlot = 337200528
		endSlot = 337200528 + rountine
		batchSize = 10 // 每批20个区块
	}

	// 使用新的批量请求方法
	maxConcurrency := 25 // 25个并发的批量请求

	fmt.Printf("🚀 开始批量请求模式\n")
	fmt.Printf("📊 总slot数: %d\n", endSlot-startSlot)
	fmt.Printf("📦 批量大小: %d (每次HTTP请求获取%d个区块)\n", batchSize, batchSize)
	fmt.Printf("⚡ 并发数: %d\n", maxConcurrency)
	fmt.Printf("🔢 HTTP请求总数: %d (vs 传统方式的 %d)\n",
		(endSlot-startSlot+batchSize-1)/batchSize, endSlot-startSlot)

	startTime := time.Now()

	// 调用批量请求函数
	results := solana.BatchGetBlockDataFastV2(
		uint64(startSlot),
		uint64(endSlot-1),
		"3ed35a0b-35f6-4adb-8caa-5c72cd36b023",
		batchSize,
		maxConcurrency,
	)

	elapsed := time.Since(startTime)

	// 处理结果 - 过滤包含特定token的交易
	totalFilteredTxs := 0

	fullBlockData := []model.ParseBlockDataDenoReq{}

	for slot, block := range results {
		if block == nil || len(block.Transactions) == 0 {
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

		if len(transactions) > 0 {
			totalFilteredTxs += len(transactions)
		}

		block.Transactions = transactions

		fullBlockData = append(fullBlockData, model.ParseBlockDataDenoReq{
			BlockNum:  strconv.Itoa(int(slot)),
			BlockData: *block,
		})

		// err := rpccall.SendParseDataToDeno(strconv.Itoa(int(slot)), *block)

		// endTime := time.Now()
		// fmt.Println("send parse data to deno time is ", endTime.Sub(startTime))

		// if err != nil {
		// 	log.Fatalf("Failed to send parse data to deno: %v", err)
		// }

	}

	startTime = time.Now()
	err := rpccall.SendMultipleParseDataToDeno(fullBlockData)
	if err != nil {
		log.Fatalf("Failed to send parse data to deno: %v", err)
	}

	endTime := time.Now()
	fmt.Println("send multiple parse data to deno time is ", endTime.Sub(startTime))

	fmt.Printf("\n🎯 最终统计:\n")
	fmt.Printf("✅ 成功获取: %d/%d 个区块\n", len(results), endSlot-startSlot)
	fmt.Printf("🔍 包含目标token的交易: %d\n", totalFilteredTxs)
	fmt.Printf("⏱️  总耗时: %v\n", elapsed)
	fmt.Printf("📈 速度: %.2f blocks/second\n", float64(len(results))/elapsed.Seconds())

	// 计算效率提升
	estimatedSingleRequests := endSlot - startSlot
	actualBatchRequests := (endSlot - startSlot + batchSize - 1) / batchSize
	fmt.Printf("🚀 请求效率: 减少了 %.1f%% 的HTTP请求 (%d -> %d)\n",
		float64(estimatedSingleRequests-actualBatchRequests)/float64(estimatedSingleRequests)*100,
		estimatedSingleRequests, actualBatchRequests)

}

func parsePerBlockData() {
	config.LoadSvcConfig()

	// 2. 创建处理器
	handler := parser.NewSolanaBlockDataHandler()

	slot := 337200528
	blockData, err := solana.GetBlockData(uint64(slot), "3ed35a0b-35f6-4adb-8caa-5c72cd36b023")
	if err != nil {
		log.Fatalf("Failed to get block data: %v", err)
	}

	fmt.Println("hash is ", blockData.Blockhash)

	versionedBlockResponse := model.VersionedBlockResponse(*blockData)

	// 3. 处理区块数据
	results, err := handler.HandleBlockData(versionedBlockResponse, uint64(slot))

	if err != nil {
		log.Fatalf("Failed to handle block data: %v", err)
	}

	// 自动完成：解析 → 转换 → 双写数据库
	// db.NewMySQLSwapTransactionDB(mysqlDB)
	// db.NewClickHouseSwapTransactionDB(clickhouseDB)
	// 写入json文件

	fmt.Println(len(results))

	jsonData, err := json.Marshal(results)
	if err != nil {
		log.Fatalf("Failed to marshal block data: %v", err)
	}
	os.WriteFile("block_data.json", jsonData, 0644)
}
