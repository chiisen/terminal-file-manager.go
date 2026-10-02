package app

import (
	"fmt"
	"path/filepath"
	"sort"
	"unicode/utf8"

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

	// 載入編號可辨識同一路徑的不同請求，避免舊結果覆蓋新狀態。
	loadID                      uint64
	previewID                   uint64
	previewPath, previewContent string
	previewLoading              bool
	previewCacheEntry           types.FileEntry
	previewCacheContent         string
	previewCacheValid           bool
	operationID                 uint64
	operationBusy               bool
	operationProgress           string
}

// 背景命令只回傳資料，所有畫面狀態由 Update 在事件迴圈內更新。
type directoryLoadedMsg struct {
	id      uint64
	path    string
	entries []types.FileEntry
	err     error
}

type metadataLoadedMsg struct {
	id      uint64
	path    string
	entries []types.FileEntry
	err     error
}

type gitLoadedMsg struct {
	id   uint64
	path string
	info *git.GitInfo
	err  error
}

// New 建立並回傳一個新的 AppState
// 參數 startPath 指定起始瀏覽的目錄路徑
func New(startPath string) *AppState {
	// 解析為絕對路徑
	absPath, _ := fs.GetAbsolutePath(startPath)

	return &AppState{
		CurrentPath:   absPath,
		Width:         80, // 預設寬度
		Height:        24, // 預設高度
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
	return m.loadDirectory()
}

// Update 處理輸入事件並回傳新的 Model 和 Command
// 參數 msg 是發生的事件（如鍵盤輸入）
func (m *AppState) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case previewLoadedMsg:
		m.applyPreview(msg)
	case operationFinishedMsg:
		return m, m.finishOperation(msg)
	case directoryLoadedMsg:
		if msg.id != m.loadID || msg.path != m.CurrentPath {
			return m, nil
		}
		if msg.err != nil {
			m.ErrorMessage = "Error loading directory: " + msg.err.Error()
			return m, nil
		}
		m.Entries = msg.entries
		m.SortEntries()
		if m.Mode == ModeSearch {
			m.OriginalEntries = m.Entries
			m.performSearch()
		}
		return m, loadMetadata(msg.path, msg.id)
	case metadataLoadedMsg:
		if msg.id != m.loadID || msg.path != m.CurrentPath {
			return m, nil
		}
		if msg.err == nil {
			// 用路徑匹配而非索引，避免排序或目錄異動造成 metadata 錯配。
			byPath := make(map[string]types.FileEntry, len(msg.entries))
			for _, entry := range msg.entries {
				byPath[entry.Path] = entry
			}
			// 搜尋列表與原始列表都需更新，Esc 還原時才不會遺失 metadata。
			for _, entries := range [][]types.FileEntry{m.Entries, m.OriginalEntries} {
				for i, entry := range entries {
					if metadata, ok := byPath[entry.Path]; ok {
						entries[i] = metadata
					}
				}
			}
			var selectedPath string
			if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
				selectedPath = m.Entries[m.Cursor].Path
			}
			m.SortEntries()
			for i, entry := range m.Entries {
				if entry.Path == selectedPath {
					m.Cursor = i
					break
				}
			}
		}
		return m, loadGitInfo(msg.path, msg.id)
	case gitLoadedMsg:
		if msg.id == m.loadID && msg.path == m.CurrentPath && msg.err == nil {
			m.GitInfo = msg.info
		}
	case tea.WindowSizeMsg:
		// 自動填滿視窗：更新寬度和高度
		m.Width = msg.Width
		m.Height = msg.Height
	case tea.KeyMsg:
		m.ErrorMessage = ""

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
	if m.operationBusy {
		switch msg.String() {
		case "d", "r", "p", "a", "A":
			return m, nil
		}
	}
	switch msg.String() {
	case "ctrl+c", "q", "Q":
		return m, tea.Quit

	case "esc":
		if m.PreviewActive {
			m.hidePreview()
		}

	// 檔案導航
	case "up", "k":
		if len(m.Entries) > 0 && m.Cursor > 0 {
			m.Cursor--
			m.hidePreview()
		}
	case "down", "j":
		if len(m.Entries) > 0 && m.Cursor < len(m.Entries)-1 {
			m.Cursor++
			m.hidePreview()
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
		m.hidePreview()
		m.SortAsc = !m.SortAsc
		m.SortEntries()
		m.StatusMessage = fmt.Sprintf("Sorted by %s (%s)", m.SortBy, map[bool]string{true: "asc", false: "desc"}[m.SortAsc])
	case "S": // 切換排序方式
		m.hidePreview()
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
			_, size := utf8.DecodeLastRuneInString(m.InputBuffer)
			m.InputBuffer = m.InputBuffer[:len(m.InputBuffer)-size]
		}
	default:
		// 處理一般字元輸入
		if msg.Type == tea.KeyRunes {
			m.InputBuffer += string(msg.Runes)
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
			_, size := utf8.DecodeLastRuneInString(m.SearchQuery)
			m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-size]
			m.performSearch()
		}

	case "up": // 在搜尋結果中移動
		if m.Cursor > 0 {
			m.Cursor--
			m.hidePreview()
		}

	case "down": // 在搜尋結果中移動
		if m.Cursor < len(m.SearchResults)-1 {
			m.Cursor++
			m.hidePreview()
		}

	default:
		// 處理一般字元輸入
		if msg.Type == tea.KeyRunes {
			m.SearchQuery += string(msg.Runes)
			m.performSearch()
		}
	}
	return m, nil
}

// performSearch 執行 fuzzy search
func (m *AppState) performSearch() {
	m.hidePreview()
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
	if m.operationBusy {
		return m, nil
	}
	if m.InputBuffer == "" {
		m.Mode = ModeNormal
		return m, nil
	}

	// 根據上一個操作的上下文來決定要執行的操作
	// 這裡我們需要記住是新增檔案還是重新命名
	// 暫時用 StatusMessage 來傳遞上下文
	name := m.InputBuffer
	var cmd tea.Cmd
	if m.StatusMessage == "newfile" {
		// 新增檔案
		newPath := filepath.Join(m.CurrentPath, m.InputBuffer)
		cmd = m.startOperation("Creating file: "+name, "Created file: "+name, "", false, func() error { return fs.CreateFile(newPath) })
	} else if m.StatusMessage == "newdir" {
		// 新增目錄
		newPath := filepath.Join(m.CurrentPath, m.InputBuffer)
		cmd = m.startOperation("Creating directory: "+name, "Created directory: "+name, "", false, func() error { return fs.CreateDirectory(newPath) })
	} else if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
		// 重新命名
		oldPath := m.Entries[m.Cursor].Path
		cmd = m.startOperation("Renaming: "+name, "Renamed to: "+name, oldPath, false, func() error { return fs.RenameFile(oldPath, name) })
	}

	m.Mode = ModeNormal
	m.InputBuffer = ""
	return m, cmd
}

// handleDeleteConfirm 處理確認刪除
func (m *AppState) handleDeleteConfirm() (tea.Model, tea.Cmd) {
	if m.operationBusy {
		return m, nil
	}
	var cmd tea.Cmd
	if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
		entry := m.Entries[m.Cursor]
		cmd = m.startOperation("Deleting: "+entry.Name, "Deleted: "+entry.Name, entry.Path, false, func() error { return fs.DeleteFile(entry.Path) })
	}

	m.Mode = ModeNormal
	return m, cmd
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
		m.hidePreview()
		return m, m.loadDirectory()
	}

	// 如果是檔案，啟動預覽並顯示訊息
	m.StatusMessage = "Previewing: " + entry.Name
	return m, m.startPreview(entry, preview.GetPreview)
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
	m.hidePreview()
	return m, m.loadDirectory()
}

// HandlePaste 處理貼上操作（公開版本）
func (m *AppState) HandlePaste() (tea.Model, tea.Cmd) {
	return m.handlePaste()
}

// handlePaste 處理貼上操作
func (m *AppState) handlePaste() (tea.Model, tea.Cmd) {
	if m.operationBusy {
		return m, nil
	}
	if m.Clipboard == "" {
		m.ErrorMessage = "Clipboard is empty"
		return m, nil
	}

	// 取得檔案名稱
	filename := filepath.Base(m.Clipboard)
	destPath := filepath.Join(m.CurrentPath, filename)

	source, isCut := m.Clipboard, m.IsCut
	progress, success := "Copying: "+filename, "Copied: "+filename
	if isCut {
		progress, success = "Moving: "+filename, "Moved: "+filename
	}
	return m, m.startOperation(progress, success, source, isCut, func() error {
		if fs.FileExists(destPath) {
			return fmt.Errorf("File already exists: %s", filename)
		}
		if isCut {
			return fs.MoveFile(source, destPath)
		}
		return fs.CopyFile(source, destPath)
	})
}

// loadDirectory 載入目前目錄的檔案列表
// 這是一個非同步命令，載入完成後會發送目錄載入完成的訊息
// 使用 Lazy Load 策略：先快速顯示，再非同步載入詳細資訊
func (m *AppState) loadDirectory() tea.Cmd {
	m.loadID++
	id, path := m.loadID, m.CurrentPath
	// 載入期間清除舊項目，避免新路徑下仍能操作上一個目錄的檔案。
	m.Entries = nil
	m.Cursor = 0
	m.hidePreview()
	m.previewCacheValid = false
	m.GitInfo = nil
	m.OriginalEntries = nil
	m.SearchResults = nil
	m.SearchQuery = ""
	return func() tea.Msg {
		entries, err := fs.LazyReadDirectory(path)
		return directoryLoadedMsg{id: id, path: path, entries: entries, err: err}
	}
}

// loadMetadata 非同步載入檔案的詳細資訊（大小、權限等）
func loadMetadata(path string, id uint64) tea.Cmd {
	return func() tea.Msg {
		entries, err := fs.ReadDirectory(path)
		return metadataLoadedMsg{id: id, path: path, entries: entries, err: err}
	}
}

// loadGitInfo 載入 Git 資訊
func loadGitInfo(path string, id uint64) tea.Cmd {
	return func() tea.Msg {
		info, err := git.GetGitInfo(path)
		return gitLoadedMsg{id: id, path: path, info: info, err: err}
	}
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
		} else if !m.PreviewActive || m.previewPath != entry.Path {
			previewContent = "Preview\n\nPress Enter to view contents.\nPress Esc to hide."
		} else if m.previewLoading {
			previewContent = "Loading preview..."
		} else {
			previewContent = m.previewContent
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
	if m.operationBusy {
		output += "\n" + ui.RenderStatusMessage(m.operationProgress)
	}

	return output
}
