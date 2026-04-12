package config_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
)

func TestValidate_Defaults(t *testing.T) {
	if err := config.Validate(config.Defaults()); err != nil {
		t.Errorf("defaults should be valid, got: %v", err)
	}
}

func TestValidate_IntervalTooShort(t *testing.T) {
	cfg := config.Defaults()
	cfg.Interval = 500 * time.Millisecond

	err := config.Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error for short interval")
	}
	if !config.IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestValidate_IntervalTooLong(t *testing.T) {
	cfg := config.Defaults()
	cfg.Interval = 25 * time.Hour

	if err := config.Validate(cfg); err == nil {
		t.Fatal("expected validation error for interval > 24h")
	}
}

func TestValidate_HookURLWithZeroTimeout(t *testing.T) {
	cfg := config.Defaults()
	cfg.AlertHook.URL = "http://example.com/hook"
	cfg.AlertHook.Timeout = 0

	if err := config.Validate(cfg); err == nil {
		t.Fatal("expected validation error for zero hook timeout")
	}
}

func TestValidate_EmptyProcNetPath(t *testing.T) {
	cfg := config.Defaults()
	cfg.ProcNet.TCP = ""

	err := config.Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error for empty TCP path")
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := config.Defaults()
	cfg.Interval = 0
	cfg.ProcNet.TCP = ""
	cfg.ProcNet.UDP = ""

	err := config.Validate(cfg)
	if err == nil {
		t.Fatal("expected multiple validation errors")
	}
	ve, ok := err.(*config.ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Errs) < 3 {
		t.Errorf("expected at least 3 errors, got %d: %v", len(ve.Errs), ve.Errs)
	}
}
