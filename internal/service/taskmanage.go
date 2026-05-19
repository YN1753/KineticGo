package service

import (
	"context"
	"encoding/json"
	"errors"
	"kineticgo/internal/model"
	"kineticgo/internal/repository"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type TaskManageService struct {
	TaskRepo  *repository.TaskRepository
	registry  map[string]func() model.TaskInstance
	running   map[uint]model.RunningTask
	scheduler *Scheduler
	rootCtx   context.Context
	mutex     sync.RWMutex
}

// 初始化部分方法和函数
func NewTaskManageService(task *repository.TaskRepository) *TaskManageService {
	t := &TaskManageService{
		TaskRepo: task,
		registry: make(map[string]func() model.TaskInstance),
		running:  make(map[uint]model.RunningTask),
	}
	t.scheduler = NewScheduler(t.runScheduled)
	return t
}

// SetRootCtx 由 app 在 OnStartup 时注入，cron 触发的任务用这个 ctx，避免和窗口 ctx 绑死
func (t *TaskManageService) SetRootCtx(ctx context.Context) {
	t.rootCtx = ctx
}

// runScheduled 是注入给 Scheduler 的回调，到点拉起任务
func (t *TaskManageService) runScheduled(scheduleId uint) {
	if t.rootCtx == nil {
		return
	}
	ctx := WithTrigger(t.rootCtx, "schedule")
	if err := t.Start(ctx, scheduleId); err != nil {
		runtime.EventsEmit(t.rootCtx, "task_log", map[string]any{
			"scheduleId": scheduleId,
			"level":      LogError,
			"message":    "定时触发失败: " + err.Error(),
			"time":       time.Now().Unix(),
		})
	}
}

// LoadEnabledCronSchedules 启动时调用：把所有启用的定时任务塞进调度器并启动 cron
func (t *TaskManageService) LoadEnabledCronSchedules() {
	list, err := t.TaskRepo.GetEnabledCronSchedules()
	if err != nil {
		if t.rootCtx != nil {
			runtime.EventsEmit(t.rootCtx, "log", "加载定时任务失败: "+err.Error())
		}
		return
	}
	for i := range list {
		if err := t.scheduler.Add(&list[i]); err != nil && t.rootCtx != nil {
			runtime.EventsEmit(t.rootCtx, "log", "加载 schedule ["+list[i].Name+"] 失败: "+err.Error())
		}
	}
	t.scheduler.Start()
}

// Shutdown 由 app 在 OnShutdown 时调用，优雅停掉所有定时任务
func (t *TaskManageService) Shutdown() {
	t.scheduler.Stop()
}
func (t *TaskManageService) Register(taskType string, constructor func(repo *repository.TaskRepository) model.TaskInstance) {
	t.registry[taskType] = func() model.TaskInstance {
		return constructor(t.TaskRepo)
	}
}
func (t *TaskManageService) RegisterSystem(name string, constructor func(repo *repository.TaskRepository) model.TaskInstance) {
	key := "system-" + name
	t.registry[key] = func() model.TaskInstance {
		return constructor(t.TaskRepo)
	}
}
func (t *TaskManageService) AutoStartSystemTasks(ctx context.Context) {
	schedules, err := t.TaskRepo.GetEnabledSystemSchedules()
	if err != nil {
		runtime.EventsEmit(ctx, "log", "获取开机启动任务失败:")
		return
	}

	for _, schedule := range schedules {
		err := t.Start(ctx, schedule.ID)
		if err != nil {
			runtime.EventsEmit(ctx, "log", "自动启动系统任务 ["+schedule.Name+"] 失败:"+err.Error())
			return
		}
	}
}

//

func (t *TaskManageService) GetTaskList() (*[]model.Task, error) {
	return t.TaskRepo.GetTaskList()
}

func (t *TaskManageService) Start(ctx context.Context, scheduleId uint) error {
	t.mutex.Lock()
	if _, ok := t.running[scheduleId]; ok {
		t.mutex.Unlock()
		return errors.New("任务已经在运行")
	}
	t.mutex.Unlock()

	task, err := t.TaskRepo.GetTaskScheduleById(scheduleId)
	if err != nil {
		return errors.New("获取任务失败")
	}
	if task.TaskType == "system" {
		task.TaskType = "system-" + task.Name
	}
	factory, ok := t.registry[task.TaskType]
	if !ok {
		return errors.New("未找到匹配的任务处理器: " + task.TaskType)
	}
	instance := factory()
	childCtx, cancel := context.WithCancel(ctx)

	// 系统任务持续运行、日志噪声大，不写 execution / log 表
	isSystem := strings.HasPrefix(task.TaskType, "system-")

	var execId uint
	if !isSystem {
		exec := &model.TaskExecution{
			OptionID:    scheduleId,
			TriggerType: triggerFrom(ctx),
			Status:      "running",
			StartTime:   time.Now(),
		}
		_ = t.TaskRepo.CreateTaskExecution(exec)
		execId = exec.ID
	}
	childCtx = withTaskLog(childCtx, t.TaskRepo, scheduleId, execId, !isSystem)

	t.mutex.Lock()
	t.running[scheduleId] = model.RunningTask{
		Instance: instance,
		Cancel:   cancel,
	}
	t.mutex.Unlock()

	go func() {
		t.TaskRepo.AddActiveTask(&t.TaskRepo.ActiveTasks)
		runErr := instance.Run(childCtx, scheduleId)

		if !isSystem && execId > 0 {
			status := "success"
			summary := "执行成功"
			if runErr != nil {
				status = "failed"
				summary = runErr.Error()
				TaskLog(childCtx, LogError, runErr.Error())
			}
			_ = t.TaskRepo.UpdateTaskExecution(execId, status, summary, time.Now())
		} else if runErr != nil {
			runtime.EventsEmit(childCtx, "log", runErr.Error())
		}

		// 定时任务跑完后刷新下次运行时间, 让前端的 "下次运行" 不再卡在首次计算结果
		if task.CronExpr != "" {
			if sched, parseErr := cron.ParseStandard(task.CronExpr); parseErr == nil {
				_ = t.TaskRepo.UpdateScheduleNextRunTime(scheduleId, sched.Next(time.Now()))
			}
		}

		t.mutex.Lock()
		delete(t.running, scheduleId)
		t.mutex.Unlock()
		t.TaskRepo.RemoveActiveTask(&t.TaskRepo.ActiveTasks)
	}()

	return nil
}

func (t *TaskManageService) Stop(scheduleId uint) error {
	t.mutex.Lock()
	task, exists := t.running[scheduleId]
	t.mutex.Unlock()
	if !exists {
		return errors.New("任务未运行，无法停止")
	}
	if task.Cancel != nil {
		task.Cancel()
	}
	if task.Instance != nil {
		err := task.Instance.Stop(scheduleId)
		if err != nil {
			runtime.EventsEmit(context.Background(), "log", err.Error())
			return err
		}
	}
	return nil
}

func (t *TaskManageService) GetRunningTaskIds() []uint {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	ids := make([]uint, 0, len(t.running))
	for id := range t.running {
		ids = append(ids, id)
	}
	return ids
}

func (t *TaskManageService) GetTaskConfigById(id uint) (json.RawMessage, error) {
	return t.TaskRepo.GetTaskConfigById(id)
}

func (t *TaskManageService) GetTaskScheduleList() ([]model.TaskSchedule, error) {
	list, err := t.TaskRepo.GetTaskScheduleList()
	if err != nil {
		return nil, err
	}
	return *list, nil
}

func prepareCron(sch *model.TaskSchedule) error { //检验cron合法性并且算出下次时间
	if sch.CronExpr == "" {
		sch.NextRunTime = time.Time{}
		return nil
	}
	schedule, err := cron.ParseStandard(sch.CronExpr)
	if err != nil {
		return errors.New("cron 表达式无效: " + err.Error())
	}
	sch.NextRunTime = schedule.Next(time.Now())
	return nil
}

func (t *TaskManageService) syncScheduler(sch *model.TaskSchedule) error { //存入cron里面
	t.scheduler.Remove(sch.ID) // 不存在是 no-op，安全
	if sch.IsEnabled && sch.CronExpr != "" {
		return t.scheduler.Add(sch)
	}
	return nil
}

func (t *TaskManageService) CreateTaskSchedule(sch *model.TaskSchedule) error {
	if err := prepareCron(sch); err != nil {
		return err
	}
	if err := t.TaskRepo.CreateTaskSchedule(sch); err != nil {
		return err
	}
	return t.syncScheduler(sch)
}

func (t *TaskManageService) UpdateTaskSchedule(sch *model.TaskSchedule) error {
	if err := prepareCron(sch); err != nil {
		return err
	}
	if err := t.TaskRepo.UpdateTaskSchedule(sch); err != nil {
		return err
	}
	return t.syncScheduler(sch)
}

func (t *TaskManageService) DeleteTaskSchedule(id uint) error {
	t.scheduler.Remove(id)
	return t.TaskRepo.DeleteTaskSchedule(id)
}

func (t *TaskManageService) GetTaskScheduleById(id uint) (*model.TaskSchedule, error) {
	return t.TaskRepo.GetTaskScheduleById(id)
}

func (t *TaskManageService) GetTaskExecutions(limit int) ([]model.TaskExecution, error) {
	return t.TaskRepo.GetTaskExecutions(limit)
}

func (t *TaskManageService) GetTaskLogsByExecution(execId uint) ([]model.TaskLog, error) {
	return t.TaskRepo.GetTaskLogsByExecution(execId)
}

func (t *TaskManageService) GetSystemTaskScheduleList() ([]model.TaskSchedule, error) {
	list, err := t.TaskRepo.GetSystemTaskScheduleList()
	if err != nil {
		return nil, err
	}
	return *list, nil
}

func (t *TaskManageService) GetTaskRepo() *repository.TaskRepository {
	return t.TaskRepo
}
