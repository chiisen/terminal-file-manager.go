package app

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"gofm/internal/remote"
	"gofm/internal/types"
)

type controlledRemote struct {
	read     func(string) ([]types.FileEntry, error)
	download func(string, io.Writer) (int64, error)
	closed   bool
}

func (c *controlledRemote) ReadDirectory(p string) ([]types.FileEntry, error) { return c.read(p) }
func (c *controlledRemote) Download(p string, w io.Writer) (int64, error)     { return c.download(p, w) }
func (c *controlledRemote) Close() error                                      { c.closed = true; return nil }

func newRemoteTest(t *testing.T) *AppState {
	t.Helper()
	m := NewRemote(remote.Config{Host: "example.test", User: "sam"}, "/home/sam", t.TempDir())
	t.Cleanup(m.Close)
	return m
}

func TestRemoteConnectDoesNotBlockEvents(t *testing.T) {
	m := newRemoteTest(t)
	started, release := make(chan struct{}), make(chan struct{})
	session := &controlledRemote{read: func(p string) ([]types.FileEntry, error) {
		if p != "/home/sam" && p != "/home/sam/sub" {
			t.Errorf("讀取路徑 %s", p)
		}
		return []types.FileEntry{{Name: "sub", Path: "/home/sam/sub", IsDir: true}}, nil
	}}
	m.remote.connect = func(ctx context.Context, cfg *remote.Config) (remoteSession, error) {
		close(started)
		select {
		case <-release:
			return session, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	cmd := m.Init()
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	<-started
	m.Update(tea.WindowSizeMsg{Width: 50, Height: 12})
	if m.Width != 50 || !strings.Contains(m.View(), "Connecting") {
		t.Fatal("連線期間 UI 未回應或未提示")
	}
	close(release)
	_, readCmd := m.Update(<-done)
	if readCmd == nil {
		t.Fatal("連線後未讀取目錄")
	}
	_, metadataCmd := m.Update(readCmd())
	if metadataCmd != nil || len(m.Entries) != 1 || m.CurrentPath != "/home/sam" {
		t.Fatal("遠端讀取錯誤或嘗試本機 metadata")
	}
	_, subCmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if subCmd == nil || m.CurrentPath != "/home/sam/sub" {
		t.Fatal("未進入遠端目錄")
	}
	_, parentCmd := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if parentCmd == nil || m.CurrentPath != "/home/sam" {
		t.Fatal("返回使用了本機路徑規則")
	}
	m.Update(subCmd())
	if len(m.Entries) != 0 {
		t.Fatal("過期遠端目錄覆蓋目前位置")
	}
}

func TestRemoteDownloadIsBackgroundAndUsesSnapshot(t *testing.T) {
	m := newRemoteTest(t)
	started, release := make(chan struct{}), make(chan struct{})
	m.remote.client = &controlledRemote{download: func(p string, w io.Writer) (int64, error) {
		if p != "/home/sam/file.txt" {
			t.Errorf("下載路徑不是快照: %s", p)
		}
		close(started)
		<-release
		n, err := io.WriteString(w, "完整內容")
		return int64(n), err
	}}
	m.Entries = []types.FileEntry{{Name: "file.txt", Path: "/home/sam/file.txt"}, {Name: "sub", Path: "/home/sam/sub", IsDir: true}}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if cmd == nil {
		t.Fatal("未開始下載")
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	<-started
	if _, duplicate := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD}); duplicate != nil {
		t.Fatal("重複下載未阻擋")
	}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Cursor != 1 {
		t.Fatal("下載期間不能導航")
	}
	close(release)
	m.Update(<-done)
	if m.operationBusy || m.ErrorMessage != "" {
		t.Fatalf("完成狀態錯誤: %s", m.ErrorMessage)
	}
	data, err := os.ReadFile(filepath.Join(m.remote.downloadDir, "file.txt"))
	if err != nil || string(data) != "完整內容" {
		t.Fatalf("下載內容錯誤 %q %v", data, err)
	}
}

func TestRemoteFailuresAndMutationGuard(t *testing.T) {
	m := newRemoteTest(t)
	m.remote.connect = func(context.Context, *remote.Config) (remoteSession, error) {
		return nil, errors.New("host key rejected")
	}
	m.Update(m.Init()())
	if !strings.Contains(m.ErrorMessage, "host key rejected") || !strings.Contains(m.View(), "host key rejected") {
		t.Fatal("連線錯誤未顯示")
	}
	m.remote.client = &controlledRemote{read: func(string) ([]types.FileEntry, error) { return nil, errors.New("disconnected") }, download: func(_ string, w io.Writer) (int64, error) {
		n, _ := io.WriteString(w, "partial")
		return int64(n), errors.New("interrupted")
	}}
	m.Update(m.loadDirectory()())
	if !strings.Contains(m.ErrorMessage, "disconnected") {
		t.Fatal("讀取斷線未顯示")
	}
	m.Entries = []types.FileEntry{{Name: "file", Path: "/home/sam/file"}}
	for _, key := range []string{"d", "r", "a", "A", "x", "y", "p"} {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		if m.Mode != ModeNormal || m.operationBusy {
			t.Fatalf("遠端修改鍵未阻擋: %s", key)
		}
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m.Update(cmd())
	if !strings.Contains(m.ErrorMessage, "interrupted") || strings.Contains(m.StatusMessage, "Downloaded:") {
		t.Fatal("中斷被回報成功")
	}
	files, _ := os.ReadDir(m.remote.downloadDir)
	if len(files) != 0 {
		t.Fatal("失敗下載殘留檔案")
	}
}

func TestRemoteQuitCancelsPendingConnect(t *testing.T) {
	m := newRemoteTest(t)
	started := make(chan struct{})
	m.remote.connect = func(ctx context.Context, _ *remote.Config) (remoteSession, error) {
		close(started)
		<-ctx.Done()
		return nil, ctx.Err()
	}
	done := make(chan tea.Msg, 1)
	cmd := m.Init()
	go func() { done <- cmd() }()
	<-started
	_, quit := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if quit == nil {
		t.Fatal("不能退出")
	}
	select {
	case msg := <-done:
		m.Update(msg)
	case <-time.After(time.Second):
		t.Fatal("退出未取消連線")
	}
}

func TestRealRemoteQuitCancelsHandshake(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn
		}
	}()
	file := filepath.Join(t.TempDir(), "known_hosts")
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	host, port, _ := net.SplitHostPort(listener.Addr().String())
	m := NewRemote(remote.Config{Host: host, Port: port, User: "test", KnownHostsPath: file}, "/", t.TempDir())
	defer m.Close()
	done := make(chan tea.Msg, 1)
	cmd := m.Init()
	go func() { done <- cmd() }()
	select {
	case conn := <-accepted:
		defer conn.Close()
	case <-time.After(3 * time.Second):
		t.Fatal("未開始握手")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("退出未中止真實握手")
	}
}

func TestRemoteCloseWaitsForDownloadCleanup(t *testing.T) {
	m := newRemoteTest(t)
	var sessionContext context.Context
	started, canceled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	client := &controlledRemote{read: func(string) ([]types.FileEntry, error) { return nil, nil }, download: func(_ string, w io.Writer) (int64, error) {
		n, _ := io.WriteString(w, "partial")
		close(started)
		<-sessionContext.Done()
		close(canceled)
		<-release
		return int64(n), sessionContext.Err()
	}}
	m.remote.connect = func(ctx context.Context, _ *remote.Config) (remoteSession, error) {
		sessionContext = ctx
		return client, nil
	}
	m.Update(m.Init()())
	m.Entries = []types.FileEntry{{Name: "file", Path: "/file"}}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	<-started
	closed := make(chan struct{})
	go func() { m.Close(); close(closed) }()
	<-canceled
	select {
	case <-closed:
		close(release)
		<-done
		t.Fatal("Close 未等待下載清理")
	case <-time.After(30 * time.Millisecond):
	}
	close(release)
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("清理後未結束")
	}
	<-done
	files, err := os.ReadDir(m.remote.downloadDir)
	if err != nil || len(files) != 0 {
		t.Fatalf("退出殘留暫存: %v %v", files, err)
	}
}

func TestRemoteCtrlCWorksDuringSearch(t *testing.T) {
	m := newRemoteTest(t)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("搜尋模式的 Ctrl+C 未退出")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("沒有退出訊息")
	}
}

func TestSlowRemoteDirectoryKeepsUIResponsive(t *testing.T) {
	m := newRemoteTest(t)
	started, release := make(chan struct{}), make(chan struct{})
	m.remote.client = &controlledRemote{read: func(p string) ([]types.FileEntry, error) {
		if p != "/home/sam" {
			t.Errorf("錯誤路徑: %s", p)
		}
		close(started)
		<-release
		return []types.FileEntry{{Name: "file", Path: "/home/sam/file"}}, nil
	}}
	cmd := m.loadDirectory()
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	<-started
	m.Update(tea.WindowSizeMsg{Width: 30, Height: 8})
	if m.Width != 30 || !strings.Contains(m.View(), "Reading remote") {
		close(release)
		<-done
		t.Fatal("慢速讀取期間 UI 未回應")
	}
	close(release)
	m.Update(<-done)
	if len(m.Entries) != 1 || m.Entries[0].Name != "file" {
		t.Fatal("目錄結果未套用")
	}
}

func TestRemoteReconnectReplacesSession(t *testing.T) {
	m := newRemoteTest(t)
	old := &controlledRemote{}
	m.remote.client = old
	client := &controlledRemote{read: func(p string) ([]types.FileEntry, error) {
		return []types.FileEntry{{Name: "ready", Path: p + "/ready"}}, nil
	}}
	m.remote.connect = func(context.Context, *remote.Config) (remoteSession, error) { return client, nil }
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	if cmd == nil {
		t.Fatal("沒有重連命令")
	}
	_, read := m.Update(cmd())
	if !old.closed || read == nil {
		t.Fatal("舊連線未釋放或新連線未載入")
	}
	m.Update(read())
	if len(m.Entries) != 1 || m.Entries[0].Name != "ready" {
		t.Fatal("重連後列表錯誤")
	}
}
