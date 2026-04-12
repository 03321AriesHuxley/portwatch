package daemon

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

// Watcher holds the state needed to poll for port changes and dispatch alerts.
type Watcher struct {
	interval   time.Duration
	procPath   string
	logger     *alert.Logger
	hook       *alert.Hook
	prevPorts  []scanner.PortEntry
}

// NewWatcher constructs a Watcher with the provided dependencies.
func NewWatcher(interval time.Duration, procPath string, logger *alert.Logger, hook *alert.Hook) *Watcher {
	return &Watcher{
		interval: interval,
		procPath: procPath,
		logger:   logger,
		hook:     hook,
	}
}

// Run starts the polling loop. It blocks until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Seed initial snapshot so the first tick only reports real changes.
	initial, err := scanner.Snapshot(w.procPath)
	if err != nil {
		log.Printf("[watcher] initial snapshot error: %v", err)
	} else {
		w.prevPorts = initial
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.poll(); err != nil {
				log.Printf("[watcher] poll error: %v", err)
			}
		}
	}
}

func (w *Watcher) poll() error {
	current, err := scanner.Snapshot(w.procPath)
	if err != nil {
		return err
	}

	diff := scanner.Diff(w.prevPorts, current)
	w.prevPorts = current

	if len(diff.Added) == 0 && len(diff.Removed) == 0 {
		return nil
	}

	w.logger.Send(diff)

	if w.hook != nil {
		if err := w.hook.Send(diff); err != nil {
			log.Printf("[watcher] hook error: %v", err)
		}
	}

	return nil
}
