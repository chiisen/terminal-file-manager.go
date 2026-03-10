package fs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gofm/internal/types"
)

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：FileSystem Layer
// 說明：負責所有檔案系統相關的操作，如讀取目錄、創建檔案、複製、刪除等
// 為何使用：將檔案系統操作封裝起來，方便測試和維護
// ══════════════════════════════════════════════════════════════════════════════

// LazyReadDirectory 快速讀取目錄（只讀取名稱，不讀取詳細資訊）
// 這是效能優化的關鍵：先快速顯示目錄內容，再非同步載入詳細資訊
func LazyReadDirectory(path string) ([]types.FileEntry, error) {
	// 解析絕對路徑
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// 讀取目錄內容
	entries, err := os.ReadDir(absPath)
	if err != nil {
		// 檢查權限錯誤
		if os.IsPermission(err) {
			return nil, &types.FileError{
				Type:    types.ErrorPermissionDenied,
				Message: "Permission denied: " + absPath,
				Path:    absPath,
			}
		}
		return nil, err
	}

	// 只讀取名稱，快速返回
	result := make([]types.FileEntry, 0, len(entries))
	for _, entry := range entries {
		var size int64
		var modTime time.Time
		// 快速取得檔案大小 (在大多數作業系統中 DirEntry 已經快取了這項資訊，成本極低)
		if info, err := entry.Info(); err == nil {
			size = info.Size()
			modTime = info.ModTime()
		}

		result = append(result, types.FileEntry{
			Name:    entry.Name(),
			Path:    filepath.Join(absPath, entry.Name()),
			IsDir:   entry.IsDir(),
			Mode:    entry.Type().String(),
			Size:    size,
			ModTime: modTime,
		})
	}

	return result, nil
}

// ReadDirectory 讀取指定目錄的檔案列表
// 參數 path 是要讀取的目錄路徑
// 回傳 FileEntry 切片和錯誤資訊
func ReadDirectory(path string) ([]types.FileEntry, error) {
	// 解析絕對路徑
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// 讀取目錄內容
	entries, err := os.ReadDir(absPath)
	if err != nil {
		// 檢查權限錯誤
		if os.IsPermission(err) {
			return nil, &types.FileError{
				Type:    types.ErrorPermissionDenied,
				Message: "Permission denied: " + absPath,
				Path:    absPath,
			}
		}
		return nil, err
	}

	// 轉換為 FileEntry
	result := make([]types.FileEntry, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			// 可能是損壞的符號連結
			// 嘗試讀取符號連結的目標
			if entry.Type()&os.ModeSymlink != 0 {
				linkPath, linkErr := os.Readlink(filepath.Join(absPath, entry.Name()))
				result = append(result, types.FileEntry{
					Name:        entry.Name(),
					Path:        filepath.Join(absPath, entry.Name()),
					IsSymlink:   true,
					SymlinkPath: linkPath,
					IsBroken:    linkErr != nil,
					Mode:        entry.Type().String(),
				})
			}
			continue
		}

		// 檢查是否是符號連結
		isSymlink := entry.Type()&os.ModeSymlink != 0
		symlinkPath := ""
		isBroken := false

		if isSymlink {
			linkPath, linkErr := os.Readlink(filepath.Join(absPath, entry.Name()))
			symlinkPath = linkPath
			isBroken = linkErr != nil
		}

		result = append(result, types.FileEntry{
			Name:        entry.Name(),
			Path:        filepath.Join(absPath, entry.Name()),
			Size:        info.Size(),
			ModTime:     info.ModTime(),
			IsDir:       entry.IsDir(),
			Mode:        info.Mode().String(),
			IsSymlink:   isSymlink,
			SymlinkPath: symlinkPath,
			IsBroken:    isBroken,
			Permission:  formatPermission(info.Mode()),
		})
	}

	return result, nil
}

// formatPermission 將 os.FileMode 轉換為權限字串（如 rwxr-xr-x）
func formatPermission(mode os.FileMode) string {
	var perms strings.Builder

	// Owner permissions
	if mode&0400 != 0 {
		perms.WriteByte('r')
	} else {
		perms.WriteByte('-')
	}
	if mode&0200 != 0 {
		perms.WriteByte('w')
	} else {
		perms.WriteByte('-')
	}
	if mode&0100 != 0 {
		if mode&04000 != 0 {
			perms.WriteByte('s')
		} else {
			perms.WriteByte('x')
		}
	} else {
		if mode&04000 != 0 {
			perms.WriteByte('S')
		} else {
			perms.WriteByte('-')
		}
	}

	// Group permissions
	if mode&0040 != 0 {
		perms.WriteByte('r')
	} else {
		perms.WriteByte('-')
	}
	if mode&0020 != 0 {
		perms.WriteByte('w')
	} else {
		perms.WriteByte('-')
	}
	if mode&0010 != 0 {
		if mode&02000 != 0 {
			perms.WriteByte('s')
		} else {
			perms.WriteByte('x')
		}
	} else {
		if mode&02000 != 0 {
			perms.WriteByte('S')
		} else {
			perms.WriteByte('-')
		}
	}

	// Other permissions
	if mode&0004 != 0 {
		perms.WriteByte('r')
	} else {
		perms.WriteByte('-')
	}
	if mode&0002 != 0 {
		perms.WriteByte('w')
	} else {
		perms.WriteByte('-')
	}
	if mode&0001 != 0 {
		if mode&01000 != 0 {
			perms.WriteByte('t')
		} else {
			perms.WriteByte('x')
		}
	} else {
		if mode&01000 != 0 {
			perms.WriteByte('T')
		} else {
			perms.WriteByte('-')
		}
	}

	return perms.String()
}

// GetParentDirectory 回傳指定目錄的父目錄
func GetParentDirectory(path string) string {
	return filepath.Dir(path)
}

// GetAbsolutePath 回傳指定路徑的絕對路徑
func GetAbsolutePath(path string) (string, error) {
	// 如果路徑為空，使用目前目錄
	if path == "" {
		return os.Getwd()
	}

	// 解析為絕對路徑
	return filepath.Abs(path)
}

// SortEntries 根據指定的排序方式對檔案列表進行排序
// 參數 entries 是要排序的檔案列表
// 參數 sortBy 是排序方式："name", "size", "modified", "type"
func SortEntries(entries []types.FileEntry, sortBy string) {
	switch sortBy {
	case "name":
		sort.Slice(entries, func(i, j int) bool {
			// 目錄優先於檔案
			if entries[i].IsDir != entries[j].IsDir {
				return entries[i].IsDir
			}
			return entries[i].Name < entries[j].Name
		})
	case "size":
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].IsDir != entries[j].IsDir {
				return entries[i].IsDir
			}
			return entries[i].Size < entries[j].Size
		})
	case "modified":
		// 需要讀取修改時間，這裡簡化處理
		fallthrough
	default:
		// 預設按名稱排序
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].IsDir != entries[j].IsDir {
				return entries[i].IsDir
			}
			return entries[i].Name < entries[j].Name
		})
	}
}

// FileExists 檢查指定路徑的檔案或目錄是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetFileInfo 取得檔案的詳細資訊
func GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// GetModTime 取得檔案的修改時間
// 這是一個輔助函數，用於取得時間格式化後的字串
func GetModTime(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}
	return info.ModTime().Format(time.RFC3339)
}
