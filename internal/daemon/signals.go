package daemon

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// WithSignals returns a context that is cancelled when the process receives
// SIGINT or SIGTERM. The returned cancel function should be deferred by the
// caller to release resources.
func WithSignals(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer signal.Stop(ch)
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx, cancel
}
