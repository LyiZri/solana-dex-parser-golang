package mysql

import (
	"fmt"
	"testing"

	"github.com/go-solana-parse/src/db"
	"github.com/go-solana-parse/src/test"
)

func TestUserReport_GetCount(t *testing.T) {

	test.TestEnvInit()
	count, err := UserReportNsp.GetCount(db.DBClient)
	if err != nil {
		t.Fatalf("获取用户报告数量失败: %v", err)
	}

	fmt.Println("用户报告数量:", count)
}
