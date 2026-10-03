# gofm - Terminal File Manager

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

高效能 Terminal 檔案管理器，使用 Go (Golang) 開發，提供鍵盤導向的檔案瀏覽、操作與管理能力。

![Terminal_File_Manager](./images/Terminal_File_Manager.gif)

## 目標

- 比 GUI 檔案總管更快
- 比 shell ls/cd 更直覺
- 支援大型目錄 (10k+ files)

## 功能特色

### 核心功能
- **檔案導航** - 鍵盤操作 (↑ / Up 上移, ↓ / Down 下移, Enter/l 開啟, h 上一層)
- **檔案操作** - 複製 (y), 貼上 (p), 刪除 (d), 重新命名 (r), 新增檔案/目錄 (a/A)
- **Fuzzy 搜尋** - 按 `/` 進入搜尋模式，即時篩選檔案
  - 搜尋時使用 ↑/↓ 選取結果，Enter 確認並開啟；Esc 還原完整列表，`j/k` 可作為查詢字元輸入。
- **排序功能** - 按 `s` 切換升序/降序，按 `S` 循環切換 (name → size → type → modified)
- **預覽面板** - 選取檔案時自動顯示預覽
  - 文字檔案預覽
  - 圖片資訊 (PNG, JPEG, GIF, BMP, WebP, ICO)
  - 二進制檔案資訊

### 進階功能
- **Git 整合** - 顯示 Git 倉庫變更狀態 (M modified, A added, D deleted)
- **外掛介面（尚未接入 TUI）** - 提供程式內註冊及示範外掛；`~/.config/gofm/plugins` 目前只建立及列出目錄，不載入外掛。後續程序協定見 [設計決策](docs/PLUGIN_DESIGN.md)。
- **SSH/SFTP 遠端瀏覽** - 透過 `--remote` 使用私鑰與 known_hosts 連線，背景瀏覽目錄並串流下載至本機。
- **Lazy Load** - 目錄快速載入，非同步載入詳細資訊

### 錯誤處理
- Permission denied 提示
- 檔案刪除後自動刷新
- 損壞的符號連結高亮顯示

## 安裝

```bash
# Clone 後編譯
go build -o gofm ./cmd/gofm
```

## 使用方式

```bash
# 執行 (預設目前目錄)
./gofm

# 指定目錄
./gofm /var/www
./gofm ~/Documents
```

## 遠端瀏覽與下載

```powershell
# Windows PowerShell；路徑含空白時加上引號
.\gofm.exe --remote sam@example.com:/home/sam --key "$env:USERPROFILE\.ssh\id_ed25519" --download-dir "D:\Downloads"

# 非預設埠與指定主機驗證檔案
.\gofm.exe --remote sam@example.com:/data --port 2222 --key "D:\Keys\id_ed25519" --known-hosts "D:\Keys\known_hosts"
```

```bash
./gofm --remote sam@example.com:/home/sam --key ~/.ssh/id_ed25519 --download-dir ~/Downloads
# IPv6 主機使用中括號
./gofm --remote 'sam@[::1]:/' --key ~/.ssh/id_ed25519
```

`--port` 預設 22；`--known-hosts` 預設 `~/.ssh/known_hosts`；`--download-dir` 預設目前工作目錄，須已存在。必須指定 `--key`；目前 CLI 不提供密碼、SSH agent 或加密私鑰解鎖入口。遠端旗標放在參數前；本機模式維持 `gofm [目錄]`。

- Enter 開啟遠端目錄，←／h 返回上層；選取檔案按 Ctrl+D 下載，Ctrl+R 重連並刷新目前目錄。搜尋、排序與自訂導航鍵位可用，Ctrl+C 在所有模式皆可退出。
- 連線、讀取及下載背景執行，顯示執行中／完成／錯誤；慢速網路期間仍可導航及調整視窗。退出取消 session 並等待下載暫存清理。
- 下載先寫唯一暫存檔，完整關閉後才發布，不覆寫既有目的檔；失敗會清理暫存並顯示原因。目的檔案系統須支援同目錄 hard link（如 NTFS、一般 Linux 本機檔案系統）；不支援時會拒絕發布，並清理暫存。包含 Windows 裝置名或不安全字元的檔名會拒絕下載。
- 本次遠端模式支援瀏覽與單檔下載；上傳、修改與遠端預覽尚未提供。遠端不執行本機 metadata／Git 讀取。

專案內 Go 程式也可使用 `remote.NewRemoteClient`／`NewRemoteClientContext`：

- `Get(remotePath)` 回傳完整小檔案內容；讀取中斷會回傳錯誤，不會把部分內容當作成功。
- `Download(remotePath, writer)` 串流寫入目的地，回傳已寫入的 byte 數與錯誤。下載大檔時可傳入本機檔案 writer，避免把整份內容保留在記憶體中。
- 寫入失敗或連線中斷時，writer 可能已有部分內容；呼叫端應依錯誤決定重試或清理。
- `DownloadToDirectory(client, remotePath, localDir)` 提供上述暫存清理與不覆寫發布行為。

## 快捷鍵

| 按鍵 | 動作 |
|------|------|
| ↑ / Up | 上移 |
| ↓ / Down | 下移 |
| Enter / l | 開啟檔案/目錄 |
| h | 返回上一層 |
| d | 刪除 |
| r | 重新命名 |
| y | 複製 |
| x | 剪下 |
| p | 貼上 |
| a | 新增檔案 |
| A | 新增目錄 |
| / | 搜尋 |
| PageUp / PageDown | 捲動預覽 |
| Esc | 關閉預覽／取消輸入或搜尋 |
| s | 切換排序方向 |
| S | 切換排序方式 |
| q | 離開 |

寬視窗以左右面板顯示列表與預覽；窄視窗顯示單一面板，Enter 開啟預覽後可按 Esc 返回列表。面板、通知與快捷鍵列會依視窗尺寸調整。

## 專案架構

```
cmd/gofm/main.go          - 入口點
├── app/app.go            - 主應internal/
用程式 (Bubble Tea Model)
├── fs/                  - 檔案系統操作
│   ├── filesystem.go    - 目錄讀取
│   └── operations.go    - 檔案操作
├── ui/components.go      - UI 元件
├── input/keymap.go       - 鍵位映射
├── preview/preview.go    - 檔案預覽
├── git/git.go            - Git 整合
├── plugin/plugin.go      - 外掛介面與示範（無動態載入）
├── remote/remote.go      - SSH/SFTP
├── logger/logger.go     - 日誌系統
├── state/state.go       - 狀態管理
└── types/types.go       - 共用類型
```

## 測試

```bash
# 執行所有測試
go test ./...

# 執行測試並顯示涵蓋率
go test -cover ./...
```

### 涵蓋率

| 套件 | 涵蓋率 |
|------|--------|
| types | 100.0% |
| state | 100.0% |
| git | 84.8% |
| preview | 80.4% |
| ui | 63.6% |
| fs | 81.6% |
| logger | 80.0% |
| plugin | 78.6% |
| input | 60.3% |
| app | 83.2% |
| remote | 70.3% |
| cmd/gofm | 60.9% |

> 註：以 `go test -cover ./...`（go1.26.1）量測。部分 UI、設定解析與錯誤分支尚未全部覆蓋；非同步載入、操作、搜尋、Unicode、自訂鍵位、SSH 握手取消及 SFTP 串流／TUI 整合已有回歸測試。

## 效能目標

| 指標 | 目標 |
|------|------|
| Startup | < 200ms |
| Navigation | < 16ms |
| Memory | < 50MB |
| Directory render | < 50ms |

## 技術棧

- **Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **SSH/SFTP**: [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh), [github.com/pkg/sftp](https://github.com/pkg/sftp)

## 自訂鍵位

啟動時讀取 `~/.config/gofm/config.toml` 的 `[keymap]`，例如：

```toml
[keymap]
down = "n"
copy = "c"
rename = "e"
quit = "z"
```

可設定 `up/down/open/back/delete/rename/copy/cut/paste/newfile/newdir/quit/select`，未指定者沿用預設；使用 Bubble Tea 按鍵名稱或單一可列印字元。任一鍵位無效或衝突時整份退回預設。方向鍵、Enter、Ctrl+C 固定可用；Esc、PageUp/PageDown、`/`、`s/S` 保留既有用途。預設 `j/k/Q` 別名亦保留，不能改派其他動作。自訂鍵位只作用於一般模式；搜尋／文字編輯仍輸入字元，Enter 確認、Esc 取消。狀態列顯示有效設定。

## SSH 主機驗證

`internal/remote.Config` 預設讀取使用者家目錄下的 `.ssh/known_hosts`；可用 `KnownHostsPath` 指定另一份驗證檔案。連線前請透過可信管道確認伺服器指紋，再以 SSH 工具建立對應紀錄。未知或已變更的金鑰、缺少或無法解析的驗證檔案均會拒絕連線，不會自動加入或略過驗證。非預設埠使用 `[host]:port` 的 known_hosts 紀錄。

## License

MIT License - see [LICENSE](LICENSE)
