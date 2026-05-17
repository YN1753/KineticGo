package model

import (
	"context"
	"time"
)

type TaskInstance interface {
	Run(ctx context.Context, scheduleId uint) error
	Stop(scheduleId uint) error
}
type Task struct { //任务类型
	ID          uint         `gorm:"primaryKey;column:id"`
	Name        string       `gorm:"column:name;not null;unique"`
	Type        string       `gorm:"column:type;index"`
	Description string       `gorm:"column:description;not null"`
	ExecMode    string       `gorm:"column:exec_mode;not null"`
	Config      TempleConfig `gorm:"column:config"`
	CreatedAt   time.Time    `gorm:"column:created_at"`
}

type TaskSchedule struct {
	ID           uint      `gorm:"primaryKey;column:id"`
	Name         string    `gorm:"column:name;not null"`
	TaskType     string    `gorm:"column:task_type;size:50;index"`
	CronExpr     string    `gorm:"column:cron_expr"`
	IntervalSecs int       `gorm:"column:interval_secs"`
	IsEnabled    bool      `gorm:"column:is_enabled;default:true"`
	NextRunTime  time.Time `gorm:"column:next_run_time;index"`

	Config TempleConfig `gorm:"column:config;type:text"`
}

type TaskExecution struct { //任务历史记录
	ID            uint      `gorm:"primaryKey;column:id"`
	OptionID      uint      `gorm:"column:option_id;index"`
	TriggerType   string    `gorm:"column:trigger_type"`
	Status        string    `gorm:"column:status"`
	ResultSummary string    `gorm:"column:result_summary;type:text"`
	StartTime     time.Time `gorm:"column:start_time"`
	EndTime       time.Time `gorm:"column:end_time"`
}
type TaskLog struct { //任务日志表
	ID          uint      `gorm:"primaryKey;column:id"`
	ExecutionID uint      `gorm:"column:execution_id;index"`
	Level       string    `gorm:"column:level"`
	Message     string    `gorm:"column:message;type:text"`
	CreatedAt   time.Time `gorm:"column:created_at;index"`
}

type RunningTask struct {
	Instance TaskInstance
	Cancel   context.CancelFunc
}
