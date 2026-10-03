package app

import (
	"context"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"gofm/internal/remote"
	"gofm/internal/types"
)

type remoteSession interface {
	ReadDirectory(string) ([]types.FileEntry, error)
	Download(string, io.Writer) (int64, error)
	Close() error
}

type remoteState struct {
	config      remote.Config
	downloadDir string
	client      remoteSession
	connect     func(context.Context, *remote.Config) (remoteSession, error)
	cancel      context.CancelFunc
	connectID   uint64
	connecting  bool
	closed      bool
	work        remoteWork
}

// 與 Model 分離的工作生命週期：關閉後禁止新工作，等待已開始的工作清理。
type remoteWork struct {
	mu     sync.Mutex
	closed bool
	wg     sync.WaitGroup
}

func (w *remoteWork) command(run tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		w.mu.Lock()
		if w.closed {
			w.mu.Unlock()
			return nil
		}
		w.wg.Add(1)
		w.mu.Unlock()
		defer w.wg.Done()
		return run()
	}
}

func (w *remoteWork) wait() {
	w.mu.Lock()
	w.closed = true
	w.mu.Unlock()
	w.wg.Wait()
}

type remoteConnectedMsg struct {
	id     uint64
	client remoteSession
	err    error
}
type remoteDownloadedMsg struct {
	id          uint64
	destination string
	written     int64
	err         error
}

func (m *AppState) quit() (tea.Model, tea.Cmd) {
	if m.remote != nil {
		m.remote.closed = true
		if m.remote.cancel != nil {
			m.remote.cancel()
		}
	}
	return m, tea.Quit
}

// NewRemote 使用 POSIX 起始路徑，實際連線由 Init 的背景命令執行。
func NewRemote(config remote.Config, remoteDir, localDir string) *AppState {
	m := New(localDir)
	m.CurrentPath = path.Clean("/" + remoteDir)
	localDir, _ = filepath.Abs(localDir)
	m.remote = &remoteState{config: config, downloadDir: localDir, connect: func(ctx context.Context, c *remote.Config) (remoteSession, error) {
		client, err := remote.NewRemoteClientContext(ctx, c)
		if err != nil {
			return nil, err
		}
		return client, nil
	}}
	return m
}

// Close 由 CLI 在事件迴圈結束後釋放連線；取消 context 也可中止握手與傳輸。
func (m *AppState) Close() {
	if m.remote == nil {
		return
	}
	m.remote.closed = true
	if m.remote.cancel != nil {
		m.remote.cancel()
	}
	if m.remote.client != nil {
		m.remote.client.Close()
		m.remote.client = nil
	}
	m.remote.work.wait()
}

func (m *AppState) connectRemote() tea.Cmd {
	r := m.remote
	if r.closed || r.connecting || m.operationBusy {
		return nil
	}
	if r.cancel != nil {
		r.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.connectID++
	id, config, connect, old := r.connectID, r.config, r.connect, r.client
	r.client = nil
	r.connecting = true
	m.loadID++
	m.Entries = nil
	m.GitInfo = nil
	m.Mode = ModeNormal
	m.hidePreview()
	m.ErrorMessage = ""
	m.StatusMessage = "Connecting to " + config.Host + "..."
	return r.work.command(func() tea.Msg {
		if old != nil {
			old.Close()
		}
		client, err := connect(ctx, &config)
		if ctx.Err() != nil {
			if client != nil {
				client.Close()
			}
			client = nil
			err = ctx.Err()
		}
		return remoteConnectedMsg{id: id, client: client, err: err}
	})
}

func (m *AppState) applyRemoteConnection(msg remoteConnectedMsg) tea.Cmd {
	r := m.remote
	if r == nil || r.closed || msg.id != r.connectID {
		if msg.client != nil {
			return func() tea.Msg { msg.client.Close(); return nil }
		}
		return nil
	}
	r.connecting = false
	if msg.err != nil {
		m.ErrorMessage = "SSH connection failed: " + msg.err.Error()
		m.StatusMessage = "Ctrl+R: reconnect"
		return nil
	}
	r.client = msg.client
	return m.loadDirectory()
}

func (m *AppState) downloadRemote() tea.Cmd {
	if m.operationBusy || m.remote.connecting {
		return nil
	}
	if m.remote.client == nil {
		m.ErrorMessage = "Not connected; Ctrl+R to reconnect"
		return nil
	}
	if m.Cursor < 0 || m.Cursor >= len(m.Entries) {
		return nil
	}
	entry := m.Entries[m.Cursor]
	if entry.IsDir {
		m.ErrorMessage = "Select a file to download"
		return nil
	}
	m.operationID++
	id, client, dir := m.operationID, m.remote.client, m.remote.downloadDir
	m.operationBusy = true
	m.operationProgress = "Downloading: " + entry.Name
	m.StatusMessage = m.operationProgress
	return m.remote.work.command(func() tea.Msg {
		destination, written, err := remote.DownloadToDirectory(client, entry.Path, dir)
		return remoteDownloadedMsg{id: id, destination: destination, written: written, err: err}
	})
}

func (m *AppState) applyRemoteDownload(msg remoteDownloadedMsg) {
	if m.remote == nil || m.remote.closed || !m.operationBusy || msg.id != m.operationID {
		return
	}
	m.operationBusy = false
	m.operationProgress = ""
	if msg.err != nil {
		m.ErrorMessage = msg.err.Error()
		m.StatusMessage = "Download failed"
		return
	}
	m.StatusMessage = fmt.Sprintf("Downloaded: %s (%d bytes)", msg.destination, msg.written)
}

func (m *AppState) remotePathLabel() string {
	port := m.remote.config.Port
	if port == "" {
		port = "22"
	}
	return fmt.Sprintf("sftp://%s@%s:%s%s", m.remote.config.User, m.remote.config.Host, port, m.CurrentPath)
}
