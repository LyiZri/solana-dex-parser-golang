package seelog

import (
	"os"

	"github.com/cihub/seelog"
)

func init() {
}

const (
	SERVICE_NAME    = "SERVICE_NAME"
	LOG_LEVEL_ERROR = "ERROR"
	LOG_LEVEL_INFO  = "INFO"
	LOG_LEVEL_DEBUG = "DEBUG"
	LOG_LEVEL_WARN  = "WARN"
)

func createAppNameFormatter(params string) seelog.FormatterFunc {
	return func(message string, level seelog.LogLevel, context seelog.LogContextInterface) interface{} {
		serviceName := os.Getenv(SERVICE_NAME)
		if serviceName == "" {
			serviceName = "None"
		}
		return serviceName
	}
}

// Errorf function
func Errorf(logger seelog.LoggerInterface, format string, params ...interface{}) error {
	defer logger.Flush()
	prefix := getPrefix(LOG_LEVEL_ERROR)
	return logger.Errorf(prefix+format, params...)
}

// Error function
func Error(logger seelog.LoggerInterface, params ...interface{}) error {
	defer logger.Flush()
	prefix := getPrefix(LOG_LEVEL_ERROR)
	var newParams []interface{}
	newParams = append(newParams, prefix)
	for _, param := range params {
		newParams = append(newParams, param)
	}
	return logger.Error(prefix, newParams)
}

// Infof function
func Infof(logger seelog.LoggerInterface, format string, params ...interface{}) {
	defer logger.Flush()
	logger.Infof(getPrefix(LOG_LEVEL_INFO)+format, params...)
}

// Info function
func Info(logger seelog.LoggerInterface, params ...interface{}) {
	defer logger.Flush()
	prefix := getPrefix(LOG_LEVEL_INFO)
	var newParams []interface{}
	newParams = append(newParams, prefix)
	for _, param := range params {
		newParams = append(newParams, param)
	}
	logger.Info(newParams...)
}

// Debugf function
func Debugf(logger seelog.LoggerInterface, format string, params ...interface{}) {
	defer logger.Flush()
	prefix := getPrefix(LOG_LEVEL_DEBUG)
	logger.Debugf(prefix+format, params...)
}

// Debug function
func Debug(logger seelog.LoggerInterface, params ...interface{}) {
	defer logger.Flush()
	prefix := getPrefix(LOG_LEVEL_DEBUG)
	logger.Debug(prefix, params)
	// var newParams []interface{}
	// newParams = append(newParams, prefix)
	// for _, param := range params {
	// 	newParams = append(newParams, param)
	// }
}

// Warnf function
func Warnf(logger seelog.LoggerInterface, format string, params ...interface{}) {
	defer logger.Flush()
	prefix := getPrefix(LOG_LEVEL_WARN)
	logger.Warnf(prefix+format, params...)
}

// Warn function
func Warn(logger seelog.LoggerInterface, params ...interface{}) {
	defer logger.Flush()
	prefix := getPrefix(LOG_LEVEL_WARN)
	logger.Warn(prefix, params)
	// var newParams []interface{}
	// newParams = append(newParams, prefix)
	// for _, param := range params {
	// 	newParams = append(newParams, param)
	// }
}

// // Flush function
// func Flush() {
// 	seelog.Flush()
// }
