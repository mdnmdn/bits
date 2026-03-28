# Provider Implementation Guide

This guide explains how the multi-provider architecture works and how to add a new provider.

---

## Architecture Overview

```
cmd/                         ← command handlers
  └── factory.go             ← loadConfig(), newResolver(), flag helpers
        ↓
internal/resolve/
  ├── resolver.go            ← Resolve() — selects provider+market, handles fallback
  ├── require.go             ← Require[T] — type-asserts to a capability interface
  └── fanout.go              ← FanOut[T] — parallel multi-symbol calls
        ↓
internal/provider/           ← capability interfaces only (no implementations)
  ├── provider.go            ← Provider base interface
  ├── exchange.go            ← ExchangeProvider
  ├── aggregator.go          ← AggregatorProvider
  ├── capability.go          ← PriceProvider, CandleProvider, TickerProvider, OrderBookProvider
  ├── stream.go              ← PriceStreamProvider, OrderBookStreamProvider
  ├── binance/               ← exchange provider (spot + futures)
  ├── bitget/                ← exchange provider (spot + futures)
  └── coingecko/             ← aggregator + streaming provider
        ↓
internal/registry/
  └── registry.go            ← NewProvider() factory — separate pkg to avoid import cycle
        ↓
internal/model/              ← provider-agnostic data types
internal/capability/         ← MarketType / Feature matrix types (no project imports)
internal/ws/
  ├── base_client.go         ← reusable WebSocket client (reconnect + backoff)
  └── client.go              ← CoinGecko-specific streaming client (ActionCable protocol)
```

Commands **never** import a provider package directly. They always receive a `provider.Provider`
from `newResolver(cfg).Resolve(...)`, then type-assert to a capability interface via
`resolve.Require[T]`.

---

## Interface Hierarchy

All interfaces live in `internal/provider/`.

### Base — every provider must implement

```go
// provider.go
type Provider interface {
    ID() string
    SetUserAgent(string)
    Capabilities() capability.CapabilityMatrix
}
```

### Composite base interfaces

```go
// exchange.go — direct exchange APIs (Binance, Bitget)
type ExchangeProvider interface {
    Provider
    ServerTime(ctx context.Context) (model.Response[model.ServerTime], error)
    ExchangeInfo(ctx context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error)
}

// aggregator.go — market data aggregators (CoinGecko)
type AggregatorProvider interface {
    Provider
    CoinMarkets(ctx context.Context, opts model.MarketOpts) (model.Response[[]model.CoinMarket], error)
}
```

### Capability interfaces (capability.go)

```go
// Batch-native: all ids sent to provider in one call.
type PriceProvider interface {
    Price(ctx context.Context, ids []string, currency string) (model.Response[[]model.CoinPrice], error)
}

type CandleProvider interface {
    Candles(ctx context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) (model.Response[[]model.Candle], error)
}

// Single-symbol; resolver fans out for multi-symbol calls.
type TickerProvider interface {
    Ticker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error)
}

type OrderBookProvider interface {
    OrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (model.Response[model.OrderBook], error)
}
```

### Streaming interfaces (stream.go)

```go
type PriceStreamProvider interface {
    WatchPrices(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error)
}

type OrderBookStreamProvider interface {
    WatchOrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (<-chan *model.OrderBook, error)
}
```

### Capability matrix

Every provider declares its capabilities so `bits capabilities` can display the matrix:

```go
func (c *Client) Capabilities() capability.CapabilityMatrix {
    return capability.NewCapabilityMatrix(
        capability.CapabilityKey{Market: capability.MarketSpot,    Feature: capability.FeaturePrice},
        capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeaturePrice},
        // …
    )
}
```

---

## Package Layout

A provider lives in `internal/provider/<name>/`. Follow this file-per-concern convention:

```
internal/provider/<name>/
├── client.go          — Client struct, NewClient(), ID(), SetUserAgent(), Capabilities()
├── market.go          — PriceProvider, CandleProvider, TickerProvider, OrderBookProvider methods
├── exchange.go        — (optional) ExchangeProvider methods (ServerTime, ExchangeInfo)
├── types.go           — provider-specific JSON structs (internal; never exported to model)
└── stream.go          — (optional) streaming methods via ws.BaseClient
```

Additional files (`auth.go`, `trading.go`, etc.) are fine. Keep `client.go` focused on
construction, identity, and capabilities; keep API logic in `market.go` / `exchange.go`.

---

## Step-by-Step: Adding a New Provider

### 1. Config

Add a config struct in `internal/config/config.go`:

```go
type MyExchangeConfig struct {
    APIKey    string `mapstructure:"api_key"`
    APISecret string `mapstructure:"api_secret"`
    BaseURL   string `mapstructure:"base_url"`
}
```

Add it to the top-level `Config` struct and wire env overrides in `applyEnvOverrides()`:

```go
cfg.MyExchange.APIKey    = getEnv("BITS_MYEXCHANGE_API_KEY",    cfg.MyExchange.APIKey)
cfg.MyExchange.APISecret = getEnv("BITS_MYEXCHANGE_API_SECRET", cfg.MyExchange.APISecret)
```

### 2. `client.go` — struct, constructor, identity, capabilities

```go
package myexchange

import (
    "github.com/mdnmdn/bits/internal/capability"
    "github.com/mdnmdn/bits/internal/config"
)

const providerID = "myexchange"

type Client struct {
    cfg       config.MyExchangeConfig
    userAgent string
}

func NewClient(cfg config.MyExchangeConfig) *Client {
    return &Client{cfg: cfg}
}

func (c *Client) ID() string             { return providerID }
func (c *Client) SetUserAgent(ua string) { c.userAgent = ua }

func (c *Client) Capabilities() capability.CapabilityMatrix {
    s := capability.MarketSpot
    f := capability.MarketFutures
    return capability.NewCapabilityMatrix(
        capability.CapabilityKey{Market: s, Feature: capability.FeaturePrice},
        capability.CapabilityKey{Market: s, Feature: capability.FeatureCandles},
        capability.CapabilityKey{Market: s, Feature: capability.FeatureTicker24h},
        capability.CapabilityKey{Market: f, Feature: capability.FeaturePrice},
    )
}
```

### 3. `types.go` — internal JSON shapes

Define private structs for deserialising the exchange's API responses. Never expose them
outside the package; convert to `model.*` types in `market.go`.

```go
package myexchange

type apiTickerResponse struct {
    Symbol    string `json:"symbol"`
    LastPrice string `json:"lastPrice"`
    // …
}
```

### 4. `market.go` — implement capability interfaces

All methods return `model.Response[T]` with `Provider` and `Market` populated.

```go
func (c *Client) Price(ctx context.Context, ids []string, currency string) (model.Response[[]model.CoinPrice], error) {
    // call API, parse response, map to []model.CoinPrice
    return model.Response[[]model.CoinPrice]{
        Data:     prices,
        Provider: providerID,
        Market:   model.MarketSpot,
    }, nil
}

func (c *Client) Ticker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
    // …
    return model.Response[model.Ticker24h]{
        Data:     ticker,
        Provider: providerID,
        Market:   market,
    }, nil
}
```

Implement only the interfaces the exchange actually supports. Commands check at runtime via
`resolve.Require[T]`; unsupported interfaces cause a clean error or fallback.

### 5. `stream.go` — streaming (optional)

Use `ws.BaseClient` from `internal/ws/base_client.go`. Implement `OnConnect` (subscriptions)
and `OnMessage` (parsing), then push results onto a channel.

```go
package myexchange

import (
    "context"
    "encoding/json"

    "github.com/mdnmdn/bits/internal/model"
    "github.com/mdnmdn/bits/internal/ws"
)

// WatchOrderBook implements provider.OrderBookStreamProvider.
func (c *Client) WatchOrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
    updates := make(chan *model.OrderBook, 100)
    base := ws.NewBaseClient(buildStreamURL(symbol, depth))
    base.UserAgent = c.userAgent

    base.OnConnect = func(ctx context.Context, conn ws.Conn) error {
        return conn.WriteJSON(map[string]any{
            "method": "SUBSCRIBE",
            "params": []string{symbol + "@depth"},
            "id":     1,
        })
    }

    base.OnMessage = func(ctx context.Context, raw []byte) error {
        var msg apiDepthMessage
        if err := json.Unmarshal(raw, &msg); err != nil {
            return err
        }
        select {
        case updates <- convertOrderBook(msg, market):
        case <-ctx.Done():
        }
        return nil
    }

    if err := base.Connect(ctx); err != nil {
        return nil, err
    }
    go func() {
        <-ctx.Done()
        base.Close()
        close(updates)
    }()
    return updates, nil
}
```

> **`ws.BaseClient` vs `ws.Client`**
>
> | | `ws.BaseClient` | `ws.Client` |
> |---|---|---|
> | Purpose | Generic reconnecting WS client | CoinGecko-specific (ActionCable + auth) |
> | Reconnect | Built-in | Built-in (independent copy) |
> | Usage | Binance, new providers | CoinGecko only |
>
> **Always use `ws.BaseClient` for new providers.** `ws.Client` is CoinGecko-specific.
> Implement protocol handshakes in `OnConnect` — it is re-called on every reconnect.

### 6. Register the provider

In `internal/registry/registry.go`, add a case to `NewProvider` and `AllProviderIDs`:

```go
func NewProvider(name string, cfg *config.Config) (provider.Provider, error) {
    switch name {
    case "coingecko", "":
        return coingecko.NewClient(cfg), nil
    case "binance":
        return binance.NewClient(cfg.Binance), nil
    case "bitget":
        return bitget.NewClient(cfg.Bitget), nil
    case "myexchange":
        return myexchange.NewClient(cfg.MyExchange), nil
    default:
        return nil, fmt.Errorf("unknown provider %q", name)
    }
}

func AllProviderIDs() []string {
    return []string{"coingecko", "binance", "bitget", "myexchange"}
}
```

---

## Data Model

All public method signatures use types from `internal/model/`:

| Type | Description |
|------|-------------|
| `Response[T]` | Generic envelope: `Data T`, `Provider`, `Market`, `Fallback`, `Errors` |
| `MarketType` | `"spot"` \| `"futures"` \| `"margin"` |
| `CoinPrice` | Current price (coin ID or symbol, price, currency, optional 24h change) |
| `Ticker24h` | 24h rolling stats (last price, change %, high, low, volume — all optional floats) |
| `Candle` | OHLCV candle with optional volume and close time |
| `OrderBook` | Bid/ask depth snapshot (`[]OrderBookEntry{Price, Quantity}`) |
| `ServerTime` | Exchange server timestamp with optional latency/skew |
| `ExchangeInfo` | Exchange symbol catalogue (`[]Symbol`) |
| `CoinMarket` | Ranked coin listing with market metadata |

Parse string-encoded floats from exchange APIs with `strconv.ParseFloat(s, 64)`.
Store provider-specific fields that don't map to model types in `Extra map[string]any`.

---

## Testing Patterns

Look at `internal/provider/binance/client_test.go` and `internal/provider/coingecko/client_test.go`
for the established pattern:

- Use `net/http/httptest.NewServer` to mock HTTP responses.
- Use `testify/assert` for assertions.
- Test each capability method in isolation.
- For market-type branching, write one test per market type.

For WebSocket methods, test `OnMessage` in isolation by calling it directly with fixture JSON,
without spinning up a real WebSocket server.

---

## Quick Checklist

```
□ internal/config/config.go            — add MyExchangeConfig, wire env vars
□ internal/provider/myexchange/
  □ client.go                          — Client struct, NewClient, ID, SetUserAgent, Capabilities
  □ types.go                           — internal JSON response structs
  □ market.go                          — PriceProvider, CandleProvider, TickerProvider, etc.
  □ exchange.go                        — (optional) ExchangeProvider (ServerTime, ExchangeInfo)
  □ stream.go                          — (optional) streaming via ws.NewBaseClient
□ internal/registry/registry.go        — add case to NewProvider and AllProviderIDs
□ _docs/architecture.md                — update capability table
□ _docs/features.md                    — note new provider in relevant commands
```

Run `go build ./...` and `go test ./...` after each file to catch issues early.
