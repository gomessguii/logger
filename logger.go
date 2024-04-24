package logger

import (
	"log"
	"os"
)

const (
	INFO  = iota
	ERR   = iota
	WARN  = iota
	DEBUG = iota
)

func Log(logLevel int, format string, v ...any) {
	if logLevel == DEBUG && os.Getenv("DEBUG_ENABLED") != "1" {
		return
	}
	var preffix string
	switch logLevel {
	case ERR:
		preffix = "\033[41m[ERR]\033[0m "
	case WARN:
		preffix = "\033[43m[WARN]\033[0m "
	case DEBUG:
		preffix = "\033[40m\033[37m[DEBUG]\033[0m "
	default:
		preffix = "\033[44m[INFO]\033[0m "
	}
	preffix += format
	log.Printf(preffix, v...)
}

func LogError(format string, v ...any) {
	Log(ERR, format, v...)
}

func LogFatal(format string, v ...any) {
	Log(ERR, format, v...)
	os.Exit(1)
}

func LogWarn(format string, v ...any) {
	Log(WARN, format, v...)
}

func LogDebug(format string, v ...any) {
	Log(DEBUG, format, v...)
}

func LogInfo(format string, v ...any) {
	Log(INFO, format, v...)
}
