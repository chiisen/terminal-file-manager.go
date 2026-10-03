package remote

import (
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDownloadToDirectoryPublishesCompleteFile(t *testing.T) {
	dir := t.TempDir()
	client := pipeRemote(t, virtualFile{size: 8*1024*1024 + 7, failAfter: -1})
	destination, count, err := DownloadToDirectory(client, "/file", dir)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(destination)
	if err != nil || info.Size() != 8*1024*1024+7 || count != 8*1024*1024+7 {
		t.Fatalf("未發布完整檔案: %v %d %v", info, count, err)
	}
	if _, _, err := DownloadToDirectory(client, "/file", dir); err == nil {
		t.Fatal("不得覆寫既有下載")
	}
	files, _ := os.ReadDir(dir)
	if len(files) != 1 || files[0].Name() != "file" {
		t.Fatalf("殘留暫存: %v", files)
	}
}

func TestDownloadToDirectoryCleansInterruptedFile(t *testing.T) {
	dir := t.TempDir()
	client := pipeRemote(t, virtualFile{size: 98304, failAfter: 32768})
	_, count, err := DownloadToDirectory(client, "/file", dir)
	if err == nil || count != 32768 {
		t.Fatalf("中斷應失敗並回報部分 bytes: %d %v", count, err)
	}
	files, _ := os.ReadDir(dir)
	if len(files) != 0 {
		t.Fatalf("中斷下載殘留檔案: %v", files)
	}
}

type writerDownload struct{}

func (writerDownload) Download(_ string, w io.Writer) (int64, error) {
	n, err := io.WriteString(w, "payload")
	return int64(n), err
}

func TestDownloadRejectsUnsafeNames(t *testing.T) {
	for _, name := range []string{"/", "/..", "/bad\\name", "/CON", "/nul.txt", "/a:b", "/trailing.", "/bad\nname", "/file?", "/COM¹.txt", "/LPT²"} {
		dir := t.TempDir()
		if _, _, err := DownloadToDirectory(writerDownload{}, name, dir); err == nil {
			t.Errorf("不安全名稱未拒絕: %q", name)
		}
		files, _ := os.ReadDir(dir)
		if len(files) != 0 {
			t.Fatalf("拒絕後仍寫入: %v", files)
		}
	}
}

func TestSSHHandshakeCancellation(t *testing.T) {
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		client, err := NewRemoteClientContext(ctx, &Config{Host: host, Port: port, User: "tester", KnownHostsPath: file})
		if client != nil {
			client.Close()
		}
		done <- err
	}()
	select {
	case conn := <-accepted:
		defer conn.Close()
	case <-time.After(3 * time.Second):
		t.Fatal("未開始握手")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "context canceled") {
			t.Fatalf("取消應回報原因: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("取消未停止握手")
	}
}
