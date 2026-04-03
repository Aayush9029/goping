package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

var ErrHelp = errors.New("help requested")

type Config struct {
	Targets     []string
	Interval    time.Duration
	Timeout     time.Duration
	Buffer      int
	Count       int
	Duration    time.Duration
	Plain       bool
	NoColor     bool
	ForceColor  bool
	IPv4        bool
	IPv6        bool
	Version string
	IsTTY       bool
}

func Parse(args []string, version string, stdout *os.File) (Config, error) {
	cfg := Config{
		Interval:    time.Second,
		Timeout:     3 * time.Second,
		Buffer:      60,
		Version: version,
		IsTTY:       term.IsTerminal(int(stdout.Fd())),
	}

	fs := flag.NewFlagSet("goping", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.DurationVar(&cfg.Interval, "interval", cfg.Interval, "probe interval")
	fs.DurationVar(&cfg.Interval, "i", cfg.Interval, "probe interval")
	fs.DurationVar(&cfg.Timeout, "timeout", cfg.Timeout, "probe timeout")
	fs.IntVar(&cfg.Buffer, "buffer", cfg.Buffer, "history size in samples")
	fs.IntVar(&cfg.Buffer, "b", cfg.Buffer, "history size in samples")
	fs.IntVar(&cfg.Count, "count", 0, "stop after this many probes per target")
	fs.IntVar(&cfg.Count, "c", 0, "stop after this many probes per target")
	fs.DurationVar(&cfg.Duration, "duration", 0, "stop after this amount of time")
	fs.BoolVar(&cfg.Plain, "plain", false, "disable the TUI and print streaming lines")
	fs.BoolVar(&cfg.NoColor, "no-color", false, "disable ANSI color")
	fs.BoolVar(&cfg.ForceColor, "color", false, "force ANSI color")
	fs.BoolVar(&cfg.IPv4, "4", false, "resolve targets as IPv4")
	fs.BoolVar(&cfg.IPv6, "6", false, "resolve targets as IPv6")
	var showHelp bool
	var showVersion bool
	fs.BoolVar(&showHelp, "help", false, "show help")
	fs.BoolVar(&showHelp, "h", false, "show help")
	fs.BoolVar(&showVersion, "version", false, "show version")
	fs.BoolVar(&showVersion, "v", false, "show version")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if showHelp {
		printUsage(stdout, version)
		return Config{}, ErrHelp
	}
	if showVersion {
		fmt.Fprintf(stdout, "goping %s\n", version)
		return Config{}, ErrHelp
	}
	if cfg.IPv4 && cfg.IPv6 {
		return Config{}, errors.New("choose either -4 or -6, not both")
	}
	if cfg.Buffer < 10 {
		return Config{}, errors.New("buffer must be at least 10")
	}
	if cfg.Interval <= 0 {
		return Config{}, errors.New("interval must be positive")
	}
	if cfg.Timeout <= 0 {
		return Config{}, errors.New("timeout must be positive")
	}
	if cfg.Count < 0 {
		return Config{}, errors.New("count must be non-negative")
	}

	cfg.Targets = normalizeTargets(fs.Args())
	if len(cfg.Targets) == 0 {
		printUsage(stdout, version)
		return Config{}, errors.New("provide at least one host or IP")
	}
	return cfg, nil
}

func (c Config) ColorEnabled() bool {
	if c.NoColor {
		return false
	}
	return c.ForceColor || c.IsTTY
}

func (c Config) UseTUI() bool {
	return !c.Plain && c.IsTTY
}

func normalizeTargets(args []string) []string {
	var targets []string
	for _, arg := range args {
		trimmed := strings.TrimSpace(arg)
		if trimmed == "" {
			continue
		}
		targets = append(targets, trimmed)
	}
	return targets
}

func printUsage(w io.Writer, version string) {
	fmt.Fprintf(w, "goping %s\n\n", version)
	fmt.Fprintln(w, "Thoughtful ping with a live TUI and a script-friendly plain mode.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  goping [options] <host> [host...]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  goping 1.1.1.1")
	fmt.Fprintln(w, "  goping github.com cloudflare.com")
	fmt.Fprintln(w, "  goping --plain --count 5 1.1.1.1")
	fmt.Fprintln(w, "  goping --interval 500ms --buffer 90 8.8.8.8")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  -i, --interval <duration>    Probe interval (default: 1s)")
	fmt.Fprintln(w, "      --timeout <duration>     Probe timeout budget (default: 3s)")
	fmt.Fprintln(w, "  -b, --buffer <samples>       Samples to retain in memory (default: 60)")
	fmt.Fprintln(w, "  -c, --count <n>              Stop after n probes per target")
	fmt.Fprintln(w, "      --duration <duration>    Stop after this amount of time")
	fmt.Fprintln(w, "      --plain                  Disable the TUI")
	fmt.Fprintln(w, "      --color                  Force ANSI color")
	fmt.Fprintln(w, "      --no-color               Disable ANSI color")
	fmt.Fprintln(w, "  -4                           Resolve targets as IPv4")
	fmt.Fprintln(w, "  -6                           Resolve targets as IPv6")
	fmt.Fprintln(w, "  -v, --version                Show version")
	fmt.Fprintln(w, "  -h, --help                   Show help")
}
