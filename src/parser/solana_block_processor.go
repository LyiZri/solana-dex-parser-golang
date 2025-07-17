package parser

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-solana-parse/src/db"
	"github.com/go-solana-parse/src/model"
)

// SolanaBlockProcessor Solana 区块处理器
type SolanaBlockProcessor struct {
	blockDataHandler *SolanaBlockDataHandler
	database         db.SwapTransactionDB
}

// ProcessorConfig 处理器配置
type ProcessorConfig struct {
	EnableDatabase  bool
	EnableFilter    bool
	BatchSize       int
	ProcessingDelay time.Duration
}

// DefaultProcessorConfig 默认处理器配置
func DefaultProcessorConfig() ProcessorConfig {
	return ProcessorConfig{
		EnableDatabase:  true,
		EnableFilter:    true,
		BatchSize:       100,
		ProcessingDelay: time.Millisecond * 100,
	}
}

// NewSolanaBlockProcessor 创建新的区块处理器
func NewSolanaBlockProcessor(database db.SwapTransactionDB) *SolanaBlockProcessor {
	return &SolanaBlockProcessor{
		blockDataHandler: NewSolanaBlockDataHandler(),
		database:         database,
	}
}

// ProcessBlock 处理单个区块 - 主要入口点
func (sbp *SolanaBlockProcessor) ProcessBlock(ctx context.Context, blockData model.Block, blockNumber uint64, config ProcessorConfig) (*BlockProcessResult, error) {
	startTime := time.Now()

	log.Printf("Processing block %d with %d transactions", blockNumber, len(blockData.Transactions))

	// 1. 使用 SolanaBlockDataHandler 处理区块数据
	swapTransactions, err := sbp.blockDataHandler.HandleBlockData(blockData, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to handle block data: %w", err)
	}

	result := &BlockProcessResult{
		BlockNumber:          blockNumber,
		ProcessingTime:       time.Since(startTime),
		TotalTransactions:    len(blockData.Transactions),
		ValidTransactions:    len(swapTransactions),
		FilteredTransactions: 0,
		DatabaseInserted:     0,
		Success:              true,
	}

	if len(swapTransactions) == 0 {
		log.Printf("No valid swap transactions found in block %d", blockNumber)
		return result, nil
	}

	// 2. 数据库存储（如果启用）
	if config.EnableDatabase && sbp.database != nil {
		err := sbp.database.InsertSwapTransactions(ctx, swapTransactions)
		if err != nil {
			log.Printf("Failed to insert transactions for block %d: %v", blockNumber, err)
			result.Success = false
			result.Error = err.Error()
		} else {
			result.DatabaseInserted = len(swapTransactions)
		}
	}

	// 3. 应用过滤器（如果启用）
	if config.EnableFilter {
		filteredData := sbp.applyFilters(swapTransactions)
		result.FilteredTransactions = len(filteredData)
		result.FilteredData = filteredData
	}

	log.Printf("Block %d processed: %d total, %d valid, %d filtered in %v",
		blockNumber, result.TotalTransactions, result.ValidTransactions,
		result.FilteredTransactions, result.ProcessingTime)

	return result, nil
}

// ProcessBlocks 批量处理多个区块
func (sbp *SolanaBlockProcessor) ProcessBlocks(ctx context.Context, blocks []BlockData, config ProcessorConfig) (*BatchProcessResult, error) {
	startTime := time.Now()

	batchResult := &BatchProcessResult{
		StartTime:         startTime,
		TotalBlocks:       len(blocks),
		ProcessedBlocks:   0,
		SuccessfulBlocks:  0,
		FailedBlocks:      0,
		TotalTransactions: 0,
		ValidTransactions: 0,
		BlockResults:      make([]BlockProcessResult, 0, len(blocks)),
	}

	for i, blockData := range blocks {
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled, stopping batch processing at block %d/%d", i, len(blocks))
			return batchResult, ctx.Err()
		default:
		}

		result, err := sbp.ProcessBlock(ctx, blockData.Block, blockData.BlockNumber, config)
		if err != nil {
			log.Printf("Failed to process block %d: %v", blockData.BlockNumber, err)
			batchResult.FailedBlocks++
			continue
		}

		batchResult.ProcessedBlocks++
		if result.Success {
			batchResult.SuccessfulBlocks++
		} else {
			batchResult.FailedBlocks++
		}

		batchResult.TotalTransactions += result.TotalTransactions
		batchResult.ValidTransactions += result.ValidTransactions
		batchResult.BlockResults = append(batchResult.BlockResults, *result)

		// 添加处理延迟以避免过载
		if config.ProcessingDelay > 0 {
			time.Sleep(config.ProcessingDelay)
		}
	}

	batchResult.ProcessingTime = time.Since(startTime)

	log.Printf("Batch processing completed: %d/%d blocks successful in %v",
		batchResult.SuccessfulBlocks, batchResult.TotalBlocks, batchResult.ProcessingTime)

	return batchResult, nil
}

// GetTokenData 获取代币数据（带过滤）
func (sbp *SolanaBlockProcessor) GetTokenData(ctx context.Context, filters db.TokenDataFilter) ([]model.TokenSwapFilterData, error) {
	if sbp.database == nil {
		return nil, fmt.Errorf("database not configured")
	}

	transactions, err := sbp.database.GetFilteredTokenData(ctx, filters)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

// applyFilters 应用过滤器逻辑
func (sbp *SolanaBlockProcessor) applyFilters(transactions []model.SwapTransaction) []model.TokenSwapFilterData {
	// 转换为 SwapTransactionToken 格式
	var tokenTransactions []model.SwapTransactionToken
	for _, tx := range transactions {
		tokenTx := model.SwapTransactionToken{
			TxHash:          tx.TxHash,
			TradeType:       tx.TradeType,
			PoolAddress:     tx.PoolAddress,
			BlockHeight:     tx.BlockHeight,
			TransactionTime: tx.TransactionTime,
			WalletAddress:   tx.WalletAddress,
			TokenAmount:     tx.TokenAmount,
			TokenSymbol:     tx.TokenSymbol,
			TokenAddress:    tx.TokenAddress,
			QuoteSymbol:     tx.QuoteSymbol,
			QuoteAmount:     tx.QuoteAmount,
			QuoteAddress:    tx.QuoteAddress,
			QuotePrice:      parseStringToFloat(tx.QuotePrice),
			USDPrice:        parseStringToFloat(tx.USDPrice),
			USDAmount:       parseStringToFloat(tx.USDAmount),
		}
		tokenTransactions = append(tokenTransactions, tokenTx)
	}

	// 使用 SolanaBlockDataHandler 的过滤逻辑
	return sbp.blockDataHandler.FilterTokenData(tokenTransactions)
}

// GetStatistics 获取处理统计信息
func (sbp *SolanaBlockProcessor) GetStatistics(ctx context.Context) (*ProcessorStatistics, error) {
	if sbp.database == nil {
		return nil, fmt.Errorf("database not configured")
	}

	// 获取最近的统计数据
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour) // 最近24小时

	filters := db.SwapTransactionFilter{
		StartTime: &startTime,
		EndTime:   &endTime,
		Limit:     1000,
	}

	transactions, err := sbp.database.GetSwapTransactions(ctx, filters)
	if err != nil {
		return nil, err
	}

	stats := &ProcessorStatistics{
		TotalTransactions: len(transactions),
		Last24Hours:       len(transactions),
		UpdatedAt:         time.Now(),
	}

	// 计算额外统计信息
	uniqueWallets := make(map[string]bool)
	uniqueTokens := make(map[string]bool)
	totalVolume := 0.0

	for _, tx := range transactions {
		uniqueWallets[tx.WalletAddress] = true
		uniqueTokens[tx.TokenAddress] = true
		totalVolume += tx.USDAmount
	}

	stats.UniqueWallets = len(uniqueWallets)
	stats.UniqueTokens = len(uniqueTokens)
	stats.TotalVolumeUSD = totalVolume

	return stats, nil
}

// 结果结构定义

// BlockData 区块数据结构
type BlockData struct {
	Block       model.Block
	BlockNumber uint64
}

// BlockProcessResult 单个区块处理结果
type BlockProcessResult struct {
	BlockNumber          uint64                      `json:"block_number"`
	ProcessingTime       time.Duration               `json:"processing_time"`
	TotalTransactions    int                         `json:"total_transactions"`
	ValidTransactions    int                         `json:"valid_transactions"`
	FilteredTransactions int                         `json:"filtered_transactions"`
	DatabaseInserted     int                         `json:"database_inserted"`
	Success              bool                        `json:"success"`
	Error                string                      `json:"error,omitempty"`
	FilteredData         []model.TokenSwapFilterData `json:"filtered_data,omitempty"`
}

// BatchProcessResult 批处理结果
type BatchProcessResult struct {
	StartTime         time.Time            `json:"start_time"`
	ProcessingTime    time.Duration        `json:"processing_time"`
	TotalBlocks       int                  `json:"total_blocks"`
	ProcessedBlocks   int                  `json:"processed_blocks"`
	SuccessfulBlocks  int                  `json:"successful_blocks"`
	FailedBlocks      int                  `json:"failed_blocks"`
	TotalTransactions int                  `json:"total_transactions"`
	ValidTransactions int                  `json:"valid_transactions"`
	BlockResults      []BlockProcessResult `json:"block_results"`
}

// ProcessorStatistics 处理器统计信息
type ProcessorStatistics struct {
	TotalTransactions int       `json:"total_transactions"`
	Last24Hours       int       `json:"last_24_hours"`
	UniqueWallets     int       `json:"unique_wallets"`
	UniqueTokens      int       `json:"unique_tokens"`
	TotalVolumeUSD    float64   `json:"total_volume_usd"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// 工具函数

func parseStringToFloat(s string) float64 {
	// TODO: 实现安全的字符串到浮点数转换
	return 0.0
}

// 全局处理器实例（可选）
var GlobalBlockProcessor *SolanaBlockProcessor
