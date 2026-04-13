package alert

import (
	"context"
	"fmt"
	"log/syslog"
	"strings"
)

// SyslogNotifier sends alert events to the local syslog daemon.
type SyslogNotifier struct {
	writer   *syslog.Writer
	priority syslog.Priority
	tag      string
}

// SyslogOption configures a SyslogNotifier.
type SyslogOption func(*SyslogNotifier)

// WithSyslogPriority sets the syslog priority level (default: LOG_INFO).
func WithSyslogPriority(p syslog.Priority) SyslogOption {
	return func(s *SyslogNotifier) {
		s.priority = p
	}
}

// WithSyslogTag sets the syslog tag string (default: "portwatch").
func WithSyslogTag(tag string) SyslogOption {
	return func(s *SyslogNotifier) {
		s.tag = tag
	}
}

// NewSyslogNotifier creates a SyslogNotifier connected to the local syslog.
func NewSyslogNotifier(opts ...SyslogOption) (*SyslogNotifier, error) {
	sn := &SyslogNotifier{
		priority: syslog.LOG_INFO | syslog.LOG_DAEMON,
		tag:      "portwatch",
	}
	for _, o := range opts {
		o(sn)
	}
	w, err := syslog.New(sn.priority, sn.tag)
	if err != nil {
		return nil, fmt.Errorf("syslog: dial: %w", err)
	}
	sn.writer = w
	return sn, nil
}

// Send writes each event as a syslog message.
func (s *SyslogNotifier) Send(_ context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	lines := make([]string, 0, len(events))
	for _, e := range events {
		lines = append(lines, FormatEvent(e))
	}
	msg := strings.Join(lines, "; ")
	if err := s.writer.Info(msg); err != nil {
		return fmt.Errorf("syslog: write: %w", err)
	}
	return nil
}

// Close releases the underlying syslog connection.
func (s *SyslogNotifier) Close() error {
	return s.writer.Close()
}
