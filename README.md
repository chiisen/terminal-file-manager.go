# gofm - Terminal File Manager

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

高效能 Terminal 檔案管理器，使用 Go (Golang) 開發，提供鍵盤導向的檔案瀏覽、操作與管理能力。

## 目標

- 比 GUI 檔案總管更快
- 比 shell ls/cd 更直覺
- 支援大型目錄 (10k+ files)

## 功能特色

### 核心功能
- **檔案導航** - vi-style 鍵盤操作 (↑/k 上移, ↓/j 下移, Enter/l 開啟, h 上一層)
- **檔案操作** - 複製 (y), 貼上 (p), 刪除 (d), 重新命名 (r), 新增檔案/目錄 (a/A)
- **Fuzzy 搜尋** - 按 `/` 進入搜尋模式，即時篩選檔案
- **排序功能** - 按 `s` 切換升序/降序，按 `S` 循環切換 (name → size → type → modified)
- **預覽面板** - 選取檔案時自動顯示預覽
  - 文字檔案預覽
  - 圖片資訊 (PNG, JPEG, GIF, BMP, WebP, ICO)
  - 二進制檔案資訊

### 進階功能
- **Git 整合** - 顯示 Git 倉庫變更狀態 (M modified, A added, D deleted)
- **外掛系統** - 支援自訂外掛 (`~/.config/gofm/plugins`)
- **遠端檔案** - SSH/SFTP 遠端檔案系統支援
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

## 快捷鍵

| 按鍵 | 動作 |
|------|------|
| ↑ / k | 上移 |
| ↓ / j | 下移 |
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
| s | 切換排序方向 |
| S | 切換排序方式 |
| q | 離開 |

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
├── plugin/plugin.go      - 外掛系統
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
| types | 100% |
| logger | 83% |
| plugin | 80% |
| preview | 81% |
| git | 85% |
| fs | 81% |

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

## License

MIT License - see [LICENSE](LICENSE)
