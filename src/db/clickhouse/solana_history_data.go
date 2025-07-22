package clickhouse

import (
	"context"
	"strconv"

	ckdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type SolanaHistoryData struct {
	TxHash          string  `json:"tx_hash" gorm:"column:tx_hash"`
	TradeType       string  `json:"trade_type" gorm:"column:trade_type"`
	PoolAddress     string  `json:"pool_address" gorm:"column:pool_address"`
	BlockHeight     uint64  `json:"block_height" gorm:"column:block_height"`
	TransactionTime uint64  `json:"transaction_time" gorm:"column:transaction_time"`
	WalletAddress   string  `json:"wallet_address" gorm:"column:wallet_address"`
	TokenAmount     float64 `json:"token_amount" gorm:"column:token_amount"`
	TokenSymbol     string  `json:"token_symbol" gorm:"column:token_symbol"`
	TokenAddress    string  `json:"token_address" gorm:"column:token_address"`
	QuoteSymbol     string  `json:"quote_symbol" gorm:"column:quote_symbol"`
	QuoteAmount     float64 `json:"quote_amount" gorm:"column:quote_amount"`
	QuoteAddress    string  `json:"quote_address" gorm:"column:quote_address"`
	QuotePrice      float64 `json:"quote_price" gorm:"column:quote_price"`
	UsdPrice        float64 `json:"usd_price" gorm:"column:usd_price"`
	UsdAmount       float64 `json:"usd_amount" gorm:"column:usd_amount"`
}

var SolanaHistoryDataNsp = &SolanaHistoryData{}

// parseDecimalToFloat64 将ClickHouse Decimal字符串转换为float64
func parseDecimalToFloat64(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

func (s *SolanaHistoryData) TableName() string {
	return "solana_history_data_new"
}

func (s *SolanaHistoryData) TableNameWithNamespace() string {
	return SolanaHistoryDataNsp.TableName()
}

func (s *SolanaHistoryData) GetTotalTxCount(db ckdriver.Conn) (uint64, error) {
	var count uint64
	err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM solana_history_data_new").Scan(&count)
	return count, err
}

func (s *SolanaHistoryData) GetTotalUniqueAddress(db ckdriver.Conn) ([]string, error) {
	var addressList []string
	rows, err := db.Query(context.Background(), "SELECT DISTINCT wallet_address FROM solana_history_data_new")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var address string
		if err := rows.Scan(&address); err != nil {
			return nil, err
		}
		addressList = append(addressList, address)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return addressList, nil
}

func (s *SolanaHistoryData) GetTotalUniqueAddressLength(db ckdriver.Conn) (int, error) {
	addressList, err := s.GetTotalUniqueAddress(db)
	if err != nil {
		return 0, err
	}
	return len(addressList), nil
}

// GetUserTransactionsByAddress 根据用户地址获取交易记录
func (s *SolanaHistoryData) GetUserTransactionsByAddress(db ckdriver.Conn, address string) ([]*SolanaHistoryData, error) {
	query := `
		SELECT tx_hash, trade_type, pool_address, block_height, transaction_time,
			   wallet_address, token_amount, token_symbol, token_address,
			   quote_symbol, quote_amount, quote_address, toString(quote_price),
			   toString(usd_price), toString(usd_amount)
		FROM solana_history_data_new 
		WHERE wallet_address = ? 
		ORDER BY transaction_time ASC
	`

	rows, err := db.Query(context.Background(), query, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*SolanaHistoryData
	for rows.Next() {
		tx := &SolanaHistoryData{}
		var quotePriceStr, usdPriceStr, usdAmountStr string
		
		err := rows.Scan(
			&tx.TxHash, &tx.TradeType, &tx.PoolAddress, &tx.BlockHeight,
			&tx.TransactionTime, &tx.WalletAddress, &tx.TokenAmount,
			&tx.TokenSymbol, &tx.TokenAddress, &tx.QuoteSymbol,
			&tx.QuoteAmount, &tx.QuoteAddress, &quotePriceStr,
			&usdPriceStr, &usdAmountStr,
		)
		if err != nil {
			return nil, err
		}
		
		// 转换Decimal字段
		if tx.QuotePrice, err = parseDecimalToFloat64(quotePriceStr); err != nil {
			return nil, err
		}
		if tx.UsdPrice, err = parseDecimalToFloat64(usdPriceStr); err != nil {
			return nil, err
		}
		if tx.UsdAmount, err = parseDecimalToFloat64(usdAmountStr); err != nil {
			return nil, err
		}
		
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

// GetUserTransactionsByAddressAndTimeRange 根据用户地址和时间范围获取交易记录
func (s *SolanaHistoryData) GetUserTransactionsByAddressAndTimeRange(db ckdriver.Conn, address string, startTime, endTime uint64) ([]*SolanaHistoryData, error) {
	query := `
		SELECT tx_hash, trade_type, pool_address, block_height, transaction_time,
			   wallet_address, token_amount, token_symbol, token_address,
			   quote_symbol, quote_amount, quote_address, quote_price,
			   usd_price, usd_amount
		FROM solana_history_data_new 
		WHERE wallet_address = ? 
		  AND transaction_time >= ? 
		  AND transaction_time <= ?
		ORDER BY transaction_time ASC
	`

	rows, err := db.Query(context.Background(), query, address, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*SolanaHistoryData
	for rows.Next() {
		tx := &SolanaHistoryData{}
		err := rows.Scan(
			&tx.TxHash, &tx.TradeType, &tx.PoolAddress, &tx.BlockHeight,
			&tx.TransactionTime, &tx.WalletAddress, &tx.TokenAmount,
			&tx.TokenSymbol, &tx.TokenAddress, &tx.QuoteSymbol,
			&tx.QuoteAmount, &tx.QuoteAddress, &tx.QuotePrice,
			&tx.UsdPrice, &tx.UsdAmount,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

// GetTokenTransactionsByAddressAndToken 根据用户地址和代币地址获取特定代币的交易记录
func (s *SolanaHistoryData) GetTokenTransactionsByAddressAndToken(db ckdriver.Conn, walletAddress, tokenAddress string) ([]*SolanaHistoryData, error) {
	query := `
		SELECT tx_hash, trade_type, pool_address, block_height, transaction_time,
			   wallet_address, token_amount, token_symbol, token_address,
			   quote_symbol, quote_amount, quote_address, quote_price,
			   usd_price, usd_amount
		FROM solana_history_data_new 
		WHERE wallet_address = ? 
		  AND token_address = ?
		ORDER BY transaction_time ASC
	`

	rows, err := db.Query(context.Background(), query, walletAddress, tokenAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*SolanaHistoryData
	for rows.Next() {
		tx := &SolanaHistoryData{}
		err := rows.Scan(
			&tx.TxHash, &tx.TradeType, &tx.PoolAddress, &tx.BlockHeight,
			&tx.TransactionTime, &tx.WalletAddress, &tx.TokenAmount,
			&tx.TokenSymbol, &tx.TokenAddress, &tx.QuoteSymbol,
			&tx.QuoteAmount, &tx.QuoteAddress, &tx.QuotePrice,
			&tx.UsdPrice, &tx.UsdAmount,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

// GetUniqueTokensByAddress 获取用户交易过的所有唯一代币地址
func (s *SolanaHistoryData) GetUniqueTokensByAddress(db ckdriver.Conn, address string) ([]string, error) {
	query := `
		SELECT DISTINCT token_address 
		FROM solana_history_data_new 
		WHERE wallet_address = ?
		ORDER BY token_address
	`

	rows, err := db.Query(context.Background(), query, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokenAddresses []string
	for rows.Next() {
		var tokenAddress string
		if err := rows.Scan(&tokenAddress); err != nil {
			return nil, err
		}
		tokenAddresses = append(tokenAddresses, tokenAddress)
	}

	return tokenAddresses, rows.Err()
}

// GetUserFirstTransaction 获取用户的第一笔交易记录
func (s *SolanaHistoryData) GetUserFirstTransaction(db ckdriver.Conn, address string) (*SolanaHistoryData, error) {
	query := `
		SELECT tx_hash, trade_type, pool_address, block_height, transaction_time,
			   wallet_address, token_amount, token_symbol, token_address,
			   quote_symbol, quote_amount, quote_address, quote_price,
			   usd_price, usd_amount
		FROM solana_history_data_new 
		WHERE wallet_address = ? 
		ORDER BY transaction_time ASC
		LIMIT 1
	`

	row := db.QueryRow(context.Background(), query, address)

	tx := &SolanaHistoryData{}
	err := row.Scan(
		&tx.TxHash, &tx.TradeType, &tx.PoolAddress, &tx.BlockHeight,
		&tx.TransactionTime, &tx.WalletAddress, &tx.TokenAmount,
		&tx.TokenSymbol, &tx.TokenAddress, &tx.QuoteSymbol,
		&tx.QuoteAmount, &tx.QuoteAddress, &tx.QuotePrice,
		&tx.UsdPrice, &tx.UsdAmount,
	)

	if err != nil {
		return nil, err
	}

	return tx, nil
}

// GetTokenPriceAtBlock 获取某个代币在指定区块高度的价格信息
// 通过查找在该区块高度之前的最后一笔交易来推导价格
func (s *SolanaHistoryData) GetTokenPriceAtBlock(db ckdriver.Conn, tokenAddress string, blockHeight uint64) (*SolanaHistoryData, error) {
	query := `
		SELECT tx_hash, trade_type, pool_address, block_height, transaction_time,
			   wallet_address, token_amount, token_symbol, token_address,
			   quote_symbol, quote_amount, quote_address, quote_price,
			   usd_price, usd_amount
		FROM solana_history_data_new 
		WHERE token_address = ? 
		  AND block_height <= ?
		ORDER BY block_height DESC, transaction_time DESC
		LIMIT 1
	`

	row := db.QueryRow(context.Background(), query, tokenAddress, blockHeight)

	tx := &SolanaHistoryData{}
	err := row.Scan(
		&tx.TxHash, &tx.TradeType, &tx.PoolAddress, &tx.BlockHeight,
		&tx.TransactionTime, &tx.WalletAddress, &tx.TokenAmount,
		&tx.TokenSymbol, &tx.TokenAddress, &tx.QuoteSymbol,
		&tx.QuoteAmount, &tx.QuoteAddress, &tx.QuotePrice,
		&tx.UsdPrice, &tx.UsdAmount,
	)

	if err != nil {
		return nil, err
	}

	return tx, nil
}
