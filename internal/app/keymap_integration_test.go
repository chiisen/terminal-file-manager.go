package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"gofm/internal/app"
	"gofm/internal/types"
)

func keyModel(t *testing.T, config string) *app.AppState {
	t.Helper()
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "gofm")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	m := app.New(home)
	m.Entries = []types.FileEntry{{Name: "中文.txt", Path: filepath.Join(home, "中文.txt")}, {Name: "other.txt", Path: filepath.Join(home, "other.txt")}}
	return m
}

func press(m *app.AppState, text string) tea.Cmd {
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(text)})
	return cmd
}

func TestCustomKeymapDrivesTUI(t *testing.T) {
	m := keyModel(t, "[keymap]\ndown = 'n'\ncopy = 'c'\nrename = 'e'\nquit = 'z'\n")
	press(m, "n")
	if m.Cursor != 1 {
		t.Fatal("自訂 down 未移動游標")
	}
	press(m, "c")
	if m.Clipboard != m.Entries[1].Path {
		t.Fatal("自訂 copy 未複製選取路徑")
	}
	press(m, "e")
	if m.Mode != app.ModeInput || m.InputBuffer != "other.txt" {
		t.Fatal("自訂 rename 未進入編輯")
	}
	press(m, "中文cz")
	if m.InputBuffer != "other.txt中文cz" {
		t.Fatal("輸入模式誤攔截字元")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd := press(m, "q"); cmd != nil {
		t.Fatal("覆寫後舊退出鍵不應仍然生效")
	}
	if cmd := press(m, "z"); cmd == nil {
		t.Fatal("自訂退出鍵未生效")
	} else if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("未退出")
	}
	m.Width = 180
	if !strings.Contains(m.View(), "z: quit") || !strings.Contains(m.View(), "c/x/p: files") {
		t.Fatal("提示未顯示有效鍵位")
	}
}

func TestCustomMutationKeysRespectBusyOperation(t *testing.T) {
	m := keyModel(t, "[keymap]\npaste = 'v'\nrename = 'e'\nnewfile = 'f'\n")
	source := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(source, []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}
	m.Clipboard = source
	cmd := press(m, "v")
	if cmd == nil {
		t.Fatal("自訂貼上應啟動背景操作")
	}
	press(m, "e")
	press(m, "f")
	if m.Mode != app.ModeNormal {
		t.Fatal("執行期間自訂異動鍵繞過防重複")
	}
	m.Update(cmd())
	if data, err := os.ReadFile(filepath.Join(m.CurrentPath, "source.txt")); err != nil || string(data) != "payload" {
		t.Fatalf("貼上失敗: %q %v", data, err)
	}
}

func TestKeymapConflictFallbackAndSearchInput(t *testing.T) {
	for _, config := range []string{"", "[keymap]\ncopy = 'd'\n", "[keymap]\nquit = 'esc'\n", "[keymap]\nopen = 'not-a-key'\n"} {
		m := keyModel(t, config)
		press(m, "j")
		if m.Cursor != 1 {
			t.Fatal("預設導航失效")
		}
		press(m, "d")
		if m.Mode != app.ModeConfirmDelete {
			t.Fatal("衝突設定不應覆蓋刪除")
		}
		m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		press(m, "/")
		press(m, "j中文q")
		if m.SearchQuery != "j中文q" {
			t.Fatal("搜尋字元被快捷鍵攔截")
		}
		m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		if cmd == nil {
			t.Fatal("固定 ctrl+c 退出失效")
		}
	}
}
