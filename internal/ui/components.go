package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"gofm/internal/types"
)

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：UI Components
// 說明：定義 TUI 的視覺元件和樣式
// 為何使用：將 UI 樣式集中管理，方便統一修改
// ══════════════════════════════════════════════════════════════════════════════

// Style definitions using Lip Gloss
var (
	// PathBarStyle 路徑列的樣式
	PathBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#1a1a2e")).
			Bold(true).
			Padding(0, 1)

	// FileListStyle 檔案列表區域的樣式
	FileListStyle = lipgloss.NewStyle()

	// SelectedStyle 選中項目的樣式
	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#16213e"))

	// DirectoryStyle 目錄項目的樣式
	DirectoryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4fc3f7"))

	// FileStyle 檔案項目的樣式
	FileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e0e0e0"))

	// PreviewStyle 預覽區域的樣式
	PreviewStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#333333")).
			Padding(0, 1)

	// StatusBarStyle 狀態列的樣式
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#0f3460")).
			Padding(0, 1)

	// ErrorStyle 錯誤訊息的樣式
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff5252"))

	// StatusMessageStyle 狀態訊息的樣式
	StatusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#69f0ae"))
)

// RenderFileList 渲染檔案列表
// 參數 entries 是檔案列表
// 參數 cursor 是目前選中的索引
// 回傳渲染後的字串
func RenderFileList(entries []types.FileEntry, cursor int, width int) string {
	if len(entries) == 0 {
		return "(empty directory)"
	}

	output := ""
	for i, entry := range entries {
		// 決定前綴
		prefix := "  "
		if i == cursor {
			prefix = "> "
		}

		// 決定樣式
		var lineStyle lipgloss.Style
		if i == cursor {
			lineStyle = SelectedStyle
		} else if entry.IsDir {
			lineStyle = DirectoryStyle
		} else {
			lineStyle = FileStyle
		}

		// 類型標記
		typeMark := " "
		if entry.IsDir {
			typeMark = "/"
		}

		// 大小
		sizeStr := FormatSize(entry.Size)

		// 組合行
		line := fmt.Sprintf("%s%s %s %s", prefix, typeMark, entry.Name, sizeStr)

		// 如果行太長，進行截斷
		if len(line) > width-2 {
			line = line[:width-5] + "..."
		}

		output += lineStyle.Render(line) + "\n"
	}

	return output
}

// RenderPathBar 渲染路徑列
func RenderPathBar(path string, width int) string {
	text := "PATH: " + path
	if len(text) > width {
		text = "..." + text[len(text)-width+3:]
	}
	return PathBarStyle.Render(text)
}

// RenderStatusBar 渲染狀態列
func RenderStatusBar(message string, width int) string {
	text := message
	if len(text) > width {
		text = text[:width-3] + "..."
	}
	return StatusBarStyle.Render(text)
}

// RenderError 渲染錯誤訊息
func RenderError(err string) string {
	if err == "" {
		return ""
	}
	return ErrorStyle.Render("[ERROR] " + err)
}

// RenderStatusMessage 渲染狀態訊息
func RenderStatusMessage(msg string) string {
	if msg == "" {
		return ""
	}
	return StatusMessageStyle.Render(msg)
}

// FormatSize 格式化檔案大小
func FormatSize(size int64) string {
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
