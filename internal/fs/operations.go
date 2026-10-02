package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：File Operations
// 說明：負責處理檔案的各種操作，包括複製、貼上、刪除、重新命名等
// 為何使用：將檔案操作邏輯集中管理，便於重用和測試
// ══════════════════════════════════════════════════════════════════════════════

// DeleteFile 刪除指定的檔案或目錄
// 如果是目錄，會遞迴刪除所有內容
func DeleteFile(path string) error {
	// 檢查是否存在
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("檔案不存在: %w", err)
	}

	if info.IsDir() {
		// 遞迴刪除目錄
		return os.RemoveAll(path)
	}

	// 刪除檔案
	return os.Remove(path)
}

// RenameFile 重新命名檔案或目錄
// 參數 oldPath 是原始路徑
// 參數 newName 是新的名稱（不包含路徑）
func RenameFile(oldPath, newName string) error {
	// 取得父目錄
	parent := filepath.Dir(oldPath)

	// 組合新路徑
	newPath := filepath.Join(parent, newName)

	// 檢查新路徑是否已存在
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("目標名稱已存在: %s", newName)
	}

	// 執行重新命名
	return os.Rename(oldPath, newPath)
}

// CopyFile 複製檔案
// 參數 src 是來源檔案路徑
// 參數 dst 是目標檔案路徑
func CopyFile(src, dst string) error {
	sourceInfo, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("無法讀取來源: %w", err)
	}
	// 符號連結保留為連結，避免遞迴追蹤目錄連結形成循環。
	if sourceInfo.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dst)
	}
	if err := validateCopyPaths(src, dst, sourceInfo); err != nil {
		return err
	}
	// 開啟來源檔案
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("無法開啟來源檔案: %w", err)
	}
	defer sourceFile.Close()

	// 取得來源檔案資訊
	sourceInfo, err = sourceFile.Stat()
	if err != nil {
		return err
	}

	// 檢查是否是目錄
	if sourceInfo.IsDir() {
		return copyDirectory(src, dst)
	}

	// 建立目標檔案
	destFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("無法建立目標檔案: %w", err)
	}
	defer destFile.Close()

	// 複製內容
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("複製失敗: %w", err)
	}

	return nil
}

// 寫入前解析實體路徑，阻擋自身、硬連結及透過連結指向的子目錄。
func validateCopyPaths(src, dst string, sourceInfo os.FileInfo) error {
	srcPath, err := filepath.EvalSymlinks(src)
	if err != nil {
		return err
	}
	srcPath, err = filepath.Abs(srcPath)
	if err != nil {
		return err
	}
	dstPath, err := resolvedDestination(dst)
	if err != nil {
		return err
	}
	if dstInfo, err := os.Stat(dst); err == nil {
		if os.SameFile(sourceInfo, dstInfo) {
			return fmt.Errorf("不可複製到來源本身: %s", dst)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	rel, err := filepath.Rel(srcPath, dstPath)
	// 不同磁碟沒有相對路徑，也不會是來源的子目錄。
	if err == nil && (rel == "." || (sourceInfo.IsDir() && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel))) {
		return fmt.Errorf("不可複製到來源本身或其子目錄: %s", dst)
	}
	return nil
}

// 目的地可能尚未存在：解析最接近的既有祖先，再接回未建立的部分。
func resolvedDestination(dst string) (string, error) {
	absPath, err := filepath.Abs(dst)
	if err != nil {
		return "", err
	}
	current := absPath
	var missing []string
	for {
		_, err := os.Lstat(current)
		if err == nil {
			resolved, err := filepath.EvalSymlinks(current)
			if err != nil {
				return "", err
			}
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return resolved, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", err
		}
		missing = append(missing, filepath.Base(current))
		current = parent
	}
}

// copyDirectory 遞迴複製目錄
func copyDirectory(src, dst string) error {
	// 取得來源目錄的資訊
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 建立目標目錄
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	// 讀取來源目錄內容
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// 遞迴複製每個項目
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// 子目錄同樣經過路徑防護，避免既有目的地連結指回來源。
		err = CopyFile(srcPath, dstPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateFile 創建新檔案
// 參數 path 是要創建的檔案路徑
func CreateFile(path string) error {
	// 檢查是否已存在
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("檔案已存在: %s", path)
	}

	// 創建檔案
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("無法創建檔案: %w", err)
	}
	defer file.Close()

	return nil
}

// CreateDirectory 創建新目錄
// 參數 path 是要創建的目錄路徑
func CreateDirectory(path string) error {
	// 檢查是否已存在
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("目錄已存在: %s", path)
	}

	// 創建目錄
	return os.MkdirAll(path, 0755)
}

// MoveFile 移動檔案（或重新命名）
// 這是 RenameFile 的包裝，提供更直觀的名稱
func MoveFile(src, dst string) error {
	return os.Rename(src, dst)
}
