package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"gofm/internal/app"
	"gofm/internal/logger"
)

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：ANSI Escape Code
// 說明：用於控制終端機輸出的特殊字序列
//       \033[2J 表示清除整個螢幕並將游標移到起始位置
// 為何使用：確保程式全螢幕顯示前先清除舊的畫面殘留
// ══════════════════════════════════════════════════════════════════════════════

// clearScreen 是 ANSI escape code，用於清除終端機畫面
const clearScreen = "\033[2J"

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：Bubble Tea
// 說明：Go 語言的 TUI 框架，採用 MVC 模式（Model-View-Controller）
//       - Model: 存放應用程式狀態 (AppState)
//       - View: 根據狀態渲染 UI
//       - Controller: 處理輸入事件 (Update 方法)
// 為何使用：事件驅動、反應式 UI，適合開發互動式終端應用
// ══════════════════════════════════════════════════════════════════════════════

func main() {
	// 初始化日誌系統
	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize logger: %v\n", err)
	}
	defer logger.Close()

	// 決定起始目錄（優先使用命令列參數，否則使用目前目錄）
	startPath := "."
	if len(os.Args) > 1 {
		startPath = os.Args[1]
	}

	// 初始化應用程式
	// tea.WithAltScreen() 切換到替代螢幕緩衝區，實現全螢幕效果
	p := tea.NewProgram(
		app.New(startPath),
		tea.WithAltScreen(),
	)

	// 執行應用程式
	if _, err := p.Run(); err != nil {
		// 發生錯誤時，印出訊息並記錄日誌
		logger.Error("Application error: %v", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
