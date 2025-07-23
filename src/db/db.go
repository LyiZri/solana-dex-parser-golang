package db

import (
	"fmt"

	ckdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/go-solana-parse/src/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DBClient *gorm.DB
var ClickHouseClient ckdriver.Conn

func InitDB() error {
	dbUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.SvcConfig.DB.User,
		config.SvcConfig.DB.Password,
		config.SvcConfig.DB.Host,
		config.SvcConfig.DB.Port,
		config.SvcConfig.DB.DbName)
	fmt.Printf("MySQL connection string: %s\n", dbUrl)
	db, err := gorm.Open(mysql.Open(dbUrl), &gorm.Config{
		// 使用默认logger替代可能有问题的自定义logger
	})
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %v", err)
	}
	DBClient = db

	// 测试连接
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping MySQL: %v", err)
	}

	fmt.Println("✅ MySQL connection and ping successful!")
	return nil
}

func InitClickHouseV2() error {
	conn, err := Connect()

	fmt.Println("conn", conn)
	if err != nil {
		return err
	}
	ClickHouseClient = conn
	return nil
}
