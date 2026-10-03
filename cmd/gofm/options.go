package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"gofm/internal/remote"
)

type options struct {
	startPath              string
	remote                 *remote.Config
	remoteDir, downloadDir string
}

const usage = `使用方式：
  gofm [本機目錄]
  gofm --remote user@host:/path --key <私鑰> [--port 22] [--known-hosts <檔案>] [--download-dir <本機目錄>]

遠端操作：Enter 開啟目錄、Ctrl+D 下載檔案、Ctrl+R 重連／刷新、Ctrl+C 退出。
遠端主機金鑰必須已存在於 known_hosts；下載不覆寫已有檔案。
`

func parseOptions(args []string) (options, error) {
	result := options{startPath: "."}
	flags := flag.NewFlagSet("gofm", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	remotePath := flags.String("remote", "", "遠端 user@host:/path")
	port := flags.String("port", "22", "SSH 埠")
	key := flags.String("key", "", "SSH 私鑰路徑")
	known := flags.String("known-hosts", "", "主機驗證檔案")
	download := flags.String("download-dir", ".", "本機下載目錄")
	if err := flags.Parse(args); err != nil {
		return result, err
	}
	if *remotePath == "" {
		remoteFlag := false
		flags.Visit(func(f *flag.Flag) { remoteFlag = true })
		if remoteFlag {
			return result, fmt.Errorf("遠端旗標須搭配 --remote")
		}
		if flags.NArg() > 1 {
			return result, fmt.Errorf("只能指定一個本機目錄")
		}
		if flags.NArg() == 1 {
			result.startPath = flags.Arg(0)
		}
		return result, nil
	}
	if flags.NArg() != 0 {
		return result, fmt.Errorf("遠端模式不可同時指定本機起始目錄")
	}
	user, host, dir, err := remote.ParseRemotePath(*remotePath)
	if err != nil {
		return result, err
	}
	portNumber, err := strconv.Atoi(*port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return result, fmt.Errorf("SSH 埠須介於 1 與 65535")
	}
	if *key == "" {
		return result, fmt.Errorf("遠端模式須提供 --key SSH 私鑰路徑")
	}
	localDir, err := filepath.Abs(*download)
	if err != nil {
		return result, err
	}
	info, err := os.Stat(localDir)
	if err != nil {
		return result, fmt.Errorf("無法存取下載目錄: %w", err)
	}
	if !info.IsDir() {
		return result, fmt.Errorf("下載目的地必須是目錄")
	}
	result.remote = &remote.Config{Host: host, Port: *port, User: user, KeyPath: *key, KnownHostsPath: *known}
	result.remoteDir, result.downloadDir = dir, localDir
	return result, nil
}
