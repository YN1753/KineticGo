package wailsapp

import (
	"context"
	"encoding/json"
	"time"

	"kineticgo/internal/model"
	"kineticgo/internal/service"
)

type App struct {
	ctx        context.Context
	taskManage *service.TaskManageService
}

func NewApp(taskManage *service.TaskManageService) *App {
	return &App{
		taskManage: taskManage,
	}
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	a.taskManage.SetRootCtx(ctx)
	a.taskManage.RegisterSystem("local_cpu", service.NewCpuService)
	a.taskManage.RegisterSystem("local_memory", service.NewMemeryService)
	a.taskManage.RegisterSystem("active_tasks", service.NewActiveTasksService)
	a.taskManage.RegisterSystem("local_network", service.NewNetworkService)
	a.taskManage.Register("campus_auth", service.NewSuseWifiService)
	a.taskManage.LoadEnabledCronSchedules()
	go func() {
		time.Sleep(1 * time.Second)
		a.taskManage.AutoStartSystemTasks(a.ctx)
	}()
}

func (a *App) OnShutdown(ctx context.Context) {
	a.taskManage.Shutdown()
}

func (a *App) GetTaskList() ([]model.Task, error) {
	tasks, err := a.taskManage.GetTaskList()
	if err != nil {
		return nil, err
	}
	return *tasks, nil
}

func (a *App) GetTaskConfigById(id uint) (json.RawMessage, error) {
	return a.taskManage.GetTaskConfigById(id)
}

func (a *App) GetTaskScheduleList() ([]model.TaskSchedule, error) {
	list, err := a.taskManage.GetTaskScheduleList()
	if err != nil {
		return nil, err
	}
	return *list, nil
}

func (a *App) CreateTaskSchedule(task model.TaskSchedule) error {
	return a.taskManage.CreateTaskSchedule(&task)
}

func (a *App) UpdateTaskSchedule(task model.TaskSchedule) error {
	return a.taskManage.UpdateTaskSchedule(&task)
}

func (a *App) DeleteTaskSchedule(id uint) error {
	return a.taskManage.DeleteTaskSchedule(id)
}

func (a *App) GetTaskScheduleById(id uint) (*model.TaskSchedule, error) {
	return a.taskManage.GetTaskScheduleById(id)
}

func (a *App) RunTask(scheduleID uint) error {
	return a.taskManage.Start(a.ctx, scheduleID)
}

func (a *App) StopTask(scheduleID uint) error {
	return a.taskManage.Stop(scheduleID)
}

func (a *App) GetRunningTaskIds() []uint {
	return a.taskManage.GetRunningTaskIds()
}

func (a *App) GetTaskExecutions(limit int) ([]model.TaskExecution, error) {
	return a.taskManage.GetTaskExecutions(limit)
}

func (a *App) GetTaskLogsByExecution(execId uint) ([]model.TaskLog, error) {
	return a.taskManage.GetTaskLogsByExecution(execId)
}

func (a *App) GetVersion() string {
	return service.Version
}

func (a *App) CheckUpdate() (*service.UpdateInfo, error) {
	return service.CheckUpdate()
}

func (a *App) ApplyUpdate(downloadURL string) error {
	return service.ApplyUpdate(downloadURL)
}
