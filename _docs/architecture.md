# bits Architecture

## Overview

`bits` is a multi-provider crypto CLI tool written in Go. It uses a capability-based provider architecture that allows different data sources (CoinGecko, Binance, Bitget) to be used interchangeably through a unified command interface.

All provider responses are wrapped in a typed `Response[T]` envelope carrying provenance (which provider and market actually served the data), enabling transparent fallback and consistent rendering.

## Provider Architecture

### Interface Hierarchy

```
internal/provider/
├── Provider (base)          → ID, SetUserAgent, Capabilities
├── ExchangeProvider         → Provider + ServerTime, ExchangeInfo
├── AggregatorProvider       → Provider + CoinMarkets
├── PriceProvider            → Price(ids, currency)            [batch-native]
├── CandleProvider           → Candles(symbol, market, interval, opts)
├── TickerProvider           → Ticker24h(symbol, market)       [fan-out for multi]
├── OrderBookProvider        → OrderBook(symbol, market, depth)
├── PriceStreamProvider      → WatchPrices(ids)
└── OrderBookStreamProvider  → WatchOrderBook(symbol, market, depth)
```

All market-sensitive calls carry a `model.MarketType` parameter (`"spot"` | `"futures"` | `"margin"`).

### Provider Capabilities

| Interface              | CoinGecko | Binance spot | Binance futures | Bitget spot | Bitget futures |
|------------------------|:---------:|:------------:|:---------------:|:-----------:|:--------------:|
| `ExchangeProvider`     | —         | Yes          | Yes             | Yes         | Yes            |
| `AggregatorProvider`   | Yes       | —            | —               | —           | —              |
| `PriceProvider`        | Yes       | Yes          | Yes             | Yes         | Yes            |
| `CandleProvider`       | Yes       | Yes          | Yes             | Yes         | Yes            |
| `TickerProvider`       | —         | Yes          | Yes             | Yes         | Yes            |
| `OrderBookProvider`    | —         | Yes          | Yes             | —           | —              |
| `PriceStreamProvider`  | Yes       | —            | —               | —           | —              |
| `OrderBookStreamProvider` | —      | Yes          | —               | —           | —              |

Use `bits capabilities` or `bits caps -p <provider>` to inspect the matrix at runtime.

### Response Envelope

Every provider call returns a typed `Response[T]` that carries provenance:

```go
type Response[T any] struct {
    Data              T
    Provider          string       // provider that served this response
    Market            MarketType   // market that served this response
    Fallback          bool         // true when automatic fallback occurred
    RequestedProvider string       // populated only when Fallback is true
    RequestedMarket   MarketType   // populated only when Fallback is true
    Errors            []ItemError  // partial failures in batch/multi-symbol calls
}
```

Renderers use provenance to display footnotes (table), top-level keys (JSON/YAML), or annotations (markdown).

## Data Model (`internal/model/`)

All providers return types from `internal/model/`. Key types:

- `Response[T]` — generic envelope with provenance fields
- `MarketType` — type alias for `capability.MarketType` (`"spot"` | `"futures"` | `"margin"`)
- `ServerTime` — exchange server timestamp with optional latency/skew fields
- `ExchangeInfo` / `Symbol` — exchange symbol catalogue
- `Candle` — OHLCV candle with optional volume and close time
- `Ticker24h` — 24h rolling statistics (all optional fields use pointer types)
- `OrderBook` / `OrderBookEntry` — order book depth snapshot
- `CoinPrice` — current price (aggregators use coin IDs; exchanges use symbols)
- `CoinMarket` / `MarketOpts` — ranked coin listing with market metadata
- `ErrUnsupportedMarket`, `ErrUnsupportedFeature` — typed errors

Optional values use pointer types (`*float64`, `*time.Time`, `*int`). Provider-specific data is preserved in `Extra map[string]any`.

## Resolution Layer (`internal/resolve/`)

The resolver selects the best provider+market for a requested feature, with optional fallback:

```go
type ResolutionOpts struct {
    Provider string          // explicit override ("" = config/default)
    Market   MarketType      // explicit market ("" = spot)
    Lock     bool            // if true, error instead of fallback
}

// Returns: actualProvider, actualMarket, wasFallback, error
func (r *Resolver) Resolve(ctx, feature, opts) (Provider, MarketType, bool, error)
```

Fallback order: requested provider → same provider with spot market → other registered providers.

`resolve.Require[T]` asserts a provider implements interface T:

```go
tp, err := resolve.Require[provider.TickerProvider](p, "ticker")
```

`resolve.FanOut` fans out single-symbol calls in parallel for multi-symbol requests, collecting results into `Response[[]T]`.

## Rendering Layer (`internal/render/`)

Rendering is decoupled from providers. Commands hand a `Response[T]` to a renderer:

- `internal/render/renderer.go` — `Format` type (`table` | `json` | `markdown` | `yaml` | `toon`) + `ParseFormat`
- `internal/render/provenance.go` — shared helpers: `FallbackFootnote`, `ProviderLabel`
- `internal/render/json/` — generic JSON renderer for any `Response[T]`
- `internal/render/table/` — table renderers: `server_time`, `exchange_info`, `price`, `ticker`, `orderbook`, `candles`, `markets`

## Processing Layer (`internal/process/`)

Optional processors enrich a `Response[T]` before rendering. Processors are composable:

```go
type Processor[T any] func(Response[T]) Response[T]

func Apply[T any](res Response[T], processors ...Processor[T]) Response[T]
```

Available processors: `TimeEnricher` (latency + clock skew), `SpreadCalculator` (bid-ask spread), `CandleStats` (VWAP, typical price, body/wick ratios).

## Provider Implementations

### CoinGecko (`internal/provider/coingecko/`)
- Raw HTTP client; uses `config.CoinGeckoConfig` for base URL and auth header
- Supports demo and paid API tiers
- Implements: `AggregatorProvider`, `PriceProvider`, `CandleProvider`, `PriceStreamProvider`
- Stream via `internal/ws/` (WebSocket, ActionCable protocol, paid plan required)

### Binance (`internal/provider/binance/`)
- Uses `go-binance/v2` library for spot/futures HTTP; gorilla/websocket for streaming
- Market routing by method parameter (not state); spot+futures clients held per instance
- Implements: `ExchangeProvider`, `PriceProvider`, `CandleProvider`, `TickerProvider`, `OrderBookProvider`, `OrderBookStreamProvider`

### Bitget (`internal/provider/bitget/`)
- Raw HTTP client; 3-part auth (key + secret + passphrase) via `internal/auth/`
- Implements: `ExchangeProvider`, `PriceProvider`, `CandleProvider`, `TickerProvider`

## Registry (`internal/registry/`)

Lives in its own package to avoid import cycles (providers must not import the registry that imports them):

```go
func NewProvider(name string, cfg *config.Config) (provider.Provider, error)
func AllProviderIDs() []string  // ["coingecko", "binance", "bitget"]
```

## Configuration (`internal/config/`)

Config file: platform-specific path (e.g. `~/Library/Application Support/bits-cli/config.yaml` on macOS, `~/.config/bits/config.yaml` on Linux).

```toml
provider = "coingecko"   # active provider

[coingecko]
api_key = ""
tier = "demo"            # demo or paid

[binance]
api_key = ""
api_secret = ""

[binance.spot]
enabled = true

[bitget]
api_key = ""
api_secret = ""
passphrase = ""
```

Environment variable overrides (`BITS_*` prefix):
- `BITS_PROVIDER` — active provider
- `BITS_COINGECKO_API_KEY`, `BITS_COINGECKO_TIER`
- `BITS_BINANCE_API_KEY`, `BITS_BINANCE_API_SECRET`
- `BITS_BITGET_API_KEY`, `BITS_BITGET_API_SECRET`, `BITS_BITGET_PASSPHRASE`

CoinGecko-specific helpers (`GetBaseURL`, `GetAuthHeader`, `IsPaid`, `MaskedKey`, `ApplyAuth`) are methods on `CoinGeckoConfig`, not the root `Config`.

## CLI Provider Resolution

The `--provider` / `-p` flag selects the provider; `--market` / `-m` selects the market type; `--lock` / `-l` disables fallback.

Resolution order: `--provider` flag → `BITS_PROVIDER` env → config file → `"coingecko"` default.

```bash
bits price BTC ETH                             # CoinGecko (default), USD
bits price BTCUSDT -p binance                  # Binance spot
bits price BTCUSDT -p binance -m futures       # Binance futures
bits ticker BTCUSDT -p coingecko               # coingecko lacks ticker → fallback to binance
bits ticker BTCUSDT -p coingecko --lock        # error: coingecko does not support ticker
```

## Adding a New Provider

1. Create `internal/provider/<name>/` with `client.go`, market files
2. Implement `Provider` base interface (`ID`, `SetUserAgent`, `Capabilities`)
3. Implement applicable capability interfaces; all methods return `model.Response[T]`
4. Register capabilities in `Capabilities()` using `capability.NewCapabilityMatrix(...)`
5. Add a config struct to `internal/config/config.go` and wire env overrides
6. Register in `internal/registry/registry.go` — one new case in `NewProvider`

## Project Structure

```
bits/
├── main.go
├── cmd/
│   ├── root.go              # RootCmd, Execute(), global flags (-p, -m, -o, -l)
│   ├── factory.go           # loadConfig(), newResolver(), flag helpers
│   ├── providers.go         # bits providers
│   ├── capabilities.go      # bits capabilities [--provider id]
│   ├── time.go              # bits time
│   ├── price.go             # bits price <id>...
│   ├── ticker.go            # bits ticker <symbol>...
│   ├── book.go              # bits book <symbol>
│   ├── candles.go           # bits candles <symbol>
│   ├── info.go              # bits info [--symbol S]
│   ├── markets.go           # bits markets
│   ├── stream.go            # bits stream (group)
│   ├── stream_price.go      # bits stream price <id>...
│   └── stream_book.go       # bits stream book <symbol>
├── internal/
│   ├── auth/
│   │   └── signature.go     # HMAC-SHA256 helpers (used by Bitget)
│   ├── capability/
│   │   └── capability.go    # MarketType, Feature, CapabilityKey, CapabilityMatrix
│   ├── config/
│   │   └── config.go        # Multi-provider config (YAML + Env + .env)
│   ├── model/               # Provider-agnostic data types
│   │   ├── market.go        # MarketType alias + constants
│   │   ├── response.go      # Response[T], ItemError
│   │   ├── exchange.go      # ServerTime, ExchangeInfo, Symbol
│   │   ├── candle.go        # Candle, CandleOpts
│   │   ├── ticker.go        # Ticker24h
│   │   ├── orderbook.go     # OrderBook, OrderBookEntry
│   │   ├── price.go         # CoinPrice
│   │   ├── coin.go          # CoinMarket, MarketOpts
│   │   └── errors.go        # ErrUnsupportedMarket, ErrUnsupportedFeature
│   ├── provider/            # Provider interfaces
│   │   ├── provider.go      # Provider base interface
│   │   ├── exchange.go      # ExchangeProvider
│   │   ├── aggregator.go    # AggregatorProvider
│   │   ├── capability.go    # PriceProvider, CandleProvider, TickerProvider, OrderBookProvider
│   │   ├── stream.go        # PriceStreamProvider, OrderBookStreamProvider
│   │   ├── binance/         # Binance implementation
│   │   ├── bitget/          # Bitget implementation
│   │   └── coingecko/       # CoinGecko implementation
│   ├── registry/
│   │   └── registry.go      # NewProvider factory (avoids import cycle)
│   ├── resolve/
│   │   ├── resolver.go      # Resolver, ResolutionOpts, Resolve()
│   │   ├── require.go       # Require[T] type assertion helper
│   │   └── fanout.go        # FanOut[T] parallel multi-symbol helper
│   ├── render/
│   │   ├── renderer.go      # Format type + ParseFormat
│   │   ├── provenance.go    # FallbackFootnote, ProviderLabel
│   │   ├── json/            # Generic JSON renderer
│   │   └── table/           # Table renderers per data type
│   ├── process/
│   │   ├── process.go       # Processor[T] type + Apply combinator
│   │   ├── time.go          # TimeEnricher
│   │   ├── orderbook.go     # SpreadCalculator
│   │   └── candles.go       # CandleStats
│   └── ws/
│       ├── base_client.go   # Generic WebSocket client (reconnect, backoff)
│       └── client.go        # CoinGecko WebSocket client (ActionCable)
├── Makefile
├── .goreleaser.yml
├── install.sh
└── .github/workflows/       # CI and Release
```
