package remote

import (
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func testSigner(t *testing.T) ssh.Signer {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

// 真實 SSH 握手與 SFTP 子系統；未知金鑰不應到達 SFTP。
func sshTestServer(t *testing.T, signer ssh.Signer) (string, string) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(10 * time.Second))
		config := &ssh.ServerConfig{NoClientAuth: true}
		config.AddHostKey(signer)
		server, channels, requests, err := ssh.NewServerConn(conn, config)
		if err != nil {
			return
		}
		defer server.Close()
		go ssh.DiscardRequests(requests)
		for incoming := range channels {
			channel, requests, err := incoming.Accept()
			if err != nil {
				return
			}
			go func() {
				defer channel.Close()
				for request := range requests {
					if request.Type == "subsystem" {
						request.Reply(true, nil)
						server := sftp.NewRequestServer(channel, sftp.InMemHandler())
						defer server.Close()
						server.Serve()
						return
					}
					request.Reply(false, nil)
				}
			}()
		}
	}()
	t.Cleanup(func() {
		listener.Close()
		select {
		case <-done:
		case <-time.After(12 * time.Second):
			t.Error("SSH server 未結束")
		}
	})
	host, port, _ := net.SplitHostPort(listener.Addr().String())
	return host, port
}

func TestSSHDefaultKnownHostsHandshake(t *testing.T) {
	for _, scenario := range []string{"known", "unknown", "changed", "missing"} {
		t.Run(scenario, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("USERPROFILE", home)
			t.Setenv("HOME", home)
			signer := testSigner(t)
			host, port := sshTestServer(t, signer)
			if scenario != "missing" {
				if err := os.MkdirAll(filepath.Join(home, ".ssh"), 0700); err != nil {
					t.Fatal(err)
				}
				content := ""
				if scenario != "unknown" {
					key := signer.PublicKey()
					if scenario == "changed" {
						key = testSigner(t).PublicKey()
					}
					content = knownhosts.Line([]string{net.JoinHostPort(host, port)}, key) + "\n"
				}
				if err := os.WriteFile(filepath.Join(home, ".ssh", "known_hosts"), []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			client, err := NewRemoteClient(&Config{Host: host, Port: port, User: "tester"})
			if client != nil {
				defer client.Close()
			}
			if scenario == "known" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatalf("%s 金鑰不應被接受", scenario)
			}
			if !strings.Contains(err.Error(), "known_hosts") {
				t.Fatalf("錯誤應指出 known_hosts：%v", err)
			}
		})
	}
}

func TestSSHExplicitKnownHosts(t *testing.T) {
	signer := testSigner(t)
	host, port := sshTestServer(t, signer)
	file := filepath.Join(t.TempDir(), "trusted_hosts")
	if err := os.WriteFile(file, []byte(knownhosts.Line([]string{net.JoinHostPort(host, port)}, signer.PublicKey())+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	config := &Config{Host: host, Port: port, User: "tester", KnownHostsPath: file}
	client, err := NewRemoteClient(config)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if config.KnownHostsPath != file {
		t.Fatal("呼叫者設定被改寫")
	}
	for _, invalid := range []*Config{nil, {}, {Host: host}, {Host: host, User: "tester", KnownHostsPath: filepath.Join(t.TempDir(), "missing")}} {
		if _, err := NewRemoteClient(invalid); err == nil {
			t.Fatal("無效設定應提早回報錯誤")
		}
	}
}
