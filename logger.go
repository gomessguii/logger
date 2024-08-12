package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	INFO  = "INFO"
	ERR   = "ERR"
	WARN  = "WARN"
	DEBUG = "DEBUG"
)

type WebhookConfig struct {
	Url       string
	SendError bool
	SendFatal bool
	SendWarn  bool
}

type Logger struct {
	ServiceName          string
	LogContextName       string
	CaptureExceptionFunc func(err error)
	WebhookConfig        WebhookConfig
}

func (l *Logger) Log(logLevel string, format string, v ...any) {
	if logLevel == DEBUG && os.Getenv("DEBUG_ENABLED") != "1" {
		return
	}

	var prefix string
	switch logLevel {
	case ERR:
		prefix += "\033[41m[ERR]\033[0m "
	case WARN:
		prefix += "\033[43m[WARN]\033[0m "
	case DEBUG:
		prefix += "\033[40m\033[37m[DEBUG]\033[0m "
	default:
		prefix += "\033[44m[INFO]\033[0m "
	}
	servicePrefix := fmt.Sprintf("\033[35m[%s]\033[0m ", l.ServiceName)
	prefix = servicePrefix + prefix + format
	log.Printf(prefix, v...)
}

func (l *Logger) LogInfo(format string, v ...any) {
	l.Log(INFO, format, v...)
}

func (l *Logger) LogError(format string, v ...any) {
	if l.CaptureExceptionFunc != nil {
		l.CaptureExceptionFunc(fmt.Errorf(fmt.Sprintf("{%s} => %s", l.LogContextName, fmt.Sprintf(format, v...))))
	}
	l.Log(ERR, format, v...)
	if l.WebhookConfig.SendError {
		l.sendWebhook(ERR, format, v...)
	}
}

func (l *Logger) LogFatal(format string, v ...any) {
	if l.CaptureExceptionFunc != nil {
		l.CaptureExceptionFunc(fmt.Errorf(fmt.Sprintf("{%s} => %s", l.LogContextName, fmt.Sprintf(format, v...))))
	}
	l.Log(ERR, format, v...)
	if l.WebhookConfig.SendFatal {
		l.sendWebhook(ERR, format, v...)
	}
	os.Exit(1)
}

func (l *Logger) LogWarn(format string, v ...any) {
	l.Log(WARN, format, v...)
	if l.WebhookConfig.SendWarn {
		l.sendWebhook(WARN, format, v...)
	}
}

func (l *Logger) LogDebug(format string, v ...any) {
	l.Log(DEBUG, format, v...)
}

func (l *Logger) sendWebhook(logLevel string, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	timestamp := time.Now().Format(time.RFC3339)

	payload := struct {
		ServiceName    string `json:"serviceName"`
		LogContextName string `json:"logContextName"`
		Message        string `json:"message"`
		Level          string `json:"level"`
		Timestamp      string `json:"timestamp"`
	}{
		ServiceName:    l.ServiceName,
		LogContextName: l.LogContextName,
		Message:        message,
		Level:          logLevel,
		Timestamp:      timestamp,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		l.Log(ERR, "Failed to marshal webhook payload: %v\n", err)
		return
	}

	resp, err := http.Post(l.WebhookConfig.Url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		l.Log(ERR, "Failed to send webhook: %v\n", err)
		return
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		l.Log(ERR, "Webhook responded with status: %s\n", resp.Status)
	}
}
