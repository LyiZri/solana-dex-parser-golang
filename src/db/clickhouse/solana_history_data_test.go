package clickhouse

import (
	"fmt"
	"testing"

	"github.com/go-solana-parse/src/db"
	"github.com/go-solana-parse/src/test"
)

func TestGetTotalTxCount(t *testing.T) {
	test.TestEnvInit()

	count, err := SolanaHistoryDataNsp.GetTotalTxCount(db.ClickHouseClient)
	if err != nil {
		fmt.Println("get total tx count error", err)
		t.Fatalf("Failed to connect to ClickHouse: %v", err)
	}

	fmt.Println(count)
}

func TestGetTotalUniqueAddress(t *testing.T) {
	test.TestEnvInit()

	addressList, err := SolanaHistoryDataNsp.GetTotalUniqueAddress(db.ClickHouseClient)
	if err != nil {
		t.Fatalf("Failed to connect to ClickHouse: %v", err)
	}

	fmt.Println(addressList)
}

func TestGetTotalUniqueAddressLength(t *testing.T) {
	test.TestEnvInit()

	addressList, err := SolanaHistoryDataNsp.GetTotalUniqueAddress(db.ClickHouseClient)
	if err != nil {
		t.Fatalf("Failed to connect to ClickHouse: %v", err)
	}

	fmt.Println(len(addressList))
}

func TestGetSolanaTokenPrice(t *testing.T) {
	test.TestEnvInit()

	price, err := SolanaHistoryDataNsp.GetSOLPriceFromTransactions(db.ClickHouseClient, 346937568)
	if err != nil {
		t.Fatalf("Failed to connect to ClickHouse: %v", err)
	}

	fmt.Println(price)
}

func TestGetTokenPriceAtBlock(t *testing.T) {
	test.TestEnvInit()

	tokenAddress := "F3QEA7LhaUmVVcPTdRCi62hAYd8bLwPbNgwbwPh6SE4A"

	price, err := SolanaHistoryDataNsp.GetTokenPriceAtBlock(db.ClickHouseClient, tokenAddress, 999999999)
	if err != nil {
		t.Fatalf("Failed to connect to ClickHouse: %v", err)
	}

	fmt.Println(price)
}
