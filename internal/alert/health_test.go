package alert_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/alert"
)

func TestHealthReport_Empty(t *testing.T) {
	var r alert.HealthReport
	var buf bytes.Buffer
	if err := r.WriteTo(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "NOTIFIER") {
		t.Error("expected header row in output")
	}
}

func TestHealthReport_WritesRows(t *testing.T) {
	var r alert.HealthReport
	r.Add("stdout", alert.MetricsSnapshot{TotalSent: 5, TotalFailed: 1, TotalFiltered: 2})
	r.Add("slack", alert.MetricsSnapshot{TotalSent: 3, TotalFailed: 0, TotalFiltered: 0})

	var buf bytes.Buffer
	if err := r.WriteTo(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"stdout", "slack", "5", "1", "2", "3"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
}

func TestHealthReport_TotalFailed(t *testing.T) {
	var r alert.HealthReport
	r.Add("a", alert.MetricsSnapshot{TotalFailed: 2})
	r.Add("b", alert.MetricsSnapshot{TotalFailed: 3})

	if got := r.TotalFailed(); got != 5 {
		t.Errorf("expected TotalFailed=5, got %d", got)
	}
}

func TestHealthReport_TotalSent(t *testing.T) {
	var r alert.HealthReport
	r.Add("a", alert.MetricsSnapshot{TotalSent: 10})
	r.Add("b", alert.MetricsSnapshot{TotalSent: 7})

	if got := r.TotalSent(); got != 17 {
		t.Errorf("expected TotalSent=17, got %d", got)
	}
}

func TestMetricsNotifier_IntegratesWithHealthReport(t *testing.T) {
	inner := &stubNotifier{}
	mn := alert.NewMetricsNotifier(inner, nil)

	_ = mn.Send(nil, []alert.Event{makeMetricsEvent()})
	_ = mn.Send(nil, []alert.Event{})

	var r alert.HealthReport
	r.Add("test", mn.Snapshot())

	if r.TotalSent() != 1 {
		t.Errorf("expected TotalSent=1, got %d", r.TotalSent())
	}
}
