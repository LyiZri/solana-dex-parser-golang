package db

import "gorm.io/gorm"

type UserReport struct {
	Address                   string `json:"address" gorm:"column:address"`
	ProfitRate                string `json:"profit_rate" gorm:"column:profit_rate"`
	FirstTradeTime            string `json:"first_trade_time" gorm:"column:first_trade_time"`
	FirstTradeTokenAddress    string `json:"first_trade_token_address" gorm:"column:first_trade_token_address"`
	FirstTradeTokenAmount     string `json:"first_trade_token_amount" gorm:"column:first_trade_token_amount"`
	TotalBuyVolumn            string `json:"total_buy_volumn" gorm:"column:total_buy_volumn"`
	TotalSellVolumn           string `json:"total_sell_volumn" gorm:"column:total_sell_volumn"`
	TotalVolumn               string `json:"total_volumn" gorm:"column:total_volumn"`
	BuyCount                  int64  `json:"buy_count" gorm:"column:buy_count"`
	SellCount                 int64  `json:"sell_count" gorm:"column:sell_count"`
	TradeCount                int64  `json:"trade_count" gorm:"column:trade_count"`
	UniqueTokenCount          int64  `json:"unique_token_count" gorm:"column:unique_token_count"`
	PnlWinCount               int64  `json:"pnl_win_count" gorm:"column:pnl_win_count"`
	PnlLossCount              int64  `json:"pnl_loss_count" gorm:"column:pnl_loss_count"`
	PnlWinRate                string `json:"pnl_win_rate" gorm:"column:pnl_win_rate"`
	TopProfitTokenAddress     string `json:"top_profit_token_address" gorm:"column:top_profit_token_address"`
	TopLossTokenAddress       string `json:"top_loss_token_address" gorm:"column:top_loss_token_address"`
	WinLevelOneCount          int64  `json:"win_level_one_count" gorm:"column:win_level_one_count"`
	WinLevelTwoCount          int64  `json:"win_level_two_count" gorm:"column:win_level_two_count"`
	WinLevelThreeCount        int64  `json:"win_level_three_count" gorm:"column:win_level_three_count"`
	WinLevelFourCount         int64  `json:"win_level_four_count" gorm:"column:win_level_four_count"`
	LossLevelOneCount         int64  `json:"loss_level_one_count" gorm:"column:loss_level_one_count"`
	LossLevelTwoCount         int64  `json:"loss_level_two_count" gorm:"column:loss_level_two_count"`
	MostHoldValueTokenAddress string `json:"most_hold_value_token_address" gorm:"column:most_hold_value_token_address"`
	MostHoldValueTokenAmount  string `json:"most_hold_value_token_amount" gorm:"column:most_hold_value_token_amount"`
	MostHoldValueTokenUSD     string `json:"most_hold_value_token_usd" gorm:"column:most_hold_value_token_usd"`
	MaxTotalHoldValue         string `json:"max_total_hold_value" gorm:"column:max_total_hold_value"`
}

var UserReportN = &UserReport{}

func (p *UserReport) TableName() string {
	return "user_report"
}

func (p *UserReport) GetUserReport(db *gorm.DB, address string) (*UserReport, error) {
	err := db.Where("address = ?", address).First(p).Error
	if err != nil {
		return nil, err
	}
	return p, nil
}
