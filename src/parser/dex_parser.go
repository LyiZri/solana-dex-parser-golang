package parser

import (
	"fmt"
	"log"
	"time"

	configsvc "github.com/go-solana-parse/src/config"
	"github.com/go-solana-parse/src/model"
)

// DexParser DEX 解析器主结构
type DexParser struct {
	// Trade parser mapping - 将来扩展用于不同协议的解析器
	parserMap map[string]interface{}
}

// NewDexParser 创建新的 DEX 解析器实例
func NewDexParser() *DexParser {
	return &DexParser{
		parserMap: make(map[string]interface{}),
	}
}

// ParseBlockData 解析整个区块数据 - 对应 TS 版本的 parseBlockData 方法
func (dp *DexParser) ParseBlockData(blockData model.VersionedBlockResponse, blockNumber uint64) ([]model.ParseResult, error) {
	start := time.Now()

	// 过滤有效交易（排除失败交易）
	var validTransactions []model.TransactionInfo
	for _, tx := range blockData.Transactions {
		if tx.Meta == nil || tx.Meta.Err == nil {
			validTransactions = append(validTransactions, tx)
		}
	}

	// 解析每个有效交易
	var parseResults []model.ParseResult
	for _, transaction := range validTransactions {
		solanaTransaction := model.SolanaTransaction{
			Transaction: transaction,
			Meta:        transaction.Meta,
			BlockTime:   blockData.BlockTime,
			Slot:        blockNumber,
		}

		result := dp.ParseAllComplete(solanaTransaction, nil)
		parseResults = append(parseResults, result)
	}

	log.Printf("parse block %d, cost: %v ms", blockNumber, time.Since(start).Milliseconds())
	return parseResults, nil
}

// ParseAllComplete 完整解析单个交易 - 对应 TS 版本的 parseAllComplete 方法
func (dp *DexParser) ParseAllComplete(tx model.SolanaTransaction, config *model.ParseConfig) model.ParseResult {
	if config == nil {
		config = &model.ParseConfig{
			TryUnknownDEX: true,
		}
	}

	parseResult := dp.parseWithClassifier(tx, *config, "all")
	return dp.enhanceParseResultComplete(parseResult)
}

// parseWithClassifier 使用分类器解析交易 - 对应 TS 版本的 parseWithClassifier 方法
func (dp *DexParser) parseWithClassifier(
	tx model.SolanaTransaction,
	config model.ParseConfig,
	parseType string,
) model.ParseResult {
	result := model.ParseResult{
		State:              true,
		Fee:                model.TokenAmount{Amount: "0", UIAmount: 0, Decimals: 9},
		Trades:             []model.TradeInfo{},
		Liquidities:        []model.PoolEvent{},
		Transfers:          []model.TransferData{},
		TokenBalanceChange: make(map[string]*model.BalanceChange),
		MoreEvents:         make(map[string]interface{}),
		Result: model.EnhancedResult{
			Trades:             []model.ResSwapStruct{},
			Liquidities:        []model.ResLpInfoStruct{},
			Tokens:             []model.ResTokenMetadataStruct{},
			TokenPrices:        []model.ResTokenPriceStruct{},
			UserTradingSummary: []model.ResUserTradingSummaryStruct{},
		},
	}

	defer func() {
		if r := recover(); r != nil {
			if config.ThrowError {
				panic(r)
			}
			result.State = false
			result.Msg = fmt.Sprintf("Parse error: %v %v", tx.Transaction.Transaction.Signatures[0], r)
			log.Printf("Parse error: %v", r)
		}
	}()

	// 创建适配器和工具
	adapter := NewTransactionAdapter(tx, config)
	utils := NewTransactionUtils(adapter)
	classifier := NewInstructionClassifier(adapter)

	// 获取 DEX 信息和验证
	dexInfo := utils.GetDexInfo(classifier)
	allProgramIDs := classifier.GetAllProgramIDs()
	transferActions := utils.GetTransferActions([]string{"mintTo", "burn", "mintToChecked", "burnChecked"})

	// 处理手续费
	result.Fee = adapter.GetFee()

	// 处理用户余额变化
	result.SOLBalanceChange = adapter.GetAccountSOLBalanceChanges(false, adapter.GetSigner())
	result.TokenBalanceChange = adapter.GetAccountTokenBalanceChanges(true, adapter.GetSigner())

	// 优先处理 Jupiter 系列协议
	if dexInfo.ProgramID != "" && configsvc.IsJupiterFamily(dexInfo.ProgramID) {
		if parseType == "trades" || parseType == "all" {
			jupiterInstructions := classifier.GetInstructions(dexInfo.ProgramID)
			trades := dp.parseJupiterTransaction(adapter, dexInfo, transferActions, jupiterInstructions)
			result.Trades = append(result.Trades, trades...)
		}
		if len(result.Trades) > 0 {
			return result
		}
	}

	// 处理每个程序的指令
	for _, programID := range allProgramIDs {
		classifiedInstructions := classifier.GetInstructions(programID)

		// 处理交易（如果需要）
		if parseType == "trades" || parseType == "all" {
			if len(config.ProgramIDs) > 0 && !contains(config.ProgramIDs, programID) {
				continue
			}
			if len(config.IgnoreProgramIDs) > 0 && contains(config.IgnoreProgramIDs, programID) {
				continue
			}

			// 尝试使用专用解析器
			if dp.hasParser(programID) {
				trades := dp.parseWithSpecificParser(programID, adapter, dexInfo, transferActions, classifiedInstructions)
				result.Trades = append(result.Trades, trades...)
			} else if config.TryUnknownDEX {
				// 处理未知 DEX 程序
				transfers := dp.getTransfersForProgram(transferActions, programID)
				if len(transfers) >= 2 && dp.hasSupportedToken(transfers, adapter) {
					trade := utils.ProcessSwapData(transfers, model.DexInfo{
						ProgramID: programID,
						AMM:       configsvc.GetProgramName(programID),
					})
					if trade != nil {
						enhancedTrade := utils.AttachTokenTransferInfo(*trade, transferActions)
						result.Trades = append(result.Trades, enhancedTrade)
					}
				}
			}
		}

		// 处理流动性（如果需要）
		if parseType == "liquidity" || parseType == "all" {
			if len(config.ProgramIDs) > 0 && !contains(config.ProgramIDs, programID) {
				continue
			}
			if len(config.IgnoreProgramIDs) > 0 && contains(config.IgnoreProgramIDs, programID) {
				continue
			}

			liquidities := dp.parseLiquidityEvents(programID, adapter, transferActions, classifiedInstructions)
			result.Liquidities = append(result.Liquidities, liquidities...)
		}
	}

	// 去重交易
	if len(result.Trades) > 0 {
		result.Trades = dp.deduplicateTrades(result.Trades)
	}

	// 处理转账（如果没有交易和流动性）
	if len(result.Trades) == 0 && len(result.Liquidities) == 0 {
		if parseType == "transfer" || parseType == "all" {
			if dexInfo.ProgramID != "" {
				classifiedInstructions := classifier.GetInstructions(dexInfo.ProgramID)
				transfers := dp.parseTransfers(dexInfo.ProgramID, adapter, dexInfo, transferActions, classifiedInstructions)
				result.Transfers = append(result.Transfers, transfers...)
			}
			if len(result.Transfers) == 0 {
				// 添加所有转账作为备用
				for _, transferList := range transferActions {
					result.Transfers = append(result.Transfers, transferList...)
				}
			}
		}
	}

	// 处理更多事件
	dp.processMoreEvents(parseType, &result, allProgramIDs, adapter, transferActions, classifier)

	return result
}

// enhanceParseResultComplete 增强解析结果 - 对应 TS 版本的 enhanceParseResultComplete 方法
func (dp *DexParser) enhanceParseResultComplete(parseResult model.ParseResult) model.ParseResult {
	// 处理 token 信息
	parseResult.Result.Tokens = dp.extractTokenMetadata(parseResult)

	// 处理 token 价格
	parseResult.Result.TokenPrices = dp.calculateTokenPrices(parseResult)

	// 处理用户交易汇总
	parseResult.Result.UserTradingSummary = dp.generateUserTradingSummary(parseResult)

	// 处理流动性数据
	dp.enhanceLiquidityData(&parseResult)

	// 格式化交易数据
	parseResult.Result.Trades = dp.formatAllTrades(parseResult)

	return parseResult
}

// 辅助方法实现

func (dp *DexParser) parseJupiterTransaction(adapter *TransactionAdapter, dexInfo model.DexInfo, transferActions map[string][]model.TransferData, instructions []model.ClassifiedInstruction) []model.TradeInfo {
	// Jupiter 特殊解析逻辑
	// TODO: 实现 Jupiter 聚合器的特殊处理
	return []model.TradeInfo{}
}

func (dp *DexParser) hasParser(programID string) bool {
	_, exists := dp.parserMap[programID]
	return exists
}

func (dp *DexParser) parseWithSpecificParser(programID string, adapter *TransactionAdapter, dexInfo model.DexInfo, transferActions map[string][]model.TransferData, instructions []model.ClassifiedInstruction) []model.TradeInfo {
	// TODO: 实现特定协议的解析器
	return []model.TradeInfo{}
}

func (dp *DexParser) getTransfersForProgram(transferActions map[string][]model.TransferData, programID string) []model.TransferData {
	for key, transfers := range transferActions {
		if len(key) >= len(programID) && key[:len(programID)] == programID {
			return transfers
		}
	}
	return []model.TransferData{}
}

func (dp *DexParser) hasSupportedToken(transfers []model.TransferData, adapter *TransactionAdapter) bool {
	for _, transfer := range transfers {
		if adapter.IsSupportedToken(transfer.Info.Mint) {
			return true
		}
	}
	return false
}

func (dp *DexParser) parseLiquidityEvents(programID string, adapter *TransactionAdapter, transferActions map[string][]model.TransferData, instructions []model.ClassifiedInstruction) []model.PoolEvent {
	// TODO: 实现流动性事件解析
	return []model.PoolEvent{}
}

func (dp *DexParser) parseTransfers(programID string, adapter *TransactionAdapter, dexInfo model.DexInfo, transferActions map[string][]model.TransferData, instructions []model.ClassifiedInstruction) []model.TransferData {
	// TODO: 实现转账解析
	return []model.TransferData{}
}

func (dp *DexParser) processMoreEvents(parseType string, result *model.ParseResult, allProgramIDs []string, adapter *TransactionAdapter, transferActions map[string][]model.TransferData, classifier *InstructionClassifier) {
	if parseType == "all" {
		// 处理 Pump.fun 事件
		if contains(allProgramIDs, configsvc.DEX_PROGRAMS["PUMP_FUN"].ID) {
			// TODO: 实现 Pump.fun 事件解析
		}

		// 处理 Raydium Launchpad 事件
		if contains(allProgramIDs, configsvc.DEX_PROGRAMS["RAYDIUM_LCP"].ID) {
			// TODO: 实现 Raydium Launchpad 事件解析
		}
	}
}

func (dp *DexParser) deduplicateTrades(trades []model.TradeInfo) []model.TradeInfo {
	seen := make(map[string]bool)
	var result []model.TradeInfo

	for _, trade := range trades {
		key := fmt.Sprintf("%s-%s", trade.IDX, trade.Signature)
		if !seen[key] {
			seen[key] = true
			result = append(result, trade)
		}
	}

	return result
}

func (dp *DexParser) extractTokenMetadata(parseResult model.ParseResult) []model.ResTokenMetadataStruct {
	// TODO: 实现代币元数据提取
	return []model.ResTokenMetadataStruct{}
}

func (dp *DexParser) calculateTokenPrices(parseResult model.ParseResult) []model.ResTokenPriceStruct {
	// TODO: 实现代币价格计算
	return []model.ResTokenPriceStruct{}
}

func (dp *DexParser) generateUserTradingSummary(parseResult model.ParseResult) []model.ResUserTradingSummaryStruct {
	// TODO: 实现用户交易汇总
	return []model.ResUserTradingSummaryStruct{}
}

func (dp *DexParser) enhanceLiquidityData(parseResult *model.ParseResult) {
	// TODO: 实现流动性数据增强
}

func (dp *DexParser) formatAllTrades(parseResult model.ParseResult) []model.ResSwapStruct {
	var result []model.ResSwapStruct

	for _, trade := range parseResult.Trades {
		result = append(result, model.ResSwapStruct{
			TransactionSignature: trade.Signature,
			BlockTime:            trade.BlockTime,
			UserAddress:          trade.Signer,
			TokenInMint:          trade.TokenInMint,
			TokenInSymbol:        trade.TokenInSymbol,
			TokenInAmount:        parseFloat64(trade.TokenInAmount),
			TokenOutMint:         trade.TokenOutMint,
			TokenOutSymbol:       trade.TokenOutSymbol,
			TokenOutAmount:       parseFloat64(trade.TokenOutAmount),
			PoolAddress:          "", // TODO: 从交易中提取池地址
			AMM:                  trade.AMM,
			Route:                trade.Route,
			SlotNumber:           trade.SlotNumber,
		})
	}

	return result
}

// 工具函数

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func parseFloat64(s string) float64 {
	// TODO: 实现字符串到 float64 的安全转换
	return 0.0
}

// 全局解析器实例
var ExportDexParserInstance = NewDexParser()
