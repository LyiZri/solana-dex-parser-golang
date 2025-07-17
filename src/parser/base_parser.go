package parser

import (
	"github.com/go-solana-parse/src/model"
)

// BaseParser 基础解析器 - 对应 TS 版本的 BaseParser
type BaseParser struct {
	adapter                *TransactionAdapter
	dexInfo                model.DexInfo
	transferActions        map[string][]model.TransferData
	classifiedInstructions []model.ClassifiedInstruction
	utils                  *TransactionUtils
}

// NewBaseParser 创建基础解析器实例
func NewBaseParser(
	adapter *TransactionAdapter,
	dexInfo model.DexInfo,
	transferActions map[string][]model.TransferData,
	classifiedInstructions []model.ClassifiedInstruction,
) *BaseParser {
	return &BaseParser{
		adapter:                adapter,
		dexInfo:                dexInfo,
		transferActions:        transferActions,
		classifiedInstructions: classifiedInstructions,
		utils:                  NewTransactionUtils(adapter),
	}
}

// ProcessTrades 处理交易（需要具体解析器实现）
type TradeProcessor interface {
	ProcessTrades() []model.TradeInfo
}

// GetTransfersForInstruction 获取特定指令的转账数据 - 对应 TS 版本的 getTransfersForInstruction
func (bp *BaseParser) GetTransfersForInstruction(
	programID string,
	outerIndex int,
	innerIndex *int,
	extraTypes []string,
) []model.TransferData {
	var key string
	if innerIndex == nil {
		key = programID + ":" + string(rune(outerIndex))
	} else {
		key = programID + ":" + string(rune(outerIndex)) + "-" + string(rune(*innerIndex))
	}

	transfers, exists := bp.transferActions[key]
	if !exists {
		return []model.TransferData{}
	}

	// 过滤支持的转账类型
	supportedTypes := []string{"transfer", "transferChecked"}
	if extraTypes != nil {
		supportedTypes = append(supportedTypes, extraTypes...)
	}

	var filtered []model.TransferData
	for _, transfer := range transfers {
		for _, supportedType := range supportedTypes {
			if transfer.Info.Source == supportedType {
				filtered = append(filtered, transfer)
				break
			}
		}
	}

	return filtered
}
