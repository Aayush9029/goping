package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/Aayush9029/goping/internal/stats"
)

var panelColors = []lipgloss.Color{
	lipgloss.Color("#38BDF8"),
	lipgloss.Color("#4ADE80"),
	lipgloss.Color("#F59E0B"),
	lipgloss.Color("#F472B6"),
	lipgloss.Color("#A78BFA"),
	lipgloss.Color("#22D3EE"),
}

func renderPanel(tracker stats.Tracker, width int, graphHeight int) string {
	color := panelColors[int(math.Abs(float64(hash(tracker.Target))))%len(panelColors)]
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(color)
	subtle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8"))

	header := titleStyle.Render(tracker.Target)
	status := renderStatusBadge(tracker.Status(), color)
	meta := subtle.Render(renderMeta(tracker))
	flow := renderFlow(tracker, width-4, color)
	statLine := renderStatLine(tracker)
	graph := renderGraph(tracker, width-4, graphHeight, color)

	body := strings.Join([]string{
		lipgloss.JoinHorizontal(lipgloss.Center, header, " ", status),
		meta,
		flow,
		statLine,
		graph,
		renderFooter(tracker),
	}, "\n")

	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Render(body)
}

func renderMeta(tracker stats.Tracker) string {
	if tracker.Addr == "" || tracker.Addr == tracker.Target {
		return fmt.Sprintf("target %s", tracker.Target)
	}
	return fmt.Sprintf("target %s  resolved %s", tracker.Target, tracker.Addr)
}

func renderStatLine(tracker stats.Tracker) string {
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B"))
	value := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E2E8F0"))
	parts := []string{
		label.Render("now") + " " + value.Render(stats.FormatDuration(tracker.Current)),
		label.Render("avg") + " " + value.Render(stats.FormatDuration(tracker.Avg())),
		label.Render("best") + " " + value.Render(stats.FormatDuration(tracker.Min)),
		label.Render("worst") + " " + value.Render(stats.FormatDuration(tracker.Max)),
		label.Render("jitter") + " " + value.Render(stats.FormatDuration(tracker.Jitter())),
	}
	return strings.Join(parts, "  ")
}

func renderStatusBadge(status string, color lipgloss.Color) string {
	background := lipgloss.Color("#1E293B")
	switch status {
	case "online":
		background = lipgloss.Color("#052E1A")
	case "degraded":
		background = lipgloss.Color("#3F2A04")
	case "down":
		background = lipgloss.Color("#3F1111")
	}
	return lipgloss.NewStyle().
		Foreground(color).
		Background(background).
		Padding(0, 1).
		Render(status)
}

func renderFooter(tracker stats.Tracker) string {
	subtle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8"))
	age := "-"
	if !tracker.LastReply.IsZero() {
		age = time.Since(tracker.LastReply).Round(time.Second).String() + " ago"
	}
	line := fmt.Sprintf("last reply %s  sent %d  recv %d  loss %.1f%%  timeouts %d", age, tracker.Sent, tracker.Received, tracker.LossPct(), tracker.Timeouts)
	if tracker.LastError != "" {
		line = fmt.Sprintf("%s  note %s", line, tracker.LastError)
	}
	return subtle.Render(truncate(line, 80))
}

func renderGraph(tracker stats.Tracker, width int, height int, color lipgloss.Color) string {
	if width < 16 {
		width = 16
	}
	if height < 4 {
		height = 4
	}

	gridStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#334155"))
	fillStyle := lipgloss.NewStyle().Foreground(color)
	pointStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E0F2FE"))
	lastPointStyle := lipgloss.NewStyle().Bold(true).Foreground(color)
	timeoutStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B"))

	samples := tracker.History
	if len(samples) > width {
		samples = samples[len(samples)-width:]
	}
	axisMax := tracker.AxisMax()
	axisMaxMs := float64(axisMax) / float64(time.Millisecond)
	levels := sampleLevels(samples, height, axisMaxMs)

	rows := make([]string, 0, height)
	for row := 0; row < height; row++ {
		var label string
		switch row {
		case 0:
			label = fmt.Sprintf("%6s", stats.FormatDuration(axisMax))
		case height / 2:
			label = fmt.Sprintf("%6s", stats.FormatDuration(axisMax/2))
		case height - 1:
			label = fmt.Sprintf("%6s", "0ms")
		default:
			label = "      "
		}

		var line strings.Builder
		line.WriteString(labelStyle.Render(label))
		line.WriteString(" ")
		for col := 0; col < width; col++ {
			cell := gridRune(row, col, height)
			cellStyle := gridStyle
			if col >= width-len(samples) {
				sampleIndex := col - (width - len(samples))
				sample := samples[sampleIndex]
				if sample.Timeout {
					if row == height-1 {
						cell = "×"
						cellStyle = timeoutStyle
					}
				} else if sample.RTT > 0 {
					level := levels[sampleIndex]
					switch {
					case row == level && sampleIndex == len(samples)-1:
						cell = "◉"
						cellStyle = lastPointStyle
					case row == level:
						cell = "●"
						cellStyle = pointStyle
					case row > level:
						cell = "▄"
						cellStyle = fillStyle
					}
				}
			}
			line.WriteString(cellStyle.Render(cell))
		}
		rows = append(rows, line.String())
	}
	return strings.Join(rows, "\n")
}

func sampleLevels(samples []stats.Sample, height int, axisMaxMs float64) []int {
	levels := make([]int, len(samples))
	for i, sample := range samples {
		if sample.Timeout || sample.RTT <= 0 {
			levels[i] = height - 1
			continue
		}
		normalized := float64(sample.RTT) / float64(time.Millisecond) / axisMaxMs
		if normalized < 0 {
			normalized = 0
		}
		if normalized > 1 {
			normalized = 1
		}
		level := height - 1 - int(math.Round(normalized*float64(height-1)))
		if level < 0 {
			level = 0
		}
		if level >= height {
			level = height - 1
		}
		levels[i] = level
	}
	return levels
}

func gridRune(row int, col int, height int) string {
	horizontal := row == height-1 || row == height/2 || row == 0
	vertical := col%5 == 0
	switch {
	case horizontal && vertical:
		return "┼"
	case horizontal:
		if row == 0 {
			return "╌"
		}
		return "─"
	case vertical:
		return "┊"
	default:
		return " "
	}
}

func renderFlow(tracker stats.Tracker, width int, color lipgloss.Color) string {
	if width < 24 {
		width = 24
	}

	subtle := lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B"))
	good := lipgloss.NewStyle().Bold(true).Foreground(color)
	bad := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F87171"))
	neutral := lipgloss.NewStyle().Foreground(lipgloss.Color("#CBD5E1"))

	txIcon := "▷"
	rxIcon := "○"
	rxStyle := neutral
	if time.Since(tracker.LastEventAt) < 600*time.Millisecond {
		switch tracker.LastEventKind {
		case 0:
			rxIcon = "◉"
			rxStyle = good
		case 1:
			rxIcon = "×"
			rxStyle = bad
		}
	}

	recent := recentPackets(tracker, 12, color)
	left := subtle.Render(fmt.Sprintf("tx %04d", tracker.Sent))
	right := subtle.Render(fmt.Sprintf("rx %04d", tracker.Received))
	rail := good.Render(txIcon) + subtle.Render("═══") + rxStyle.Render(rxIcon)
	line := lipgloss.JoinHorizontal(lipgloss.Center, left, "  ", rail, "  ", right)
	if recent != "" {
		line = lipgloss.JoinHorizontal(lipgloss.Center, line, "  ", subtle.Render("recent"), " ", recent)
	}
	return line
}

func recentPackets(tracker stats.Tracker, count int, color lipgloss.Color) string {
	if count <= 0 || len(tracker.History) == 0 {
		return ""
	}
	if len(tracker.History) < count {
		count = len(tracker.History)
	}

	good := lipgloss.NewStyle().Foreground(color)
	bad := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F87171"))
	last := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E0F2FE"))

	var out strings.Builder
	window := tracker.History[len(tracker.History)-count:]
	for i, sample := range window {
		switch {
		case sample.Timeout:
			out.WriteString(bad.Render("×"))
		case i == len(window)-1:
			out.WriteString(last.Render("◉"))
		default:
			out.WriteString(good.Render("•"))
		}
	}
	return out.String()
}

func hash(text string) int {
	value := 0
	for _, r := range text {
		value = value*31 + int(r)
	}
	return value
}

func truncate(text string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= width {
		return text
	}
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
}
