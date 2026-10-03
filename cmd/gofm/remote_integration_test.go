package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/sftp"
	"gofm/internal/app"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func TestRemoteCLIToTUIDownload(t *testing.T) {
	_, hostKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	hostSigner, err := ssh.NewSignerFromKey(hostKey)
	if err != nil {
		t.Fatal(err)
	}
	_, clientKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	clientSigner, err := ssh.NewSignerFromKey(clientKey)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "id_ed25519")
	private, err := x509.MarshalPKCS8PrivateKey(clientKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: private}), 0600); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	host, port, _ := net.SplitHostPort(listener.Addr().String())
	knownFile := filepath.Join(dir, "known_hosts")
	if err := os.WriteFile(knownFile, []byte(knownhosts.Line([]string{listener.Addr().String()}, hostSigner.PublicKey())+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	handlers := sftp.InMemHandler()
	if err := handlers.FileCmd.Filecmd(&sftp.Request{Method: "Mkdir", Filepath: "/docs"}); err != nil {
		t.Fatal(err)
	}
	writer, err := handlers.FilePut.Filewrite(&sftp.Request{Method: "Put", Filepath: "/docs/sample.txt", Flags: 0x1a})
	if err != nil {
		t.Fatal(err)
	}
	payload := bytes.Repeat([]byte("中文資料\n"), 32769)
	if _, err := writer.WriteAt(payload, 0); err != nil {
		t.Fatal(err)
	}
	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
	done := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(10 * time.Second))
		config := &ssh.ServerConfig{PublicKeyCallback: func(meta ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if meta.User() != "sam" || !bytes.Equal(key.Marshal(), clientSigner.PublicKey().Marshal()) {
				return nil, fmt.Errorf("authentication rejected")
			}
			return nil, nil
		}}
		config.AddHostKey(hostSigner)
		server, channels, requests, err := ssh.NewServerConn(conn, config)
		if err != nil {
			done <- err
			return
		}
		defer server.Close()
		go ssh.DiscardRequests(requests)
		incoming, ok := <-channels
		if !ok {
			done <- fmt.Errorf("missing channel")
			return
		}
		channel, channelRequests, err := incoming.Accept()
		if err != nil {
			done <- err
			return
		}
		defer channel.Close()
		request, ok := <-channelRequests
		if !ok || request.Type != "subsystem" {
			done <- fmt.Errorf("missing subsystem")
			return
		}
		request.Reply(true, nil)
		go ssh.DiscardRequests(channelRequests)
		sftpServer := sftp.NewRequestServer(channel, handlers)
		defer sftpServer.Close()
		err = sftpServer.Serve()
		if err == io.EOF {
			err = nil
		}
		done <- err
	}()
	settings, err := parseOptions([]string{"--remote", "sam@" + host + ":/", "--port", port, "--key", keyFile, "--known-hosts", knownFile, "--download-dir", dir})
	if err != nil {
		t.Fatal(err)
	}
	m := app.NewRemote(*settings.remote, settings.remoteDir, settings.downloadDir)
	defer m.Close()
	_, read := m.Update(m.Init()())
	if read == nil {
		t.Fatalf("連線失敗: %s", m.ErrorMessage)
	}
	m.Update(read())
	if len(m.Entries) != 1 || m.Entries[0].Name != "docs" || !m.Entries[0].IsDir {
		t.Fatalf("根目錄錯誤: %v %s", m.Entries, m.ErrorMessage)
	}
	_, read = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if read == nil {
		t.Fatal("未讀取子目錄")
	}
	m.Update(read())
	if m.CurrentPath != "/docs" || len(m.Entries) != 1 || m.Entries[0].Path != "/docs/sample.txt" || m.Entries[0].Size != int64(len(payload)) || m.Entries[0].ModTime.IsZero() {
		t.Fatalf("遠端 metadata 錯誤: %v", m.Entries)
	}
	m.Width = 150
	if !strings.Contains(m.View(), "sftp://sam@127.0.0.1:") || !strings.Contains(m.View(), "Ctrl+D: download") {
		t.Fatal("未顯示遠端入口")
	}
	_, download := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if download == nil {
		t.Fatal("未啟動下載")
	}
	m.Update(download())
	data, err := os.ReadFile(filepath.Join(dir, "sample.txt"))
	if err != nil || !bytes.Equal(data, payload) {
		t.Fatalf("完整下載不符: %d %v %s", len(data), err, m.ErrorMessage)
	}
	if !strings.Contains(m.StatusMessage, "Downloaded:") {
		t.Fatalf("未回報成功: %s", m.StatusMessage)
	}
	_, parent := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if parent == nil {
		t.Fatal("未返回根目錄")
	}
	m.Update(parent())
	if m.CurrentPath != "/" {
		t.Fatal("根路徑錯誤")
	}
	m.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("session 未關閉")
	}
}
