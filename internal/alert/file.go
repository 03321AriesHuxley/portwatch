package alert

import (
	"context"
	"fmt"
	"os"
	"time"
)

// FileNotifier appends alert events to a log file on disk.
type FileNotifier struct {
	path   string
	prefix string
}

// NewFileNotifier returns a Notifier that appends events to the given file path.
// The file is opened and closed on each Send to avoid holding file descriptors.
func NewFileNotifier(path string, prefix string) *FileNotifier {
	if prefix == "" {
		prefix = "[portwatch]"
	}
	return &FileNotifier{path: path, prefix: prefix}
}

// Send appends each event as a timestamped line to the log file.
func (f *FileNotifier) Send(_ context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	file, err := os.OpenFile(f.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("filenotifier: open %q: %w", f.path, err)
	}
	defer file.Close()

	ts := time.Now().UTC().Format(time.RFC3339)
	for _, e := range events {
		line := fmt.Sprintf("%s %s %s\n", ts, f.prefix, FormatEvent(e))
		if _, werr := file.WriteString(line); werr != nil {
			return fmt.Errorf("filenotifier: write: %w", werr)
		}
	}
	return nil
}
