package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：Logger (日誌記錄)
// 說明：提供簡單的日誌記錄功能，輸出到 ~/.config/gofm/log.txt
// 為何使用：記錄應用程式的執行狀態和錯誤，方便 Debug
// ══════════════════════════════════════════════════════════════════════════════

// fileLogger 結構
type fileLogger struct {
	file *os.File
}

// LogLevel 日誌等級
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// 全域日誌實例
var defaultLogger *fileLogger

// Init 初始化日誌系統
func Init() error {
	// 建立配置目錄
	home := os.Getenv("HOME")
	configDir := filepath.Join(home, ".config", "gofm")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// 建立日誌檔案
	logPath := filepath.Join(configDir, "log.txt")
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defaultLogger = &fileLogger{file: file}

	// 寫入啟動訊息
	Info("gofm started")

	return nil
}

// Close 關閉日誌檔案
func Close() {
	if defaultLogger != nil && defaultLogger.file != nil {
		Info("gofm closed")
		defaultLogger.file.Close()
	}
}

// write 寫入日誌
func (l *fileLogger) write(level LogLevel, format string, args ...interface{}) {
	if l == nil || l.file == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	levelStr := ""
	switch level {
	case LevelDebug:
		levelStr = "DEBUG"
	case LevelInfo:
		levelStr = "INFO"
	case LevelWarn:
		levelStr = "WARN"
	case LevelError:
		levelStr = "ERROR"
	}

	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, levelStr, message)

	l.file.WriteString(logLine)
}

// Debug 寫入 Debug 等級日誌
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.write(LevelDebug, format, args...)
	}
}

// Info 寫入 Info 等級日誌
func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.write(LevelInfo, format, args...)
	}
}

// Warn 寫入 Warn 等級日誌
func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.write(LevelWarn, format, args...)
	}
}

// Error 寫入 Error 等級日誌
func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.write(LevelError, format, args...)
	}
}

// GetLogPath 回傳日誌檔案路徑
func GetLogPath() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config", "gofm", "log.txt")
}
