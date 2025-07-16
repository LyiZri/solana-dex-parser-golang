package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-solana-parse/src/solana"
)

// BlockHandler handles HTTP requests for block data
type BlockHandler struct{}

// GetBlockHandler returns block information for a given block number
// URL pattern: /block/{blockNumber}
func (h *BlockHandler) GetBlockHandler(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Extract block number from URL path or query parameter
	blockNumStr := r.URL.Query().Get("block")
	if blockNumStr == "" {
		// Try to get from path parameter (assuming mux router usage)
		blockNumStr = r.PathValue("blockNumber")
	}

	if blockNumStr == "" {
		http.Error(w, `{"error":"block number is required"}`, http.StatusBadRequest)
		return
	}

	// Convert string to uint64
	blockNumber, err := strconv.ParseUint(blockNumStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid block number"}`, http.StatusBadRequest)
		return
	}

	// Get block data using our GetBlock function
	ctx := context.Background()
	block, err := solana.GetBlock(ctx, blockNumber)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get block: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Return block data as JSON
	if err := json.NewEncoder(w).Encode(block); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}

// GetBlockSummaryHandler returns a summary of block information
// URL pattern: /block/{blockNumber}/summary
func (h *BlockHandler) GetBlockSummaryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	blockNumStr := r.URL.Query().Get("block")
	if blockNumStr == "" {
		blockNumStr = r.PathValue("blockNumber")
	}

	if blockNumStr == "" {
		http.Error(w, `{"error":"block number is required"}`, http.StatusBadRequest)
		return
	}

	blockNumber, err := strconv.ParseUint(blockNumStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid block number"}`, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	block, err := solana.GetBlock(ctx, blockNumber)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get block: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Create a summary response
	summary := map[string]interface{}{
		"blockNumber":      blockNumber,
		"blockhash":        block.Blockhash,
		"parentSlot":       block.ParentSlot,
		"transactionCount": len(block.Transactions),
		"rewardCount":      len(block.Rewards),
	}

	if block.BlockHeight != nil {
		summary["blockHeight"] = *block.BlockHeight
	}

	if block.BlockTime != nil {
		summary["blockTime"] = *block.BlockTime
	}

	if err := json.NewEncoder(w).Encode(summary); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}
