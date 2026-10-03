package ui

import (
	"fmt"
	"strings"
	"time"

	"gofm/internal/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
	if width <= 0 || height <= 0 {
		return ""
	}
	if len(entries) == 0 {
		return FitText("(empty directory)", width)
	}
	// height 為真正可用列數；即使只有一列，也必須顯示目前游標。
	cursor = max(0, min(cursor, len(entries)-1))
	start := max(0, cursor-height/2)
	start = min(start, max(0, len(entries)-height))
	end := min(start+height, len(entries))
	var output strings.Builder
	for i := start; i < end; i++ {
		entry := entries[i]
		// 決定前綴與游標（放大版）
		prefix := "  "
		if i == cursor {
			prefix = "▶ "
		} else if selected[entry.Path] {
			prefix = "✓ " // 選取標記
		}

		// 決定圖示與基本樣式
		icon := "📄 "
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
			icon = "📁 "
		}

		// 檔案大小和時間格式化
		sizeStr := ""
		timeStr := FormatTime(entry.ModTime)
		if !entry.IsDir {
			sizeStr = FormatSize(entry.Size)
		}

		// 按終端 cell 寬度排版，窄視窗逐步隱藏時間、大小與圖示。
		padding := 0
		if width >= 3 {
			padding = 1
		}
		innerWidth := width - 2*padding
		if innerWidth < 8 {
			icon = ""
		}
		var suffix string
		if innerWidth >= 36 {
			suffix = " " + padLeft(FitText(sizeStr, 10), 10) + " " + padLeft(timeStr, 10)
		} else if innerWidth >= 24 {
			suffix = " " + padLeft(FitText(sizeStr, 10), 10)
		}
		nameWidth := max(0, innerWidth-ansi.StringWidth(prefix+icon+suffix))
		name := FitText(singleLine(entry.Name), nameWidth)
		line := prefix + icon + name + strings.Repeat(" ", max(0, nameWidth-ansi.StringWidth(name))) + suffix
		line = ansi.Truncate(line, innerWidth, "")
		if i > start {
			output.WriteByte('\n')
		}
		output.WriteString(lineStyle.Padding(0, padding).Width(width).MaxWidth(width).MaxHeight(1).Render(line))
	}
	return output.String()
}

// RenderPathBar 渲染路徑列
// 自動填滿視窗：使用 width 參數來決定顯示寬度
func RenderPathBar(path string, width int) string {
	path = singleLine(path)
	available := width - 8 // PATH: 與左右 padding。
	if available >= 3 && ansi.StringWidth(path) > available {
		path = FitTail(path, available)
	}
	return renderBar(PathBarStyle, "PATH: "+path, width)
}

// RenderStatusBar 渲染狀態列
// 自動填滿視窗：使用 width 參數來決定顯示寬度
func RenderStatusBar(message string, width int) string {
	return renderBar(StatusBarStyle, singleLine(message), width)
}

// RenderError 渲染錯誤訊息
func RenderError(err string) string {
	if err == "" {
		return ""
	}
	return ErrorStyle.Render("[ERROR] " + singleLine(err))
}

// RenderStatusMessage 渲染狀態訊息
func RenderStatusMessage(msg string) string {
	if msg == "" {
		return ""
	}
	return StatusMessageStyle.Render(singleLine(msg))
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
	if t.IsZero() {
		return "-"
	}
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
