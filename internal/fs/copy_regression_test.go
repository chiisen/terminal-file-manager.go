package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyRejectsSameFileWithoutTruncation(t *testing.T) {
	p := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(p, []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := CopyFile(p, p); err == nil {
		t.Error("應拒絕複製到同一檔案")
	}
	data, err := os.ReadFile(p)
	if err != nil || string(data) != "keep" {
		t.Fatalf("來源遭截斷: %q, %v", data, err)
	}
}

func TestCopyRejectsDescendantBeforeCreatingDestination(t *testing.T) {
	for _, suffix := range []string{"copy", filepath.Join("child", "copy")} {
		t.Run(suffix, func(t *testing.T) {
			src := t.TempDir()
			if err := os.Mkdir(filepath.Join(src, "child"), 0755); err != nil {
				t.Fatal(err)
			}
			dst := filepath.Join(src, suffix)
			if err := CopyFile(src, dst); err == nil {
				t.Fatal("應拒絕複製到子目錄")
			}
			if _, err := os.Lstat(dst); !os.IsNotExist(err) {
				t.Fatalf("拒絕前已寫入目的地: %v", err)
			}
		})
	}
}

func TestCopyRejectsSymlinkDestinationInsideSource(t *testing.T) {
	src := t.TempDir()
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(src, alias); err != nil {
		t.Skipf("目前環境不支援符號連結: %v", err)
	}
	dst := filepath.Join(alias, "copy")
	if err := CopyFile(src, dst); err == nil {
		t.Fatal("符號連結不得繞過子目錄檢查")
	}
	if _, err := os.Lstat(dst); !os.IsNotExist(err) {
		t.Fatalf("拒絕前已建立目的地: %v", err)
	}
}

func TestCopyRejectsHardLinkWithoutTruncation(t *testing.T) {
	src := filepath.Join(t.TempDir(), "original")
	dst := filepath.Join(filepath.Dir(src), "alias")
	if err := os.WriteFile(src, []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(src, dst); err != nil {
		t.Skipf("不支援硬連結: %v", err)
	}
	if err := CopyFile(src, dst); err == nil {
		t.Fatal("應拒絕同一檔案的硬連結")
	}
	data, err := os.ReadFile(src)
	if err != nil || string(data) != "keep" {
		t.Fatalf("來源遭截斷: %q %v", data, err)
	}
}

func TestCopyPreservesDirectorySymlinkCycle(t *testing.T) {
	src := t.TempDir()
	link := filepath.Join(src, "loop")
	if err := os.Symlink(src, link); err != nil {
		t.Skipf("不支援符號連結: %v", err)
	}
	dst := filepath.Join(t.TempDir(), "copy")
	if err := CopyFile(src, dst); err != nil {
		t.Fatal(err)
	}
	if target, err := os.Readlink(filepath.Join(dst, "loop")); err != nil || target != src {
		t.Fatalf("連結未保留: %q %v", target, err)
	}
}
