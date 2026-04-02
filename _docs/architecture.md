# bits Architecture

## Overview

> ⚠️ **Status: Work in Progress** — The library API may change rapidly.

`bits` is a multi-provider crypto library and CLI tool written in Go. It uses a capability-based provider architecture that allows different data sources (CoinGecko, Binance, Bitget, WhiteBit, Crypto.com, MEXC) to be used interchangeably through a unified command interface.

**Architectural Principle**: The bits library (`github.com/mdnmdn/bits`) is the first citizen of this project. The CLI is a thin wrapper that uses the public library interface. External projects can import and extend the CLI commands.

## Provider Architecture

### Interface Hierarchy

```
provider/
├── Provider (base)          → ID, SetUserAgent, Capabilities
├── ExchangeProvider         → Provider + ServerTime, ExchangeInfo
├── AggregatorProvider       → Provider + CoinMarkets
├── PriceProvider            → Price(ids, currency)            [batch-native]
├── CandleProvider           → Candles(symbol, market, interval, opts)
├── TickerProvider           → Ticker24h(symbol, market)       [fan-out for multi]
├── OrderBookProvider        → OrderBook(symbol, market, depth)
├── PriceStreamProvider      → StartPriceStream, SubscribePrice, StopPriceStream
└── OrderBookStreamProvider  → StartOrderBookStream, SubscribeOrderBook, StopOrderBookStream
```

### Provider Capabilities

| Interface              | CoinGecko | Binance spot | Binance futures | Bitget spot | Bitget futures | WhiteBit spot | WhiteBit futures | Crypto.com | MEXC |
|------------------------|:---------:|:------------:|:---------------:|:-----------:|:--------------:|:-------------:|:----------------:|:----------:|:----:|
| `ExchangeProvider`     | —         | Yes          | Yes             | Yes         | Yes            | Yes           | Yes              | Yes        | Yes  |
| `AggregatorProvider`   | Yes       | —            | —               | —           | —              | —             | —                | —          | —    |
| `PriceProvider`        | Yes       | Yes          | Yes             | Yes         | Yes            | Yes           | Yes              | Yes        | Yes  |
| `CandleProvider`       | Yes       | Yes          | Yes             | Yes         | Yes            | Yes           | Yes              | —          | —    |
| `TickerProvider`       | —         | Yes          | Yes             | Yes         | Yes            | Yes           | Yes              | Yes        | Yes  |
| `OrderBookProvider`    | —         | Yes          | Yes             | —           | —              | —             | —                | —          | —    |
| `PriceStreamProvider`  | Yes       | —            | —               | —           | —              | Yes           | Yes              | Yes        | Yes  |
| `OrderBookStreamProvider` | —      | Yes          | —               | —           | —              | Yes           | Yes              | Yes        | Yes  |

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

// ItemError pairs a symbol with its typed error.
type ItemError struct {
    Symbol string
    Err    *ProviderError
}
```

Renderers use provenance to display footnotes (table), top-level keys (JSON/YAML), or annotations (markdown).

## Data Model (`model/`)

All providers return types from `model/`. Key types:

- `Response[T]` — generic envelope with provenance fields
- `ItemError` — per-symbol failure: `Symbol string` + `Err *ProviderError`
- `ProviderError` — typed cross-provider error (see Error Handling section below)
- `MarketType` — type alias for `capability.MarketType` (`"spot"` | `"futures"` | `"margin"`)
- `ServerTime` — exchange server timestamp with optional latency/skew fields
- `ExchangeInfo` / `Symbol` — exchange symbol catalogue
- `Candle` — OHLCV candle with optional volume and close time
- `Ticker24h` — 24h rolling statistics (all optional fields use pointer types)
- `OrderBook` / `OrderBookEntry` — order book depth snapshot
- `CoinPrice` — current price (aggregators use coin IDs; exchanges use symbols)
- `CoinMarket` / `MarketOpts` — ranked coin listing with market metadata

Optional values use pointer types (`*float64`, `*time.Time`, `*int`). Provider-specific data is preserved in `Extra map[string]any`.

## Error Handling

All errors that cross a provider boundary are `*model.ProviderError`. This type normalises
heterogeneous provider errors (HTTP status codes, API envelope codes, network failures) into a
single inspectable value without losing raw provider detail.

### `ErrorKind` — normalised category

```go
type ErrorKind int

const (
    ErrKindUnknown           ErrorKind = iota // unclassified
    ErrKindAuth                               // 401 / 403 / invalid API key
    ErrKindRateLimit                          // 429 / quota exceeded
    ErrKindNotFound                           // 404 / symbol not found
    ErrKindInvalidRequest                     // 400 / bad parameters
    ErrKindServerError                        // 5xx / provider-side transient failure
    ErrKindNetwork                            // connection refused, DNS failure
    ErrKindCanceled                           // context.Canceled / DeadlineExceeded
    ErrKindParse                              // unexpected response shape
    ErrKindUnsupportedMarket                  // market type not supported by this provider
    ErrKindUnsupportedFeature                 // feature not supported by this provider
)

func (k ErrorKind) Retryable() bool  // true for RateLimit, ServerError, Network
```

### `ProviderError` — the error type

```go
type ProviderError struct {
    Kind            ErrorKind // normalised category
    ProviderID      string    // "binance", "bitget", … empty for resolution-layer errors
    ProviderCode    string    // provider's native error code ("00000", "40001", …)
    ProviderMessage string    // provider's native message text
    HTTPStatus      int       // 0 when not an HTTP error
    Cause           error     // underlying cause (net error, json error, …)
}
```

`ProviderError` implements `Unwrap()` and `Is()` so standard `errors.As` / `errors.Is` chains
work correctly. The sentinel vars `ErrUnsupportedMarket` and `ErrUnsupportedFeature` are
`*ProviderError` values and match via `Kind` comparison.

### Caller usage

```go
res, err := client.Price(ctx, []string{"BTCUSDT"}, "")
if err != nil {
    var pe *model.ProviderError
    if errors.As(err, &pe) {
        switch pe.Kind {
        case model.ErrKindRateLimit:
            // back off and retry
        case model.ErrKindAuth:
            // surface config problem to user
        case model.ErrKindCanceled:
            return // caller canceled — do not retry
        }
    }
}

// Partial failures in batch calls:
for _, ie := range res.Errors {
    if ie.Err.Kind == model.ErrKindNotFound {
        // skip missing symbols gracefully
    }
}
```

### `WrapError` — migration shim

`model.WrapError(providerID, err)` wraps any `error` as `*ProviderError{Kind: ErrKindUnknown}`.
If the error is already a `*ProviderError` it is returned as-is. Use this at sites where a
typed error is required but only a plain `error` is available (e.g. third-party library errors).

### Provider responsibility

- Providers **never** log errors — they return them.
- Every `fmt.Errorf` at a provider boundary is a bug; use `providerErr` / `httpErr` / `apiErr`
  helpers defined in each provider's `errors.go`.
- HTTP status codes are checked on every response; non-2xx is returned as `httpErr(status, body)`.
- API-envelope failures (`code != "00000"`, `!success`, etc.) are returned as `apiErr(code, msg)`.
- Network and context errors are wrapped with the appropriate `ErrKindNetwork` / `ErrKindCanceled`.

See `_docs/error-handling.md` for the full reference.

## Library API (`bits/`)

The `bits` package provides a high-level client:

```go
import "github.com/mdnmdn/bits"

// Multi-provider client
client := bits.NewClient(cfg, bits.WithSymbolEngine())
client.GetPriceWithResolution(ctx, "BTC-USDT", "binance", "spot")
client.ComparePricesWithResolution(ctx, "BTC-USDT", []string{"binance", "bitget"}, "spot")

// Provider-specific client (stateful, for WebSocket)
p := bits.NewProvider(cfg, "binance", bits.WithSymbolEngine())
p.ID()
p.Capabilities()
p.Price(ctx, []string{"BTCUSDT"}, "")
p.Ticker24h(ctx, "BTCUSDT", "spot")
p.StartPriceStream(ctx, []string{"bitcoin"})
```

When a provider doesn't support a capability, methods return a "not implemented" error instead of panicking (via NullProvider).

## CLI Architecture (`command/`)

The CLI uses the public bits library interface. Commands are defined in the `command/` package and can be imported and extended by external projects:

```go
import "github.com/mdnmdn/bits/command"

func main() {
    // Add custom commands to the CLI
    command.Root.AddCommand(myCustomCmd)
    
    // Or add multiple at once
    command.Root.AddCommands(cmd1, cmd2, cmd3)
    
    if err := command.Root.Execute(); err != nil {
        // handle error
    }
}
```

### Command Pattern

Every CLI command uses the public bits library:

```go
func runPrice(cmd *cobra.Command, args []string) error {
    cfg, _, err := command.LoadConfig()           // load config
    providerID, market, format := command.ResolveOptions(cmd)

    client := bits.NewProvider(cfg, providerID, bits.WithSymbolEngine())
    res, err := client.Price(ctx, args, currency)

    if ok, err := command.RenderGeneric(w, format, res); ok || err != nil {
        return err
    }
    return table.RenderPrices(w, res)
}
```

## Rendering Layer (`render/`)

Rendering is decoupled from providers. Commands hand a `Response[T]` to a renderer:

- `render/renderer.go` — `Format` type (`table` | `json` | `markdown` | `yaml` | `toon`) + `ParseFormat`
- `render/provenance.go` — shared helpers: `FallbackFootnote`, `ProviderLabel`
- `render/json/` — generic JSON renderer for any `Response[T]`
- `render/table/` — table renderers: `server_time`, `exchange_info`, `price`, `ticker`, `orderbook`, `candles`, `markets`

## Processing Layer (`process/`)

Optional processors enrich a `Response[T]` before rendering. Processors are composable:

```go
type Processor[T any] func(Response[T]) Response[T]

func Apply[T any](res Response[T], processors ...Processor[T]) Response[T]
```

Available processors: `TimeEnricher` (latency + clock skew), `SpreadCalculator` (bid-ask spread), `CandleStats` (VWAP, typical price, body/wick ratios).

## Provider Implementations

### CoinGecko (`provider/coingecko/`)
- Raw HTTP client; uses `config.CoinGeckoConfig` for base URL and auth header
- Supports demo and paid API tiers
- Implements: `AggregatorProvider`, `PriceProvider`, `CandleProvider`, `PriceStreamProvider`
- Stream via `internal/ws/` (WebSocket, ActionCable protocol, paid plan required)

### Binance (`provider/binance/`)
- Uses `go-binance/v2` library for spot/futures HTTP; gorilla/websocket for streaming
- Market routing by method parameter (not state); spot+futures clients held per instance
- Implements: `ExchangeProvider`, `PriceProvider`, `CandleProvider`, `TickerProvider`, `OrderBookProvider`, `OrderBookStreamProvider`

### Bitget (`provider/bitget/`)
- Raw HTTP client; 3-part auth (key + secret + passphrase) via `internal/auth/`
- Implements: `ExchangeProvider`, `PriceProvider`, `CandleProvider`, `TickerProvider`

### WhiteBit (`provider/whitebit/`)
- Raw HTTP client; public endpoints require no auth
- Implements: `ExchangeProvider`, `PriceProvider`, `CandleProvider`, `TickerProvider`, `PriceStreamProvider`, `OrderBookStreamProvider`

### Crypto.com (`provider/cryptocom/`)
- Raw HTTP client; spot only
- Implements: `ExchangeProvider`, `PriceProvider`, `TickerProvider`, `PriceStreamProvider`, `OrderBookStreamProvider`

### MEXC (`provider/mexc/`)
- Raw HTTP client; spot only with protobuf parsing for some endpoints
- Implements: `ExchangeProvider`, `PriceProvider`, `TickerProvider`, `PriceStreamProvider`, `OrderBookStreamProvider`

## Registry (`provider/registry/`)

Lives in its own package to avoid import cycles (providers must not import the registry that imports them):

```go
func NewProvider(name string, cfg *config.Config) (provider.Provider, error)
func AllProviderIDs() []string  // ["coingecko", "binance", "bitget", "whitebit", "cryptocom"]
```

## Configuration (`config/`)

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

The `--provider` / `-p` flag selects the provider; `--market` / `-m` selects the market type.

Resolution order: `--provider` flag → `BITS_PROVIDER` env → config file → `"coingecko"` default.

```bash
bits price BTC ETH                              # CoinGecko (default), USD
bits price BTCUSDT -p binance                   # Binance spot
bits price BTCUSDT -p binance -m futures        # Binance futures
bits ticker BTCUSDT -p coingecko               # error: coingecko does not support ticker
```

## Adding a New Provider

1. Create `provider/<name>/` with `client.go`, market files
2. Implement `Provider` base interface (`ID`, `SetUserAgent`, `Capabilities`)
3. Implement applicable capability interfaces; all methods return `model.Response[T]`
4. Register capabilities in `Capabilities()` using `capability.NewCapabilityMatrix(...)`
5. Add a config struct to `config/config.go` and wire env overrides
6. Register in `provider/registry/registry.go` — one new case in `NewProvider`

## Project Structure

```
bits/
├── bits.go                   # Main library: Client, NewProvider, NullProvider
├── command/                  # CLI commands (exported for external use)
│   ├── commands.go           # Root command definition
│   ├── helpers.go            # Config loading, option resolution, render dispatch
│   ├── price.go              # price command
│   ├── ticker.go             # ticker command
│   ├── book.go               # book command
│   ├── candles.go            # candles command
│   ├── info.go               # info command
│   ├── time.go               # time command
│   ├── markets.go            # markets command
│   ├── stream.go             # stream commands
│   └── providers.go          # providers, capabilities commands
├── cmd/bits/main.go          # CLI entry point
├── capability/               # Capability system
├── config/                   # Multi-provider config (YAML + Env + .env)
├── model/                    # Provider-agnostic data types
├── provider/                 # Provider interfaces + implementations
│   ├── binance/              # Binance implementation (spot + futures)
│   ├── bitget/               # Bitget implementation (spot + futures)
│   ├── coingecko/            # CoinGecko implementation
│   ├── whitebit/             # WhiteBit implementation
│   ├── cryptocom/            # Crypto.com implementation
│   ├── mexc/                 # MEXC implementation
│   └── registry/             # NewProvider factory
├── resolve/                  # Resolver, symbol resolution
├── render/                   # Output renderers (exported)
│   ├── renderer.go           # Format type + ParseFormat
│   ├── provenance.go         # FallbackFootnote, ProviderLabel
│   ├── table/                # Table renderers
│   ├── json/                 # JSON renderer
│   ├── yaml/                 # YAML renderer
│   ├── markdown/             # Markdown renderer
│   └── toon/                 # TOON renderer
├── process/                  # Data processors (exported)
│   ├── process.go            # Processor[T] type + Apply
│   ├── time.go                # TimeEnricher
│   ├── orderbook.go           # SpreadCalculator
│   └── candles.go            # CandleStats
├── internal/                 # CLI-only internals
│   ├── auth/                 # HMAC-SHA256 helpers
│   ├── logger/               # Structured logging
│   ├── tui/                  # Terminal UI components
│   └── ws/                   # WebSocket clients
└── examples/                 # Example applications
```