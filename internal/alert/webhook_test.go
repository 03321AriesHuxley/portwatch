package alert_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

func makeWebhookEvent(kind alert.EventKind, port uint16) alert.Event {
	return alert.Event{
		Kind: kind,
		Entry: scanner.Entry{
			Proto:        "tcp",
			LocalAddress: "0.0.0.0",
			LocalPort:    port,
			PID:          1234,
		},
	}
}

func TestWebhookNotifier_SendEmpty(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()

	n := alert.NewWebhookNotifier(ts.URL)
	if err := n.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("expected no HTTP call for empty events")
	}
}

func TestWebhookNotifier_SendWithEvents(t *testing.T) {
	var received map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	events := []alert.Event{
		makeWebhookEvent(alert.EventAdded, 8080),
		makeWebhookEvent(alert.EventRemoved, 443),
	}

	n := alert.NewWebhookNotifier(ts.URL)
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received == nil {
		t.Fatal("expected payload to be received")
	}
	evs, ok := received["events"].([]interface{})
	if !ok || len(evs) != 2 {
		t.Fatalf("expected 2 events in payload, got %v", received["events"])
	}
}

func TestWebhookNotifier_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := alert.NewWebhookNotifier(ts.URL)
	err := n.Send(context.Background(), []alert.Event{makeWebhookEvent(alert.EventAdded, 9090)})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestWebhookNotifier_InvalidURL(t *testing.T) {
	n := alert.NewWebhookNotifier("http://127.0.0.1:0/no-server")
	err := n.Send(context.Background(), []alert.Event{makeWebhookEvent(alert.EventAdded, 22)})
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}
