package user_report_processor

import (
	"fmt"
	"log"

	ckdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/go-solana-parse/src/db"
	"github.com/go-solana-parse/src/db/mysql"
	"github.com/go-solana-parse/src/model"
	"github.com/go-solana-parse/src/service"
)

// UserReportProcessor 用户报告处理器
type UserReportProcessor struct {
	calculator       *service.UserReportCalculator
	clickhouseClient ckdriver.Conn
}

var UserReportProcessorNsp = &UserReportProcessor{}

// NewUserReportProcessor 创建新的用户报告处理器
func NewUserReportProcessor() *UserReportProcessor {
	return &UserReportProcessor{
		clickhouseClient: db.ClickHouseClient,
	}
}

// 延迟初始化calculator，确保ClickHouse连接已建立
func (processor *UserReportProcessor) getCalculator() *service.UserReportCalculator {
	if processor.calculator == nil {
		processor.calculator = service.NewUserReportCalculator()
	}
	return processor.calculator
}

// ProcessAllUserReports 处理所有用户报告
func (processor *UserReportProcessor) ProcessAllUserReports() error {
	log.Println("开始处理所有用户报告...")

	// 步骤 1: 从 ClickHouse view表中获取所有钱包地址，按交易量升序排序
	addresses, err := processor.getAllUniqueAddressesOrderByTradeCount()
	if err != nil {
		return fmt.Errorf("获取地址列表失败: %v", err)
	}

	log.Printf("找到 %d 个钱包地址需要处理（按交易量升序排序）\n", len(addresses))

	// 步骤 2: 针对每个地址进行逐个分析
	successCount := 0
	errorCount := 0

	for i, address := range addresses {
		log.Printf("处理地址 %d/%d: %s\n", i+1, len(addresses), address)

		_, err = processor.ProcessSingleUserReport(address)
		if err != nil {
			log.Printf("处理地址 %s 失败: %v\n", address, err)
			errorCount++
			continue
		}

		successCount++

		// 每处理100个地址打印一次进度
		if (i+1)%100 == 0 {
			log.Printf("进度: %d/%d 已完成，成功: %d，失败: %d\n", i+1, len(addresses), successCount, errorCount)
		}
	}

	log.Printf("用户报告处理完成！总计: %d，成功: %d，失败: %d\n", len(addresses), successCount, errorCount)
	return nil
}

// ProcessSingleUserReport 处理单个用户报告
func (processor *UserReportProcessor) ProcessSingleUserReport(address string) (*mysql.UserReport, error) {
	log.Printf("开始处理用户 %s 的报告...\n", address)

	// 计算用户报告
	userReport, err := processor.getCalculator().CalculateUserReport(address)
	if err != nil {
		return nil, fmt.Errorf("计算用户报告失败: %v", err)
	}

	// 保存到数据库
	err = mysql.UserReportNsp.SaveOrUpdateUserReport(db.DBClient, userReport)
	if err != nil {
		return nil, fmt.Errorf("保存用户报告失败: %v", err)
	}

	log.Printf("用户 %s 的报告处理完成\n", address)
	return userReport, nil
}

// getAllUniqueAddresses 获取所有唯一地址
func (processor *UserReportProcessor) getAllUniqueAddresses() ([]string, error) {
	return processor.getCalculator().GetAllUniqueAddresses()
}

// getAllUniqueAddressesOrderByTradeCount 获取所有唯一地址，按交易量升序排序
func (processor *UserReportProcessor) getAllUniqueAddressesOrderByTradeCount() ([]string, error) {
	return processor.getCalculator().GetAllUniqueAddressesOrderByTradeCount()
}

// GetAllUniqueAddresses 获取所有唯一地址（公开方法）
func (processor *UserReportProcessor) GetAllUniqueAddresses() ([]string, error) {
	return processor.getAllUniqueAddresses()
}

// GetAllUniqueAddressesOrderByTradeCount 获取所有唯一地址，按交易量升序排序（公开方法）
func (processor *UserReportProcessor) GetAllUniqueAddressesOrderByTradeCount() ([]string, error) {
	return processor.getAllUniqueAddressesOrderByTradeCount()
}

// GetUserReportSummary 获取用户报告汇总信息
func (processor *UserReportProcessor) GetUserReportSummary() (*model.UserReportSummary, error) {
	return mysql.UserReportNsp.GetUserReportSummary(db.DBClient)
}

// GetUserReportByAddress 根据地址获取用户报告
func (processor *UserReportProcessor) GetUserReportByAddress(address string) (*mysql.UserReport, error) {
	return mysql.UserReportNsp.GetUserReportByAddress(db.DBClient, address)
}

// GetAllUserReports 获取所有用户报告
func (processor *UserReportProcessor) GetAllUserReports(limit, offset int) ([]*mysql.UserReport, error) {
	return mysql.UserReportNsp.GetAllUserReports(db.DBClient, limit, offset)
}

// DeleteUserReport 删除用户报告
func (processor *UserReportProcessor) DeleteUserReport(address string) error {
	return mysql.UserReportNsp.DeleteUserReport(db.DBClient, address)
}
