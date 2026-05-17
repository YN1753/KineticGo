package service

import (
	"context"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"kineticgo/internal/model"
	"kineticgo/internal/repository"
	"time"
)

// cpu部分
type CpuService struct {
	TaskRepo   *repository.TaskRepository
	CpuPercent float64
}

func NewCpuService(TaskRepo *repository.TaskRepository) model.TaskInstance {
	return CpuService{
		CpuPercent: 0,
		TaskRepo:   TaskRepo,
	}
}
func (c CpuService) Run(ctx context.Context, scheduleId uint) error {
	ticker := time.NewTicker(1 * time.Second)

	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			return nil

		case <-ticker.C:
			percent, _ := cpu.Percent(0, false)
			runtime.EventsEmit(ctx, "stats_update", map[string]interface{}{
				"cpuPercent": percent[0],
			})
		}
	}
}

func (c CpuService) Stop(scheduleId uint) error {
	return nil
}

// 内存部分
type MemoryService struct {
	TaskRepo      *repository.TaskRepository
	MemeryPercent float64
}

func NewMemeryService(TaskRepo *repository.TaskRepository) model.TaskInstance {
	return MemoryService{
		MemeryPercent: 0,
		TaskRepo:      TaskRepo,
	}
}

func (m MemoryService) Run(ctx context.Context, scheduleId uint) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			info, _ := mem.VirtualMemory()
			runtime.EventsEmit(ctx, "stats_update", map[string]interface{}{
				"memPercent": info.UsedPercent,
			})
		}
	}

}
func (m MemoryService) Stop(scheduleId uint) error {

	return nil
}

// 活跃任务
type ActiveTasksService struct {
	TaskRepo *repository.TaskRepository
}

func NewActiveTasksService(TaskRepo *repository.TaskRepository) model.TaskInstance {
	return ActiveTasksService{
		TaskRepo: TaskRepo,
	}
}
func (a ActiveTasksService) Run(ctx context.Context, scheduleId uint) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			runtime.EventsEmit(ctx, "stats_update", map[string]interface{}{
				"activeTasks": a.TaskRepo.ActiveTasks,
			})
		}
	}
}
func (a ActiveTasksService) Stop(scheduleId uint) error {

	return nil
}

//本地网络

type NetworkService struct {
	TaskRepo *repository.TaskRepository
}

func NewNetworkService(TaskRepo *repository.TaskRepository) model.TaskInstance {
	return NetworkService{
		TaskRepo: TaskRepo,
	}
}

func (n NetworkService) Run(ctx context.Context, scheduleId uint) error {
	ticker := time.NewTicker(1 * time.Second)
	var lastRecv uint64 = 0
	var lastSent uint64 = 0
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			currentStats, err := net.IOCounters(false)
			if err != nil || len(currentStats) == 0 {
				continue
			}
			if lastRecv == 0 && lastSent == 0 {
				lastRecv = currentStats[0].BytesRecv
				lastSent = currentStats[0].BytesSent
				continue
			}
			currentRecv := currentStats[0].BytesRecv
			currentSent := currentStats[0].BytesSent

			var downloadSpeed, uploadSpeed uint64
			if currentRecv >= lastRecv {
				downloadSpeed = currentRecv - lastRecv
			}
			if currentSent >= lastSent {
				uploadSpeed = currentSent - lastSent
			}
			lastRecv = currentRecv
			lastSent = currentSent

			runtime.EventsEmit(ctx, "stats_update", map[string]interface{}{
				"downloadSpeed": downloadSpeed,
				"uploadSpeed":   uploadSpeed,
			})
		}
	}
}
func (n NetworkService) Stop(scheduleId uint) error {
	return nil
}
