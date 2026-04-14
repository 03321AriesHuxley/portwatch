package alert

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
	"time"
)

// WriteSnapshotReport writes a human-readable table of all notifier snapshots
// to w. Rows are sorted alphabetically by notifier name.
func WriteSnapshotReport(w io.Writer, store *SnapshotStore) error {
	snaps := store.All()
	sort.Slice(snaps, func(i, j int) bool {
		return snaps[i].Name < snaps[j].Name
	})

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NOTIFIER\tSENT\tFAILED\tFILTERED\tLAST SENT\tLAST ERROR")

	for _, s := range snaps {
		lastSent := "-"
		if !s.LastSent.IsZero() {
			lastSent = s.LastSent.Format(time.RFC3339)
		}
		lastErr := "-"
		if s.LastError != nil {
			lastErr = s.LastError.Error()
		}
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%s\t%s\n",
			s.Name, s.Sent, s.Failed, s.Filtered, lastSent, lastErr)
	}

	return tw.Flush()
}
