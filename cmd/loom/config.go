package main

import (
	"flag"
	"io"
	"time"
)

// Config holds parsed CLI configuration.
type Config struct {
	Watch    bool
	Interval time.Duration
	BeadsDir string
	Version  bool
}

// ParseFlags parses command-line flags and environment variables into a Config.
func ParseFlags(args []string, environ []string) (Config, error) {
	fs := flag.NewFlagSet("loom", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var cfg Config
	fs.BoolVar(&cfg.Watch, "watch", false, "Start in watch mode (auto-refresh)")
	fs.DurationVar(&cfg.Interval, "interval", 5*time.Second, "Polling interval for watch mode")
	fs.StringVar(&cfg.BeadsDir, "beads-dir", "", "Path to .beads directory")
	fs.BoolVar(&cfg.Version, "version", false, "Print version and exit")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
