# OPEN_QUESTIONS.md

本文件收斂目前已知、尚未決定的開放問題（供 `AI_GITHUB_PROJECT_GUIDE.md` SOP 匯入追蹤）。

## 1. 安全：SSH HostKey 驗證（P1）✅ 已完成（#10）

- 預設使用 `~/.ssh/known_hosts`，可透過 `Config.KnownHostsPath` 指定驗證檔案。
- 未知、變更或無法讀取的主機金鑰一律拒絕；不提供自動接受或跳過驗證。
- 以本機真實 SSH/SFTP 握手驗證相符、未知、變更、缺檔與非預設埠情境。

## 2. 外掛設計決策（P2）✅ 已完成釐清（#12）

- 現況：`internal/plugin/plugin.go` 的 `LoadPlugins()` 只建目錄＋列出，不實際載入 `.so`/腳本。
- 決策：後續使用外部程序與 stdin/stdout JSON，以支援 Windows／Linux 並降低工具鏈耦合。
- [PLUGIN_DESIGN.md](PLUGIN_DESIGN.md) 記錄協定、信任邊界、錯誤處理與後續驗收；動態載入本身尚未實作。

## 3. 效能目標量測（P2）✅ 已落地（2026-09-14，i7-12700H）

- 基準：`internal/ui/components_bench_test.go`、`internal/app/app_bench_test.go`（10k 筆）。
- 實測：排序 name asc ~9.8ms / size desc ~9.2ms；視窗化渲染 ~0.086ms（目標 navigation < 16ms ✅）；全量 10k 渲染 ~1.48s（ pathological case，實際 UI 只畫可見列，不影響）。
- 待辦：進 CI（`go test -bench=. -run=^$`）即結案。

## 4. 排序 `modified` 實作（P3）✅ 已完成

- `internal/app` 與 `internal/fs` 共用 `SortByModified`，依 `FileEntry.ModTime` 實作升降序。
- 目錄優先；未知時間在各群組內置後；相同時間以名稱升序排列。
- 單元測試涵蓋升降序、零值與相同時間；以真實檔案時間驗證載入後排序（#6）。
