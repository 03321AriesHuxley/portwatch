package alert_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
)

func TestWriteSnapshotReport_Empty(t *testing.T) {
	store := alert.NewSnapshotStore()
	var buf bytes.Buffer
	if err := alert.WriteSnapshotReport(&buf, store); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "NOTIFIER") {
		t.Error("expected header row in output")
	}
}

func TestWriteSnapshotReport_ContainsName(t *testing.T) {
	store := alert.NewSnapshotStore()
	store.Record(alert.NotifierSnapshot{
		Name:     "slack",
		Sent:     5,
		Failed:   1,
		Filtered: 2,
		LastSent: time.Now(),
	})

	var buf bytes.Buffer
	_ = alert.WriteSnapshotReport(&buf, store)
	out := buf.String()

	for _, want := range []string{"slack", "5", "1", "2"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
}

func TestWriteSnapshotReport_ShowsLastError(t *testing.T) {
	store := alert.NewSnapshotStore()
	store.Record(alert.NotifierSnapshot{
		Name:      "pagerduty",
		Failed:    3,
		LastError: errors.New("connection refused"),
	})

	var buf bytes.Buffer
	_ = alert.WriteSnapshotReport(&buf, store)

	if !strings.Contains(buf.String(), "connection refused") {
		t.Error("expected last error message in output")
	}
}

func TestWriteSnapshotReport_SortedAlphabetically(t *testing.T) {
	store := alert.NewSnapshotStore()
	store.Record(alert.NotifierSnapshot{Name: "zebra", Sent: 1})
	store.Record(alert.NotifierSnapshot{Name: "alpha", Sent: 2})
	store.Record(alert.NotifierSnapshot{Name: "mango", Sent: 3})

	var buf bytes.Buffer
	_ = alert.WriteSnapshotReport(&buf, store)
	out := buf.String()

	alphaIdx := strings.Index(out, "alpha")
	mangoIdx := strings.Index(out, "mango")
	zebraIdx := strings.Index(out, "zebra")

	if !(alphaIdx < mangoIdx && mangoIdx < zebraIdx) {
		t.Errorf("expected alphabetical order, got indices alpha=%d mango=%d zebra=%d",
			alphaIdx, mangoIdx, zebraIdx)
	}
}
