package state

import (
	"gofm/internal/app"
	"gofm/internal/types"
)

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：State Manager
// 說明：負責管理應用程式的狀態，提供狀態轉換和查詢的統一介面
// 為何使用：將狀態管理邏輯分離出來，讓程式碼更具可維護性
// ══════════════════════════════════════════════════════════════════════════════

// Manager 負責管理應用程式的狀態
type Manager struct {
	// state 是目前的應用程式狀態
	state *app.AppState
}

// New 建立一個新的狀態管理器
func New() *Manager {
	return &Manager{
		state: app.New("."),
	}
}

// GetState 回傳目前的應用程式狀態
func (m *Manager) GetState() *app.AppState {
	return m.state
}

// SetPath 設定目前瀏覽的路徑
func (m *Manager) SetPath(path string) {
	m.state.CurrentPath = path
	m.state.Cursor = 0 // 重置游標位置
}

// SetEntries 設定目錄中的檔案列表
func (m *Manager) SetEntries(entries []types.FileEntry) {
	m.state.Entries = entries
}

// MoveCursorUp 將游標向上移動
func (m *Manager) MoveCursorUp() {
	if m.state.Cursor > 0 {
		m.state.Cursor--
	}
}

// MoveCursorDown 將游標向下移動
func (m *Manager) MoveCursorDown() {
	if m.state.Cursor < len(m.state.Entries)-1 {
		m.state.Cursor++
	}
}

// GetSelectedEntry 回傳目前選中的檔案
func (m *Manager) GetSelectedEntry() *types.FileEntry {
	if m.state.Cursor >= 0 && m.state.Cursor < len(m.state.Entries) {
		return &m.state.Entries[m.state.Cursor]
	}
	return nil
}

// ToggleSelected 切換檔案的選中狀態
func (m *Manager) ToggleSelected(path string) {
	if m.state.Selected[path] {
		delete(m.state.Selected, path)
	} else {
		m.state.Selected[path] = true
	}
}

// GetSelectedPaths 回傳所有選中的檔案路徑
func (m *Manager) GetSelectedPaths() []string {
	paths := make([]string, 0, len(m.state.Selected))
	for path := range m.state.Selected {
		paths = append(paths, path)
	}
	return paths
}
