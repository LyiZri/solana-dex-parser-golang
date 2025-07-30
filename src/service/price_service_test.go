package service

import (
	"fmt"
	"testing"

	"github.com/go-solana-parse/src/db"
	"github.com/go-solana-parse/src/test"
)

func TestPriceService_GetSOLPriceAtBlock(t *testing.T) {
	test.TestEnvInit()

	priceService := NewPriceService(db.ClickHouseClient)

	price, err := priceService.GetSOLPriceAtBlock(341972793)
	if err != nil {
		t.Fatalf("Failed to get SOL price: %v", err)
	}

	fmt.Println("SOL price", price)

}
