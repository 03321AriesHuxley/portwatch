package alert

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type mockPDClient struct {
	statusCode int
	recorded   []pdPayload
	err        error
}

func (m *mockPDClient) Do(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	var p pdPayload
	body, _ := io.ReadAll(req.Body)
	_ = json.Unmarshal(body, &p)
	m.recorded = append(m.recorded, p)
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

func makePDEvent(kind EventKind, port uint16) Event {
	return Event{Kind: kind, Entry: makeFormatterEntry(port)}
}

func TestPagerDutyNotifier_SendEmpty(t *testing.T) {
	client := &mockPDClient{statusCode: 202}
	n := NewPagerDutyNotifier("key123", WithPagerDutyHTTPClient(client))
	if err := n.Send(context.Background(), nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(client.recorded) != 0 {
		t.Fatalf("expected no requests, got %d", len(client.recorded))
	}
}

func TestPagerDutyNotifier_SendWithEvents(t *testing.T) {
	client := &mockPDClient{statusCode: 202}
	n := NewPagerDutyNotifier("key-abc", WithPagerDutyHTTPClient(client))
	events := []Event{makePDEvent(EventAdded, 8080), makePDEvent(EventRemoved, 443)}
	if err := n.Send(context.Background(), events); err != nil {
("unexpected error: %v", err)
	}
	if len(client.recorded) != 1.Fatalf("expected 1 request, got %d", len(client.recorded))
	}
	got := client.recorded[0]
	if diff := cmp.Diff("key-abc", got.RoutingKey); diff != "" {
		t.Errorf("routing key mismatch (-want +got):\n%s", diff)
	}
	if got.EventAction != "trigger" {
		t.Errorf("expected event_action=trigger, got %s", got.EventAction)
	}
	if got.Payload.Source != "portwatch" {
		t.Errorf("expected source=portwatch, got %s", got.Payload.Source)
	}
	if got.Payload.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestPagerDutyNotifier_NonOKStatus(t *testing.T) {
	client := &mockPDClient{statusCode: 400}
	n := NewPagerDutyNotifier("key", WithPagerDutyHTTPClient(client))
	err := n.Send(context.Background(), []Event{makePDEvent(EventAdded, 9090)})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestPagerDutyNotifier_ImplementsNotifier(t *testing.T) {
	var _ Notifier = NewPagerDutyNotifier("key")
}
