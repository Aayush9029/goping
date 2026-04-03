<p align="center">
  <img src="assets/icon.png" width="128" alt="goping">
  <h1 align="center">goping</h1>
  <p align="center">Ping with live stats — like <code>ping</code> but with a persistent stats bar.</p>
</p>

## Install

```bash
brew install aayush9029/tap/goping
```

Or tap first:

```bash
brew tap aayush9029/tap
brew install goping
```

## Usage

```bash
goping 1.1.1.1
goping github.com cloudflare.com
goping --plain --count 5 8.8.8.8
goping --interval 500ms 1.1.1.1
goping --duration 10s apple.com
```

## Options

| Option | Description |
| --- | --- |
| `-i`, `--interval` | Probe interval (default: 1s) |
| `--timeout` | Probe timeout (default: 3s) |
| `-b`, `--buffer` | Samples to retain (default: 60) |
| `-c`, `--count` | Stop after n probes per target |
| `--duration` | Stop after a fixed runtime |
| `--plain` | Disable TUI, print streaming lines |
| `--color` / `--no-color` | Force or disable ANSI color |
| `-4` / `-6` | IPv4 or IPv6 resolution |

## How it works

1. Launches the system `ping` for each target — no raw sockets needed.
2. Parses replies and timeouts into a live stats model (min, avg, max, jitter, loss).
3. TUI shows scrolling ping lines with a sticky stats footer. Multi-target gets color-coded labels.
4. `--plain` streams formatted lines for scripts and quick shell use.

## License

MIT
