package alert

import "time"

// SetClock allows tests to inject a deterministic clock into DigestNotifier.
func (d *DigestNotifier) SetClock(fn func() time.Time) {
	d.clock = fn
}
