package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

func TestHook_SendEmpty(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()

	h := alert.NewHook(ts.URL)
	err := h.Send(scanner.Diff{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if called {
		t.Fatal("expected webhook not to be called for empty diff")
	}
}

func TestHook_SendWithChanges(t *testing.T) {
	var received alert.Event

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	h := alert.NewHook(ts.URL)
	diff := scanner.Diff{
		Added:   []scanner.Socket{{LocalAddress: "0.0.0.0", LocalPort: 8080}},
		Removed: []scanner.Socket{},
	}

	if err := h.Send(diff); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(received.Added) != 1 {
		t.Errorf("expected 1 added socket, got %d", len(received.Added))
	}
	if received.Added[0].LocalPort != 8080 {
		t.Errorf("expected port 8080, got %d", received.Added[0].LocalPort)
	}
}

func TestHook_SendNonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	h := alert.NewHook(ts.URL)
	diff := scanner.Diff{
		Added: []scanner.Socket{{LocalAddress: "127.0.0.1", LocalPort: 9090}},
	}

	err := h.Send(diff)
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}
