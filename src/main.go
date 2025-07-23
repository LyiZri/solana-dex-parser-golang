package main

import (
	"sync"

	"github.com/go-solana-parse/src/config"
	"github.com/go-solana-parse/src/db"
	"github.com/go-solana-parse/src/processor/user_report_processor"
)

var twg sync.WaitGroup

var wg sync.WaitGroup

func main() {
	configInit()
	user_report_processor.UserReportProcessorNsp.ProcessAllUserReports()
}

func configInit() {
	config.LoadSvcConfig()
	db.InitDB()
	db.InitClickHouseV2()
}
