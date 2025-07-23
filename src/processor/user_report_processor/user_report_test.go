package user_report_processor

import (
	"testing"

	"github.com/go-solana-parse/src/test"
)

func TestUserReportProcessor(t *testing.T) {

	test.TestEnvInit()

	var address = "CskjdjYDE7VD2vLkmnaT2nuK8Tps5e1sEP34oQPCBKuH"

	userProcessor := NewUserReportProcessor()

	_, err := userProcessor.ProcessSingleUserReport(address)
	if err != nil {
		t.Fatalf("Failed to process user report: %v", err)
	}
}
