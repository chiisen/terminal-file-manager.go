package plugin

import (
	"fmt"
	"os"
	"path/filepath"
)

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：Plugin System
// 說明：提供外掛系統，可以載入和使用第三方外掛
// 為何使用：讓使用者擴展 gofm 的功能
// ══════════════════════════════════════════════════════════════════════════════

// Plugin 外掛介面
type Plugin interface {
	// Name 回傳外掛名稱
	Name() string
	// Description 回傳外掛描述
	Description() string
	// Init 初始化外掛
	Init() error
	// Execute 執行外掛
	Execute(args []string) (string, error)
}

// Manager 外掛管理器
type Manager struct {
	plugins map[string]Plugin
}

// NewManager 建立外掛管理器
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
	}
}

// GetPluginsDir 回傳外掛目錄路徑
func GetPluginsDir() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config", "gofm", "plugins")
}

// LoadPlugins 載入所有外掛
// 目前這是一個簡化的實現，真正的外掛系統需要動態載入 .so 或 .go 檔案
func (m *Manager) LoadPlugins() error {
	pluginsDir := GetPluginsDir()

	// 檢查外掛目錄是否存在
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		// 建立外掛目錄
		if err := os.MkdirAll(pluginsDir, 0755); err != nil {
			return err
		}
		return nil
	}

	// 讀取外掛目錄中的所有檔案
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return err
	}

	// 目前暫時不自動載入外掛（需要實現動態載入）
	// 這裡只是預留結構
	for _, entry := range entries {
		if entry.IsDir() {
			// 嘗試載入目錄中外掛
			pluginPath := filepath.Join(pluginsDir, entry.Name())
			fmt.Printf("Found plugin directory: %s\n", pluginPath)
		}
	}

	return nil
}

// RegisterPlugin 註冊外掛
func (m *Manager) RegisterPlugin(p Plugin) error {
	name := p.Name()
	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin already exists: %s", name)
	}

	if err := p.Init(); err != nil {
		return fmt.Errorf("failed to init plugin %s: %w", name, err)
	}

	m.plugins[name] = p
	return nil
}

// GetPlugin 取得外掛
func (m *Manager) GetPlugin(name string) (Plugin, error) {
	p, ok := m.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return p, nil
}

// ListPlugins 列出所有已註冊的外掛
func (m *Manager) ListPlugins() []Plugin {
	plugins := make([]Plugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// ExecutePlugin 執行外掛
func (m *Manager) ExecutePlugin(name string, args []string) (string, error) {
	p, err := m.GetPlugin(name)
	if err != nil {
		return "", err
	}
	return p.Execute(args)
}

// CreateSamplePlugin 創建一個範例外掛
func CreateSamplePlugin() Plugin {
	return &samplePlugin{
		name:        "sample",
		description: "A sample plugin for demonstration",
	}
}

// samplePlugin 範例外掛
type samplePlugin struct {
	name        string
	description string
}

func (p *samplePlugin) Name() string {
	return p.name
}

func (p *samplePlugin) Description() string {
	return p.description
}

func (p *samplePlugin) Init() error {
	return nil
}

func (p *samplePlugin) Execute(args []string) (string, error) {
	return fmt.Sprintf("Sample plugin executed with %d arguments", len(args)), nil
}
