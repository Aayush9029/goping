package stats

import (
	"testing"
	"time"

	"github.com/Aayush9029/goping/internal/ping"
)

func TestTrackerMetrics(t *testing.T) {
	tracker := NewTracker("example.com", 8)
	base := time.Now()

	tracker.Apply(ping.Event{Target: "example.com", Kind: ping.EventReply, Seq: 0, RTT: 10 * time.Millisecond, At: base})
	tracker.Apply(ping.Event{Target: "example.com", Kind: ping.EventReply, Seq: 1, RTT: 14 * time.Millisecond, At: base.Add(time.Second)})
	tracker.Apply(ping.Event{Target: "example.com", Kind: ping.EventTimeout, Seq: 2, Line: "Request timeout for icmp_seq 2", At: base.Add(2 * time.Second)})

	if tracker.Sent != 3 {
		t.Fatalf("expected sent=3, got %d", tracker.Sent)
	}
	if tracker.Received != 2 {
		t.Fatalf("expected received=2, got %d", tracker.Received)
	}
	if got := tracker.Avg(); got != 12*time.Millisecond {
		t.Fatalf("expected avg=12ms, got %s", got)
	}
	if got := tracker.Jitter(); got != 4*time.Millisecond {
		t.Fatalf("expected jitter=4ms, got %s", got)
	}
	if got := tracker.LossPct(); got < 33.3 || got > 33.4 {
		t.Fatalf("expected loss around 33.3, got %.2f", got)
	}
}
