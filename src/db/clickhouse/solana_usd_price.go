package clickhouse

import (
	"context"
	"fmt"

	ckdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type SolanaUsdPrice struct {
	BlockHeight uint64  `ch:"block_height"`
	UsdPrice    float64 `ch:"usd_price"`
}

// TableName 返回表名
func (s *SolanaUsdPrice) TableName() string {
	return "solana_usd_price"
}

// GetSolanaUsdPrice 获取指定区块高度的SOL价格
func (s *SolanaUsdPrice) GetSolanaUsdPrice(db ckdriver.Conn, blockHeight uint64) (float64, error) {
	query := `SELECT usd_price FROM ` + s.TableName() + ` WHERE block_height = ?`
	row := db.QueryRow(context.Background(), query, blockHeight)

	var usdPrice float64
	err := row.Scan(&usdPrice)
	if err != nil {
		return 0, fmt.Errorf("未找到区块高度 %d 的SOL价格: %v", blockHeight, err)
	}

	return usdPrice, nil
}

// GetSolanaUsdPriceAtOrBefore 获取指定区块高度或之前最近的SOL价格
func (s *SolanaUsdPrice) GetSolanaUsdPriceAtOrBefore(db ckdriver.Conn, blockHeight uint64) (uint64, float64, error) {
	query := `
		SELECT block_height, usd_price 
		FROM ` + s.TableName() + `
		WHERE block_height <= ? 
		ORDER BY block_height DESC 
		LIMIT 1
	`
	row := db.QueryRow(context.Background(), query, blockHeight)

	var foundBlockHeight uint64
	var usdPrice float64
	err := row.Scan(&foundBlockHeight, &usdPrice)
	if err != nil {
		return 0, 0, fmt.Errorf("未找到区块高度 %d 或之前的SOL价格: %v", blockHeight, err)
	}

	return foundBlockHeight, usdPrice, nil
}

// InsertSolanaUsdPrice 插入SOL价格数据
func (s *SolanaUsdPrice) InsertSolanaUsdPrice(db ckdriver.Conn, blockHeight uint64, usdPrice float64) error {
	query := `INSERT INTO ` + s.TableName() + ` (block_height, usd_price) VALUES (?, ?)`
	err := db.Exec(context.Background(), query, blockHeight, usdPrice)
	if err != nil {
		return fmt.Errorf("插入SOL价格失败: %v", err)
	}
	return nil
}

// BatchInsertSolanaUsdPrice 批量插入SOL价格数据
func (s *SolanaUsdPrice) BatchInsertSolanaUsdPrice(db ckdriver.Conn, prices []SolanaUsdPrice) error {
	if len(prices) == 0 {
		return nil
	}

	batch, err := db.PrepareBatch(context.Background(),
		"INSERT INTO "+s.TableName()+" (block_height, usd_price)")
	if err != nil {
		return fmt.Errorf("准备批量插入失败: %v", err)
	}

	for _, price := range prices {
		err := batch.Append(price.BlockHeight, price.UsdPrice)
		if err != nil {
			return fmt.Errorf("添加批量数据失败: %v", err)
		}
	}

	err = batch.Send()
	if err != nil {
		return fmt.Errorf("发送批量数据失败: %v", err)
	}

	return nil
}

// ExistsSolanaUsdPrice 检查指定区块高度的价格是否已存在
func (s *SolanaUsdPrice) ExistsSolanaUsdPrice(db ckdriver.Conn, blockHeight uint64) (bool, error) {
	query := `SELECT COUNT(*) FROM ` + s.TableName() + ` WHERE block_height = ?`
	row := db.QueryRow(context.Background(), query, blockHeight)

	var count uint64
	err := row.Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查价格存在性失败: %v", err)
	}

	return count > 0, nil
}

// GetPriceRangeCount 获取指定区块高度范围内的价格记录数量
func (s *SolanaUsdPrice) GetPriceRangeCount(db ckdriver.Conn, startBlock, endBlock uint64) (uint64, error) {
	query := `SELECT COUNT(*) FROM ` + s.TableName() + ` WHERE block_height BETWEEN ? AND ?`
	row := db.QueryRow(context.Background(), query, startBlock, endBlock)

	var count uint64
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("获取价格范围计数失败: %v", err)
	}

	return count, nil
}

// GetLatestSolanaUsdPrice 获取最新的SOL价格记录
func (s *SolanaUsdPrice) GetLatestSolanaUsdPrice(db ckdriver.Conn) (uint64, float64, error) {
	query := `
		SELECT block_height, usd_price 
		FROM ` + s.TableName() + `
		ORDER BY block_height DESC 
		LIMIT 1
	`
	row := db.QueryRow(context.Background(), query)

	var blockHeight uint64
	var usdPrice float64
	err := row.Scan(&blockHeight, &usdPrice)
	if err != nil {
		return 0, 0, fmt.Errorf("获取最新SOL价格失败: %v", err)
	}

	return blockHeight, usdPrice, nil
}
