package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Aayush9029/goping/internal/config"
	"github.com/Aayush9029/goping/internal/ping"
	"github.com/Aayush9029/goping/internal/stats"
)

type streamClosedMsg struct{}

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
	events   []ping.Event
	offset   int // lines scrolled up from bottom; 0 = follow latest
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
	return waitForEvent(m.stream)
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
		if msg.Kind == ping.EventReply || msg.Kind == ping.EventTimeout || msg.Kind == ping.EventError {
			m.events = append(m.events, msg)
			maxEvents := m.cfg.Buffer * max(len(m.order), 1)
			if maxEvents < 200 {
				maxEvents = 200
			}
			if len(m.events) > maxEvents {
				m.events = m.events[len(m.events)-maxEvents:]
			}
		}
		return m, waitForEvent(m.stream)
	case streamClosedMsg:
		m.closed = true
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.cancel()
			return m, tea.Quit
		case "up", "k":
			ceiling := len(m.events) - m.viewportHeight()
			if ceiling < 0 {
				ceiling = 0
			}
			if m.offset < ceiling {
				m.offset++
			}
			return m, nil
		case "down", "j":
			if m.offset > 0 {
				m.offset--
			}
			return m, nil
		case "G", "end":
			m.offset = 0
			return m, nil
		case "g", "home":
			ceiling := len(m.events) - m.viewportHeight()
			if ceiling < 0 {
				ceiling = 0
			}
			m.offset = ceiling
			return m, nil
		}
	}
	return m, nil
}

// viewportHeight returns lines available for ping output.
// Layout: header(1) + blank(1) + viewport(?) + blank(1) + stats(nTargets) + help(1)
func (m Model) viewportHeight() int {
	chrome := 4 + len(m.order)
	h := m.height - chrome
	if h < 1 {
		h = 1
	}
	return h
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
