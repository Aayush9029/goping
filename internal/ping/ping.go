package ping

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Aayush9029/goping/internal/config"
)

type EventKind int

const (
	EventReply EventKind = iota
	EventTimeout
	EventError
	EventResolved
)

type Event struct {
	Target string
	Addr   string
	Bytes  int
	Seq    int
	RTT    time.Duration
	Kind   EventKind
	Line   string
	At     time.Time
}

type runner struct {
	cfg config.Config
}

func Start(parent context.Context, cfg config.Config) (<-chan Event, func() error, error) {
	ctx := parent
	var cancel context.CancelFunc = func() {}
	if cfg.Duration > 0 {
		ctx, cancel = context.WithTimeout(parent, cfg.Duration)
	}

	r := runner{cfg: cfg}
	events := make(chan Event, len(cfg.Targets)*8)
	var wg sync.WaitGroup

	for _, target := range cfg.Targets {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			r.runTarget(ctx, target, events)
		}(target)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(events)
		close(done)
	}()

	return events, func() error {
		<-done
		cancel()
		if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil
		}
		return nil
	}, nil
}

func (r runner) runTarget(ctx context.Context, target string, events chan<- Event) {
	cmd := exec.CommandContext(ctx, r.binary(), r.args(target)...)
	cmd.Env = append(os.Environ(), "LANG=C", "LC_ALL=C")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		events <- Event{Target: target, Kind: EventError, Line: err.Error(), At: time.Now()}
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		events <- Event{Target: target, Kind: EventError, Line: err.Error(), At: time.Now()}
		return
	}

	if err := cmd.Start(); err != nil {
		events <- Event{Target: target, Kind: EventError, Line: err.Error(), At: time.Now()}
		return
	}

	var scanWG sync.WaitGroup
	scanWG.Add(2)
	go func() {
		defer scanWG.Done()
		r.scan(ctx, target, stdout, events)
	}()
	go func() {
		defer scanWG.Done()
		r.scan(ctx, target, stderr, events)
	}()

	waitErr := normalizeWaitErr(cmd.Wait())
	scanWG.Wait()
	if ctx.Err() == nil && waitErr != nil {
		events <- Event{Target: target, Kind: EventError, Line: waitErr.Error(), At: time.Now()}
	}
}

func (r runner) scan(ctx context.Context, target string, reader io.Reader, events chan<- Event) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		event, ok := ParseLine(target, line)
		if !ok {
			continue
		}
		select {
		case events <- event:
		case <-ctx.Done():
			return
		}
	}
}

func (r runner) binary() string {
	if custom := strings.TrimSpace(os.Getenv("GOPING_PING_BIN")); custom != "" {
		return custom
	}
	return "ping"
}

func (r runner) args(target string) []string {
	args := []string{"-n", "-i", intervalString(r.cfg.Interval)}
	switch runtime.GOOS {
	case "linux":
		args = append(args, "-O")
	case "darwin":
		args = append(args, "-W", strconv.Itoa(int(r.cfg.Timeout/time.Millisecond)))
	}
	if r.cfg.Count > 0 {
		args = append(args, "-c", strconv.Itoa(r.cfg.Count))
	}
	if r.cfg.IPv4 {
		args = append(args, "-4")
	}
	if r.cfg.IPv6 {
		args = append(args, "-6")
	}
	args = append(args, target)
	return args
}

func intervalString(d time.Duration) string {
	if d < time.Millisecond {
		return "0.001"
	}
	seconds := float64(d) / float64(time.Second)
	return strconv.FormatFloat(seconds, 'f', 3, 64)
}

func normalizeWaitErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signal() == syscall.SIGKILL {
			return nil
		}
	}
	return fmt.Errorf("ping exited: %w", err)
}
