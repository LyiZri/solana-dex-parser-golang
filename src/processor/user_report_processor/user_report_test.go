package user_report_processor

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-solana-parse/src/test"
)

func TestUserReportProcessor(t *testing.T) {

	test.TestEnvInit()

	var address = "5TLRz619uQoDEPtyUK2z4NLVMQF6xrV9hYPRFMXqbNRV"

	userProcessor := NewUserReportProcessor()

	userReport, err := userProcessor.ProcessSingleUserReport(address)
	if err != nil {
		t.Fatalf("Failed to process user report: %v", err)
	}

	json, err := json.Marshal(userReport)
	if err != nil {
		t.Fatalf("Failed to marshal user report: %v", err)
	}

	fmt.Println(string(json))
}
