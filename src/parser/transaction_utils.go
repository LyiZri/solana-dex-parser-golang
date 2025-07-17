package parser

import (
	"github.com/go-solana-parse/src/config"
	"github.com/go-solana-parse/src/model"
)

// TransactionUtils 交易工具类
type TransactionUtils struct {
	adapter *TransactionAdapter
}

// NewTransactionUtils 创建新的交易工具实例
func NewTransactionUtils(adapter *TransactionAdapter) *TransactionUtils {
	return &TransactionUtils{
		adapter: adapter,
	}
}

// GetDexInfo 获取 DEX 信息
func (tu *TransactionUtils) GetDexInfo(classifier *InstructionClassifier) model.DexInfo {
	allProgramIDs := classifier.GetAllProgramIDs()

	for _, programID := range allProgramIDs {
		if program, exists := config.GetDexProgramByID(programID); exists {
			return model.DexInfo{
				ProgramID: programID,
				AMM:       program.Name,
			}
		}
	}

	return model.DexInfo{}
}

// GetTransferActions 获取转账操作
func (tu *TransactionUtils) GetTransferActions(actionTypes []string) map[string][]model.TransferData {
	// TODO: 实现转账操作提取逻辑
	// 这里需要从交易指令中提取 transfer、mintTo、burn 等操作
	return make(map[string][]model.TransferData)
}

// ProcessSwapData 处理交换数据
func (tu *TransactionUtils) ProcessSwapData(transfers []model.TransferData, dexInfo model.DexInfo) *model.TradeInfo {
	// TODO: 实现交换数据处理逻辑
	// 根据转账数据构建交易信息
	return nil
}

// AttachTokenTransferInfo 附加代币转账信息
func (tu *TransactionUtils) AttachTokenTransferInfo(trade model.TradeInfo, transferActions map[string][]model.TransferData) model.TradeInfo {
	// TODO: 实现代币转账信息附加逻辑
	return trade
}
