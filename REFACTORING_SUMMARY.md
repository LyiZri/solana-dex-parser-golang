# 代码重构总结 - 数据库操作分离

## 重构目标

按照你的要求，将所有直接与 ClickHouse 数据库进行操作的方法集中到 `solana_history_data.go` 文件中，实现数据访问层与业务逻辑层的分离。

## 重构前后对比

### 重构前的问题
- 数据库查询逻辑散布在多个文件中
- `UserReportCalculator` 中包含直接的 SQL 查询代码
- 职责不清晰，难以维护和测试

### 重构后的改进
- ✅ 所有 ClickHouse 数据库操作集中在 `solana_history_data.go`
- ✅ 业务逻辑与数据访问分离
- ✅ 更好的可测试性和可维护性
- ✅ 遵循单一职责原则

## 具体改动

### 1. `solana_history_data.go` 新增方法

增加了以下数据库操作方法，涵盖所有 ClickHouse 查询需求：

```go
// 基础查询方法
func (s *SolanaHistoryData) GetUserTransactionsByAddress(db ckdriver.Conn, address string) ([]*SolanaHistoryData, error)
func (s *SolanaHistoryData) GetUserTransactionsByAddressAndTimeRange(db ckdriver.Conn, address string, startTime, endTime uint64) ([]*SolanaHistoryData, error)
func (s *SolanaHistoryData) GetTokenTransactionsByAddressAndToken(db ckdriver.Conn, walletAddress, tokenAddress string) ([]*SolanaHistoryData, error)

// 优化查询方法
func (s *SolanaHistoryData) GetUniqueTokensByAddress(db ckdriver.Conn, address string) ([]string, error)
func (s *SolanaHistoryData) GetUserFirstTransaction(db ckdriver.Conn, address string) (*SolanaHistoryData, error)
func (s *SolanaHistoryData) GetTokenPriceAtBlock(db ckdriver.Conn, tokenAddress string, blockHeight uint64) (*SolanaHistoryData, error)
```

### 2. `user_report_calculator.go` 重构

**重构前：**
```go
func (calc *UserReportCalculator) getUserTransactions(address string) ([]*clickhouse.SolanaHistoryData, error) {
    query := `SELECT tx_hash, trade_type, ... FROM solana_history_data_new WHERE wallet_address = ? ORDER BY transaction_time ASC`
    rows, err := calc.clickhouseClient.Query(context.Background(), query, address)
    // ... 复杂的查询逻辑
}
```

**重构后：**
```go
func (calc *UserReportCalculator) getUserTransactions(address string) ([]*clickhouse.SolanaHistoryData, error) {
    solanaData := &clickhouse.SolanaHistoryData{}
    return solanaData.GetUserTransactionsByAddress(calc.clickhouseClient, address)
}
```

### 3. `price_service.go` 改进

**重构前：**
```go
func (ps *PriceService) GetTokenPriceAtBlock(tokenAddress string, blockHeight uint64) float64 {
    // TODO: 实现价格获取逻辑
    return 0.0
}
```

**重构后：**
```go
func (ps *PriceService) GetTokenPriceAtBlock(tokenAddress string, blockHeight uint64) float64 {
    solanaData := &clickhouse.SolanaHistoryData{}
    priceData, err := solanaData.GetTokenPriceAtBlock(ps.clickhouseClient, tokenAddress, blockHeight)
    // 根据查询结果计算价格
}
```

## 重构带来的好处

### 1. 职责分离
- **数据访问层** (`solana_history_data.go`): 专门负责 ClickHouse 数据库操作
- **业务逻辑层** (`user_report_calculator.go`): 专注于用户报告计算逻辑
- **服务层** (`price_service.go`): 处理价格相关的业务逻辑

### 2. 代码复用
- 所有数据库查询方法可以在不同模块间复用
- 避免重复的 SQL 查询代码
- 统一的错误处理和数据转换

### 3. 可维护性
- 数据库查询逻辑集中管理
- 修改数据库结构时只需更新一个文件
- 更容易进行性能优化

### 4. 可测试性
- 可以单独测试数据访问层
- 可以 mock 数据库操作进行业务逻辑测试
- 更好的单元测试覆盖率

### 5. 扩展性
- 新增查询方法只需在 `solana_history_data.go` 中添加
- 支持更复杂的查询条件和优化
- 便于添加缓存、连接池等功能

## 新增的查询功能

### 1. 时间范围查询
```go
// 支持按时间范围获取交易记录
transactions, err := solanaData.GetUserTransactionsByAddressAndTimeRange(db, address, startTime, endTime)
```

### 2. 特定代币查询
```go
// 获取用户对特定代币的所有交易
transactions, err := solanaData.GetTokenTransactionsByAddressAndToken(db, walletAddress, tokenAddress)
```

### 3. 优化查询
```go
// 直接获取用户的唯一代币列表，避免重复数据传输
tokens, err := solanaData.GetUniqueTokensByAddress(db, address)

// 直接获取第一笔交易，无需获取全部再排序
firstTx, err := solanaData.GetUserFirstTransaction(db, address)
```

### 4. 价格查询
```go
// 获取代币在特定区块高度的价格信息
priceData, err := solanaData.GetTokenPriceAtBlock(db, tokenAddress, blockHeight)
```

## 性能优化机会

重构后的架构为以下性能优化提供了基础：

1. **查询优化**: 可以在数据访问层添加查询优化逻辑
2. **批量查询**: 支持批量获取多个用户或代币的数据
3. **缓存机制**: 可以在数据访问层添加查询结果缓存
4. **索引优化**: 集中的查询逻辑便于数据库索引优化

## 后续扩展建议

1. **添加聚合查询方法**
   ```go
   func (s *SolanaHistoryData) GetUserTradeStatsByAddress(db ckdriver.Conn, address string) (*UserTradeStats, error)
   ```

2. **支持批量操作**
   ```go
   func (s *SolanaHistoryData) GetMultipleUsersTransactions(db ckdriver.Conn, addresses []string) (map[string][]*SolanaHistoryData, error)
   ```

3. **添加缓存层**
   ```go
   type CachedSolanaHistoryData struct {
       *SolanaHistoryData
       cache Cache
   }
   ```

4. **查询构建器**
   ```go
   func (s *SolanaHistoryData) QueryBuilder() *SolanaQueryBuilder
   ```

## 总结

此次重构成功实现了数据访问与业务逻辑的分离，提高了代码的可维护性、可测试性和可扩展性。所有的 ClickHouse 数据库操作现在都集中在 `solana_history_data.go` 文件中，符合你的架构要求。

重构后的代码结构更加清晰，为后续的功能扩展和性能优化奠定了良好的基础。 