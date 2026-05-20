package service

import (
	"context"
	"sync"
	"time"

	"kineticgo/internal/model"
	"kineticgo/internal/repository"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// 日志级别
const (
	LogInfo    = "info"
	LogWarn    = "warn"
	LogError   = "error"
	LogSuccess = "success"
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

// ============================================================
//  日志前端推送 — 串行化 + 节流（解决 macOS 频繁 EventsEmit 被杀）
// ============================================================

// logEntry 单条日志记录
type logEntry struct {
	ctx        context.Context
	scheduleId uint
	level      string
	message    string
	time       time.Time
}

var (
	logCh     chan logEntry
	logInit   sync.Once
	logStopCh chan struct{}
)

// StartLogEmitter 在应用启动时调用，启动后台日志消费 goroutine.
// 所有 TaskLog 产生的日志都先进 channel，由单消费者串行推送到前端，
// 避免多个任务 goroutine 并发调用 runtime.EventsEmit 压垮 macOS 主线程.
func StartLogEmitter() {
	logInit.Do(func() {
		logCh = make(chan logEntry, 100)
		logStopCh = make(chan struct{})
		go func() {
			for {
				select {
				case entry, ok := <-logCh:
					if !ok {
						return
					}
					runtime.EventsEmit(entry.ctx, "task_log", map[string]any{
						"scheduleId": entry.scheduleId,
						"level":      entry.level,
						"message":    entry.message,
						"time":       entry.time.Unix(),
					})
					// 节流：每次推送后 sleep 20ms，防止短时间内高频 EventsEmit
					// 8 条日志总延迟增加约 140ms，前端完全无感知
					time.Sleep(20 * time.Millisecond)

				case <-logStopCh:
					return
				}
			}
		}()
	})
}

// StopLogEmitter 在应用退出时调用，优雅关闭日志消费 goroutine.
func StopLogEmitter() {
	if logStopCh != nil {
		close(logStopCh)
	}
	if logCh != nil {
		close(logCh)
	}
}

// TaskLog 任务内部统一日志通道：emit 给前端 + 必要时写库.
// 前端推送改为异步 channel，不再在业务 goroutine 中直接调用 EventsEmit.
func TaskLog(ctx context.Context, level, message string) {
	scheduleId, _ := ctx.Value(keyScheduleId).(uint)
	now := time.Now()

	// 异步推送到前端（非阻塞，channel 满则丢弃，避免阻塞业务）
	select {
	case logCh <- logEntry{
		ctx:        ctx,
		scheduleId: scheduleId,
		level:      level,
		message:    message,
		time:       now,
	}:
	default:
		// channel 已满（100 条），直接丢弃前端推送，保留持久化
	}

	// 持久化逻辑不变
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
