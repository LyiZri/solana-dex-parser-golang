package clickhouse

import (
	"fmt"
	"testing"

	"github.com/go-solana-parse/src/db"
	"github.com/go-solana-parse/src/test"
)

func TestGetSolanaUsdPrice(t *testing.T) {
	test.TestEnvInit()

	solanaUsdPrice := &SolanaUsdPrice{}
	price, err := solanaUsdPrice.GetSolanaUsdPrice(db.ClickHouseClient, 347599307)
	if err != nil {
		t.Fatalf("Failed to get solana usd price: %v", err)
	}

	fmt.Println(price)
}
