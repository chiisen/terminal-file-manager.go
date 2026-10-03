package ui

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"gofm/internal/types"
)

func TestBarsAndFileNamesFitDisplayWidth(t *testing.T) {
	for _, width := range []int{1, 2, 7, 20, 40, 80} {
		entries := []types.FileEntry{{Name: strings.Repeat("檔名🙂", 20), Path: "one", Size: 1024, ModTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)}}
		for _, view := range []string{RenderPathBar("/目錄/"+strings.Repeat("🙂中文", 20), width), RenderStatusBar(strings.Repeat("中文🙂", 20), width), RenderFileList(entries, 0, nil, width, 8)} {
			if !utf8.ValidString(view) {
				t.Fatalf("width=%d 截斷破壞 UTF-8", width)
			}
			if lipgloss.Width(view) > width {
				t.Fatalf("width=%d 渲染超出: %d", width, lipgloss.Width(view))
			}
		}
	}
}

func TestFileListAlignsUnicodeNamesAndShowsCursorInOneRow(t *testing.T) {
	entries := []types.FileEntry{{Name: "中文🙂", Size: 1024}, {Name: "abc", Size: 1024}}
	view := RenderFileList(entries, 0, nil, 80, 8)
	var sizeColumns []int
	for _, line := range strings.Split(strings.TrimSuffix(view, "\n"), "\n") {
		if idx := strings.Index(line, "1.0 KB"); idx >= 0 {
			sizeColumns = append(sizeColumns, lipgloss.Width(line[:idx]))
		}
	}
	if len(sizeColumns) != 2 || sizeColumns[0] != sizeColumns[1] {
		t.Fatalf("Unicode 欄位未對齊: %v", sizeColumns)
	}
	if view := RenderFileList(entries, 1, nil, 20, 1); !strings.Contains(view, "abc") || lipgloss.Height(view) > 1 {
		t.Fatalf("小視窗無法顯示游標: %q", view)
	}
}

func TestTailKeepsCompleteEmojiAndUnicode(t *testing.T) {
	for _, tc := range []struct {
		text  string
		width int
		want  string
	}{
		{"abcdef中文👨‍👩‍👧‍👦", 5, "...👨‍👩‍👧‍👦"},
		{"abcdef中文🙂", 7, "...文🙂"},
		{"中文", 2, "文"},
		{"中文", 0, ""},
	} {
		if got := FitTail(tc.text, tc.width); got != tc.want {
			t.Fatalf("width=%d got=%q want=%q", tc.width, got, tc.want)
		}
	}
}
