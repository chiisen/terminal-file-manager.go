package state

import (
	"testing"

	"gofm/internal/types"
)

// 💡 概念：狀態管理器測試
// 說明：驗證游標移動、選取切換等純邏輯，不依賴檔案系統
// 為何使用：state 是導航核心，迴歸測試可防止游標越界等 bug

func TestNew(t *testing.T) {
	m := New()
	if m == nil || m.GetState() == nil {
		t.Fatal("New 應回傳非 nil 的 Manager 與 State")
	}
	if m.GetState().Cursor != 0 {
		t.Errorf("初始 Cursor = %d; want 0", m.GetState().Cursor)
	}
}

func TestSetPath(t *testing.T) {
	m := New()
	m.GetState().Cursor = 5
	m.SetPath("/tmp")
	if m.GetState().CurrentPath != "/tmp" {
		t.Errorf("CurrentPath = %q; want %q", m.GetState().CurrentPath, "/tmp")
	}
	if m.GetState().Cursor != 0 {
		t.Errorf("SetPath 後 Cursor 應重置為 0, got %d", m.GetState().Cursor)
	}
}

func TestMoveCursor(t *testing.T) {
	m := New()
	m.SetEntries([]types.FileEntry{
		{Name: "a"}, {Name: "b"}, {Name: "c"},
	})

	// 上移邊界：已在頂部不應變成負數
	m.MoveCursorUp()
	if m.GetState().Cursor != 0 {
		t.Errorf("頂部上移後 Cursor = %d; want 0", m.GetState().Cursor)
	}

	m.MoveCursorDown()
	m.MoveCursorDown()
	if m.GetState().Cursor != 2 {
		t.Errorf("下移兩次後 Cursor = %d; want 2", m.GetState().Cursor)
	}

	// 下移邊界：已在底部不應越界
	m.MoveCursorDown()
	if m.GetState().Cursor != 2 {
		t.Errorf("底部下移後 Cursor = %d; want 2", m.GetState().Cursor)
	}

	m.MoveCursorUp()
	if m.GetState().Cursor != 1 {
		t.Errorf("上移後 Cursor = %d; want 1", m.GetState().Cursor)
	}
}

func TestGetSelectedEntry(t *testing.T) {
	m := New()
	if got := m.GetSelectedEntry(); got != nil {
		t.Errorf("空列表時應回傳 nil, got %+v", got)
	}

	m.SetEntries([]types.FileEntry{{Name: "a", Path: "/tmp/a"}})
	got := m.GetSelectedEntry()
	if got == nil || got.Name != "a" {
		t.Errorf("GetSelectedEntry = %+v; want Name=a", got)
	}

	// 游標越界時回傳 nil
	m.GetState().Cursor = 99
	if got := m.GetSelectedEntry(); got != nil {
		t.Errorf("越界時應回傳 nil, got %+v", got)
	}
}

func TestToggleSelected(t *testing.T) {
	m := New()
	m.ToggleSelected("/tmp/a")
	if !m.GetState().Selected["/tmp/a"] {
		t.Error("Toggle 後應為選中狀態")
	}
	m.ToggleSelected("/tmp/a")
	if m.GetState().Selected["/tmp/a"] {
		t.Error("再次 Toggle 後應取消選中")
	}
}

func TestGetSelectedPaths(t *testing.T) {
	m := New()
	m.ToggleSelected("/tmp/a")
	m.ToggleSelected("/tmp/b")
	paths := m.GetSelectedPaths()
	if len(paths) != 2 {
		t.Errorf("選中路徑數 = %d; want 2", len(paths))
	}
}
