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
	ID          uint         `gorm:"primaryKey;column:id" json:"ID"`
	Name        string       `gorm:"column:name;not null;unique" json:"Name"`
	Type        string       `gorm:"column:type;index" json:"Type"`
	Description string       `gorm:"column:description;not null" json:"Description"`
	ExecMode    string       `gorm:"column:exec_mode;not null" json:"ExecMode"`
	Config      TempleConfig `gorm:"column:config" json:"Config"`
	CreatedAt   time.Time    `gorm:"column:created_at" json:"CreatedAt"`
}

type TaskSchedule struct {
	ID           uint      `gorm:"primaryKey;column:id" json:"ID"`
	Name         string    `gorm:"column:name;not null" json:"Name"`
	TaskType     string    `gorm:"column:task_type;size:50;index" json:"TaskType"`
	CronExpr     string    `gorm:"column:cron_expr" json:"CronExpr"`
	IntervalSecs int       `gorm:"column:interval_secs" json:"IntervalSecs"`
	IsEnabled    bool      `gorm:"column:is_enabled;default:true" json:"IsEnabled"`
	NextRunTime  time.Time `gorm:"column:next_run_time;index" json:"NextRunTime"`
	Option       string    `gorm:"column:option;size:255" json:"Option"` // 备注字段

	Config TempleConfig `gorm:"column:config;type:text" json:"Config"`
}

type TaskExecution struct { //任务历史记录
	ID            uint      `gorm:"primaryKey;column:id" json:"ID"`
	OptionID      uint      `gorm:"column:option_id;index" json:"OptionID"`
	TriggerType   string    `gorm:"column:trigger_type" json:"TriggerType"`
	Status        string    `gorm:"column:status" json:"Status"`
	ResultSummary string    `gorm:"column:result_summary;type:text" json:"ResultSummary"`
	StartTime     time.Time `gorm:"column:start_time" json:"StartTime"`
	EndTime       time.Time `gorm:"column:end_time" json:"EndTime"`
}
type TaskLog struct { //任务日志表
	ID          uint      `gorm:"primaryKey;column:id" json:"ID"`
	ExecutionID uint      `gorm:"column:execution_id;index" json:"ExecutionID"`
	Level       string    `gorm:"column:level" json:"Level"`
	Message     string    `gorm:"column:message;type:text" json:"Message"`
	CreatedAt   time.Time `gorm:"column:created_at;index" json:"CreatedAt"`
}

type RunningTask struct {
	Instance TaskInstance
	Cancel   context.CancelFunc
}
