package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestPlainModeE2E(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--plain", "--duration", "350ms", "test.local")
	cmd.Env = append(os.Environ(), "GOPING_PING_BIN="+repoRoot()+"/testdata/fake_ping.sh")
	cmd.Dir = repoRoot()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run failed: %v\n%s", err, string(output))
	}

	text := string(output)
	for _, needle := range []string{
		"goping test.local every 1s",
		"[test.local]",
		"time=12.3ms",
		"timeout",
		"sent=",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("expected output to contain %q\n%s", needle, text)
		}
	}
}

func repoRoot() string {
	return "/Users/yush/Developer/oss/public/brew-tools/goping"
}
