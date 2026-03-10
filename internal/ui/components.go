package ui

import (
	"fmt"
	"time"

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
	// PathBarStyle 路徑列的樣式（放大版）
	PathBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#1a1a2e")).
			Bold(true).
			Padding(1, 2).
			Height(2)

	// FileListStyle 檔案列表區域的樣式
	FileListStyle = lipgloss.NewStyle()

	// SelectedStyle 選中項目的樣式（放大版）
	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#5e81ac")). // 北歐風格藍 (Nord Blue)
			Bold(true).
			Padding(0, 1)

	// DirectoryStyle 目錄項目的樣式（放大版）
	DirectoryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#88c0d0")). // 冰藍色
			Bold(true).
			Padding(0, 1)

	// FileStyle 檔案項目的樣式（放大版）
	FileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#eceff4")). // 柔和雪白
			Padding(0, 1)

	// PreviewStyle 預覽區域的樣式（放大版）
	PreviewStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#333333")).
			Padding(1, 2)

	// StatusBarStyle 狀態列的樣式（放大版）
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#0f3460")).
			Padding(1, 2).
			Height(2)

	// ErrorStyle 錯誤訊息的樣式
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff5252")).
			Padding(0, 1)

	// StatusMessageStyle 狀態訊息的樣式
	StatusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#69f0ae")).
				Padding(0, 1)
)

// RenderFileList 渲染檔案列表（放大版 + 滾動支援）
// 參數 entries 是檔案列表
// 參數 cursor 是目前選中的索引
// 參數 selected 是已被選取的檔案 map（key 為檔案路徑）
// 參數 height 是可用高度
// 回傳渲染後的字串
func RenderFileList(entries []types.FileEntry, cursor int, selected map[string]bool, width int, height int) string {
	if len(entries) == 0 {
		return "(empty directory)"
	}

	// 計算可顯示的起始位置（滾動邏輯）
	start := 0
	if len(entries) > height-4 { // 預留標題列和狀態列的空間
		// 讓游標保持在畫面中央
		start = cursor - height/2
		if start < 0 {
			start = 0
		}
		if start+height-4 > len(entries) {
			start = len(entries) - height + 4
			if start < 0 {
				start = 0
			}
		}
	}

	// 計算結束位置
	end := start + height - 4
	if end > len(entries) {
		end = len(entries)
	}

	output := ""
	for i := start; i < end; i++ {
		entry := entries[i]
		// 決定前綴與游標（放大版）
		prefix := "   "
		if i == cursor {
			prefix = " ▶ "
		} else if selected[entry.Path] {
			prefix = " ✓ " // 選取標記
		}

		// 決定圖示與基本樣式
		icon := " 📄 "
		var lineStyle lipgloss.Style
		isSelected := selected[entry.Path]
		if i == cursor {
			lineStyle = SelectedStyle
		} else if isSelected {
			// 選取的檔案（但不是目前游標）：使用不同的背景色
			lineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#4a6fa5")).
				Padding(0, 1)
		} else if entry.IsDir {
			lineStyle = DirectoryStyle
		} else {
			lineStyle = FileStyle
		}

		if entry.IsDir {
			icon = " 📁 "
		}

		// 檔案大小和時間格式化
		sizeStr := ""
		timeStr := FormatTime(entry.ModTime)
		if !entry.IsDir {
			sizeStr = FormatSize(entry.Size)
		}

		// 排版計算：確保檔名過長時會被截斷，並讓 Size 和 Time 欄位靠右對齊
		maxNameLen := width - 40 // 預留給圖示、游標、Size、Time、Padding 的空間（放大版）
		if maxNameLen < 10 {
			maxNameLen = 10
		}

		nameStr := entry.Name
		if len(nameStr) > maxNameLen {
			nameStr = nameStr[:maxNameLen-3] + "..."
		}

		// %-*s 保證檔名區塊寬度固定，讓後方的 sizeStr 和 timeStr 能整齊切齊
		line := fmt.Sprintf("%s%s %-*s %10s %8s", prefix, icon, maxNameLen, nameStr, sizeStr, timeStr)

		// 若為選取狀態，讓背景色能覆蓋整行寬度
		if i == cursor || isSelected {
			output += lineStyle.Width(width).Render(line) + "\n"
		} else {
			output += lineStyle.Render(line) + "\n"
		}
	}

	return output
}

// RenderPathBar 渲染路徑列
// 自動填滿視窗：使用 width 參數來決定顯示寬度
func RenderPathBar(path string, width int) string {
	// 確保最小寬度
	if width < 10 {
		width = 10
	}
	text := "PATH: " + path
	if len(text) > width {
		// 從右側開始顯示，保留空間給 "PATH: "
		pathStart := len("PATH: ")
		availablePathWidth := width - pathStart - 3 // 3 個點
		if availablePathWidth > 0 {
			text = "PATH: ..." + path[len(path)-availablePathWidth:]
		} else {
			text = "PATH:"
		}
	}
	// 使用 Width() 確保填滿整行
	return PathBarStyle.Width(width).Render(text)
}

// RenderStatusBar 渲染狀態列
// 自動填滿視窗：使用 width 參數來決定顯示寬度
func RenderStatusBar(message string, width int) string {
	// 確保最小寬度
	if width < 10 {
		width = 10
	}
	text := message
	if len(text) > width {
		text = text[:width-3] + "..."
	}
	// 使用 Width() 確保填滿整行
	return StatusBarStyle.Width(width).Render(text)
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

// FormatTime 格式化時間顯示
func FormatTime(t time.Time) string {
	now := time.Now()

	// 如果是今天的檔案，顯示時間
	if t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day() {
		return t.Format("15:04")
	}

	// 如果是今年的檔案，顯示月日
	if t.Year() == now.Year() {
		return t.Format("01/02")
	}

	// 顯示年月日
	return t.Format("2006/01/02")
}
