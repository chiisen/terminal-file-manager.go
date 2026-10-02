# CHANGELOG.md

## [Unreleased]

### 新增

- `internal/input/keymap.go`：實作 `config.toml` 的 `[keymap]` 解析（`LoadKeymapFromPath`，標準函式庫、無新依賴）。
- 核心單元測試：`internal/app/app_test.go`、`internal/state/state_test.go`、`internal/ui/components_test.go`、`internal/remote/remote_test.go`、`internal/input/keymap_test.go`。
- 新增 `docs/OPEN_QUESTIONS.md`（HostKey 驗證、外掛載入、效能基準、`modified` 排序四項開放問題）。
- 效能基準：`internal/app/app_bench_test.go`、`internal/ui/components_bench_test.go`（10k 排序 ~9–10ms、視窗渲染 ~0.086ms）。

### 修正

- 修正目錄、詳細資訊與 Git 的非同步載入：背景命令回傳完成訊息，由事件迴圈更新狀態；以請求編號忽略過期結果，依路徑匹配 metadata 並保留搜尋與游標狀態（#1）。
- 複製前阻擋來源自身、硬連結及來源子目錄，解析目的地祖先的符號連結；複製來源符號連結時保留連結，避免遞迴循環與來源截斷（#2）。
- 修正中文及 emoji 的文字辨識、控制字元清理與預覽截斷，保留有效 UTF-8 內容（#3）。
- 移除會遺失快速輸入的全域按鍵節流，支援 Unicode、多字元貼上與按字元退格；搜尋模式的 `j/k` 改為查詢字元（#4）。
- 新增非同步載入、過期結果、複製路徑防護與 Unicode 的回歸測試。

- Windows 相容性：`internal/plugin/plugin.go`、`internal/logger/logger.go`（含測試）由 `os.Getenv("HOME")` 改為 `os.UserHomeDir()` 優先、`HOME` 備援。
- 遠端路徑：`internal/remote/remote.go` 遠端拼接由 `filepath.Join` 改為 `path.Join`（POSIX 語義）；修正 `InsecureIgnoreHostKey` 註解錯字並加註生產環境警告。

### 文件

- `TODO.md`：`config.toml` 標註完成；`event.go` 條目修正為「已合併至 keymap.go + app.go」。
- `docs/PRD.md` §8：包結構更新為實作現況（ui/components.go、input/keymap.go 合併說明）。
- `README.md`：覆蓋率表新增 input/app/state/ui/remote 列（確切數字待跑 `go test -cover ./...` 回填）。
