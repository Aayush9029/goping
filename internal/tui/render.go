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

	header := titleStyle.Render(strings.ToUpper(tracker.Target))
	status := renderStatusBadge(tracker.Status(), color)
	meta := subtle.Render(tracker.Subtitle())
	statLine := subtle.Render(fmt.Sprintf(
		"now %s  avg %s  best %s  worst %s  jitter %s",
		stats.FormatDuration(tracker.Current),
		stats.FormatDuration(tracker.Avg()),
		stats.FormatDuration(tracker.Min),
		stats.FormatDuration(tracker.Max),
		stats.FormatDuration(tracker.Jitter()),
	))
	graph := renderGraph(tracker, width-4, graphHeight, color)

	body := strings.Join([]string{
		lipgloss.JoinHorizontal(lipgloss.Center, header, " ", status),
		meta,
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
	line := fmt.Sprintf("last reply %s  loss %.1f%%  timeouts %d", age, tracker.LossPct(), tracker.Timeouts)
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
	barStyle := lipgloss.NewStyle().Foreground(color)
	timeoutStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B"))

	samples := tracker.History
	if len(samples) > width {
		samples = samples[len(samples)-width:]
	}
	axisMax := tracker.AxisMax()
	axisMaxMs := float64(axisMax) / float64(time.Millisecond)

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
				sample := samples[col-(width-len(samples))]
				if sample.Timeout {
					if row == height-1 {
						cell = "×"
						cellStyle = timeoutStyle
					}
				} else if sample.RTT > 0 {
					normalized := float64(sample.RTT) / float64(time.Millisecond) / axisMaxMs
					barHeight := int(math.Round(normalized * float64(height)))
					if barHeight < 1 {
						barHeight = 1
					}
					if height-row <= barHeight {
						cell = "█"
						cellStyle = barStyle
					}
				}
			}
			line.WriteString(cellStyle.Render(cell))
		}
		rows = append(rows, line.String())
	}
	return strings.Join(rows, "\n")
}

func gridRune(row int, col int, height int) string {
	horizontal := row == height-1 || row == height/2
	vertical := col%5 == 0
	switch {
	case horizontal && vertical:
		return "┼"
	case horizontal:
		return "─"
	case vertical:
		return "┊"
	default:
		return " "
	}
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
