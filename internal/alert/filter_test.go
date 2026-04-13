package alert

import (
	"errors"
	"testing"

	"github.com/user/portwatch/internal/scanner"
)

type captureNotifier struct {
	received []Event
	err      error
}

func (c *captureNotifier) Send(events []Event) error {
	c.received = append(c.received, events...)
	return c.err
}

func makeFilterEntry(port uint16, addr string) scanner.Entry {
	return scanner.Entry{LocalPort: port, LocalAddr: addr}
}

func TestFilter_AllowsNonExcluded(t *testing.T) {
	cap := &captureNotifier{}
	f := NewFilter(cap, nil, nil)
	events := []Event{
		{Kind: KindAdded, Entry: makeFilterEntry(8080, "0.0.0.0")},
	}
	if err := f.Send(events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.received) != 1 {
		t.Fatalf("expected 1 event forwarded, got %d", len(cap.received))
	}
}

func TestFilter_ExcludesPort(t *testing.T) {
	cap := &captureNotifier{}
	f := NewFilter(cap, []uint16{8080}, nil)
	events := []Event{
		{Kind: KindAdded, Entry: makeFilterEntry(8080, "0.0.0.0")},
		{Kind: KindAdded, Entry: makeFilterEntry(9090, "0.0.0.0")},
	}
	if err := f.Send(events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.received) != 1 {
		t.Fatalf("expected 1 event, got %d", len(cap.received))
	}
	if cap.received[0].Entry.LocalPort != 9090 {
		t.Errorf("expected port 9090, got %d", cap.received[0].Entry.LocalPort)
	}
}

func TestFilter_ExcludesAddress(t *testing.T) {
	cap := &captureNotifier{}
	f := NewFilter(cap, nil, []string{"127."})
	events := []Event{
		{Kind: KindAdded, Entry: makeFilterEntry(3000, "127.0.0.1")},
		{Kind: KindAdded, Entry: makeFilterEntry(4000, "0.0.0.0")},
	}
	if err := f.Send(events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.received) != 1 {
		t.Fatalf("expected 1 event, got %d", len(cap.received))
	}
	if cap.received[0].Entry.LocalPort != 4000 {
		t.Errorf("expected port 4000, got %d", cap.received[0].Entry.LocalPort)
	}
}

func TestFilter_AllSuppressedReturnsNil(t *testing.T) {
	cap := &captureNotifier{}
	f := NewFilter(cap, []uint16{22}, nil)
	events := []Event{
		{Kind: KindAdded, Entry: makeFilterEntry(22, "0.0.0.0")},
	}
	if err := f.Send(events); err != nil {
		t.Fatalf("expected nil when all events suppressed, got %v", err)
	}
	if len(cap.received) != 0 {
		t.Errorf("expected inner notifier not called")
	}
}

func TestFilter_PropagatesInnerError(t *testing.T) {
	want := errors.New("inner failure")
	cap := &captureNotifier{err: want}
	f := NewFilter(cap, nil, nil)
	events := []Event{
		{Kind: KindAdded, Entry: makeFilterEntry(8080, "0.0.0.0")},
	}
	if err := f.Send(events); !errors.Is(err, want) {
		t.Errorf("expected %v, got %v", want, err)
	}
}
