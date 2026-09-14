# CHANGELOG.md

## [Unreleased]

### 新增

- `internal/input/keymap.go`：實作 `config.toml` 的 `[keymap]` 解析（`LoadKeymapFromPath`，標準函式庫、無新依賴）。
- 核心單元測試：`internal/app/app_test.go`、`internal/state/state_test.go`、`internal/ui/components_test.go`、`internal/remote/remote_test.go`、`internal/input/keymap_test.go`。
- 新增 `docs/OPEN_QUESTIONS.md`（HostKey 驗證、外掛載入、效能基準、`modified` 排序四項開放問題）。
- 效能基準：`internal/app/app_bench_test.go`、`internal/ui/components_bench_test.go`（10k 排序 ~9–10ms、視窗渲染 ~0.086ms）。

### 修正

- Windows 相容性：`internal/plugin/plugin.go`、`internal/logger/logger.go`（含測試）由 `os.Getenv("HOME")` 改為 `os.UserHomeDir()` 優先、`HOME` 備援。
- 遠端路徑：`internal/remote/remote.go` 遠端拼接由 `filepath.Join` 改為 `path.Join`（POSIX 語義）；修正 `InsecureIgnoreHostKey` 註解錯字並加註生產環境警告。

### 文件

- `TODO.md`：`config.toml` 標註完成；`event.go` 條目修正為「已合併至 keymap.go + app.go」。
- `docs/PRD.md` §8：包結構更新為實作現況（ui/components.go、input/keymap.go 合併說明）。
- `README.md`：覆蓋率表新增 input/app/state/ui/remote 列（確切數字待跑 `go test -cover ./...` 回填）。
