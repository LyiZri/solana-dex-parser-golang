package db

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/go-solana-parse/src/config"
)

func Connect() (driver.Conn, error) {
	cfg := config.SvcConfig.ClickHouse
	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
			Auth: clickhouse.Auth{
				Database: cfg.DbName,
				Username: cfg.User,
				Password: cfg.Password,
			},
			Protocol: clickhouse.HTTP,
			ClientInfo: clickhouse.ClientInfo{
				Products: []struct {
					Name    string
					Version string
				}{
					{Name: "go-solana-parse", Version: "0.1"},
				},
			},
			Debug: true,
		})
	)

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}
	return conn, nil
}
