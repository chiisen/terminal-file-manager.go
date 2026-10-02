package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"gofm/internal/types"
)

func searchKey(m *AppState, query string) {
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	if query != "" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(query)})
	}
}

func TestSearchEnterOpensChosenResultAndRestoresList(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a-jk.txt", "b-jk.txt", "other.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0644); err != nil {
			t.Fatal(err)
		}
	}
	m := New(dir)
	m.Update(m.Init()())
	searchKey(m, "jk")
	if m.SearchQuery != "jk" || len(m.Entries) != 2 {
		t.Fatal("j/k 被當成導航")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Mode != ModeNormal || len(m.Entries) != 3 || m.Entries[m.Cursor].Name != "b-jk.txt" || cmd == nil {
		t.Fatal("Enter 未確認目前選取或未恢復完整列表")
	}
	m.Update(cmd())
	if !strings.Contains(m.View(), "b-jk.txt") || !m.PreviewActive {
		t.Fatal("搜尋確認後未開啟檔案預覽")
	}
	if m.OriginalEntries != nil || m.SearchQuery != "" || m.SearchResults != nil {
		t.Fatal("搜尋上下文未清理")
	}
}

func TestSearchEnterOpensDirectory(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "chosen")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	m := New(dir)
	m.Update(m.Init()())
	searchKey(m, "chosen")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Mode != ModeNormal || m.CurrentPath != sub || cmd == nil {
		t.Fatal("搜尋確認未進入目錄")
	}
	m.Update(cmd())
	if len(m.Entries) != 0 {
		t.Fatal("未載入子目錄")
	}
}

func TestSearchEmptyQueryNavigationAndNoMatches(t *testing.T) {
	m := New(t.TempDir())
	m.Entries = []types.FileEntry{{Name: "one", Path: "one"}, {Name: "two", Path: "two"}}
	searchKey(m, "")
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Cursor != 1 {
		t.Fatal("空查詢不能導航完整列表")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("missing")})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil || m.Mode != ModeSearch || len(m.Entries) != 0 {
		t.Fatal("空結果誤觸操作")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.Mode != ModeNormal || len(m.Entries) != 2 {
		t.Fatal("Esc 未恢復列表")
	}
}
