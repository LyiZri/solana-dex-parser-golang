# 用户报告分析系统

## 概述

本系统是一个完整的用户交易数据分析解决方案，用于计算 Solana 区块链上用户的交易表现和投资组合指标。系统从 ClickHouse 中读取交易数据，进行复杂的 PnL 计算，并生成详细的用户报告。

## 功能特性

### 1. 基础交易指标
- **首次交易信息**：记录用户第一笔交易的时间、代币地址和数量
- **交易量统计**：计算总买入量、总卖出量和总交易量（SOL本位）
- **交易次数统计**：统计买入次数、卖出次数和总交易次数
- **代币种类统计**：计算用户交易的唯一代币数量

### 2. 盈亏分析（PnL）
- **盈亏统计**：计算盈利代币数量、亏损代币数量和胜率
- **最佳/最差代币**：识别盈利最多和亏损最大的代币地址
- **盈利分布**：统计不同盈利区间的代币数量
  - 0-200%、200-500%、500-1000%、>1000%
- **亏损分布**：统计不同亏损区间的代币数量
  - 0-50%、>50%

### 3. 投资组合分析
- **历史最高持仓**：记录历史最高持仓价值的代币信息
- **总体盈亏率**：计算用户整体的投资回报率
- **持仓价值追踪**：跟踪历史最高钱包总持仓价值

## 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   ClickHouse    │────│ UserReportCalc  │────│     MySQL       │
│ (交易数据源)     │    │   (计算引擎)     │    │  (报告存储)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                       ┌─────────────────┐
                       │   PriceService  │
                       │   (价格服务)     │
                       └─────────────────┘
```

### 核心组件

1. **UserReportCalculator** - 用户报告计算器
   - 负责计算各种交易指标
   - 实现 PnL 计算逻辑
   - 处理代币持仓分析

2. **PriceService** - 价格服务（占位符实现）
   - 提供代币价格查询接口
   - 支持历史价格获取
   - 缓存价格数据以提高性能

3. **UserReportProcessor** - 用户报告处理器
   - 协调整个处理流程
   - 管理数据库操作
   - 提供批量处理功能

## 使用方法

### 1. 初始化系统

```go
import (
    "github.com/go-solana-parse/src/db"
    userdb "github.com/go-solana-parse/src/db/user"
)

// 初始化数据库连接
err := db.InitDB()
if err != nil {
    log.Fatal("MySQL 初始化失败:", err)
}

err = db.InitClickHouseV2()
if err != nil {
    log.Fatal("ClickHouse 初始化失败:", err)
}
```

### 2. 处理单个用户报告

```go
// 创建处理器
processor := userdb.NewUserReportProcessor(db.DBClient, db.ClickHouseClient)

// 处理单个用户
userReport, err := processor.ProcessSingleUserReport("用户钱包地址")
if err != nil {
    log.Printf("处理失败: %v", err)
} else {
    log.Printf("用户报告: %+v", userReport)
}
```

### 3. 批量处理所有用户

```go
// 处理所有用户（适合后台任务）
err := processor.ProcessAllUserReports()
if err != nil {
    log.Printf("批量处理失败: %v", err)
}
```

### 4. 获取汇总信息

```go
// 获取用户报告汇总
summary, err := processor.GetUserReportSummary()
if err != nil {
    log.Printf("获取汇总失败: %v", err)
} else {
    log.Printf("总用户数: %d", summary.TotalUsers)
    log.Printf("盈利用户数: %d", summary.ProfitableUsers)
    log.Printf("盈利比例: %.2f%%", summary.ProfitableRatio*100)
}
```

## 数据结构

### UserReport - 用户报告
```go
type UserReport struct {
    Address                   string // 用户地址
    ProfitRate               string // 盈亏率
    FirstTradeTime           string // 首次交易时间
    FirstTradeTokenAddress   string // 首次交易代币地址
    FirstTradeTokenAmount    string // 首次交易代币数量
    TotalBuyVolumn          string // 总买入量(SOL)
    TotalSellVolumn         string // 总卖出量(SOL)
    TotalVolumn             string // 总交易量(SOL)
    BuyCount                int64  // 买入次数
    SellCount               int64  // 卖出次数
    TradeCount              int64  // 总交易次数
    UniqueTokenCount        int64  // 唯一代币数量
    PnlWinCount             int64  // 盈利代币数量
    PnlLossCount            int64  // 亏损代币数量
    PnlWinRate              string // 胜率
    TopProfitTokenAddress   string // 最佳盈利代币
    TopLossTokenAddress     string // 最大亏损代币
    WinLevelOneCount        int64  // 盈利分布 0-200%
    WinLevelTwoCount        int64  // 盈利分布 200-500%
    WinLevelThreeCount      int64  // 盈利分布 500-1000%
    WinLevelFourCount       int64  // 盈利分布 >1000%
    LossLevelOneCount       int64  // 亏损分布 0-50%
    LossLevelTwoCount       int64  // 亏损分布 >50%
    MostHoldValueTokenAddress string // 历史最高持仓代币地址
    MostHoldValueTokenAmount  string // 历史最高持仓数量
    MostHoldValueTokenUSD     string // 历史最高持仓USD价值
    MaxTotalHoldValue         string // 历史最高总持仓价值
}
```

### TokenPnLData - 代币盈亏数据
```go
type TokenPnLData struct {
    TokenAddress     string  // 代币地址
    TotalBuyAmount   float64 // 总买入数量
    TotalSellAmount  float64 // 总卖出数量
    TotalBuyValue    float64 // 总买入价值（SOL本位）
    TotalSellValue   float64 // 总卖出价值（SOL本位）
    AvgBuyPrice      float64 // 平均买入价格
    AvgSellPrice     float64 // 平均卖出价格
    RealizedPnL      float64 // 已实现盈亏
    UnrealizedPnL    float64 // 未实现盈亏
    CurrentHolding   float64 // 当前持仓
    MaxHoldingValue  float64 // 历史最高持仓价值
    MaxHoldingAmount float64 // 历史最高持仓数量
    MaxHoldingUSD    float64 // 历史最高持仓USD价值
}
```

## 计算逻辑

### 1. 已实现盈亏计算
```
已实现盈亏 = 用户卖出个数 * 用户卖出价格 - 用户卖出个数 * 用户平均买入价格
```

### 2. 未实现盈亏计算
```
未实现盈亏 = Math.min(用户当前持仓, 用户未卖出个数) * (当前价格 - 用户平均买入价格)
```

### 3. 平均买入价格计算
```
用户平均买入价格 = (pre买入总花费 + 当次总花费) / (pre买入总数量 + 当次购买总数量)
```

### 4. 胜率计算
```
胜率 = 盈利代币数量 / 交易代币总数量
```

## 性能优化

1. **批量处理**：支持分批处理大量用户数据
2. **价格缓存**：缓存历史价格数据避免重复查询
3. **增量更新**：支持更新已存在的用户报告
4. **错误恢复**：单个用户处理失败不影响整体流程

## TODO 和扩展计划

### 当前版本限制
- 价格服务仅为占位符实现，需要完整的价格获取逻辑
- 未实现时间维度分析（1d、3d、7d、1M、3M）
- 缺少转账记录的处理逻辑

### 未来扩展
1. **完整价格服务实现**
   - 实现基于交易数据的价格推导
   - 支持 SOL-USDC 价格获取
   - 实现代币价格缓存机制

2. **时间维度分析**
   - 支持不同时间窗口的 PnL 计算
   - 实现时间序列数据分析

3. **高级功能**
   - 支持转账记录分析
   - 实现更复杂的投资组合指标
   - 添加风险分析功能

## 文件结构

```
src/db/user/
├── user_report.go              # 用户报告数据结构
├── user_report_calculator.go   # 用户报告计算器
├── user_report_processor.go    # 用户报告处理器
└── price_service.go            # 价格服务（占位符）

src/examples/
└── user_report_example.go      # 使用示例
```

## 依赖项

- `gorm.io/gorm` - MySQL ORM
- `github.com/ClickHouse/clickhouse-go/v2` - ClickHouse 客户端
- Go 1.23+ - 语言版本要求

## 注意事项

1. **数据一致性**：确保 ClickHouse 和 MySQL 数据的一致性
2. **性能考虑**：大量用户处理时建议分批执行
3. **错误处理**：单个用户处理失败不会影响整体流程
4. **价格精度**：使用 float64 进行价格计算，注意精度问题

这个系统为 Solana 用户交易分析提供了一个完整的解决方案，可以根据具体需求进行扩展和定制。 