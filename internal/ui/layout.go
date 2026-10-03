package ui

import (
	"strings"
	"unicode"

	"gofm/internal/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// FitText 以 grapheme 與終端顯示寬度截斷，不切開中文或 emoji。
func FitText(text string, width int) string {
	if width <= 0 {
		return ""
	}
	tail := ""
	if width >= 3 {
		tail = "..."
	}
	return ansi.Truncate(text, width, tail)
}

// FitTail 保留輸入或路徑尾端，讓小視窗仍可看見編輯位置。
func FitTail(text string, width int) string {
	if width <= 0 {
		return ""
	}
	text = singleLine(text)
	textWidth := ansi.StringWidth(text)
	if textWidth <= width {
		return text
	}
	prefix := ""
	if width >= 3 {
		prefix = "..."
	}
	available := width - len(prefix)
	trim := textWidth - available
	tail := ansi.TruncateLeft(text, trim, "")
	// 起點可能落在雙寬字元中間，從左側多移一格，保留右側完整字元。
	for ansi.StringWidth(tail) > available {
		trim++
		tail = ansi.TruncateLeft(text, trim, "")
	}
	return prefix + tail
}

// InputStatus 優先顯示編輯尾端與游標，剩餘空間才顯示快捷鍵提示。
func InputStatus(label, value, hint string, width int) string {
	innerWidth := width
	if width >= 3 {
		innerWidth -= 2
	}
	prefix := label + ": "
	if ansi.StringWidth(prefix)+1 > innerWidth {
		prefix = ""
	}
	value = FitTail(value, max(0, innerWidth-ansi.StringWidth(prefix)-1))
	text := prefix + value + "_"
	remaining := innerWidth - ansi.StringWidth(text) - 2
	if remaining > 0 {
		text += "  " + FitText(singleLine(hint), remaining)
	}
	return text
}

func singleLine(text string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, text)
}

func padLeft(text string, width int) string {
	return strings.Repeat(" ", max(0, width-ansi.StringWidth(text))) + text
}

func renderBar(style lipgloss.Style, text string, width int) string {
	if width <= 0 {
		return ""
	}
	padding := 0
	if width >= 3 {
		padding = 1
	}
	text = FitText(text, width-2*padding)
	return style.Padding(0, padding).Height(1).Width(width).MaxWidth(width).MaxHeight(1).Render(text)
}

// FitBlock 限制區塊寬高並補齊空白列，避免面板與狀態列超出終端。
func FitBlock(text string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	lines := strings.Split(text, "\n")
	rows := make([]string, height)
	for i := range rows {
		if i < len(lines) {
			rows[i] = ansi.Truncate(lines[i], width, "")
		}
		rows[i] += strings.Repeat(" ", max(0, width-ansi.StringWidth(rows[i])))
	}
	return strings.Join(rows, "\n")
}

func panelSize(width, height int) (innerWidth, innerHeight int, border bool) {
	if width >= 6 && height >= 3 {
		return width - 4, height - 2, true
	}
	return width, height, false
}

func renderPanel(style lipgloss.Style, text string, width, height int) string {
	innerWidth, innerHeight, border := panelSize(width, height)
	content := FitBlock(text, innerWidth, innerHeight)
	if !border {
		return FitBlock(content, width, height)
	}
	return style.Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#333333")).Padding(0, 1).Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render(content)
}

// RenderFilePanel 將可見檔案列配置於實際面板尺寸，低矮視窗省略邊框。
func RenderFilePanel(entries []types.FileEntry, cursor int, selected map[string]bool, width, height int) string {
	innerWidth, innerHeight, _ := panelSize(width, height)
	return renderPanel(FileListStyle, RenderFileList(entries, cursor, selected, innerWidth, innerHeight), width, height)
}

// PreviewLines 回傳寬度適配後的預覽列；不讓控制字元影響面板排版。
func PreviewLines(content string, width int) []string {
	if width <= 0 {
		return nil
	}
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\t", "    ")
	content = strings.Map(func(r rune) rune {
		if r != '\n' && unicode.IsControl(r) {
			return ' '
		}
		return r
	}, content)
	return strings.Split(ansi.Wrap(content, width, ""), "\n")
}

// RenderPreview 只渲染捲動位置起算的可見列，內容不會推開其他區域。
func RenderPreview(content string, width, height, offset int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	innerWidth, innerHeight, _ := panelSize(width, height)
	lines := PreviewLines(content, innerWidth)
	offset = max(0, min(offset, max(0, len(lines)-innerHeight)))
	return renderPanel(PreviewStyle, strings.Join(lines[offset:], "\n"), width, height)
}

// PreviewMaxOffset 將捲動位置限制在實際最後一頁，避免 PageUp 在底部失效。
func PreviewMaxOffset(content string, width, height int) int {
	if width <= 0 || height <= 0 {
		return 0
	}
	innerWidth, innerHeight, _ := panelSize(width, height)
	return max(0, len(PreviewLines(content, innerWidth))-innerHeight)
}
