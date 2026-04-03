package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Aayush9029/goping/internal/ping"
	"github.com/Aayush9029/goping/internal/stats"
)

var targetColors = []lipgloss.Color{
	"#38BDF8",
	"#4ADE80",
	"#F59E0B",
	"#F472B6",
	"#A78BFA",
	"#22D3EE",
}

var (
	dimStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B"))
	textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#CBD5E1"))
	goodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ADE80"))
	warnStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FBBF24"))
	badStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F87171"))
	boldStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E2E8F0"))
)

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	return m.renderHeader() + "\n\n" +
		m.renderViewport() + "\n\n" +
		m.renderStats() + "\n" +
		m.renderHelp()
}

func (m Model) renderHeader() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7DD3FC"))
	interval := dimStyle.Render("every " + m.cfg.Interval.String())

	if len(m.order) == 1 {
		target := m.order[0]
		tracker := m.trackers[target]
		addr := ""
		if tracker != nil && tracker.Addr != "" && tracker.Addr != target {
			addr = " (" + tracker.Addr + ")"
		}
		return title.Render("goping") + " " + target + addr + " " + interval
	}

	return title.Render("goping") + " " + strings.Join(m.order, ", ") + " " + interval
}

func (m Model) renderViewport() string {
	vpH := m.viewportHeight()

	if len(m.events) == 0 {
		lines := make([]string, vpH)
		lines[0] = dimStyle.Render("waiting for reply...")
		return strings.Join(lines, "\n")
	}

	multi := len(m.order) > 1
	colorMap := m.targetColorMap()
	padW := 0
	if multi {
		padW = m.maxTargetLen()
	}

	end := len(m.events) - m.offset
	if end < 0 {
		end = 0
	}
	start := end - vpH
	if start < 0 {
		start = 0
	}

	lines := make([]string, 0, vpH)
	for i := start; i < end; i++ {
		lines = append(lines, formatEvent(m.events[i], multi, colorMap, padW))
	}

	// Pad top with blank lines when there aren't enough events yet.
	for len(lines) < vpH {
		lines = append([]string{""}, lines...)
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderStats() string {
	colorMap := m.targetColorMap()
	order := stableOrder(m.order)
	lines := make([]string, 0, len(order))

	for _, target := range order {
		tracker := m.trackers[target]
		if tracker == nil {
			continue
		}

		color := colorMap[target]
		name := lipgloss.NewStyle().Bold(true).Foreground(color).Render(target)

		loss := tracker.LossPct()
		var lossVal string
		switch {
		case loss == 0:
			lossVal = goodStyle.Render(fmt.Sprintf("%.1f%%", loss))
		case loss < 10:
			lossVal = warnStyle.Render(fmt.Sprintf("%.1f%%", loss))
		default:
			lossVal = badStyle.Render(fmt.Sprintf("%.1f%%", loss))
		}

		recv := fmt.Sprintf("%d/%d", tracker.Received, tracker.Sent)
		parts := []string{
			dimStyle.Render("──") + " " + name,
			dimStyle.Render("avg ") + boldStyle.Render(stats.FormatDuration(tracker.Avg())),
			dimStyle.Render("min ") + boldStyle.Render(stats.FormatDuration(tracker.Min)),
			dimStyle.Render("max ") + boldStyle.Render(stats.FormatDuration(tracker.Max)),
			dimStyle.Render("jitter ") + boldStyle.Render(stats.FormatDuration(tracker.Jitter())),
			dimStyle.Render("loss ") + lossVal,
			dimStyle.Render("recv ") + boldStyle.Render(recv),
		}
		lines = append(lines, strings.Join(parts, "  "))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderHelp() string {
	help := dimStyle.Render("q quit  ↑↓ scroll")
	if m.offset > 0 {
		help += "  " + dimStyle.Render(fmt.Sprintf("↓ %d below", m.offset))
	}
	return help
}

func formatEvent(event ping.Event, multi bool, colorMap map[string]lipgloss.Color, padW int) string {
	var prefix string
	if multi {
		color := colorMap[event.Target]
		nameStyle := lipgloss.NewStyle().Foreground(color)
		prefix = nameStyle.Render(fmt.Sprintf("%-*s", padW, event.Target)) + "  "
	}

	switch event.Kind {
	case ping.EventReply:
		addr := event.Addr
		if addr == "" {
			addr = event.Target
		}
		return prefix +
			textStyle.Render(fmt.Sprintf("%d bytes from %s:", event.Bytes, addr)) +
			"  " + dimStyle.Render("seq=") + textStyle.Render(fmt.Sprintf("%d", event.Seq)) +
			"  " + dimStyle.Render("time=") + goodStyle.Render(stats.FormatDuration(event.RTT))

	case ping.EventTimeout:
		seq := ""
		if event.Seq >= 0 {
			seq = "  " + dimStyle.Render("seq=") + textStyle.Render(fmt.Sprintf("%d", event.Seq))
		}
		return prefix + badStyle.Render("timeout") + seq

	case ping.EventError:
		return prefix + badStyle.Render(strings.TrimSpace(event.Line))

	default:
		return ""
	}
}

func (m Model) targetColorMap() map[string]lipgloss.Color {
	order := stableOrder(m.order)
	colors := make(map[string]lipgloss.Color, len(order))
	for i, target := range order {
		colors[target] = targetColors[i%len(targetColors)]
	}
	return colors
}

func (m Model) maxTargetLen() int {
	n := 0
	for _, t := range m.order {
		if len(t) > n {
			n = len(t)
		}
	}
	return n
}

func stableOrder(items []string) []string {
	out := append([]string(nil), items...)
	sort.Strings(out)
	return out
}
