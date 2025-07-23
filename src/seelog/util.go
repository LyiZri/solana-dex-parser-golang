package seelog

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
	"github.com/petermattis/goid"
	"github.com/quanhengzhuang/requestid"
)

func getPrefix(level string) string {
	callerInfo := getCallerName()
	requestID := requestid.Get()
	if requestID == nil {
		requestIDStr := fmt.Sprintf("%+v", uuid.New())
		requestid.Set(requestIDStr)
	}
	prefix := fmt.Sprintf("%v %s [%d] %s: ", requestid.Get(), level, goid.Get(), callerInfo)
	return prefix
}

func getCallerName() string {
	_, file, line, _ := runtime.Caller(4)
	return fmt.Sprintf("%s.%d", filepath.Base(file), line)
}
