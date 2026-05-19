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

	if Version == "dev" {
		// dev 版本不参与版本比较
		info.HasUpdate = false
		return info, nil
	}

	if gh.TagName == Version || strings.TrimPrefix(gh.TagName, "v") == Version {
		info.HasUpdate = false
		return info, nil
	}

	info.HasUpdate = true

	// 简化：直接返回浏览器下载链接，不做自动更新
	// Windows -> .exe, macOS -> .dmg
	ext := ".exe"
	if runtime.GOOS == "darwin" {
		ext = ".dmg"
	}
	for _, a := range gh.Assets {
		if strings.HasSuffix(a.Name, ext) {
			info.DownloadURL = a.URL
			break
		}
	}

	return info, nil
}

// ApplyUpdate 直接打开浏览器下载页面，让用户手动下载安装
func ApplyUpdate(downloadURL string) error {
	// 打开浏览器下载
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", downloadURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", downloadURL)
	default:
		cmd = exec.Command("xdg-open", downloadURL)
	}
	return cmd.Start()
}
