package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// UpdateInfo 检查更新返回结构
type UpdateInfo struct {
	HasUpdate    bool   `json:"hasUpdate"`
	CurrentVer   string `json:"currentVersion"`
	LatestVer    string `json:"latestVersion"`
	ReleaseNotes string `json:"releaseNotes"`
	DownloadURL  string `json:"downloadUrl"`
}

// CheckUpdate 查询 GitHub Releases 最新版本
func CheckUpdate() (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", GitHubOwner, GitHubRepo)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API 返回 %d", resp.StatusCode)
	}

	var gh struct {
		TagName string `json:"tag_name"`
		Body    string `json:"body"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gh); err != nil {
		return nil, err
	}

	info := &UpdateInfo{
		CurrentVer:   Version,
		LatestVer:    gh.TagName,
		ReleaseNotes: gh.Body,
	}

	ext := ".exe"
	switch runtime.GOOS {
	case "darwin":
		ext = ".dmg"
	case "linux":
		ext = ".AppImage"
	}
	for _, a := range gh.Assets {
		if strings.HasSuffix(a.Name, ext) {
			info.DownloadURL = a.URL
			break
		}
	}

	if Version == "dev" {
		// dev 版本：只要远端有安装包就提示可下载
		info.HasUpdate = info.DownloadURL != ""
		return info, nil
	}

	if compareVersions(gh.TagName, Version) <= 0 {
		info.HasUpdate = false
		return info, nil
	}
	info.HasUpdate = true
	return info, nil
}

// compareVersions 比较 tag 与当前版本，支持 v 前缀；latest > current 返回 1
func compareVersions(latest, current string) int {
	normalize := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.TrimPrefix(s, "v")
		s = strings.TrimPrefix(s, "V")
		return s
	}
	latest, current = normalize(latest), normalize(current)
	if latest == current {
		return 0
	}
	lp := strings.Split(latest, ".")
	cp := strings.Split(current, ".")
	for i := 0; i < len(lp) || i < len(cp); i++ {
		var lv, cv int
		if i < len(lp) {
			fmt.Sscanf(lp[i], "%d", &lv)
		}
		if i < len(cp) {
			fmt.Sscanf(cp[i], "%d", &cv)
		}
		if lv != cv {
			if lv > cv {
				return 1
			}
			return -1
		}
	}
	return 0
}

// ApplyUpdate 直接打开浏览器下载页面，让用户手动下载安装
func ApplyUpdate(downloadURL string) error {
	if strings.TrimSpace(downloadURL) == "" {
		return fmt.Errorf("下载链接为空")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", downloadURL)
	case "windows":
		// start 的第一个空参数是窗口标题，避免 URL 被当成标题
		cmd = exec.Command("cmd", "/c", "start", "", downloadURL)
	default:
		cmd = exec.Command("xdg-open", downloadURL)
	}
	return cmd.Start()
}
