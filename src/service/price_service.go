package service

import (
	"fmt"
	"log"
	"sort"
	"sync"

	ckdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/go-solana-parse/src/config"
	"github.com/go-solana-parse/src/db/clickhouse"
)

// PriceCache 内存缓存结构（用于快速访问最近使用的价格）
type PriceCache struct {
	// 使用切片存储，按区块高度排序，便于二分查找
	BlockHeights []uint64  // 区块高度列表，已排序
	Prices       []float64 // 对应的SOL价格列表
	maxSize      int       // 最大缓存大小
	mutex        sync.RWMutex
}

// PriceService 代币价格服务（混合策略：持久化 + 内存缓存）
type PriceService struct {
	clickhouseClient ckdriver.Conn                 // 数据库连接
	solPriceCache    *PriceCache                   // 内存缓存
	solPriceDB       *clickhouse.SolanaUsdPrice    // 持久化存储（solana_usd_price表）
	historyDataDB    *clickhouse.SolanaHistoryData // 历史数据查询（solana_history_data_new表）
}

// NewPriceService 创建新的价格服务
func NewPriceService(clickhouseClient ckdriver.Conn) *PriceService {
	if clickhouseClient == nil {
		panic("ClickHouse client cannot be nil")
	}
	return &PriceService{
		clickhouseClient: clickhouseClient,
		solPriceCache: &PriceCache{
			BlockHeights: make([]uint64, 0),
			Prices:       make([]float64, 0),
			maxSize:      10000000, // 最多缓存10000000个价格点
		},
		solPriceDB:    &clickhouse.SolanaUsdPrice{},
		historyDataDB: &clickhouse.SolanaHistoryData{},
	}
}

// GetSOLPriceAtBlock 获取SOL在指定区块高度的价格（混合策略）
func (ps *PriceService) GetSOLPriceAtBlock(blockHeight uint64) (float64, error) {
	// 1. 先从内存缓存查找
	if price, found := ps.solPriceCache.getPrice(blockHeight); found {
		return price, nil
	}

	// 2. 内存缓存未命中，从持久化存储查找
	foundBlockHeight, price, err := ps.solPriceDB.GetSolanaUsdPriceAtOrBefore(ps.clickhouseClient, blockHeight)
	if err == nil {
		// 找到了持久化的价格，加入内存缓存
		ps.solPriceCache.setPrice(foundBlockHeight, price)
		return price, nil
	}
	log.Printf("开始从持久化存储获取交易数据")

	// 3. 持久化存储也没有，从原始交易数据计算
	calculatedPrice, err := ps.historyDataDB.GetSOLPriceFromTransactions(ps.clickhouseClient, blockHeight)
	if err != nil {
		return 0.0, fmt.Errorf("计算SOL价格失败: %v", err)
	}

	// 4. 将计算出的价格存储到持久化和缓存
	err = ps.solPriceDB.InsertSolanaUsdPrice(ps.clickhouseClient, blockHeight, calculatedPrice)
	if err != nil {
		// 持久化失败不影响返回结果，但记录日志
	}

	ps.solPriceCache.setPrice(blockHeight, calculatedPrice)

	return calculatedPrice, nil
}

// GetTokenPriceAtBlock 获取某个代币在指定区块高度的USD价格
func (ps *PriceService) GetTokenPriceAtBlock(tokenAddress string, blockHeight uint64) (float64, error) {
	// 如果是USDC，价格固定为1
	usdcAddress := config.USDC_ADDRESS
	if tokenAddress == usdcAddress {
		return 1.0, nil
	}

	// 查询该代币在指定区块高度的交易数据
	priceData, err := ps.historyDataDB.GetTokenPriceAtBlock(ps.clickhouseClient, tokenAddress, blockHeight)
	if err != nil {
		return 0.0, fmt.Errorf("查询代币交易数据失败: %v", err)
	}

	// 如果交易对手是USDC，直接计算USD价格
	if priceData.QuoteAddress == usdcAddress {
		if priceData.TokenAmount > 0 {
			return priceData.QuoteAmount / priceData.TokenAmount, nil
		}
	}

	// 如果交易对手是SOL，需要获取SOL价格再折算
	solAddress := config.SOL_ADDRESS
	wsolAddress := config.WSOL_ADDRESS
	if priceData.QuoteAddress == solAddress || priceData.QuoteAddress == wsolAddress {

		solPrice, err := ps.GetSOLPriceAtBlock(blockHeight)
		if err != nil {
			return 0.0, fmt.Errorf("获取SOL价格失败: %v", err)
		}

		if priceData.TokenAmount > 0 {
			tokenPriceInSOL := priceData.QuoteAmount / priceData.TokenAmount
			return tokenPriceInSOL * solPrice, nil
		}
	}

	return 0.0, fmt.Errorf("无法计算代币价格")
}

// BatchCalculateAndStorePrices 批量计算并存储SOL价格（用于历史数据预处理）
func (ps *PriceService) BatchCalculateAndStorePrices(startBlock, endBlock uint64) error {
	fmt.Printf("开始批量计算SOL价格，区块范围: %d - %d\n", startBlock, endBlock)

	// 获取该范围内所有SOL-USDC交易
	transactions, err := ps.historyDataDB.GetSOLTransactionsInRange(ps.clickhouseClient, startBlock, endBlock)
	if err != nil {
		return fmt.Errorf("获取SOL交易数据失败: %v", err)
	}

	var prices []clickhouse.SolanaUsdPrice
	for _, tx := range transactions {
		// 检查是否已存在该区块的价格
		exists, err := ps.solPriceDB.ExistsSolanaUsdPrice(ps.clickhouseClient, tx.BlockHeight)
		if err != nil || exists {
			continue // 跳过已存在的或检查失败的
		}

		// 计算价格
		price, err := tx.CalculateSOLPriceFromTransaction()
		if err != nil {
			continue // 跳过无效的交易数据
		}

		prices = append(prices, clickhouse.SolanaUsdPrice{
			BlockHeight: tx.BlockHeight,
			UsdPrice:    price,
		})
	}

	// 批量插入
	if len(prices) > 0 {
		err = ps.solPriceDB.BatchInsertSolanaUsdPrice(ps.clickhouseClient, prices)
		if err != nil {
			return fmt.Errorf("批量插入SOL价格失败: %v", err)
		}
		fmt.Printf("成功插入 %d 条SOL价格记录\n", len(prices))
	}

	return nil
}

// getPrice 从内存缓存获取价格（使用二分查找）
func (pc *PriceCache) getPrice(blockHeight uint64) (float64, bool) {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()

	if len(pc.BlockHeights) == 0 {
		return 0, false
	}

	// 使用二分查找找到小于等于指定区块高度的最大值
	idx := sort.Search(len(pc.BlockHeights), func(i int) bool {
		return pc.BlockHeights[i] > blockHeight
	})

	if idx == 0 {
		return 0, false
	}

	return pc.Prices[idx-1], true
}

// setPrice 设置价格到内存缓存
func (pc *PriceCache) setPrice(blockHeight uint64, price float64) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	// 检查是否已存在该区块高度
	idx := sort.Search(len(pc.BlockHeights), func(i int) bool {
		return pc.BlockHeights[i] >= blockHeight
	})

	if idx < len(pc.BlockHeights) && pc.BlockHeights[idx] == blockHeight {
		// 更新现有价格
		pc.Prices[idx] = price
		return
	}

	// 插入新的价格记录，保持排序
	pc.BlockHeights = append(pc.BlockHeights, 0)
	pc.Prices = append(pc.Prices, 0)

	// 向后移动元素
	copy(pc.BlockHeights[idx+1:], pc.BlockHeights[idx:])
	copy(pc.Prices[idx+1:], pc.Prices[idx:])

	// 插入新值
	pc.BlockHeights[idx] = blockHeight
	pc.Prices[idx] = price

	// 检查缓存大小，超出限制时移除最旧的记录
	if len(pc.BlockHeights) > pc.maxSize {
		pc.BlockHeights = pc.BlockHeights[1:]
		pc.Prices = pc.Prices[1:]
	}
}

// GetCacheStats 获取缓存统计信息
func (ps *PriceService) GetCacheStats() map[string]interface{} {
	ps.solPriceCache.mutex.RLock()
	defer ps.solPriceCache.mutex.RUnlock()

	var minBlock, maxBlock uint64
	if len(ps.solPriceCache.BlockHeights) > 0 {
		minBlock = ps.solPriceCache.BlockHeights[0]
		maxBlock = ps.solPriceCache.BlockHeights[len(ps.solPriceCache.BlockHeights)-1]
	}

	return map[string]interface{}{
		"cache_size":     len(ps.solPriceCache.BlockHeights),
		"max_cache_size": ps.solPriceCache.maxSize,
		"min_block":      minBlock,
		"max_block":      maxBlock,
	}
}
