package app

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gofm/internal/types"
)

func TestViewStaysWithinTerminalDimensions(t *testing.T) {
	for _, size := range [][2]int{{120, 30}, {80, 24}, {50, 12}, {20, 6}, {10, 3}, {1, 1}, {0, 0}, {-1, -1}} {
		m := New(t.TempDir())
		m.Width, m.Height = size[0], size[1]
		m.Entries = []types.FileEntry{{Name: "中文🙂.txt", Path: "selected"}}
		m.PreviewActive = true
		// 測試超長預覽也不能將狀態列推出視窗。
		m.previewPath = "selected"
		m.previewContent = strings.Repeat("PREVIEW 中文🙂 "+strings.Repeat("x", 100)+"\n", 100)
		m.StatusMessage = "Done"
		view := m.View()
		if !utf8.ValidString(view) {
			t.Fatalf("size=%v 無效 UTF-8", size)
		}
		if size[0] <= 0 || size[1] <= 0 {
			if view != "" {
				t.Fatalf("零或負尺寸應為空畫面: %v", size)
			}
			continue
		}
		if lipgloss.Width(view) > size[0] || lipgloss.Height(view) > size[1] {
			t.Fatalf("size=%v 實際畫面 %dx%d", size, lipgloss.Width(view), lipgloss.Height(view))
		}
		if size[0] >= 20 && !strings.Contains(view, "nav") {
			t.Fatalf("size=%v 狀態列不可見", size)
		}
	}
}

func TestWideViewPlacesPreviewBesideList(t *testing.T) {
	m := New(t.TempDir())
	m.Width, m.Height = 100, 12
	m.Entries = []types.FileEntry{{Name: "selected.txt", Path: "selected"}}
	m.PreviewActive = true
	m.previewPath, m.previewContent = "selected", "PREVIEW_MARKER"
	t.Log("寬視窗實際渲染:\n" + m.View())
	for _, line := range strings.Split(m.View(), "\n") {
		if strings.Contains(line, "selected.txt") && strings.Contains(line, "PREVIEW_MARKER") {
			return
		}
	}
	t.Fatal("列表與預覽沒有左右並排")
}

func TestPreviewScrollClampsAndReturnsFromBottom(t *testing.T) {
	m := New(t.TempDir())
	m.Width, m.Height = 80, 12
	m.Entries = []types.FileEntry{{Name: "selected.txt", Path: "selected"}}
	m.PreviewActive = true
	m.previewPath = "selected"
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, fmt.Sprintf("line%02d", i))
	}
	m.previewContent = strings.Join(lines, "\n")
	for i := 0; i < 30; i++ {
		m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	}
	if !strings.Contains(m.View(), "line99") {
		t.Fatal("無法捲動至底部")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if strings.Contains(m.View(), "line99") {
		t.Fatal("PageUp 在底部無效")
	}
	for i := 0; i < 30; i++ {
		m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	}
	if m.previewOffset != 0 || !strings.Contains(m.View(), "line00") {
		t.Fatal("無法捲動回頂部")
	}
}

func TestNarrowViewSwitchesBetweenPreviewAndList(t *testing.T) {
	m := New(t.TempDir())
	m.Width, m.Height = 50, 8
	m.Entries = []types.FileEntry{{Name: "中文🙂.txt", Path: "selected"}}
	m.PreviewActive = true
	m.previewPath, m.previewContent = "selected", "NARROW_PREVIEW 中文🙂"
	view := m.View()
	t.Log("窄視窗實際渲染:\n" + view)
	if !strings.Contains(view, "NARROW_PREVIEW") || strings.Contains(view, "中文🙂.txt") {
		t.Fatal("窄視窗未切換為單一預覽面板")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if view := m.View(); !strings.Contains(view, "中文🙂.txt") || strings.Contains(view, "NARROW_PREVIEW") {
		t.Fatal("Esc 未切回列表")
	}
}

func TestNarrowInputKeepsEditedTailAndConfirmationVisible(t *testing.T) {
	m := New(t.TempDir())
	m.Width, m.Height = 30, 6
	m.Mode = ModeInput
	m.StatusMessage = "newfile"
	m.InputBuffer = strings.Repeat("中文", 30) + "尾🙂"
	if view := m.View(); !strings.Contains(view, "尾🙂_") {
		t.Fatalf("窄視窗看不到目前輸入與游標: %q", view)
	}
	m.Mode = ModeSearch
	m.SearchQuery = m.InputBuffer
	if !strings.Contains(m.View(), "尾🙂_") {
		t.Fatal("窄視窗看不到搜尋輸入游標")
	}
	m.Mode = ModeConfirmDelete
	m.Entries = []types.FileEntry{{Name: strings.Repeat("中文🙂", 30), Path: "selected"}}
	if !strings.Contains(m.View(), "[y/n]") {
		t.Fatal("長檔名遮住刪除確認按鍵")
	}
}
