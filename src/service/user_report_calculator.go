package service

import (
	"fmt"
	"math"
	"sort"

	"github.com/go-solana-parse/src/config"
	"github.com/go-solana-parse/src/db"
	"github.com/go-solana-parse/src/db/clickhouse"
	"github.com/go-solana-parse/src/db/mysql"
	"github.com/go-solana-parse/src/model"
	"github.com/shopspring/decimal"
)

const (
	TRADE_TYPE_BUY  = "BUY"
	TRADE_TYPE_SELL = "SELL"
)

// UserReportCalculator 用户报告计算器
type UserReportCalculator struct {
	priceService *PriceService
}

// NewUserReportCalculator 创建新的用户报告计算器
func NewUserReportCalculator() *UserReportCalculator {
	if db.ClickHouseClient == nil {
		panic("ClickHouse client is not initialized. Please call db.InitClickHouseV2() first.")
	}
	return &UserReportCalculator{
		priceService: NewPriceService(db.ClickHouseClient),
	}
}

// CalculateUserReport 计算单个用户的报告
func (calc *UserReportCalculator) CalculateUserReport(address string) (*mysql.UserReport, error) {
	// 获取用户所有交易记录
	transactions, err := calc.getUserTransactions(address)
	if err != nil {
		return nil, fmt.Errorf("获取用户交易记录失败: %v", err)
	}

	if len(transactions) == 0 {
		return nil, fmt.Errorf("用户 %s 没有交易记录", address)
	}

	// 按时间排序交易记录
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].TransactionTime < transactions[j].TransactionTime
	})

	// 初始化用户报告
	userReport := &mysql.UserReport{
		UserAddr: address,
	}

	// 计算基础指标
	calc.calculateBasicMetrics(userReport, transactions)

	// 计算 PnL 相关指标
	tokenPnLMap, err := calc.calculateTokenPnL(transactions)
	if err != nil {
		return nil, fmt.Errorf("计算代币PnL失败: %v", err)
	}

	calc.calculatePnLMetrics(userReport, tokenPnLMap)

	// 计算投资组合指标
	calc.calculatePortfolioMetrics(userReport, tokenPnLMap)

	return userReport, nil
}

// getAllUniqueAddresses 获取所有唯一地址
func (calc *UserReportCalculator) GetAllUniqueAddresses() ([]string, error) {
	solanaData := &clickhouse.SolanaHistoryData{}
	return solanaData.GetTotalUniqueAddress(db.ClickHouseClient)
}

// GetAllUniqueAddressesOrderByTradeCount 获取所有唯一地址，按交易量升序排序
func (calc *UserReportCalculator) GetAllUniqueAddressesOrderByTradeCount() ([]string, error) {
	viewData := &clickhouse.ViewSolanaWalletTradeCount{}
	return viewData.GetWalletAddressesByTradeCountASC(db.ClickHouseClient)
}

// getUserTransactions 获取用户所有交易记录
func (calc *UserReportCalculator) getUserTransactions(address string) ([]*clickhouse.SolanaHistoryData, error) {
	solanaData := &clickhouse.SolanaHistoryData{}
	return solanaData.GetUserTransactionsByAddress(db.ClickHouseClient, address)
}

// calculateBasicMetrics 计算基础指标
func (calc *UserReportCalculator) calculateBasicMetrics(userReport *mysql.UserReport, transactions []*clickhouse.SolanaHistoryData) {
	if len(transactions) == 0 {
		return
	}

	// 第一笔交易信息
	firstTx := transactions[0]
	userReport.FirstTx = int64(firstTx.TransactionTime)
	userReport.FirstTokenAddr = firstTx.TokenAddress
	userReport.FirstTokenAmount = decimal.NewFromFloat(firstTx.TokenAmount).Round(6)
	userReport.FirstSolAmount = decimal.NewFromFloat(firstTx.QuoteAmount).Round(6)

	var totalBuyVolume, totalSellVolume float64
	var buyCount, sellCount int64
	uniqueTokens := make(map[string]bool)

	for _, tx := range transactions {
		// 统计买卖交易量（U本位）

		quotePrice := 1.0

		if tx.QuoteAddress == config.SOL_ADDRESS || tx.QuoteAddress == config.WSOL_ADDRESS {
			fmt.Printf("获取SOL价格: %d\n", tx.BlockHeight)
			solPrice, err := calc.priceService.GetSOLPriceAtBlock(tx.BlockHeight)
			if err != nil {
				fmt.Printf("获取SOL价格失败: %v\n", err)
			}
			if err == nil && solPrice > 0 {
				quotePrice = solPrice
			}
		}

		if tx.TradeType == TRADE_TYPE_BUY {
			totalBuyVolume += tx.QuoteAmount * quotePrice // 买入时花费的SOL
			buyCount++
		} else if tx.TradeType == TRADE_TYPE_SELL {
			totalSellVolume += tx.QuoteAmount * quotePrice // 卖出时获得的SOL
			sellCount++
		}

		// 统计唯一代币
		uniqueTokens[tx.TokenAddress] = true
	}

	userReport.TxAmountUsd = decimal.NewFromFloat(totalBuyVolume + totalSellVolume).Round(6)
	userReport.TxBuyAmountUsd = decimal.NewFromFloat(totalBuyVolume).Round(6)
	userReport.TxSellAmountUsd = decimal.NewFromFloat(totalSellVolume).Round(6)
	userReport.TxCount = buyCount + sellCount
	userReport.TxBuyCount = buyCount
	userReport.TxSellCount = sellCount
	userReport.TokenCount = int64(len(uniqueTokens))
}

// calculateTokenPnL 计算每个代币的盈亏数据
func (calc *UserReportCalculator) calculateTokenPnL(transactions []*clickhouse.SolanaHistoryData) (map[string]*model.TokenPnLData, error) {
	tokenMap := make(map[string]*model.TokenPnLData)

	for _, tx := range transactions {
		tokenAddr := tx.TokenAddress

		if tokenAddr == "3LNJzpzLzobb9kghSwH1Ro1W3NPB4G1Vc9vTdj2P3Jij" {
			fmt.Printf("代币地址: %s 交易类型: %s 交易数量: %f 交易价格: %f\n", tokenAddr, tx.TradeType, tx.TokenAmount, tx.QuoteAmount)
		}

		if tokenMap[tokenAddr] == nil {
			tokenMap[tokenAddr] = &model.TokenPnLData{
				TokenAddress: tokenAddr,
			}
		}

		token := tokenMap[tokenAddr]

		if tx.TradeType == TRADE_TYPE_BUY {
			// 更新买入数据
			token.TotalBuyAmount += tx.TokenAmount
			if tx.QuoteAddress == config.SOL_ADDRESS || tx.QuoteAddress == config.WSOL_ADDRESS {
				solPrice, err := calc.priceService.GetSOLPriceAtBlock(tx.BlockHeight)
				if err == nil && solPrice > 0 {
					token.TotalBuyValue += tx.QuoteAmount * solPrice
				}
			} else {
				token.TotalBuyValue += tx.QuoteAmount
			}

			// 计算加权平均买入价格
			if token.TotalBuyAmount > 0 {
				token.AvgBuyPrice = token.TotalBuyValue / token.TotalBuyAmount
				if token.TokenAddress == "3LNJzpzLzobb9kghSwH1Ro1W3NPB4G1Vc9vTdj2P3Jij" {
					fmt.Printf("代币地址: %s 总买入价值: %f 总买入数量: %f 平均买入价格: %f\n tx.QuoteAmount: %f\n", token.TokenAddress, token.TotalBuyValue, token.TotalBuyAmount, token.AvgBuyPrice, tx.QuoteAmount)
				}
			}

			// 更新当前持仓
			token.CurrentHolding += tx.TokenAmount

		} else if tx.TradeType == TRADE_TYPE_SELL {
			// 更新卖出数据
			if token.TotalBuyAmount == 0 {
				continue
			}

			currentSellUsdValue := tx.QuoteAmount

			token.TotalSellAmount += tx.TokenAmount
			if tx.QuoteAddress == config.SOL_ADDRESS || tx.QuoteAddress == config.WSOL_ADDRESS {
				solPrice, err := calc.priceService.GetSOLPriceAtBlock(tx.BlockHeight)
				if err == nil && solPrice > 0 {
					currentSellUsdValue = tx.QuoteAmount * solPrice
				}
			}

			token.TotalSellValue += currentSellUsdValue

			// 计算加权平均卖出价格
			if token.TotalSellAmount > 0 {
				token.AvgSellPrice = token.TotalSellValue / token.TotalSellAmount
			}

			// 计算已实现盈亏
			if token.AvgBuyPrice > 0 {
				sellPrice := currentSellUsdValue / tx.TokenAmount

				realizedSellTokenAmount := math.Min(tx.TokenAmount, token.CurrentHolding)

				realizedPnL := (sellPrice - token.AvgBuyPrice) * realizedSellTokenAmount
				if token.TokenAddress == "3LNJzpzLzobb9kghSwH1Ro1W3NPB4G1Vc9vTdj2P3Jij" {
					fmt.Printf("代币地址: %s 卖出价格: %f 卖出数量: %f 真实卖出数量: %f 平均买入价格: %f 当前持仓: %f 盈亏: %f\n", token.TokenAddress, sellPrice, tx.TokenAmount, realizedSellTokenAmount, token.AvgBuyPrice, token.CurrentHolding, realizedPnL)
				}
				token.RealizedPnL += realizedPnL
			}

			// 更新当前持仓
			token.CurrentHolding -= tx.TokenAmount
			if token.CurrentHolding < 0 {
				token.CurrentHolding = 0 // 避免负数持仓
			}
		}

		// 计算历史最高持仓价值
		currentPrice, err := calc.priceService.GetTokenPriceAtBlock(tokenAddr, tx.BlockHeight)
		if err == nil && currentPrice > 0 {
			currentValue := token.CurrentHolding * currentPrice
			if currentValue > token.MaxHoldingValue {
				token.MaxHoldingValue = currentValue
				token.MaxHoldingAmount = token.CurrentHolding
			}
		}
	}

	return tokenMap, nil
}

// calculatePnLMetrics 计算盈亏相关指标
func (calc *UserReportCalculator) calculatePnLMetrics(userReport *mysql.UserReport, tokenPnLMap map[string]*model.TokenPnLData) {
	var pnlWinCount, pnlLossCount int64
	var topProfitPnL, topLossPnL float64
	var topProfitToken, topLossToken string
	var topProfitTokenData *model.TokenPnLData // 保存最盈利代币的完整数据

	// 盈利分布计数器
	var winLevel1, winLevel2, winLevel3, winLevel4 int64 // 0-200%, 200-500%, 500-1000%, >1000%
	var lossLevel1, lossLevel2 int64                     // 0-50%, >50%

	for tokenAddr, tokenData := range tokenPnLMap {
		fmt.Printf("代币地址: %s 总盈亏: %f 当前价格: %f 平均买入价格: %f 平均卖出价格: %f 当前持仓: %f 总买入: %f 总卖出: %f\n", tokenAddr, tokenData.RealizedPnL, tokenData.UnrealizedPnL, tokenData.AvgBuyPrice, tokenData.AvgSellPrice, tokenData.CurrentHolding, tokenData.TotalBuyValue, tokenData.TotalSellValue)
		if tokenData.TotalBuyValue == 0 {
			continue // 跳过没有买入记录的代币
		}

		// 计算总盈亏（已实现 + 未实现）
		// 计算未实现盈亏（使用最新可用价格）
		// 注意：这里使用一个很大的区块高度来获取最新价格
		currentPrice, err := calc.priceService.GetTokenPriceAtBlock(tokenAddr, 999999999)
		if err == nil && currentPrice > 0 && tokenData.AvgBuyPrice > 0 {
			unrealizedPnL := (currentPrice - tokenData.AvgBuyPrice) * tokenData.CurrentHolding
			tokenData.UnrealizedPnL = unrealizedPnL
		}

		totalPnL := tokenData.RealizedPnL + tokenData.UnrealizedPnL

		fmt.Printf("代币地址: %s 总盈亏: %f 当前价格: %f 平均买入价格: %f 当前持仓: %f 总买入: %f 总卖出: %f\n",
			tokenAddr, totalPnL, currentPrice, tokenData.AvgBuyPrice, tokenData.CurrentHolding, tokenData.TotalBuyValue, tokenData.TotalSellValue)

		// 统计盈亏代币数量
		if totalPnL > 0 {
			pnlWinCount++

			// 计算盈利率
			profitRate := totalPnL / tokenData.TotalBuyValue

			// 盈利分布统计
			if profitRate <= 2.0 { // 0-200%
				winLevel1++
			} else if profitRate <= 5.0 { // 200-500%
				winLevel2++
			} else if profitRate <= 10.0 { // 500-1000%
				winLevel3++
			} else { // >1000%
				winLevel4++
			}

			// 记录最高盈利代币
			if totalPnL > topProfitPnL {
				topProfitPnL = totalPnL
				topProfitToken = tokenAddr
				topProfitTokenData = tokenData // 保存完整的代币数据用于后续计算
			}
		} else if totalPnL < 0 {
			pnlLossCount++

			// 计算亏损率
			lossRate := math.Abs(totalPnL) / tokenData.TotalBuyValue

			// 亏损分布统计
			if lossRate <= 0.5 { // 0-50%
				lossLevel1++
			} else { // >50%
				lossLevel2++
			}

			// 记录最大亏损代币
			if totalPnL < topLossPnL {
				topLossPnL = totalPnL
				topLossToken = tokenAddr
			}
		}
	}

	// 设置用户报告字段
	userReport.TokenWinCount = pnlWinCount
	userReport.TokenLossCount = pnlLossCount

	// 计算胜率
	if userReport.TokenCount > 0 {
		winRate := float64(pnlWinCount) / float64(userReport.TokenCount)
		userReport.WinRate = decimal.NewFromFloat(winRate).Round(6)
	}

	userReport.MostEarnTokenAddr = topProfitToken
	userReport.MostLossTokenAddr = topLossToken

	// 计算新增的三个字段
	// 1. TopProfitUsdAmount - 最多盈利的代币盈利金额（USD）
	if topProfitPnL > 0 {
		// 将SOL本位的盈利转换为USD金额
		// 这里假设盈利已经是USD金额，如果是SOL本位需要转换
		userReport.MostEarnTokenAmountUsd = decimal.NewFromFloat(topProfitPnL).Round(6)
	} else {
		userReport.MostEarnTokenAmountUsd = decimal.NewFromFloat(0)
	}

	// 2. TopLossUsdAmount - 最多损失的代币损失金额（USD）
	if topLossPnL < 0 {
		// 取绝对值，转换为正数表示损失金额
		userReport.MostLossTokenAmountUsd = decimal.NewFromFloat(math.Abs(topLossPnL)).Round(4)
	} else {
		userReport.MostLossTokenAmountUsd = decimal.NewFromFloat(0)
	}

	// 3. TopProfitWinRate - 最多盈利的代币的胜率/翻了多少倍
	if topProfitTokenData != nil && topProfitTokenData.TotalBuyValue > 0 {
		// 计算盈利倍数：盈利金额 / 投入成本
		profitMultiplier := topProfitPnL / topProfitTokenData.TotalBuyValue
		// 例如：投入100，盈利200，倍数为2（翻了2倍）

		// 临时解决方案：限制最大值以适应当前MySQL字段 DECIMAL(10,2) 的限制
		// TODO: 将MySQL字段改为 DECIMAL(10,4) 后可以移除此限制
		const MAX_WIN_RATE = 99999999.99 // MySQL DECIMAL(10,2) 的最大值
		if profitMultiplier > MAX_WIN_RATE {
			profitMultiplier = MAX_WIN_RATE
		}

		// 限制精度为4位小数，避免MySQL DECIMAL字段溢出
		userReport.MostEarnTokenWinRate = decimal.NewFromFloat(profitMultiplier).Round(2)
	} else {
		userReport.MostEarnTokenWinRate = decimal.NewFromFloat(0)
	}

	// 盈利分布
	userReport.MetricE0E200 = winLevel1
	userReport.MetricE200E500 = winLevel2
	userReport.MetricE500E1000 = winLevel3
	userReport.MetricE1000 = winLevel4

	// 亏损分布
	userReport.MetricL50L0 = lossLevel1
	userReport.MetricL50 = lossLevel2
}

// calculatePortfolioMetrics 计算投资组合指标
func (calc *UserReportCalculator) calculatePortfolioMetrics(userReport *mysql.UserReport, tokenPnLMap map[string]*model.TokenPnLData) {
	var maxTotalHoldValue float64
	var mostHoldValueToken string
	var mostHoldValueUSD float64

	for tokenAddr, tokenData := range tokenPnLMap {
		// 寻找历史最高持仓价值的代币
		if tokenData.MaxHoldingValue > mostHoldValueUSD {
			mostHoldValueUSD = tokenData.MaxHoldingValue
			mostHoldValueToken = tokenAddr
		}

		// 累计历史最高总持仓价值
		maxTotalHoldValue += tokenData.MaxHoldingValue
	}

	userReport.MostHoldTokenAddr = mostHoldValueToken
	userReport.MostHoldTokenAmountUsd = decimal.NewFromFloat(mostHoldValueUSD).Round(6)
	userReport.MostWalletHoldUsd = decimal.NewFromFloat(maxTotalHoldValue).Round(6)

	// 计算用户整体盈亏率
	var totalInvestment, totalCurrentValue, totalProfit float64
	for _, tokenData := range tokenPnLMap {
		totalInvestment += tokenData.TotalBuyValue

		// 计算当前价值
		if tokenData.CurrentHolding > 0 {
			currentPrice, err := calc.priceService.GetTokenPriceAtBlock(tokenData.TokenAddress, 999999999)
			if err == nil && currentPrice > 0 {
				totalCurrentValue += tokenData.CurrentHolding * currentPrice
			}
		}
		totalProfit += tokenData.TotalSellValue
	}

	if totalInvestment > 0 {
		profitRate := (totalCurrentValue - totalInvestment + totalProfit) / totalInvestment
		userReport.TotalPnlUsd = decimal.NewFromFloat(profitRate).Round(6)
	}

}
