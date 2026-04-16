package alert

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestWriteReplayReport_Empty(t *testing.T) {
	store := NewReplayStore(10)
	var buf bytes.Buffer
	WriteReplayReport(&buf, store, time.Now().Add(-time.Hour))
	if !strings.Contains(buf.String(), "no replayed") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestWriteReplayReport_ShowsBatches(t *testing.T) {
	store := NewReplayStore(10)
	before := time.Now()
	store.Record(makeReplayEvents(8080))
	store.Record(makeReplayEvents(9090))
	var buf bytes.Buffer
	WriteReplayReport(&buf, store, before)
	out := buf.String()
	if !strings.Contains(out, "2 batch(es)") {
		t.Fatalf("expected batch count: %q", out)
	}
	if !strings.Contains(out, "2 event(s)") {
		t.Fatalf("expected event count: %q", out)
	}
}

func TestWriteReplayReport_ContainsPortInfo(t *testing.T) {
	store := NewReplayStore(10)
	before := time.Now()
	store.Record(makeReplayEvents(1234))
	var buf bytes.Buffer
	WriteReplayReport(&buf, store, before)
	if !strings.Contains(buf.String(), "1234") {
		t.Fatalf("expected port 1234 in output")
	}
}
