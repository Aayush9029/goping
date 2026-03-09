package ping

import (
	"testing"
	"time"
)

func TestParseReplyLine(t *testing.T) {
	event, ok := ParseLine("example.com", "64 bytes from 1.1.1.1: icmp_seq=3 ttl=57 time=14.7 ms")
	if !ok {
		t.Fatal("expected line to parse")
	}
	if event.Kind != EventReply {
		t.Fatalf("expected reply event, got %v", event.Kind)
	}
	if event.Seq != 3 {
		t.Fatalf("expected seq 3, got %d", event.Seq)
	}
	if event.Addr != "1.1.1.1" {
		t.Fatalf("expected addr, got %q", event.Addr)
	}
	if event.RTT != 14700000*time.Nanosecond {
		t.Fatalf("unexpected rtt: %s", event.RTT)
	}
}

func TestParseTimeoutLine(t *testing.T) {
	event, ok := ParseLine("example.com", "Request timeout for icmp_seq 2")
	if !ok {
		t.Fatal("expected line to parse")
	}
	if event.Kind != EventTimeout {
		t.Fatalf("expected timeout event, got %v", event.Kind)
	}
	if event.Seq != 2 {
		t.Fatalf("expected seq 2, got %d", event.Seq)
	}
}
