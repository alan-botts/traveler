# traveler

A flight search CLI built with Go, Cobra, and Bubbletea. Searches Google Flights for one-way flights and displays results in an interactive terminal UI or headless text output.

## Install

```bash
go install github.com/alan-botts/traveler@latest
```

Or build from source:

```bash
git clone https://github.com/alan-botts/traveler.git
cd traveler
go build -o travel .
```

## Usage

```bash
# Interactive TUI
./travel flights PHX OAK 2026-04-01

# Headless (for scripts/piping)
./travel flights PHX OAK 2026-04-01 --headless
```

Arguments:
- `origin` — 3-letter IATA airport code (e.g. PHX, SFO, JFK)
- `destination` — 3-letter IATA airport code
- `date` — travel date in YYYY-MM-DD format

## How it works

Traveler queries Google Flights' internal API endpoints (the same ones the Google Flights web UI uses) via TLS-fingerprinted HTTP requests.

## Disclaimer

This tool accesses Google's undocumented internal API and uses TLS fingerprinting to impersonate a browser. This may violate Google's Terms of Service. Use at your own risk and for personal/educational purposes only. The authors are not responsible for any consequences of using this tool.

## License

MIT — see [LICENSE](LICENSE).
