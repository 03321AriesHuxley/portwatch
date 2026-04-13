package alert_test

import (
	"context"
	"errors"

	"github.com/yourusername/portwatch/internal/alert"
)

// captureNotifier records events it receives without error.
type captureNotifier struct {
	Received []alert.Event
}

func (c *captureNotifier) Send(_ context.Context, events []alert.Event) error {
	c.Received = append(c.Received, events...)
	return nil
}

// failNotifier always returns the configured error.
type failNotifier struct {
	err error
}

func (f *failNotifier) Send(_ context.Context, _ []alert.Event) error {
	if f.err != nil {
		return f.err
	}
	return errors.New("failNotifier: generic failure")
}
