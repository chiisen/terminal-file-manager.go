package ui

import (
	"fmt"

	"gofm/internal/types"

	"github.com/charmbracelet/lipgloss"
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
			Background(lipgloss.Color("#5e81ac")). // 北歐風格藍 (Nord Blue)
			Bold(true)

	// DirectoryStyle 目錄項目的樣式
	DirectoryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#88c0d0")). // 冰藍色
			Bold(true)

	// FileStyle 檔案項目的樣式
	FileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#eceff4")) // 柔和雪白

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
		// 決定前綴與游標
		prefix := "  "
		if i == cursor {
			prefix = "▶ "
		}

		// 決定圖示與基本樣式
		icon := "📄"
		var lineStyle lipgloss.Style
		if i == cursor {
			lineStyle = SelectedStyle
		} else if entry.IsDir {
			lineStyle = DirectoryStyle
		} else {
			lineStyle = FileStyle
		}

		if entry.IsDir {
			icon = "📁"
		}

		// 檔案大小格式化
		sizeStr := ""
		if !entry.IsDir {
			sizeStr = FormatSize(entry.Size)
		}

		// 排版計算：確保檔名過長時會被截斷，並讓 Size 欄位靠右對齊
		maxNameLen := width - 22 // 預留給圖示、游標、Size、Padding 的空間
		if maxNameLen < 10 {
			maxNameLen = 10
		}

		nameStr := entry.Name
		if len(nameStr) > maxNameLen {
			nameStr = nameStr[:maxNameLen-3] + "..."
		}

		// %-*s 保證檔名區塊寬度固定，讓後方的 sizeStr 能整齊切齊
		line := fmt.Sprintf("%s%s %-*s %9s ", prefix, icon, maxNameLen, nameStr, sizeStr)

		// 若為選取狀態，讓背景色能覆蓋整行寬度
		if i == cursor {
			output += lineStyle.Width(width).Render(line) + "\n"
		} else {
			output += lineStyle.Render(line) + "\n"
		}
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
