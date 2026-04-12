package config

import (
	"errors"
	"fmt"
	"time"
)

// ValidationError collects all constraint violations found in a Config.
type ValidationError struct {
	Errs []string
}

func (v *ValidationError) Error() string {
	msg := "config validation failed:"
	for _, e := range v.Errs {
		msg += "\n  - " + e
	}
	return msg
}

// Validate checks Config fields for invalid or unsafe values.
// It returns a *ValidationError listing every violation, or nil if valid.
func Validate(cfg Config) error {
	var errs []string

	if cfg.Interval < 1*time.Second {
		errs = append(errs, fmt.Sprintf("interval must be >= 1s, got %s", cfg.Interval))
	}
	if cfg.Interval > 24*time.Hour {
		errs = append(errs, fmt.Sprintf("interval must be <= 24h, got %s", cfg.Interval))
	}

	if cfg.AlertHook.URL != "" {
		if cfg.AlertHook.Timeout <= 0 {
			errs = append(errs, "alert_hook.timeout must be > 0 when url is set")
		}
	}

	if cfg.ProcNet.TCP == "" {
		errs = append(errs, "proc_net.tcp path must not be empty")
	}
	if cfg.ProcNet.UDP == "" {
		errs = append(errs, "proc_net.udp path must not be empty")
	}

	if len(errs) > 0 {
		return &ValidationError{Errs: errs}
	}
	return nil
}

// IsValidationError reports whether err is a *ValidationError.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
