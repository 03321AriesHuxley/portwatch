package daemon

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

// Runner periodically scans port state and emits alerts on changes.
type Runner struct {
	procNetPath string
	interval    time.Duration
	logger      *alert.Logger
	hook        *alert.Hook
}

// NewRunner constructs a Runner with the supplied configuration.
func NewRunner(procNetPath string, interval time.Duration, logger *alert.Logger, hook *alert.Hook) *Runner {
	return &Runner{
		procNetPath: procNetPath,
		interval:    interval,
		logger:      logger,
		hook:        hook,
	}
}

// Run starts the monitoring loop. It blocks until ctx is cancelled.
func (r *Runner) Run(ctx context.Context) error {
	prev, err := scanner.Snapshot(r.procNetPath)
	if err != nil {
		return fmt.Errorf("initial snapshot: %w", err)
	}

	return RunEvery(ctx, r.interval, func() {
		curr, err := scanner.Snapshot(r.procNetPath)
		if err != nil {
			log.Printf("portwatch: snapshot error: %v", err)
			return
		}

		diff := scanner.Diff(prev, curr)
		if len(diff.Added) == 0 && len(diff.Removed) == 0 {
			prev = curr
			return
		}

		if r.logger != nil {
			r.logger.Send(ctx, diff)
		}
		if r.hook != nil {
			if err := r.hook.Send(ctx, diff); err != nil {
				log.Printf("portwatch: hook error: %v", err)
			}
		}

		prev = curr
	})
}
