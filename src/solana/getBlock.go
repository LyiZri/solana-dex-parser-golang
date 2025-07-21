// Package solana provides functions to interact with Solana blockchain using Helius RPC
// This module specifically handles block data retrieval using the getBlock method.
//
// Usage:
//
//	import "github.com/go-solana-parse/src/solana"
//
//	ctx := context.Background()
//	block, err := solana.GetBlock(ctx, 250000000)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Block hash: %s\n", block.Blockhash)
package solana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-solana-parse/src/model"
	"golang.org/x/time/rate"
)

// Global HTTP client optimized for high concurrency
var (
	httpClient     *http.Client
	clientOnce     sync.Once
	rateLimiter    *rate.Limiter
	highPerfClient *http.Client
)

// initHTTPClient initializes the optimized HTTP client
func initHTTPClient() {
	// Create rate limiter: 500 requests per second with burst capacity of 100
	rateLimiter = rate.NewLimiter(rate.Limit(500), 100)

	// Configure transport for high concurrency
	transport := &http.Transport{
		MaxIdleConns:        1000,             // Maximum idle connections
		MaxIdleConnsPerHost: 100,              // Maximum idle connections per host
		MaxConnsPerHost:     200,              // Maximum connections per host
		IdleConnTimeout:     90 * time.Second, // How long idle connections stay open
		DisableCompression:  false,            // Enable compression
		DisableKeepAlives:   false,            // Enable keep-alives
		WriteBufferSize:     32 * 1024,        // 32KB write buffer
		ReadBufferSize:      32 * 1024,        // 32KB read buffer
	}

	httpClient = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second, // Total request timeout
	}
}

// getHTTPClient returns the singleton HTTP client
func getHTTPClient() *http.Client {
	clientOnce.Do(initHTTPClient)
	return httpClient
}

// getHighPerfClient 获取高性能HTTP客户端（单例）
func getHighPerfClient() *http.Client {
	clientOnce.Do(func() {
		// 自定义Transport，大幅增加连接池
		transport := &http.Transport{
			// 连接池配置
			MaxIdleConns:        1000,             // 最大空闲连接数
			MaxIdleConnsPerHost: 500,              // 每个host的最大空闲连接数
			MaxConnsPerHost:     500,              // 每个host的最大连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时

			// TCP连接优化
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,  // 连接超时
				KeepAlive: 30 * time.Second, // keep-alive
			}).DialContext,

			// HTTP/2和TLS优化
			ForceAttemptHTTP2:     true,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,

			// 禁用压缩以减少CPU使用
			DisableCompression: true,
		}

		highPerfClient = &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second, // 总超时时间
		}
	})
	return highPerfClient
}

// GetBlockData 获取指定slot的区块交易数据 - 高性能版本
func GetBlockData(slotNum uint64, apiKey string) (*model.Block, error) {
	// 使用高性能客户端
	client := getHighPerfClient()

	// 构建请求URL
	url := fmt.Sprintf("https://mainnet.helius-rpc.com/?api-key=%s", apiKey)

	// 构建请求参数
	params := []interface{}{
		slotNum,
		map[string]interface{}{
			"maxSupportedTransactionVersion": 0,
			"transactionDetails":             "full",
			"encoding":                       "json",
			"rewards":                        false,
		},
	}

	// 构建请求体
	requestBody := model.RPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "getBlock",
		Params:  params,
	}

	// 将请求体转换为JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive") // 强制keep-alive

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// 解析响应
	var response model.RPCResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// 检查RPC错误
	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: %s (code: %d)", response.Error.Message, response.Error.Code)
	}

	// 检查是否找到区块
	if response.Result == nil {
		return nil, fmt.Errorf("block not found for slot %d", slotNum)
	}

	return response.Result, nil
}

// BatchGetBlockRequest 批量getBlock请求结构
type BatchGetBlockRequest []model.RPCRequest

// BatchGetBlockResponse 批量getBlock响应结构
type BatchGetBlockResponse []model.GetBlockResponse

// GetMultipleBlocksData 批量获取多个区块数据
// slotNums: 要获取的slot数组
// apiKey: Helius API密钥
// batchSize: 每次批量请求的数量(建议10-20)
func GetMultipleBlocksData(slotNums []uint64, apiKey string, batchSize int) map[uint64]*model.Block {
	results := make(map[uint64]*model.Block)
	client := getHighPerfClient()

	wg := sync.WaitGroup{}

	// 分批处理
	for i := 0; i < len(slotNums); i += batchSize {
		end := i + batchSize
		if end > len(slotNums) {
			end = len(slotNums)
		}

		batch := slotNums[i:end]

		wg.Add(1)
		go func() {
			defer wg.Done()

			batchResults := processBatch(batch, apiKey, client)

			// 合并结果
			for slot, block := range batchResults {
				results[slot] = block
			}
		}()
	}

	wg.Wait()

	return results
}

// processBatch 处理一批区块请求
func processBatch(slotNums []uint64, apiKey string, client *http.Client) map[uint64]*model.Block {
	if len(slotNums) == 0 {
		return make(map[uint64]*model.Block)
	}

	// 构建批量请求
	var batchRequest BatchGetBlockRequest
	for i, slotNum := range slotNums {
		request := model.RPCRequest{
			JSONRPC: "2.0",
			ID:      i + 1, // 每个请求需要唯一ID
			Method:  "getBlock",
			Params: []interface{}{
				slotNum,
				map[string]interface{}{
					"maxSupportedTransactionVersion": 0,
					"transactionDetails":             "full",
					"encoding":                       "json",
					"rewards":                        false,
				},
			},
		}
		batchRequest = append(batchRequest, request)
	}

	// 发送批量请求
	url := fmt.Sprintf("https://mainnet.helius-rpc.com/?api-key=%s", apiKey)

	jsonData, err := json.Marshal(batchRequest)

	if err != nil {
		return make(map[uint64]*model.Block)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return make(map[uint64]*model.Block)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("发送批量请求失败: %v\n", err)
		return make(map[uint64]*model.Block)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取批量响应失败: %v\n", err)
		return make(map[uint64]*model.Block)
	}

	// 解析批量响应
	var batchResponse BatchGetBlockResponse
	if err := json.Unmarshal(body, &batchResponse); err != nil {
		fmt.Printf("解析批量响应失败: %v\n", err)
		fmt.Println("body ", string(body))
		return make(map[uint64]*model.Block)
	}

	// 处理结果
	results := make(map[uint64]*model.Block)
	for i, response := range batchResponse {
		if i >= len(slotNums) {
			continue
		}

		slotNum := slotNums[i]

		if response.Error != nil {
			continue
		}

		if response.Result != nil {
			results[slotNum] = response.Result
		}
	}

	return results
}

func BatchGetBlockDataFastV2(startSlot, endSlot uint64, apiKey string, batchSize, maxConcurrency int) map[uint64]*model.Block {
	// 生成所有要请求的slot
	var allSlots []uint64
	for slot := startSlot; slot <= endSlot; slot++ {
		allSlots = append(allSlots, slot)
	}

	results := make(map[uint64]*model.Block)

	// 按batchSize分组
	var batches [][]uint64
	for i := 0; i < len(allSlots); i += batchSize {
		end := i + batchSize
		if end > len(allSlots) {
			end = len(allSlots)
		}
		batches = append(batches, allSlots[i:end])
	}

	fmt.Printf("📦 创建 %d 个批次，每批 %d 个slot\n", len(batches), batchSize)

	startTime := time.Now()

	for i, batch := range batches {

		// 处理这一批
		batchResults := GetMultipleBlocksData(batch, apiKey, batchSize)

		// 合并结果
		for slot, block := range batchResults {
			results[slot] = block
		}
		fmt.Printf("✅ 批次 %d 完成: %d/%d 个slot成功\n",
			i, len(batchResults), len(batch))
	}

	elapsed := time.Since(startTime)

	fmt.Printf("\n🎯 批量请求完成!\n")
	fmt.Printf("总耗时: %v\n", elapsed)
	fmt.Printf("成功率: %.2f%% (%d/%d)\n",
		float64(len(results))/float64(len(allSlots))*100, len(results), len(allSlots))
	fmt.Printf("平均速度: %.2f blocks/second\n", float64(len(results))/elapsed.Seconds())
	fmt.Printf("HTTP请求数: %d (vs %d)\n", len(batches), len(allSlots))

	return results
}
