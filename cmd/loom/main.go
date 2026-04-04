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
		fmt.Fprintf(stdout, "loom %s\n", version)
		return nil
	}

	exec := &datasource.BdExecutor{WorkDir: cfg.BeadsDir}
	client := datasource.NewClient(exec)
	cache := datasource.NewCache(client, cfg.Interval)
	_ = cache // wired to views in a future issue

	app := tui.NewApp()
	p := tea.NewProgram(app, tea.WithOutput(stderr))
	_, err = p.Run()
	return err
}
