package alert

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// StdoutNotifier writes alert events to an io.Writer (defaults to os.Stdout).
type StdoutNotifier struct {
	out    io.Writer
	prefix string
}

// StdoutOption configures a StdoutNotifier.
type StdoutOption func(*StdoutNotifier)

// WithWriter overrides the default output writer.
func WithWriter(w io.Writer) StdoutOption {
	return func(s *StdoutNotifier) {
		s.out = w
	}
}

// WithPrefix sets a log line prefix (e.g. "[portwatch]").
func WithPrefix(p string) StdoutOption {
	return func(s *StdoutNotifier) {
		s.prefix = p
	}
}

// NewStdoutNotifier returns a Notifier that prints events to stdout.
func NewStdoutNotifier(opts ...StdoutOption) *StdoutNotifier {
	s := &StdoutNotifier{
		out:    os.Stdout,
		prefix: "[portwatch]",
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Send writes each event as a formatted line to the configured writer.
func (s *StdoutNotifier) Send(_ context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	ts := time.Now().UTC().Format(time.RFC3339)
	for _, e := range events {
		_, err := fmt.Fprintf(s.out, "%s %s %s\n", ts, s.prefix, FormatEvent(e))
		if err != nil {
			return err
		}
	}
	return nil
}
