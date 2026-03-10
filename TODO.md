# TODO.md

## MVP (Minimum Viable Product) ✅ 已完成

### Phase 1: 專案初始化 ✅

- [x] 初始化 Go module (`go mod init gofm`)
- [x] 安裝 Bubble Tea 與 Lip Gloss 依賴
- [x] 建立 `cmd/gofm/main.go` 入口點

### Phase 2: 資料模型與核心架構 ✅

- [x] 定義 `FileEntry` struct (Name, Path, Size, Mode, ModTime, IsDir)
- [x] 定義 `AppState` struct (CurrentPath, Entries, Cursor, Selected)
- [x] 建立 `internal/state/state.go` 狀態管理
- [x] 建立 `internal/fs/filesystem.go` 檔案系統讀取

### Phase 3: 檔案導航 (File Navigation) ✅

- [x] 實現目錄瀏覽 (↑/k 上移, ↓/j 下移)
- [x] 實現進入資料夾 (Enter/l)
- [x] 實現返回上一層 (h)

### Phase 4: 檔案操作 (File Operations) ✅

- [x] 實現刪除功能 (d)
- [x] 實現重新命名 (r)
- [x] 實現複製 (y)
- [x] 實現貼上 (p)
- [x] 實現新增檔案 (a)
- [x] 實現新增目錄 (A)

### Phase 5: UI 佈局 ✅

- [x] 建立路徑標頭 (PATH bar)
- [x] 建立檔案列表面板 (左側)
- [x] 建立預覽面板 (右側)
- [x] 建立狀態列 (Status bar)

### Phase 6: 鍵位綁定系統 ✅

- [x] 建立 `internal/input/keymap.go` 預設鍵位映射
- [x] 建立 `internal/input/event.go` 事件處理
- [x] 實現 config.toml 載入邏輯 (~/.config/gofm/config.toml) - 部分完成

### Phase 7: 狀態機 ✅

- [x] 实现 Normal 狀態
- [x] 实现 Search 狀態 (Post-MVP)
- [x] 实现 Rename 狀態
- [x] 实现 ConfirmDelete 狀態

### Phase 8: 錯誤處理 ✅

- [x] 處理 permission denied (顯示訊息)
- [x] 處理 file deleted (自動刷新)
- [x] 處理 broken symlink (高亮顯示)

### Phase 9: 日誌系統 ✅

- [x] 建立日誌記錄 (~/.config/gofm/log.txt)

### Phase 10: 效能優化 ✅

- [x] 實現 Lazy load metadata (先讀取目錄名稱，再非同步 stat)

---

## Post-MVP ✅ 已完成

- [x] 搜尋功能 (fuzzy search, 高亮結果)
- [x] 排序功能 (name, size, modified, type)
- [x] 預覽面板 (文字預覽, 圖片資訊, 二進制資訊)
- [x] Git 整合 (M modified, A added, D deleted)
- [x] 外掛系統 (~/.config/gofm/plugins)
- [x] 遠端檔案系統 (SSH, SFTP)

---

## 效能目標

| 指標 | 目標 |
|------|------|
| Startup | < 200ms |
| Navigation | < 16ms |
| Memory | < 50MB |
| Directory render | < 50ms |
