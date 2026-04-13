package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookHTTPClient is the interface used for sending webhook requests.
type WebhookHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// webhookNotifier sends port change events as JSON payloads to a generic webhook URL.
type webhookNotifier struct {
	url    string
	client WebhookHTTPClient
}

type webhookPayload struct {
	Timestamp string         `json:"timestamp"`
	Events    []webhookEvent `json:"events"`
}

type webhookEvent struct {
	Kind    string `json:"kind"`
	Proto   string `json:"proto"`
	Address string `json:"address"`
	Port    uint16 `json:"port"`
	PID     int    `json:"pid"`
}

// WithWebhookHTTPClient overrides the HTTP client used by the webhook notifier.
func WithWebhookHTTPClient(c WebhookHTTPClient) func(*webhookNotifier) {
	return func(w *webhookNotifier) {
		w.client = c
	}
}

// NewWebhookNotifier creates a Notifier that POSTs JSON event payloads to url.
func NewWebhookNotifier(url string, opts ...func(*webhookNotifier)) Notifier {
	wn := &webhookNotifier{
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
	for _, o := range opts {
		o(wn)
	}
	return wn
}

func (wn *webhookNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	we := make([]webhookEvent, len(events))
	for i, e := range events {
		we[i] = webhookEvent{
			Kind:    string(e.Kind),
			Proto:   e.Entry.Proto,
			Address: e.Entry.LocalAddress,
			Port:    e.Entry.LocalPort,
			PID:     e.Entry.PID,
		}
	}

	payload := webhookPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Events:    we,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshaltreq, err := http.NewRequestWithContext(ctx, http.MethodPost, wn.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := wn.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}
