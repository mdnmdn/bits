# Project Features

This document provides a comprehensive map of the features available in the Multi-Provider Crypto CLI (`bits`).

## 1. Multi-Provider Support

`bits` supports multiple data providers, allowing you to fetch market data from different sources:

- **CoinGecko** (Default): Market listings, prices, OHLCV candles, and live price streaming.
- **Binance**: Server time, exchange info, prices, candles, 24h ticker, order book, live order book streaming.
- **Bitget**: Server time, exchange info, prices, candles, 24h ticker.

Use the `--provider` / `-p` flag to select a provider, `--market` / `-m` for market type, and `--lock` / `-l` to prevent automatic fallback:

```bash
bits price BTC ETH                             # CoinGecko (default), USD
bits price BTCUSDT -p binance                  # Binance spot
bits price BTCUSDT -p binance -m futures       # Binance futures
bits ticker BTCUSDT -p coingecko               # no ticker on coingecko → fallback to binance
bits ticker BTCUSDT -p coingecko --lock        # error: coingecko does not support ticker
```

### Automatic Provider Fallback

When a provider doesn't support the requested feature or market, `bits` transparently falls back to the first capable provider. The output always indicates when a fallback occurred:

- **table**: footnote `† served by binance (requested: coingecko)`
- **json / yaml**: top-level `fallback: true`, `provider`, `requested_provider` keys
- **markdown**: blockquote `> † served by binance (requested: coingecko)`
- **toon**: styled amber note below the output box

Use `--lock` / `-l` to disable fallback and fail with an explicit error instead.

---

## 2. CLI Commands

### Global Flags

Applied to every command:

```
-p, --provider  string    provider id: coingecko | binance | bitget  (default: from config)
-m, --market    string    market type: spot | futures | margin        (default: spot)
-o, --output    string    output format: table | json | yaml | markdown | toon  (default: table)
-l, --lock                disable provider fallback
```

---

### `bits time`
- Show exchange server timestamp.
- Automatically computes latency and estimated clock skew (via `TimeEnricher` processor).
- Requires an exchange provider (Binance, Bitget).

```bash
bits time -p binance
bits time -p bitget
```

---

### `bits info`
- Display exchange symbol catalogue.
- Filter to a single symbol with `--symbol`.
- Supports spot and futures markets.

```bash
bits info -p binance
bits info -p binance -m futures
bits info -p binance --symbol BTCUSDT
```

---

### `bits price`
- Fetch current price(s) for one or more coin IDs or trading symbols.
- CoinGecko uses coin IDs (`bitcoin`, `ethereum`); exchanges use trading symbols (`BTCUSDT`).
- Batch-native: all IDs are sent to the provider in a single call.

```bash
bits price bitcoin ethereum                 # CoinGecko, USD
bits price bitcoin --currency eur           # CoinGecko, EUR
bits price BTCUSDT -p binance               # Binance spot
bits price BTCUSDT -p binance -m futures
bits price bitcoin -o json                  # JSON output with provenance
bits price bitcoin -o yaml                  # YAML output with provenance
bits price bitcoin -o toon                  # styled terminal output
```

---

### `bits ticker`
- 24h rolling statistics for one or more trading symbols.
- Multi-symbol: calls are fanned out in parallel; results collected into one response.
- Partial failures reported in `errors` field without aborting the whole request.

```bash
bits ticker BTCUSDT -p binance
bits ticker BTCUSDT ETHUSDT -p binance     # multi-symbol fan-out
bits ticker BTCUSDT -p binance -m futures
```

---

### `bits book`
- Order book snapshot (bids and asks).
- Customisable depth via `--depth` (default: 20).

```bash
bits book BTCUSDT -p binance
bits book BTCUSDT -p binance --depth 50
bits book BTCUSDT -p binance -m futures
```

---

### `bits candles`
- OHLCV candle history.
- `--interval` selects granularity (e.g. `1m`, `5m`, `1h`, `1d`).
- `--from` / `--to` accept RFC3339 or `YYYY-MM-DD`.
- `--limit` caps the number of candles returned.

```bash
bits candles BTCUSDT -p binance --interval 1h
bits candles BTCUSDT -p binance -m futures --from 2024-01-01 --to 2024-02-01
bits candles BTCUSDT -p bitget --limit 50
```

---

### `bits markets`
- Ranked coin listing by market cap.
- Aggregator-only (CoinGecko); automatically routes there if another provider is selected.
- Pagination via `--page` / `--per-page`.

```bash
bits markets
bits markets --currency eur
bits markets --page 2 --per-page 50
```

---

### `bits stream price`
- Live price feed via WebSocket.
- CoinGecko only; requires a paid API plan.
- Outputs continuously to stdout; one item per line/block.

```bash
bits stream price bitcoin ethereum
bits stream price bitcoin -o json        # JSONL: one compact JSON object per line
bits stream price bitcoin -o yaml        # one YAML document per update (--- separated)
bits stream price bitcoin -o markdown    # one markdown bullet per update
bits stream price bitcoin -o toon        # colored inline line per update
```

---

### `bits stream book`
- Live order book feed via WebSocket.
- Binance only; `--depth` controls subscription depth.
- Outputs continuously to stdout; one item per line/block.

```bash
bits stream book BTCUSDT -p binance
bits stream book BTCUSDT -p binance --depth 10
bits stream book BTCUSDT -p binance -o json      # JSONL
bits stream book BTCUSDT -p binance -o toon      # colored inline line
```

---

### `bits providers`
- List all registered providers with active marker.

```bash
bits providers
```

---

### `bits capabilities` (alias: `bits caps`)
- Display the capability matrix: which features each provider supports per market type.
- Filter to a single provider with `--provider` / `-p`.
- Requires no API key — reads static provider declarations.

```bash
bits capabilities
bits caps -p binance
bits capabilities -p coingecko
```

---

## 3. Output Formats

All commands support the `-o` / `--output` flag:

| Format     | Description |
|------------|-------------|
| `table`    | Human-readable tabwriter layout (default) |
| `json`     | JSON envelope with data + provenance fields (pretty-printed) |
| `yaml`     | YAML envelope with data + provenance fields |
| `markdown` | Markdown document: `# provider/market` heading, data as fenced YAML block |
| `toon`     | Lipgloss-styled terminal output: colored header + rounded box around YAML data |

All non-table formats include `provider`, `market`, and (when applicable) `fallback`,
`requested_provider`, `requested_market`, and `errors` provenance fields.

### Streaming format behaviour

Streaming commands (`bits stream price`, `bits stream book`) emit **continuous compact output**
— one item per line or YAML block, suitable for piping or tailing. No screen clearing occurs.

| Format     | Streaming behaviour |
|------------|---------------------|
| `json`     | JSONL — one compact JSON object per line |
| `yaml`     | One YAML document per update, separated by `---` |
| `markdown` | One markdown bullet (`- **SYMBOL** price currency  _change%_`) per update |
| `toon`     | Compact colored line per update (bold symbol, green price, red/green change) |
| `table`    | Plain compact line per update (same as before) |

---

## 4. Technical Architecture Features

### Typed Response Envelope
Every provider call returns `model.Response[T]`, carrying provenance (provider, market, fallback
info) and partial-failure errors alongside the data. No out-of-band signalling needed.

### Capability-Based Routing
The resolver (`internal/resolve/`) selects providers by declared capabilities, not hard-coded
conditionals. Fallback is transparent and always annotated.

### Processing Pipeline
An optional processor layer (`internal/process/`) enriches responses before rendering:
- `TimeEnricher` — computes latency and clock skew from server time
- `SpreadCalculator` — computes bid-ask spread and mid price for order books
- `CandleStats` — computes VWAP, typical price, body/wick ratios for candles

### Parallel Fan-Out
Multi-symbol commands (e.g. `bits ticker BTCUSDT ETHUSDT`) fan out in parallel using
`resolve.FanOut`. Partial failures are collected in `Response.Errors` and don't abort the whole
request.

### WebSocket Streaming
The CoinGecko WebSocket client (`internal/ws/`) implements the ActionCable protocol with
automatic reconnection and exponential backoff with jitter. Binance order book streaming uses
gorilla/websocket with the combined-stream API.

### Secure Configuration
- API keys masked in display (`CoinGeckoConfig.MaskedKey()`)
- Auth applied per-request (`CoinGeckoConfig.ApplyAuth()`)
- Platform-specific config directories; `.env` file support alongside YAML
- `BITS_*` env vars override all config file values
