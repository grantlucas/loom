package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

func TestParseFlags_Defaults(t *testing.T) {
	cfg, err := ParseFlags([]string{}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Watch {
		t.Error("expected Watch=false")
	}
	if cfg.Interval != 5*time.Second {
		t.Errorf("expected Interval=5s, got %v", cfg.Interval)
	}
	if cfg.BeadsDir != ".beads" {
		t.Errorf("expected BeadsDir=.beads, got %q", cfg.BeadsDir)
	}
	if cfg.Version {
		t.Error("expected Version=false")
	}
}

func TestParseFlags_Watch(t *testing.T) {
	cfg, err := ParseFlags([]string{"--watch"}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Watch {
		t.Error("expected Watch=true")
	}
}

func TestParseFlags_Interval(t *testing.T) {
	cfg, err := ParseFlags([]string{"--interval", "10s"}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 10*time.Second {
		t.Errorf("expected Interval=10s, got %v", cfg.Interval)
	}
}

func TestParseFlags_BeadsDir(t *testing.T) {
	cfg, err := ParseFlags([]string{"--beads-dir", "/tmp/test"}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BeadsDir != "/tmp/test" {
		t.Errorf("expected BeadsDir=/tmp/test, got %q", cfg.BeadsDir)
	}
}

func TestParseFlags_Version(t *testing.T) {
	cfg, err := ParseFlags([]string{"--version"}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Version {
		t.Error("expected Version=true")
	}
}

func TestParseFlags_InvalidFlag(t *testing.T) {
	_, err := ParseFlags([]string{"--bogus"}, io.Discard)
	if err == nil {
		t.Error("expected error for invalid flag")
	}
}

func TestRun_HelpFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--help"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("--help should not return error, got: %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "Usage") {
		t.Errorf("expected help output to contain 'Usage', got %q", output)
	}
}

func TestRun_VersionFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--version"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "loom") {
		t.Errorf("expected output to contain 'loom', got %q", output)
	}
}

func TestRun_InvalidFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--bogus"}, &stdout, &stderr)
	if err == nil {
		t.Error("expected error for invalid flag")
	}
}
