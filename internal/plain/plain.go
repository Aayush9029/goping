package plain

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/Aayush9029/goping/internal/config"
	"github.com/Aayush9029/goping/internal/ping"
	"github.com/Aayush9029/goping/internal/stats"
)

func Run(ctx context.Context, cfg config.Config, stream <-chan ping.Event, out io.Writer) error {
	trackers := make(map[string]*stats.Tracker, len(cfg.Targets))
	for _, target := range cfg.Targets {
		trackers[target] = stats.NewTracker(target, cfg.Buffer)
	}

	palette := newPalette(cfg.ColorEnabled())
	fmt.Fprintf(out, "%s %s every %s\n", palette.Title.Render("goping"), strings.Join(cfg.Targets, ", "), cfg.Interval)

	for {
		select {
		case <-ctx.Done():
			printSummary(out, palette, cfg.Targets, trackers)
			return nil
		case event, ok := <-stream:
			if !ok {
				printSummary(out, palette, cfg.Targets, trackers)
				return nil
			}

			tracker := trackers[event.Target]
			if tracker == nil {
				tracker = stats.NewTracker(event.Target, cfg.Buffer)
				trackers[event.Target] = tracker
			}
			tracker.Apply(event)
			printEvent(out, palette, tracker, event)
		}
	}
}

type palette struct {
	Title lipgloss.Style
	Label lipgloss.Style
	Good  lipgloss.Style
	Warn  lipgloss.Style
	Bad   lipgloss.Style
	Dim   lipgloss.Style
}

func newPalette(enabled bool) palette {
	if !enabled {
		return palette{
			Title: lipgloss.NewStyle().Bold(true),
			Label: lipgloss.NewStyle().Bold(true),
			Good:  lipgloss.NewStyle(),
			Warn:  lipgloss.NewStyle(),
			Bad:   lipgloss.NewStyle(),
			Dim:   lipgloss.NewStyle(),
		}
	}
	return palette{
		Title: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7DD3FC")),
		Label: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E5E7EB")),
		Good:  lipgloss.NewStyle().Foreground(lipgloss.Color("#4ADE80")),
		Warn:  lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24")),
		Bad:   lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")),
		Dim:   lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")),
	}
}

func printEvent(out io.Writer, palette palette, tracker *stats.Tracker, event ping.Event) {
	host := palette.Label.Render("[" + tracker.Target + "]")
	timestamp := palette.Dim.Render(time.Now().Format("15:04:05"))

	switch event.Kind {
	case ping.EventReply:
		fmt.Fprintf(
			out,
			"%s %s %s from %s  seq=%d  time=%s  avg=%s  best=%s  worst=%s  jitter=%s  loss=%.1f%%\n",
			timestamp,
			host,
			palette.Good.Render(fmt.Sprintf("%d bytes", event.Bytes)),
			nonEmpty(tracker.Addr, event.Addr, tracker.Target),
			event.Seq,
			palette.Good.Render(stats.FormatDuration(event.RTT)),
			stats.FormatDuration(tracker.Avg()),
			stats.FormatDuration(tracker.Min),
			stats.FormatDuration(tracker.Max),
			stats.FormatDuration(tracker.Jitter()),
			tracker.LossPct(),
		)
	case ping.EventTimeout:
		fmt.Fprintf(
			out,
			"%s %s %s seq=%d  loss=%.1f%%  %s\n",
			timestamp,
			host,
			palette.Bad.Render("timeout"),
			event.Seq,
			tracker.LossPct(),
			palette.Dim.Render(strings.TrimSpace(event.Line)),
		)
	case ping.EventError:
		fmt.Fprintf(out, "%s %s %s\n", timestamp, host, palette.Bad.Render(strings.TrimSpace(event.Line)))
	}
}

func printSummary(out io.Writer, palette palette, targets []string, trackers map[string]*stats.Tracker) {
	for _, target := range targets {
		tracker := trackers[target]
		if tracker == nil {
			continue
		}
		fmt.Fprintf(
			out,
			"%s sent=%d recv=%d loss=%.1f%% avg=%s best=%s worst=%s jitter=%s\n",
			palette.Title.Render("--- "+tracker.Target+" ---"),
			tracker.Sent,
			tracker.Received,
			tracker.LossPct(),
			stats.FormatDuration(tracker.Avg()),
			stats.FormatDuration(tracker.Min),
			stats.FormatDuration(tracker.Max),
			stats.FormatDuration(tracker.Jitter()),
		)
	}
}

func nonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return "-"
}
