package ui

import (
	"fmt"
	"testing"
	"time"

	"gofm/internal/types"
)

// 💡 概念：基準測試 (Benchmark)
// 說明：量測 10k 檔案目錄的渲染成本，對照 PRD 目標（directory render < 50ms）
// 為何使用：效能迴歸無人把關時，bench 是唯一的量化防線（見 docs/OPEN_QUESTIONS.md 項 3）

// genLargeEntries 產生 n 筆混合目錄/檔案的測試資料
func genLargeEntries(n int) []types.FileEntry {
	entries := make([]types.FileEntry, n)
	now := time.Now()
	for i := range entries {
		if i%10 == 0 {
			entries[i] = types.FileEntry{
				Name:    fmt.Sprintf("dir_%05d", i),
				Path:    fmt.Sprintf("/tmp/dir_%05d", i),
				IsDir:   true,
				ModTime: now,
			}
			continue
		}
		entries[i] = types.FileEntry{
			Name:    fmt.Sprintf("file_%05d.txt", i),
			Path:    fmt.Sprintf("/tmp/file_%05d.txt", i),
			Size:    int64(i * 137 % 1000000),
			ModTime: now,
		}
	}
	return entries
}

// BenchmarkRenderFileList_Full10k 全量渲染 10k 筆（height 撐大到全部可見）
func BenchmarkRenderFileList_Full10k(b *testing.B) {
	entries := genLargeEntries(10000)
	selected := map[string]bool{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderFileList(entries, 5000, selected, 80, len(entries)+4)
	}
}

// BenchmarkRenderFileList_Windowed 正常視窗高度渲染（只畫可見列，驗證滾動優化的效果）
func BenchmarkRenderFileList_Windowed(b *testing.B) {
	entries := genLargeEntries(10000)
	selected := map[string]bool{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderFileList(entries, 5000, selected, 80, 24)
	}
}
