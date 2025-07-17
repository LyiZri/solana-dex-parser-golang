package parser

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/go-solana-parse/src/config"
	"github.com/go-solana-parse/src/model"
)

// SolanaBlockDataHandler Solana 区块数据处理器
type SolanaBlockDataHandler struct {
	dexParser *DexParser
}

// NewSolanaBlockDataHandler 创建新的区块数据处理器
func NewSolanaBlockDataHandler() *SolanaBlockDataHandler {
	return &SolanaBlockDataHandler{
		dexParser: NewDexParser(),
	}
}

// HandleBlockData 处理区块数据 - 对应 TS 版本的 handleBlockData 方法
func (handler *SolanaBlockDataHandler) HandleBlockData(blockData model.VersionedBlockResponse, blockNumber uint64) ([]model.SwapTransaction, error) {
	log.Printf("SolanaBlockDataHandler.HandleBlockData blockNumber: %d", blockNumber)

	// 1. 调用 DexParser 解析区块数据
	parseResult, err := handler.dexParser.ParseBlockData(blockData, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to parse block data: %w", err)
	}

	// 2. 过滤出包含交易的结果
	var filteredTransactions []model.ParseResult
	for _, tx := range parseResult {
		if len(tx.Result.Trades) > 0 && len(tx.Trades) > 0 {
			filteredTransactions = append(filteredTransactions, tx)
		}
	}

	// 3. 转换数据格式并收集
	var swapTransactionArray []model.SwapTransaction
	for _, tx := range filteredTransactions {
		for tradeIndex := range tx.Trades {
			swapTransaction, err := handler.ConvertData(tx, tradeIndex, blockNumber)
			if err != nil {
				log.Printf("SolanaBlockDataHandler.ConvertData error: %v", err)
				continue
			}
			if swapTransaction != nil {
				swapTransactionArray = append(swapTransactionArray, *swapTransaction)
			}
		}
	}

	log.Printf("Processed %d swap transactions from block %d", len(swapTransactionArray), blockNumber)
	return swapTransactionArray, nil
}

// ConvertData 转换数据 - 对应 TS 版本的 convertData 方法
func (handler *SolanaBlockDataHandler) ConvertData(parseResult model.ParseResult, index int, blockNumber uint64) (*model.SwapTransaction, error) {
	if index >= len(parseResult.Result.Trades) || index >= len(parseResult.Trades) {
		return nil, fmt.Errorf("index out of range")
	}

	tradeDetail := parseResult.Result.Trades[index]
	tradeType := parseResult.Trades[index].Type

	txHash := tradeDetail.TransactionSignature
	transactionTime := tradeDetail.BlockTime
	walletAddress := tradeDetail.UserAddress

	var tokenAmount, tokenSymbol, tokenAddress string
	var quoteSymbol, quoteAmount, quoteAddress string
	var poolAddress string = tradeDetail.PoolAddress

	// 根据交易类型确定输入输出
	if tradeType == model.TradeTypeBuy {
		tokenAmount = formatFloat(tradeDetail.TokenOutAmount)
		tokenSymbol = tradeDetail.TokenOutSymbol
		tokenAddress = tradeDetail.TokenOutMint
		quoteSymbol = tradeDetail.TokenInSymbol
		quoteAmount = formatFloat(tradeDetail.TokenInAmount)
		quoteAddress = tradeDetail.TokenInMint
	} else {
		tokenAmount = formatFloat(tradeDetail.TokenInAmount)
		tokenSymbol = tradeDetail.TokenInSymbol
		tokenAddress = tradeDetail.TokenInMint
		quoteSymbol = tradeDetail.TokenOutSymbol
		quoteAmount = formatFloat(tradeDetail.TokenOutAmount)
		quoteAddress = tradeDetail.TokenOutMint
	}

	// 计算基础价格
	tokenAmountFloat, err := strconv.ParseFloat(tokenAmount, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token amount: %v", err)
	}

	quoteAmountFloat, err := strconv.ParseFloat(quoteAmount, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse quote amount: %v", err)
	}

	if tokenAmountFloat == 0 {
		return nil, fmt.Errorf("token amount is zero")
	}

	quotePrice := quoteAmountFloat / tokenAmountFloat
	quotePriceStr := strconv.FormatFloat(quotePrice, 'f', -1, 64)

	// 地址到符号映射
	if symbol, exists := config.SOLANA_DEX_ADDRESS_TO_NAME[quoteAddress]; exists {
		quoteSymbol = symbol
	} else {
		log.Printf("quoteSymbol not support %s", quoteAddress)
		return nil, nil
	}

	// 获取 USD 价格（这里需要实现价格服务）
	quoteTokenUSDPrice, err := handler.getTokenPrice(quoteSymbol, "USDT")
	if err != nil {
		log.Printf("Failed to get USD price for %s: %v", quoteSymbol, err)
		quoteTokenUSDPrice = 0
	}

	// 计算 USD 价格和金额
	usdPrice := quotePrice * quoteTokenUSDPrice
	usdAmount := quoteTokenUSDPrice * quoteAmountFloat

	usdPriceStr := strconv.FormatFloat(usdPrice, 'f', -1, 64)
	usdAmountStr := strconv.FormatFloat(usdAmount, 'f', -1, 64)

	data := &model.SwapTransaction{
		TxHash:          txHash,
		TransactionTime: transactionTime,
		WalletAddress:   walletAddress,
		TokenAmount:     tokenAmountFloat,
		TokenSymbol:     tokenSymbol,
		TokenAddress:    tokenAddress,
		QuoteSymbol:     quoteSymbol,
		QuoteAmount:     quoteAmountFloat,
		QuoteAddress:    quoteAddress,
		QuotePrice:      quotePriceStr,
		USDPrice:        usdPriceStr,
		USDAmount:       usdAmountStr,
		TradeType:       string(tradeType),
		PoolAddress:     poolAddress,
		BlockHeight:     blockNumber,
	}

	return data, nil
}

// FilterTokenData 过滤代币数据 - 对应 TS 版本的 filterTokenData 方法
func (handler *SolanaBlockDataHandler) FilterTokenData(data []model.SwapTransactionToken) []model.TokenSwapFilterData {
	var result []model.TokenSwapFilterData

	for _, transaction := range data {
		// 检查黑名单代币
		if config.IsBlacklistedToken(transaction.TokenAddress) ||
			config.IsBlacklistedToken(transaction.QuoteAddress) {
			continue
		}

		// 检查黑名单钱包
		if config.IsBlacklistedWallet(transaction.WalletAddress) {
			continue
		}

		// 检查 MEV 机器人
		if config.IsMEVBot(transaction.WalletAddress) {
			continue
		}

		// 检查基础代币
		tokenIsBase := config.IsBaseToken(strings.ToLower(transaction.TokenAddress))
		quoteIsBase := config.IsBaseToken(strings.ToLower(transaction.QuoteAddress))

		if !tokenIsBase && !quoteIsBase {
			continue
		}

		if tokenIsBase && quoteIsBase {
			continue
		}

		// 检查最小交易金额
		minAmount, _ := model.SNAP_SHOT_CONFIG.MinTransactionAmount.Float64()
		if transaction.USDAmount < minAmount {
			continue
		}

		filteredData := model.TokenSwapFilterData{
			UserAddress:     transaction.WalletAddress,
			PoolAddress:     transaction.PoolAddress,
			TxHash:          transaction.TxHash,
			IsBuy:           transaction.TradeType == string(model.ESwapTradeTypeBUY),
			BlockHeight:     transaction.BlockHeight,
			TokenSymbol:     transaction.TokenSymbol,
			TokenAddress:    transaction.TokenAddress,
			QuoteSymbol:     transaction.QuoteSymbol,
			QuoteAddress:    transaction.QuoteAddress,
			QuotePrice:      transaction.QuotePrice,
			USDPrice:        transaction.USDPrice,
			USDAmount:       transaction.USDAmount,
			TransactionTime: transaction.TransactionTime,
			TokenAmount:     transaction.TokenAmount,
			QuoteAmount:     transaction.QuoteAmount,
		}

		result = append(result, filteredData)
	}

	return result
}

// getTokenPrice 获取代币价格（占位符实现）
func (handler *SolanaBlockDataHandler) getTokenPrice(tokenSymbol, baseCurrency string) (float64, error) {
	// TODO: 实现真实的价格服务集成
	// 这里需要调用实际的价格 API

	// 基础代币的价格映射
	priceMap := map[string]float64{
		"SOL":  100.0, // 示例价格
		"USDC": 1.0,
		"USDT": 1.0,
		"WSOL": 100.0,
	}

	if price, exists := priceMap[tokenSymbol]; exists {
		return price, nil
	}

	return 0, fmt.Errorf("price not found for token %s", tokenSymbol)
}

// formatFloat 格式化浮点数
func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// 全局处理器实例
var ExportSolanaBlockDataHandler = NewSolanaBlockDataHandler()
