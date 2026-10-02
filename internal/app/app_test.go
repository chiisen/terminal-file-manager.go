package app

import (
	"os"
	"path/filepath"
	"testing"

	"gofm/internal/types"
)

// 💡 概念：AppState 核心邏輯測試
// 說明：鎖定排序、搜尋、導航等純邏輯；檔案建立類測試一律用 t.TempDir() 隔離
// 為何使用：app.go 是 800+ 行的核心，零測試時任何重構都高風險

func TestNew_Defaults(t *testing.T) {
	m := New(t.TempDir())
	if m.Cursor != 0 {
		t.Errorf("初始 Cursor = %d; want 0", m.Cursor)
	}
	if m.Mode != ModeNormal {
		t.Errorf("初始 Mode = %v; want ModeNormal", m.Mode)
	}
	if m.Selected == nil {
		t.Error("Selected map 應初始化，不可為 nil")
	}
	if m.SortBy != "name" || !m.SortAsc {
		t.Errorf("預設排序 = %q/%v; want name/true", m.SortBy, m.SortAsc)
	}
}

func TestSetMode(t *testing.T) {
	m := New(t.TempDir())
	m.SetMode(ModeSearch)
	if m.Mode != ModeSearch {
		t.Errorf("Mode = %v; want ModeSearch", m.Mode)
	}
}

func TestSortEntries_NameDirsFirst(t *testing.T) {
	m := New(t.TempDir())
	m.SortBy = "name"
	m.SortAsc = true
	m.Entries = []types.FileEntry{
		{Name: "z.txt", IsDir: false, Size: 10},
		{Name: "b", IsDir: true},
		{Name: "a.txt", IsDir: false, Size: 5},
	}
	m.SortEntries()

	if !m.Entries[0].IsDir {
		t.Errorf("目錄應排在最前，got %+v", m.Entries[0])
	}
	if m.Entries[1].Name != "a.txt" || m.Entries[2].Name != "z.txt" {
		t.Errorf("升序應為 a.txt,z.txt，got %q,%q", m.Entries[1].Name, m.Entries[2].Name)
	}
}

func TestSortEntries_SizeDesc(t *testing.T) {
	m := New(t.TempDir())
	m.SortBy = "size"
	m.SortAsc = false
	m.Entries = []types.FileEntry{
		{Name: "small", Size: 10},
		{Name: "big", Size: 1000},
	}
	m.SortEntries()
	if m.Entries[0].Name != "big" {
		t.Errorf("降序首位應為 big，got %q", m.Entries[0].Name)
	}
}

func TestSortEntries_NilSafe(t *testing.T) {
	m := New(t.TempDir())
	m.Entries = nil
	// 不應 panic
	m.SortEntries()
}

func TestSortEntries_ModifiedFallsBackToName(t *testing.T) {
	// 💡 現況：modified 尚未實作，fallthrough 用名稱排序（見 docs/OPEN_QUESTIONS.md 項 4）
	// 此測試鎖定目前行為，實作真排序時需同步更新
	m := New(t.TempDir())
	m.SortBy = "modified"
	m.SortAsc = true
	m.Entries = []types.FileEntry{
		{Name: "b.txt"},
		{Name: "a.txt"},
	}
	m.SortEntries()
	if m.Entries[0].Name != "a.txt" || m.Entries[1].Name != "b.txt" {
		t.Errorf("modified 暫代行為應為名稱升序，got %q,%q", m.Entries[0].Name, m.Entries[1].Name)
	}
}

func TestCycleSortBy(t *testing.T) {
	m := New(t.TempDir())
	order := []string{"size", "type", "modified", "name"}
	for _, want := range order {
		m.cycleSortBy()
		if m.SortBy != want {
			t.Errorf("cycleSortBy 後 = %q; want %q", m.SortBy, want)
		}
	}
}

func TestGetExt(t *testing.T) {
	tests := []struct{ in, want string }{
		{"a.txt", ".txt"},
		{"archive.tar.gz", ".gz"},
		{"noext", ""},
		{".hidden", ""},
	}
	for _, tt := range tests {
		if got := getExt(tt.in); got != tt.want {
			t.Errorf("getExt(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}

func TestToLowerContains(t *testing.T) {
	if got := toLower("AbC"); got != "abc" {
		t.Errorf("toLower = %q; want abc", got)
	}
	if !contains("hello world", "world") {
		t.Error("contains 應找到子字串")
	}
	if contains("abc", "z") {
		t.Error("contains 不應誤判")
	}
	if indexOf("abc", "b") != 1 {
		t.Errorf("indexOf = %d; want 1", indexOf("abc", "b"))
	}
	if indexOf("abc", "z") != -1 {
		t.Errorf("找不到時 indexOf = %d; want -1", indexOf("abc", "z"))
	}
}

func TestPerformSearch(t *testing.T) {
	m := New(t.TempDir())
	m.OriginalEntries = []types.FileEntry{
		{Name: "main.go"}, {Name: "README.md"}, {Name: "go.mod"},
	}
	m.SearchQuery = "go"
	m.performSearch()

	if len(m.Entries) != 2 {
		t.Fatalf("搜尋 go 應得 2 筆，got %d", len(m.Entries))
	}
	if len(m.SearchResults) != 2 {
		t.Errorf("SearchResults 長度 = %d; want 2", len(m.SearchResults))
	}
	if m.Cursor != 0 {
		t.Errorf("搜尋後 Cursor 應重置為 0，got %d", m.Cursor)
	}

	// 空查詢還原原始列表
	m.SearchQuery = ""
	m.performSearch()
	if len(m.Entries) != 3 {
		t.Errorf("清空查詢應還原 3 筆，got %d", len(m.Entries))
	}
}

func TestHandleOpen_Dir(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	m := New(tmp)
	m.Entries = []types.FileEntry{{Name: "sub", Path: sub, IsDir: true}}
	m.Cursor = 0

	m.HandleOpen()
	if m.CurrentPath != sub {
		t.Errorf("進入目錄後 CurrentPath = %q; want %q", m.CurrentPath, sub)
	}
	if m.Cursor != 0 {
		t.Errorf("進入目錄後 Cursor 應重置，got %d", m.Cursor)
	}
}

func TestHandleOpen_File(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "a.txt")
	if err := os.WriteFile(f, []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}
	m := New(tmp)
	m.Entries = []types.FileEntry{{Name: "a.txt", Path: f, IsDir: false}}
	m.Cursor = 0

	m.HandleOpen()
	if !m.PreviewActive {
		t.Error("開啟檔案應啟動預覽 (PreviewActive=true)")
	}
}

func TestHandleOpen_EmptySafe(t *testing.T) {
	m := New(t.TempDir())
	m.Entries = nil
	// 不應 panic
	m.HandleOpen()
}

func TestHandleBack(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	m := New(sub)
	m.HandleBack()
	if m.CurrentPath != tmp {
		t.Errorf("返回後 CurrentPath = %q; want %q", m.CurrentPath, tmp)
	}
}

func TestHandleInputSubmit_NewFile(t *testing.T) {
	tmp := t.TempDir()
	m := New(tmp)
	m.StatusMessage = "newfile"
	m.InputBuffer = "created.txt"
	m.Mode = ModeInput

	_, cmd := m.handleInputSubmit()
	m.Update(cmd())
	if _, err := os.Stat(filepath.Join(tmp, "created.txt")); err != nil {
		t.Errorf("應建立新檔案，err=%v", err)
	}
	if m.Mode != ModeNormal {
		t.Errorf("提交後 Mode = %v; want ModeNormal", m.Mode)
	}
}

func TestHandleInputSubmit_NewDir(t *testing.T) {
	tmp := t.TempDir()
	m := New(tmp)
	m.StatusMessage = "newdir"
	m.InputBuffer = "newfolder"
	m.Mode = ModeInput

	_, cmd := m.handleInputSubmit()
	m.Update(cmd())
	info, err := os.Stat(filepath.Join(tmp, "newfolder"))
	if err != nil || !info.IsDir() {
		t.Errorf("應建立新目錄，err=%v", err)
	}
}

func TestHandleInputSubmit_EmptyBuffer(t *testing.T) {
	m := New(t.TempDir())
	m.InputBuffer = ""
	m.Mode = ModeInput
	m.handleInputSubmit()
	if m.Mode != ModeNormal {
		t.Errorf("空輸入應直接回 Normal，got %v", m.Mode)
	}
}
