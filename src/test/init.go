package test

import (
	"github.com/go-solana-parse/src/config"
	"github.com/go-solana-parse/src/db"
)

func TestEnvInit() {
	config.LoadSvcConfigFromPath()
	db.InitDB()
	db.InitClickHouseV2()
}
