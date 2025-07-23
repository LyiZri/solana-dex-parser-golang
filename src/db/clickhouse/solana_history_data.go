package clickhouse

import (
	"context"
	"fmt"
	"strconv"

	ckdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/go-solana-parse/src/config"
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
	query := "SELECT COUNT(*) FROM " + s.TableName()
	err := db.QueryRow(context.Background(), query).Scan(&count)
	return count, err
}

func (s *SolanaHistoryData) GetTotalUniqueAddress(db ckdriver.Conn) ([]string, error) {
	query := "SELECT DISTINCT wallet_address FROM " + s.TableName()
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []string
	for rows.Next() {
		var address string
		err := rows.Scan(&address)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}

func (s *SolanaHistoryData) GetTotalUniqueAddressLength(db ckdriver.Conn) (int, error) {
	addressList, err := s.GetTotalUniqueAddress(db)
	if err != nil {
		return 0, err
	}

	return len(addressList), nil
}

func (s *SolanaHistoryData) GetUserTransactionsByAddress(db ckdriver.Conn, address string) ([]*SolanaHistoryData, error) {
	query := `
		SELECT tx_hash, trade_type, pool_address, block_height, transaction_time,
			   wallet_address, token_amount, token_symbol, token_address,
			   quote_symbol, quote_amount, quote_address, toString(quote_price),
			   toString(usd_price), toString(usd_amount)
		FROM ` + s.TableName() + `
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

	return transactions, nil
}

func (s *SolanaHistoryData) GetUserTransactionsByAddressAndTimeRange(db ckdriver.Conn, address string, startTime, endTime uint64) ([]*SolanaHistoryData, error) {
	query := `
		SELECT tx_hash, trade_type, pool_address, block_height, transaction_time,
			   wallet_address, token_amount, token_symbol, token_address,
			   quote_symbol, quote_amount, quote_address, toString(quote_price),
			   toString(usd_price), toString(usd_amount)
		FROM ` + s.TableName() + `
		WHERE wallet_address = ?
		  AND transaction_time BETWEEN ? AND ?
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

	return transactions, nil
}

func (s *SolanaHistoryData) GetTokenTransactionsByAddressAndToken(db ckdriver.Conn, walletAddress, tokenAddress string) ([]*SolanaHistoryData, error) {
	query := `
		SELECT tx_hash, trade_type, pool_address, block_height, transaction_time,
			   wallet_address, token_amount, token_symbol, token_address,
			   quote_symbol, quote_amount, quote_address, toString(quote_price),
			   toString(usd_price), toString(usd_amount)
		FROM ` + s.TableName() + `
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

	return transactions, nil
}

func (s *SolanaHistoryData) GetUniqueTokensByAddress(db ckdriver.Conn, address string) ([]string, error) {
	query := `
		SELECT DISTINCT token_address
		FROM ` + s.TableName() + `
		WHERE wallet_address = ?
	`

	rows, err := db.Query(context.Background(), query, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		err := rows.Scan(&token)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (s *SolanaHistoryData) GetUserFirstTransaction(db ckdriver.Conn, address string) (*SolanaHistoryData, error) {
	query := `
		SELECT tx_hash, trade_type, pool_address, block_height, transaction_time,
			   wallet_address, token_amount, token_symbol, token_address,
			   quote_symbol, quote_amount, quote_address, toString(quote_price),
			   toString(usd_price), toString(usd_amount)
		FROM ` + s.TableName() + `
		WHERE wallet_address = ?
		ORDER BY transaction_time ASC
		LIMIT 1
	`

	row := db.QueryRow(context.Background(), query, address)

	tx := &SolanaHistoryData{}
	var quotePriceStr, usdPriceStr, usdAmountStr string

	err := row.Scan(
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

	return tx, nil
}

// 通过查找在该区块高度之前的最后一笔交易来推导价格
func (s *SolanaHistoryData) GetTokenPriceAtBlock(db ckdriver.Conn, tokenAddress string, blockHeight uint64) (*SolanaHistoryData, error) {
	query := `
		SELECT tx_hash, trade_type, pool_address, block_height, transaction_time,
			   wallet_address, token_amount, token_symbol, token_address,
			   quote_symbol, quote_amount, quote_address, toString(quote_price),
			   toString(usd_price), toString(usd_amount)
		FROM ` + s.TableName() + `
		WHERE token_address = ? 
		  AND block_height <= ?
		ORDER BY block_height DESC, transaction_time DESC
		LIMIT 1
	`

	row := db.QueryRow(context.Background(), query, tokenAddress, blockHeight)

	tx := &SolanaHistoryData{}
	var quotePriceStr, usdPriceStr, usdAmountStr string

	err := row.Scan(
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

	return tx, nil
}

// ===== SOL价格相关方法 =====

// GetSOLPriceFromTransactions 从交易数据中查询SOL价格
func (s *SolanaHistoryData) GetSOLPriceFromTransactions(db ckdriver.Conn, blockHeight uint64) (float64, error) {
	query := `
		SELECT quote_amount, token_amount, toString(usd_price)
		FROM ` + s.TableName() + `
		WHERE (quote_address = ? OR token_address = ?)
		  AND (token_address = ? OR quote_address = ?)
		  AND block_height <= ?
		  AND trade_type IS NOT NULL
		ORDER BY block_height DESC, transaction_time DESC
		LIMIT 1
	`

	wsolAddress := config.WSOL_ADDRESS
	usdcAddress := config.USDC_ADDRESS

	row := db.QueryRow(context.Background(), query,
		wsolAddress, wsolAddress, usdcAddress, usdcAddress, blockHeight)

	var quoteAmount, tokenAmount float64
	var usdPriceStr string
	err := row.Scan(&quoteAmount, &tokenAmount, &usdPriceStr)
	if err != nil {
		return 0.0, fmt.Errorf("未找到SOL-USDC交易记录: %v", err)
	}

	if tokenAmount > 0 {
		return tokenAmount / quoteAmount, nil
	}

	return 0.0, fmt.Errorf("无效的交易数据")
}

// GetSOLTransactionsInRange 获取指定区块范围内的SOL-USDC交易数据
func (s *SolanaHistoryData) GetSOLTransactionsInRange(db ckdriver.Conn, startBlock, endBlock uint64) ([]SOLTransactionData, error) {
	query := `
		SELECT DISTINCT block_height, quote_amount, token_amount, toString(usd_price)
		FROM ` + s.TableName() + `
		WHERE (quote_address = ? OR token_address = ?)
		  AND (token_address = ? OR quote_address = ?)
		  AND block_height BETWEEN ? AND ?
		  AND trade_type IS NOT NULL
		ORDER BY block_height ASC
	`

	wsolAddress := config.WSOL_ADDRESS
	usdcAddress := config.USDC_ADDRESS

	rows, err := db.Query(context.Background(), query,
		wsolAddress, wsolAddress, usdcAddress, usdcAddress, startBlock, endBlock)
	if err != nil {
		return nil, fmt.Errorf("查询SOL-USDC交易失败: %v", err)
	}
	defer rows.Close()

	var transactions []SOLTransactionData
	for rows.Next() {
		var tx SOLTransactionData
		var usdPriceStr string

		err := rows.Scan(&tx.BlockHeight, &tx.QuoteAmount, &tx.TokenAmount, &usdPriceStr)
		if err != nil {
			continue
		}

		// 转换Decimal字段
		tx.UsdPrice, err = parseDecimalToFloat64(usdPriceStr)
		if err != nil {
			// 如果转换失败，设为0继续处理
			tx.UsdPrice = 0
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// ===== 数据结构定义 =====

// SOLTransactionData SOL交易数据结构
type SOLTransactionData struct {
	BlockHeight uint64  `json:"block_height"`
	QuoteAmount float64 `json:"quote_amount"`
	TokenAmount float64 `json:"token_amount"`
	UsdPrice    float64 `json:"usd_price"`
}

// CalculateSOLPriceFromTransaction 从交易数据计算SOL价格
func (tx *SOLTransactionData) CalculateSOLPriceFromTransaction() (float64, error) {
	// 如果有直接的USD价格，使用它
	if tx.UsdPrice > 0 {
		return tx.UsdPrice, nil
	}

	// 否则根据交易比例计算（USDC视为1美元）
	if tx.TokenAmount > 0 {
		return tx.QuoteAmount / tx.TokenAmount, nil
	}

	return 0.0, fmt.Errorf("无效的交易数据")
}
