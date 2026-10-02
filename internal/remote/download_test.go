package remote

import (
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/sftp"
)

// 透過真實 SFTP 封包測試成功與中斷，不依賴外部 SSH 主機。
type virtualFile struct {
	size, failAfter int64
	readUntil       *atomic.Int64
}

func (f virtualFile) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	if r.Filepath != "/file" {
		return nil, os.ErrNotExist
	}
	return f, nil
}

func (f virtualFile) ReadAt(p []byte, off int64) (int, error) {
	if f.failAfter >= 0 && off >= f.failAfter {
		return 0, errors.New("read interrupted")
	}
	if off >= f.size {
		return 0, io.EOF
	}
	n := len(p)
	if int64(n) > f.size-off {
		n = int(f.size - off)
	}
	if f.failAfter >= 0 && int64(n) > f.failAfter-off {
		n = int(f.failAfter - off)
	}
	for i := 0; i < n; i++ {
		p[i] = 'x'
	}
	if f.readUntil != nil {
		f.readUntil.Store(off + int64(n))
	}
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func pipeRemote(t *testing.T, f virtualFile) *RemoteClient {
	t.Helper()
	clientConn, serverConn := net.Pipe()
	server := sftp.NewRequestServer(serverConn, sftp.Handlers{FileGet: f})
	done := make(chan error, 1)
	go func() { done <- server.Serve() }()
	var client *sftp.Client
	t.Cleanup(func() {
		if client != nil {
			client.Close()
		}
		clientConn.Close()
		server.Close()
		serverConn.Close()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("SFTP server 未結束")
		}
	})
	var err error
	client, err = sftp.NewClientPipe(clientConn, clientConn)
	if err != nil {
		t.Fatal(err)
	}
	return &RemoteClient{client: client}
}

func TestGetPropagatesInterruptedSFTPRead(t *testing.T) {
	c := pipeRemote(t, virtualFile{size: 98304, failAfter: 32768})
	data, err := c.Get("/file")
	if err == nil {
		t.Fatalf("中斷讀取被當成功: %d bytes", len(data))
	}
	if data != nil {
		t.Fatal("Get 失敗不應回傳可誤用的部分內容")
	}
}

func TestGetCompletesAtEOF(t *testing.T) {
	c := pipeRemote(t, virtualFile{size: 32773, failAfter: -1})
	data, err := c.Get("/file")
	if err != nil || string(data) != strings.Repeat("x", 32773) {
		t.Fatalf("EOF 或尾段讀取失敗: %d bytes %v", len(data), err)
	}
}

type streamingSink struct {
	written   int64
	maxChunk  int
	readUntil *atomic.Int64
	fail      error
}

func (w *streamingSink) Write(p []byte) (int, error) {
	if w.fail != nil {
		return 0, w.fail
	}
	if w.readUntil != nil && w.readUntil.Load() > w.written+int64(len(p)) {
		return 0, errors.New("下載先累積完整內容才寫出")
	}
	for _, b := range p {
		if b != 'x' {
			return 0, errors.New("內容不正確")
		}
	}
	w.written += int64(len(p))
	if len(p) > w.maxChunk {
		w.maxChunk = len(p)
	}
	return len(p), nil
}

func TestDownloadStreamsBeforeEntireFileIsRead(t *testing.T) {
	var readUntil atomic.Int64
	c := pipeRemote(t, virtualFile{size: 8*1024*1024 + 7, failAfter: -1, readUntil: &readUntil})
	sink := &streamingSink{readUntil: &readUntil}
	n, err := c.Download("/file", sink)
	if err != nil || n != 8*1024*1024+7 || sink.written != n {
		t.Fatalf("下載內容不完整: %d %v", n, err)
	}
	if sink.maxChunk > 32768 {
		t.Fatalf("不是固定緩衝串流: %d", sink.maxChunk)
	}
}

func TestDownloadReportsPartialCountAndReadError(t *testing.T) {
	c := pipeRemote(t, virtualFile{size: 98304, failAfter: 32768})
	n, err := c.Download("/file", io.Discard)
	if err == nil || n != 32768 {
		t.Fatalf("中斷未傳遞正確已寫入數: %d %v", n, err)
	}
}

func TestDownloadReportsWriterAndOpenErrors(t *testing.T) {
	c := pipeRemote(t, virtualFile{size: 32773, failAfter: -1})
	diskErr := errors.New("disk full")
	n, err := c.Download("/file", &streamingSink{fail: diskErr})
	if n != 0 || !errors.Is(err, diskErr) {
		t.Fatalf("writer 錯誤遺失: %d %v", n, err)
	}
	if _, err := c.Download("/missing", io.Discard); err == nil {
		t.Fatal("開啟錯誤遺失")
	}
	if _, err := c.Download("/file", nil); err == nil {
		t.Fatal("nil writer 未檢查")
	}
	if _, err := (&RemoteClient{}).Download("/file", io.Discard); err == nil {
		t.Fatal("未連線 client 未檢查")
	}
}
