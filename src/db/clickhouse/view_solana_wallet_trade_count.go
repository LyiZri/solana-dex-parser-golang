package clickhouse

import (
	"context"
	"fmt"

	ckdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// ViewSolanaWalletTradeCount 钱包交易统计view结构
type ViewSolanaWalletTradeCount struct {
	WalletAddress string `json:"wallet_address"`
	TradeCount    uint64 `json:"trade_count"`
}

var ViewSolanaWalletTradeCountNsp = &ViewSolanaWalletTradeCount{}

// GetWalletAddressesByTradeCountASC 从view表获取钱包地址，按交易量升序排序
func (v *ViewSolanaWalletTradeCount) GetWalletAddressesByTradeCountASC(db ckdriver.Conn) ([]string, error) {
	query := `
		SELECT wallet_address 
		FROM solana_data.solana_wallet_trade_count_view 
		ORDER BY trade_count ASC
	`

	rows, err := db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("查询钱包地址按交易量排序失败: %v", err)
	}
	defer rows.Close()

	var addresses []string
	for rows.Next() {
		var address string
		err := rows.Scan(&address)
		if err != nil {
			return nil, fmt.Errorf("扫描钱包地址失败: %v", err)
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}

// GetWalletTradeCountList 获取钱包地址和交易量列表，按交易量升序排序
func (v *ViewSolanaWalletTradeCount) GetWalletTradeCountList(db ckdriver.Conn) ([]*ViewSolanaWalletTradeCount, error) {
	query := `
		SELECT wallet_address, trade_count 
		FROM solana_data.solana_wallet_trade_count_view 
		ORDER BY trade_count ASC
	`

	rows, err := db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("查询钱包交易统计失败: %v", err)
	}
	defer rows.Close()

	var walletTradeCounts []*ViewSolanaWalletTradeCount
	for rows.Next() {
		var walletTradeCount ViewSolanaWalletTradeCount
		err := rows.Scan(&walletTradeCount.WalletAddress, &walletTradeCount.TradeCount)
		if err != nil {
			return nil, fmt.Errorf("扫描钱包交易统计失败: %v", err)
		}
		walletTradeCounts = append(walletTradeCounts, &walletTradeCount)
	}

	return walletTradeCounts, nil
}

// GetWalletAddressesByTradeCountWithLimit 从view表获取钱包地址，按交易量升序排序，支持分页
func (v *ViewSolanaWalletTradeCount) GetWalletAddressesByTradeCountWithLimit(db ckdriver.Conn, limit, offset int) ([]string, error) {
	query := `
		SELECT wallet_address 
		FROM solana_data.solana_wallet_trade_count_view 
		ORDER BY trade_count ASC
		LIMIT ? OFFSET ?
	`

	rows, err := db.Query(context.Background(), query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询钱包地址按交易量排序(分页)失败: %v", err)
	}
	defer rows.Close()

	var addresses []string
	for rows.Next() {
		var address string
		err := rows.Scan(&address)
		if err != nil {
			return nil, fmt.Errorf("扫描钱包地址失败: %v", err)
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}

// GetTotalWalletCount 获取总钱包数量
func (v *ViewSolanaWalletTradeCount) GetTotalWalletCount(db ckdriver.Conn) (uint64, error) {
	var count uint64
	query := "SELECT COUNT(*) FROM solana_data.solana_wallet_trade_count_view"

	err := db.QueryRow(context.Background(), query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询钱包总数失败: %v", err)
	}

	return count, nil
}
