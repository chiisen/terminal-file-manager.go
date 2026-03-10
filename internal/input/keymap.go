package input

import (
	"fmt"
	"os"

	"gofm/internal/app"

	tea "github.com/charmbracelet/bubbletea"
)

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：Keymap (鍵位映射)
// 說明：定義鍵盤快捷鍵與操作的對應關係
// 為何使用：允許使用者自定義鍵位，提高效率
// ══════════════════════════════════════════════════════════════════════════════

// KeyAction 代表一個鍵位動作
type KeyAction string

// 預定義的鍵位動作
const (
	ActionUp      KeyAction = "up"
	ActionDown    KeyAction = "down"
	ActionOpen    KeyAction = "open"
	ActionBack    KeyAction = "back"
	ActionDelete  KeyAction = "delete"
	ActionRename  KeyAction = "rename"
	ActionCopy    KeyAction = "copy"
	ActionCut     KeyAction = "cut"
	ActionPaste   KeyAction = "paste"
	ActionNewFile KeyAction = "newfile"
	ActionNewDir  KeyAction = "newdir"
	ActionQuit    KeyAction = "quit"
	ActionSelect  KeyAction = "select"
)

// Keymap 儲存鍵位映射
type Keymap struct {
	Up      string
	Down    string
	Open    string
	Back    string
	Delete  string
	Rename  string
	Copy    string
	Cut     string
	Paste   string
	NewFile string
	NewDir  string
	Quit    string
	Select  string
}

// DefaultKeymap 回傳預設的鍵位映射
func DefaultKeymap() *Keymap {
	return &Keymap{
		Up:      "up",
		Down:    "down",
		Open:    "l",
		Back:    "h",
		Delete:  "d",
		Rename:  "r",
		Copy:    "y",
		Cut:     "x",
		Paste:   "p",
		NewFile: "a",
		NewDir:  "A",
		Quit:    "q",
		Select:  " ",
	}
}

// HandleKey 根據鍵位映射處理鍵盤輸入
// 參數 key 是按鍵字串
// 回傳對應的動作
func (km *Keymap) HandleKey(key string) KeyAction {
	switch key {
	case "up", km.Up:
		return ActionUp
	case "down", km.Down:
		return ActionDown
	case "enter", km.Open:
		return ActionOpen
	case km.Back:
		return ActionBack
	case km.Delete:
		return ActionDelete
	case km.Rename:
		return ActionRename
	case km.Copy:
		return ActionCopy
	case km.Cut:
		return ActionCut
	case km.Paste:
		return ActionPaste
	case km.NewFile:
		return ActionNewFile
	case km.NewDir:
		return ActionNewDir
	case "ctrl+c", km.Quit:
		return ActionQuit
	case km.Select:
		return ActionSelect
	default:
		return ""
	}
}

// ApplyAction 根據動作執行對應的操作
// 參數 m 是目前的 AppState
// 回傳更新後的 Model 和 Command
func (km *Keymap) ApplyAction(m *app.AppState, action KeyAction) (tea.Model, tea.Cmd) {
	switch action {
	case ActionUp:
		if m.Cursor > 0 {
			m.Cursor--
		}
	case ActionDown:
		if m.Cursor < len(m.Entries)-1 {
			m.Cursor++
		}
	case ActionOpen:
		return m.HandleOpen()
	case ActionBack:
		return m.HandleBack()
	case ActionDelete:
		if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
			m.SetMode(app.ModeConfirmDelete)
		}
	case ActionRename:
		if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
			m.SetMode(app.ModeInput)
			m.InputBuffer = m.Entries[m.Cursor].Name
		}
	case ActionCopy:
		if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
			m.Clipboard = m.Entries[m.Cursor].Path
			m.IsCut = false
			m.StatusMessage = "Copied: " + m.Entries[m.Cursor].Name
		}
	case ActionCut:
		if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
			m.Clipboard = m.Entries[m.Cursor].Path
			m.IsCut = true
			m.StatusMessage = "Cut: " + m.Entries[m.Cursor].Name
		}
	case ActionPaste:
		return m.HandlePaste()
	case ActionNewFile:
		m.SetMode(app.ModeInput)
		m.InputBuffer = ""
		m.StatusMessage = "newfile"
	case ActionNewDir:
		m.SetMode(app.ModeInput)
		m.InputBuffer = ""
		m.StatusMessage = "newdir"
	case ActionQuit:
		return m, tea.Quit
	case ActionSelect:
		if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
			path := m.Entries[m.Cursor].Path
			if m.Selected[path] {
				delete(m.Selected, path)
			} else {
				m.Selected[path] = true
			}
		}
	}
	return m, nil
}

// LoadKeymap 從配置檔載入鍵位映射
// 目前，這只是一個簡單的實現，未來可以擴展為讀取 config.toml
func LoadKeymap() *Keymap {
	// 檢查配置檔是否存在
	configPath := os.Getenv("HOME") + "/.config/gofm/config.toml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置檔不存在，使用預設值
		return DefaultKeymap()
	}

	// TODO: 實現從 TOML 檔案讀取配置
	// 目前只返回預設值
	return DefaultKeymap()
}

// String 回傳鍵位映射的字串表示
func (km *Keymap) String() string {
	return fmt.Sprintf("Keymap: up=%s down=%s open=%s back=%s delete=%s rename=%s copy=%s cut=%s paste=%s newfile=%s newdir=%s quit=%s select=%s",
		km.Up, km.Down, km.Open, km.Back, km.Delete, km.Rename,
		km.Copy, km.Cut, km.Paste, km.NewFile, km.NewDir, km.Quit, km.Select)
}
