package remote

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	"gofm/internal/types"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// ══════════════════════════════════════════════════════════════════════════════
// 💡 概念：Remote Filesystem (SSH/SFTP)
// 說明：提供遠端檔案系統的支援，透過 SSH/SFTP 協議連接到遠端伺服器
// 為何使用：讓使用者可以透過網路瀏覽和操作遠端伺服器上的檔案
// ══════════════════════════════════════════════════════════════════════════════

// Config SSH 連線配置
type Config struct {
	Host     string // 主機地址
	Port     string // 連接埠 (預設 22)
	User     string // 使用者名稱
	Password string // 密碼 (如果使用金鑰認證則為空)
	KeyPath  string // SSH 金鑰路徑
}

// RemoteClient SFTP 客戶端
type RemoteClient struct {
	config *Config
	client *sftp.Client
	ssh    *ssh.Client
}

// NewRemoteClient 建立遠端客戶端
func NewRemoteClient(config *Config) (*RemoteClient, error) {
	// 設定預設連接埠
	if config.Port == "" {
		config.Port = "22"
	}

	// 建立 SSH 客戶端
	var authMethods []ssh.AuthMethod

	if config.KeyPath != "" {
		// 使用金鑰認證
		key, err := os.ReadFile(config.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("無法讀取 SSH 金鑰: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("無法解析 SSH 金鑰: %w", err)
		}

		authMethods = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else if config.Password != "" {
		// 使用密碼認證
		authMethods = []ssh.AuthMethod{ssh.Password(config.Password)}
	}

	// 建立 SSH 客戶端配置
	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: authMethods,
		// ⚠️ 注意：InsecureIgnoreHostKey 會跳過主機金鑰驗證，僅適合測試/內網；
		// 生產環境應改用 known_hosts 驗證的 HostKeyCallback
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// 連接到 SSH 伺服器
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("無法連接到 SSH 伺服器: %w", err)
	}

	// 建立 SFTP 客戶端
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("無法建立 SFTP 客戶端: %w", err)
	}

	return &RemoteClient{
		config: config,
		client: sftpClient,
		ssh:    sshClient,
	}, nil
}

// Close 關閉連線
func (c *RemoteClient) Close() error {
	if c.client != nil {
		c.client.Close()
	}
	if c.ssh != nil {
		c.ssh.Close()
	}
	return nil
}

// ReadDirectory 讀取遠端目錄
func (c *RemoteClient) ReadDirectory(dirPath string) ([]types.FileEntry, error) {
	entries, err := c.client.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	result := make([]types.FileEntry, 0, len(entries))
	for _, entry := range entries {
		// sftp.FileInfo 已經實現了 os.FileInfo 介面，不需要再調用 Info()
		// 💡 注意：遠端一律是 POSIX 路徑，必須用 path.Join；
		// filepath.Join 在 Windows 會產生反斜線，導致遠端路徑錯誤
		result = append(result, types.FileEntry{
			Name:  entry.Name(),
			Path:  path.Join(dirPath, entry.Name()),
			Size:  entry.Size(),
			IsDir: entry.IsDir(),
			Mode:  entry.Mode().String(),
		})
	}

	return result, nil
}

// Get 取得小檔案內容；大檔案請使用 Download 串流到檔案或其他 writer。
func (c *RemoteClient) Get(remotePath string) ([]byte, error) {
	var content bytes.Buffer
	if _, err := c.Download(remotePath, &content); err != nil {
		return nil, err
	}
	return content.Bytes(), nil
}

// Download 串流下載，不累積完整內容；回傳已寫入的 byte 數與讀取／寫入錯誤。
func (c *RemoteClient) Download(remotePath string, dst io.Writer) (written int64, err error) {
	if c == nil || c.client == nil {
		return 0, fmt.Errorf("SFTP 尚未連線")
	}
	if dst == nil {
		return 0, fmt.Errorf("下載目的 writer 不可為 nil")
	}
	file, err := c.client.Open(remotePath)
	if err != nil {
		return 0, fmt.Errorf("無法開啟遠端檔案 %s: %w", remotePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("關閉遠端檔案 %s 失敗: %w", remotePath, closeErr)
		}
	}()
	// 隱藏 WriterTo／ReaderFrom 快速路徑，固定以 32KB buffer 順序讀取與寫出。
	written, err = io.CopyBuffer(struct{ io.Writer }{dst}, struct{ io.Reader }{file}, make([]byte, 32*1024))
	if err != nil {
		return written, fmt.Errorf("下載 %s 失敗（已寫入 %d bytes）: %w", remotePath, written, err)
	}
	return written, nil
}

// Put 上傳檔案
func (c *RemoteClient) Put(localPath, remotePath string) error {
	localFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	remoteFile, err := c.client.Create(remotePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	_, err = remoteFile.ReadFrom(localFile)
	return err
}

// Mkdir 建立目錄
func (c *RemoteClient) Mkdir(remotePath string) error {
	return c.client.MkdirAll(remotePath)
}

// Remove 刪除檔案或目錄
func (c *RemoteClient) Remove(remotePath string) error {
	return c.client.Remove(remotePath)
}

// Rename 重新命名
func (c *RemoteClient) Rename(oldPath, newPath string) error {
	return c.client.Rename(oldPath, newPath)
}

// Stat 取得檔案資訊
func (c *RemoteClient) Stat(remotePath string) (os.FileInfo, error) {
	return c.client.Stat(remotePath)
}

// ParseRemotePath 解析遠端路徑
// 格式: user@host:/path/to/dir 或 user@host:path/to/dir
func ParseRemotePath(remotePath string) (user, host, remoteDir string, err error) {
	// 檢查是否為遠端路徑
	if !isRemotePath(remotePath) {
		return "", "", "", fmt.Errorf("無效的遠端路徑格式: %s", remotePath)
	}

	// 解析格式: user@host:/path
	// 或 user@host:path
	parts := splitAtFirst(remotePath, "@")
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("無效的使用者@主機格式: %s", remotePath)
	}

	user = parts[0]
	hostParts := splitAtFirst(parts[1], ":")
	if len(hostParts) != 2 {
		return "", "", "", fmt.Errorf("無效的主機:路徑格式: %s", remotePath)
	}

	host = hostParts[0]
	remoteDir = hostParts[1]

	// 確保路徑以 / 開頭
	if !path.IsAbs(remoteDir) {
		remoteDir = "/" + remoteDir
	}

	return user, host, remoteDir, nil
}

// isRemotePath 檢查是否為遠端路徑
func isRemotePath(p string) bool {
	// 包含 @ 和 : 且 : 在 @ 之後
	atIdx := -1
	colonIdx := -1

	for i := 0; i < len(p); i++ {
		if p[i] == '@' && atIdx == -1 {
			atIdx = i
		}
		if p[i] == ':' && atIdx != -1 && colonIdx == -1 {
			colonIdx = i
			break
		}
	}

	return atIdx != -1 && colonIdx != -1
}

// splitAtFirst 在第一個分隔符處分割字串
func splitAtFirst(s, sep string) []string {
	for i := 0; i < len(s); i++ {
		if len(s) >= i+len(sep) && s[i:i+len(sep)] == sep {
			return []string{s[:i], s[i+len(sep):]}
		}
	}
	return []string{s}
}
