package service

import (
	"kineticgo/internal/model"
	"sync"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron    *cron.Cron
	entries map[uint]cron.EntryID // scheduleID -> entry
	mu      sync.Mutex
	runFn   func(scheduleID uint) // 注入：到点要做什么
}

func NewScheduler(runFn func(uint)) *Scheduler {
	c := cron.New(cron.WithSeconds(), cron.WithChain(
		cron.Recover(cron.DefaultLogger),            // panic 不拖死调度器
		cron.SkipIfStillRunning(cron.DefaultLogger), // 上次没跑完就跳过本次
	))
	return &Scheduler{cron: c, entries: map[uint]cron.EntryID{}, runFn: runFn}
}

func (s *Scheduler) Add(sch *model.TaskSchedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.entries[sch.ID]; ok { // 已存在先移除，等价于 Update
		s.cron.Remove(old)
		delete(s.entries, sch.ID)
	}
	id, err := s.cron.AddFunc(sch.CronExpr, func() { s.runFn(sch.ID) })
	if err != nil {
		return err
	}
	s.entries[sch.ID] = id
	return nil
}

func (s *Scheduler) Remove(scheduleID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.entries[scheduleID]; ok {
		s.cron.Remove(id)
		delete(s.entries, scheduleID)
	}
}

func (s *Scheduler) Start() { s.cron.Start() }
func (s *Scheduler) Stop()  { <-s.cron.Stop().Done() } // 等待正在跑的退出
