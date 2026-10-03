package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoteOptionsAndLocalCompatibility(t *testing.T) {
	dir := t.TempDir()
	options, err := parseOptions([]string{"--remote", "sam@example.test:/home/sam", "--port", "2222", "--key", "id_ed25519", "--known-hosts", "trusted", "--download-dir", dir})
	if err != nil {
		t.Fatal(err)
	}
	if options.remote == nil || options.remote.Host != "example.test" || options.remote.User != "sam" || options.remote.Port != "2222" || options.remote.KeyPath != "id_ed25519" || options.remote.KnownHostsPath != "trusted" || options.remoteDir != "/home/sam" || options.downloadDir != dir {
		t.Fatalf("設定遺失: %+v", options)
	}
	local, err := parseOptions([]string{dir})
	if err != nil || local.remote != nil || local.startPath != dir {
		t.Fatalf("本機模式改變: %+v %v", local, err)
	}
	ipv6, err := parseOptions([]string{"--remote", "sam@[::1]:/", "--key", "id"})
	if err != nil || ipv6.remote.Host != "::1" {
		t.Fatalf("IPv6 解析錯誤: %+v %v", ipv6, err)
	}
}

func TestOptionsRejectConflictingAndInvalidInputs(t *testing.T) {
	notDir := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(notDir, nil, 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"--remote", "sam@host:/", "local"},
		{"--remote", "sam@host:/"},
		{"--key", "id"},
		{"--remote", "sam@host:/", "--key", "id", "--port", "0"},
		{"--remote", "sam@host:/", "--key", "id", "--port", "65536"},
		{"--remote", "sam@host:/", "--key", "id", "--download-dir", notDir},
		{"--remote", "@host:/", "--key", "id"},
		{"--remote", "sam@:/", "--key", "id"},
		{"--remote", "sam@[bad:/", "--key", "id"},
		{"one", "two"},
	} {
		if _, err := parseOptions(args); err == nil {
			t.Errorf("無效參數未拒絕: %v", args)
		}
	}
}
