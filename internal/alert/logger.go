package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/scanner"
)

// Logger writes port change events as human-readable lines to a writer.
type Logger struct {
	out io.Writer
}

// NewLogger creates a Logger writing to the given writer.
// If w is nil, os.Stdout is used.
func NewLogger(w io.Writer) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return &Logger{out: w}
}

// Log prints added and removed sockets from a Diff with a timestamp.
func (l *Logger) Log(diff scanner.Diff) {
	if len(diff.Added) == 0 && len(diff.Removed) == 0 {
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)

	for _, s := range diff.Added {
		fmt.Fprintf(l.out, "%s [OPEN]   %s:%d\n", now, s.LocalAddress, s.LocalPort)
	}
	for _, s := range diff.Removed {
		fmt.Fprintf(l.out, "%s [CLOSE]  %s:%d\n", now, s.LocalAddress, s.LocalPort)
	}
}
