package parser

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/go-solana-parse/src/config"
	"github.com/go-solana-parse/src/model"
)

// TransactionAdapter 交易适配器
type TransactionAdapter struct {
	transaction model.SolanaTransaction
	config      model.ParseConfig
}

// NewTransactionAdapter 创建新的交易适配器
func NewTransactionAdapter(tx model.SolanaTransaction, cfg model.ParseConfig) *TransactionAdapter {
	return &TransactionAdapter{
		transaction: tx,
		config:      cfg,
	}
}

// GetSigner 获取交易签名者
func (ta *TransactionAdapter) GetSigner() string {
	if len(ta.transaction.Transaction.Transaction.Signatures) > 0 {
		// 通常第一个签名者是交易发起者
		if len(ta.transaction.Transaction.Transaction.Message.AccountKeys) > 0 {
			return ta.transaction.Transaction.Transaction.Message.AccountKeys[0]
		}
	}
	return ""
}

// GetFee 获取交易手续费
func (ta *TransactionAdapter) GetFee() model.TokenAmount {
	if ta.transaction.Meta != nil {
		return model.TokenAmount{
			Amount:   strconv.FormatUint(ta.transaction.Meta.Fee, 10),
			UIAmount: float64(ta.transaction.Meta.Fee) / 1e9, // SOL 小数位数
			Decimals: 9,
		}
	}
	return model.TokenAmount{
		Amount:   "0",
		UIAmount: 0,
		Decimals: 9,
	}
}

// GetAccountSOLBalanceChanges 获取账户 SOL 余额变化
func (ta *TransactionAdapter) GetAccountSOLBalanceChanges(excludeWSOL bool, account string) *model.BalanceChange {
	if ta.transaction.Meta == nil {
		return nil
	}

	// 查找账户在 accountKeys 中的索引
	accountIndex := -1
	for i, key := range ta.transaction.Transaction.Transaction.Message.AccountKeys {
		if key == account {
			accountIndex = i
			break
		}
	}

	if accountIndex == -1 || accountIndex >= len(ta.transaction.Meta.PreBalances) || accountIndex >= len(ta.transaction.Meta.PostBalances) {
		return nil
	}

	preBalance := ta.transaction.Meta.PreBalances[accountIndex]
	postBalance := ta.transaction.Meta.PostBalances[accountIndex]

	change := int64(postBalance) - int64(preBalance)

	return &model.BalanceChange{
		Before: new(big.Int).SetUint64(preBalance),
		After:  new(big.Int).SetUint64(postBalance),
		Change: new(big.Int).SetInt64(change),
	}
}

// GetAccountTokenBalanceChanges 获取账户代币余额变化
func (ta *TransactionAdapter) GetAccountTokenBalanceChanges(includeAuthority bool, account string) map[string]*model.BalanceChange {
	changes := make(map[string]*model.BalanceChange)

	if ta.transaction.Meta == nil {
		return changes
	}

	// 处理前置代币余额
	preTokenBalances := make(map[string]model.TokenBalance)
	for _, balance := range ta.transaction.Meta.PreTokenBalances {
		key := fmt.Sprintf("%s-%d", balance.Mint, balance.AccountIndex)
		preTokenBalances[key] = balance
	}

	// 处理后置代币余额
	for _, postBalance := range ta.transaction.Meta.PostTokenBalances {
		key := fmt.Sprintf("%s-%d", postBalance.Mint, postBalance.AccountIndex)

		// 检查是否是目标账户
		if postBalance.AccountIndex < len(ta.transaction.Transaction.Transaction.Message.AccountKeys) {
			accountKey := ta.transaction.Transaction.Transaction.Message.AccountKeys[postBalance.AccountIndex]
			if accountKey != account && !includeAuthority {
				continue
			}
		}

		preBalance := uint64(0)
		if pre, exists := preTokenBalances[key]; exists {
			if amount, err := strconv.ParseUint(pre.UiTokenAmount.Amount, 10, 64); err == nil {
				preBalance = amount
			}
		}

		postAmount := uint64(0)
		if amount, err := strconv.ParseUint(postBalance.UiTokenAmount.Amount, 10, 64); err == nil {
			postAmount = amount
		}

		change := int64(postAmount) - int64(preBalance)

		changes[postBalance.Mint] = &model.BalanceChange{
			Before: new(big.Int).SetUint64(preBalance),
			After:  new(big.Int).SetUint64(postAmount),
			Change: new(big.Int).SetInt64(change),
		}
	}

	return changes
}

// IsSupportedToken 检查是否为支持的代币
func (ta *TransactionAdapter) IsSupportedToken(mint string) bool {
	// 检查是否为基础代币
	if config.IsBaseToken(mint) {
		return true
	}

	// 检查是否在代币映射中
	if _, exists := config.SOLANA_DEX_ADDRESS_TO_NAME[mint]; exists {
		return true
	}

	return false
}

// GetInstructions 获取所有指令
func (ta *TransactionAdapter) GetInstructions() []model.TransactionInstruction {
	return ta.transaction.Transaction.Transaction.Message.Instructions
}

// GetInnerInstructions 获取内部指令
func (ta *TransactionAdapter) GetInnerInstructions() []model.InnerInstruction {
	if ta.transaction.Meta != nil {
		return ta.transaction.Meta.InnerInstructions
	}
	return []model.InnerInstruction{}
}

// GetAccountKeys 获取账户密钥列表
func (ta *TransactionAdapter) GetAccountKeys() []string {
	return ta.transaction.Transaction.Transaction.Message.AccountKeys
}

// GetLogMessages 获取日志消息
func (ta *TransactionAdapter) GetLogMessages() []string {
	if ta.transaction.Meta != nil {
		return ta.transaction.Meta.LogMessages
	}
	return []string{}
}

// GetSignature 获取交易签名
func (ta *TransactionAdapter) GetSignature() string {
	if len(ta.transaction.Transaction.Transaction.Signatures) > 0 {
		return ta.transaction.Transaction.Transaction.Signatures[0]
	}
	return ""
}

// GetSlot 获取区块槽位
func (ta *TransactionAdapter) GetSlot() uint64 {
	return ta.transaction.Slot
}

// GetBlockTime 获取区块时间
func (ta *TransactionAdapter) GetBlockTime() uint64 {
	if ta.transaction.BlockTime != nil {
		return uint64(*ta.transaction.BlockTime)
	}
	return 0
}

// IsSuccess 检查交易是否成功
func (ta *TransactionAdapter) IsSuccess() bool {
	return ta.transaction.Meta == nil || ta.transaction.Meta.Err == nil
}

// GetProgramIdFromInstruction 从指令获取程序 ID
func (ta *TransactionAdapter) GetProgramIdFromInstruction(instruction model.TransactionInstruction) string {
	accountKeys := ta.GetAccountKeys()
	if instruction.ProgramIdIndex < len(accountKeys) {
		return accountKeys[instruction.ProgramIdIndex]
	}
	return ""
}
