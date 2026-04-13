package alert

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeSlackEvent(kind Kind, port uint16) Event {
	return Event{
		Kind:  kind,
		Entry: makeFormatterEntry(port),
	}
}

func TestSlackNotifier_SendEmpty(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()

	n := NewSlackNotifier(ts.URL)
	err := n.Send(context.Background(), []Event{})
	require.NoError(t, err)
	assert.False(t, called, "expected no HTTP call for empty events")
}

func TestSlackNotifier_SendWithEvents(t *testing.T) {
	var gotBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		gotBody = buf
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	events := []Event{
		makeSlackEvent(KindAdded, 8080),
		makeSlackEvent(KindRemoved, 9090),
	}

	n := NewSlackNotifier(ts.URL)
	err := n.Send(context.Background(), events)
	require.NoError(t, err)
	assert.Contains(t, string(gotBody), "8080")
	assert.Contains(t, string(gotBody), "9090")
}

func TestSlackNotifier_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	events := []Event{makeSlackEvent(KindAdded, 443)}
	n := NewSlackNotifier(ts.URL)
	err := n.Send(context.Background(), events)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestSlackNotifier_InvalidURL(t *testing.T) {
	n := NewSlackNotifier("http://127.0.0.1:0/no-server")
	events := []Event{makeSlackEvent(KindAdded, 22)}
	err := n.Send(context.Background(), events)
	require.Error(t, err)
}

func TestSlackNotifier_WithCustomClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	customClient := &http.Client{}
	n := NewSlackNotifier(ts.URL, WithSlackHTTPClient(customClient))
	assert.Equal(t, customClient, n.client)

	err := n.Send(context.Background(), []Event{makeSlackEvent(KindAdded, 3000)})
	require.NoError(t, err)
}
