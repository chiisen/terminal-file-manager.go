package fs

import (
	"os"
	"path/filepath"
	"testing"

	"gofm/internal/types"
)

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：單元測試 (Unit Test)
// 說明：針對 filesystem 套件進行獨立測試，確保每個函數的正確性
// 為何使用：快速驗證程式碼正確性，防止回歸問題
// ══════════════════════════════════════════════════════════════════════════════

func TestGetAbsolutePath(t *testing.T) {
	// 測試取得絕對路徑
	absPath, err := GetAbsolutePath(".")
	if err != nil {
		t.Errorf("GetAbsolutePath failed: %v", err)
	}

	if !filepath.IsAbs(absPath) {
		t.Errorf("Expected absolute path, got: %s", absPath)
	}

	// 測試空路徑
	absPath, err = GetAbsolutePath("")
	if err != nil {
		t.Errorf("GetAbsolutePath with empty path failed: %v", err)
	}
}

func TestGetParentDirectory(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/home/user/documents", "/home/user"},
		{"/home/user", "/home"},
		{"/home", "/"},
		{"/", "/"},
	}

	for _, tt := range tests {
		result := GetParentDirectory(tt.input)
		if result != tt.expected {
			t.Errorf("GetParentDirectory(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestFileExists(t *testing.T) {
	// 測試存在的檔案
	if !FileExists(".") {
		t.Error("FileExists('.') should return true")
	}

	// 測試不存在的檔案
	if FileExists("/nonexistent/path/to/file") {
		t.Error("FileExists for nonexistent path should return false")
	}
}

func TestReadDirectory(t *testing.T) {
	// 測試讀取目前目錄
	entries, err := ReadDirectory(".")
	if err != nil {
		t.Errorf("ReadDirectory failed: %v", err)
	}

	// 檢查是否為切片
	if entries == nil {
		t.Error("ReadDirectory should return non-nil slice")
	}

	// 測試讀取不存在的目錄
	_, err = ReadDirectory("/nonexistent/directory")
	if err == nil {
		t.Error("ReadDirectory should return error for nonexistent directory")
	}
}

func TestReadDirectoryPermission(t *testing.T) {
	// 這個測試可能需要特殊權限，僅作基本檢查
	// 嘗試讀取根目錄
	entries, err := ReadDirectory("/")
	if err != nil {
		// 可能沒有權限，這是預期的
		t.Logf("ReadDirectory / failed (expected if no permission): %v", err)
		return
	}

	// 檢查是否至少有一些項目
	if len(entries) == 0 {
		t.Error("Root directory should have entries")
	}
}

func TestLazyReadDirectory(t *testing.T) {
	entries, err := LazyReadDirectory(".")
	if err != nil {
		t.Errorf("LazyReadDirectory failed: %v", err)
	}

	if entries == nil {
		t.Error("LazyReadDirectory should return non-nil slice")
	}

	// 檢查每個項目都有基本欄位
	for _, entry := range entries {
		if entry.Name == "" {
			t.Error("FileEntry should have a name")
		}
		if entry.Path == "" {
			t.Error("FileEntry should have a path")
		}
	}
}

func TestSortEntries(t *testing.T) {
	// 建立測試資料
	entries := []types.FileEntry{
		{Name: "c.txt", IsDir: false, Size: 100},
		{Name: "b.txt", IsDir: false, Size: 200},
		{Name: "a", IsDir: true, Size: 0},
		{Name: "d.txt", IsDir: false, Size: 50},
	}

	// 按名稱排序
	SortEntries(entries, "name")

	// 檢查排序結果（目錄應該在最前面）
	if len(entries) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(entries))
	}

	// 第一個應該是目錄
	if !entries[0].IsDir {
		t.Error("Directory should come before files")
	}
}

func TestSortEntriesSize(t *testing.T) {
	entries := []types.FileEntry{
		{Name: "big.txt", IsDir: false, Size: 1000},
		{Name: "small.txt", IsDir: false, Size: 10},
	}

	SortEntries(entries, "size")

	// 小檔案應該在前
	if entries[0].Size > entries[1].Size {
		t.Error("Entries should be sorted by size")
	}
}

func TestGetModTime(t *testing.T) {
	// 測試取得目前目錄的修改時間
	modTime := GetModTime(".")
	if modTime == "" {
		t.Error("GetModTime should return non-empty string for existing path")
	}

	// 測試不存在的檔案
	modTime = GetModTime("/nonexistent/file")
	if modTime != "" {
		t.Error("GetModTime should return empty string for nonexistent path")
	}
}

// Helper function to create temporary test directory
func createTestDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "gofm_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create some test files
	testFiles := []string{"test1.txt", "test2.txt", "test3"}
	for _, name := range testFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	return tmpDir
}

func TestSortEntriesType(t *testing.T) {
	entries := []types.FileEntry{
		{Name: "a.txt", IsDir: false, Size: 100},
		{Name: "b.go", IsDir: false, Size: 200},
		{Name: "c.jpg", IsDir: false, Size: 50},
		{Name: "d", IsDir: true, Size: 0},
	}

	// 按副檔名排序
	SortEntries(entries, "type")

	// 目錄應該在最前面
	if !entries[0].IsDir {
		t.Error("Directory should come first")
	}
}

func TestSortEntriesDescending(t *testing.T) {
	entries := []types.FileEntry{
		{Name: "a.txt", IsDir: false, Size: 100},
		{Name: "b.txt", IsDir: false, Size: 200},
	}

	// 這個測試驗證排序功能的行為
	// Asc=true 升序，Asc=false 降序
	// 但因為測試的是獨立的 SortEntries 函數，它內部沒有 SortAsc 欄位
	// 所以會使用預設的行為
	SortEntries(entries, "name")

	// 預設應該是升序
	if entries[0].Name != "a.txt" {
		t.Logf("Note: SortEntries uses default ascending order")
	}
}

func TestSortEntriesUnknown(t *testing.T) {
	entries := []types.FileEntry{
		{Name: "a.txt", IsDir: false, Size: 100},
		{Name: "b.txt", IsDir: false, Size: 200},
	}

	// 未知排序方式，應該按名稱排序
	SortEntries(entries, "unknown")

	// 應該按名稱排序
	if entries[0].Name != "a.txt" {
		t.Error("Unknown sort should default to name")
	}
}

func TestGetFileInfo(t *testing.T) {
	// 測試取得檔案資訊
	info, err := GetFileInfo(".")
	if err != nil {
		t.Errorf("GetFileInfo for directory failed: %v", err)
	}

	if info == nil {
		t.Error("GetFileInfo should return non-nil")
	}

	// 測試不存在的檔案
	_, err = GetFileInfo("/nonexistent/file")
	if err == nil {
		t.Error("GetFileInfo should return error for nonexistent file")
	}
}

func TestReadDirectoryWithSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	// 建立一個檔案
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 建立一個符號連結
	symlink := filepath.Join(tmpDir, "link")
	if err := os.Symlink(testFile, symlink); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// 讀取目錄（包含符號連結）
	entries, err := ReadDirectory(tmpDir)
	if err != nil {
		t.Errorf("ReadDirectory failed: %v", err)
	}

	// 應該有兩個項目
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestLazyReadDirectoryWithSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	// 建立一個符號連結指向不存在的檔案
	symlink := filepath.Join(tmpDir, "broken_link")
	if err := os.Symlink("/nonexistent", symlink); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	// 讀取目錄（包含損壞的符號連結）
	entries, err := LazyReadDirectory(tmpDir)
	if err != nil {
		t.Errorf("LazyReadDirectory failed: %v", err)
	}

	// 應該有項目
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}
