package remote

import (
	"testing"
)

// 💡 概念：遠端路徑解析測試
// 說明：只測純字串解析，不建立 SSH 連線（避免依賴外部伺服器）
// 為何使用：路徑格式錯誤應早失敗，而非連線後才報錯

func TestParseRemotePath_Valid(t *testing.T) {
	tests := []struct {
		input           string
		user, host, dir string
	}{
		{"sam@example.com:/home/sam", "sam", "example.com", "/home/sam"},
		{"sam@example.com:home/sam", "sam", "example.com", "/home/sam"},
		{"root@192.168.1.1:/var/www", "root", "192.168.1.1", "/var/www"},
	}
	for _, tt := range tests {
		user, host, dir, err := ParseRemotePath(tt.input)
		if err != nil {
			t.Errorf("ParseRemotePath(%q) 意外錯誤: %v", tt.input, err)
			continue
		}
		if user != tt.user || host != tt.host || dir != tt.dir {
			t.Errorf("ParseRemotePath(%q) = (%q,%q,%q); want (%q,%q,%q)",
				tt.input, user, host, dir, tt.user, tt.host, tt.dir)
		}
	}
}

func TestParseRemotePath_Invalid(t *testing.T) {
	invalid := []string{
		"",
		"/local/path",
		"no-at-sign:/path",
		"user@host-without-colon",
		"userhost/path",
	}
	for _, in := range invalid {
		if _, _, _, err := ParseRemotePath(in); err == nil {
			t.Errorf("ParseRemotePath(%q) 應回傳錯誤", in)
		}
	}
}

func TestIsRemotePath(t *testing.T) {
	if !isRemotePath("sam@host:/path") {
		t.Error("標準遠端路徑應判為 true")
	}
	if isRemotePath("/local/path") {
		t.Error("本地路徑應判為 false")
	}
	if isRemotePath("nodomain") {
		t.Error("無 @: 的字串應判為 false")
	}
	// : 在 @ 之前不算遠端路徑
	if isRemotePath("C:/windows@file") {
		t.Error("冒號在 @ 之前不應判為遠端路徑")
	}
}

func TestSplitAtFirst(t *testing.T) {
	got := splitAtFirst("a@b@c", "@")
	if len(got) != 2 || got[0] != "a" || got[1] != "b@c" {
		t.Errorf("應只在第一個分隔符分割，got %v", got)
	}
	got = splitAtFirst("noseparator", "@")
	if len(got) != 1 || got[0] != "noseparator" {
		t.Errorf("無分隔符應回傳原字串，got %v", got)
	}
}
