package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"gofm/internal/app"
	"gofm/internal/logger"
)

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
	p := tea.NewProgram(app.New(startPath))

	// 執行應用程式
	if _, err := p.Run(); err != nil {
		// 發生錯誤時，印出訊息並記錄日誌
		logger.Error("Application error: %v", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
