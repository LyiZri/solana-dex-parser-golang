package mysql

import (
	"time"

	"github.com/shopspring/decimal"
)

type SmartSeasonOne struct {
	ID                     int64           `json:"id" gorm:"column:id"`
	UserId                 int64           `json:"user_id" gorm:"column:user_id"`
	UserAddr               string          `json:"user_addr" gorm:"column:user_addr"`
	AirdropStatus          int64           `json:"airdrop_status" gorm:"column:airdrop_status"`
	FirstTx                int64           `json:"first_tx" gorm:"column:first_tx"`
	FirstTokenSymbol       string          `json:"first_token_symbol" gorm:"column:first_token_symbol"`
	FirstTokenAddr         string          `json:"first_token_addr" gorm:"column:first_token_addr"`
	FirstTokenAmount       decimal.Decimal `json:"first_token_amount" gorm:"column:first_token_amount"`
	FirstSolAmount         decimal.Decimal `json:"first_sol_amount" gorm:"column:first_sol_amount"`
	FirstTxEvent           string          `json:"first_tx_event" gorm:"column:first_tx_event"`
	TokenCount             int64           `json:"token_count" gorm:"column:token_count"`
	TokenWinCount          int64           `json:"token_win_count" gorm:"column:token_win_count"`
	TokenLossCount         int64           `json:"token_loss_count" gorm:"column:token_loss_count"`
	WinRate                decimal.Decimal `json:"win_rate" gorm:"column:win_rate"`
	TxCount                int64           `json:"tx_count" gorm:"column:tx_count"`
	TxBuyCount             int64           `json:"tx_buy_count" gorm:"column:tx_buy_count"`
	TxSellCount            int64           `json:"tx_sell_count" gorm:"column:tx_sell_count"`
	TxAmountUsd            decimal.Decimal `json:"tx_amount_usd" gorm:"column:tx_amount_usd"`
	TxBuyAmountUsd         decimal.Decimal `json:"tx_buy_amount_usd" gorm:"column:tx_buy_amount_usd"`
	TxSellAmountUsd        decimal.Decimal `json:"tx_sell_amount_usd" gorm:"column:tx_sell_amount_usd"`
	MostHoldTokenSymbol    string          `json:"most_hold_token_symbol" gorm:"column:most_hold_token_symbol"`
	MostHoldTokenAddr      string          `json:"most_hold_token_addr" gorm:"column:most_hold_token_addr"`
	MostHoldTokenAmountUsd decimal.Decimal `json:"most_hold_token_amount_usd" gorm:"column:most_hold_token_amount_usd"`
	MostWalletHoldUsd      decimal.Decimal `json:"most_wallet_hold_usd" gorm:"column:most_wallet_hold_usd"`
	MostEarnTokenSymbol    string          `json:"most_earn_token_symbol" gorm:"column:most_earn_token_symbol"`
	MostEarnTokenAddr      string          `json:"most_earn_token_addr" gorm:"column:most_earn_token_addr"`
	MostEarnTokenAmountUsd decimal.Decimal `json:"most_earn_token_amount_usd" gorm:"column:most_earn_token_amount_usd"`
	MostEarnTokenWinRate   decimal.Decimal `json:"most_earn_token_win_rate" gorm:"column:most_earn_token_win_rate"`
	MostLossTokenSymbol    string          `json:"most_loss_token_symbol" gorm:"column:most_loss_token_symbol"`
	MostLossTokenAmountUsd decimal.Decimal `json:"most_loss_token_amount_usd" gorm:"column:most_loss_token_amount_usd"`
	TotalPnlUsd            decimal.Decimal `json:"total_pnl_usd" gorm:"column:total_pnl_usd"`
	MetricL50              int64           `json:"metric_l50" gorm:"column:metric_l50"`
	MetricL50L0            int64           `json:"metric_l50_l0" gorm:"column:metric_l50_l0"`
	MetricE0E200           int64           `json:"metric_e0_e200" gorm:"column:metric_e0_e200"`
	MetricE200E500         int64           `json:"metric_e200_e500" gorm:"column:metric_e200_e500"`
	MetricE500E1000        int64           `json:"metric_e500_e1000" gorm:"column:metric_e500_e1000"`
	MetricE1000            int64           `json:"metric_e1000" gorm:"column:metric_e1000"`
	Title                  string          `json:"title" gorm:"column:title"`
	SmartBox               int64           `json:"smart_box" gorm:"column:smart_box"`
	CreatedAt              time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt              time.Time       `json:"updated_at" gorm:"column:updated_at"`
}
