package workspace

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"

	"rolecraft-ai/internal/models"
)

type Scheduler struct {
	db       *gorm.DB
	runner   *Runner
	interval time.Duration
	cancel   context.CancelFunc
}

func NewScheduler(db *gorm.DB, runner *Runner, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Scheduler{
		db:       db,
		runner:   runner,
		interval: interval,
	}
}

func (s *Scheduler) Start(parent context.Context) {
	if s.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(parent)
	s.cancel = cancel

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// startup warm scan
		s.scanAndRun(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.scanAndRun(ctx)
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
}

func (s *Scheduler) scanAndRun(ctx context.Context) {
	now := time.Now()
	var due []models.Work
	if err := s.db.
		Where("trigger_type <> ? AND next_run_at IS NOT NULL AND next_run_at <= ? AND async_status IN ?", "manual", now, []string{"scheduled", "idle"}).
		Order("next_run_at ASC").
		Limit(20).
		Find(&due).Error; err != nil {
		log.Printf("workspace scheduler query failed: %v", err)
		return
	}
	if len(due) == 0 {
		return
	}

	for _, item := range due {
		work, claimed, err := s.runner.ClaimWork(item.ID, item.UserID)
		if err != nil {
			log.Printf("workspace scheduler claim failed: work=%s err=%v", item.ID, err)
			continue
		}
		if !claimed {
			continue
		}
		_, err = s.runner.ExecuteClaimed(ctx, &work, "scheduler")
		if err != nil {
			log.Printf("workspace scheduler run failed: work=%s err=%v", item.ID, err)
		}
	}
}
