# CHANGELOG.md

## [Unreleased]

### 新增

- 新增 SSH/SFTP TUI 入口與埠、私鑰、known_hosts、下載目錄旗標；背景瀏覽、重連與串流下載，完整下載後發布且不覆寫目的檔。取消握手／傳輸並等待暫存清理，補真實 SSH 金鑰認證整合與慢速事件測試（#13）。
- 將自訂鍵位設定接入 TUI 的既有操作入口；輸入模式保留字元輸入，衝突／無效設定回退預設，狀態列顯示有效鍵位，固定方向鍵、Enter 與 Ctrl+C（#11）。
- `internal/input/keymap.go`：實作 `config.toml` 的 `[keymap]` 解析（`LoadKeymapFromPath`，標準函式庫、無新依賴）。
- 核心單元測試：`internal/app/app_test.go`、`internal/state/state_test.go`、`internal/ui/components_test.go`、`internal/remote/remote_test.go`、`internal/input/keymap_test.go`。
- 新增 `docs/OPEN_QUESTIONS.md`（HostKey 驗證、外掛載入、效能基準、`modified` 排序四項開放問題）。
- 效能基準：`internal/app/app_bench_test.go`、`internal/ui/components_bench_test.go`（10k 排序 ~9–10ms、視窗渲染 ~0.086ms）。

### 修正

- SSH 預設使用 known_hosts 驗證主機金鑰，支援指定驗證檔案，拒絕未知／變更金鑰並提供明確錯誤；新增真實握手及非預設埠測試（#10）。
- 修正 TUI 左右面板排版與寬高限制；窄視窗切換列表／預覽，保留狀態列與輸入尾端；以終端顯示寬度處理中文與 emoji 截斷及欄位對齊，新增預覽翻頁捲動與視窗尺寸回歸測試（#9）。
- 遠端讀取正確傳遞 SFTP 中斷與 writer 錯誤，新增固定緩衝的 `Download` 串流 API；以真實 SFTP 封包測試 EOF、中斷及大檔串流，README 說明尚未接入遠端 TUI（#8）。
- 搜尋 Enter 確認目前選取並開啟檔案／目錄，保留對應游標且還原完整列表；修正空查詢導航與結果數量，空結果保持搜尋模式（#7）。
- 實作修改時間升降序排序，目錄優先、未知時間置後，同時間依名稱排序；檔案系統與 UI 共用排序邏輯，新增真實檔案時間回歸測試（#6）。
- 預覽改為背景命令與單份快取，渲染不再讀檔並忽略過期結果；檔案操作非同步執行，顯示執行／成功／失敗狀態並阻擋重複提交，失敗移動保留剪貼簿（#5）。
- 修正目錄、詳細資訊與 Git 的非同步載入：背景命令回傳完成訊息，由事件迴圈更新狀態；以請求編號忽略過期結果，依路徑匹配 metadata 並保留搜尋與游標狀態（#1）。
- 複製前阻擋來源自身、硬連結及來源子目錄，解析目的地祖先的符號連結；複製來源符號連結時保留連結，避免遞迴循環與來源截斷（#2）。
- 修正中文及 emoji 的文字辨識、控制字元清理與預覽截斷，保留有效 UTF-8 內容（#3）。
- 移除會遺失快速輸入的全域按鍵節流，支援 Unicode、多字元貼上與按字元退格；搜尋模式的 `j/k` 改為查詢字元（#4）。
- 新增非同步載入、過期結果、複製路徑防護與 Unicode 的回歸測試。

- Windows 相容性：`internal/plugin/plugin.go`、`internal/logger/logger.go`（含測試）由 `os.Getenv("HOME")` 改為 `os.UserHomeDir()` 優先、`HOME` 備援。
- 遠端路徑：`internal/remote/remote.go` 遠端拼接由 `filepath.Join` 改為 `path.Join`（POSIX 語義）；修正 `InsecureIgnoreHostKey` 註解錯字並加註生產環境警告。

### 文件

- 修正 README、TODO 與 PRD 的外掛可用程度；新增外部程序協定設計決策、信任邊界及後續實作驗收，保留動態載入為待辦（#12）。
- `TODO.md`：`config.toml` 標註完成；`event.go` 條目修正為「已合併至 keymap.go + app.go」。
- `docs/PRD.md` §8：包結構更新為實作現況（ui/components.go、input/keymap.go 合併說明）。
- `README.md`：覆蓋率表新增 input/app/state/ui/remote 列（確切數字待跑 `go test -cover ./...` 回填）。
