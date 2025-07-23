package model

// TokenPnLData 代币盈亏数据
type TokenPnLData struct {
	TokenAddress     string  // 代币地址
	TotalBuyAmount   float64 // 总买入数量
	TotalSellAmount  float64 // 总卖出数量
	TotalBuyValue    float64 // 总买入价值（SOL本位）
	TotalSellValue   float64 // 总卖出价值（SOL本位）
	AvgBuyPrice      float64 // 平均买入价格
	AvgSellPrice     float64 // 平均卖出价格
	RealizedPnL      float64 // 已实现盈亏
	UnrealizedPnL    float64 // 未实现盈亏
	CurrentHolding   float64 // 当前持仓
	MaxHoldingValue  float64 // 历史最高持仓价值
	MaxHoldingAmount float64 // 历史最高持仓数量
	MaxHoldingUSD    float64 // 历史最高持仓USD价值
}

// UserReportSummary 用户报告汇总
type UserReportSummary struct {
	TotalUsers      int64   `json:"total_users"`
	ProfitableUsers int64   `json:"profitable_users"`
	LossUsers       int64   `json:"loss_users"`
	ProfitableRatio float64 `json:"profitable_ratio"`
}

// TransactionPriceData 交易价格数据
type TransactionPriceData struct {
	TokenAddress string
	BlockHeight  uint64
	TokenAmount  float64
	QuoteAmount  float64
	QuoteAddress string
	USDPrice     float64
}
