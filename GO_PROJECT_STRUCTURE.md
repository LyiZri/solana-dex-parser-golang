# Go 项目结构重构 - 符合 Go 语言规范

## 重构目标

按照 Go 语言的最佳实践和规范，重新组织项目结构，将业务逻辑、数据模型、服务和数据访问层进行清晰分离。

## 新的项目结构

```
src/
├── config/                    # 配置管理
│   ├── config.go             # 配置结构体和加载逻辑
│   ├── constants.go          # 常量定义
│   └── dex_programs.go       # DEX 程序配置
│
├── db/                       # 数据访问层 (Data Access Layer)
│   ├── clickhouse/          # ClickHouse 数据库操作
│   │   └── solana_history_data.go  # Solana 历史数据查询方法
│   ├── dblogger/            # 数据库日志
│   ├── db.go                # 数据库连接管理
│   └── swap_transaction_db.go
│
├── model/                    # 数据模型 (Data Models)
│   └── user_report.go       # 用户报告相关的数据结构
│
├── service/                  # 业务服务层 (Business Service Layer)
│   ├── user_report_calculator.go  # 用户报告计算服务
│   └── price_service.go     # 价格服务
│
├── processor/                # 处理器层 (Processor Layer)
│   └── user_report_processor.go   # 用户报告处理器
│
├── parser/                   # 解析器 (Parsers)
│   ├── base_parser.go
│   ├── dex_parser.go
│   ├── instruction_classifier.go
│   ├── solana_block_data_handler.go
│   ├── solana_block_processor.go
│   ├── transaction_adapter.go
│   └── transaction_utils.go
│
├── rpc_call/                 # RPC 调用
│   └── rpc_call.go
│
├── solana/                   # Solana 相关功能
│   ├── batch_rpc.go
│   ├── concurrency_config.go
│   └── getBlock.go
│
├── util/                     # 工具类
│   └── rpc_call.go
│
├── seelog/                   # 日志组件
│   ├── app.go
│   ├── const.go
│   ├── seelog.go
│   └── util.go
│
├── examples/                 # 示例代码
│   └── user_report_example.go
│
└── main.go                   # 主程序入口
```

## 分层架构说明

### 1. 数据访问层 (Data Access Layer) - `src/db/`
**职责**：负责所有数据库操作和数据持久化
- `clickhouse/` - ClickHouse 数据库的所有查询方法
- `db.go` - 数据库连接管理
- 不包含业务逻辑，只负责数据的 CRUD 操作

### 2. 数据模型层 (Data Models) - `src/model/`
**职责**：定义所有数据结构和模型
- `UserReport` - 用户报告数据结构
- `TokenPnLData` - 代币盈亏数据结构
- `UserReportSummary` - 用户报告汇总结构
- `TransactionPriceData` - 交易价格数据结构

### 3. 业务服务层 (Business Service Layer) - `src/service/`
**职责**：实现核心业务逻辑和计算
- `UserReportCalculator` - 用户报告计算服务
- `PriceService` - 价格计算和缓存服务
- 包含算法实现和业务规则

### 4. 处理器层 (Processor Layer) - `src/processor/`
**职责**：协调不同服务，管理工作流程
- `UserReportProcessor` - 协调计算服务和数据存储
- 处理批量操作和错误恢复
- 管理数据库事务

### 5. 示例层 (Examples) - `src/examples/`
**职责**：提供使用示例和演示代码
- 展示如何使用各个组件
- 提供测试和验证代码

## 包的依赖关系

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  examples   │───▶│ processor   │───▶│   service   │
└─────────────┘    └─────────────┘    └─────────────┘
                           │                   │
                           ▼                   ▼
                   ┌─────────────┐    ┌─────────────┐
                   │    model    │    │     db      │
                   └─────────────┘    └─────────────┘
```

**依赖规则**：
- `examples` → `processor` → `service` → `model`
- `service` → `db` (数据访问)
- `processor` → `model` (数据模型)
- **禁止反向依赖**：低层不能依赖高层

## 符合 Go 语言规范的特点

### 1. 清晰的包职责
- 每个包都有明确的单一职责
- 包名简洁且具有描述性
- 避免了循环依赖

### 2. 分层架构
- 数据访问与业务逻辑分离
- 业务逻辑与表示层分离
- 模型定义独立

### 3. 可测试性
```go
// 可以单独测试每一层
func TestUserReportCalculator(t *testing.T) {
    // 可以 mock 数据访问层
}

func TestPriceService(t *testing.T) {
    // 可以 mock ClickHouse 连接
}
```

### 4. 可维护性
- 修改数据库查询只需更改 `db` 包
- 修改业务逻辑只需更改 `service` 包
- 添加新模型只需更改 `model` 包

## 使用方式

### 基本用法
```go
import (
    "github.com/go-solana-parse/src/db"
    "github.com/go-solana-parse/src/processor"
)

func main() {
    // 初始化数据库
    db.InitDB()
    db.InitClickHouseV2()
    
    // 创建处理器
    userProcessor := processor.NewUserReportProcessor(
        db.DBClient, 
        db.ClickHouseClient,
    )
    
    // 处理用户报告
    report, err := userProcessor.ProcessSingleUserReport("address")
}
```

### 扩展新功能
```go
// 1. 在 model 包中定义新的数据结构
// src/model/new_model.go

// 2. 在 db 包中添加数据访问方法
// src/db/clickhouse/new_data_access.go

// 3. 在 service 包中实现业务逻辑
// src/service/new_service.go

// 4. 在 processor 包中协调服务
// src/processor/new_processor.go
```

## 与原结构的对比

### 重构前的问题
```
src/db/user/
├── user_report_calculator.go  # 混合了业务逻辑和数据访问
├── user_report_processor.go   # 不符合 Go 包组织规范
├── price_service.go           # 业务逻辑放在了 db 包中
└── user_report.go            # 数据模型与数据访问混合
```

### 重构后的优势
```
src/
├── model/          # 纯数据模型
├── service/        # 纯业务逻辑
├── processor/      # 协调层
└── db/            # 纯数据访问
```

## 性能和扩展性

### 1. 更好的缓存策略
- 价格服务可以独立实现缓存
- 数据访问层可以添加连接池
- 业务逻辑层可以添加计算缓存

### 2. 更容易的水平扩展
- 各层可以独立部署
- 服务可以独立扩展
- 数据访问可以分库分表

### 3. 更好的监控和调试
- 每层都可以独立监控
- 错误定位更精确
- 性能瓶颈更容易识别

## 总结

此次重构完全符合 Go 语言的项目组织规范：

✅ **分层清晰**：数据访问、业务逻辑、协调层分离
✅ **职责单一**：每个包都有明确的职责
✅ **依赖合理**：遵循依赖倒置原则
✅ **易于测试**：每层都可以独立测试
✅ **便于维护**：修改影响范围可控
✅ **符合规范**：遵循 Go 语言最佳实践

新的结构为项目的长期发展和维护奠定了坚实的基础。 