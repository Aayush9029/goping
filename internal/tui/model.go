package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Aayush9029/goping/internal/config"
	"github.com/Aayush9029/goping/internal/ping"
	"github.com/Aayush9029/goping/internal/stats"
)

type streamClosedMsg struct{}
type tickMsg time.Time

type Model struct {
	cfg      config.Config
	trackers map[string]*stats.Tracker
	order    []string
	stream   <-chan ping.Event
	cancel   context.CancelFunc
	width    int
	height   int
	started  time.Time
	closed   bool
}

func NewModel(cfg config.Config, stream <-chan ping.Event, cancel context.CancelFunc) Model {
	trackers := make(map[string]*stats.Tracker, len(cfg.Targets))
	for _, target := range cfg.Targets {
		trackers[target] = stats.NewTracker(target, cfg.Buffer)
	}
	return Model{
		cfg:      cfg,
		trackers: trackers,
		order:    append([]string(nil), cfg.Targets...),
		stream:   stream,
		cancel:   cancel,
		started:  time.Now(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(waitForEvent(m.stream), tick())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case ping.Event:
		tracker := m.trackers[msg.Target]
		if tracker == nil {
			tracker = stats.NewTracker(msg.Target, m.cfg.Buffer)
			m.trackers[msg.Target] = tracker
			m.order = append(m.order, msg.Target)
		}
		tracker.Apply(msg)
		return m, waitForEvent(m.stream)
	case streamClosedMsg:
		m.closed = true
		return m, tea.Quit
	case tickMsg:
		return m, tick()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.cancel()
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading goping..."
	}

	header := m.renderHeader()
	panels := m.renderPanels()
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Render("q quit  ctrl+c quit  plain mode: goping --plain host")

	content := []string{header, panels, footer}
	return strings.Join(content, "\n\n")
}

func waitForEvent(stream <-chan ping.Event) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-stream
		if !ok {
			return streamClosedMsg{}
		}
		return event
	}
}

func tick() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) renderHeader() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7DD3FC")).
		Render("goping")

	elapsed := time.Since(m.started).Round(time.Second)
	if elapsed < 0 {
		elapsed = 0
	}

	statusParts := []string{
		fmt.Sprintf("%d targets", len(m.order)),
		fmt.Sprintf("interval %s", m.cfg.Interval),
		fmt.Sprintf("buffer %d", m.cfg.Buffer),
		fmt.Sprintf("elapsed %s", elapsed),
	}
	if m.cfg.Duration > 0 {
		remaining := m.cfg.Duration - time.Since(m.started)
		if remaining < 0 {
			remaining = 0
		}
		statusParts = append(statusParts, fmt.Sprintf("remaining %s", remaining.Round(time.Second)))
	}

	line := lipgloss.NewStyle().Foreground(lipgloss.Color("#CBD5E1")).Render(strings.Join(statusParts, "  •  "))
	return title + "\n" + line
}

func (m Model) renderPanels() string {
	if len(m.order) == 0 {
		return ""
	}

	width := max(44, m.width-4)
	cols := 1
	if width >= 120 && len(m.order) > 1 {
		cols = 2
	}
	panelWidth := (width - ((cols - 1) * 2)) / cols

	var panels []string
	for _, target := range stableOrder(m.order) {
		tracker := m.trackers[target]
		if tracker == nil {
			continue
		}
		panels = append(panels, renderPanel(*tracker, panelWidth, m.cfg.GraphHeight))
	}
	return lipgloss.JoinVertical(lipgloss.Left, chunkPanels(panels, cols)...)
}

func chunkPanels(panels []string, cols int) []string {
	var rows []string
	for i := 0; i < len(panels); i += cols {
		end := i + cols
		if end > len(panels) {
			end = len(panels)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, panels[i:end]...))
	}
	return rows
}

func stableOrder(items []string) []string {
	out := append([]string(nil), items...)
	sort.Strings(out)
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
