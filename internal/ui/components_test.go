package ui

import (
	"strings"
	"testing"
	"time"

	"gofm/internal/types"
)

// 💡 概念：UI 純函數測試
// 說明：Format/Render 系列不依賴終端機，適合單元測試鎖定排版行為
// 為何使用：防止截斷、空目錄等邊界條件迴歸

func TestFormatSize(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}
	for _, tt := range tests {
		if got := FormatSize(tt.input); got != tt.want {
			t.Errorf("FormatSize(%d) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatTime(t *testing.T) {
	now := time.Now()
	// 用今天中午測試（離午夜邊界最遠，避免跨日 flake）
	noon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	got := FormatTime(noon)
	if got != "12:00" {
		t.Errorf("今天的檔案 FormatTime = %q; want %q", got, "12:00")
	}

	// 去年檔案顯示 YYYY/MM/DD
	old := time.Date(2020, 5, 6, 7, 8, 9, 0, time.Local)
	if got := FormatTime(old); got != "2020/05/06" {
		t.Errorf("舊檔案 FormatTime = %q; want %q", got, "2020/05/06")
	}
}

func TestRenderFileListEmpty(t *testing.T) {
	got := RenderFileList(nil, 0, map[string]bool{}, 80, 24)
	if got != "(empty directory)" {
		t.Errorf("空目錄渲染 = %q; want %q", got, "(empty directory)")
	}
}

func TestRenderFileListBasic(t *testing.T) {
	entries := []types.FileEntry{
		{Name: "dir1", Path: "/tmp/dir1", IsDir: true, ModTime: time.Now()},
		{Name: "a.txt", Path: "/tmp/a.txt", Size: 100, ModTime: time.Now()},
	}
	got := RenderFileList(entries, 0, map[string]bool{}, 80, 24)
	if !strings.Contains(got, "dir1") || !strings.Contains(got, "a.txt") {
		t.Errorf("列表應包含檔名，got:\n%s", got)
	}
	if !strings.Contains(got, "▶") {
		t.Errorf("游標行應有 ▶ 標記，got:\n%s", got)
	}
}

func TestRenderPathBar(t *testing.T) {
	// 短路徑直接顯示
	if got := RenderPathBar("/tmp", 80); !strings.Contains(got, "/tmp") {
		t.Errorf("短路徑應完整顯示，got %q", got)
	}
	// 超長路徑截斷並保留前綴
	long := "/very/long/path/that/definitely/exceeds/width/limit/here"
	got := RenderPathBar(long, 20)
	if !strings.Contains(got, "PATH:") {
		t.Errorf("截斷後仍應有 PATH: 前綴，got %q", got)
	}
}

func TestRenderStatusBar(t *testing.T) {
	if got := RenderStatusBar("hello", 80); !strings.Contains(got, "hello") {
		t.Errorf("狀態列應包含訊息，got %q", got)
	}
	// 超長訊息截斷為 ... 結尾（去除 ANSI 後檢查）
	got := RenderStatusBar(strings.Repeat("x", 100), 20)
	if !strings.Contains(got, "...") {
		t.Errorf("超長訊息應截斷，got %q", got)
	}
}

func TestRenderErrorAndStatus(t *testing.T) {
	if got := RenderError(""); got != "" {
		t.Errorf("空錯誤應回傳空字串，got %q", got)
	}
	if got := RenderError("boom"); !strings.Contains(got, "boom") {
		t.Errorf("錯誤訊息應包含原文，got %q", got)
	}
	if got := RenderStatusMessage(""); got != "" {
		t.Errorf("空狀態應回傳空字串，got %q", got)
	}
	if got := RenderStatusMessage("ok"); !strings.Contains(got, "ok") {
		t.Errorf("狀態訊息應包含原文，got %q", got)
	}
}
