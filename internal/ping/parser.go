package ping

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	replyPattern   = regexp.MustCompile(`(?i)^(\d+)\s+bytes\s+from\s+([^: ]+).*icmp_(?:seq|req)=(\d+).*time[=<]?\s*([0-9.]+)\s*ms`)
	timeoutPattern = regexp.MustCompile(`(?i)(?:request timeout for icmp_seq|no answer yet for icmp_seq=|icmp_seq[=\s])\s*(\d+)(?:.*(?:unreachable|timeout))?`)
	headerPattern  = regexp.MustCompile(`(?i)^PING\s+(.+?)\s+\(([^)]+)\)`)
)

func ParseLine(target string, line string) (Event, bool) {
	now := time.Now()

	if match := headerPattern.FindStringSubmatch(line); len(match) == 3 {
		return Event{
			Target: target,
			Addr:   strings.TrimSpace(match[2]),
			Kind:   EventResolved,
			Line:   line,
			At:     now,
		}, true
	}

	if match := replyPattern.FindStringSubmatch(line); len(match) == 5 {
		bytes, _ := strconv.Atoi(match[1])
		seq, _ := strconv.Atoi(match[3])
		ms, _ := strconv.ParseFloat(match[4], 64)
		return Event{
			Target: target,
			Addr:   strings.TrimSpace(match[2]),
			Bytes:  bytes,
			Seq:    seq,
			RTT:    time.Duration(ms * float64(time.Millisecond)),
			Kind:   EventReply,
			Line:   line,
			At:     now,
		}, true
	}

	if strings.Contains(strings.ToLower(line), "request timeout") ||
		strings.Contains(strings.ToLower(line), "no answer yet") ||
		strings.Contains(strings.ToLower(line), "unreachable") {
		seq := parseSeq(line)
		return Event{
			Target: target,
			Seq:    seq,
			Kind:   EventTimeout,
			Line:   line,
			At:     now,
		}, true
	}

	if strings.Contains(strings.ToLower(line), "unknown host") ||
		strings.Contains(strings.ToLower(line), "cannot resolve") {
		return Event{
			Target: target,
			Kind:   EventError,
			Line:   line,
			At:     now,
		}, true
	}

	return Event{}, false
}

func parseSeq(line string) int {
	match := timeoutPattern.FindStringSubmatch(line)
	if len(match) < 2 {
		return -1
	}
	seq, err := strconv.Atoi(match[1])
	if err != nil {
		return -1
	}
	return seq
}
