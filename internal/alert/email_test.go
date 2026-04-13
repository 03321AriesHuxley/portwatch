package alert

import (
	"context"
	"net/smtp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeEmailEvent(kind Kind, port uint16) Event {
	return Event{
		Kind:  kind,
		Entry: makeFormatterEntry(port),
	}
}

func TestEmailNotifier_SendEmpty(t *testing.T) {
	called := false
	n := NewEmailNotifier(EmailConfig{
		Host: "localhost", Port: 25,
		From: "a@example.com", To: []string{"b@example.com"},
	}, WithEmailSendFunc(func(_ string, _ smtp.Auth, _ string, _ []string, _ []byte) error {
		called = true
		return nil
	}))

	err := n.Send(context.Background(), nil)
	require.NoError(t, err)
	assert.False(t, called, "send should not be called for empty events")
}

func TestEmailNotifier_SendWithEvents(t *testing.T) {
	var capturedAddr string
	var capturedTo []string
	var capturedMsg []byte

	n := NewEmailNotifier(EmailConfig{
		Host: "smtp.example.com", Port: 587,
		Username: "user", Password: "pass",
		From: "alerts@example.com",
		To:   []string{"ops@example.com", "admin@example.com"},
	}, WithEmailSendFunc(func(addr string, _ smtp.Auth, _ string, to []string, msg []byte) error {
		capturedAddr = addr
		capturedTo = to
		capturedMsg = msg
		return nil
	}))

	events := []Event{
		makeEmailEvent(KindAdded, 8080),
		makeEmailEvent(KindRemoved, 9090),
	}

	err := n.Send(context.Background(), events)
	require.NoError(t, err)

	assert.Equal(t, "smtp.example.com:587", capturedAddr)
	assert.Equal(t, []string{"ops@example.com", "admin@example.com"}, capturedTo)

	msgStr := string(capturedMsg)
	assert.True(t, strings.Contains(msgStr, "Subject: portwatch: 2 port change(s) detected"))
	assert.True(t, strings.Contains(msgStr, "From: alerts@example.com"))
	assert.True(t, strings.Contains(msgStr, "To: ops@example.com, admin@example.com"))
}

func TestEmailNotifier_ImplementsNotifier(t *testing.T) {
	var _ Notifier = NewEmailNotifier(EmailConfig{})
}
