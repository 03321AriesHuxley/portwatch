package daemon

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

// Runner holds the dependencies needed to run the polling loop.
type Runner struct {
	interval time.Duration
	procPath string
	logger   *alert.Logger
	hook     *alert.Hook
}

// NewRunner creates a Runner from the provided options.
func NewRunner(interval time.Duration, procPath string, logger *alert.Logger, hook *alert.Hook) *Runner {
	return &Runner{
		interval: interval,
		procPath: procPath,
		logger:   logger,
		hook:     hook,
	}
}

// Run starts the polling loop, blocking until ctx is cancelled.
func (r *Runner) Run(ctx context.Context) error {
	prev, err := scanner.Snapshot(r.procPath)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	log.Printf("portwatch: started, polling every %s", r.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("portwatch: shutting down")
			return ctx.Err()
		case <-ticker.C:
			curr, err := scanner.Snapshot(r.procPath)
			if err != nil {
				log.Printf("portwatch: snapshot error: %v", err)
				continue
			}

			diff := scanner.Diff(prev, curr)
			if len(diff.Opened)+len(diff.Closed) > 0 {
				r.logger.Send(diff)
				if r.hook != nil {
					if err := r.hook.Send(diff); err != nil {
						log.Printf("portwatch: hook error: %v", err)
					}
				}
			}
			prev = curr
		}
	}
}
