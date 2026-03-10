package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：Git Integration
// 說明：負責偵測 Git 倉庫狀態，並提供檔案的 Git 狀態標記
// 為何使用：讓使用者快速識別哪些檔案被修改、新增或刪除
// ══════════════════════════════════════════════════════════════════════════════

// GitStatus 代表 Git 狀態
type GitStatus string

const (
	StatusModified  GitStatus = "M" // 已修改
	StatusAdded     GitStatus = "A" // 新增
	StatusDeleted   GitStatus = "D" // 刪除
	StatusUntracked GitStatus = "?" // 未追蹤
	StatusStaged    GitStatus = "S" // 已暫存
	StatusUnknown   GitStatus = ""
)

// GitInfo 儲存 Git 資訊
type GitInfo struct {
	IsRepo    bool           // 是否為 Git 倉庫
	RootPath  string         // Git 倉庫根目錄
	StatusMap map[string]GitStatus // 檔案路徑對應的狀態
}

// GetGitInfo 取得目錄的 Git 資訊
// 參數 path 是要檢查的目錄路徑
// 回傳 GitInfo
func GetGitInfo(path string) (*GitInfo, error) {
	// 向上尋找 .git 目錄
	gitRoot, err := findGitRoot(path)
	if err != nil {
		return &GitInfo{
			IsRepo:    false,
			RootPath:  "",
			StatusMap: make(map[string]GitStatus),
		}, nil
	}

	// 執行 git status 取得狀態
	statusMap, err := getGitStatus(gitRoot)
	if err != nil {
		return &GitInfo{
			IsRepo:    true,
			RootPath:  gitRoot,
			StatusMap: make(map[string]GitStatus),
		}, err
	}

	return &GitInfo{
		IsRepo:    true,
		RootPath:  gitRoot,
		StatusMap: statusMap,
	}, nil
}

// findGitRoot 尋找 Git 倉庫根目錄
func findGitRoot(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	// 向上尋找 .git 目錄
	current := absPath
	for {
		gitPath := filepath.Join(current, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return current, nil
		}

		// 檢查是否已經到達根目錄
		parent := filepath.Dir(current)
		if parent == current {
			return "", os.ErrNotExist
		}
		current = parent
	}
}

// getGitStatus 取得 Git 狀態
func getGitStatus(rootPath string) (map[string]GitStatus, error) {
	// 執行 git status --porcelain 取得簡潔格式的狀態
	cmd := exec.Command("git", "status", "--porcelain", "-uall")
	cmd.Dir = rootPath

	output, err := cmd.Output()
	if err != nil {
		// 如果不是 Git 倉庫，回傳空 map
		if strings.Contains(err.Error(), "not a git repository") {
			return make(map[string]GitStatus), nil
		}
		return nil, err
	}

	statusMap := make(map[string]GitStatus)

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		// 解析狀態行：XY filename
		indexStatus := line[0]
		workTreeStatus := line[1]
		filename := strings.TrimSpace(line[3:])

		// 決定狀態
		var status GitStatus

		// 優先處理 index status (已暫存的變更)
		switch indexStatus {
		case 'A':
			status = StatusStaged
		case 'M':
			status = StatusModified
		case 'D':
			status = StatusDeleted
		default:
			// 如果 index 沒有狀態，使用 worktree 的狀態
			switch workTreeStatus {
			case '?':
				status = StatusUntracked
			case 'M', ' ':
				status = StatusModified
			case 'D':
				status = StatusDeleted
			default:
				status = StatusUnknown
			}
		}

		if status != StatusUnknown {
			statusMap[filename] = status
		}
	}

	return statusMap, nil
}

// GetFileStatus 取得特定檔案的 Git 狀態
// 參數 info 是 GitInfo
// 參數 filename 是檔案名稱
// 回傳 GitStatus
func (info *GitInfo) GetFileStatus(filename string) GitStatus {
	if !info.IsRepo {
		return StatusUnknown
	}

	// 嘗試直接匹配
	if status, ok := info.StatusMap[filename]; ok {
		return status
	}

	// 嘗試相對路徑匹配
	relPath, err := filepath.Rel(info.RootPath, filename)
	if err != nil {
		return StatusUnknown
	}

	if status, ok := info.StatusMap[relPath]; ok {
		return status
	}

	// 嘗試移除前導 "./"
	relPath = strings.TrimPrefix(relPath, "./")
	if status, ok := info.StatusMap[relPath]; ok {
		return status
	}

	return StatusUnknown
}

// GetStatusIcon 取得狀態對應的圖示
func GetStatusIcon(status GitStatus) string {
	switch status {
	case StatusModified:
		return "M"
	case StatusAdded, StatusStaged:
		return "A"
	case StatusDeleted:
		return "D"
	case StatusUntracked:
		return "?"
	default:
		return ""
	}
}
