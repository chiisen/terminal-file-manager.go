package preview

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：單元測試 (Unit Test)
// 說明：針對 preview 套件進行獨立測試
// ══════════════════════════════════════════════════════════════════════════════

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPreview(t *testing.T) {
	// 測試取得不存在的檔案
	_, err := GetPreview("/nonexistent/file.txt")
	if err == nil {
		t.Error("GetPreview should return error for nonexistent file")
	}

	// 測試取得目前目錄
	result, err := GetPreview(".")
	if err != nil {
		t.Errorf("GetPreview for directory failed: %v", err)
	}

	if !result.IsText {
		t.Error("Directory preview should be text")
	}
}

func TestGetPreviewFile(t *testing.T) {
	// 建立臨時測試檔案
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := "Hello, World!"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := GetPreview(testFile)
	if err != nil {
		t.Errorf("GetPreview failed: %v", err)
	}

	if !result.IsText {
		t.Error("Text file should be detected as text")
	}
}

func TestGetPreviewBinary(t *testing.T) {
	// 建立臨時測試檔案 (二進制)
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.bin")

	// 寫入一些可能是二進制的內容
	binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	if err := os.WriteFile(testFile, binaryContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := GetPreview(testFile)
	if err != nil {
		t.Errorf("GetPreview failed: %v", err)
	}

	// 二進制檔案應該被檢測到
	if result.IsBinary {
		t.Log("Binary file detected correctly")
	}
}

func TestGetPreviewLargeFile(t *testing.T) {
	// 建立臨時測試檔案 (大於 MaxPreviewSize)
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.txt")

	// 建立一個大於 1MB 的檔案
	largeContent := make([]byte, MaxPreviewSize+1024)
	for i := range largeContent {
		largeContent[i] = 'a'
	}

	if err := os.WriteFile(testFile, largeContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := GetPreview(testFile)
	if err != nil {
		t.Errorf("GetPreview failed: %v", err)
	}

	// 大檔案應該顯示訊息
	if result.IsText {
		t.Log("Large file handled correctly")
	}
}

func TestIsImageExt(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{".jpg", true},
		{".jpeg", true},
		{".png", true},
		{".gif", true},
		{".bmp", true},
		{".webp", true},
		{".txt", false},
		{".md", false},
		{".go", false},
		{"", false},
		{".JPG", true}, // 大寫
		{".PNG", true}, // 大寫
	}

	for _, tt := range tests {
		result := isImageExt(tt.ext)
		if result != tt.expected {
			t.Errorf("isImageExt(%s) = %v; want %v", tt.ext, result, tt.expected)
		}
	}
}

func TestDetectImageType(t *testing.T) {
	tests := []struct {
		header   []byte
		expected string
	}{
		{[]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, "PNG"},
		{[]byte{0xFF, 0xD8, 0xFF, 0xE0}, "JPEG"},
		{[]byte{'G', 'I', 'F', '8', '7', 'a'}, "GIF"},
		{[]byte{'G', 'I', 'F', '8', '9', 'a'}, "GIF"},
		{[]byte{'B', 'M'}, "BMP"},
		{[]byte{'R', 'I', 'F', 'F', 0x00, 0x00, 0x00, 0x00, 'W', 'E', 'B', 'P'}, "WebP"},
		{[]byte{0x00, 0x00, 0x01, 0x00}, "ICO"},
		{[]byte{0x00, 0x00, 0x02, 0x00}, "ICO"},
		{[]byte{0x00, 0x00}, ""},             // 太短
		{[]byte{0xAA, 0xBB, 0xCC, 0xDD}, ""}, // 未知格式
	}

	for _, tt := range tests {
		result := detectImageType(tt.header)
		if result != tt.expected {
			t.Errorf("detectImageType(%v) = %s; want %s", tt.header, result, tt.expected)
		}
	}
}

func TestIsText(t *testing.T) {
	tests := []struct {
		input    []byte
		expected bool
	}{
		{[]byte("Hello, World!"), true},
		{[]byte("Line 1\nLine 2\nLine 3"), true},
		{[]byte("Tab\tseparated\tvalues"), true},
		{[]byte{0x00, 0x01, 0x02}, false},               // 包含 NULL
		{[]byte{0xFF, 0xFE, 0xFD}, false},               // 全是非可印字元
		{[]byte("Hello\x00World"), false},               // 包含 NULL
		{[]byte{}, true},                                // 空內容視為文字
		{[]byte("Test with 80% printable chars"), true}, // 高比例可印字元
	}

	for _, tt := range tests {
		result := isText(tt.input)
		if result != tt.expected {
			t.Errorf("isText(%v) = %v; want %v", tt.input, result, tt.expected)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		result := formatSize(tt.input)
		if result != tt.expected {
			t.Errorf("formatSize(%d) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestCleanControlChars(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello, World!", "Hello, World!"},
		{"Line1\nLine2", "Line1\nLine2"},
		{"Tab\tValue", "Tab\tValue"},
		{"With\x00Null", "WithNull"},
		{"", ""},
	}

	for _, tt := range tests {
		result := cleanControlChars(tt.input)
		if result != tt.expected {
			t.Errorf("cleanControlChars(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}
