package input

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
// 說明：讀取 ~/.config/gofm/config.toml 的 [keymap] 區段，未指定或缺檔時用預設值
// 為何如此：避免缺檔或格式錯誤導致當機，保證鍵盤操作永遠可用
func LoadKeymap() *Keymap {
	// 💡 概念：os.UserHomeDir
	// 說明：跨平台取得家目錄（Windows 用 USERPROFILE，Unix 用 HOME）
	// 為何使用：原本只讀 HOME 在 Windows 會失效
	home, err := os.UserHomeDir()
	if err != nil {
		return DefaultKeymap()
	}
	configPath := filepath.Join(home, ".config", "gofm", "config.toml")
	return LoadKeymapFromPath(configPath)
}

// LoadKeymapFromPath 從指定路徑載入鍵位映射（方便測試）
// 參數 path 是 config.toml 路徑
// 檔案不存在或解析失敗時回傳預設值
func LoadKeymapFromPath(path string) *Keymap {
	km := DefaultKeymap()

	// 檔案不存在直接用預設值
	data, err := os.ReadFile(path)
	if err != nil {
		return km
	}

	// 解析 [keymap] 區段並覆寫預設值
	overrides := parseKeymapTOML(string(data))
	applyKeymapOverrides(km, overrides)
	return km
}

// parseKeymapTOML 解析 TOML 文字中的 [keymap] 區段
// 說明：只用標準函式庫實作的最小解析器，支援 key = "value"、註解與單雙引號
// 為何如此：config 格式單純，不需為此引入第三方 TOML 依賴
func parseKeymapTOML(content string) map[string]string {
	result := make(map[string]string)
	inKeymap := false

	for _, rawLine := range strings.Split(content, "\n") {
		// 去除前後空白與 \r（Windows 換行）
		line := strings.TrimSpace(strings.TrimSuffix(rawLine, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 區段標頭，例如 [keymap]
		if strings.HasPrefix(line, "[") {
			section := strings.ToLower(strings.TrimSpace(line))
			inKeymap = section == "[keymap]"
			continue
		}
		if !inKeymap {
			continue
		}

		// 只處理 key = value 形式
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(line[:idx]))
		value := strings.TrimSpace(line[idx+1:])

		// 去除行尾註解（不在引號內才算註解）
		value = stripInlineComment(value)
		value = strings.TrimSpace(value)
		// 去除前後單雙引號
		value = unquote(value)
		if key == "" || value == "" {
			continue
		}
		result[key] = value
	}
	return result
}

// stripInlineComment 去除 value 後方的行內註解（# 開頭，且不在引號內）
func stripInlineComment(s string) string {
	var quote rune
	for i, r := range s {
		switch {
		case quote != 0 && r == quote:
			quote = 0
		case quote == 0 && (r == '"' || r == '\''):
			quote = r
		case quote == 0 && r == '#':
			return s[:i]
		}
	}
	return s
}

// unquote 去除字串前後成對的單引號或雙引號
func unquote(s string) string {
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// applyKeymapOverrides 將解析出的鍵位覆寫到 Keymap，未指定者維持預設
func applyKeymapOverrides(km *Keymap, overrides map[string]string) {
	// 💡 概念：map 查表覆寫
	// 說明：用輔助函式避免重複的 if 判斷
	// 為何使用：13 個欄位逐一處理時保持程式碼簡潔
	set := func(key, current string) string {
		if v, ok := overrides[key]; ok && v != "" {
			return v
		}
		return current
	}

	km.Up = set("up", km.Up)
	km.Down = set("down", km.Down)
	km.Open = set("open", km.Open)
	km.Back = set("back", km.Back)
	km.Delete = set("delete", km.Delete)
	km.Rename = set("rename", km.Rename)
	km.Copy = set("copy", km.Copy)
	km.Cut = set("cut", km.Cut)
	km.Paste = set("paste", km.Paste)
	km.NewFile = set("newfile", km.NewFile)
	km.NewDir = set("newdir", km.NewDir)
	km.Quit = set("quit", km.Quit)
	km.Select = set("select", km.Select)
}

// String 回傳鍵位映射的字串表示
func (km *Keymap) String() string {
	return fmt.Sprintf("Keymap: up=%s down=%s open=%s back=%s delete=%s rename=%s copy=%s cut=%s paste=%s newfile=%s newdir=%s quit=%s select=%s",
		km.Up, km.Down, km.Open, km.Back, km.Delete, km.Rename,
		km.Copy, km.Cut, km.Paste, km.NewFile, km.NewDir, km.Quit, km.Select)
}
