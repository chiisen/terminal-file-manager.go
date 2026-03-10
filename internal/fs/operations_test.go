package fs

import (
	"os"
	"path/filepath"
	"testing"
)

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：單元測試 (Unit Test)
// 說明：針對 operations.go 進行測試
// ══════════════════════════════════════════════════════════════════════════════

func TestDeleteFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 建立測試檔案
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 刪除檔案
	err := DeleteFile(testFile)
	if err != nil {
		t.Errorf("DeleteFile failed: %v", err)
	}

	// 確認檔案已刪除
	if _, err := os.Stat(testFile); err == nil {
		t.Error("File should be deleted")
	}

	// 測試刪除不存在的檔案
	err = DeleteFile("/nonexistent/file")
	if err == nil {
		t.Error("DeleteFile should return error for nonexistent file")
	}
}

func TestDeleteDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// 建立測試目錄
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	// 刪除目錄
	err := DeleteFile(testDir)
	if err != nil {
		t.Errorf("DeleteFile failed: %v", err)
	}

	// 確認目錄已刪除
	if _, err := os.Stat(testDir); err == nil {
		t.Error("Directory should be deleted")
	}
}

func TestRenameFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 建立測試檔案
	testFile := filepath.Join(tmpDir, "old.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 重新命名
	err := RenameFile(testFile, "new.txt")
	if err != nil {
		t.Errorf("RenameFile failed: %v", err)
	}

	// 確認新檔案存在
	newPath := filepath.Join(tmpDir, "new.txt")
	if _, err := os.Stat(newPath); err != nil {
		t.Error("New file should exist")
	}

	// 確認舊檔案不存在
	if _, err := os.Stat(testFile); err == nil {
		t.Error("Old file should not exist")
	}

	// 測試重新命名到已存在的檔案
	err = RenameFile(newPath, "new.txt")
	if err == nil {
		t.Error("RenameFile should return error when target exists")
	}
}

func TestCreateFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 建立新檔案
	newFile := filepath.Join(tmpDir, "newfile.txt")
	err := CreateFile(newFile)
	if err != nil {
		t.Errorf("CreateFile failed: %v", err)
	}

	// 確認檔案存在
	if _, err := os.Stat(newFile); err != nil {
		t.Error("File should exist")
	}

	// 測試建立已存在的檔案
	err = CreateFile(newFile)
	if err == nil {
		t.Error("CreateFile should return error when file exists")
	}
}

func TestCreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// 建立新目錄
	newDir := filepath.Join(tmpDir, "newdir")
	err := CreateDirectory(newDir)
	if err != nil {
		t.Errorf("CreateDirectory failed: %v", err)
	}

	// 確認目錄存在
	if _, err := os.Stat(newDir); err != nil {
		t.Error("Directory should exist")
	}

	// 測試建立已存在的目錄
	err = CreateDirectory(newDir)
	if err == nil {
		t.Error("CreateDirectory should return error when directory exists")
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 建立測試檔案
	srcFile := filepath.Join(tmpDir, "source.txt")
	content := []byte("test content")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 複製檔案
	dstFile := filepath.Join(tmpDir, "dest.txt")
	err := CopyFile(srcFile, dstFile)
	if err != nil {
		t.Errorf("CopyFile failed: %v", err)
	}

	// 確認目的檔案存在
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Error("Destination file should exist")
	}

	// 確認內容相同
	if string(dstContent) != string(content) {
		t.Error("Destination content should match source")
	}

	// 確認來源檔案仍然存在
	if _, err := os.Stat(srcFile); err != nil {
		t.Error("Source file should still exist")
	}
}

func TestCopyFileToExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// 建立兩個檔案
	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")

	os.WriteFile(srcFile, []byte("source"), 0644)
	os.WriteFile(dstFile, []byte("dest"), 0644)

	// 複製到已存在的檔案（應該成功，會覆蓋）
	err := CopyFile(srcFile, dstFile)
	if err != nil {
		t.Errorf("CopyFile should succeed even when destination exists: %v", err)
	}

	// 確認內容已被覆蓋
	content, _ := os.ReadFile(dstFile)
	if string(content) != "source" {
		t.Error("Destination should be overwritten with source content")
	}
}

func TestMoveFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 建立測試檔案
	srcFile := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 移動檔案
	dstFile := filepath.Join(tmpDir, "dest.txt")
	err := MoveFile(srcFile, dstFile)
	if err != nil {
		t.Errorf("MoveFile failed: %v", err)
	}

	// 確認目的檔案存在
	if _, err := os.Stat(dstFile); err != nil {
		t.Error("Destination file should exist")
	}

	// 確認來源檔案不存在
	if _, err := os.Stat(srcFile); err == nil {
		t.Error("Source file should not exist after move")
	}
}

func TestCopyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// 建立測試目錄結構
	srcDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// 在來源目錄中建立檔案
	if err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file in source dir: %v", err)
	}

	// 複製目錄
	dstDir := filepath.Join(tmpDir, "dest")
	err := CopyFile(srcDir, dstDir)
	if err != nil {
		t.Errorf("CopyFile for directory failed: %v", err)
	}

	// 確認目的目錄存在
	if _, err := os.Stat(dstDir); err != nil {
		t.Error("Destination directory should exist")
	}

	// 確認來源目錄仍然存在
	if _, err := os.Stat(srcDir); err != nil {
		t.Error("Source directory should still exist")
	}
}
