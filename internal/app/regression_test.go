package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// 背景命令不得修改畫面狀態，完成訊息必須經 Update 才套用。
func TestDirectoryCommandUpdatesThroughMessages(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "中文.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	m := New(dir)
	msg := m.Init()()
	if len(m.Entries) != 0 {
		t.Fatal("背景命令直接修改 Entries")
	}
	_, metadata := m.Update(msg)
	if len(m.Entries) != 1 || m.Entries[0].Name != "中文.txt" {
		t.Fatalf("完成訊息未載入列表: %+v", m.Entries)
	}
	if metadata == nil {
		t.Fatal("未排程 metadata")
	}
	_, gitCmd := m.Update(metadata())
	if m.Entries[0].Size != 5 || m.Entries[0].Permission == "" {
		t.Fatalf("metadata 未套用: %+v", m.Entries[0])
	}
	if gitCmd == nil {
		t.Fatal("未排程 Git 載入")
	}
	m.Update(gitCmd())
	if m.GitInfo == nil {
		t.Fatal("Git 完成訊息未套用")
	}
}

// 快速導航時，舊路徑的命令必須讀取原先快照並被忽略。
func TestDirectoryIgnoresOutdatedResults(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(a, "old.txt"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(b, "new.txt"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	m := New(a)
	old := m.Init()
	m.CurrentPath = b
	current := m.Init()
	oldMsg := old()
	if len(m.Entries) != 0 {
		t.Fatal("舊命令修改目前狀態")
	}
	m.Update(current())
	m.Update(oldMsg)
	if len(m.Entries) != 1 || m.Entries[0].Name != "new.txt" {
		t.Fatalf("舊結果覆蓋目前目錄: %+v", m.Entries)
	}
}

func TestDirectorySamePathReloadIgnoresOldStages(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(p, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	m := New(dir)
	old := m.Init()()
	_, metadata := m.Update(old)
	oldMetadata := metadata()
	_, gitCmd := m.Update(oldMetadata)
	oldGit := gitCmd()
	if err := os.WriteFile(p, []byte("new content"), 0644); err != nil {
		t.Fatal(err)
	}
	current := m.Init()
	_, latestMetadata := m.Update(current())
	m.Update(latestMetadata())
	m.Update(old)
	m.Update(oldMetadata)
	m.Update(oldGit)
	if len(m.Entries) != 1 || m.Entries[0].Size != 11 {
		t.Fatalf("同路徑舊資料覆蓋新載入: %+v", m.Entries)
	}
	if m.GitInfo != nil {
		t.Fatal("過期 Git 資訊已套用")
	}
}

func TestMetadataUsesPathsAndPreservesSearch(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{"a.txt": "a", "b.txt": "bbbb"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	m := New(dir)
	_, metadata := m.Update(m.Init()())
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	m.Update(metadata())
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if len(m.Entries) != 2 {
		t.Fatalf("搜尋列表未還原: %+v", m.Entries)
	}
	for _, entry := range m.Entries {
		want := int64(1)
		if entry.Name == "b.txt" {
			want = 4
		}
		if entry.Size != want || entry.Permission == "" {
			t.Fatalf("搜尋還原的 metadata 錯配或遺失: %+v", entry)
		}
	}
}

func TestMetadataMatchesAfterSorting(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{"a.txt": "aaaa", "b.txt": "b"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	m := New(dir)
	m.SortBy = "size"
	_, metadata := m.Update(m.Init()())
	m.Cursor = 1
	selected := m.Entries[m.Cursor].Path
	m.Update(metadata())
	if m.Entries[0].Name != "b.txt" || m.Entries[0].Size != 1 || m.Entries[1].Size != 4 {
		t.Fatalf("排序後 metadata 錯配: %+v", m.Entries)
	}
	if m.Entries[m.Cursor].Path != selected {
		t.Fatal("metadata 更新改變游標選取")
	}
}

func TestGitInfoLoadedForRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("未安裝 Git")
	}
	dir := t.TempDir()
	if output, err := exec.Command("git", "init", "-q", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, output)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}
	m := New(dir)
	_, metadata := m.Update(m.Init()())
	_, gitCmd := m.Update(metadata())
	m.Update(gitCmd())
	if m.GitInfo == nil || !m.GitInfo.IsRepo || len(m.GitInfo.StatusMap) != 1 {
		t.Fatalf("Git 狀態未載入: %+v", m.GitInfo)
	}
}

func TestUnicodeNewFileThroughUpdate(t *testing.T) {
	dir := t.TempDir()
	m := New(dir)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("中文🙂.txt"), Paste: true})
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if _, err := os.Stat(filepath.Join(dir, "中文🙂.txt")); err != nil {
		t.Fatalf("快速輸入未建立 Unicode 檔案: %v", err)
	}
}

// 真實背景命令與事件更新交錯執行，讓 race detector 檢查共享 Model 存取。
func TestDirectoryCommandsDoNotRaceWithUpdates(t *testing.T) {
	m := New(t.TempDir())
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		cmd := m.Init()
		wg.Add(1)
		go func() { defer wg.Done(); cmd() }()
		m.Update(tea.WindowSizeMsg{Width: 80 + i, Height: 24})
		m.View()
	}
	wg.Wait()
}

func TestDirectoryLoadErrorDisplayed(t *testing.T) {
	m := New(filepath.Join(t.TempDir(), "missing"))
	m.Update(m.Init()())
	if m.ErrorMessage == "" {
		t.Fatal("載入錯誤未顯示")
	}
}

// 快速輸入、多字元貼上及退格都必須保留完整 Unicode 字元。
func TestRapidUnicodeInputAndBackspace(t *testing.T) {
	for _, mode := range []AppMode{ModeInput, ModeSearch} {
		t.Run(map[AppMode]string{ModeInput: "input", ModeSearch: "search"}[mode], func(t *testing.T) {
			m := New(t.TempDir())
			m.Mode = mode
			m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("中文🙂")})
			m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ab"), Paste: true})
			m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
			m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
			m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
			got := m.InputBuffer
			if mode == ModeSearch {
				got = m.SearchQuery
			}
			if got != "中文" {
				t.Fatalf("輸入或退格遺失字元: %q", got)
			}
		})
	}
}
