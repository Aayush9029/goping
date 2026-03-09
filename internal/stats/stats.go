package stats

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Aayush9029/goping/internal/ping"
)

type Sample struct {
	RTT      time.Duration
	Timeout  bool
	Recorded time.Time
}

type Tracker struct {
	Target              string
	Addr                string
	Sent                int
	Received            int
	Timeouts            int
	Current             time.Duration
	Min                 time.Duration
	Max                 time.Duration
	LastReply           time.Time
	LastError           string
	Started             time.Time
	History             []Sample
	Capacity            int
	ConsecutiveTimeouts int
	LastEventAt         time.Time
	LastEventKind       ping.EventKind
	sumRTT              time.Duration
	prevRTT             time.Duration
	jitterSum           time.Duration
	jitterCount         int
}

func NewTracker(target string, capacity int) *Tracker {
	return &Tracker{
		Target:   target,
		Started:  time.Now(),
		Capacity: capacity,
	}
}

func (t *Tracker) Apply(event ping.Event) {
	if event.Target != "" {
		t.Target = event.Target
	}
	if event.Addr != "" {
		t.Addr = event.Addr
	}
	if event.Seq >= 0 && event.Seq+1 > t.Sent {
		t.Sent = event.Seq + 1
	}

	switch event.Kind {
	case ping.EventResolved:
		if event.Addr != "" {
			t.Addr = event.Addr
		}
	case ping.EventReply:
		t.LastEventAt = event.At
		t.LastEventKind = event.Kind
		t.Received++
		t.Current = event.RTT
		t.LastReply = event.At
		t.LastError = ""
		t.ConsecutiveTimeouts = 0
		t.sumRTT += event.RTT
		if t.Min == 0 || event.RTT < t.Min {
			t.Min = event.RTT
		}
		if event.RTT > t.Max {
			t.Max = event.RTT
		}
		if t.prevRTT > 0 {
			diff := t.prevRTT - event.RTT
			if diff < 0 {
				diff = -diff
			}
			t.jitterSum += diff
			t.jitterCount++
		}
		t.prevRTT = event.RTT
		t.appendSample(Sample{RTT: event.RTT, Recorded: event.At})
	case ping.EventTimeout:
		t.LastEventAt = event.At
		t.LastEventKind = event.Kind
		t.Timeouts++
		t.LastError = strings.TrimSpace(event.Line)
		t.ConsecutiveTimeouts++
		t.appendSample(Sample{Timeout: true, Recorded: event.At})
	case ping.EventError:
		t.LastEventAt = event.At
		t.LastEventKind = event.Kind
		t.LastError = strings.TrimSpace(event.Line)
	}
}

func (t *Tracker) appendSample(sample Sample) {
	t.History = append(t.History, sample)
	if len(t.History) > t.Capacity {
		t.History = t.History[len(t.History)-t.Capacity:]
	}
}

func (t *Tracker) Avg() time.Duration {
	if t.Received == 0 {
		return 0
	}
	return time.Duration(int64(t.sumRTT) / int64(t.Received))
}

func (t *Tracker) Jitter() time.Duration {
	if t.jitterCount == 0 {
		return 0
	}
	return time.Duration(int64(t.jitterSum) / int64(t.jitterCount))
}

func (t *Tracker) LossPct() float64 {
	if t.Sent == 0 {
		return 0
	}
	lost := t.Sent - t.Received
	if lost < 0 {
		lost = 0
	}
	return float64(lost) / float64(t.Sent) * 100
}

func (t *Tracker) Status() string {
	switch {
	case t.Received == 0 && t.Timeouts > 0:
		return "down"
	case t.ConsecutiveTimeouts >= 2:
		return "degraded"
	case t.Received > 0:
		return "online"
	default:
		return "warming"
	}
}

func (t *Tracker) Subtitle() string {
	addr := t.Target
	if t.Addr != "" {
		addr = fmt.Sprintf("%s  %s", t.Target, t.Addr)
	}
	return fmt.Sprintf("%s  sent %d  recv %d  loss %.1f%%", addr, t.Sent, t.Received, t.LossPct())
}

func (t *Tracker) AxisMax() time.Duration {
	maxValue := t.Max
	if maxValue == 0 {
		maxValue = 100 * time.Millisecond
	}
	niceMs := niceCeil(float64(maxValue) / float64(time.Millisecond))
	return time.Duration(niceMs * float64(time.Millisecond))
}

func niceCeil(value float64) float64 {
	if value <= 1 {
		return 1
	}
	power := math.Pow(10, math.Floor(math.Log10(value)))
	scaled := value / power
	switch {
	case scaled <= 1:
		return 1 * power
	case scaled <= 2:
		return 2 * power
	case scaled <= 5:
		return 5 * power
	default:
		return 10 * power
	}
}

func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return "-"
	}
	ms := float64(d) / float64(time.Millisecond)
	switch {
	case ms < 1:
		return fmt.Sprintf("%.2fms", ms)
	case ms < 100:
		return fmt.Sprintf("%.1fms", ms)
	default:
		return fmt.Sprintf("%.0fms", ms)
	}
}
