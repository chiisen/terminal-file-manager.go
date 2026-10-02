package app

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"gofm/internal/fs"
	"gofm/internal/types"
)

func TestPreviewLoadsInCommandAndViewUsesCache(t *testing.T) {
	p := filepath.Join(t.TempDir(), "a.txt")
	if err := os.WriteFile(p, []byte("cached preview"), 0644); err != nil {
		t.Fatal(err)
	}
	m := New(filepath.Dir(p))
	m.Entries = []types.FileEntry{{Name: "a.txt", Path: p}}
	_, cmd := m.HandleOpen()
	if cmd == nil {
		t.Fatal("預覽沒有背景命令")
	}
	m.Update(cmd())
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if !strings.Contains(m.View(), "cached preview") {
			t.Fatal("View 重新讀取檔案或未保留預覽")
		}
	}
	if _, again := m.HandleOpen(); again != nil {
		t.Fatal("相同檔案未使用預覽快取")
	}
}

func TestPreviewIgnoresOldSelection(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{"a.txt": "old preview", "b.txt": "new preview"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	m := New(dir)
	m.Entries = []types.FileEntry{{Name: "a.txt", Path: filepath.Join(dir, "a.txt")}, {Name: "b.txt", Path: filepath.Join(dir, "b.txt")}}
	_, old := m.HandleOpen()
	if old == nil {
		t.Fatal("沒有背景預覽")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, current := m.HandleOpen()
	m.Update(current())
	m.Update(old())
	if view := m.View(); !strings.Contains(view, "new preview") || strings.Contains(view, "old preview") {
		t.Fatal("舊預覽覆蓋目前選取")
	}
}

func TestPasteAndDeleteRunOnlyInCommands(t *testing.T) {
	src := filepath.Join(t.TempDir(), "a.txt")
	if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	m := New(t.TempDir())
	m.Clipboard = src
	_, copyCmd := m.HandlePaste()
	if copyCmd == nil {
		t.Fatal("複製没有背景命令")
	}
	dst := filepath.Join(m.CurrentPath, "a.txt")
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Fatal("命令執行前已複製檔案")
	}
	if _, duplicate := m.HandlePaste(); duplicate != nil {
		t.Fatal("允許重複提交")
	}
	_, reload := m.Update(copyCmd())
	if data, err := os.ReadFile(dst); err != nil || string(data) != "content" {
		t.Fatalf("複製失敗: %q %v", data, err)
	}
	if reload == nil {
		t.Fatal("操作後未刷新")
	}
	m.Update(reload())
	_, deleteCmd := m.handleDeleteConfirm()
	if _, err := os.Stat(dst); err != nil {
		t.Fatal("背景命令前已刪除")
	}
	m.Update(deleteCmd())
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Fatal("未刪除檔案")
	}
}

func TestFailedCutKeepsClipboard(t *testing.T) {
	src := filepath.Join(t.TempDir(), "missing.txt")
	m := New(t.TempDir())
	m.Clipboard = src
	m.IsCut = true
	_, cmd := m.HandlePaste()
	if cmd == nil {
		t.Fatal("沒有背景移動命令")
	}
	m.Update(cmd())
	if m.Clipboard != src || !m.IsCut || m.ErrorMessage == "" {
		t.Fatal("失敗移動遺失剪貼簿或錯誤")
	}
}

// 用 channel 控制真實複製開始與完成，驗證操作期間事件迴圈仍可更新。
func TestSlowOperationAllowsUpdatesAndRejectsDuplicates(t *testing.T) {
	src := filepath.Join(t.TempDir(), "a.txt")
	if err := os.WriteFile(src, []byte("slow copy"), 0644); err != nil {
		t.Fatal(err)
	}
	m := New(t.TempDir())
	dst := filepath.Join(m.CurrentPath, "a.txt")
	m.Entries = []types.FileEntry{{Name: "a"}, {Name: "b"}}
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	cmd := m.startOperation("Copying a.txt...", "Copied: a.txt", src, false, func() error {
		close(started)
		<-release
		return fs.CopyFile(src, dst)
	})
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("操作未開始")
	}
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Width != 100 || m.Cursor != 1 || !strings.Contains(m.View(), "Copying") {
		t.Fatal("操作阻塞 UI 或未顯示執行狀態")
	}
	if _, duplicate := m.HandlePaste(); duplicate != nil {
		t.Fatal("允許重複操作")
	}
	unblock()
	select {
	case msg := <-done:
		m.Update(msg)
	case <-time.After(5 * time.Second):
		t.Fatal("操作未完成")
	}
	if m.operationBusy || m.ErrorMessage != "" || !strings.Contains(m.StatusMessage, "Copied:") {
		t.Fatal("成功狀態未套用")
	}
	if data, err := os.ReadFile(dst); err != nil || string(data) != "slow copy" {
		t.Fatalf("實際操作未完成: %q %v", data, err)
	}
}

func TestMoveUsesSnapshotAndPreservesNewClipboard(t *testing.T) {
	src := filepath.Join(t.TempDir(), "old.txt")
	if err := os.WriteFile(src, []byte("move"), 0644); err != nil {
		t.Fatal(err)
	}
	m := New(t.TempDir())
	m.Clipboard, m.IsCut = src, true
	_, cmd := m.HandlePaste()
	dst := filepath.Join(m.CurrentPath, "old.txt")
	m.CurrentPath = t.TempDir()
	m.Clipboard, m.IsCut = "new clipboard", false
	_, reload := m.Update(cmd())
	if reload != nil || m.Clipboard != "new clipboard" {
		t.Fatal("背景完成覆蓋新目錄或剪貼簿")
	}
	if data, err := os.ReadFile(dst); err != nil || string(data) != "move" {
		t.Fatalf("移動未使用原始快照: %q %v", data, err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatal("來源未移動")
	}
	if !strings.Contains(m.StatusMessage, "Moved:") {
		t.Fatal("移動誤標示為複製")
	}
}
