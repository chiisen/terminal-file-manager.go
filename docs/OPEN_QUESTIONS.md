# OPEN_QUESTIONS.md

本文件收斂目前已知、尚未決定的開放問題（供 `AI_GITHUB_PROJECT_GUIDE.md` SOP 匯入追蹤）。

## 1. 安全：SSH HostKey 驗證（P1）

- 現況：`internal/remote/remote.go` 使用 `ssh.InsecureIgnoreHostKey()`，僅適合測試/內網。
- 待決定：是否引入 `known_hosts` 驗證（`golang.org/x/crypto/ssh/knownhosts`）？預設開啟還是用 flag 選擇？
- 影響：生產環境中間人攻擊風險。

## 2. 外掛動態載入（P2）

- 現況：`internal/plugin/plugin.go` 的 `LoadPlugins()` 只建目錄＋列出，不實際載入 `.so`/腳本。
- 待決定：外掛機制要 Go plugin（`.so`，版本耦合嚴格）還是外部進程（exec 腳本，鬆耦合）？
- 影響：TODO Post-MVP「外掛系統」目前只是骨架。

## 3. 效能目標量測（P2）✅ 已落地（2026-09-14，i7-12700H）

- 基準：`internal/ui/components_bench_test.go`、`internal/app/app_bench_test.go`（10k 筆）。
- 實測：排序 name asc ~9.8ms / size desc ~9.2ms；視窗化渲染 ~0.086ms（目標 navigation < 16ms ✅）；全量 10k 渲染 ~1.48s（ pathological case，實際 UI 只畫可見列，不影響）。
- 待辦：進 CI（`go test -bench=. -run=^$`）即結案。

## 4. 排序 `modified` 實作（P3）✅ 已完成

- `internal/app` 與 `internal/fs` 共用 `SortByModified`，依 `FileEntry.ModTime` 實作升降序。
- 目錄優先；未知時間在各群組內置後；相同時間以名稱升序排列。
- 單元測試涵蓋升降序、零值與相同時間；以真實檔案時間驗證載入後排序（#6）。
