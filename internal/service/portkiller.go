package service

import (
	"context"
	"fmt"
	"kineticgo/internal/model"
	"kineticgo/internal/repository"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// PortKiller 定义占用端口的进程元数据
type PortKiller struct {
	Port        int    `json:"port"`
	ProcessName string `json:"name"`
	Pid         int    `json:"pid"`
	Path        string `json:"path"`
	IsCritical  bool   `json:"isCritical"`
}

type PortKillerService struct {
	Port     PortKiller
	TaskRepo repository.TaskRepository
}

func NewPortKillerService(repo *repository.TaskRepository) model.TaskInstance {
	return &PortKillerService{
		TaskRepo: *repo,
	}
}

func (p *PortKillerService) Run(ctx context.Context, pid uint) error {
	TaskLog(ctx, LogInfo, fmt.Sprintf("开始尝试注销端口占用进程 (PID: %d)...", pid))
	err := KillPortByPid(pid)
	if err != nil {
		return err
	}
	TaskLog(ctx, LogSuccess, fmt.Sprintf("进程 (PID: %d) 已成功强制终结，端口已释放。", pid))
	return nil
}

func (p *PortKillerService) Stop(scheduleId uint) error {
	return nil
}
func GetPortMessage(ctx context.Context) ([]PortKiller, error) {
	sysType := runtime.GOOS
	var cmd *exec.Cmd

	if sysType == "windows" {
		cmd = exec.Command("cmd", "/C", "netstat -ano -p tcp")
	} else {
		cmd = exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-P", "-n", "-F", "pcn")
	}

	var sysProcAttr *syscall.SysProcAttr
	if sysType == "windows" {
		sysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		cmd.SysProcAttr = sysProcAttr
	}

	out, err := cmd.Output()
	if err != nil {
		return []PortKiller{}, nil
	}

	// 性能优化：预先获取所有进程映射，避免循环内调用外部指令
	procMap := make(map[int]string)
	if sysType == "windows" {
		procMap = getWindowsProcessMap(sysProcAttr)
	}

	lines := strings.Split(string(out), "\n")
	seen := make(map[int]bool)
	var list []PortKiller

	if sysType == "windows" {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "TCP") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 5 {
				continue
			}
			if fields[3] != "LISTENING" {
				continue
			}
			localAddr := fields[1]
			pidStr := fields[4]
			portIdx := strings.LastIndex(localAddr, ":")
			if portIdx == -1 {
				continue
			}
			port, _ := strconv.Atoi(localAddr[portIdx+1:])
			if seen[port] || port == 0 {
				continue
			}
			seen[port] = true
			pid, _ := strconv.Atoi(pidStr)

			name := procMap[pid]
			if name == "" {
				name = "未知进程"
			}

			list = append(list, PortKiller{
				Port:        port,
				ProcessName: name,
				Pid:         pid,
				Path:        "", // 移除慢速的 wmic 路径查询
				IsCritical:  checkIsCritical(port, pid, name),
			})
		}
	} else {
		// macOS 解析逻辑保持不变，但增加安全性判断
		var currentPort PortKiller
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) < 2 {
				continue
			}
			tag := line[0]
			val := line[1:]
			switch tag {
			case 'p':
				pid, _ := strconv.Atoi(val)
				currentPort.Pid = pid
			case 'c':
				currentPort.ProcessName = val
			case 'n':
				idx := strings.LastIndex(val, ":")
				if idx != -1 {
					portStr := val[idx+1:]
					port, _ := strconv.Atoi(portStr)
					if port > 0 && !seen[port] {
						currentPort.Port = port
						currentPort.IsCritical = checkIsCritical(port, currentPort.Pid, currentPort.ProcessName)
						list = append(list, currentPort)
						seen[port] = true
					}
				}
				currentPort = PortKiller{}
			}
		}
	}

	return list, nil
}

// 批量获取进程名，极大提升性能
func getWindowsProcessMap(attr *syscall.SysProcAttr) map[int]string {
	m := make(map[int]string)
	cmd := exec.Command("tasklist", "/nh", "/fo", "csv")
	cmd.SysProcAttr = attr
	out, err := cmd.Output()
	if err != nil {
		return m
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 2 {
			name := strings.Trim(fields[0], "\"")
			pidStr := strings.Trim(fields[1], "\"")
			pid, _ := strconv.Atoi(pidStr)
			m[pid] = name
		}
	}
	return m
}

func checkIsCritical(port int, pid int, name string) bool {
	// 1. 系统核心 PID 保护
	if pid <= 4 || pid == 0 {
		return true
	}
	// 2. 众所周知的核心服务名保护
	lowerName := strings.ToLower(name)
	systemProcs := []string{"system", "svchost.exe", "lsass.exe", "services.exe", "wininit.exe", "csrss.exe", "smss.exe", "launchd", "kernel_task"}
	for _, sp := range systemProcs {
		if lowerName == sp {
			return true
		}
	}
	// 3. 核心端口保护
	criticalPorts := []int{22, 80, 443, 135, 445, 3389, 53, 123, 161, 162}
	for _, cp := range criticalPorts {
		if port == cp {
			return true
		}
	}
	// 4. 常见的数据库/基础服务警告（可选，这里设为 false 仅作提醒）
	return false
}

func KillPortByPid(pid uint) error {
	if pid <= 4 {
		return fmt.Errorf("系统核心进程 (PID: %d) 受内核安全保护，拒绝强杀", pid)
	}

	sysType := runtime.GOOS
	var cmd *exec.Cmd

	if sysType == "windows" {
		cmd = exec.Command("taskkill", "/F", "/PID", strconv.Itoa(int(pid)))
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	} else {
		cmd = exec.Command("kill", "-9", strconv.Itoa(int(pid)))
	}

	return cmd.Run()
}

func getWindowsProcessInfo(pid int, attr *syscall.SysProcAttr) (string, string) {
	nameCmd := exec.Command("tasklist", "/fi", fmt.Sprintf("pid eq %d", pid), "/nh", "/fo", "csv")
	nameCmd.SysProcAttr = attr
	out, err := nameCmd.Output()
	procName := "未知进程"
	if err == nil && len(out) > 0 {
		fields := strings.Split(string(out), ",")
		if len(fields) > 0 {
			procName = strings.Trim(fields[0], "\"")
		}
	}

	pathCmd := exec.Command("wmic", "process", "where", fmt.Sprintf("processid=%d", pid), "get", "ExecutablePath", "/value")
	pathCmd.SysProcAttr = attr
	pathOut, pathErr := pathCmd.Output()
	procPath := "未知 / 系统保护"
	if pathErr == nil && len(pathOut) > 0 {
		lines := strings.Split(string(pathOut), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "ExecutablePath=") {
				procPath = strings.TrimSpace(strings.Split(line, "=")[1])
				break
			}
		}
	}

	return procName, procPath
}

func getUnixProcessPath(pid int) string {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=")
	out, err := cmd.Output()
	if err == nil && len(out) > 0 {
		return strings.TrimSpace(string(out))
	}
	return "未知"
}

func checkIsCriticalPort(port int) bool {
	criticalPorts := []int{22, 80, 443, 135, 445, 3389}
	for _, cp := range criticalPorts {
		if port == cp {
			return true
		}
	}
	return false
}
