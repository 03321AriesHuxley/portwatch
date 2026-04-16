package alert

import (
	"fmt"
	"io"
	"time"
)

// WriteReplayReport writes a human-readable summary of replayed batches to w.
func WriteReplayReport(w io.Writer, store *ReplayStore, since time.Time) {
	batches := store.Since(since)
	if len(batches) == 0 {
		fmt.Fprintln(w, "no replayed events")
		return
	}
	total := 0
	for _, b := range batches {
		total += len(b)
	}
	fmt.Fprintf(w, "replay: %d batch(es), %d event(s) since %s\n",
		len(batches), total, since.Format(time.RFC3339))
	for i, batch := range batches {
		fmt.Fprintf(w, "  batch %d:\n", i+1)
		for _, e := range batch {
			fmt.Fprintf(w, "    %s\n", FormatEvent(e))
		}
	}
}
