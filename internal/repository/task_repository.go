package repository

import (
	"encoding/json"
	"errors"
	"kineticgo/internal/model"
	"sync/atomic"

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
	err := t.Db.Model(&model.Task{}).Find(&tasks).Error
	if err != nil {
		return nil, errors.New("获取task列表失败")
	}
	return &tasks, nil
}

func (t *TaskRepository) GetTaskScheduleList() (*[]model.TaskSchedule, error) {
	var taskSchedules []model.TaskSchedule
	err := t.Db.Model(&model.TaskSchedule{}).Find(&taskSchedules).Error
	if err != nil {
		return nil, errors.New("获取taskSchedule列表失败")
	}
	return &taskSchedules, nil
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
func (t *TaskRepository) DeleteTaskSchedule(id uint) error {
	return t.Db.Delete(&model.TaskSchedule{}, id).Error
}
func (t *TaskRepository) GetTaskScheduleById(id uint) (*model.TaskSchedule, error) {
	var task model.TaskSchedule
	err := t.Db.Model(&model.TaskSchedule{}).Where("id = ? ", id).Find(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (t *TaskRepository) GetEnabledSystemSchedules() ([]model.TaskSchedule, error) {
	var schedules []model.TaskSchedule

	err := t.Db.Where("task_type LIKE ? AND is_enabled = ?", "system-%", true).Find(&schedules).Error
	if err != nil {
		return nil, err
	}
	return schedules, nil
}
