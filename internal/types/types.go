package types

import "time"

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：共用類型定義
// 說明：存放應用程式中共用的資料結構定義
// 為何使用：避免套件間的循環依賴問題
// ══════════════════════════════════════════════════════════════════════════════

// FileEntry 代表目錄中的單個檔案或資料夾
type FileEntry struct {
	Name        string    // 檔案名稱
	Path        string    // 完整路徑
	Size        int64     // 檔案大小（位元組）
	ModTime     time.Time // 檔案修改時間
	IsDir       bool      // 是否為目錄
	Mode        string    // 權限模式（如 drwxr-xr-x）
	IsSymlink   bool      // 是否為符號連結
	SymlinkPath string    // 符號連結指向的路徑（如果存在）
	IsBroken    bool      // 是否為損壞的符號連結
	Permission  string    // 權限字串（如 "rwxr-xr-x"）
}

// ErrorType 代表錯誤類型
type ErrorType int

const (
	ErrorNone ErrorType = iota
	ErrorPermissionDenied
	ErrorNotFound
	ErrorBrokenSymlink
	ErrorOther
)

// FileError 代表檔案操作錯誤
type FileError struct {
	Type    ErrorType // 錯誤類型
	Message string    // 錯誤訊息
	Path    string    // 發生錯誤的路徑
}

// Error 回傳錯誤訊息（實現 error 介面）
func (e *FileError) Error() string {
	return e.Message
}
