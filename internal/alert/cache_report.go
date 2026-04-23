package alert

import (
	"fmt"
	"io"
	"os"
	"time"
)

// WriteCacheReport writes a human-readable summary of the CacheNotifier
// state to w. If w is nil, os.Stderr is used.
func WriteCacheReport(c *CacheNotifier, w io.Writer) {
	if w == nil {
		w = os.Stderr
	}

	events, at := c.Cached()
	if events == nil {
		fmt.Fprintln(w, "cache: empty")
		return
	}

	age := time.Since(at).Round(time.Millisecond)
	fmt.Fprintf(w, "cache: %d event(s), stored %s ago\n", len(events), age)
	for i, ev := range events {
		fmt.Fprintf(w, "  [%d] %s %s:%d/%s\n",
			i+1,
			ev.Kind,
			ev.Entry.Address,
			ev.Entry.Port,
			ev.Entry.Protocol,
		)
	}
}
