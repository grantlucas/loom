package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "loom")
	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Dir = filepath.Join(projectRoot(t), "cmd", "loom")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return binary
}

func projectRoot(t *testing.T) string {
	t.Helper()
	// Walk up from this test file to find go.mod
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (go.mod)")
		}
		dir = parent
	}
}

func TestBinaryRunsAndExitsCleanly(t *testing.T) {
	binary := buildBinary(t)
	cmd := exec.Command(binary, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("binary exited with error: %v\n%s", err, out)
	}
}

func TestVersionFlag(t *testing.T) {
	binary := buildBinary(t)
	cmd := exec.Command(binary, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--version exited with error: %v\n%s", err, out)
	}
	output := string(out)
	if !strings.Contains(output, "loom") {
		t.Errorf("expected version output to contain 'loom', got: %q", output)
	}
	if !strings.Contains(output, "dev") {
		t.Errorf("expected version output to contain 'dev' (default version), got: %q", output)
	}
}
