# Solana DEX Parser - Go Implementation

A comprehensive Solana DEX (Decentralized Exchange) transaction parsing system written in Go, equivalent to the TypeScript version's `SolanaBlockDataHandler.handleBlockData()` functionality.

## 🌟 Features

### Multi-Protocol DEX Support (40+ Protocols)
- **Jupiter Aggregator Family**: Jupiter, Jupiter DCA, Limit Orders, VA, and multiple keeper programs
- **Raydium Series**: V4, Route, CPMM, Concentrated Liquidity, Launchpad
- **Orca Family**: Whirlpool, V1, V2
- **Meteora Products**: DLMM, Pools, DAMM
- **Meme Token Platforms**: Pump.fun, Moonshot, Boop.fun
- **Trading Bots**: Banana Gun, Mintech, Bloom, Maestro, Nova, Apepro
- **Major DEXs**: Phoenix, Openbook, Aldrin, Crema, GooseFX, Lifinity, Mercurial, Saber, and more

### Core Capabilities
- **Real-time Block Processing**: Parse Solana blocks and extract DEX transactions
- **Multi-layer Transaction Analysis**: Support for complex aggregated swaps and routing
- **Smart Filtering**: Blacklist filtering, MEV detection, minimum transaction thresholds
- **Data Standardization**: Convert raw blockchain data to standardized swap transactions
- **Database Integration**: PostgreSQL support with optimized schemas
- **Performance Optimized**: Concurrent processing with configurable batch sizes

## 🏗️ Architecture

### System Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Block Data     │───▶│  DexParser      │───▶│  Database       │
│  (Raw Solana)   │    │  (Core Engine)  │    │  (PostgreSQL)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                        │                        │
         ▼                        ▼                        ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ SolanaBlockData │    │ Transaction     │    │ SwapTransaction │
│ Handler         │    │ Classifier      │    │ DB              │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Key Components

#### 1. Configuration Layer (`src/config/`)
- **`dex_programs.go`**: 40+ DEX protocol definitions with program IDs
- **`constants.go`**: Token mappings, blacklists, trading types, and filtering rules

#### 2. Parser Layer (`src/parser/`)
- **`solana_block_data_handler.go`**: Main processing engine (equivalent to TS `handleBlockData`)
- **`dex_parser.go`**: Core DEX transaction parsing logic
- **`transaction_adapter.go`**: Transaction data extraction and balance tracking
- **`transaction_utils.go`**: Utility functions for DEX info and transfer processing
- **`instruction_classifier.go`**: Program ID extraction and instruction analysis
- **`types.go`**: Comprehensive type definitions for all data structures
- **`solana_block_processor.go`**: High-level block processing orchestrator

#### 3. Database Layer (`src/db/`)
- **`swap_transaction_db.go`**: Database interfaces and PostgreSQL implementation
- Optimized table schemas with proper indexing
- Batch insertion capabilities
- Advanced filtering and querying

## 🚀 Usage

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/go-solana-parse/src/parser"
    "github.com/go-solana-parse/src/db"
)

func main() {
    // 1. Create database connection
    database := initPostgresDB() // Your DB setup
    swapDB := db.NewPostgresSwapTransactionDB(database)
    
    // 2. Create block processor
    processor := parser.NewSolanaBlockProcessor(swapDB)
    
    // 3. Configure processing
    config := parser.DefaultProcessorConfig()
    config.EnableDatabase = true
    config.EnableFilter = true
    
    // 4. Process a block
    ctx := context.Background()
    blockData := getSolanaBlock() // Your block data source
    
    result, err := processor.ProcessBlock(ctx, blockData, blockNumber, config)
    if err != nil {
        log.Fatalf("Processing failed: %v", err)
    }
    
    log.Printf("Processed %d swap transactions", result.ValidTransactions)
}
```

### Advanced Features

#### Batch Processing
```go
// Process multiple blocks
blocks := []parser.BlockData{...}
batchResult, err := processor.ProcessBlocks(ctx, blocks, config)
```

#### Data Filtering and Analysis
```go
// Get filtered token data
filters := db.TokenDataFilter{
    StartTime:    &startTime,
    EndTime:      &endTime,
    MinUSDAmount: &minAmount,
    Limit:        100,
}

filteredData, err := processor.GetTokenData(ctx, filters)
```

#### Statistics and Monitoring
```go
// Get processing statistics
stats, err := processor.GetStatistics(ctx)
log.Printf("Total volume: $%.2f", stats.TotalVolumeUSD)
log.Printf("Unique wallets: %d", stats.UniqueWallets)
```

## 🔧 Configuration

### Database Schema

```sql
CREATE TABLE swap_transactions (
    id SERIAL PRIMARY KEY,
    tx_hash VARCHAR(100) NOT NULL,
    transaction_time TIMESTAMP NOT NULL,
    wallet_address VARCHAR(50) NOT NULL,
    token_amount NUMERIC NOT NULL,
    token_symbol VARCHAR(20),
    token_address VARCHAR(50) NOT NULL,
    quote_symbol VARCHAR(20),
    quote_amount NUMERIC NOT NULL,
    quote_address VARCHAR(50) NOT NULL,
    quote_price VARCHAR(50),
    usd_price VARCHAR(50),
    usd_amount VARCHAR(50),
    trade_type VARCHAR(10),
    pool_address VARCHAR(50),
    block_height BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tx_hash, wallet_address, token_address)
);
```

### Processing Configuration

```go
config := parser.ProcessorConfig{
    EnableDatabase:  true,
    EnableFilter:    true,
    BatchSize:       100,
    ProcessingDelay: time.Millisecond * 50,
}
```

## 📊 Data Flow

### 1. Block Data Input
```
Raw Solana Block → Transaction Extraction → Instruction Analysis
```

### 2. DEX Detection and Parsing
```
Program ID Matching → Protocol-Specific Parsing → Transaction Classification
```

### 3. Data Standardization
```
Raw Trade Data → Price Calculation → USD Conversion → Standardized Format
```

### 4. Filtering and Storage
```
Blacklist Filtering → MEV Detection → Database Storage → Query Interface
```

## 🎯 Supported DEX Protocols

### Jupiter Ecosystem (8 programs)
- Main Jupiter Aggregator
- DCA (Dollar Cost Averaging)
- Limit Orders V1 & V2
- Vote-Aggregated programs
- Multiple Keeper programs

### Raydium Family (5 programs)
- V4 AMM
- Route program
- CPMM (Constant Product Market Maker)
- Concentrated Liquidity
- Launchpad

### Orca Suite (3 programs)
- Whirlpool (main)
- V1 legacy
- V2 upgrade

### Meteora Products (3 programs)
- DLMM (Dynamic Liquidity Market Maker)
- Standard Pools
- DAMM (Dynamic AMM)

### Meme Token Platforms (4 programs)
- Pump.fun
- Pumpswap
- Moonshot
- Boop.fun

### Trading Bots (6 programs)
- Banana Gun
- Mintech
- Bloom
- Maestro
- Nova
- Apepro

### Major DEXs (12+ programs)
- Phoenix
- Openbook
- Aldrin
- Crema
- GooseFX
- Lifinity
- Mercurial
- Saber
- Saros
- SolFi
- Stabble
- Sanctum
- Photon
- OKX DEX

## 🔍 Key Features Comparison with TypeScript Version

| Feature | TypeScript | Go Implementation | Status |
|---------|------------|-------------------|---------|
| 40+ DEX Support | ✅ | ✅ | Complete |
| Jupiter Aggregator | ✅ | ✅ | Complete |
| Multi-layer Parsing | ✅ | ✅ | Complete |
| Price Calculation | ✅ | ✅ | Complete |
| Database Integration | ✅ | ✅ | Enhanced |
| Filtering System | ✅ | ✅ | Complete |
| Batch Processing | ⚠️ | ✅ | Improved |
| Concurrent Processing | ⚠️ | ✅ | New |
| Type Safety | ⚠️ | ✅ | Improved |
| Memory Management | ⚠️ | ✅ | Optimized |

## 🚦 Performance Optimizations

### Concurrent Processing
- Parallel transaction parsing
- Batch database operations
- Configurable processing delays

### Memory Management
- Efficient struct packing
- Minimal memory allocations
- Proper resource cleanup

### Database Optimizations
- Indexed queries
- Prepared statements
- Connection pooling
- Upsert operations for deduplication

## 📈 Monitoring and Analytics

### Processing Metrics
- Transactions per second
- Success/failure rates
- Processing latency
- Memory usage

### Business Metrics
- Trading volume (USD)
- Unique wallet count
- Token diversity
- DEX usage distribution

## 🛠️ Installation and Setup

### Prerequisites
```bash
# Go 1.19+
go version

# PostgreSQL 12+
psql --version
```

### Installation
```bash
# Clone the repository
git clone <repository-url>
cd go-solana-explore

# Install dependencies
go mod tidy

# Set up database
createdb solana_dex

# Run migrations (if available)
# go run migrations/migrate.go
```

### Environment Variables
```bash
export DATABASE_URL="postgres://user:pass@localhost/solana_dex?sslmode=disable"
export SOLANA_RPC_URL="https://api.mainnet-beta.solana.com"
```

## 🔮 Future Enhancements

### Planned Features
- [ ] Real-time streaming processing
- [ ] GraphQL API layer
- [ ] Machine learning price prediction
- [ ] Cross-chain DEX support
- [ ] Enhanced MEV detection
- [ ] WebSocket real-time feeds

### Performance Improvements
- [ ] Redis caching layer
- [ ] Horizontal scaling support
- [ ] Advanced query optimization
- [ ] Metrics and alerting

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Implement your changes
4. Add comprehensive tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Original TypeScript implementation team
- Solana development community
- DEX protocol developers
- Open source contributors

---

**Note**: This Go implementation provides equivalent functionality to the TypeScript `SolanaBlockDataHandler.handleBlockData()` system with enhanced performance, type safety, and additional features for enterprise-scale Solana DEX transaction processing. 