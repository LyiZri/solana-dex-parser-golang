package model

import "math/big"

// ParseConfig 解析配置
type ParseConfig struct {
	TryUnknownDEX    bool     `json:"try_unknown_dex"`
	ProgramIDs       []string `json:"program_ids"`
	IgnoreProgramIDs []string `json:"ignore_program_ids"`
	ThrowError       bool     `json:"throw_error"`
}

// SnapShotConfig 快照配置
type SnapShotConfig struct {
	MinTransactionAmount *big.Float `json:"min_transaction_amount"`
}

// SNAP_SHOT_CONFIG 快照配置实例
var SNAP_SHOT_CONFIG = SnapShotConfig{
	MinTransactionAmount: big.NewFloat(1.0), // 最小交易金额 1 USD
}

// TradeType 交易类型
type TradeType string

const (
	TradeTypeBuy  TradeType = "BUY"
	TradeTypeSell TradeType = "SELL"
)

// SwapTradeType 与 TS 版本对应的交易类型
type SwapTradeType string

const (
	ESwapTradeTypeBUY  SwapTradeType = "BUY"
	ESwapTradeTypeSELL SwapTradeType = "SELL"
)

// TokenAmount 代币数量结构
type TokenAmount struct {
	Amount   string  `json:"amount"`
	UIAmount float64 `json:"ui_amount"`
	Decimals uint8   `json:"decimals"`
}

// BalanceChange 余额变化
type BalanceChange struct {
	Before *big.Int `json:"before"`
	After  *big.Int `json:"after"`
	Change *big.Int `json:"change"`
}

// DexInfo DEX 信息
type DexInfo struct {
	ProgramID string `json:"program_id"`
	AMM       string `json:"amm"`
}

// TransferData 转账数据
type TransferData struct {
	Index            int          `json:"index"`
	InstructionIndex int          `json:"instruction_index"`
	Info             TransferInfo `json:"info"`
}

// TransferInfo 转账信息
type TransferInfo struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Owner       string `json:"owner"`
	Mint        string `json:"mint"`
	Amount      string `json:"amount"`
	Decimals    uint8  `json:"decimals"`
}

// TradeInfo 交易信息
type TradeInfo struct {
	Signature      string    `json:"signature"`
	Type           TradeType `json:"type"`
	Signer         string    `json:"signer"`
	TokenInMint    string    `json:"token_in_mint"`
	TokenInSymbol  string    `json:"token_in_symbol"`
	TokenInAmount  string    `json:"token_in_amount"`
	TokenOutMint   string    `json:"token_out_mint"`
	TokenOutSymbol string    `json:"token_out_symbol"`
	TokenOutAmount string    `json:"token_out_amount"`
	SlotNumber     uint64    `json:"slot_number"`
	BlockTime      uint64    `json:"block_time"`
	AMM            string    `json:"amm"`
	AMMs           []string  `json:"amms"`
	Route          string    `json:"route"`
	IDX            string    `json:"idx"`
}

// PoolEvent 流动性池事件
type PoolEvent struct {
	Signature   string `json:"signature"`
	Type        string `json:"type"`
	Signer      string `json:"signer"`
	PoolAddress string `json:"pool_address"`
	SlotNumber  uint64 `json:"slot_number"`
	BlockTime   uint64 `json:"block_time"`
	AMM         string `json:"amm"`
}
