package alert_test

import (
	"context"
	"log/syslog"
	"testing"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

func makeSyslogEntry(proto, addr string, port uint16) scanner.Entry {
	return scanner.Entry{Protocol: proto, LocalAddr: addr, LocalPort: port}
}

// TestSyslogNotifier_New verifies that a SyslogNotifier can be constructed.
// On systems without syslog this will be skipped.
func TestSyslogNotifier_New(t *testing.T) {
	sn, err := alert.NewSyslogNotifier(
		alert.WithSyslogTag("portwatch-test"),
		alert.WithSyslogPriority(syslog.LOG_INFO|syslog.LOG_LOCAL0),
	)
	if err != nil {
		t.Skipf("syslog unavailable: %v", err)
	}
	t.Cleanup(func() { _ = sn.Close() })
}

// TestSyslogNotifier_SendEmpty ensures no error is returned for empty events.
func TestSyslogNotifier_SendEmpty(t *testing.T) {
	sn, err := alert.NewSyslogNotifier(alert.WithSyslogTag("portwatch-test"))
	if err != nil {
		t.Skipf("syslog unavailable: %v", err)
	}
	t.Cleanup(func() { _ = sn.Close() })

	if err := sn.Send(context.Background(), nil); err != nil {
		t.Fatalf("expected no error on empty send, got: %v", err)
	}
}

// TestSyslogNotifier_SendWithEvents verifies events are written without error.
func TestSyslogNotifier_SendWithEvents(t *testing.T) {
	sn, err := alert.NewSyslogNotifier(alert.WithSyslogTag("portwatch-test"))
	if err != nil {
		t.Skipf("syslog unavailable: %v", err)
	}
	t.Cleanup(func() { _ = sn.Close() })

	events := []alert.Event{
		{Kind: alert.EventAdded, Entry: makeSyslogEntry("tcp", "0.0.0.0", 8080)},
		{Kind: alert.EventRemoved, Entry: makeSyslogEntry("tcp", "127.0.0.1", 9090)},
	}

	if err := sn.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestSyslogNotifier_ImplementsNotifier checks interface compliance.
func TestSyslogNotifier_ImplementsNotifier(t *testing.T) {
	sn, err := alert.NewSyslogNotifier(alert.WithSyslogTag("portwatch-test"))
	if err != nil {
		t.Skipf("syslog unavailable: %v", err)
	}
	t.Cleanup(func() { _ = sn.Close() })

	var _ interface {
		Send(context.Context, []alert.Event) error
	} = sn
}
