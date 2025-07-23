package mysql

import (
	"fmt"
	"log"
	"time"

	"github.com/go-solana-parse/src/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// type UserReport struct {
// 	Address                   string `json:"address" gorm:"column:address"`
// 	ProfitRate                string `json:"profit_rate" gorm:"column:profit_rate"`
// 	FirstTradeTime            string `json:"first_trade_time" gorm:"column:first_trade_time"`
// 	FirstTradeTokenAddress    string `json:"first_trade_token_address" gorm:"column:first_trade_token_address"`
// 	FirstTradeTokenAmount     string `json:"first_trade_token_amount" gorm:"column:first_trade_token_amount"`
// 	TotalBuyVolumn            string `json:"total_buy_volumn" gorm:"column:total_buy_volumn"`
// 	TotalSellVolumn           string `json:"total_sell_volumn" gorm:"column:total_sell_volumn"`
// 	TotalVolumn               string `json:"total_volumn" gorm:"column:total_volumn"`
// 	BuyCount                  int64  `json:"buy_count" gorm:"column:buy_count"`
// 	SellCount                 int64  `json:"sell_count" gorm:"column:sell_count"`
// 	TradeCount                int64  `json:"trade_count" gorm:"column:trade_count"`
// 	UniqueTokenCount          int64  `json:"unique_token_count" gorm:"column:unique_token_count"`
// 	PnlWinCount               int64  `json:"pnl_win_count" gorm:"column:pnl_win_count"`
// 	PnlLossCount              int64  `json:"pnl_loss_count" gorm:"column:pnl_loss_count"`
// 	PnlWinRate                string `json:"pnl_win_rate" gorm:"column:pnl_win_rate"`
// 	TopProfitTokenAddress     string `json:"top_profit_token_address" gorm:"column:top_profit_token_address"`
// 	TopLossTokenAddress       string `json:"top_loss_token_address" gorm:"column:top_loss_token_address"`
// 	TopProfitUsdAmount        string `json:"top_profit_usd_amount" gorm:"column:top_profit_usd_amount"`
// 	TopLossUsdAmount          string `json:"top_loss_usd_amount" gorm:"column:top_loss_usd_amount"`
// 	TopProfitWinRate          string `json:"top_profit_win_rate" gorm:"column:top_profit_win_rate"`
// 	WinLevelOneCount          int64  `json:"win_level_one_count" gorm:"column:win_level_one_count"`
// 	WinLevelTwoCount          int64  `json:"win_level_two_count" gorm:"column:win_level_two_count"`
// 	WinLevelThreeCount        int64  `json:"win_level_three_count" gorm:"column:win_level_three_count"`
// 	WinLevelFourCount         int64  `json:"win_level_four_count" gorm:"column:win_level_four_count"`
// 	LossLevelOneCount         int64  `json:"loss_level_one_count" gorm:"column:loss_level_one_count"`
// 	LossLevelTwoCount         int64  `json:"loss_level_two_count" gorm:"column:loss_level_two_count"`
// 	MostHoldValueTokenAddress string `json:"most_hold_value_token_address" gorm:"column:most_hold_value_token_address"`
// 	MostHoldValueTokenAmount  string `json:"most_hold_value_token_amount" gorm:"column:most_hold_value_token_amount"`
// 	MostHoldValueTokenUSD     string `json:"most_hold_value_token_usd" gorm:"column:most_hold_value_token_usd"`
// 	MaxTotalHoldValue         string `json:"max_total_hold_value" gorm:"column:max_total_hold_value"`
// }

type UserReport struct {
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
	MostLossTokenAddr      string          `json:"most_loss_token_addr" gorm:"column:most_loss_token_addr"`
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

// TableName 返回表名
func (u *UserReport) TableName() string {
	return "smart_season_1"
}

var UserReportNsp = &UserReport{}

// InsertUserReport 插入用户报告
func (p *UserReport) InsertUserReport(db *gorm.DB, userReport *UserReport) error {
	err := db.Create(userReport).Error
	if err != nil {
		return err
	}
	return nil
}

// SaveOrUpdateUserReport 保存或更新用户报告到 MySQL 数据库
func (p *UserReport) SaveOrUpdateUserReport(db *gorm.DB, userReport *UserReport) error {
	if db == nil {
		log.Printf("MySQL 数据库连接为空，跳过保存用户报告: %s\n", userReport.UserAddr)
		return nil
	}

	// 检查用户报告是否已存在
	existingReport := &UserReport{}
	err := db.Where("user_addr = ?", userReport.UserAddr).First(existingReport).Error

	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		err = db.Create(userReport).Error
		if err != nil {
			return fmt.Errorf("创建用户报告失败: %v", err)
		}
		log.Printf("创建新的用户报告: %s\n", userReport.UserAddr)
	} else if err != nil {
		return fmt.Errorf("查询用户报告失败: %v", err)
	} else {
		// 已存在，更新记录
		err = db.Model(existingReport).Where("user_addr = ?", userReport.UserAddr).Updates(userReport).Error
		if err != nil {
			return fmt.Errorf("更新用户报告失败: %v", err)
		}
		log.Printf("更新用户报告: %s\n", userReport.UserAddr)
	}

	return nil
}

// GetUserReportSummary 获取用户报告汇总信息
func (p *UserReport) GetUserReportSummary(db *gorm.DB) (*model.UserReportSummary, error) {
	if db == nil {
		return nil, fmt.Errorf("MySQL 数据库连接为空")
	}

	summary := &model.UserReportSummary{}

	// 统计总用户数
	err := db.Model(&UserReport{}).Count(&summary.TotalUsers).Error
	if err != nil {
		return nil, fmt.Errorf("统计总用户数失败: %v", err)
	}

	// 统计盈利用户数
	err = db.Model(&UserReport{}).Where("profit_rate > ?", "0").Count(&summary.ProfitableUsers).Error
	if err != nil {
		return nil, fmt.Errorf("统计盈利用户数失败: %v", err)
	}

	// 统计亏损用户数
	err = db.Model(&UserReport{}).Where("profit_rate < ?", "0").Count(&summary.LossUsers).Error
	if err != nil {
		return nil, fmt.Errorf("统计亏损用户数失败: %v", err)
	}

	// 计算盈利用户比例
	if summary.TotalUsers > 0 {
		summary.ProfitableRatio = float64(summary.ProfitableUsers) / float64(summary.TotalUsers)
	}

	return summary, nil
}

// GetUserReportByAddress 根据地址获取用户报告
func (p *UserReport) GetUserReportByAddress(db *gorm.DB, address string) (*UserReport, error) {
	if db == nil {
		return nil, fmt.Errorf("MySQL 数据库连接为空")
	}

	userReport := &UserReport{}
	err := db.Where("address = ?", address).First(userReport).Error
	if err != nil {
		return nil, err
	}

	return userReport, nil
}

// GetAllUserReports 获取所有用户报告
func (p *UserReport) GetAllUserReports(db *gorm.DB, limit, offset int) ([]*UserReport, error) {
	if db == nil {
		return nil, fmt.Errorf("MySQL 数据库连接为空")
	}

	var userReports []*UserReport
	query := db.Model(&UserReport{})

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&userReports).Error
	if err != nil {
		return nil, fmt.Errorf("获取用户报告列表失败: %v", err)
	}

	return userReports, nil
}

// DeleteUserReport 删除用户报告
func (p *UserReport) DeleteUserReport(db *gorm.DB, address string) error {
	if db == nil {
		return fmt.Errorf("MySQL 数据库连接为空")
	}

	err := db.Where("address = ?", address).Delete(&UserReport{}).Error
	if err != nil {
		return fmt.Errorf("删除用户报告失败: %v", err)
	}

	return nil
}

func (p *UserReport) GetCount(db *gorm.DB) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("MySQL 数据库连接为空")
	}

	var count int64
	err := db.Model(&UserReport{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("获取用户报告数量失败: %v", err)
	}

	return count, nil
}
