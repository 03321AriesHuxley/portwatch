package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const pagerDutyEventsURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutyClient is the HTTP interface used by NewPagerDutyNotifier.
type PagerDutyClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type pagerDutyNotifier struct {
	routingKey string
	client     PagerDutyClient
	eventsURL  string
}

type pdPayload struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	Payload     pdDetails `json:"payload"`
}

type pdDetails struct {
	Summary  string `json:"summary"`
	Source   string `json:"source"`
	Severity string `json:"severity"`
}

// WithPagerDutyHTTPClient overrides the HTTP client used for PagerDuty requests.
func WithPagerDutyHTTPClient(c PagerDutyClient) func(*pagerDutyNotifier) {
	return func(n *pagerDutyNotifier) { n.client = c }
}

// NewPagerDutyNotifier creates a Notifier that sends alerts to PagerDuty.
func NewPagerDutyNotifier(routingKey string, opts ...func(*pagerDutyNotifier)) Notifier {
	n := &pagerDutyNotifier{
		routingKey: routingKey,
		client:     &http.Client{Timeout: 10 * time.Second},
		eventsURL:  pagerDutyEventsURL,
	}
	for _, o := range opts {
		o(n)
	}
	return n
}

func (n *pagerDutyNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	summary := FormatSummary(events)
	body := pdPayload{
		RoutingKey:  n.routingKey,
		EventAction: "trigger",
		Payload: pdDetails{
			Summary:  summary,
			Source:   "portwatch",
			Severity: "warning",
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.eventsURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("pagerduty: send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}
