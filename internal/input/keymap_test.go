package input

import (
	"os"
	"path/filepath"
	"testing"
)

// 💡 概念：測試驅動開發 (TDD)
// 說明：先寫測試描述期望行為，再實作功能讓測試通過
// 為何使用：確保 config.toml 解析正確，防止回歸

// 測試自訂 config.toml 能覆寫預設鍵位
func TestLoadKeymapFromPath_CustomValues(t *testing.T) {
	// 建立暫存 config.toml
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")
	content := `[keymap]
up = "k"
down = "j"
open = "l"
back = "h"
delete = "d"
copy = "y"
paste = "p"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("建立測試 config 失敗: %v", err)
	}

	km := LoadKeymapFromPath(cfgPath)

	if km.Up != "k" {
		t.Errorf("Up = %q; want %q", km.Up, "k")
	}
	if km.Down != "j" {
		t.Errorf("Down = %q; want %q", km.Down, "j")
	}
	if km.Open != "l" {
		t.Errorf("Open = %q; want %q", km.Open, "l")
	}
	if km.Back != "h" {
		t.Errorf("Back = %q; want %q", km.Back, "h")
	}
	if km.Copy != "y" {
		t.Errorf("Copy = %q; want %q", km.Copy, "y")
	}
	if km.Paste != "p" {
		t.Errorf("Paste = %q; want %q", km.Paste, "p")
	}
}

// 測試檔案不存在時回傳預設值
func TestLoadKeymapFromPath_MissingFile(t *testing.T) {
	km := LoadKeymapFromPath(filepath.Join(t.TempDir(), "not-exist.toml"))
	def := DefaultKeymap()

	if km.Up != def.Up || km.Down != def.Down || km.Open != def.Open {
		t.Errorf("缺少檔案時應回傳預設值, got %+v", km)
	}
}

// 測試部分覆寫：未指定的欄位維持預設
func TestLoadKeymapFromPath_Partial(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")
	content := `[keymap]
up = "k"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("建立測試 config 失敗: %v", err)
	}

	km := LoadKeymapFromPath(cfgPath)
	def := DefaultKeymap()

	if km.Up != "k" {
		t.Errorf("Up = %q; want %q", km.Up, "k")
	}
	if km.Down != def.Down {
		t.Errorf("Down = %q; want預設 %q", km.Down, def.Down)
	}
}

// 測試無效內容不當機，回傳預設值
func TestLoadKeymapFromPath_Invalid(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(cfgPath, []byte(":::invalid:::\n"), 0644); err != nil {
		t.Fatalf("建立測試 config 失敗: %v", err)
	}

	km := LoadKeymapFromPath(cfgPath)
	if km == nil {
		t.Fatal("無效 config 應回傳預設值，不可為 nil")
	}
}
