package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	// 注释掉不可用的驱动，实际使用时可以安装这些包
	// _ "github.com/go-sql-driver/mysql"     // MySQL driver
	// _ "github.com/ClickHouse/clickhouse-go/v2" // ClickHouse driver

	"go-solana-explore/src/parser"
)

// SwapTransactionDB 交换交易数据库接口
type SwapTransactionDB interface {
	InsertSwapTransactions(ctx context.Context, transactions []parser.SwapTransaction) error
	GetSwapTransactions(ctx context.Context, filters SwapTransactionFilter) ([]parser.SwapTransactionToken, error)
	GetFilteredTokenData(ctx context.Context, filters TokenDataFilter) ([]parser.TokenSwapFilterData, error)
}

// SwapTransactionFilter 查询过滤器
type SwapTransactionFilter struct {
	StartTime    *time.Time
	EndTime      *time.Time
	WalletAddr   *string
	TokenAddr    *string
	MinUSDAmount *float64
	Limit        int
	Offset       int
}

// TokenDataFilter 代币数据过滤器
type TokenDataFilter struct {
	TokenAddresses []string
	StartTime      *time.Time
	EndTime        *time.Time
	MinUSDAmount   *float64
	Limit          int
}

// MySQLSwapTransactionDB MySQL 实现
type MySQLSwapTransactionDB struct {
	db *sql.DB
}

// ClickHouseSwapTransactionDB ClickHouse 实现
type ClickHouseSwapTransactionDB struct {
	db *sql.DB
}

// NewMySQLSwapTransactionDB 创建新的 MySQL 数据库实例
func NewMySQLSwapTransactionDB(db *sql.DB) *MySQLSwapTransactionDB {
	return &MySQLSwapTransactionDB{db: db}
}

// NewClickHouseSwapTransactionDB 创建新的 ClickHouse 数据库实例
func NewClickHouseSwapTransactionDB(db *sql.DB) *ClickHouseSwapTransactionDB {
	return &ClickHouseSwapTransactionDB{db: db}
}

// MySQL 实现

// InsertSwapTransactions MySQL 批量插入交换交易
func (m *MySQLSwapTransactionDB) InsertSwapTransactions(ctx context.Context, transactions []parser.SwapTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	// 对应 TS 版本的 insertToTokenTable 和 insertToWalletTable
	err := m.insertToTokenTable(ctx, transactions)
	if err != nil {
		return fmt.Errorf("failed to insert to token table: %w", err)
	}

	err = m.insertToWalletTable(ctx, transactions)
	if err != nil {
		return fmt.Errorf("failed to insert to wallet table: %w", err)
	}

	return nil
}

// insertToTokenTable 插入到代币表 (对应 TS 版本的 solana_swap_transactions_token)
func (m *MySQLSwapTransactionDB) insertToTokenTable(ctx context.Context, transactions []parser.SwapTransaction) error {
	query := `
	INSERT INTO solana_swap_transactions_token (
		tx_hash, transaction_time, wallet_address, token_amount, token_symbol, 
		token_address, quote_symbol, quote_amount, quote_address, quote_price, 
		usd_price, usd_amount, trade_type, pool_address, block_height
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		transaction_time = VALUES(transaction_time),
		token_amount = VALUES(token_amount),
		quote_amount = VALUES(quote_amount),
		usd_amount = VALUES(usd_amount)
	`

	stmt, err := m.db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, tx := range transactions {
		_, err := stmt.ExecContext(ctx,
			tx.TxHash, time.Unix(int64(tx.TransactionTime), 0), tx.WalletAddress,
			tx.TokenAmount, tx.TokenSymbol, tx.TokenAddress,
			tx.QuoteSymbol, tx.QuoteAmount, tx.QuoteAddress,
			tx.QuotePrice, tx.USDPrice, tx.USDAmount,
			tx.TradeType, tx.PoolAddress, tx.BlockHeight,
		)
		if err != nil {
			log.Printf("Failed to insert token transaction %s: %v", tx.TxHash, err)
			continue
		}
	}

	log.Printf("✅ 插入 %d 条记录到 solana_swap_transactions_token", len(transactions))
	return nil
}

// insertToWalletTable 插入到钱包表 (对应 TS 版本的 solana_swap_transactions_wallet)
func (m *MySQLSwapTransactionDB) insertToWalletTable(ctx context.Context, transactions []parser.SwapTransaction) error {
	query := `
	INSERT INTO solana_swap_transactions_wallet (
		tx_hash, transaction_time, wallet_address, token_amount, token_symbol, 
		token_address, quote_symbol, quote_amount, quote_address, quote_price, 
		usd_price, usd_amount, trade_type, pool_address, block_height
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		transaction_time = VALUES(transaction_time),
		token_amount = VALUES(token_amount),
		quote_amount = VALUES(quote_amount),
		usd_amount = VALUES(usd_amount)
	`

	stmt, err := m.db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, tx := range transactions {
		_, err := stmt.ExecContext(ctx,
			tx.TxHash, time.Unix(int64(tx.TransactionTime), 0), tx.WalletAddress,
			tx.TokenAmount, tx.TokenSymbol, tx.TokenAddress,
			tx.QuoteSymbol, tx.QuoteAmount, tx.QuoteAddress,
			tx.QuotePrice, tx.USDPrice, tx.USDAmount,
			tx.TradeType, tx.PoolAddress, tx.BlockHeight,
		)
		if err != nil {
			log.Printf("Failed to insert wallet transaction %s: %v", tx.TxHash, err)
			continue
		}
	}

	log.Printf("✅ 插入 %d 条记录到 solana_swap_transactions_wallet", len(transactions))
	return nil
}

// GetSwapTransactions MySQL 查询交换交易 (对应 TS 版本的 getXDaysData)
func (m *MySQLSwapTransactionDB) GetSwapTransactions(ctx context.Context, filters SwapTransactionFilter) ([]parser.SwapTransactionToken, error) {
	query := `
	SELECT tx_hash, trade_type, pool_address, block_height, 
		   UNIX_TIMESTAMP(transaction_time) as transaction_time,
		   wallet_address, token_amount, token_symbol, token_address,
		   quote_symbol, quote_amount, quote_address, quote_price,
		   usd_price, usd_amount
	FROM solana_swap_transactions_token
	WHERE 1=1
	`

	args := []interface{}{}

	if filters.StartTime != nil {
		query += " AND transaction_time >= ?"
		args = append(args, *filters.StartTime)
	}

	if filters.EndTime != nil {
		query += " AND transaction_time <= ?"
		args = append(args, *filters.EndTime)
	}

	if filters.WalletAddr != nil {
		query += " AND wallet_address = ?"
		args = append(args, *filters.WalletAddr)
	}

	if filters.TokenAddr != nil {
		query += " AND token_address = ?"
		args = append(args, *filters.TokenAddr)
	}

	if filters.MinUSDAmount != nil {
		query += " AND CAST(usd_amount AS DECIMAL(20,8)) >= ?"
		args = append(args, *filters.MinUSDAmount)
	}

	query += " ORDER BY transaction_time DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	if filters.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filters.Offset)
	}

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query swap transactions: %w", err)
	}
	defer rows.Close()

	var transactions []parser.SwapTransactionToken
	for rows.Next() {
		var tx parser.SwapTransactionToken
		err := rows.Scan(
			&tx.TxHash, &tx.TradeType, &tx.PoolAddress, &tx.BlockHeight,
			&tx.TransactionTime, &tx.WalletAddress, &tx.TokenAmount,
			&tx.TokenSymbol, &tx.TokenAddress, &tx.QuoteSymbol,
			&tx.QuoteAmount, &tx.QuoteAddress, &tx.QuotePrice,
			&tx.USDPrice, &tx.USDAmount,
		)
		if err != nil {
			log.Printf("Failed to scan transaction: %v", err)
			continue
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// GetFilteredTokenData MySQL 获取过滤后的代币数据
func (m *MySQLSwapTransactionDB) GetFilteredTokenData(ctx context.Context, filters TokenDataFilter) ([]parser.TokenSwapFilterData, error) {
	dbFilters := SwapTransactionFilter{
		StartTime:    filters.StartTime,
		EndTime:      filters.EndTime,
		MinUSDAmount: filters.MinUSDAmount,
		Limit:        filters.Limit,
	}

	transactions, err := m.GetSwapTransactions(ctx, dbFilters)
	if err != nil {
		return nil, err
	}

	// 使用 SolanaBlockDataHandler 进行过滤
	handler := parser.ExportSolanaBlockDataHandler
	return handler.FilterTokenData(transactions), nil
}

// ClickHouse 实现

// InsertSwapTransactions ClickHouse 批量插入交换交易
func (c *ClickHouseSwapTransactionDB) InsertSwapTransactions(ctx context.Context, transactions []parser.SwapTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	// ClickHouse 使用批量插入优化
	err := c.insertToTokenTable(ctx, transactions)
	if err != nil {
		return fmt.Errorf("failed to insert to token table: %w", err)
	}

	err = c.insertToWalletTable(ctx, transactions)
	if err != nil {
		return fmt.Errorf("failed to insert to wallet table: %w", err)
	}

	return nil
}

// insertToTokenTable ClickHouse 插入到代币表
func (c *ClickHouseSwapTransactionDB) insertToTokenTable(ctx context.Context, transactions []parser.SwapTransaction) error {
	query := `
	INSERT INTO solana_swap_transactions_token (
		tx_hash, transaction_time, wallet_address, token_amount, token_symbol, 
		token_address, quote_symbol, quote_amount, quote_address, quote_price, 
		usd_price, usd_amount, trade_type, pool_address, block_height
	) VALUES
	`

	values := []string{}
	args := []interface{}{}

	for _, tx := range transactions {
		values = append(values, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		args = append(args,
			tx.TxHash, time.Unix(int64(tx.TransactionTime), 0), tx.WalletAddress,
			tx.TokenAmount, tx.TokenSymbol, tx.TokenAddress,
			tx.QuoteSymbol, tx.QuoteAmount, tx.QuoteAddress,
			tx.QuotePrice, tx.USDPrice, tx.USDAmount,
			tx.TradeType, tx.PoolAddress, tx.BlockHeight,
		)
	}

	query += strings.Join(values, ",")

	_, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert transactions: %w", err)
	}

	log.Printf("✅ 插入 %d 条记录到 solana_swap_transactions_token", len(transactions))
	return nil
}

// insertToWalletTable ClickHouse 插入到钱包表
func (c *ClickHouseSwapTransactionDB) insertToWalletTable(ctx context.Context, transactions []parser.SwapTransaction) error {
	query := `
	INSERT INTO solana_swap_transactions_wallet (
		tx_hash, transaction_time, wallet_address, token_amount, token_symbol, 
		token_address, quote_symbol, quote_amount, quote_address, quote_price, 
		usd_price, usd_amount, trade_type, pool_address, block_height
	) VALUES
	`

	values := []string{}
	args := []interface{}{}

	for _, tx := range transactions {
		values = append(values, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		args = append(args,
			tx.TxHash, time.Unix(int64(tx.TransactionTime), 0), tx.WalletAddress,
			tx.TokenAmount, tx.TokenSymbol, tx.TokenAddress,
			tx.QuoteSymbol, tx.QuoteAmount, tx.QuoteAddress,
			tx.QuotePrice, tx.USDPrice, tx.USDAmount,
			tx.TradeType, tx.PoolAddress, tx.BlockHeight,
		)
	}

	query += strings.Join(values, ",")

	_, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert transactions: %w", err)
	}

	log.Printf("✅ 插入 %d 条记录到 solana_swap_transactions_wallet", len(transactions))
	return nil
}

// GetSwapTransactions ClickHouse 查询交换交易
func (c *ClickHouseSwapTransactionDB) GetSwapTransactions(ctx context.Context, filters SwapTransactionFilter) ([]parser.SwapTransactionToken, error) {
	query := `
	SELECT tx_hash, trade_type, pool_address, block_height, 
		   toUnixTimestamp(transaction_time) as transaction_time,
		   wallet_address, token_amount, token_symbol, token_address,
		   quote_symbol, quote_amount, quote_address, quote_price,
		   usd_price, usd_amount
	FROM solana_swap_transactions_token
	WHERE 1=1
	`

	args := []interface{}{}

	if filters.StartTime != nil {
		query += " AND transaction_time >= ?"
		args = append(args, *filters.StartTime)
	}

	if filters.EndTime != nil {
		query += " AND transaction_time <= ?"
		args = append(args, *filters.EndTime)
	}

	if filters.WalletAddr != nil {
		query += " AND wallet_address = ?"
		args = append(args, *filters.WalletAddr)
	}

	if filters.TokenAddr != nil {
		query += " AND token_address = ?"
		args = append(args, *filters.TokenAddr)
	}

	if filters.MinUSDAmount != nil {
		query += " AND toFloat64(usd_amount) >= ?"
		args = append(args, *filters.MinUSDAmount)
	}

	query += " ORDER BY transaction_time DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	if filters.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filters.Offset)
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query swap transactions: %w", err)
	}
	defer rows.Close()

	var transactions []parser.SwapTransactionToken
	for rows.Next() {
		var tx parser.SwapTransactionToken
		err := rows.Scan(
			&tx.TxHash, &tx.TradeType, &tx.PoolAddress, &tx.BlockHeight,
			&tx.TransactionTime, &tx.WalletAddress, &tx.TokenAmount,
			&tx.TokenSymbol, &tx.TokenAddress, &tx.QuoteSymbol,
			&tx.QuoteAmount, &tx.QuoteAddress, &tx.QuotePrice,
			&tx.USDPrice, &tx.USDAmount,
		)
		if err != nil {
			log.Printf("Failed to scan transaction: %v", err)
			continue
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// GetFilteredTokenData ClickHouse 获取过滤后的代币数据
func (c *ClickHouseSwapTransactionDB) GetFilteredTokenData(ctx context.Context, filters TokenDataFilter) ([]parser.TokenSwapFilterData, error) {
	dbFilters := SwapTransactionFilter{
		StartTime:    filters.StartTime,
		EndTime:      filters.EndTime,
		MinUSDAmount: filters.MinUSDAmount,
		Limit:        filters.Limit,
	}

	transactions, err := c.GetSwapTransactions(ctx, dbFilters)
	if err != nil {
		return nil, err
	}

	// 使用 SolanaBlockDataHandler 进行过滤
	handler := parser.ExportSolanaBlockDataHandler
	return handler.FilterTokenData(transactions), nil
}

// 工具函数：对应 TS 版本的数据库查询方法

// GetDataByBlockHeightRange 基于区块高度范围获取交易数据 (对应 TS 版本)
func (m *MySQLSwapTransactionDB) GetDataByBlockHeightRange(ctx context.Context, startBlockHeight, endBlockHeight uint64) ([]parser.SwapTransactionToken, error) {
	query := `
	SELECT tx_hash, trade_type, pool_address, block_height, 
		   UNIX_TIMESTAMP(transaction_time) as transaction_time,
		   wallet_address, token_amount, token_symbol, token_address,
		   quote_symbol, quote_amount, quote_address, quote_price,
		   usd_price, usd_amount
	FROM solana_swap_transactions_token
	WHERE block_height >= ? AND block_height <= ?
	ORDER BY block_height ASC
	`

	rows, err := m.db.QueryContext(ctx, query, startBlockHeight, endBlockHeight)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions by block height: %w", err)
	}
	defer rows.Close()

	var transactions []parser.SwapTransactionToken
	for rows.Next() {
		var tx parser.SwapTransactionToken
		err := rows.Scan(
			&tx.TxHash, &tx.TradeType, &tx.PoolAddress, &tx.BlockHeight,
			&tx.TransactionTime, &tx.WalletAddress, &tx.TokenAmount,
			&tx.TokenSymbol, &tx.TokenAddress, &tx.QuoteSymbol,
			&tx.QuoteAmount, &tx.QuoteAddress, &tx.QuotePrice,
			&tx.USDPrice, &tx.USDAmount,
		)
		if err != nil {
			log.Printf("Failed to scan transaction: %v", err)
			continue
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// GetDataByBlockHeightRange ClickHouse 版本
func (c *ClickHouseSwapTransactionDB) GetDataByBlockHeightRange(ctx context.Context, startBlockHeight, endBlockHeight uint64) ([]parser.SwapTransactionToken, error) {
	query := `
	SELECT tx_hash, trade_type, pool_address, block_height, 
		   toUnixTimestamp(transaction_time) as transaction_time,
		   wallet_address, token_amount, token_symbol, token_address,
		   quote_symbol, quote_amount, quote_address, quote_price,
		   usd_price, usd_amount
	FROM solana_swap_transactions_token
	WHERE block_height >= ? AND block_height <= ?
	ORDER BY block_height ASC
	`

	rows, err := c.db.QueryContext(ctx, query, startBlockHeight, endBlockHeight)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions by block height: %w", err)
	}
	defer rows.Close()

	var transactions []parser.SwapTransactionToken
	for rows.Next() {
		var tx parser.SwapTransactionToken
		err := rows.Scan(
			&tx.TxHash, &tx.TradeType, &tx.PoolAddress, &tx.BlockHeight,
			&tx.TransactionTime, &tx.WalletAddress, &tx.TokenAmount,
			&tx.TokenSymbol, &tx.TokenAddress, &tx.QuoteSymbol,
			&tx.QuoteAmount, &tx.QuoteAddress, &tx.QuotePrice,
			&tx.USDPrice, &tx.USDAmount,
		)
		if err != nil {
			log.Printf("Failed to scan transaction: %v", err)
			continue
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}
