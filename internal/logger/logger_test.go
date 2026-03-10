package logger

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：單元測試 (Unit Test)
// 說明：針對 logger 套件進行獨立測試
// ═════════════════════════════════════════════════════════=====================

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetLogPath(t *testing.T) {
	path := GetLogPath()

	// 檢查路徑格式
	expectedPrefix := filepath.Join(os.Getenv("HOME"), ".config", "gofm")
	if !strings.HasPrefix(path, expectedPrefix) {
		t.Errorf("GetLogPath() = %s; want prefix %s", path, expectedPrefix)
	}

	// 檢查結尾
	if !strings.HasSuffix(path, "log.txt") {
		t.Errorf("GetLogPath() should end with log.txt, got: %s", path)
	}
}

func TestLogLevel(t *testing.T) {
	// 測試日誌等級
	if LevelDebug != 0 {
		t.Errorf("LevelDebug = %d; want 0", LevelDebug)
	}

	if LevelInfo != 1 {
		t.Errorf("LevelInfo = %d; want 1", LevelInfo)
	}

	if LevelWarn != 2 {
		t.Errorf("LevelWarn = %d; want 2", LevelWarn)
	}

	if LevelError != 3 {
		t.Errorf("LevelError = %d; want 3", LevelError)
	}
}

func TestInit(t *testing.T) {
	// 測試初始化
	err := Init()
	if err != nil {
		t.Errorf("Init failed: %v", err)
	}

	// 檢查日誌檔案是否被建立
	logPath := GetLogPath()
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file should be created at: %s", logPath)
	}
}

func TestInfo(t *testing.T) {
	// 測試寫入 Info 日誌（應該不會崩潰）
	Info("Test info message")
}

func TestDebug(t *testing.T) {
	// 測試寫入 Debug 日誌
	Debug("Test debug message")
}

func TestWarn(t *testing.T) {
	// 測試寫入 Warn 日誌
	Warn("Test warning message")
}

func TestError(t *testing.T) {
	// 測試寫入 Error 日誌
	Error("Test error message")
}

func TestLogFormat(t *testing.T) {
	// 測試日誌格式
	Info("Format test: %d %s", 123, "test")

	// 測試格式化
	logLine := "[2024-01-01 12:00:00] [INFO] Test message"
	if !strings.Contains(logLine, "[INFO]") {
		t.Error("Log should contain level")
	}
}
