package model

import (
	"math/big"
)

// ParseResult 解析结果
type ParseResult struct {
	State              bool                      `json:"state"`
	Fee                TokenAmount               `json:"fee"`
	Trades             []TradeInfo               `json:"trades"`
	Liquidities        []PoolEvent               `json:"liquidities"`
	Transfers          []TransferData            `json:"transfers"`
	SOLBalanceChange   *BalanceChange            `json:"sol_balance_change"`
	TokenBalanceChange map[string]*BalanceChange `json:"token_balance_change"`
	MoreEvents         map[string]interface{}    `json:"more_events"`
	Msg                string                    `json:"msg"`
	Result             EnhancedResult            `json:"result"`
}

// EnhancedResult 增强的解析结果
type EnhancedResult struct {
	Trades             []ResSwapStruct               `json:"trades"`
	Liquidities        []ResLpInfoStruct             `json:"liquidities"`
	Tokens             []ResTokenMetadataStruct      `json:"tokens"`
	TokenPrices        []ResTokenPriceStruct         `json:"token_prices"`
	UserTradingSummary []ResUserTradingSummaryStruct `json:"user_trading_summary"`
}

// ResSwapStruct 标准化交易结构（对应 TS 版本）
type ResSwapStruct struct {
	TransactionSignature string  `json:"transaction_signature"`
	BlockTime            uint64  `json:"block_time"`
	UserAddress          string  `json:"user_address"`
	TokenInMint          string  `json:"token_in_mint"`
	TokenInSymbol        string  `json:"token_in_symbol"`
	TokenInAmount        float64 `json:"token_in_amount"`
	TokenOutMint         string  `json:"token_out_mint"`
	TokenOutSymbol       string  `json:"token_out_symbol"`
	TokenOutAmount       float64 `json:"token_out_amount"`
	PoolAddress          string  `json:"pool_address"`
	AMM                  string  `json:"amm"`
	Route                string  `json:"route"`
	SlotNumber           uint64  `json:"slot_number"`
}

// ResLpInfoStruct 流动性池信息结构
type ResLpInfoStruct struct {
	Signature   string `json:"signature"`
	Type        string `json:"type"`
	Signer      string `json:"signer"`
	PoolAddress string `json:"pool_address"`
	SlotNumber  uint64 `json:"slot_number"`
	BlockTime   uint64 `json:"block_time"`
	AMM         string `json:"amm"`
}

// ResTokenMetadataStruct 代币元数据结构
type ResTokenMetadataStruct struct {
	Mint     string `json:"mint"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Decimals uint8  `json:"decimals"`
}

// ResTokenPriceStruct 代币价格结构
type ResTokenPriceStruct struct {
	Mint      string     `json:"mint"`
	Symbol    string     `json:"symbol"`
	USDPrice  *big.Float `json:"usd_price"`
	Timestamp uint64     `json:"timestamp"`
}

// ResUserTradingSummaryStruct 用户交易汇总结构
type ResUserTradingSummaryStruct struct {
	UserAddress   string     `json:"user_address"`
	TotalVolume   *big.Float `json:"total_volume"`
	TradeCount    int        `json:"trade_count"`
	TokensTraded  []string   `json:"tokens_traded"`
	LastTradeTime uint64     `json:"last_trade_time"`
}

// SwapTransaction 最终标准化的交易结构（对应 TS 版本的 SwapTransaction）
type SwapTransaction struct {
	TxHash          string  `json:"tx_hash"`
	TransactionTime uint64  `json:"transaction_time"`
	WalletAddress   string  `json:"wallet_address"`
	TokenAmount     float64 `json:"token_amount"`
	TokenSymbol     string  `json:"token_symbol"`
	TokenAddress    string  `json:"token_address"`
	QuoteSymbol     string  `json:"quote_symbol"`
	QuoteAmount     float64 `json:"quote_amount"`
	QuoteAddress    string  `json:"quote_address"`
	QuotePrice      string  `json:"quote_price"`
	USDPrice        string  `json:"usd_price"`
	USDAmount       string  `json:"usd_amount"`
	TradeType       string  `json:"trade_type"`
	PoolAddress     string  `json:"pool_address"`
	BlockHeight     uint64  `json:"block_height"`
}

// VersionedBlockResponse 版本化区块响应（与现有 solana.Block 兼容）
type VersionedBlockResponse = Block

// SolanaTransaction Solana 交易结构
type SolanaTransaction struct {
	Transaction Transaction      `json:"transaction"`
	Meta        *TransactionMeta `json:"meta"`
	BlockTime   *int64           `json:"blockTime"`
	Slot        uint64           `json:"slot"`
}

// TokenSwapFilterData 代币交换过滤数据（对应 TS 版本）
type TokenSwapFilterData struct {
	UserAddress     string  `json:"user_address"`
	PoolAddress     string  `json:"pool_address"`
	TxHash          string  `json:"tx_hash"`
	IsBuy           bool    `json:"is_buy"`
	BlockHeight     uint64  `json:"block_height"`
	TokenSymbol     string  `json:"token_symbol"`
	TokenAddress    string  `json:"token_address"`
	QuoteSymbol     string  `json:"quote_symbol"`
	QuoteAddress    string  `json:"quote_address"`
	QuotePrice      float64 `json:"quote_price"`
	USDPrice        float64 `json:"usd_price"`
	USDAmount       float64 `json:"usd_amount"`
	TransactionTime uint64  `json:"transaction_time"`
	TokenAmount     float64 `json:"token_amount"`
	QuoteAmount     float64 `json:"quote_amount"`
}

// SwapTransactionToken 数据库查询结果结构（对应 TS 版本）
type SwapTransactionToken struct {
	TxHash          string  `json:"tx_hash"`
	TradeType       string  `json:"trade_type"`
	PoolAddress     string  `json:"pool_address"`
	BlockHeight     uint64  `json:"block_height"`
	TransactionTime uint64  `json:"transaction_time"`
	WalletAddress   string  `json:"wallet_address"`
	TokenAmount     float64 `json:"token_amount"`
	TokenSymbol     string  `json:"token_symbol"`
	TokenAddress    string  `json:"token_address"`
	QuoteSymbol     string  `json:"quote_symbol"`
	QuoteAmount     float64 `json:"quote_amount"`
	QuoteAddress    string  `json:"quote_address"`
	QuotePrice      float64 `json:"quote_price"`
	USDPrice        float64 `json:"usd_price"`
	USDAmount       float64 `json:"usd_amount"`
}

// ClassifiedInstruction 分类指令 - 对应 TS 版本的 ClassifiedInstruction
type ClassifiedInstruction struct {
	Instruction interface{} `json:"instruction"`
	ProgramID   string      `json:"programId"`
	OuterIndex  int         `json:"outerIndex"`
	InnerIndex  *int        `json:"innerIndex,omitempty"`
}
