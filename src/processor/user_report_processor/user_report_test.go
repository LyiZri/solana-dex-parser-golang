package user_report_processor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-solana-parse/src/db"
	"github.com/go-solana-parse/src/test"
)

func TestUserReportProcessor(t *testing.T) {

	test.TestEnvInit()

	var address = "CZP27EA547NXG7peLE2z9SUCtkwRbmEHJwGCAAQaCgfD"

	userProcessor := NewUserReportProcessor(db.DBClient, db.ClickHouseClient)

	userReport, err := userProcessor.ProcessSingleUserReport(address)
	if err != nil {
		t.Fatalf("Failed to process user report: %v", err)
	}

	// 将结果转换为JSON格式
	jsonData, err := json.MarshalIndent(userReport, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal user report to JSON: %v", err)
	}

	// 创建输出目录
	outputDir := "test_results"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// 生成文件名，包含时间戳
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("user_report_%s_%s.json", address, timestamp)
	filepath := filepath.Join(outputDir, filename)

	// 保存JSON文件
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write JSON file: %v", err)
	}

	// 输出结果信息
	fmt.Printf("✅ 用户报告生成成功!\n")
	fmt.Printf("📊 用户地址: %s\n", address)
	fmt.Printf("📁 文件保存至: %s\n", filepath)
	fmt.Printf("📄 文件大小: %d bytes\n", len(jsonData))
	
	// 也输出到控制台便于调试
	fmt.Printf("\n📋 报告摘要:\n%s\n", string(jsonData))
}
