package alert

import (
	"context"
	"testing"

	"github.com/user/portwatch/internal/scanner"
)

func makeRedactEvent(addr string, meta map[string]string) Event {
	return Event{
		Kind:  "added",
		Entry: scanner.Entry{Addr: addr, Port: 8080, Protocol: "tcp"},
		Meta:  meta,
	}
}

func TestRedactNotifier_EmptyEvents(t *testing.T) {
	var got []Event
	next := NotifierFunc(func(_ context.Context, evs []Event) error {
		got = evs
		return nil
	})
	n := NewRedactNotifier(next)
	if err := n.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no events forwarded")
	}
}

func TestRedactNotifier_MetaRedaction(t *testing.T) {
	var got []Event
	next := NotifierFunc(func(_ context.Context, evs []Event) error {
		got = evs
		return nil
	})
	n := NewRedactNotifier(next,
		WithMetaRedaction("token", func(v string) string { return "***" }),
	)
	events := []Event{
		makeRedactEvent("127.0.0.1", map[string]string{"token": "secret123", "user": "alice"}),
	}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(got))
	}
	if got[0].Meta["token"] != "***" {
		t.Errorf("expected token redacted, got %q", got[0].Meta["token"])
	}
	if got[0].Meta["user"] != "alice" {
		t.Errorf("expected user preserved, got %q", got[0].Meta["user"])
	}
}

func TestRedactNotifier_NilMetaUnchanged(t *testing.T) {
	var got []Event
	next := NotifierFunc(func(_ context.Context, evs []Event) error {
		got = evs
		return nil
	})
	n := NewRedactNotifier(next,
		WithMetaRedaction("token", func(v string) string { return "***" }),
	)
	events := []Event{makeRedactEvent("127.0.0.1", nil)}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Meta != nil {
		t.Errorf("expected nil meta preserved")
	}
}

func TestMaskRight(t *testing.T) {
	cases := []struct {
		input string
		n     int
		want  string
	}{
		{"secret", 2, "se****"},
		{"secret", 0, "******"},
		{"secret", 10, "secret"},
		{"", 2, ""},
	}
	for _, tc := range cases {
		got := MaskRight(tc.input, tc.n)
		if got != tc.want {
			t.Errorf("MaskRight(%q, %d) = %q, want %q", tc.input, tc.n, got, tc.want)
		}
	}
}

func TestMaskIP_IPv4(t *testing.T) {
	got := maskIP("192.168.1.100")
	if got != "192.168.***.*" {
		t.Errorf("unexpected masked IP: %q", got)
	}
}

func TestMaskIP_WithPort(t *testing.T) {
	got := maskIP("10.0.0.1:9090")
	if got != "10.0.***.*:9090" {
		t.Errorf("unexpected masked addr: %q", got)
	}
}
