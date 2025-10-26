package scheduler

import (
	"context"
	"log"
	"time"
)

type Job struct {
	Interval time.Duration
	Task     func(ctx context.Context)
	cancel   context.CancelFunc
}

func (j *Job) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	j.cancel = cancel

	go func() {
		ticker := time.NewTicker(j.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				go func() {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("job panicked: %v", r)
						}
					}()
					j.Task(ctx)
				}()
			case <-ctx.Done():
				log.Println("scheduler stopped")
				return
			}
		}
	}()
}

func (j *Job) Stop() {
	if j.cancel != nil {
		j.cancel()
	}
}
