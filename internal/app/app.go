package app

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"gofm/internal/fs"
	"gofm/internal/git"
	"gofm/internal/preview"
	"gofm/internal/types"
	"gofm/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：AppState (Model)
// 說明：存放應用程式的所有狀態資料，是 Bubble Tea 的核心資料結構
//       所有 UI 渲染都基於這個結構的內容
// 為何使用：集中管理狀態，方便追蹤與調試
// ══════════════════════════════════════════════════════════════════════════════

// AppMode 代表應用程式的不同操作模式
type AppMode int

const (
	ModeNormal        AppMode = iota // 一般導航模式
	ModeInput                        // 輸入模式（如重新命名）
	ModeConfirmDelete                // 確認刪除模式
	ModeSearch                       // 搜尋模式
)

// AppState 代表應用程式的完整狀態
type AppState struct {
	// CurrentPath 是目前瀏覽的目錄路徑
	CurrentPath string

	// Width 是終端機視窗的寬度
	Width int

	// Height 是終端機視窗的高度
	Height int

	// LastKeyTime 是上次按鍵的時間，用於防呆（避免按鈕沒放開連按）
	LastKeyTime time.Time

	// Entries 是目前目錄中的檔案列表
	Entries []types.FileEntry

	// Cursor 是目前選中的項目索引（從 0 開始）
	Cursor int

	// Selected 是已被選中的檔案 map（key 為檔案路徑）
	Selected map[string]bool

	// ErrorMessage 是顯示給使用者的錯誤訊息
	ErrorMessage string

	// StatusMessage 是顯示給使用者的狀態訊息
	StatusMessage string

	// Mode 是目前的操作模式
	Mode AppMode

	// InputBuffer 是輸入模式下的文字緩衝區
	InputBuffer string

	// Clipboard 是剪貼簿（用於複製/貼上）
	Clipboard string

	// IsCut 是否為剪下模式（而非複製）
	IsCut bool

	// Search 相關欄位
	SearchQuery     string            // 搜尋關鍵字
	SearchResults   []int             // 搜尋結果的索引列表
	OriginalEntries []types.FileEntry // 搜尋前的原始列表

	// Sort 相關欄位
	SortBy  string // 排序方式: "name", "size", "modified", "type"
	SortAsc bool   // 是否升序排列

	// 預覽狀態
	PreviewActive bool

	// Git 相關欄位
	GitInfo *git.GitInfo // Git 倉庫資訊
}

// New 建立並回傳一個新的 AppState
// 參數 startPath 指定起始瀏覽的目錄路徑
func New(startPath string) *AppState {
	// 解析為絕對路徑
	absPath, _ := fs.GetAbsolutePath(startPath)

	return &AppState{
		CurrentPath:   absPath,
		Width:        80,  // 預設寬度
		Height:       24,  // 預設高度
		Entries:       []types.FileEntry{},
		Cursor:        0,
		Selected:      make(map[string]bool),
		ErrorMessage:  "",
		StatusMessage: "",
		Mode:          ModeNormal,
		InputBuffer:   "",
		Clipboard:     "",
		IsCut:         false,
		SortBy:        "name",
		SortAsc:       true,
		PreviewActive: false,
	}
}

// SetMode 設定應用程式的模式
func (m *AppState) SetMode(mode AppMode) {
	m.Mode = mode
}

// Init 是 Bubble Tea 的生命週期方法
// 用於初始化程式並回傳初始命令（通常是 nil，表示不執行額外命令）
func (m *AppState) Init() tea.Cmd {
	// 載入目錄內容
	return m.loadDirectory
}

// Update 處理輸入事件並回傳新的 Model 和 Command
// 參數 msg 是發生的事件（如鍵盤輸入）
func (m *AppState) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 清除錯誤訊息
	m.ErrorMessage = ""

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// 自動填滿視窗：更新寬度和高度
		m.Width = msg.Width
		m.Height = msg.Height
	case tea.KeyMsg:
		// 防呆檢查：避免按鈕沒放開連按（150ms 內不處理重複按鍵）
		now := time.Now()
		if now.Sub(m.LastKeyTime) < 150*time.Millisecond {
			return m, nil
		}
		m.LastKeyTime = now

		switch m.Mode {
		case ModeNormal:
			return m.handleNormalMode(msg)
		case ModeInput:
			return m.handleInputMode(msg)
		case ModeConfirmDelete:
			return m.handleConfirmDeleteMode(msg)
		case ModeSearch:
			return m.handleSearchMode(msg)
		}
	}
	return m, nil
}

// handleNormalMode 處理一般導航模式的鍵盤輸入
func (m *AppState) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "Q":
		return m, tea.Quit

	case "esc":
		if m.PreviewActive {
			m.PreviewActive = false
		}

	// 檔案導航
	case "up", "k":
		if len(m.Entries) > 0 && m.Cursor > 0 {
			m.Cursor--
			m.PreviewActive = false
		}
	case "down", "j":
		if len(m.Entries) > 0 && m.Cursor < len(m.Entries)-1 {
			m.Cursor++
			m.PreviewActive = false
		}

	// 進入目錄 (Enter, l 或 right)
	case "enter", "l", "right":
		m.StatusMessage = "Opening..."
		return m.handleOpen()

	// 返回上一層 (h 或 left)
	case "h", "left":
		return m.handleBack()

	// 選取/取消選取檔案 (space)
	case " ":
		if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
			entry := m.Entries[m.Cursor]
			if m.Selected[entry.Path] {
				delete(m.Selected, entry.Path)
				m.StatusMessage = "Unselected: " + entry.Name
			} else {
				m.Selected[entry.Path] = true
				m.StatusMessage = "Selected: " + entry.Name
			}
		}

	// 檔案操作
	case "d": // 刪除
		if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
			m.Mode = ModeConfirmDelete
		}
	case "r": // 重新命名
		if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
			m.Mode = ModeInput
			m.InputBuffer = m.Entries[m.Cursor].Name
		}
	case "y": // 複製
		if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
			m.Clipboard = m.Entries[m.Cursor].Path
			m.IsCut = false
			m.StatusMessage = "Copied: " + m.Entries[m.Cursor].Name
		}
	case "x": // 剪下
		if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
			m.Clipboard = m.Entries[m.Cursor].Path
			m.IsCut = true
			m.StatusMessage = "Cut: " + m.Entries[m.Cursor].Name
		}
	case "p": // 貼上
		return m.handlePaste()

	case "a": // 新增檔案
		m.Mode = ModeInput
		m.InputBuffer = ""
		m.StatusMessage = "newfile" // 標記為新增檔案模式
	case "A": // 新增目錄
		m.Mode = ModeInput
		m.InputBuffer = ""
		m.StatusMessage = "newdir" // 標記為新增目錄模式

	// 搜尋模式
	case "/": // 進入搜尋模式
		m.OriginalEntries = m.Entries // 儲存原始列表
		m.SearchQuery = ""
		m.SearchResults = []int{}
		m.Mode = ModeSearch

	// 排序功能
	case "s": // 切換排序方向
		m.SortAsc = !m.SortAsc
		m.SortEntries()
		m.StatusMessage = fmt.Sprintf("Sorted by %s (%s)", m.SortBy, map[bool]string{true: "asc", false: "desc"}[m.SortAsc])
	case "S": // 切換排序方式
		m.cycleSortBy()
		m.SortEntries()
		m.StatusMessage = fmt.Sprintf("Sorted by %s (%s)", m.SortBy, map[bool]string{true: "asc", false: "desc"}[m.SortAsc])
	}
	return m, nil
}

// handleInputMode 處理輸入模式的鍵盤輸入
func (m *AppState) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.handleInputSubmit()
	case "esc":
		m.Mode = ModeNormal
		m.InputBuffer = ""
	case "backspace":
		if len(m.InputBuffer) > 0 {
			m.InputBuffer = m.InputBuffer[:len(m.InputBuffer)-1]
		}
	default:
		// 處理一般字元輸入
		if len(msg.String()) == 1 {
			m.InputBuffer += msg.String()
		}
	}
	return m, nil
}

// handleConfirmDeleteMode 處理確認刪除模式的鍵盤輸入
func (m *AppState) handleConfirmDeleteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		return m.handleDeleteConfirm()
	case "n", "esc":
		m.Mode = ModeNormal
	}
	return m, nil
}

// handleSearchMode 處理搜尋模式的鍵盤輸入
func (m *AppState) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc": // 退出搜尋
		m.Entries = m.OriginalEntries
		m.OriginalEntries = nil
		m.SearchQuery = ""
		m.SearchResults = nil
		m.Mode = ModeNormal
		m.Cursor = 0

	case "enter": // 確認搜尋，停留在搜尋結果
		if len(m.SearchResults) > 0 && m.Cursor < len(m.SearchResults) {
			// 移動到第一個搜尋結果
			m.Cursor = 0
		}

	case "backspace": // 刪除字元
		if len(m.SearchQuery) > 0 {
			m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
			m.performSearch()
		}

	case "up", "k": // 在搜尋結果中移動
		if m.Cursor > 0 {
			m.Cursor--
			m.PreviewActive = false
		}

	case "down", "j": // 在搜尋結果中移動
		if m.Cursor < len(m.SearchResults)-1 {
			m.Cursor++
			m.PreviewActive = false
		}

	default:
		// 處理一般字元輸入
		if len(msg.String()) == 1 {
			m.SearchQuery += msg.String()
			m.performSearch()
		}
	}
	return m, nil
}

// performSearch 執行 fuzzy search
func (m *AppState) performSearch() {
	if m.SearchQuery == "" {
		m.SearchResults = nil
		m.Entries = m.OriginalEntries
		return
	}

	// Fuzzy search: 找名稱包含搜尋關鍵字的項目
	results := []int{}
	query := toLower(m.SearchQuery)

	for i, entry := range m.OriginalEntries {
		if contains(toLower(entry.Name), query) {
			results = append(results, i)
		}
	}

	m.SearchResults = results

	// 更新顯示的項目為搜尋結果
	if len(results) > 0 {
		m.Entries = make([]types.FileEntry, len(results))
		for i, idx := range results {
			m.Entries[i] = m.OriginalEntries[idx]
		}
		m.Cursor = 0
	} else {
		m.Entries = []types.FileEntry{}
	}
}

// toLower 轉換字串為小寫（輔助函數）
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}

// contains 檢查字串 s 是否包含 sub
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (len(sub) == 0 || indexOf(s, sub) >= 0)
}

// indexOf 找尋 sub 在 s 中的位置
func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// SortEntries 根據目前的排序設定對目錄項目進行排序
func (m *AppState) SortEntries() {
	if m.Entries == nil {
		return
	}

	// 複製切片以避免修改原始資料
	entries := make([]types.FileEntry, len(m.Entries))
	copy(entries, m.Entries)

	switch m.SortBy {
	case "name":
		sort.SliceStable(entries, func(i, j int) bool {
			// 目錄優先
			if entries[i].IsDir != entries[j].IsDir {
				return entries[i].IsDir
			}
			if m.SortAsc {
				return entries[i].Name < entries[j].Name
			}
			return entries[i].Name > entries[j].Name
		})
	case "size":
		sort.SliceStable(entries, func(i, j int) bool {
			if entries[i].IsDir != entries[j].IsDir {
				return entries[i].IsDir
			}
			if m.SortAsc {
				return entries[i].Size < entries[j].Size
			}
			return entries[i].Size > entries[j].Size
		})
	case "type":
		sort.SliceStable(entries, func(i, j int) bool {
			if entries[i].IsDir != entries[j].IsDir {
				return entries[i].IsDir
			}
			// 根據副檔名排序
			ext1 := getExt(entries[i].Name)
			ext2 := getExt(entries[j].Name)
			if m.SortAsc {
				return ext1 < ext2
			}
			return ext1 > ext2
		})
	case "modified":
		// 需要讀取修改時間，這裡暫時按名稱排序
		fallthrough
	default:
		sort.SliceStable(entries, func(i, j int) bool {
			if entries[i].IsDir != entries[j].IsDir {
				return entries[i].IsDir
			}
			if m.SortAsc {
				return entries[i].Name < entries[j].Name
			}
			return entries[i].Name > entries[j].Name
		})
	}

	m.Entries = entries
}

// cycleSortBy 循環切換排序方式
func (m *AppState) cycleSortBy() {
	switch m.SortBy {
	case "name":
		m.SortBy = "size"
	case "size":
		m.SortBy = "type"
	case "type":
		m.SortBy = "modified"
	case "modified":
		m.SortBy = "name"
	default:
		m.SortBy = "name"
	}
}

// getExt 取得檔案的副檔名
func getExt(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			if i > 0 {
				return filename[i:]
			}
			return ""
		}
		if filename[i] == '/' || filename[i] == '\\' {
			return ""
		}
	}
	return ""
}

// handleInputSubmit 處理輸入模式下的 Enter 鍵
func (m *AppState) handleInputSubmit() (tea.Model, tea.Cmd) {
	if m.InputBuffer == "" {
		m.Mode = ModeNormal
		return m, nil
	}

	// 根據上一個操作的上下文來決定要執行的操作
	// 這裡我們需要記住是新增檔案還是重新命名
	// 暫時用 StatusMessage 來傳遞上下文
	if m.StatusMessage == "newfile" {
		// 新增檔案
		newPath := filepath.Join(m.CurrentPath, m.InputBuffer)
		err := fs.CreateFile(newPath)
		if err != nil {
			m.ErrorMessage = err.Error()
		} else {
			m.StatusMessage = "Created file: " + m.InputBuffer
		}
	} else if m.StatusMessage == "newdir" {
		// 新增目錄
		newPath := filepath.Join(m.CurrentPath, m.InputBuffer)
		err := fs.CreateDirectory(newPath)
		if err != nil {
			m.ErrorMessage = err.Error()
		} else {
			m.StatusMessage = "Created directory: " + m.InputBuffer
		}
	} else if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
		// 重新命名
		oldPath := m.Entries[m.Cursor].Path
		err := fs.RenameFile(oldPath, m.InputBuffer)
		if err != nil {
			m.ErrorMessage = err.Error()
		} else {
			m.StatusMessage = "Renamed to: " + m.InputBuffer
		}
	}

	m.Mode = ModeNormal
	m.InputBuffer = ""
	return m, m.loadDirectory
}

// handleDeleteConfirm 處理確認刪除
func (m *AppState) handleDeleteConfirm() (tea.Model, tea.Cmd) {
	if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
		entry := m.Entries[m.Cursor]
		err := fs.DeleteFile(entry.Path)
		if err != nil {
			m.ErrorMessage = err.Error()
		} else {
			m.StatusMessage = "Deleted: " + entry.Name
		}
	}

	m.Mode = ModeNormal
	return m, m.loadDirectory
}

// HandleOpen 處理進入目錄或開啟檔案的操作（公開版本）
func (m *AppState) HandleOpen() (tea.Model, tea.Cmd) {
	return m.handleOpen()
}

// handleOpen 處理進入目錄或開啟檔案的操作
func (m *AppState) handleOpen() (tea.Model, tea.Cmd) {
	// 防禦性檢查：確保有項目且游標在範圍內
	if len(m.Entries) == 0 || m.Cursor < 0 || m.Cursor >= len(m.Entries) {
		return m, nil
	}

	entry := m.Entries[m.Cursor]

	if entry.IsDir {
		// 進入子目錄（保持選取狀態）
		m.CurrentPath = entry.Path
		m.Cursor = 0
		m.PreviewActive = false
		return m, m.loadDirectory
	}

	// 如果是檔案，啟動預覽並顯示訊息
	m.PreviewActive = true
	m.StatusMessage = "Previewing: " + entry.Name
	return m, nil
}

// HandleBack 返回上一層目錄（公開版本）
func (m *AppState) HandleBack() (tea.Model, tea.Cmd) {
	return m.handleBack()
}

// handleBack 返回上一層目錄
func (m *AppState) handleBack() (tea.Model, tea.Cmd) {
	parent := fs.GetParentDirectory(m.CurrentPath)

	// 確保不會超出根目錄
	if parent == m.CurrentPath {
		m.ErrorMessage = "Already at root"
		return m, nil
	}

	m.CurrentPath = parent
	m.Cursor = 0
	m.PreviewActive = false
	return m, m.loadDirectory
}

// HandlePaste 處理貼上操作（公開版本）
func (m *AppState) HandlePaste() (tea.Model, tea.Cmd) {
	return m.handlePaste()
}

// handlePaste 處理貼上操作
func (m *AppState) handlePaste() (tea.Model, tea.Cmd) {
	if m.Clipboard == "" {
		m.ErrorMessage = "Clipboard is empty"
		return m, nil
	}

	// 取得檔案名稱
	filename := filepath.Base(m.Clipboard)
	destPath := filepath.Join(m.CurrentPath, filename)

	// 檢查目標是否已存在
	if fs.FileExists(destPath) {
		m.ErrorMessage = "File already exists: " + filename
		return m, nil
	}

	// 執行複製或移動
	var err error
	if m.IsCut {
		err = fs.MoveFile(m.Clipboard, destPath)
		m.Clipboard = ""
		m.IsCut = false
	} else {
		err = fs.CopyFile(m.Clipboard, destPath)
	}

	if err != nil {
		m.ErrorMessage = err.Error()
	} else {
		if m.IsCut {
			m.StatusMessage = "Moved: " + filename
		} else {
			m.StatusMessage = "Copied: " + filename
		}
	}

	return m, m.loadDirectory
}

// loadDirectory 載入目前目錄的檔案列表
// 這是一個非同步命令，載入完成後會發送目錄載入完成的訊息
// 使用 Lazy Load 策略：先快速顯示，再非同步載入詳細資訊
func (m *AppState) loadDirectory() tea.Msg {
	// 首先快速載入目錄名稱（延遲載入策略）
	entries, err := fs.LazyReadDirectory(m.CurrentPath)
	if err != nil {
		m.ErrorMessage = "Error loading directory: " + err.Error()
		return nil
	}

	m.Entries = entries

	// 回傳一個非同步命令來載入詳細資訊（大小、權限等）
	// 這讓 UI 可以立即顯示目錄，然後在背景載入詳細資訊
	return m.loadMetadata
}

// loadMetadata 非同步載入檔案的詳細資訊（大小、權限等）
func (m *AppState) loadMetadata() tea.Msg {
	entries, err := fs.ReadDirectory(m.CurrentPath)
	if err != nil {
		// 載入失敗，但目錄已顯示，這裡不顯示錯誤
		return nil
	}

	// 更新現有的項目（保留已選擇的項目）
	for i := range m.Entries {
		if i < len(entries) {
			m.Entries[i].Size = entries[i].Size
			m.Entries[i].Mode = entries[i].Mode
			m.Entries[i].IsSymlink = entries[i].IsSymlink
			m.Entries[i].SymlinkPath = entries[i].SymlinkPath
			m.Entries[i].IsBroken = entries[i].IsBroken
			m.Entries[i].Permission = entries[i].Permission
		}
	}

	// 載入 Git 資訊（在背景執行，不阻塞 UI）
	m.loadGitInfo()

	return nil
}

// loadGitInfo 載入 Git 資訊
func (m *AppState) loadGitInfo() {
	info, err := git.GetGitInfo(m.CurrentPath)
	if err != nil {
		// Git 載入失敗不是嚴重錯誤，只是沒有狀態顯示
		return
	}
	m.GitInfo = info
}

// View 回傳目前狀態的 UI 渲染結果
// 這個方法會在每次狀態更新後被呼叫
func (m *AppState) View() string {
	// 自動填滿視窗：計算可用寬度
	// 路徑列使用完整寬度
	pathBar := ui.RenderPathBar(m.CurrentPath, m.Width)

	// 如果是 Git 倉庫，添加 Git 狀態資訊
	var gitStatusInfo string
	if m.GitInfo != nil && m.GitInfo.IsRepo {
		// 計算有多少檔案有變更
		changedCount := 0
		for range m.GitInfo.StatusMap {
			changedCount++
		}
		if changedCount > 0 {
			gitStatusInfo = fmt.Sprintf(" [Git: %d changes]", changedCount)
		} else {
			gitStatusInfo = " [Git: clean]"
		}
	}

	// 自動填滿視窗：計算檔案列表和預覽區域的寬度
	// 預覽區域佔 40%，檔案列表佔 60%
	previewWidth := m.Width * 40 / 100
	if previewWidth < 30 {
		previewWidth = 30 // 最小寬度
	}
	fileListWidth := m.Width - previewWidth - 1 // -1 為分隔線
	if fileListWidth < 30 {
		fileListWidth = 30
	}

	// 渲染檔案列表
	var fileList string
	if len(m.Entries) == 0 {
		fileList = "(empty directory)"
	} else {
		fileList = ui.RenderFileList(m.Entries, m.Cursor, m.Selected, fileListWidth, m.Height)
	}

	// 渲染預覽面板
	var previewContent string
	if len(m.Entries) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Entries) {
		entry := m.Entries[m.Cursor]
		if entry.IsDir {
			previewContent = "Directory\n\n" + entry.Name
		} else if !m.PreviewActive {
			previewContent = "Preview\n\nPress Enter to view contents.\nPress Esc to hide."
		} else {
			p, err := preview.GetPreview(entry.Path)
			if err != nil {
				previewContent = fmt.Sprintf("Error: %v", err)
			} else {
				previewContent = p.Content
			}
		}
	} else {
		previewContent = "Preview\n\nSelect a file\nto preview"
	}
	preview := ui.PreviewStyle.Render(previewContent)

	// 組合主要區域
	mainContent := fmt.Sprintf("%s\n%s", fileList, preview)

	// 底部狀態列
	var statusBar string
	switch m.Mode {
	case ModeNormal:
		sortIndicator := fmt.Sprintf(" [%s %s]", m.SortBy, map[bool]string{true: "↑", false: "↓"}[m.SortAsc])
		statusBar = "↑↓/kj: nav  Enter/l: open  ←/h: parent  space: select  d: delete  r: rename  y: copy  x: cut  p: paste  a: new file  A: new dir  /: search  s: sort order  S: sort by" + sortIndicator + "  ctrl+c: quit"
	case ModeInput:
		switch m.StatusMessage {
		case "newfile":
			statusBar = "New File - Enter: confirm  Esc: cancel  |  Input: " + m.InputBuffer + "_"
		case "newdir":
			statusBar = "New Directory - Enter: confirm  Esc: cancel  |  Input: " + m.InputBuffer + "_"
		default:
			statusBar = "Rename - Enter: confirm  Esc: cancel  |  Input: " + m.InputBuffer + "_"
		}
	case ModeConfirmDelete:
		if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
			statusBar = fmt.Sprintf("Delete %s? [y/n]", m.Entries[m.Cursor].Name)
		}
	case ModeSearch:
		resultCount := len(m.SearchResults)
		statusBar = fmt.Sprintf("Search: %s_%s", m.SearchQuery, fmt.Sprintf(" [%d results, ↑/↓ to move, Enter to select, Esc to exit]", resultCount))
	}

	// 組合輸出
	output := pathBar + gitStatusInfo + "\n\n"
	output += mainContent + "\n\n"
	output += ui.RenderStatusBar(statusBar, m.Width)

	// 顯示錯誤/狀態訊息
	if m.ErrorMessage != "" {
		output += "\n" + ui.RenderError(m.ErrorMessage)
	}
	if m.StatusMessage != "" && m.Mode == ModeNormal && m.StatusMessage != "newfile" && m.StatusMessage != "newdir" {
		output += "\n" + ui.RenderStatusMessage(m.StatusMessage)
	}

	return output
}
