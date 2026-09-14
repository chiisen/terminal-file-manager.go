package app

import (
	"fmt"
	"testing"

	"gofm/internal/types"
)

// 💡 概念：基準測試 (Benchmark)
// 說明：量測 10k 檔案的排序成本（UI 每次切換排序都會走這條路）
// 為何使用：大目錄排序是導航延遲的主要來源，需量化監控

// genSortEntries 產生 n 筆待排序資料（檔名亂序，逼出真實比較成本）
func genSortEntries(n int) []types.FileEntry {
	entries := make([]types.FileEntry, n)
	for i := range entries {
		// 用乘法打散順序，避免已排序輸入低估成本
		j := (i * 7919) % n
		entries[i] = types.FileEntry{
			Name:  fmt.Sprintf("file_%05d.log", j),
			Path:  fmt.Sprintf("/tmp/file_%05d.log", j),
			Size:  int64(j * 131 % 500000),
			IsDir: j%20 == 0,
		}
	}
	return entries
}

// benchmarkSortEntries 是排序基準的共用骨架（每輪用新切片，避免已排序輸入）
func benchmarkSortEntries(b *testing.B, sortBy string, asc bool) {
	base := genSortEntries(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		m := New(b.TempDir())
		m.SortBy = sortBy
		m.SortAsc = asc
		m.Entries = make([]types.FileEntry, len(base))
		copy(m.Entries, base)
		b.StartTimer()
		m.SortEntries()
	}
}

func BenchmarkSortEntries_NameAsc10k(b *testing.B) {
	benchmarkSortEntries(b, "name", true)
}

func BenchmarkSortEntries_SizeDesc10k(b *testing.B) {
	benchmarkSortEntries(b, "size", false)
}
