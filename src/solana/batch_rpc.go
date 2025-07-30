package solana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-solana-parse/src/config"
	"github.com/go-solana-parse/src/model"
	"golang.org/x/time/rate"
)

// BatchRPCRequest represents a single request in a batch
type BatchRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// BatchRPCResponse represents a single response in a batch
type BatchRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Result  interface{}     `json:"result"`
	Error   *model.RPCError `json:"error,omitempty"`
}

// BatchBlockResult represents the result of a batch block fetch
type BatchBlockResult struct {
	Slot     uint64
	Block    model.Block
	Error    error
	Duration time.Duration
}

// BatchRPCConfig holds configuration for batch RPC requests
type BatchRPCConfig struct {
	MaxBatchSize         int           // Maximum number of requests per batch
	BatchTimeout         time.Duration // Timeout for each batch request
	MaxRequestsPerSecond int           // Rate limiting
	BurstCapacity        int           // Burst capacity for rate limiter
	HTTPPoolSize         int           // HTTP connection pool size
	RetryAttempts        int           // Number of retry attempts
	RetryDelay           time.Duration // Delay between retries
}

// DefaultBatchRPCConfig returns optimized default configuration for batch RPC
func DefaultBatchRPCConfig() *BatchRPCConfig {
	return &BatchRPCConfig{
		MaxBatchSize:         50,               // 50 blocks per batch request
		BatchTimeout:         60 * time.Second, // 60 seconds per batch
		MaxRequestsPerSecond: 100,              // Conservative rate limit for batch requests
		BurstCapacity:        20,               // Allow bursts
		HTTPPoolSize:         100,              // HTTP connections
		RetryAttempts:        3,                // Retry attempts
		RetryDelay:           1 * time.Second,  // Retry delay
	}
}

// HighPerformanceBatchRPCConfig returns configuration optimized for maximum batch throughput
func HighPerformanceBatchRPCConfig() *BatchRPCConfig {
	return &BatchRPCConfig{
		MaxBatchSize:         50,                     // 优化批次大小
		BatchTimeout:         30 * time.Second,       // 30 second timeout
		MaxRequestsPerSecond: 300,                    // 提高速率限制
		BurstCapacity:        100,                    // 提高突发容量
		HTTPPoolSize:         300,                    // 提高连接池
		RetryAttempts:        3,                      // 更多重试
		RetryDelay:           300 * time.Millisecond, // 平衡重试速度
	}
}

// BatchRPCFetcher provides high-performance batch RPC functionality
type BatchRPCFetcher struct {
	config      *BatchRPCConfig
	rateLimiter *rate.Limiter
	httpClient  *http.Client
	stats       *BatchRPCStats
	mu          sync.RWMutex
}

// BatchRPCStats tracks batch RPC performance metrics
type BatchRPCStats struct {
	TotalBatches      int64
	TotalBlocks       int64
	SuccessfulBatches int64
	FailedBatches     int64
	SuccessfulBlocks  int64
	FailedBlocks      int64
	TotalRetries      int64
	TotalDataFetched  int64
	AverageBatchSize  float64
	StartTime         time.Time
	EndTime           time.Time
	mu                sync.RWMutex
}

// NewBatchRPCFetcher creates a new optimized batch RPC fetcher
func NewBatchRPCFetcher(config *BatchRPCConfig) *BatchRPCFetcher {
	if config == nil {
		config = DefaultBatchRPCConfig()
	}

	// Create rate limiter
	rateLimiter := rate.NewLimiter(rate.Limit(config.MaxRequestsPerSecond), config.BurstCapacity)

	// Configure HTTP client for batch requests
	transport := &http.Transport{
		MaxIdleConns:        config.HTTPPoolSize,
		MaxIdleConnsPerHost: config.HTTPPoolSize / 2, // 增加每主机连接数
		MaxConnsPerHost:     config.HTTPPoolSize,     // 最大化连接数
		IdleConnTimeout:     30 * time.Second,        // 减少空闲超时
		DisableCompression:  false,
		DisableKeepAlives:   false,
		WriteBufferSize:     256 * 1024, // 更大的写缓冲区
		ReadBufferSize:      256 * 1024, // 更大的读缓冲区
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.BatchTimeout,
	}

	return &BatchRPCFetcher{
		config:      config,
		rateLimiter: rateLimiter,
		httpClient:  httpClient,
		stats:       &BatchRPCStats{StartTime: time.Now()},
	}
}

// FetchBlocksBatch fetches multiple blocks using concurrent batch JSON-RPC requests
func (f *BatchRPCFetcher) FetchBlocksBatch(ctx context.Context, startSlot uint64, count int) ([]BatchBlockResult, error) {
	f.stats.mu.Lock()
	f.stats.StartTime = time.Now()
	f.stats.TotalBlocks = int64(count)
	f.stats.mu.Unlock()

	// Calculate number of batches
	numBatches := (count + f.config.MaxBatchSize - 1) / f.config.MaxBatchSize

	// Create channels for batch jobs and results
	batchJobs := make(chan []uint64, numBatches)
	results := make(chan []BatchBlockResult, numBatches)

	// Start concurrent workers
	const maxConcurrentBatches = 15 // 优化并发数量
	var wg sync.WaitGroup

	for i := 0; i < maxConcurrentBatches && i < numBatches; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for slots := range batchJobs {
				batchResults, err := f.processBatchWithRetry(ctx, slots)
				if err != nil {
					// Send empty results on error
					results <- []BatchBlockResult{}
				} else {
					results <- batchResults
				}
			}
		}()
	}

	// Send batch jobs
	go func() {
		defer close(batchJobs)
		for processed := 0; processed < count; processed += f.config.MaxBatchSize {
			batchEnd := processed + f.config.MaxBatchSize
			if batchEnd > count {
				batchEnd = count
			}

			batchCount := batchEnd - processed
			batchStartSlot := startSlot + uint64(processed)

			slots := make([]uint64, batchCount)
			for i := 0; i < batchCount; i++ {
				slots[i] = batchStartSlot + uint64(i)
			}

			batchJobs <- slots
		}
	}()

	// Collect results with progress tracking
	var allResults []BatchBlockResult
	processedBatches := 0

	go func() {
		wg.Wait()
		close(results)
	}()

	for batchResults := range results {
		allResults = append(allResults, batchResults...)
		processedBatches++

		f.stats.mu.Lock()
		f.stats.TotalBatches++
		f.stats.mu.Unlock()

		// Simple progress log
		elapsed := time.Since(f.stats.StartTime)
		successCount := 0
		for _, r := range allResults {
			if r.Error == nil {
				successCount++
			}
		}

		blocksProcessed := len(allResults)
		if blocksProcessed > 0 {
			successRate := float64(successCount) / float64(blocksProcessed) * 100
			fmt.Printf("Progress: %d/%d blocks | Time: %v | Success: %.1f%%\n",
				blocksProcessed, count, elapsed.Round(time.Second), successRate)
		}
	}

	f.stats.mu.Lock()
	f.stats.EndTime = time.Now()
	if f.stats.TotalBatches > 0 {
		f.stats.AverageBatchSize = float64(f.stats.TotalBlocks) / float64(f.stats.TotalBatches)
	}
	f.stats.mu.Unlock()

	return allResults, nil
}

// processBatchWithRetry processes a batch of slots with retry logic
func (f *BatchRPCFetcher) processBatchWithRetry(ctx context.Context, slots []uint64) ([]BatchBlockResult, error) {
	var results []BatchBlockResult
	var err error

	for attempt := 0; attempt <= f.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			f.stats.mu.Lock()
			f.stats.TotalRetries++
			f.stats.mu.Unlock()

			// Wait before retry
			select {
			case <-ctx.Done():
				return results, ctx.Err()
			case <-time.After(f.config.RetryDelay):
			}
		}

		// Wait for rate limiter (one request per batch)
		if err = f.rateLimiter.Wait(ctx); err != nil {
			continue
		}

		// Process the batch
		results, err = f.processBatch(ctx, slots)
		if err == nil {
			break // Success, no need to retry
		}
	}

	return results, err
}

// processBatch sends a single batch JSON-RPC request for multiple blocks
func (f *BatchRPCFetcher) processBatch(ctx context.Context, slots []uint64) ([]BatchBlockResult, error) {
	startTime := time.Now()

	// Create batch request
	batchRequest := make([]BatchRPCRequest, len(slots))
	for i, slot := range slots {
		batchRequest[i] = BatchRPCRequest{
			JSONRPC: "2.0",
			ID:      fmt.Sprintf("block-%d", slot),
			Method:  "getBlock",
			Params: []interface{}{
				slot,
				map[string]interface{}{
					"encoding":                       "json",
					"transactionDetails":             "full",
					"rewards":                        true,
					"maxSupportedTransactionVersion": 0,
				},
			},
		}
	}

	// Convert to JSON
	requestData, err := json.Marshal(batchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch request: %w", err)
	}

	// Create HTTP request
	rpcURL := config.SvcConfig.Solana.RpcUrl
	if rpcURL == "" {
		return nil, fmt.Errorf("Solana RPC URL not configured")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", rpcURL, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := f.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// Parse batch response
	var batchResponse []BatchRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&batchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode batch response: %w", err)
	}

	// Process responses
	results := make([]BatchBlockResult, 0, len(slots))
	duration := time.Since(startTime)

	for _, slot := range slots {
		result := BatchBlockResult{
			Slot:     slot,
			Duration: duration,
		}

		// Find corresponding response
		found := false
		for _, resp := range batchResponse {
			if resp.ID == fmt.Sprintf("block-%d", slot) {
				found = true
				if resp.Error != nil {
					result.Error = fmt.Errorf("RPC error: code=%d, message=%s", resp.Error.Code, resp.Error.Message)
					f.stats.mu.Lock()
					f.stats.FailedBlocks++
					f.stats.mu.Unlock()
				} else if resp.Result == nil {
					result.Error = fmt.Errorf("block %d not found", slot)
					f.stats.mu.Lock()
					f.stats.FailedBlocks++
					f.stats.mu.Unlock()
				} else {
					// Convert result to block struct
					resultBytes, err := json.Marshal(resp.Result)
					if err != nil {
						result.Error = fmt.Errorf("failed to marshal result: %w", err)
						f.stats.mu.Lock()
						f.stats.FailedBlocks++
						f.stats.mu.Unlock()
					} else {
						if err := json.Unmarshal(resultBytes, &result.Block); err != nil {
							result.Error = fmt.Errorf("failed to unmarshal block data: %w", err)
							f.stats.mu.Lock()
							f.stats.FailedBlocks++
							f.stats.mu.Unlock()
						} else {
							f.stats.mu.Lock()
							f.stats.SuccessfulBlocks++
							f.stats.TotalDataFetched += int64(len(result.Block.Transactions))
							f.stats.mu.Unlock()
						}
					}
				}
				break
			}
		}

		if !found {
			result.Error = fmt.Errorf("no response found for block %d", slot)
			f.stats.mu.Lock()
			f.stats.FailedBlocks++
			f.stats.mu.Unlock()
		}

		results = append(results, result)
	}

	return results, nil
}

// GetStats returns current batch RPC performance statistics
func (f *BatchRPCFetcher) GetStats() BatchRPCStats {
	f.stats.mu.RLock()
	defer f.stats.mu.RUnlock()

	stats := *f.stats
	if stats.EndTime.IsZero() {
		stats.EndTime = time.Now()
	}

	return stats
}

// PrintDetailedBatchStats prints essential batch RPC performance statistics
func (f *BatchRPCFetcher) PrintDetailedBatchStats() {
	stats := f.GetStats()
	duration := stats.EndTime.Sub(stats.StartTime)

	fmt.Printf("\nStats: %d blocks | %v | %.1f%% success | %.1f blocks/sec\n",
		stats.TotalBlocks,
		duration.Round(time.Second),
		float64(stats.SuccessfulBlocks)/float64(stats.TotalBlocks)*100,
		float64(stats.TotalBlocks)/duration.Seconds())
}
