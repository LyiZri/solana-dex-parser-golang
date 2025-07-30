# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

### Building the Project
```bash
go mod tidy              # Install/update dependencies
go build ./src/main.go   # Build the main executable
go run ./src/main.go     # Run directly without building
```

### Testing
```bash
go test ./src/...        # Run all tests
go test ./src/db/clickhouse -v  # Run specific package tests with verbose output
```

### Multi-Process Execution Scripts
```bash
./test_multiprocess.sh                # Small-scale test (9000 blocks)
./run_continuous_multiprocess.sh      # Large-scale production run
./stop_all_processes.sh              # Stop all running processes
```

### Configuration
- Edit `config-yaml/config.yaml` for database and RPC settings
- Load configuration with `config.LoadSvcConfig()`

## Architecture Overview

This is a Solana DEX (Decentralized Exchange) transaction parsing system that processes blockchain data and extracts standardized swap transactions. The system supports 40+ DEX protocols including Jupiter, Raydium, Orca, Meteora, and major trading bots.

### Core Components

#### 1. Configuration Layer (`src/config/`)
- `dex_programs.go`: DEX protocol definitions with program IDs for 40+ protocols
- `constants.go`: Token mappings, blacklists, trading types, filtering rules
- `config.go`: Configuration loading from YAML files

#### 2. Parser Layer (`src/parser/`)
- `solana_block_data_handler.go`: Main processing engine (equivalent to TypeScript `handleBlockData`)
- `dex_parser.go`: Core DEX transaction parsing logic
- `transaction_adapter.go`: Transaction data extraction and balance tracking
- `transaction_utils.go`: Utility functions for DEX info and transfer processing
- `instruction_classifier.go`: Program ID extraction and instruction analysis
- `solana_block_processor.go`: High-level block processing orchestrator

#### 3. Database Layer (`src/db/`)
- `db.go`: Database connection management
- `init_clickhouse.go`: ClickHouse initialization
- `swap_transaction_db.go`: Database interfaces and implementations
- `clickhouse/solana_history_data.go`: ClickHouse-specific operations
- `user/user_report.go`: User reporting functionality

#### 4. RPC and External Communication (`src/rpc_call/`, `src/solana/`)
- `rpc_call.go`: External API communication for parsing
- `batch_rpc.go`: Batch RPC operations for efficiency
- `getBlock.go`: Solana block data retrieval
- `concurrency_config.go`: Concurrent processing configuration

#### 5. Data Models (`src/model/`)
- `parser.go`: Core parsing data structures
- `transaction.go`: Transaction-related types
- `rpc_call.go`: RPC communication structures
- `config.go`: Configuration data models

### Multi-Process Architecture

The system uses a sophisticated multi-process architecture for high-throughput block processing:

- **Reverse Processing**: Processes blocks from newest to oldest for optimal data freshness
- **Concurrent Workers**: Multiple Go processes handle different block ranges simultaneously
- **Batch Processing**: Groups blocks into configurable batch sizes for efficient RPC calls
- **Failure Tracking**: Records failed block processing attempts with timestamped logs
- **Memory Management**: Built-in garbage collection and memory monitoring

### Database Integration

**Primary Database**: ClickHouse for high-performance analytics
- Connection configured in `config.yaml`
- Optimized for time-series blockchain data
- Supports complex aggregation queries

**Secondary Database**: MySQL for operational data
- User reports and metadata
- Transaction deduplication

### Key Processing Flow

1. **Block Retrieval**: Fetch Solana block data via RPC
2. **Transaction Filtering**: Filter for token-related transactions (`TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA`)
3. **DEX Detection**: Match program IDs against 40+ supported DEX protocols
4. **Data Extraction**: Parse transaction details, amounts, tokens, and prices
5. **Standardization**: Convert to unified swap transaction format
6. **Storage**: Insert into ClickHouse for analytics

### Performance Characteristics

- **Multi-Core Optimization**: Leverages `runtime.GOMAXPROCS()` for CPU utilization
- **Concurrent Processing**: Configurable worker pools for parallel execution
- **Batch Operations**: Reduces network overhead with batched RPC calls
- **Memory Efficiency**: Periodic garbage collection and memory monitoring
- **Failure Recovery**: Automatic retry logic and failure logging

### Development Patterns

- **Error Handling**: Comprehensive error tracking with context
- **Logging**: Structured logging with progress indicators and timing
- **Testing**: Isolated test environment with `src/test/init.go`
- **Configuration**: YAML-based configuration with environment flexibility
- **Modularity**: Clear separation between parsing, storage, and communication layers

### External Dependencies

- **Solana RPC**: Helius mainnet endpoint for block data
- **ClickHouse**: Primary analytics database
- **MySQL**: Secondary operational database  
- **External Parser**: Sends data to localhost:8080 for additional processing