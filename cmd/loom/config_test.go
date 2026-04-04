package main

import (
	"testing"
	"time"
)

func TestParseFlags_Defaults(t *testing.T) {
	cfg, err := ParseFlags([]string{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Watch {
		t.Error("expected Watch=false")
	}
	if cfg.Interval != 5*time.Second {
		t.Errorf("expected Interval=5s, got %v", cfg.Interval)
	}
	if cfg.BeadsDir != "" {
		t.Errorf("expected BeadsDir=\"\", got %q", cfg.BeadsDir)
	}
	if cfg.Version {
		t.Error("expected Version=false")
	}
}
