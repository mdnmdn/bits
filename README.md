# bits

**The crypto Swiss Army knife for your terminal.**

One CLI. Three exchanges. Every format. Real-time streams, snapshots, order books, candles — all from the same command set, no matter which provider you're pointing at.


---

## What it does

```sh
bits price bitcoin ethereum                    # prices from CoinGecko
bits price BTCUSDT -p binance -m futures       # or Binance futures
bits ticker BTCUSDT ETHUSDT -p binance         # 24h stats, parallel fan-out
bits book BTCUSDT -p binance --depth 50        # order book snapshot
bits candles BTCUSDT -p bitget --interval 1h   # OHLCV history
bits stream price bitcoin -o json | jq .price  # live WebSocket → jq
bits capabilities                              # what can each provider do?
```

Pick a provider with `-p`, a market with `-m`, an output format with `-o`. That's it.

---

## Providers

| Provider | Markets | What it gives you |
|---|---|---|
| **CoinGecko** | — | Prices, candles, ranked markets, live price stream |
| **Binance** | spot · futures | Server time, exchange info, prices, candles, ticker, order book, live book stream |
| **Bitget** | spot · futures | Server time, exchange info, prices, candles, ticker |

Switch providers with `-p coingecko / -p binance / -p bitget`.

**Fallback is controlled by whether you use `-p`:**
- No `-p` → fallback allowed; `bits` auto-routes to a capable provider.
- With `-p` → no fallback by default; `bits` errors if that provider can't serve the request.
- `-p ... -f` → opt-in to fallback even when the provider is explicit.

```sh
bits ticker BTCUSDT                     # no -p → auto-routes to binance
bits ticker BTCUSDT -p coingecko        # error: coingecko does not support ticker
bits ticker BTCUSDT -p coingecko -f     # -f → falls back to binance
```

---

## Output formats

Every command supports `-o` with five formats:

| Flag | Output |
|---|---|
| `table` | Aligned tabwriter — human-readable (default) |
| `json` | Pretty-printed JSON envelope with provenance metadata |
| `yaml` | Same as JSON but YAML |
| `markdown` | Markdown doc — heading + fenced YAML block |
| `toon` | Token friendly structured format |

Streaming commands (`bits stream`) emit continuous compact output per update:

| Flag | Streaming output |
|---|---|
| `json` | JSONL — one compact JSON object per line |
| `yaml` | One YAML doc per update, `---` separated |
| `markdown` | One markdown bullet per update |
| `toon` | Colored inline line per update |

---

## Install

```sh
# Homebrew
brew install mdnmdn/bits/bits

# Shell script
curl -sSfL https://raw.githubusercontent.com/mdnmdn/bits/main/install.sh | sh

# Go
go install github.com/mdnmdn/bits@latest
```

Or download a binary from [Releases](https://github.com/mdnmdn/bits/releases).

---

## Usage as a Library

`bits` can be used as a Go library to build your own crypto tools. Its core components (models, providers, and registry) are available in the `pkg/` directory.

### Quick Start

```go
import (
	"context"
	"fmt"
	"github.com/mdnmdn/bits/pkg/bits"
	"github.com/mdnmdn/bits/pkg/config"
)

func main() {
	// Initialize with spot enabled for Binance
	cfg := &config.Config{
		Binance: config.BinanceConfig{
			Spot: config.MarketConfig{Enabled: true},
		},
	}
	client := bits.NewClient(cfg)

	// Fetch price from Binance
	res, _ := client.GetPrice(context.Background(), "BTCUSDT", "binance")
	fmt.Printf("BTC Price: %.2f\n", res.Data.Price)
}
```

Check the `examples/` directory for more detailed use cases, including concurrent price comparison across multiple exchanges. To run an example, use:

```sh
go run ./examples/basic_usage
```

---

## Setup

Config file: `~/Library/Application Support/bits-cli/config.yaml` (macOS) or `~/.config/bits/config.yaml` (Linux).

```yaml
provider: coingecko

[coingecko]
api_key: ""
tier: demo        # demo | paid

[binance]
api_key: ""
api_secret: ""

[bitget]
api_key: ""
api_secret: ""
passphrase: ""
```

All values accept `BITS_*` environment variable overrides:

```sh
BITS_PROVIDER=binance
BITS_COINGECKO_API_KEY=your_key   BITS_COINGECKO_TIER=paid
BITS_BINANCE_API_KEY=your_key     BITS_BINANCE_API_SECRET=your_secret
BITS_BITGET_API_KEY=your_key      BITS_BITGET_API_SECRET=your_secret   BITS_BITGET_PASSPHRASE=your_pass
```

---

## Command reference

### Global flags

```
-p, --provider       string   coingecko | binance | bitget  (default: from config)
-m, --market         string   spot | futures | margin        (default: spot)
-o, --output         string   table | json | yaml | markdown | toon  (default: table)
-f, --allow-fallback          allow fallback even when --provider is set
```

---

### `bits price`

Current price for one or more coin IDs (CoinGecko) or symbols (exchanges). Batch-native — one API call regardless of how many IDs you pass.

```sh
bits price bitcoin ethereum
bits price bitcoin --currency eur
bits price BTCUSDT ETHUSDT -p binance -m futures
bits price bitcoin -o toon
```

---

### `bits ticker`

24h rolling stats — last price, change %, high, low, volume. Multi-symbol calls fan out in parallel; partial failures don't abort the rest.

```sh
bits ticker BTCUSDT -p binance
bits ticker BTCUSDT ETHUSDT SOLUSDT -p binance
bits ticker BTCUSDT -p binance -m futures -o json
```

---

### `bits book`

Order book depth snapshot — bids and asks side by side.

```sh
bits book BTCUSDT -p binance
bits book BTCUSDT -p binance --depth 100
bits book BTCUSDT -p binance -m futures -o yaml
```

---

### `bits candles`

OHLCV candle history with flexible time range.

```sh
bits candles BTCUSDT -p binance --interval 1h
bits candles BTCUSDT -p binance -m futures --from 2024-01-01 --to 2024-06-01
bits candles bitcoin --limit 100 -o json
```

| Flag | Default | Description |
|---|---|---|
| `--interval` | `1h` | `1m` `5m` `1h` `4h` `1d` etc. |
| `--from` | — | RFC3339 or `YYYY-MM-DD` |
| `--to` | — | RFC3339 or `YYYY-MM-DD` |
| `--limit` | — | Max candles (0 = provider default) |

---

### `bits time`

Exchange server time with computed round-trip latency and clock skew. Exchanges only (Binance, Bitget).

```sh
bits time -p binance
bits time -p bitget -o json
```

---

### `bits info`

Full symbol catalogue for an exchange. Use `--symbol` to filter.

```sh
bits info -p binance
bits info -p binance -m futures
bits info -p binance --symbol BTCUSDT
```

---

### `bits markets`

Ranked coin list by market cap. CoinGecko only (aggregator feature); automatically routed there.

```sh
bits markets
bits markets --currency eur --per-page 50
bits markets --page 3 -o yaml
```

---

### `bits stream price`

Live WebSocket price feed — CoinGecko paid plan required. One update per line; Ctrl+C to stop.

```sh
bits stream price bitcoin ethereum
bits stream price bitcoin -o json | jq .price    # pipe prices
bits stream price bitcoin -o yaml                # YAML docs, --- separated
bits stream price bitcoin -o toon                # colored live lines
```

---

### `bits stream book`

Live WebSocket order book feed — Binance only. One update per line; Ctrl+C to stop.

```sh
bits stream book BTCUSDT -p binance
bits stream book BTCUSDT -p binance --depth 5
bits stream book BTCUSDT -p binance -o json      # JSONL
```

---

### `bits providers`

List all registered providers and which one is currently active.

```sh
bits providers
```

---

### `bits capabilities` / `bits caps`

The capability matrix: which features each provider supports, by market type. No API key needed.

```sh
bits capabilities
bits caps -p binance
```

---

## Development

```sh
make build    # → ./bits
make test     # go test -race ./...
make lint     # golangci-lint
```

## Tech Stack

| Package | Purpose |
|---|---|
| [cobra](https://github.com/spf13/cobra) | CLI framework |
| [viper](https://github.com/spf13/viper) | Config (YAML + env vars) |
| [go-binance/v2](https://github.com/adshao/go-binance) | Binance HTTP client |
| [gorilla/websocket](https://github.com/gorilla/websocket) | WebSocket streaming |
| [lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal styling (`toon` format) |
| [goreleaser](https://goreleaser.com) | Cross-platform release builds |

## License

MIT — see [LICENSE](LICENSE).
