package seelog

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/google/uuid"
	"github.com/petermattis/goid"
)

// Local request ID implementation to replace the problematic package
var requestIDStore sync.Map

// setRequestID sets a request ID for the current goroutine
func setRequestID(id string) {
	goroutineID := goid.Get()
	requestIDStore.Store(goroutineID, id)
}

// getRequestID gets the request ID for the current goroutine
func getRequestID() interface{} {
	goroutineID := goid.Get()
	if id, exists := requestIDStore.Load(goroutineID); exists {
		return id
	}
	return nil
}

func getPrefix(level string) string {
	callerInfo := getCallerName()
	requestID := getRequestID()
	if requestID == nil {
		requestIDStr := fmt.Sprintf("%+v", uuid.New())
		setRequestID(requestIDStr)
		requestID = requestIDStr
	}
	prefix := fmt.Sprintf("%v %s [%d] %s: ", requestID, level, goid.Get(), callerInfo)
	return prefix
}

func getCallerName() string {
	_, file, line, _ := runtime.Caller(4)
	return fmt.Sprintf("%s.%d", filepath.Base(file), line)
}
