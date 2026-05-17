package service

import (
	"context"
	"encoding/json"
	"errors"
	"kineticgo/internal/model"
	"kineticgo/internal/repository"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type TaskManageService struct {
	TaskRepo *repository.TaskRepository
	registry map[string]func() model.TaskInstance
	running  map[uint]model.RunningTask
	mutex    sync.RWMutex
}

// 初始化部分方法和函数
func NewTaskManageService(task *repository.TaskRepository) *TaskManageService {
	return &TaskManageService{
		TaskRepo: task,
		registry: make(map[string]func() model.TaskInstance),
		running:  make(map[uint]model.RunningTask),
	}
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

	t.mutex.Lock()
	t.running[scheduleId] = model.RunningTask{
		Instance: instance,
		Cancel:   cancel,
	}
	t.mutex.Unlock()

	go func() {
		t.TaskRepo.AddActiveTask(&t.TaskRepo.ActiveTasks)
		err := instance.Run(childCtx, scheduleId)
		if err != nil {
			runtime.EventsEmit(childCtx, "log", err.Error())
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

func (t *TaskManageService) GetTaskScheduleList() (*[]model.TaskSchedule, error) {
	return t.TaskRepo.GetTaskScheduleList()
}

func (t *TaskManageService) CreateTaskSchedule(task *model.TaskSchedule) error {
	return t.TaskRepo.CreateTaskSchedule(task)
}

func (t *TaskManageService) UpdateTaskSchedule(task *model.TaskSchedule) error {
	return t.TaskRepo.UpdateTaskSchedule(task)
}

func (t *TaskManageService) DeleteTaskSchedule(id uint) error {
	return t.TaskRepo.DeleteTaskSchedule(id)
}

func (t *TaskManageService) GetTaskScheduleById(id uint) (*model.TaskSchedule, error) {
	return t.TaskRepo.GetTaskScheduleById(id)
}
