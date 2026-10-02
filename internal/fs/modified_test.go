package fs

import (
	"gofm/internal/types"
	"testing"
	"time"
)

func TestSortEntriesModifiedUsesTimes(t *testing.T) {
	old := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []types.FileEntry{{Name: "a-new", ModTime: old.Add(time.Hour)}, {Name: "unknown"}, {Name: "z-old", ModTime: old}, {Name: "dir", IsDir: true}}
	SortEntries(entries, "modified")
	for i, want := range []string{"dir", "z-old", "a-new", "unknown"} {
		if entries[i].Name != want {
			t.Fatalf("index=%d got=%q want=%q", i, entries[i].Name, want)
		}
	}
}
