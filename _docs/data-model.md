# Data Model — Multi-Provider Interface

## Provider Categories

Two distinct provider archetypes, both implementing the base `Provider` interface:

```
Provider (base: ID, SetUserAgent, Capabilities)
├── ExchangeProvider     — direct exchange (Binance, Bitget, ...)
└── AggregatorProvider   — market data aggregator (CoinGecko, CoinMarketCap, ...)
```

Capability interfaces are shared and may be implemented by either category.

---

## Market Types

All market-aware interfaces accept a `MarketType` parameter. The existing
`capability.MarketType` is the canonical type:

```go
// already defined in internal/capability/capability.go
type MarketType string  // "spot" | "futures" | "margin"
```

When a provider does not support a requested market type, it returns
`model.ErrUnsupportedMarket` (or the caller checks capabilities first).

---

## Provider Configuration

The existing config and registry structure is kept as-is. Each provider defines its
own typed config struct in its own package; `internal/config/config.go` composes them
into a single root `Config`. The registry signature `NewProvider(name string, cfg *config.Config)`
is unchanged.

```go
// internal/config/config.go
type Config struct {
    Provider  string         // active provider id
    CoinGecko coingecko.Config
    Binance   binance.Config
    Bitget    bitget.Config
}

// internal/provider/registry.go — unchanged
func NewProvider(name string, cfg *config.Config) (Provider, error)
```

Each provider config struct lives in its own package and is responsible for its own
field definitions and env var overrides (e.g. `BITS_BINANCE_API_KEY`). Adding a new
provider means adding a new field to `Config` and a new case in `NewProvider` —
a minimal, contained change.

---

## Base Interface

```go
type Provider interface {
    ID()             string
    SetUserAgent(string)
    Capabilities()   capability.CapabilityMatrix
}
```

---

## Exchange Provider

Direct exchange APIs (Binance, Bitget). Adds server time and symbol catalogue.

```go
type ExchangeProvider interface {
    Provider
    ServerTime(ctx context.Context) (*model.ServerTime, error)
    ExchangeInfo(ctx context.Context, market model.MarketType) (*model.ExchangeInfo, error)
}
```

---

## Aggregator Provider

Multi-exchange data aggregators (CoinGecko, CoinMarketCap). Works with coin IDs,
aggregates across markets. Market type is optional/advisory for aggregators.

```go
type AggregatorProvider interface {
    Provider
    // Market listings — spot-centric but currency-aware
    CoinMarkets(ctx context.Context, opts model.MarketOpts) ([]model.CoinMarket, error)
}
```

---

## Shared Capability Interfaces

Any provider may implement these. All market-sensitive calls carry `MarketType`.

```go
// Price — aggregators use coin IDs, exchanges use symbols
type PriceProvider interface {
    Price(ctx context.Context, ids []string, currency string) ([]model.CoinPrice, error)
}

// OHLCV candles
type CandleProvider interface {
    Candles(ctx context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) ([]model.Candle, error)
}

// 24h rolling ticker
type TickerProvider interface {
    Ticker24h(ctx context.Context, symbol string, market model.MarketType) (*model.Ticker24h, error)
}

// Order book depth
type OrderBookProvider interface {
    OrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (*model.OrderBook, error)
}
```

> Other capability interfaces (Search, Trending, Historical, Detail, GainersLosers)
> are out of scope for the initial migration — they remain in their current form.

---

## Streaming Interfaces

Streaming interfaces use the client as a stateful object that manages the WebSocket connection.
All methods return a single channel where updates for all subscribed symbols flow.

```go
type StreamStatus string

const (
    StreamStatusRunning StreamStatus = "running"
    StreamStatusStopped StreamStatus = "stopped"
    StreamStatusError   StreamStatus = "error"
)

type PriceStreamProvider interface {
    // StartPriceStream initiates a price stream for multiple symbols.
    // Returns a single channel where all price updates flow.
    StartPriceStream(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error)

    // SubscribePrice adds new symbols to an existing price stream.
    SubscribePrice(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error)

    // UnsubscribePrice removes symbols from the price stream.
    UnsubscribePrice(ctx context.Context, ids []string) error

    // SubscribedPrices returns the list of currently subscribed symbol IDs.
    SubscribedPrices() []string

    // StopPriceStream stops all price streams and closes all channels.
    StopPriceStream() error

    // PriceStreamStatus returns the current status of the price stream.
    PriceStreamStatus() StreamStatus

    // GetLastPrice returns the last received price for a symbol.
    GetLastPrice(id string) (*model.CoinPrice, error)

    // ReconnectPrice reconnects the price stream.
    ReconnectPrice(ctx context.Context) error

    // GetDataChannelPrice returns the current price channel.
    GetDataChannelPrice() <-chan *model.CoinPrice
}

type OrderBookStreamProvider interface {
    // StartOrderBookStream initiates an order book stream for multiple symbols.
    // Returns a single channel where all order book updates flow.
    StartOrderBookStream(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error)

    // SubscribeOrderBook adds new symbols to an existing order book stream.
    SubscribeOrderBook(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error)

    // UnsubscribeOrderBook removes symbols from the order book stream.
    UnsubscribeOrderBook(ctx context.Context, symbols []string) error

    // SubscribedOrderBooks returns the list of currently subscribed symbols.
    SubscribedOrderBooks() []string

    // StopOrderBookStream stops all order book streams and closes all channels.
    StopOrderBookStream() error

    // OrderBookStreamStatus returns the current status of the order book stream.
    OrderBookStreamStatus() StreamStatus

    // GetLastOrderBook returns the last received order book for a symbol.
    GetLastOrderBook(symbol string) (*model.OrderBook, error)

    // ReconnectOrderBook reconnects the order book stream.
    ReconnectOrderBook(ctx context.Context) error

    // GetDataChannelOrderBook returns the current order book channel.
    GetDataChannelOrderBook() <-chan *model.OrderBook
}
```

---

## Data Model

### Design Conventions

- Optional values use pointer types (`*float64`, `*time.Time`, `*int`).
- Provider-specific or unmapped data is preserved in `Extra map[string]any`.
- `MarketType` is a string alias — same type as `capability.MarketType`, re-exported from `model` for convenience.

```go
// model/market.go
type MarketType = capability.MarketType

const (
    MarketSpot    MarketType = capability.MarketSpot
    MarketFutures MarketType = capability.MarketFutures
    MarketMargin  MarketType = capability.MarketMargin
)
```

---

### ServerTime

```go
type ServerTime struct {
    Time  time.Time
    Extra map[string]any
}
```

---

### ExchangeInfo

```go
type SymbolStatus string

const (
    SymbolStatusTrading SymbolStatus = "trading"
    SymbolStatusBreak   SymbolStatus = "break"
    SymbolStatusHalt    SymbolStatus = "halt"
)

type Symbol struct {
    Symbol         string
    BaseAsset      string        // e.g. "BTC"
    QuoteAsset     string        // e.g. "USDT"
    Status         SymbolStatus
    Market         MarketType
    PricePrecision *int
    QtyPrecision   *int
    MinPrice       *float64
    MaxPrice       *float64
    MinQty         *float64
    MaxQty         *float64
    StepSize       *float64      // quantity increment
    MakerFee       *float64
    TakerFee       *float64
    Extra          map[string]any
}

type ExchangeInfo struct {
    ExchangeID string
    Market     MarketType
    ServerTime *time.Time
    Symbols    []Symbol
    Extra      map[string]any
}
```

---

### Candle (OHLCV)

Replaces `OHLCData [][]float64`. Explicit fields, optional volume and close time.

```go
type Candle struct {
    OpenTime  time.Time
    Open      float64
    High      float64
    Low       float64
    Close     float64
    Volume    *float64    // base asset volume; absent in some aggregators
    CloseTime *time.Time  // absent in some providers
    Extra     map[string]any
}

type CandleOpts struct {
    From  *time.Time
    To    *time.Time
    Limit *int
}
```

---

### Ticker24h

Current type refined: optional fields become pointers; `Extra` added.

```go
type Ticker24h struct {
    Symbol             string
    Market             MarketType
    LastPrice          float64
    PriceChange        *float64
    PriceChangePercent *float64
    HighPrice          *float64
    LowPrice           *float64
    Volume             *float64    // base asset volume
    QuoteVolume        *float64
    OpenPrice          *float64
    WeightedAvgPrice   *float64
    BidPrice           *float64
    AskPrice           *float64
    OpenTime           *time.Time
    CloseTime          *time.Time
    Extra              map[string]any
}
```

---

### OrderBook

Current type refined: optional update ID and timestamp; `Market`; `Extra`.

```go
type OrderBook struct {
    Symbol       string
    Market       MarketType
    Bids         []OrderBookEntry
    Asks         []OrderBookEntry
    LastUpdateID *int64
    Time         *time.Time
    Extra        map[string]any
}

type OrderBookEntry struct {
    Price    float64
    Quantity float64
}
```

---

### CoinPrice

Replaces `PriceResponse map[string]map[string]float64`.

```go
type CoinPrice struct {
    ID        string     // coin id (aggregators) or trading symbol (exchanges)
    Symbol    string
    Currency  string
    Price     float64
    Change24h *float64   // percent; optional
    Extra     map[string]any
}
```

---

### CoinMarket

Refines `MarketCoin`. Optional fields; `Currency` made explicit.

```go
type CoinMarket struct {
    ID                string
    Symbol            string
    Name              string
    Currency          string
    Price             float64
    MarketCap         *float64
    MarketCapRank     *int
    Volume24h         *float64
    PriceChangePct24h *float64
    High24h           *float64
    Low24h            *float64
    Extra             map[string]any
}

type MarketOpts struct {
    Currency string
    PerPage  int
    Page     int
    Order    string
    Category string
}
```

---

## CLI Command Structure

### Global Flags

Applied to every command, resolved before the command runs:

```
-p, --provider  string    provider id: coingecko | binance | bitget  (default: from config)
-m, --market    string    market type: spot | futures | margin        (default: spot)
-o, --output    string    output format: table | json | csv           (default: table)
```

Resolution order for `--provider`: flag → `BITS_PROVIDER` env → config file → `"coingecko"`.

---

### Commands

```
bits time                           Server timestamp
bits info [--symbol <sym>]          Exchange symbol catalogue (filterable)
bits price <id|symbol>...           Current price(s); --currency usd
bits candles <symbol>               OHLCV history; --interval 1h; --from/--to; --limit N
bits ticker <symbol>                24h rolling statistics
bits book <symbol> [--depth N]      Order book snapshot (default depth: 20)
bits markets [--currency usd]       Ranked coin listing (aggregators only)

bits stream price <id|symbol>...    Live price feed
bits stream book <symbol>           Live order book feed

bits providers                      List registered providers and their IDs
bits capabilities [--provider id]   Capability matrix (existing command, kept)
```

Examples:

```sh
bits price BTC ETH                              # CoinGecko (default), USD
bits price BTCUSDT -p binance                   # Binance spot price
bits price BTCUSDT -p binance -m futures        # Binance futures price

bits candles BTCUSDT -p binance --interval 1h
bits candles BTCUSDT -p binance -m futures --from 2024-01-01 --to 2024-02-01

bits ticker BTCUSDT -p bitget
bits ticker BTCUSDT -p binance -m futures

bits book BTCUSDT -p binance --depth 50
bits book BTCUSDT -p binance -m futures

bits info -p binance --symbol BTCUSDT
bits info -p binance -m futures

bits time -p binance

bits markets --currency eur -p coingecko

bits stream price BTC ETH -p coingecko
bits stream book BTCUSDT -p binance
```

---

### Command → Interface Mapping

Each command declares a required interface. If the selected provider does not implement it,
the command fails immediately with a clear error — no fallback, no silent degradation.

| Command | Required Interface | Market-aware |
|---|---|:---:|
| `bits time` | `ExchangeProvider` | — |
| `bits info` | `ExchangeProvider` | Yes |
| `bits price` | `PriceProvider` | — |
| `bits candles` | `CandleProvider` | Yes |
| `bits ticker` | `TickerProvider` | Yes |
| `bits book` | `OrderBookProvider` | Yes |
| `bits markets` | `AggregatorProvider` | — |
| `bits stream price` | `PriceStreamProvider` | — |
| `bits stream book` | `OrderBookStreamProvider` | Yes |

---

### Command Framework Pattern

Commands use a typed capability helper instead of scattered type assertions:

```go
// cmd/require.go
func require[T any](p provider.Provider, feature string) (T, error) {
    v, ok := p.(T)
    if !ok {
        var zero T
        return zero, fmt.Errorf("provider %q does not support %s", p.ID(), feature)
    }
    return v, nil
}
```

Usage in a command:

```go
func runTicker(cmd *cobra.Command, args []string) error {
    client, market := resolveProvider(cmd)   // reads -p and -m flags
    tp, err := require[provider.TickerProvider](client, "ticker")
    if err != nil {
        return err
    }
    ticker, err := tp.Ticker24h(cmd.Context(), args[0], market)
    // ...
}
```

This keeps all type assertion logic in one place and makes command code flat and readable.

---

### File Structure (target `cmd/`)

```
cmd/
├── root.go              # global flags, provider resolution, output flag
├── require.go           # require[T] helper + resolveProvider
├── time.go              # bits time
├── info.go              # bits info
├── price.go             # bits price
├── candles.go           # bits candles
├── ticker.go            # bits ticker
├── book.go              # bits book
├── markets.go           # bits markets
├── stream.go            # bits stream price / book (subcommands)
├── providers.go         # bits providers
└── capabilities.go      # bits capabilities (existing, kept)
```

---

## Response Envelope

Every provider call returns a typed `Response[T]` wrapper. This carries provenance
(which provider and market actually served the data) so the rendering layer can surface
fallback information without any out-of-band signalling.

```go
// Response[T] wraps any provider result with its provenance.
type Response[T any] struct {
    Data              T
    Provider          string     // provider that actually served this response
    Market            MarketType // market that actually served this response
    Fallback          bool       // true when a different provider was selected automatically
    RequestedProvider string     // original request; populated only when Fallback is true
    RequestedMarket   MarketType // original request; populated only when Fallback is true
    Errors            []ItemError // partial failures in batch/multi-symbol calls
}

// ItemError pairs a symbol (or id) with the error that occurred for it.
type ItemError struct {
    Symbol string
    Err    error
}
```

Renderers receive a `Response[T]` and may use provenance freely — e.g., appending a
`[fallback: binance→coingecko]` footnote in table mode, or including `"provider"` and
`"fallback"` keys in JSON output.

---

## Provider Resolution & Fallback

### Resolution Order

```
--provider flag  →  BITS_PROVIDER env  →  config file  →  "coingecko" default
```

### Fallback Behaviour

When the resolved provider does not support the requested feature, the resolver
tries the next registered provider that does.

```go
type ResolutionOpts struct {
    Provider string      // explicit override ("" = use config/default)
    Market   MarketType  // explicit market ("" = spot)
    Lock     bool        // if true, disable fallback — return error instead
}
```

- **Default (Lock=false):** transparent fallback to the first capable provider.
  `Response.Fallback` is set to `true` so the rendering layer can note it.
- **Locked (Lock=true):** no fallback. If the requested provider+market does not
  support the feature, the command fails immediately with a descriptive error.

CLI surface:

```
--lock   / -l     disable provider fallback; error if capability is missing
```

Examples:

```sh
bits ticker BTCUSDT -p coingecko          # coingecko lacks ticker → falls back to binance
bits ticker BTCUSDT -p coingecko --lock   # error: coingecko does not support ticker
bits candles BTCUSDT -p binance -m margin # margin not supported → falls back to spot (or error with --lock)
```

---

## Multi-Symbol Calls

Commands that accept identifiers take one or more positional arguments. Interfaces
that are naturally batched (e.g. `PriceProvider`) pass the full slice to the provider.
Interfaces that are inherently single-symbol (e.g. `TickerProvider`) are fanned out
by the resolver layer — one call per symbol, results collected into a single
`Response[[]Ticker24h]`.

```go
// Batch-native: provider receives all ids at once
type PriceProvider interface {
    Price(ctx context.Context, ids []string, currency string) (Response[[]CoinPrice], error)
}

// Single-symbol: resolver fans out automatically
type TickerProvider interface {
    Ticker24h(ctx context.Context, symbol string, market MarketType) (Response[Ticker24h], error)
}
// → caller receives Response[[]Ticker24h] assembled by the resolver
```

Partial failures (one symbol errors, others succeed) are collected in
`Response.Errors` rather than failing the whole call. The renderer decides
how to display them (inline warning, footnote, separate error block).

---

## Rendering Architecture

Rendering is completely decoupled from providers. A command fetches data, obtains
a `Response[T]`, and hands it to a `Renderer`. No provider types or interfaces leak
into the rendering layer.

```go
type Format string

const (
    FormatTable    Format = "table"
    FormatJSON     Format = "json"
    FormatMarkdown Format = "markdown"
    FormatYAML     Format = "yaml"
    FormatToon     Format = "toon"   // ASCII / decorative
)

type Renderer[T any] interface {
    Render(w io.Writer, res Response[T]) error
}
```

A renderer registry maps `(type, format)` → `Renderer`:

```go
// internal/render/registry.go
func Render[T any](w io.Writer, format Format, res Response[T]) error
```

Each data type registers its own renderers. The command only selects the format:

```go
// in cmd/ticker.go
res, err := resolver.Ticker(ctx, symbols, opts)
return render.Render(os.Stdout, outputFormat, res)
```

### What each format does with provenance

| Format | Fallback display |
|---|---|
| `table` | footnote: `† served by binance (requested: coingecko)` |
| `json` | top-level `"provider"`, `"market"`, `"fallback"` keys |
| `markdown` | blockquote note below the table |
| `yaml` | top-level `provider:`, `market:`, `fallback:` keys |
| `toon` | inline label or colour accent |

---

## Processing Layer

An optional intermediate layer sits between the provider response and the renderer.
Processors receive a `Response[T]` and return an enriched `Response[T]` of the same
type (using `Extra` for computed fields, or optional typed fields on the struct).

```go
type Processor[T any] func(Response[T]) Response[T]
```

Processors are composable:

```go
func Process[T any](res Response[T], processors ...Processor[T]) Response[T] {
    for _, p := range processors {
        res = p(res)
    }
    return res
}
```

### Example: time latency and skew

`ServerTime` gains optional computed fields populated by a processor:

```go
type ServerTime struct {
    Time      time.Time
    LocalTime *time.Time     // set by processor: time.Now() at receipt
    Latency   *time.Duration // LocalTime - Time (round-trip approximation)
    ClockSkew *time.Duration // estimated difference: Latency / 2
    Extra     map[string]any
}

// TimeEnricher is a Processor[ServerTime]
func TimeEnricher(res Response[ServerTime]) Response[ServerTime] {
    now := time.Now()
    latency := now.Sub(res.Data.Time)
    skew := latency / 2
    res.Data.LocalTime = &now
    res.Data.Latency = &latency
    res.Data.ClockSkew = &skew
    return res
}
```

The `toon` renderer for `ServerTime` can then display a clock diagram with skew.
The `table` renderer shows latency as a column. The `json` renderer includes all fields.

### Other processor examples

| Processor | Input type | What it computes |
|---|---|---|
| `TimeEnricher` | `ServerTime` | latency, clock skew |
| `SpreadCalculator` | `OrderBook` | bid-ask spread, mid price |
| `PriceChangeClassifier` | `[]CoinMarket` | adds `"bullish"/"bearish"` label to `Extra` |
| `CandleStats` | `[]Candle` | VWAP, typical price, body/wick ratios |

Processors are applied by the command before handing off to the renderer.
Commands choose which processors to apply; renderers remain unaware.

---

## Provider × Interface Matrix (target)

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

> Capabilities are declared at runtime via the existing `CapabilityMatrix` system.
> Use `bits capabilities` to inspect them. The matrix above is the desired end state.
