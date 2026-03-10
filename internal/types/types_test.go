package types

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：單元測試 (Unit Test)
// 說明：針對 types 套件進行獨立測試
// ══════════════════════════════════════════════════════════════════════════════

import (
	"testing"
)

func TestFileEntry(t *testing.T) {
	entry := FileEntry{
		Name:        "test.txt",
		Path:        "/home/user/test.txt",
		Size:        1024,
		IsDir:       false,
		Mode:        "-rw-r--r--",
		IsSymlink:   false,
		SymlinkPath: "",
		IsBroken:    false,
		Permission:  "rw-r--r--",
	}

	if entry.Name != "test.txt" {
		t.Errorf("Name = %s; want test.txt", entry.Name)
	}

	if entry.Size != 1024 {
		t.Errorf("Size = %d; want 1024", entry.Size)
	}

	if entry.IsDir {
		t.Error("IsDir should be false")
	}
}

func TestFileError(t *testing.T) {
	err := &FileError{
		Type:    ErrorPermissionDenied,
		Message: "Permission denied",
		Path:    "/home/user/test.txt",
	}

	// 測試 Error 方法（實現 error 介面）
	if err.Error() != "Permission denied" {
		t.Errorf("Error() = %s; want Permission denied", err.Error())
	}
}

func TestErrorType(t *testing.T) {
	tests := []struct {
		value    ErrorType
		expected string
	}{
		{ErrorNone, "ErrorNone"},
		{ErrorPermissionDenied, "ErrorPermissionDenied"},
		{ErrorNotFound, "ErrorNotFound"},
		{ErrorBrokenSymlink, "ErrorBrokenSymlink"},
		{ErrorOther, "ErrorOther"},
	}

	for _, tt := range tests {
		// 簡單的類型測試
		var et ErrorType = tt.value
		_ = et
	}
}

func TestErrorTypeValues(t *testing.T) {
	// 測試 ErrorType 的 iota 值
	if ErrorNone != 0 {
		t.Errorf("ErrorNone = %d; want 0", ErrorNone)
	}

	if ErrorPermissionDenied != 1 {
		t.Errorf("ErrorPermissionDenied = %d; want 1", ErrorPermissionDenied)
	}

	if ErrorNotFound != 2 {
		t.Errorf("ErrorNotFound = %d; want 2", ErrorNotFound)
	}

	if ErrorBrokenSymlink != 3 {
		t.Errorf("ErrorBrokenSymlink = %d; want 3", ErrorBrokenSymlink)
	}

	if ErrorOther != 4 {
		t.Errorf("ErrorOther = %d; want 4", ErrorOther)
	}
}
