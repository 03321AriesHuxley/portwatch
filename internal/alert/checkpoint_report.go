package alert

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
	"time"
)

// WriteCheckpointReport writes a human-readable table of checkpoint records
// from store to w. The report is sorted alphabetically by notifier name.
func WriteCheckpointReport(w io.Writer, store *CheckpointStore, now time.Time) error {
	store.mu.RLock()
	names := make([]string, 0, len(store.records))
	for name := range store.records {
		names = append(names, name)
	}
	store.mu.RUnlock()

	sort.Strings(names)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NOTIFIER\tLAST SUCCESS\tAGE")

	if len(names) == 0 {
		fmt.Fprintln(tw, "(none)\t-\t-")
		return tw.Flush()
	}

	for _, name := range names {
		t, ok := store.Last(name)
		if !ok {
			continue
		}
		age := now.Sub(t).Truncate(time.Second)
		fmt.Fprintf(tw, "%s\t%s\t%s ago\n", name, t.Format(time.RFC3339), age)
	}

	return tw.Flush()
}
