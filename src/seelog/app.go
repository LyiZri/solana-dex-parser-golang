package seelog

// Errorf function
func AppErrorf(format string, params ...interface{}) error {
	return Errorf(appLogger, format, params...)
}

// Error function
func AppError(params ...interface{}) error {
	return Error(appLogger, params...)
}

// Infof function
func AppInfof(format string, params ...interface{}) {
	Infof(appLogger, format, params...)
}

// Info function
func AppInfo(params ...interface{}) {
	Info(appLogger, params...)
}

// Debugf function
func AppDebugf(format string, params ...interface{}) {
	Debugf(appLogger, format, params...)
}

// Debug function
func AppDebug(params ...interface{}) {
	Debug(appLogger, params...)
}

// Warnf function
func AppWarnf(format string, params ...interface{}) {
	Warnf(appLogger, format, params...)
}

// Warn function
func AppWarn(params ...interface{}) {
	Warn(appLogger, params...)
}
