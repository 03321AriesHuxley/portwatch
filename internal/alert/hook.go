package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/portwatch/internal/scanner"
)

// Event represents a port change alert payload.
type Event struct {
	Timestamp time.Time        `json:"timestamp"`
	Added     []scanner.Socket `json:"added"`
	Removed   []scanner.Socket `json:"removed"`
}

// Hook sends alert events to a configured webhook URL.
type Hook struct {
	URL    string
	Client *http.Client
}

// NewHook creates a Hook with the given URL and a default HTTP client.
func NewHook(url string) *Hook {
	return &Hook{
		URL: url,
		Client: &http.Client{Timeout: 5 * time.Second},
	}
}

// Send posts the diff result as a JSON event to the webhook URL.
func (h *Hook) Send(diff scanner.Diff) error {
	if len(diff.Added) == 0 && len(diff.Removed) == 0 {
		return nil
	}

	event := Event{
		Timestamp: time.Now().UTC(),
		Added:     diff.Added,
		Removed:   diff.Removed,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("alert: marshal event: %w", err)
	}

	resp, err := h.Client.Post(h.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("alert: post to %s: %w", h.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("alert: unexpected status %d from %s", resp.StatusCode, h.URL)
	}

	return nil
}
