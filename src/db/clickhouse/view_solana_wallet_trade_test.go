package clickhouse

import (
	"fmt"
	"testing"

	"github.com/go-solana-parse/src/db"
	"github.com/go-solana-parse/src/test"
)

func TestViewSolanaWalletTradeCount(t *testing.T) {
	test.TestEnvInit()

	addresses, err := ViewSolanaWalletTradeCountNsp.GetWalletAddressesByTradeCountASC(db.ClickHouseClient)
	if err != nil {
		t.Fatalf("Failed to get wallet addresses by trade count: %v", err)
	}

	fmt.Println("wallet addresses by trade count", len(addresses))

}
