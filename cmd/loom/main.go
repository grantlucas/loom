package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/grantlucas/loom/internal/datasource"
	"github.com/grantlucas/loom/internal/tui"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "loom: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	cfg, err := ParseFlags(args, stdout)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}

	if cfg.Version {
		_, _ = fmt.Fprintf(stdout, "loom %s\n", version)
		return nil
	}

	exec := &datasource.BdExecutor{WorkDir: cfg.BeadsDir}
	retryExec := datasource.NewRetryExecutor(exec)
	client := datasource.NewClient(retryExec)
	cache := datasource.NewCache(client, cfg.Interval)

	app := tui.NewApp(cache, cfg.Interval, cfg.Watch)
	p := tea.NewProgram(app, tea.WithOutput(stderr), tea.WithAltScreen())
	_, err = p.Run()
	return err
}
