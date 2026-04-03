<p align="center">
  <img src="assets/icon.png" width="128" alt="goping">
  <h1 align="center">goping</h1>
  <p align="center">Ping with live stats — like ping but with a persistent stats bar</p>
</p>

<p align="center">
  <a href="https://github.com/Aayush9029/goping/releases/latest"><img src="https://img.shields.io/github/v/release/Aayush9029/goping" alt="Release"></a>
  <a href="https://github.com/Aayush9029/goping/blob/main/LICENSE"><img src="https://img.shields.io/github/license/Aayush9029/goping" alt="License"></a>
</p>

<p align="center">
  <img src="assets/demo.gif" alt="goping demo" width="800">
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
goping 1.1.1.1                        # ping a single target
goping github.com cloudflare.com       # multi-target with color labels
goping --plain --count 5 8.8.8.8      # streaming output, 5 probes
goping --interval 500ms 1.1.1.1       # custom interval
goping --duration 10s apple.com        # stop after 10 seconds
```

## Options

| Option | Description |
|--------|-------------|
| `-i, --interval` | Probe interval (default: 1s) |
| `--timeout` | Probe timeout (default: 3s) |
| `-c, --count` | Stop after n probes per target |
| `--duration` | Stop after a fixed runtime |
| `--plain` | Disable TUI, print streaming lines |
| `-4` / `-6` | IPv4 or IPv6 resolution |

## License

MIT
