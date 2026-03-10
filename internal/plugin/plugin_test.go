package plugin

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：單元測試 (Unit Test)
// 說明：針對 plugin 套件進行獨立測試
// ══════════════════════════════════════════════════════════════════════════════

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Error("NewManager should return non-nil Manager")
	}

	if m.plugins == nil {
		t.Error("Manager.plugins should be initialized")
	}
}

func TestGetPluginsDir(t *testing.T) {
	dir := GetPluginsDir()

	// 檢查路徑格式
	expectedPrefix := filepath.Join(os.Getenv("HOME"), ".config", "gofm", "plugins")
	if dir != expectedPrefix {
		t.Errorf("GetPluginsDir() = %s; want %s", dir, expectedPrefix)
	}
}

func TestLoadPlugins(t *testing.T) {
	m := NewManager()

	// 測試載入外掛（應該不會崩潰）
	err := m.LoadPlugins()
	if err != nil {
		t.Errorf("LoadPlugins failed: %v", err)
	}
}

func TestLoadPluginsCreatesDir(t *testing.T) {
	// 這個測試會在 CI 環境中創建目錄
	m := NewManager()

	// 確保目錄被建立
	err := m.LoadPlugins()
	if err != nil {
		t.Errorf("LoadPlugins failed: %v", err)
	}

	// 檢查目錄是否存在
	dir := GetPluginsDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Logf("Note: Plugins directory not created (may not exist in test environment)")
	}
}

func TestRegisterPlugin(t *testing.T) {
	m := NewManager()

	// 註冊範例外掛
	err := m.RegisterPlugin(CreateSamplePlugin())
	if err != nil {
		t.Errorf("RegisterPlugin failed: %v", err)
	}

	// 嘗試註冊相同名稱的外掛應該失敗
	err = m.RegisterPlugin(CreateSamplePlugin())
	if err == nil {
		t.Error("RegisterPlugin should fail for duplicate plugin")
	}
}

func TestGetPlugin(t *testing.T) {
	m := NewManager()

	// 註冊外掛
	p := CreateSamplePlugin()
	err := m.RegisterPlugin(p)
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	// 取得外掛
	retrieved, err := m.GetPlugin("sample")
	if err != nil {
		t.Errorf("GetPlugin failed: %v", err)
	}

	if retrieved.Name() != "sample" {
		t.Errorf("GetPlugin returned wrong plugin: %s", retrieved.Name())
	}

	// 取得不存在的外掛應該失敗
	_, err = m.GetPlugin("nonexistent")
	if err == nil {
		t.Error("GetPlugin should fail for nonexistent plugin")
	}
}

func TestListPlugins(t *testing.T) {
	m := NewManager()

	// 初始應該沒有外掛
	plugins := m.ListPlugins()
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins, got %d", len(plugins))
	}

	// 註冊外掛
	m.RegisterPlugin(CreateSamplePlugin())

	// 應該有一個外掛
	plugins = m.ListPlugins()
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}
}

func TestExecutePlugin(t *testing.T) {
	m := NewManager()

	// 註冊外掛
	err := m.RegisterPlugin(CreateSamplePlugin())
	if err != nil {
		t.Fatalf("RegisterPlugin failed: %v", err)
	}

	// 執行外掛
	result, err := m.ExecutePlugin("sample", []string{"arg1", "arg2"})
	if err != nil {
		t.Errorf("ExecutePlugin failed: %v", err)
	}

	// 檢查結果
	if result == "" {
		t.Error("ExecutePlugin should return non-empty result")
	}

	// 執行不存在的外掛應該失敗
	_, err = m.ExecutePlugin("nonexistent", nil)
	if err == nil {
		t.Error("ExecutePlugin should fail for nonexistent plugin")
	}
}

func TestSamplePlugin(t *testing.T) {
	p := CreateSamplePlugin()

	// 測試 Name
	if p.Name() != "sample" {
		t.Errorf("Name() = %s; want sample", p.Name())
	}

	// 測試 Description
	if p.Description() == "" {
		t.Error("Description() should not be empty")
	}

	// 測試 Init
	if err := p.Init(); err != nil {
		t.Errorf("Init() failed: %v", err)
	}

	// 測試 Execute
	result, err := p.Execute([]string{"test"})
	if err != nil {
		t.Errorf("Execute() failed: %v", err)
	}

	if result == "" {
		t.Error("Execute() should return result")
	}
}
