package app

import (
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"gofm/internal/preview"
	"gofm/internal/types"
)

// 預覽與檔案操作的完成訊息不攜帶 Model 指標，背景工作不修改畫面狀態。
type previewLoadedMsg struct {
	id      uint64
	entry   types.FileEntry
	content string
	err     error
}

type operationFinishedMsg struct {
	id                         uint64
	directory, source, success string
	isCut                      bool
	err                        error
}

func (m *AppState) hidePreview() {
	m.PreviewActive = false
	m.previewID++
	m.previewLoading = false
	m.previewContent = ""
	m.previewPath = ""
	m.previewOffset = 0
}

// 單份快取有固定記憶體上限；路徑、大小與修改時間共同辨識預覽版本。
func (m *AppState) startPreview(entry types.FileEntry, read func(string) (*preview.PreviewResult, error)) tea.Cmd {
	m.previewID++
	id := m.previewID
	m.PreviewActive = true
	m.previewPath = entry.Path
	m.previewContent = ""
	m.previewOffset = 0
	if m.previewCacheValid && m.previewCacheEntry.Path == entry.Path && m.previewCacheEntry.Size == entry.Size && m.previewCacheEntry.ModTime.Equal(entry.ModTime) {
		m.previewLoading = false
		m.previewContent = m.previewCacheContent
		return nil
	}
	m.previewLoading = true
	return func() tea.Msg {
		result, err := read(entry.Path)
		var content string
		if err == nil {
			content = result.Content
		}
		return previewLoadedMsg{id: id, entry: entry, content: content, err: err}
	}
}

// 使用不可變的參數快照排程檔案操作，busy 只由事件迴圈設定與解除。
func (m *AppState) startOperation(progress, success, source string, isCut bool, run func() error) tea.Cmd {
	if m.operationBusy {
		return nil
	}
	m.operationBusy = true
	m.operationID++
	id, directory := m.operationID, m.CurrentPath
	m.operationProgress = progress
	m.StatusMessage = progress
	return func() tea.Msg {
		return operationFinishedMsg{id: id, directory: directory, source: source, isCut: isCut, success: success, err: run()}
	}
}

func (m *AppState) applyPreview(msg previewLoadedMsg) {
	if msg.id != m.previewID || !m.PreviewActive || msg.entry.Path != m.previewPath || m.Cursor < 0 || m.Cursor >= len(m.Entries) || m.Entries[m.Cursor].Path != msg.entry.Path {
		return
	}
	m.previewLoading = false
	if msg.err != nil {
		m.previewContent = fmt.Sprintf("Error: %v", msg.err)
		return
	}
	m.previewContent = msg.content
	m.previewCacheEntry, m.previewCacheContent, m.previewCacheValid = msg.entry, msg.content, true
}

func (m *AppState) finishOperation(msg operationFinishedMsg) tea.Cmd {
	if !m.operationBusy || msg.id != m.operationID {
		return nil
	}
	m.operationBusy = false
	m.operationProgress = ""
	if msg.err != nil {
		m.ErrorMessage = msg.err.Error()
		m.StatusMessage = "Operation failed"
	} else {
		m.StatusMessage = msg.success
		// 使用者可能在背景移動期間複製別的檔案，不可清除新的剪貼簿。
		if msg.isCut && m.IsCut && m.Clipboard == msg.source {
			m.Clipboard = ""
			m.IsCut = false
		}
	}
	m.previewCacheValid = false
	if m.CurrentPath == msg.directory || (msg.source != "" && filepath.Dir(msg.source) == m.CurrentPath) {
		m.Mode = ModeNormal
		return m.loadDirectory()
	}
	return nil
}
