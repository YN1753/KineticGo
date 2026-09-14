package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kineticgo/internal/model"
	"kineticgo/internal/repository"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	// catchupWindow 启动/唤醒后，错过的调度点在此窗口内则补跑一次
	catchupWindow = 3 * time.Hour
	// retryDelays 失败后的调度层重试间隔（仅 schedule/catchup/retry 触发）
	maxRetryAttempts = 3
)

var retryDelays = []time.Duration{30 * time.Second, 2 * time.Minute, 5 * time.Minute}

type TaskManageService struct {
	TaskRepo  *repository.TaskRepository
	registry  map[string]func() model.TaskInstance
	running   map[uint]model.RunningTask
	scheduler *Scheduler
	rootCtx   context.Context
	mutex     sync.RWMutex
	// retryTimers 记录每个 schedule 的待执行重试，便于 Stop/Shutdown 时取消
	retryTimers map[uint]*time.Timer
}

// 初始化部分方法和函数
func NewTaskManageService(task *repository.TaskRepository) *TaskManageService {
	t := &TaskManageService{
		TaskRepo:    task,
		registry:    make(map[string]func() model.TaskInstance),
		running:     make(map[uint]model.RunningTask),
		retryTimers: make(map[uint]*time.Timer),
		rootCtx:     context.Background(),
	}
	t.scheduler = NewScheduler(t.runScheduled)
	return t
}

// SetRootCtx 由 app 在 OnStartup 时注入，cron 触发的任务用这个 ctx，避免和窗口 ctx 绑死
// SetRootCtx 由 app 在 OnStartup 时调用。任务取消树固定使用 Background，
// 与窗口 ctx 解绑；前端推送通过 service.SetAppEventCtx 注入窗口 ctx。
func (t *TaskManageService) SetRootCtx(_ context.Context) {
	t.rootCtx = context.Background()
}

// runScheduled 是注入给 Scheduler 的回调，到点拉起任务
func (t *TaskManageService) runScheduled(scheduleId uint) {
	root := t.rootCtx
	if root == nil {
		root = context.Background()
	}
	// 触发瞬间就刷新 next_run_time，避免长任务运行期间 UI 展示过期值
	t.refreshNextRunTime(scheduleId)

	ctx := WithTrigger(root, "schedule")
	if err := t.Start(ctx, scheduleId); err != nil {
		// 上次仍在跑导致的 skip 要让用户在会话日志里看见，而不是只打到 stdout
		level := LogError
		if err.Error() == "任务已经在运行" {
			level = LogWarn
		}
		emitTaskLog(root, scheduleId, level, "定时触发跳过: "+err.Error())
	}
}

func emitTaskLog(ctx context.Context, scheduleId uint, level, message string) {
	// 前端推送统一走窗口 ctx，任务 ctx 只携带业务上下文
	_ = ctx
	runtime.EventsEmit(eventCtx(), "task_log", map[string]any{
		"scheduleId": scheduleId,
		"level":      level,
		"message":    message,
		"time":       time.Now().Unix(),
	})
}

func (t *TaskManageService) refreshNextRunTime(scheduleId uint) {
	task, err := t.TaskRepo.GetTaskScheduleById(scheduleId)
	if err != nil || task.CronExpr == "" {
		return
	}
	sched, parseErr := secondParser.Parse(task.CronExpr)
	if parseErr != nil {
		return
	}
	_ = t.TaskRepo.UpdateScheduleNextRunTime(scheduleId, sched.Next(time.Now()))
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

// PrepareSchedules 启动时收尾上次异常退出的执行记录，并对错过的 cron 窗口做一次补跑
func (t *TaskManageService) PrepareSchedules() {
	if err := t.TaskRepo.MarkInterruptedExecutions(); err != nil && t.rootCtx != nil {
		runtime.EventsEmit(t.rootCtx, "log", "标记中断执行记录失败: "+err.Error())
	}
	t.CatchUpMissedTasks()
}

// CatchUpMissedTasks 扫描启用中的 cron 任务，若最近一次应触发点落在 catch-up 窗口内
// 且该点之后没有成功记录，则立刻补跑一次（去重靠 HasSuccessSince）。
func (t *TaskManageService) CatchUpMissedTasks() {
	list, err := t.TaskRepo.GetEnabledCronSchedules()
	if err != nil {
		return
	}
	now := time.Now()
	for i := range list {
		sch := list[i]
		sched, parseErr := secondParser.Parse(sch.CronExpr)
		if parseErr != nil {
			continue
		}
		lastDue := lastDueBefore(sched, now, catchupWindow)
		if lastDue.IsZero() {
			continue
		}
		// 窗口外的漏跑不再追
		if now.Sub(lastDue) > catchupWindow {
			continue
		}
		ok, herr := t.TaskRepo.HasSuccessSince(sch.ID, lastDue)
		if herr != nil || ok {
			continue
		}
		// 正在跑就不要叠
		t.mutex.RLock()
		_, busy := t.running[sch.ID]
		t.mutex.RUnlock()
		if busy {
			continue
		}

		ctx := WithTrigger(t.rootCtx, "catchup")
		if t.rootCtx == nil {
			ctx = WithTrigger(context.Background(), "catchup")
		}
		emitTaskLog(ctx, sch.ID, LogWarn,
			"检测到漏跑窗口，开始补跑（应于 "+lastDue.Format("01-02 15:04:05")+" 触发）")
		if err := t.Start(ctx, sch.ID); err != nil {
			emitTaskLog(ctx, sch.ID, LogError, "补跑启动失败: "+err.Error())
		}
	}
}

// lastDueBefore 计算 t 之前（含回看窗口内）最近一次调度点；无则返回零值
func lastDueBefore(sched cron.Schedule, now time.Time, lookback time.Duration) time.Time {
	cursor := now.Add(-lookback)
	var last time.Time
	// 保护：避免异常 cron 导致死循环
	for i := 0; i < 10000; i++ {
		next := sched.Next(cursor)
		if next.IsZero() || next.After(now) {
			return last
		}
		last = next
		cursor = next
	}
	return last
}

// Shutdown 由 app 在 OnShutdown 时调用，优雅停掉所有定时任务
func (t *TaskManageService) Shutdown() {
	t.scheduler.Stop()
	t.mutex.Lock()
	for id, timer := range t.retryTimers {
		timer.Stop()
		delete(t.retryTimers, id)
	}
	// 取消仍在跑的任务，避免关停后 goroutine 继续写库/推日志
	for id, rt := range t.running {
		if rt.Cancel != nil {
			rt.Cancel()
		}
		delete(t.running, id)
	}
	t.mutex.Unlock()
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
		runtime.EventsEmit(ctx, "log", "获取开机启动任务失败:"+err.Error())
		return
	}

	for _, schedule := range schedules {
		if err := t.Start(ctx, schedule.ID); err != nil {
			// 单个系统任务失败不应拖死其余监控
			runtime.EventsEmit(ctx, "log", "自动启动系统任务 ["+schedule.Name+"] 失败:"+err.Error())
		}
	}
}

//

func (t *TaskManageService) GetTaskList() (*[]model.Task, error) {
	return t.TaskRepo.GetTaskList()
}

func (t *TaskManageService) Start(ctx context.Context, scheduleId uint) error {
	// 整段持锁完成「判重 + 占坑」，消除 check-then-act 竞态
	t.mutex.Lock()
	if _, ok := t.running[scheduleId]; ok {
		t.mutex.Unlock()
		return errors.New("任务已经在运行")
	}
	// 先占坑，避免 DB I/O 期间并发 Start 双开
	t.running[scheduleId] = model.RunningTask{}
	t.mutex.Unlock()

	// 任一后续步骤失败时释放占坑
	releaseSlot := func() {
		t.mutex.Lock()
		delete(t.running, scheduleId)
		t.mutex.Unlock()
	}

	task, err := t.TaskRepo.GetTaskScheduleById(scheduleId)
	if err != nil {
		releaseSlot()
		return errors.New("获取任务失败")
	}
	if task.TaskType == "system" {
		task.TaskType = "system-" + task.Name
	}
	factory, ok := t.registry[task.TaskType]
	if !ok {
		releaseSlot()
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
	trigger := triggerFrom(ctx)
	attempt := attemptFrom(ctx)

	t.mutex.Lock()
	t.running[scheduleId] = model.RunningTask{
		Instance: instance,
		Cancel:   cancel,
	}
	t.mutex.Unlock()

	go func() {
		// 任务结束时必须 cancel，避免 context 泄漏
		defer cancel()

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
			emitTaskLog(childCtx, scheduleId, LogError, runErr.Error())
		}

		// 调度/补跑/重试失败后，在窗口内做有限次延迟重试
		if runErr != nil && !isSystem && shouldRetryTrigger(trigger) {
			t.scheduleRetry(scheduleId, attempt)
		}

		t.mutex.Lock()
		delete(t.running, scheduleId)
		t.mutex.Unlock()
		t.TaskRepo.RemoveActiveTask(&t.TaskRepo.ActiveTasks)
	}()

	return nil
}

func shouldRetryTrigger(trigger string) bool {
	switch trigger {
	case "schedule", "catchup", "retry":
		return true
	default:
		return false
	}
}

// scheduleRetry 在任务失败后按退避间隔再拉起一次；同一 schedule 只保留一个待触发 timer
func (t *TaskManageService) scheduleRetry(scheduleId uint, failedAttempt int) {
	nextAttempt := failedAttempt + 1
	if nextAttempt > maxRetryAttempts {
		return
	}
	delay := retryDelays[min(nextAttempt-1, len(retryDelays)-1)]

	t.mutex.Lock()
	if old, ok := t.retryTimers[scheduleId]; ok {
		old.Stop()
	}
	timer := time.AfterFunc(delay, func() {
		t.mutex.Lock()
		delete(t.retryTimers, scheduleId)
		t.mutex.Unlock()

		root := t.rootCtx
		if root == nil {
			root = context.Background()
		}
		ctx := WithTrigger(root, "retry")
		ctx = WithAttempt(ctx, nextAttempt)
		emitTaskLog(ctx, scheduleId, LogWarn,
			fmt.Sprintf("第 %d 次自动重试（间隔 %s）", nextAttempt, delay))
		if err := t.Start(ctx, scheduleId); err != nil {
			emitTaskLog(ctx, scheduleId, LogError, "重试启动失败: "+err.Error())
		}
	})
	t.retryTimers[scheduleId] = timer
	t.mutex.Unlock()
}

// RunImmediate 运行一个即时任务（不依赖 schedule 表，但记录 execution 和 log）
func (t *TaskManageService) RunImmediate(ctx context.Context, taskType string, targetID uint) error {
	factory, ok := t.registry[taskType]
	if !ok {
		return errors.New("未找到匹配的任务处理器: " + taskType)
	}

	instance := factory()
	childCtx, cancel := context.WithCancel(ctx)

	// 创建一个虚拟的 execution 记录 (OptionID 设为 0 表示非预设任务)
	exec := &model.TaskExecution{
		OptionID:      0,
		TriggerType:   "immediate",
		Status:        "running",
		StartTime:     time.Now(),
		ResultSummary: taskType + " 即时任务",
	}
	_ = t.TaskRepo.CreateTaskExecution(exec)
	execId := exec.ID

	childCtx = withTaskLog(childCtx, t.TaskRepo, 0, execId, true)

	go func() {
		defer cancel()
		t.TaskRepo.AddActiveTask(&t.TaskRepo.ActiveTasks)
		runErr := instance.Run(childCtx, targetID)

		status := "success"
		summary := "即时任务执行成功"
		if runErr != nil {
			status = "failed"
			summary = runErr.Error()
			TaskLog(childCtx, LogError, runErr.Error())
		}
		_ = t.TaskRepo.UpdateTaskExecution(execId, status, summary, time.Now())
		t.TaskRepo.RemoveActiveTask(&t.TaskRepo.ActiveTasks)
	}()

	return nil
}

func (t *TaskManageService) Stop(scheduleId uint) error {
	// 用户手动停止时取消待触发的自动重试
	t.mutex.Lock()
	if timer, ok := t.retryTimers[scheduleId]; ok {
		timer.Stop()
		delete(t.retryTimers, scheduleId)
	}
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

var secondParser = cron.NewParser(
	cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

func prepareCron(sch *model.TaskSchedule) error { //检验cron合法性并且算出下次时间
	if sch.CronExpr == "" {
		sch.NextRunTime = time.Time{}
		return nil
	}
	schedule, err := secondParser.Parse(sch.CronExpr)
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
