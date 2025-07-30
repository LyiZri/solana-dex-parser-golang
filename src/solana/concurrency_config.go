package solana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/go-solana-parse/src/config"
	"github.com/go-solana-parse/src/model"
	"golang.org/x/time/rate"
)

// ConcurrencyConfig holds configuration for optimized batch processing
type ConcurrencyConfig struct {
	// Rate limiting
	MaxRequestsPerSecond int // Maximum requests per second (based on your API limit)
	BurstCapacity        int // Burst capacity for rate limiter

	// Concurrency settings
	MaxConcurrentWorkers int // Maximum number of concurrent workers
	BatchSize            int // Number of blocks to process in one batch

	// Timeout settings
	RequestTimeout time.Duration // Individual request timeout
	BatchTimeout   time.Duration // Total batch timeout

	// Performance tuning
	HTTPPoolSize  int           // HTTP connection pool size
	RetryAttempts int           // Number of retry attempts for failed requests
	RetryDelay    time.Duration // Delay between retries
}

// DefaultConcurrencyConfig returns optimized default configuration
func DefaultConcurrencyConfig() *ConcurrencyConfig {
	return &ConcurrencyConfig{
		MaxRequestsPerSecond: 500,                   // Your Helius API limit
		BurstCapacity:        100,                   // Allow bursts up to 100 requests
		MaxConcurrentWorkers: runtime.NumCPU() * 20, // 20 workers per CPU core
		BatchSize:            100,                   // Process 100 blocks at a time
		RequestTimeout:       30 * time.Second,      // 30s per request
		BatchTimeout:         5 * time.Minute,       // 5 minutes for entire batch
		HTTPPoolSize:         200,                   // 200 HTTP connections
		RetryAttempts:        3,                     // Retry failed requests 3 times
		RetryDelay:           1 * time.Second,       // 1 second between retries
	}
}

// HighPerformanceConfig returns configuration optimized for maximum throughput
func HighPerformanceConfig() *ConcurrencyConfig {
	return &ConcurrencyConfig{
		MaxRequestsPerSecond: 500,                    // Max out your API limit
		BurstCapacity:        200,                    // Higher burst capacity
		MaxConcurrentWorkers: runtime.NumCPU() * 50,  // More aggressive concurrency
		BatchSize:            500,                    // Larger batches
		RequestTimeout:       15 * time.Second,       // Shorter timeouts for faster failure detection
		BatchTimeout:         10 * time.Minute,       // Longer batch timeout for larger batches
		HTTPPoolSize:         500,                    // Larger HTTP pool
		RetryAttempts:        2,                      // Fewer retries for speed
		RetryDelay:           500 * time.Millisecond, // Faster retries
	}
}

// ConservativeConfig returns configuration optimized for reliability
func ConservativeConfig() *ConcurrencyConfig {
	return &ConcurrencyConfig{
		MaxRequestsPerSecond: 300,                   // Leave some headroom
		BurstCapacity:        50,                    // Conservative burst
		MaxConcurrentWorkers: runtime.NumCPU() * 10, // Fewer workers
		BatchSize:            50,                    // Smaller batches
		RequestTimeout:       60 * time.Second,      // Longer timeouts
		BatchTimeout:         15 * time.Minute,      // More time for completion
		HTTPPoolSize:         100,                   // Smaller HTTP pool
		RetryAttempts:        5,                     // More retries
		RetryDelay:           2 * time.Second,       // Longer retry delays
	}
}

// OptimizedBatchFetcher provides advanced batch fetching with performance monitoring
type OptimizedBatchFetcher struct {
	config      *ConcurrencyConfig
	rateLimiter *rate.Limiter
	stats       *BatchStats
	httpClient  *http.Client
	mu          sync.RWMutex
}

// BatchStats tracks performance metrics
type BatchStats struct {
	TotalRequests    int64
	SuccessfulReqs   int64
	FailedReqs       int64
	RetryCount       int64
	TotalDataFetched int64
	StartTime        time.Time
	EndTime          time.Time
	mu               sync.RWMutex
}

// NewOptimizedBatchFetcher creates a new optimized batch fetcher
func NewOptimizedBatchFetcher(config *ConcurrencyConfig) *OptimizedBatchFetcher {
	if config == nil {
		config = DefaultConcurrencyConfig()
	}

	// Create rate limiter with the specified configuration
	rateLimiter := rate.NewLimiter(rate.Limit(config.MaxRequestsPerSecond), config.BurstCapacity)

	// Configure HTTP client
	transport := &http.Transport{
		MaxIdleConns:        config.HTTPPoolSize,
		MaxIdleConnsPerHost: config.HTTPPoolSize / 5,
		MaxConnsPerHost:     config.HTTPPoolSize / 2,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		DisableKeepAlives:   false,
		WriteBufferSize:     64 * 1024, // 64KB buffers
		ReadBufferSize:      64 * 1024,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.RequestTimeout,
	}

	return &OptimizedBatchFetcher{
		config:      config,
		rateLimiter: rateLimiter,
		stats:       &BatchStats{StartTime: time.Now()},
		httpClient:  httpClient,
	}
}

// FetchBlocksAdvanced fetches blocks with advanced performance optimization
func (f *OptimizedBatchFetcher) FetchBlocksAdvanced(ctx context.Context, startSlot uint64, count int) ([]model.BlockResult, error) {
	f.stats.mu.Lock()
	f.stats.StartTime = time.Now()
	f.stats.TotalRequests = int64(count)
	f.stats.mu.Unlock()

	// Create context with batch timeout
	batchCtx, cancel := context.WithTimeout(ctx, f.config.BatchTimeout)
	defer cancel()

	// Process in chunks if count is larger than batch size
	var allResults []model.BlockResult
	var resultMu sync.Mutex

	for processed := 0; processed < count; processed += f.config.BatchSize {
		batchEnd := processed + f.config.BatchSize
		if batchEnd > count {
			batchEnd = count
		}

		batchCount := batchEnd - processed
		batchStartSlot := startSlot + uint64(processed)

		// Process this batch
		batchResults, err := f.processBatch(batchCtx, batchStartSlot, batchCount)
		if err != nil {
			return allResults, fmt.Errorf("batch processing failed: %w", err)
		}

		resultMu.Lock()
		allResults = append(allResults, batchResults...)
		resultMu.Unlock()

		// Progress update
		fmt.Printf("Progress: %d/%d blocks processed (%.1f%%)\n",
			batchEnd, count, float64(batchEnd)/float64(count)*100)
	}

	f.stats.mu.Lock()
	f.stats.EndTime = time.Now()
	f.stats.mu.Unlock()

	return allResults, nil
}

// processBatch processes a single batch of blocks
func (f *OptimizedBatchFetcher) processBatch(ctx context.Context, startSlot uint64, count int) ([]model.BlockResult, error) {
	// Create channels
	slotChan := make(chan uint64, count)
	resultChan := make(chan model.BlockResult, count)

	// Send slots to channel
	for i := 0; i < count; i++ {
		slotChan <- startSlot + uint64(i)
	}
	close(slotChan)

	// Determine worker count
	workerCount := f.config.MaxConcurrentWorkers
	if workerCount > count {
		workerCount = count
	}

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.worker(ctx, slotChan, resultChan)
		}()
	}

	// Close result channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make([]model.BlockResult, 0, count)
	for result := range resultChan {
		results = append(results, result)

		// Update stats
		f.stats.mu.Lock()
		if result.Error != nil {
			f.stats.FailedReqs++
		} else {
			f.stats.SuccessfulReqs++
			f.stats.TotalDataFetched += int64(len(result.Block.Transactions))
		}
		f.stats.mu.Unlock()
	}

	return results, nil
}

// worker processes individual block requests with retry logic
func (f *OptimizedBatchFetcher) worker(ctx context.Context, slotChan <-chan uint64, resultChan chan<- model.BlockResult) {
	for slot := range slotChan {
		var block model.Block
		var err error

		// Retry logic
		for attempt := 0; attempt <= f.config.RetryAttempts; attempt++ {
			if attempt > 0 {
				// Update retry stats
				f.stats.mu.Lock()
				f.stats.RetryCount++
				f.stats.mu.Unlock()

				// Wait before retry
				select {
				case <-ctx.Done():
					err = ctx.Err()
					goto sendResult
				case <-time.After(f.config.RetryDelay):
				}
			}

			// Wait for rate limiter
			if err = f.rateLimiter.Wait(ctx); err != nil {
				break
			}

			// Make the request
			block, err = f.fetchBlockWithCustomClient(ctx, slot)
			if err == nil {
				break // Success, no need to retry
			}
		}

	sendResult:
		resultChan <- model.BlockResult{
			Block: block,
			Slot:  slot,
			Error: err,
		}
	}
}

// fetchBlockWithCustomClient uses the optimized HTTP client
func (f *OptimizedBatchFetcher) fetchBlockWithCustomClient(ctx context.Context, blockNumber uint64) (model.Block, error) {
	// This is similar to GetBlock but uses the fetcher's custom HTTP client
	var block model.Block

	// Prepare RPC request
	request := model.RPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "getBlock",
		Params: []interface{}{
			blockNumber,
			map[string]interface{}{
				"encoding":                       "json",
				"transactionDetails":             "full",
				"rewards":                        true,
				"maxSupportedTransactionVersion": 0,
			},
		},
	}

	// Convert request to JSON
	requestData, err := json.Marshal(request)
	if err != nil {
		return block, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	rpcURL := config.SvcConfig.Solana.RpcUrl
	if rpcURL == "" {
		return block, fmt.Errorf("Solana RPC URL not configured")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", rpcURL, bytes.NewBuffer(requestData))
	if err != nil {
		return block, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request using custom client
	resp, err := f.httpClient.Do(httpReq)
	if err != nil {
		return block, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return block, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var rpcResp model.RPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return block, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for RPC error
	if rpcResp.Error != nil {
		return block, fmt.Errorf("RPC error: code=%d, message=%s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	// Check if result is null (block not found)
	if rpcResp.Result == nil {
		return block, fmt.Errorf("block %d not found", blockNumber)
	}

	// Convert result to block struct
	resultBytes, err := json.Marshal(rpcResp.Result)
	if err != nil {
		return block, fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(resultBytes, &block); err != nil {
		return block, fmt.Errorf("failed to unmarshal block data: %w", err)
	}

	return block, nil
}

// GetStats returns current performance statistics
func (f *OptimizedBatchFetcher) GetStats() BatchStats {
	f.stats.mu.RLock()
	defer f.stats.mu.RUnlock()

	stats := *f.stats
	if stats.EndTime.IsZero() {
		stats.EndTime = time.Now()
	}

	return stats
}

// PrintDetailedStats prints comprehensive performance statistics
func (f *OptimizedBatchFetcher) PrintDetailedStats() {
	stats := f.GetStats()
	duration := stats.EndTime.Sub(stats.StartTime)

	fmt.Printf("\n🔥 ADVANCED PERFORMANCE STATISTICS 🔥\n")
	fmt.Printf("=====================================\n")
	fmt.Printf("Total Duration: %v\n", duration)
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Successful: %d (%.2f%%)\n", stats.SuccessfulReqs, float64(stats.SuccessfulReqs)/float64(stats.TotalRequests)*100)
	fmt.Printf("Failed: %d (%.2f%%)\n", stats.FailedReqs, float64(stats.FailedReqs)/float64(stats.TotalRequests)*100)
	fmt.Printf("Retries: %d\n", stats.RetryCount)
	fmt.Printf("Total Transactions Fetched: %d\n", stats.TotalDataFetched)

	if duration.Seconds() > 0 {
		fmt.Printf("Request Rate: %.2f req/sec\n", float64(stats.TotalRequests)/duration.Seconds())
		fmt.Printf("Success Rate: %.2f successful/sec\n", float64(stats.SuccessfulReqs)/duration.Seconds())
	}

	if stats.SuccessfulReqs > 0 {
		fmt.Printf("Avg Transactions per Block: %.2f\n", float64(stats.TotalDataFetched)/float64(stats.SuccessfulReqs))
	}

	fmt.Printf("=====================================\n")
}
