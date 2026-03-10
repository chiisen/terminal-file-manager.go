package preview

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：Preview System
// 說明：負責產生檔案的預覽內容，包括文字、圖片和二進制檔案
// 為何使用：讓使用者在不開啟檔案的情況下快速查看檔案內容
// ══════════════════════════════════════════════════════════════════════════════

const (
	// MaxPreviewSize 預覽的最大檔案大小（1MB）
	MaxPreviewSize = 1024 * 1024
)

// PreviewResult 預覽結果
type PreviewResult struct {
	Content   string // 預覽內容
	IsText    bool   // 是否為文字檔
	IsBinary  bool   // 是否為二進制檔
	IsImage   bool   // 是否為圖片檔
	ImageInfo string // 圖片資訊
}

// GetPreview 取得檔案的預覽
// 參數 path 是檔案路徑
// 回傳 PreviewResult
func GetPreview(path string) (*PreviewResult, error) {
	// 取得檔案資訊
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// 檢查是否為目錄
	if info.IsDir() {
		return &PreviewResult{
			Content: fmt.Sprintf("[Directory]\n%s", filepath.Base(path)),
			IsText:  true,
		}, nil
	}

	// 檢查檔案大小
	if info.Size() > MaxPreviewSize {
		return &PreviewResult{
			Content: fmt.Sprintf("[File too large]\nSize: %d bytes\nMax preview: %d bytes", info.Size(), MaxPreviewSize),
			IsText:  true,
		}, nil
	}

	// 檢查副檔名
	ext := filepath.Ext(path)
	if isImageExt(ext) {
		return getImagePreview(path, info)
	}

	// 嘗試讀取為文字檔
	return getTextPreview(path, info)
}

// isImageExt 檢查是否為圖片副檔名
func isImageExt(ext string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".ico", ".tiff"}
	ext = toLower(ext)
	for _, e := range imageExts {
		if ext == e {
			return true
		}
	}
	return false
}

// getImagePreview 取得圖片預覽
func getImagePreview(path string, info os.FileInfo) (*PreviewResult, error) {
	// 讀取檔案前幾個位元組來驗證圖片格式
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 讀取檔頭
	header := make([]byte, 32)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return nil, err
	}
	header = header[:n]

	// 驗證圖片格式
	imageType := detectImageType(header)
	if imageType == "" {
		return &PreviewResult{
			Content:  "[Unknown image format]",
			IsText:   true,
			IsImage:  false,
			IsBinary: true,
		}, nil
	}

	return &PreviewResult{
		Content: fmt.Sprintf("[%s Image]\nFile: %s\nSize: %s\nModified: %s",
			imageType,
			info.Name(),
			formatSize(info.Size()),
			info.ModTime().Format("2006-01-02 15:04:05")),
		IsText:    true,
		IsImage:   true,
		ImageInfo: fmt.Sprintf("%s - %s", imageType, formatSize(info.Size())),
	}, nil
}

// detectImageType 偵測圖片類型
func detectImageType(header []byte) string {
	if len(header) < 2 {
		return ""
	}

	// BMP (只需要 2 bytes) - 必須最前面檢查
	if header[0] == 'B' && header[1] == 'M' {
		return "BMP"
	}

	// PNG (需要 8 bytes)
	if len(header) >= 8 && bytes.Equal(header[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return "PNG"
	}

	// JPEG
	if len(header) >= 3 && header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
		return "JPEG"
	}

	// GIF (需要 6 bytes)
	if len(header) >= 6 && (bytes.Equal(header[:6], []byte("GIF87a")) || bytes.Equal(header[:6], []byte("GIF89a"))) {
		return "GIF"
	}

	// WebP
	if len(header) >= 12 && bytes.Equal(header[:4], []byte("RIFF")) && bytes.Equal(header[8:12], []byte("WEBP")) {
		return "WebP"
	}

	// ICO
	if len(header) >= 4 && (bytes.Equal(header[:4], []byte{0x00, 0x00, 0x01, 0x00}) || bytes.Equal(header[:4], []byte{0x00, 0x00, 0x02, 0x00})) {
		return "ICO"
	}

	return ""
}

// getTextPreview 取得文字檔預覽
func getTextPreview(path string, info os.FileInfo) (*PreviewResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 讀取檔案內容（最多 4KB）
	buf := make([]byte, 4096)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	buf = buf[:n]

	// 檢查是否為文字檔
	if !isText(buf) {
		return &PreviewResult{
			Content: fmt.Sprintf("[Binary File]\nFile: %s\nSize: %s\nType: %s",
				info.Name(),
				formatSize(info.Size()),
				filepath.Ext(path)),
			IsText:   true,
			IsBinary: true,
		}, nil
	}

	// 讀取完整內容用於預覽（最多 64KB）
	previewSize := int64(65536)
	if info.Size() < previewSize {
		previewSize = info.Size()
	}

	previewBuf := make([]byte, previewSize)
	file.Seek(0, 0)
	n, err = file.Read(previewBuf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	previewBuf = previewBuf[:n]

	// 移除控制字元
	content := cleanControlChars(string(previewBuf))

	return &PreviewResult{
		Content: fmt.Sprintf("[Text File]\nFile: %s\nSize: %s\n\n%s",
			info.Name(),
			formatSize(info.Size()),
			content),
		IsText: true,
	}, nil
}

// isText 檢查是否為文字檔
func isText(buf []byte) bool {
	if len(buf) == 0 {
		return true
	}

	// 檢查是否包含 NULL 字元
	for _, b := range buf {
		if b == 0 {
			return false
		}
	}

	// 計算可列印字元的比例
	printable := 0
	for _, b := range buf {
		if (b >= 32 && b <= 126) || b == '\n' || b == '\r' || b == '\t' {
			printable++
		}
	}

	return float64(printable)/float64(len(buf)) > 0.8
}

// cleanControlChars 移除控制字元
func cleanControlChars(s string) string {
	result := make([]byte, 0, len(s))
	for _, c := range s {
		if c >= 32 || c == '\n' || c == '\r' || c == '\t' {
			result = append(result, byte(c))
		}
	}
	return string(result)
}

// toLower 轉換字串為小寫
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}

// formatSize 格式化檔案大小
func formatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
}
