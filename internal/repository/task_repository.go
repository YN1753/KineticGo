package repository

import (
	"encoding/json"
	"errors"
	"kineticgo/internal/model"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

type TaskRepository struct {
	Db          *gorm.DB
	ActiveTasks int64
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{
		Db:          db,
		ActiveTasks: 0,
	}
}

func (t *TaskRepository) AddActiveTask(activeTasks *int64) {
	atomic.AddInt64(activeTasks, 1)
}
func (t *TaskRepository) RemoveActiveTask(activeTasks *int64) {
	atomic.AddInt64(activeTasks, -1)
}

func (t *TaskRepository) GetTaskList() (*[]model.Task, error) {
	var tasks []model.Task
	err := t.Db.Model(&model.Task{}).Where("NOT type=?", "system").Find(&tasks).Error
	if err != nil {
		return nil, errors.New("获取task列表失败")
	}
	return &tasks, nil
}

func (t *TaskRepository) GetTaskScheduleList() (*[]model.TaskSchedule, error) {
	var taskSchedules []model.TaskSchedule
	err := t.Db.Model(&model.TaskSchedule{}).Where("task_type NOT LIKE ?", "system-%").Find(&taskSchedules).Error
	if err != nil {
		return nil, errors.New("获取taskSchedule列表失败")
	}
	return &taskSchedules, nil
}
func (t *TaskRepository) GetSystemTaskScheduleList() (*[]model.TaskSchedule, error) {
	var taskSchedules []model.TaskSchedule
	err := t.Db.Model(&model.TaskSchedule{}).Where("task_type LIKE ?", "system-%").Find(&taskSchedules).Error
	if err != nil {
		return nil, errors.New("获取taskSchedule列表失败")
	}
	return &taskSchedules, nil
}
func (t *TaskRepository) ChangeSystemTaskStatus(enable bool, scheduleId uint) error {
	return t.Db.Model(&model.TaskSchedule{}).Where("id = ?", scheduleId).Update("is_enabled", enable).Error
}

func (t *TaskRepository) GetTaskConfigById(id uint) (json.RawMessage, error) {
	var task model.Task
	err := t.Db.Model(&model.Task{}).Where("id = ? ", id).Find(&task).Error
	if err != nil {
		return nil, errors.New("获取task失败")
	}
	return task.Config.ToJson(), nil
}

func (t *TaskRepository) CreateTaskSchedule(task *model.TaskSchedule) error {
	return t.Db.Create(task).Error
}
func (t *TaskRepository) UpdateTaskSchedule(task *model.TaskSchedule) error {
	return t.Db.Save(task).Error
}

// UpdateScheduleNextRunTime 只刷新 next_run_time 字段, 避免和业务侧 Config 更新冲突.
func (t *TaskRepository) UpdateScheduleNextRunTime(id uint, nextTime time.Time) error {
	return t.Db.Model(&model.TaskSchedule{}).Where("id = ?", id).
		Update("next_run_time", nextTime).Error
}
func (t *TaskRepository) DeleteTaskSchedule(id uint) error {
	return t.Db.Delete(&model.TaskSchedule{}, id).Error
}
func (t *TaskRepository) GetTaskScheduleById(id uint) (*model.TaskSchedule, error) {
	var task model.TaskSchedule
	err := t.Db.Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// MarkInterruptedExecutions 启动时把上次异常退出仍停留在 running 的执行记录标成 interrupted.
func (t *TaskRepository) MarkInterruptedExecutions() error {
	return t.Db.Model(&model.TaskExecution{}).
		Where("status = ?", "running").
		Updates(map[string]any{
			"status":         "interrupted",
			"result_summary": "应用退出时任务仍在运行，已标记为中断",
			"end_time":       time.Now(),
		}).Error
}

// HasSuccessSince 判断该调度在 since 之后是否已有成功执行（用于 catch-up 去重）.
func (t *TaskRepository) HasSuccessSince(scheduleId uint, since time.Time) (bool, error) {
	var count int64
	err := t.Db.Model(&model.TaskExecution{}).
		Where("option_id = ? AND status = ? AND start_time >= ?", scheduleId, "success", since).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (t *TaskRepository) GetEnabledSystemSchedules() ([]model.TaskSchedule, error) {
	var schedules []model.TaskSchedule

	err := t.Db.Where("task_type LIKE ? AND is_enabled = ?", "system-%", true).Find(&schedules).Error
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

func (t *TaskRepository) GetEnabledCronSchedules() ([]model.TaskSchedule, error) {
	var schedules []model.TaskSchedule
	err := t.Db.Where("cron_expr <> '' AND is_enabled = ?", true).Find(&schedules).Error
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

func (t *TaskRepository) CreateTaskLog(log *model.TaskLog) error {
	return t.Db.Create(log).Error
}

func (t *TaskRepository) CreateTaskExecution(task *model.TaskExecution) error {
	return t.Db.Create(task).Error
}

func (t *TaskRepository) UpdateTaskExecution(id uint, status, summary string, endTime time.Time) error {
	return t.Db.Model(&model.TaskExecution{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":         status,
			"result_summary": summary,
			"end_time":       endTime,
		}).Error
}

func (t *TaskRepository) GetTaskExecutions(limit int) ([]model.TaskExecution, error) {
	var list []model.TaskExecution
	err := t.Db.Order("id DESC").Limit(limit).Find(&list).Error
	return list, err
}

func (t *TaskRepository) GetTaskLogsByExecution(execId uint) ([]model.TaskLog, error) {
	var list []model.TaskLog
	err := t.Db.Where("execution_id = ?", execId).Order("id ASC").Find(&list).Error
	return list, err
}
