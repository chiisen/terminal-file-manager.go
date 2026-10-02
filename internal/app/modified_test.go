package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadedMetadataSupportsModifiedSorting(t *testing.T) {
	dir := t.TempDir()
	old := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for name, stamp := range map[string]time.Time{"z-old": old, "a-new": old.Add(time.Hour)} {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("a"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
	m := New(dir)
	m.SortBy = "modified"
	_, metadata := m.Update(m.Init()())
	m.Update(metadata())
	if len(m.Entries) != 2 || m.Entries[0].Name != "z-old" || !m.Entries[0].ModTime.Equal(old) {
		t.Fatalf("讀取後未依時間排序: %+v", m.Entries)
	}
}
