package preview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPreviewPreservesUnicode(t *testing.T) {
	for _, tc := range []struct{ content, want string }{
		{"繁體中文🙂", "繁體中文🙂"},
		{"hello 中文🙂\x01 world", "hello 中文🙂 world"},
		{strings.Repeat("a", 4095) + "中", strings.Repeat("a", 4095) + "中"},
		{strings.Repeat("a", 65535) + "中", strings.Repeat("a", 65535)},
		{strings.Repeat("a", 65532) + "🙂", strings.Repeat("a", 65532) + "🙂"},
	} {
		p := filepath.Join(t.TempDir(), "中文.txt")
		if err := os.WriteFile(p, []byte(tc.content), 0644); err != nil {
			t.Fatal(err)
		}
		result, err := GetPreview(p)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsBinary || !utf8.ValidString(result.Content) {
			t.Fatalf("文字誤判或 UTF-8 損壞: binary=%v", result.IsBinary)
		}
		_, content, ok := strings.Cut(result.Content, "\n\n")
		if !ok || content != tc.want {
			t.Fatalf("內容或截斷錯誤: got %d bytes, want %d bytes", len(content), len(tc.want))
		}
	}
}

func TestUnicodeTextAndControlCleaning(t *testing.T) {
	if !isText([]byte("中文🙂")) {
		t.Error("中文應視為文字")
	}
	if isText([]byte{'a', 0xff}) || isText([]byte("中\x00文")) {
		t.Error("無效 UTF-8 或 NULL 應視為二進位")
	}
	if got := cleanControlChars("中\x01文🙂\u0085\n"); got != "中文🙂\n" {
		t.Errorf("清理後內容 = %q", got)
	}
}
