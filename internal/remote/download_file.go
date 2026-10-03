package remote

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"
)

type Downloader interface {
	Download(string, io.Writer) (int64, error)
}

// DownloadToDirectory 先寫暫存，再以同目錄 hard link 發布，避免覆寫或露出半份檔案。
func DownloadToDirectory(client Downloader, remotePath, localDir string) (destination string, written int64, err error) {
	name := path.Base(remotePath)
	if !safeDownloadName(name) {
		return "", 0, fmt.Errorf("不安全的下載檔名: %q", name)
	}
	if client == nil {
		return "", 0, fmt.Errorf("SFTP 尚未連線")
	}
	destination = filepath.Join(localDir, name)
	if _, statErr := os.Lstat(destination); statErr == nil {
		return destination, 0, fmt.Errorf("目的檔案已存在: %s", name)
	} else if !os.IsNotExist(statErr) {
		return destination, 0, statErr
	}
	file, err := os.CreateTemp(localDir, ".gofm-download-*")
	if err != nil {
		return destination, 0, fmt.Errorf("無法建立下載暫存檔: %w", err)
	}
	tempPath := file.Name()
	defer func() {
		file.Close()
		if cleanupErr := os.Remove(tempPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			err = fmt.Errorf("下載暫存清理失敗: %v（原始錯誤: %v）", cleanupErr, err)
		}
	}()
	written, err = client.Download(remotePath, file)
	if err != nil {
		return destination, written, err
	}
	if err = file.Sync(); err != nil {
		return destination, written, err
	}
	if err = file.Close(); err != nil {
		return destination, written, err
	}
	// Link 的目的地存在時必定失敗，涵蓋傳輸期間另一程序建立目的檔的競態。
	if err = os.Link(tempPath, destination); err != nil {
		return destination, written, fmt.Errorf("無法發布下載（不覆寫，目的檔案系統須支援 hard link）: %w", err)
	}
	return destination, written, nil
}

func safeDownloadName(name string) bool {
	if name == "" || name == "." || name == ".." || name == "/" || strings.TrimRight(name, " .") != name {
		return false
	}
	for _, r := range name {
		if unicode.IsControl(r) || strings.ContainsRune(`<>:"/\|?*`, r) {
			return false
		}
	}
	stem := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
	switch stem {
	case "CON", "PRN", "AUX", "NUL", "CONIN$", "CONOUT$":
		return false
	}
	if runes := []rune(stem); len(runes) == 4 && (strings.HasPrefix(stem, "COM") || strings.HasPrefix(stem, "LPT")) && strings.ContainsRune("123456789¹²³", runes[3]) {
		return false
	}
	return true
}
