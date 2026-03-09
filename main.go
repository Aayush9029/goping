package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Aayush9029/goping/internal/config"
	"github.com/Aayush9029/goping/internal/ping"
	"github.com/Aayush9029/goping/internal/plain"
	"github.com/Aayush9029/goping/internal/tui"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := config.Parse(args, version, os.Stdout)
	if err != nil {
		if errors.Is(err, config.ErrHelp) {
			return nil
		}
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	stream, wait, err := ping.Start(ctx, cfg)
	if err != nil {
		return err
	}

	if cfg.UseTUI() {
		model := tui.NewModel(cfg, stream, stop)
		program := tea.NewProgram(
			model,
			tea.WithAltScreen(),
			tea.WithContext(ctx),
		)
		if _, err := program.Run(); err != nil {
			return err
		}
		return wait()
	}

	if err := plain.Run(ctx, cfg, stream, os.Stdout); err != nil {
		return err
	}
	return wait()
}
