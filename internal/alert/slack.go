package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackNotifier sends alert events to a Slack incoming webhook.
type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

type slackPayload struct {
	Text string `json:"text"`
}

// SlackOption configures a SlackNotifier.
type SlackOption func(*SlackNotifier)

// WithSlackHTTPClient overrides the default HTTP client.
func WithSlackHTTPClient(c *http.Client) SlackOption {
	return func(s *SlackNotifier) {
		s.client = c
	}
}

// NewSlackNotifier creates a Notifier that posts events to a Slack webhook URL.
func NewSlackNotifier(webhookURL string, opts ...SlackOption) *SlackNotifier {
	s := &SlackNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Send posts a formatted summary of events to the Slack webhook.
func (s *SlackNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	text := FormatEvents(events)
	payload := slackPayload{Text: text}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("slack: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}
