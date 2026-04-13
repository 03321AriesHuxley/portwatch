package alert

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// HealthReport aggregates MetricsSnapshots from named notifiers.
type HealthReport struct {
	entries []healthEntry
}

type healthEntry struct {
	name     string
	snapshot MetricsSnapshot
}

// Add registers a named snapshot into the report.
func (r *HealthReport) Add(name string, snap MetricsSnapshot) {
	r.entries = append(r.entries, healthEntry{name: name, snapshot: snap})
}

// WriteTo renders the report as a human-readable table to w.
func (r *HealthReport) WriteTo(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NOTIFIER\tSENT\tFAILED\tFILTERED")
	for _, e := range r.entries {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\n",
			e.name,
			e.snapshot.TotalSent,
			e.snapshot.TotalFailed,
			e.snapshot.TotalFiltered,
		)
	}
	return tw.Flush()
}

// TotalFailed returns the sum of failed sends across all registered notifiers.
func (r *HealthReport) TotalFailed() int64 {
	var total int64
	for _, e := range r.entries {
		total += e.snapshot.TotalFailed
	}
	return total
}

// TotalSent returns the sum of successful sends across all registered notifiers.
func (r *HealthReport) TotalSent() int64 {
	var total int64
	for _, e := range r.entries {
		total += e.snapshot.TotalSent
	}
	return total
}
