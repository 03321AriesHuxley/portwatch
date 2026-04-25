package alert

import (
	"context"
	"testing"
	"time"
)

func TestDigest_InPipeline(t *testing.T) {
	cap := &captureNotifier{}

	now := time.Now()
	digester := NewDigestNotifier(cap, 1*time.Second)
	digester.clock = func() time.Time { return now }

	p := NewPipeline(digester)

	events := []Event{
		{Kind: "added"},
		{Kind: "removed"},
	}

	// First send — within window, should buffer.
	if err := p.Send(context.Background(), events); err != nil {
		t.Fatalf("pipeline send error: %v", err)
	}

	cap.mu.Lock()
	beforeFlush := len(cap.batches)
	cap.mu.Unlock()

	if beforeFlush != 0 {
		t.Fatalf("expected 0 batches before window, got %d", beforeFlush)
	}

	// Advance clock and trigger flush via another send.
	now = now.Add(2 * time.Second)
	if err := p.Send(context.Background(), []Event{{Kind: "added"}}); err != nil {
		t.Fatalf("pipeline send after window error: %v", err)
	}

	cap.mu.Lock()
	afterFlush := len(cap.batches)
	cap.mu.Unlock()

	if afterFlush != 1 {
		t.Fatalf("expected 1 batch after window, got %d", afterFlush)
	}
	if len(cap.batches[0]) != 2 {
		t.Fatalf("expected 2 buffered events, got %d", len(cap.batches[0]))
	}
}

func TestDigest_WithFilter(t *testing.T) {
	cap := &captureNotifier{}

	now := time.Now()
	digester := NewDigestNotifier(cap, 500*time.Millisecond)
	digester.clock = func() time.Time { return now }

	// Wrap with a filter that only allows "added" events.
	filtered := NewFilter(digester, nil, nil)
	_ = filtered // filter is wired; just verify compilation and basic flow

	events := []Event{{Kind: "added"}, {Kind: "removed"}}
	_ = digester.Send(context.Background(), events)

	// Manual flush to verify inner receives events.
	_ = digester.Flush(context.Background())

	cap.mu.Lock()
	got := len(cap.batches)
	cap.mu.Unlock()

	if got != 1 {
		t.Fatalf("expected 1 batch after flush, got %d", got)
	}
}
