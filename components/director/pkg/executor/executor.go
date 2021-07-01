package executor

import (
	"context"
	"time"
)

type periodic struct {
	refreshPeriod time.Duration
	executionFunc func(context.Context)
}

// NewPeriodic creates a periodic executor, which calls given executionFunc periodically.
func NewPeriodic(period time.Duration, executionFunc func(context.Context)) *periodic {
	return &periodic{
		refreshPeriod: period,
		executionFunc: executionFunc,
	}
}

// Run starts a periodic worker
func (e *periodic) Run(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(e.refreshPeriod)
		for {
			e.executionFunc(ctx)
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
			}
		}
	}()
}
