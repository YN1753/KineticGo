package service

import (
	"context"
	"encoding/json"
	"fmt"
	"kineticgo/internal/model"
	"kineticgo/internal/repository"
	"os/exec"
	"runtime"
	"strings"
)

// AppLauncherService 本地应用快捷启动器服务
type AppLauncherService struct {
	TaskRepo *repository.TaskRepository
}

// AppLauncherConfig 启动器配置结构
type AppLauncherConfig struct {
	Paths string `json:"paths"` // 换行符分隔的路径或 URL
}

// NewAppLauncherService 构造函数
func NewAppLauncherService(repo *repository.TaskRepository) model.TaskInstance {
	return &AppLauncherService{
		TaskRepo: repo,
	}
}

// Run 执行启动任务
func (s *AppLauncherService) Run(ctx context.Context, scheduleId uint) error {
	// 强制关闭持久化日志到数据库，仅通过 Wails Events 推送前端实时显示
	ctx = context.WithValue(ctx, keyPersist, false)

	// 从数据库获取当前任务调度的配置
	schedule, err := s.TaskRepo.GetTaskScheduleById(scheduleId)
	if err != nil {
		TaskLog(ctx, LogError, "获取任务配置失败: "+err.Error())
		return err
	}

	var config AppLauncherConfig
	if err := json.Unmarshal([]byte(schedule.Config), &config); err != nil {
		TaskLog(ctx, LogError, "解析配置 JSON 失败: "+err.Error())
		return err
	}

	// 解析路径列表（按换行符分割）
	paths := strings.Split(config.Paths, "\n")
	launchedCount := 0

	for _, p := range paths {
		target := strings.TrimSpace(p)
		if target == "" {
			continue
		}

		// 异步非阻塞拉起每一个目标，不干扰主流程
		go s.launch(ctx, scheduleId, target)
		launchedCount++
	}

	if launchedCount == 0 {
		TaskLog(ctx, LogWarn, "未发现有效的启动路径，请检查任务配置")
	}

	return nil
}

// launch 跨平台异步拉起核心实现
func (s *AppLauncherService) launch(ctx context.Context, scheduleId uint, target string) {
	var cmd *exec.Cmd
	sysType := runtime.GOOS

	switch sysType {
	case "windows":
		// Windows: 使用 cmd /C start "" "target" 来拉起，兼容 exe/bat/url
		// 注意：start 命令的第一个空双引号是用于窗口标题占位，必不可少
		cmd = exec.Command("cmd", "/C", "start", "", target)
		// 使用通用的 getSysProcAttr() 来隐藏后台 cmd 黑窗口
		cmd.SysProcAttr = getSysProcAttr()
	case "darwin":
		// macOS: 使用标准的 open 命令
		cmd = exec.Command("open", target)
	default:
		// Linux: 使用标准的 xdg-open 命令
		cmd = exec.Command("xdg-open", target)
	}

	// 核心要求：必须使用 Start() 异步非阻塞运行，绝不能使用 Run() 或 Output()
	err := cmd.Start()
	if err != nil {
		// 失败日志实时推送到前端
		TaskLog(ctx, LogError, fmt.Sprintf("拉起失败 [%s]: %v", target, err))
	} else {
		// 成功日志实时推送到前端
		TaskLog(ctx, LogSuccess, fmt.Sprintf("已成功异步拉起目标: %s", target))
	}
}

// Stop 停止任务
func (s *AppLauncherService) Stop(scheduleId uint) error {
	// 启动器任务属于即时触发型，进程一旦拉起即脱离父进程管理，无需 Stop 逻辑
	return nil
}
