# traveler

Part of the **alan-botts tools** family: [movie-watcher](https://github.com/alan-botts/movie-watcher) | [divine](https://github.com/alan-botts/divine) | [strangerloops.com/tools](https://strangerloops.com/tools/)

A command-line tool that searches Google Flights using their undocumented internal API. It uses TLS fingerprinting to mimic a real Chrome browser, so Google's servers treat requests as legitimate browser traffic. If the RPC fails, it falls back to the flight data embedded in the public Google Flights search page and verifies that Google understood the requested route and date. Results are displayed in an interactive terminal UI (via [Bubbletea](https://github.com/charmbracelet/bubbletea)) or printed as plain text for scripting.

## Example

```
$ ./travel flights LAX SFO 2026-04-17 --headless
Searching flights: LAX -> SFO on 2026-04-17...

Found 36 flights: LAX → SFO on 2026-04-17

--- Flight 1 ---
  Price:    $108
  Duration: 1h 38m
  F9 2857  LAX 20:01 → SFO 21:39 (1h 38m)

--- Flight 2 ---
  Price:    $185
  Duration: 1h 30m
  WN 1105  LAX 00:00 → SFO 07:30 (1h 30m)

--- Flight 3 ---
  Price:    $185
  Duration: 1h 25m
  DL 1421  LAX 06:20 → SFO 07:45 (1h 25m)

--- Flight 4 ---
  Price:    $185
  Duration: 1h 34m
  UA 1372  LAX 07:15 → SFO 08:49 (1h 34m)

--- Flight 5 ---
  Price:    $185
  Duration: 1h 25m
  DL 1715  LAX 08:25 → SFO 09:50 (1h 25m)
```

## Features

- **Real Google Flights data** -- queries the same internal API that google.com/travel/flights uses
- **TLS fingerprinting** -- presents a Chrome 131 TLS fingerprint so requests are indistinguishable from browser traffic
- **Interactive TUI** -- browse results with keyboard navigation (j/k, arrows, Home/End)
- **Headless mode** -- `--headless` flag prints plain text for piping into scripts, `jq`, `grep`, etc.
- **One-way flight search** -- specify origin, destination, and date with standard IATA airport codes
- **Rate limiting** -- built-in rate limiter (10 req/s) to avoid hammering Google's servers
- **International dates** -- headless output includes local departure/arrival dates and handles omitted zero-minute values. Prices are USD, one-way, for one adult in economy; fares are search quotes, not booking confirmations.
- **No API key required** -- no accounts, tokens, or configuration needed

## Installation

Requires **Go 1.23** or later.

### macOS

Install Go via Homebrew if you don't have it:

```bash
brew install go
```

Then build traveler:

```bash
git clone https://github.com/alan-botts/traveler.git
cd traveler
go build -o travel .
```

The `travel` binary is now in the current directory. Move it somewhere on your PATH if you like:

```bash
sudo mv travel /usr/local/bin/
```

### Ubuntu / Debian

Install Go (option A -- snap):

```bash
sudo snap install go --classic
```

Or (option B -- from the official tarball):

```bash
# Download Go 1.23+ from https://go.dev/dl/
wget https://go.dev/dl/go1.23.6.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

Then build:

```bash
git clone https://github.com/alan-botts/traveler.git
cd traveler
go build -o travel .
```

### One-liner with `go install`

If you already have Go configured:

```bash
go install github.com/alan-botts/traveler@latest
```

This installs the binary as `traveler` in your `$GOPATH/bin` (or `$HOME/go/bin`). Rename it if you prefer:

```bash
mv "$(go env GOPATH)/bin/traveler" "$(go env GOPATH)/bin/travel"
```

## Usage

```
travel flights <origin> <destination> <date> [flags]
```

### Arguments

| Argument      | Description                              | Example      |
|---------------|------------------------------------------|--------------|
| `origin`      | 3-letter IATA airport code (departure)   | `LAX`, `JFK` |
| `destination` | 3-letter IATA airport code (arrival)     | `SFO`, `ORD` |
| `date`        | Travel date in YYYY-MM-DD format         | `2026-04-17`  |

### Flags

| Flag         | Description                                        |
|--------------|----------------------------------------------------|
| `--headless` | Print results as plain text instead of launching TUI |

### Examples

```bash
# Interactive TUI -- browse results with keyboard
./travel flights LAX SFO 2026-05-15

# Headless -- plain text output for scripts
./travel flights LAX SFO 2026-06-01 --headless

# Pipe to grep to find nonstop flights under $300
./travel flights LAX SFO 2026-04-17 --headless | grep -A3 "Flight" | grep "Price"
```

### TUI Controls

| Key              | Action              |
|------------------|---------------------|
| `j` / `Down`     | Move cursor down    |
| `k` / `Up`       | Move cursor up      |
| `g` / `Home`     | Jump to first result|
| `G` / `End`      | Jump to last result |
| `q` / `Esc`      | Quit                |

## How It Works

Google Flights has no public API. The web UI at google.com/travel/flights communicates with an internal RPC endpoint. Traveler reverse-engineers that protocol:

1. **Request encoding** -- The search parameters (airports, date, trip type, passenger count) are packed into a deeply nested JSON array, then double-encoded: the inner structure is JSON-stringified, wrapped in `[null, "<json>"]`, JSON-stringified again, URL-encoded, and sent as `f.req=...` in a POST body.

2. **TLS fingerprinting** -- Google checks TLS handshake characteristics (cipher suites, extensions, ordering) to distinguish real browsers from bots. Traveler uses [bogdanfinn/tls-client](https://github.com/bogdanfinn/tls-client) to present Chrome 131's exact TLS fingerprint, making requests look identical to a real browser at the network level.

3. **Response decoding** -- The response is also multi-layered: a XSSI prevention prefix (`)]}'`) is stripped, the outer JSON is parsed, and then an inner JSON string at `[0][2]` is parsed again to reach the actual flight data. Flight itineraries are extracted from deeply nested array indices (there is no schema -- positions were mapped by hand).

4. **Rate limiting** -- A mutex-protected rate limiter caps requests at 10/second to be a polite client.

## Disclaimer

This tool accesses Google's **undocumented internal API** and uses TLS fingerprinting to impersonate a browser. This may violate [Google's Terms of Service](https://policies.google.com/terms). Use at your own risk and for **personal or educational purposes only**. The authors are not responsible for any consequences of using this tool.

Google may change their API structure, response format, or fingerprinting detection at any time, which could break this tool without notice.

## License

MIT -- see [LICENSE](LICENSE).

## Contributing

Contributions are welcome. Feel free to open issues or submit pull requests.

Some areas that could use work:

- Round-trip flight search
- Additional sort options (duration, departure time)
- Output formats (JSON, CSV)
- Airport code autocomplete/validation against a real IATA database
- Caching results to avoid redundant requests

## Verification

Run `go test ./...` and `go vet ./...`. Regression fixtures cover RPC error
envelopes, metadata preceding results, malformed/empty responses, search-page
query verification, and whole-hour/midnight times. The compact shopping fixture
is derived from a public Google Flights response with booking tokens removed.

Both transports use undocumented Google formats and may change. If neither
returns valid data, the CLI reports both errors instead of inventing fares.
