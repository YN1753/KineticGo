package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

	if gh.TagName == Version {
		info.HasUpdate = false
		return info, nil
	}

	info.HasUpdate = true

	ext := "exe"
	if runtime.GOOS == "darwin" {
		ext = "zip"
	}
	for _, a := range gh.Assets {
		if filepath.Ext(a.Name) == "."+ext {
			info.DownloadURL = a.URL
			break
		}
	}

	return info, nil
}

// ApplyUpdate 执行更新
// Windows：下载新 exe → 写批处理脚本 → 启动脚本 → 退出自身
// macOS：返回错误提示手动下载
func ApplyUpdate(downloadURL string) error {
	if runtime.GOOS == "darwin" {
		return fmt.Errorf("macOS 暂不支持自动更新，请前往 Release 页面手动下载：%s", downloadURL)
	}

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tmpExe := filepath.Join(os.TempDir(), "KineticGo_update.exe")
	f, err := os.Create(tmpExe)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		return err
	}

	currentExe, err := os.Executable()
	if err != nil {
		return err
	}

	batPath := filepath.Join(os.TempDir(), "kineticgo_update.bat")
	batContent := fmt.Sprintf(`@echo off
timeout /t 2 /nobreak >nul
move /Y "%s" "%s"
start "" "%s"
del "%%~f0"
`, tmpExe, currentExe, currentExe)

	if err := os.WriteFile(batPath, []byte(batContent), 0644); err != nil {
		return err
	}

	cmd := exec.Command("cmd", "/c", batPath)
	if err := cmd.Start(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
