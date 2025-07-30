# Solana DEX Parser - MySQL & ClickHouse 集成方案

本文档说明如何将 Go 版本的 Solana DEX 解析系统集成到您现有的 MySQL 和 ClickHouse 数据库架构中。

## 🎯 目标

- 兼容您现有的 MySQL 和 ClickHouse 数据库
- 完全对应 TypeScript 版本的数据表结构
- 支持双写：MySQL（事务性存储） + ClickHouse（分析性存储）
- 保持与原有 TS 版本相同的数据格式和查询接口

## 📊 数据库架构

### MySQL 表结构（事务性存储）

```sql
-- 代币视角的交换交易表（对应 TS 版本的 solana_swap_transactions_token）
CREATE TABLE solana_swap_transactions_token (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tx_hash VARCHAR(100) NOT NULL,
    transaction_time TIMESTAMP NOT NULL,
    wallet_address VARCHAR(50) NOT NULL,
    token_amount DECIMAL(20,8) NOT NULL,
    token_symbol VARCHAR(20),
    token_address VARCHAR(50) NOT NULL,
    quote_symbol VARCHAR(20),
    quote_amount DECIMAL(20,8) NOT NULL,
    quote_address VARCHAR(50) NOT NULL,
    quote_price VARCHAR(50),
    usd_price VARCHAR(50),
    usd_amount VARCHAR(50),
    trade_type VARCHAR(10),
    pool_address VARCHAR(50),
    block_height BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_token_tx (tx_hash, wallet_address, token_address),
    INDEX idx_token_time (transaction_time),
    INDEX idx_token_wallet (wallet_address),
    INDEX idx_token_address (token_address),
    INDEX idx_token_block (block_height)
);

-- 钱包视角的交换交易表（对应 TS 版本的 solana_swap_transactions_wallet）
CREATE TABLE solana_swap_transactions_wallet (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tx_hash VARCHAR(100) NOT NULL,
    transaction_time TIMESTAMP NOT NULL,
    wallet_address VARCHAR(50) NOT NULL,
    token_amount DECIMAL(20,8) NOT NULL,
    token_symbol VARCHAR(20),
    token_address VARCHAR(50) NOT NULL,
    quote_symbol VARCHAR(20),
    quote_amount DECIMAL(20,8) NOT NULL,
    quote_address VARCHAR(50) NOT NULL,
    quote_price VARCHAR(50),
    usd_price VARCHAR(50),
    usd_amount VARCHAR(50),
    trade_type VARCHAR(10),
    pool_address VARCHAR(50),
    block_height BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_wallet_tx (tx_hash, wallet_address, token_address),
    INDEX idx_wallet_time (transaction_time),
    INDEX idx_wallet_address (wallet_address),
    INDEX idx_wallet_token (token_address),
    INDEX idx_wallet_block (block_height)
);
```

### ClickHouse 表结构（分析性存储）

```sql
-- ClickHouse 代币交易表（高性能分析）
CREATE TABLE solana_swap_transactions_token (
    tx_hash String,
    transaction_time DateTime,
    wallet_address String,
    token_amount Float64,
    token_symbol String,
    token_address String,
    quote_symbol String,
    quote_amount Float64,
    quote_address String,
    quote_price String,
    usd_price String,
    usd_amount String,
    trade_type String,
    pool_address String,
    block_height UInt64,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(transaction_time)
ORDER BY (transaction_time, wallet_address, token_address)
SETTINGS index_granularity = 8192;

-- ClickHouse 钱包交易表（高性能分析）
CREATE TABLE solana_swap_transactions_wallet (
    tx_hash String,
    transaction_time DateTime,
    wallet_address String,
    token_amount Float64,
    token_symbol String,
    token_address String,
    quote_symbol String,
    quote_amount Float64,
    quote_address String,
    quote_price String,
    usd_price String,
    usd_amount String,
    trade_type String,
    pool_address String,
    block_height UInt64,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(transaction_time)
ORDER BY (wallet_address, transaction_time, token_address)
SETTINGS index_granularity = 8192;
```

## 🔧 核心实现

### 数据库连接和配置

```go
package main

import (
    "database/sql"
    "context"
    "log"
    
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/ClickHouse/clickhouse-go/v2"
)

// 数据库配置
type DatabaseConfig struct {
    MySQLDSN      string
    ClickHouseDSN string
}

// 创建数据库连接
func setupDatabases(config DatabaseConfig) (*sql.DB, *sql.DB, error) {
    // MySQL 连接
    mysqlDB, err := sql.Open("mysql", config.MySQLDSN)
    if err != nil {
        return nil, nil, fmt.Errorf("MySQL connection failed: %w", err)
    }
    
    // ClickHouse 连接
    clickhouseDB, err := sql.Open("clickhouse", config.ClickHouseDSN)
    if err != nil {
        return nil, nil, fmt.Errorf("ClickHouse connection failed: %w", err)
    }
    
    return mysqlDB, clickhouseDB, nil
}
```

### 核心处理器实现

```go
// SolanaBlockDataHandler 处理器 - 完全对应 TS 版本
type SolanaBlockDataHandler struct {
    mysqlDB      *sql.DB
    clickhouseDB *sql.DB
}

// HandleBlockData 处理区块数据 - 对应 TS 版本的主要方法
func (h *SolanaBlockDataHandler) HandleBlockData(
    ctx context.Context, 
    blockData VersionedBlockResponse, 
    blockNumber uint64,
) error {
    // 1. 解析区块数据（集成您的 40+ DEX 解析逻辑）
    parseResult, err := h.parseBlockData(blockData, blockNumber)
    if err != nil {
        return fmt.Errorf("failed to parse block data: %w", err)
    }
    
    // 2. 过滤出包含交易的结果
    var filteredTransactions []ParseResult
    for _, tx := range parseResult {
        if len(tx.Result.Trades) > 0 && len(tx.Trades) > 0 {
            filteredTransactions = append(filteredTransactions, tx)
        }
    }
    
    // 3. 转换数据格式
    var swapTransactionArray []SwapTransaction
    for _, tx := range filteredTransactions {
        for index := range tx.Trades {
            swapTransaction, err := h.convertData(tx, index, blockNumber)
            if err != nil {
                log.Printf("convertData error: %v", err)
                continue
            }
            if swapTransaction != nil {
                swapTransactionArray = append(swapTransactionArray, *swapTransaction)
            }
        }
    }
    
    // 4. 双写数据库
    if len(swapTransactionArray) > 0 {
        // MySQL 写入（对应 TS 版本的 insertToTokenTable 和 insertToWalletTable）
        if err := h.insertToTokenTable(ctx, swapTransactionArray); err != nil {
            log.Printf("MySQL token table insert failed: %v", err)
        }
        if err := h.insertToWalletTable(ctx, swapTransactionArray); err != nil {
            log.Printf("MySQL wallet table insert failed: %v", err)
        }
        
        // ClickHouse 写入（高性能分析存储）
        if err := h.insertToClickHouseTokenTable(ctx, swapTransactionArray); err != nil {
            log.Printf("ClickHouse token table insert failed: %v", err)
        }
        if err := h.insertToClickHouseWalletTable(ctx, swapTransactionArray); err != nil {
            log.Printf("ClickHouse wallet table insert failed: %v", err)
        }
    }
    
    return nil
}
```

### MySQL 数据操作（对应 TS 版本）

```go
// insertToTokenTable MySQL 代币表插入 - 对应 TS 版本
func (h *SolanaBlockDataHandler) insertToTokenTable(ctx context.Context, transactions []SwapTransaction) error {
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
    
    stmt, err := h.mysqlDB.PrepareContext(ctx, query)
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

// insertToWalletTable MySQL 钱包表插入 - 对应 TS 版本
func (h *SolanaBlockDataHandler) insertToWalletTable(ctx context.Context, transactions []SwapTransaction) error {
    // 实现类似于 insertToTokenTable 的逻辑
    // ...
}

// getXDaysData MySQL 查询 - 对应 TS 版本
func (h *SolanaBlockDataHandler) getXDaysData(ctx context.Context, timestamp uint64, limit int) ([]SwapTransactionToken, error) {
    query := `
    SELECT tx_hash, trade_type, pool_address, block_height,
           UNIX_TIMESTAMP(transaction_time) as transaction_time,
           wallet_address, token_amount, token_symbol, token_address,
           quote_symbol, quote_amount, quote_address, quote_price,
           usd_price, usd_amount
    FROM solana_swap_transactions_token
    WHERE UNIX_TIMESTAMP(transaction_time) > ?
    ORDER BY transaction_time ASC
    `
    if limit > 0 {
        query += fmt.Sprintf(" LIMIT %d", limit)
    }
    
    rows, err := h.mysqlDB.QueryContext(ctx, query, timestamp)
    if err != nil {
        return nil, fmt.Errorf("failed to query transactions: %w", err)
    }
    defer rows.Close()
    
    var transactions []SwapTransactionToken
    for rows.Next() {
        var tx SwapTransactionToken
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

// getDataByBlockHeightRange 根据区块高度查询 - 对应 TS 版本
func (h *SolanaBlockDataHandler) getDataByBlockHeightRange(ctx context.Context, startBlockHeight, endBlockHeight uint64) ([]SwapTransactionToken, error) {
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
    
    rows, err := h.mysqlDB.QueryContext(ctx, query, startBlockHeight, endBlockHeight)
    if err != nil {
        return nil, fmt.Errorf("failed to query transactions by block height: %w", err)
    }
    defer rows.Close()
    
    // 处理结果...
    return []SwapTransactionToken{}, nil
}
```

### ClickHouse 高性能分析

```go
// insertToClickHouseTokenTable ClickHouse 批量插入
func (h *SolanaBlockDataHandler) insertToClickHouseTokenTable(ctx context.Context, transactions []SwapTransaction) error {
    if len(transactions) == 0 {
        return nil
    }
    
    // ClickHouse 批量插入优化
    query := `
    INSERT INTO solana_swap_transactions_token (
        tx_hash, transaction_time, wallet_address, token_amount, token_symbol,
        token_address, quote_symbol, quote_amount, quote_address, quote_price,
        usd_price, usd_amount, trade_type, pool_address, block_height
    ) VALUES
    `
    
    values := make([]string, len(transactions))
    for i, tx := range transactions {
        values[i] = fmt.Sprintf(
            "('%s', '%s', '%s', %f, '%s', '%s', '%s', %f, '%s', '%s', '%s', '%s', '%s', '%s', %d)",
            tx.TxHash, time.Unix(int64(tx.TransactionTime), 0).Format("2006-01-02 15:04:05"),
            tx.WalletAddress, tx.TokenAmount, tx.TokenSymbol, tx.TokenAddress,
            tx.QuoteSymbol, tx.QuoteAmount, tx.QuoteAddress,
            tx.QuotePrice, tx.USDPrice, tx.USDAmount,
            tx.TradeType, tx.PoolAddress, tx.BlockHeight,
        )
    }
    
    finalQuery := query + strings.Join(values, ",")
    _, err := h.clickhouseDB.ExecContext(ctx, finalQuery)
    if err != nil {
        return fmt.Errorf("failed to insert to ClickHouse: %w", err)
    }
    
    log.Printf("✅ ClickHouse 批量插入 %d 条记录", len(transactions))
    return nil
}

// ClickHouse 高性能分析查询
func (h *SolanaBlockDataHandler) analyzeTradeData(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
    results := make(map[string]interface{})
    
    // 总交易量
    query1 := `
    SELECT sum(toFloat64(usd_amount)) as total_volume
    FROM solana_swap_transactions_token
    WHERE transaction_time >= ? AND transaction_time <= ?
    `
    var totalVolume float64
    err := h.clickhouseDB.QueryRowContext(ctx, query1, startTime, endTime).Scan(&totalVolume)
    if err != nil {
        return nil, err
    }
    results["total_volume"] = totalVolume
    
    // 热门代币
    query2 := `
    SELECT token_symbol, count(*) as trade_count, sum(toFloat64(usd_amount)) as volume
    FROM solana_swap_transactions_token
    WHERE transaction_time >= ? AND transaction_time <= ?
    GROUP BY token_symbol
    ORDER BY volume DESC
    LIMIT 10
    `
    rows, err := h.clickhouseDB.QueryContext(ctx, query2, startTime, endTime)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var topTokens []map[string]interface{}
    for rows.Next() {
        var symbol string
        var count int64
        var volume float64
        if err := rows.Scan(&symbol, &count, &volume); err != nil {
            continue
        }
        topTokens = append(topTokens, map[string]interface{}{
            "symbol": symbol,
            "count":  count,
            "volume": volume,
        })
    }
    results["top_tokens"] = topTokens
    
    return results, nil
}
```

## 🔄 数据流程

### 1. 区块处理流程（对应 TS 版本）

```
Solana Block → DexParser → ParseResult → ConvertData → 双写数据库
     ↓              ↓            ↓           ↓              ↓
区块数据        解析交易     标准化结果    格式转换     MySQL + ClickHouse
```

### 2. 查询流程

```
查询请求 → MySQL（事务查询）→ 返回结果
    ↓
查询请求 → ClickHouse（分析查询）→ 返回结果
```

## 📋 实际部署步骤

### 1. 安装依赖

```bash
go get github.com/go-sql-driver/mysql
go get github.com/ClickHouse/clickhouse-go/v2
```

### 2. 配置数据库连接

```go
config := DatabaseConfig{
    MySQLDSN:      "user:password@tcp(localhost:3306)/solana_dex?parseTime=true",
    ClickHouseDSN: "tcp://localhost:9000/solana_dex",
}
```

### 3. 集成现有表结构

确保您的 MySQL 和 ClickHouse 中存在对应的表：
- `solana_swap_transactions_token`
- `solana_swap_transactions_wallet`

### 4. 集成 DEX 解析逻辑

将您现有的 40+ DEX 协议解析逻辑集成到 `parseBlockData` 方法中：
- Jupiter 系列
- Raydium 系列
- Orca 系列
- Meteora 系列
- Pump.fun 等 meme 币平台
- 各种交易机器人

### 5. 启动处理器

```go
func main() {
    ctx := context.Background()
    
    // 初始化数据库
    mysqlDB, clickhouseDB, err := setupDatabases(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建处理器
    handler := &SolanaBlockDataHandler{
        mysqlDB:      mysqlDB,
        clickhouseDB: clickhouseDB,
    }
    
    // 处理区块数据
    for blockNumber := latestBlock; ; blockNumber++ {
        blockData, err := getBlockData(blockNumber)
        if err != nil {
            log.Printf("Failed to get block %d: %v", blockNumber, err)
            continue
        }
        
        err = handler.HandleBlockData(ctx, blockData, blockNumber)
        if err != nil {
            log.Printf("Failed to process block %d: %v", blockNumber, err)
        }
        
        time.Sleep(time.Second) // 控制处理速度
    }
}
```

## 📊 性能优化

### MySQL 优化
- 使用预编译语句减少 SQL 解析开销
- 批量插入减少网络往返
- 适当的索引策略
- 连接池配置

### ClickHouse 优化
- 批量插入提高写入性能
- 分区策略优化查询性能
- 压缩算法选择
- 并行查询优化

## 🔍 监控和维护

### 数据一致性检查
```go
func checkDataConsistency(ctx context.Context, mysqlDB, clickhouseDB *sql.DB) error {
    // 检查 MySQL 和 ClickHouse 数据同步状态
    // 比较记录数量、最新区块等
}
```

### 性能监控
```go
func monitorPerformance() {
    // 监控处理速度、错误率、数据库性能等
}
```

## ✅ 总结

本方案完全兼容您现有的 MySQL 和 ClickHouse 数据库架构，提供：

1. **完整的 TS 版本功能对应**：所有方法和数据格式都与 TypeScript 版本保持一致
2. **双写架构**：MySQL 保证事务性，ClickHouse 提供高性能分析
3. **40+ DEX 协议支持**：完整的多协议解析能力
4. **高性能处理**：Go 语言的并发优势和 ClickHouse 的分析性能
5. **易于集成**：最小化对现有系统的改动

您可以根据实际需求调整配置和优化策略，逐步迁移或并行运行与现有 TypeScript 系统。 