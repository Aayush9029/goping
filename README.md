# goping

Thoughtful ping with a live TUI, better per-host stats, and a btop-style grid graph.

<img width="1800" height="1100" alt="CleanShot 2026-03-09 at 11 37 23@2x" src="https://github.com/user-attachments/assets/694ad1a3-7c5b-4340-980b-dbca1761533f" />


## Installation

```bash
brew install aayush9029/tap/goping
```

## Usage

```bash
goping 1.1.1.1
goping github.com cloudflare.com
goping --plain --count 5 8.8.8.8
goping --interval 500ms --buffer 90 1.1.1.1
goping --duration 10s --plain apple.com
```

## Options

| Option | Description |
| --- | --- |
| `-i`, `--interval <duration>` | Probe interval between ping requests |
| `--timeout <duration>` | Probe timeout budget |
| `-b`, `--buffer <samples>` | Samples to retain for graphing |
| `-c`, `--count <n>` | Stop after `n` probes per target |
| `--duration <duration>` | Stop after a fixed runtime |
| `--plain` | Disable the TUI and print streaming lines |
| `--graph-height <rows>` | Rows per graph in the TUI |
| `--color`, `--no-color` | Force or disable ANSI color |
| `-4`, `-6` | Prefer IPv4 or IPv6 resolution |

## How it works

1. `goping` launches the system `ping` command for each target so it works cleanly on macOS without raw-socket privileges.
2. Every reply or timeout is parsed into a shared stats model with live min, avg, max, jitter, loss, and recent history.
3. In the TUI, each target gets a bordered panel with a grid graph, health badge, and compact summary line.
4. In `--plain` mode, probes stream as formatted lines so the tool still feels familiar in scripts and quick shell use.

## Requirements

- macOS with the built-in `ping` command
- A terminal that supports ANSI color for the full TUI experience

## License

MIT
