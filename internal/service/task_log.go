package service

import (
	"context"
	"time"

	"kineticgo/internal/model"
	"kineticgo/internal/repository"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// 日志级别
const (
	LogInfo  = "info"
	LogWarn  = "warn"
	LogError = "error"
)

// ctx key（避免和外部 string key 冲突）
type taskCtxKey string

const (
	keyScheduleId taskCtxKey = "schedule_id"
	keyExecId     taskCtxKey = "exec_id"
	keyRepo       taskCtxKey = "task_repo"
	keyPersist    taskCtxKey = "persist_log"
	keyTrigger    taskCtxKey = "trigger_type"
)

// WithTrigger 标记本次 Start 是手动还是定时触发
func WithTrigger(ctx context.Context, trigger string) context.Context {
	return context.WithValue(ctx, keyTrigger, trigger)
}

func triggerFrom(ctx context.Context) string {
	if v, ok := ctx.Value(keyTrigger).(string); ok && v != "" {
		return v
	}
	return "manual"
}

// withTaskLog 把日志上下文注入 ctx，TaskLog 调用时从中读
func withTaskLog(ctx context.Context, repo *repository.TaskRepository, scheduleId, execId uint, persist bool) context.Context {
	ctx = context.WithValue(ctx, keyScheduleId, scheduleId)
	ctx = context.WithValue(ctx, keyExecId, execId)
	ctx = context.WithValue(ctx, keyRepo, repo)
	ctx = context.WithValue(ctx, keyPersist, persist)
	return ctx
}

// TaskLog 任务内部统一日志通道：emit 给前端 + 必要时写库
func TaskLog(ctx context.Context, level, message string) {
	scheduleId, _ := ctx.Value(keyScheduleId).(uint)
	now := time.Now()

	runtime.EventsEmit(ctx, "task_log", map[string]any{
		"scheduleId": scheduleId,
		"level":      level,
		"message":    message,
		"time":       now.Unix(),
	})

	persist, _ := ctx.Value(keyPersist).(bool)
	if !persist {
		return
	}
	repo, _ := ctx.Value(keyRepo).(*repository.TaskRepository)
	execId, _ := ctx.Value(keyExecId).(uint)
	if repo == nil || execId == 0 {
		return
	}
	_ = repo.CreateTaskLog(&model.TaskLog{
		ExecutionID: execId,
		Level:       level,
		Message:     message,
		CreatedAt:   now,
	})
}
