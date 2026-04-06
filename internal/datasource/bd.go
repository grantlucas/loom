package datasource

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Executor runs bd CLI commands and returns raw output.
type Executor interface {
	Execute(args ...string) ([]byte, error)
}

// BdExecutor runs the real bd binary.
type BdExecutor struct {
	BinPath string
	WorkDir string
}

// Execute runs a bd command with the given arguments.
func (e *BdExecutor) Execute(args ...string) ([]byte, error) {
	binPath := e.BinPath
	if binPath == "" {
		binPath = "bd"
	}
	cmd := exec.Command(binPath, args...)
	if e.WorkDir != "" {
		cmd.Dir = e.WorkDir
	}
	out, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) || isPathNotFound(binPath, err) {
			return nil, fmt.Errorf("%w: %w", ErrBdNotFound, err)
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "no beads database found") {
				return nil, fmt.Errorf("%w: %s", ErrProjectNotInitialized, strings.TrimSpace(stderr))
			}
			if strings.Contains(stderr, "exclusive lock") {
				return nil, fmt.Errorf("%w: %s", ErrDatabaseLocked, strings.TrimSpace(stderr))
			}
			return nil, &BdError{Args: args, Stderr: strings.TrimSpace(stderr), Err: err}
		}
		return nil, fmt.Errorf("bd %v: %w", args, err)
	}
	return out, nil
}

// isPathNotFound returns true when the error indicates the binary path
// does not exist on disk (absolute or relative path, not a PATH lookup).
func isPathNotFound(binPath string, err error) bool {
	var pathErr *os.PathError
	if errors.As(err, &pathErr) && os.IsNotExist(pathErr) {
		return true
	}
	return false
}

// Client provides a high-level API for fetching Beads data.
type Client struct {
	exec Executor
}

// NewClient creates a Client with the given executor.
func NewClient(exec Executor) *Client {
	return &Client{exec: exec}
}

// ListIssues runs bd list --all --json and returns parsed issues.
func (c *Client) ListIssues() ([]Issue, error) {
	data, err := c.exec.Execute("list", "--all", "--json")
	if err != nil {
		return nil, err
	}
	return ParseIssueList(data)
}

// GetIssue runs bd show <id> --json and returns parsed issue detail.
func (c *Client) GetIssue(id string) (*IssueDetail, error) {
	data, err := c.exec.Execute("show", id, "--json")
	if err != nil {
		return nil, err
	}
	return ParseIssueDetail(data)
}

// ListReady runs bd ready --json and returns parsed ready issues.
func (c *Client) ListReady() ([]Issue, error) {
	data, err := c.exec.Execute("ready", "--json")
	if err != nil {
		return nil, err
	}
	return ParseIssueList(data)
}
