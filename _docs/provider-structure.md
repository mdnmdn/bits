# Provider Implementation Guide

This guide explains how the multi-provider architecture works and how to add a new provider.

---

## Architecture Overview

```
cmd/                         ← command handlers
  └── client_factory.go      ← creates providers via registry; resolves --provider flag
        ↓
internal/provider/
  ├── types.go               ← core interfaces (Provider, TickerProvider, …)
  ├── capabilities.go        ← re-exports from internal/capability
  ├── registry.go            ← NewProvider() factory, AllCapabilities()
  ├── coingecko/             ← full-featured REST + streaming provider
  ├── binance/               ← exchange provider (spot + futures)
  └── bitget/                ← exchange provider (spot + futures)
        ↓
internal/model/              ← provider-agnostic data types (PriceResponse, OHLCData, …)
internal/capability/         ← MarketType / Feature matrix types (no project imports)
internal/ws/
  ├── base_client.go         ← reusable WebSocket client (reconnect + backoff)
  └── client.go              ← CoinGecko-specific streaming client (protocol-specific)
```

Commands **never** import a provider package directly. They always receive a `provider.Provider`
(or a capability-checked sub-interface) from `newAPIClient`.

---

## Interface Hierarchy

All interfaces live in `internal/provider/types.go`.

### Required — every provider must implement this

```go
type Provider interface {
    ID() string                                                              // e.g. "binance"
    SetUserAgent(userAgent string)
    SimplePrice(ctx, ids []string, vsCurrency string) (PriceResponse, error)
    CoinOHLC(ctx, id, vsCurrency, days, interval string) (OHLCData, error)
}
```

### Optional REST capabilities

| Interface | Methods | Who |
|-----------|---------|-----|
| `SymbolPricer` | `SimplePriceBySymbols` | CoinGecko |
| `MarketLister` | `CoinMarkets`, `FetchAllMarkets` | CoinGecko |
| `Searcher` | `Search` | CoinGecko |
| `TrendingProvider` | `SearchTrending` | CoinGecko |
| `HistoricalProvider` | `CoinHistory`, `CoinMarketChart`, `CoinMarketChartRange`, `CoinOHLCRange` | CoinGecko |
| `GainersLosersProvider` | `TopGainersLosers` | CoinGecko |
| `DetailProvider` | `CoinDetail` | CoinGecko |
| `TickerProvider` | `Ticker24h` | Binance, Bitget |
| `OrderBookProvider` | `OrderBook` | Binance |

### Optional streaming capabilities

| Interface | Methods | Who |
|-----------|---------|-----|
| `OrderBookStreamProvider` | `WatchOrderBook` | Binance |
| `PriceStreamProvider` | `WatchPrices` | CoinGecko (`ws.Client`) |

Streaming providers are distinct from REST providers. Commands get them via `newStreamer` (for
`PriceStreamProvider`) or by a direct type assertion on the REST client (for `OrderBookStreamProvider`).

### Capability declaration (always implement)

```go
type CapabilityProvider interface {      // in internal/capability
    Capabilities() CapabilityMatrix      // map[CapabilityKey]bool
}
```

Every new provider must implement `Capabilities()` so that `bits capabilities` can display the
matrix. See `internal/capability/capability.go` for the full list of `Feature` and `MarketType` constants.

---

## Package Layout

A provider lives in `internal/provider/<name>/`. Follow this file-per-concern convention:

```
internal/provider/<name>/
├── client.go          — Client struct, NewClient(), ID(), SetUserAgent()
├── market.go          — SimplePrice(), CoinOHLC(), and any optional REST methods
├── types.go           — provider-specific JSON structs (internal; not exported to model)
├── capabilities.go    — Capabilities() method
└── ws.go              — (optional) streaming methods that use ws.BaseClient
```

Additional files for trading (`trading.go`), auth helpers (`auth.go`), etc. are fine. Keep
`client.go` focused on construction and identity; keep API logic in `market.go`.

---

## Step-by-Step: Adding a New Provider

### 1. Config

Add a config struct in `internal/config/config.go`:

```go
type MyExchangeConfig struct {
    APIKey     string `mapstructure:"api_key"`
    APISecret  string `mapstructure:"api_secret"`
    BaseURL    string `mapstructure:"base_url"`
    MarketType string `mapstructure:"market_type"`
    UseTestnet bool   `mapstructure:"use_testnet"`
}
```

Add it to the top-level `Config` struct and wire env overrides in `applyEnvOverrides()`:

```go
cfg.MyExchange.APIKey    = getEnv("BITS_MYEXCHANGE_API_KEY",    cfg.MyExchange.APIKey)
cfg.MyExchange.APISecret = getEnv("BITS_MYEXCHANGE_API_SECRET", cfg.MyExchange.APISecret)
```

Also add a `myexchange:` block to the example config section in the same file.

### 2. `client.go` — struct + constructor

```go
package myexchange

import (
    "github.com/mdnmdn/bits/internal/config"
)

const providerID = "myexchange"

type Client struct {
    config     config.MyExchangeConfig
    marketType string
    userAgent  string
    // http.Client, SDK client, etc.
}

func NewClient(cfg config.MyExchangeConfig) *Client {
    mt := cfg.MarketType
    if mt == "" {
        mt = config.MarketTypeSpot
    }
    return &Client{
        config:     cfg,
        marketType: mt,
    }
}

func (c *Client) ID() string             { return providerID }
func (c *Client) SetUserAgent(ua string) { c.userAgent = ua }
```

### 3. `types.go` — internal JSON shapes

Define private structs for deserializing the exchange's API responses. Never expose these outside
the package; convert to `model.*` types in `market.go`.

```go
package myexchange

type apiTickerResponse struct {
    Symbol    string `json:"symbol"`
    LastPrice string `json:"lastPrice"`
    // …
}
```

### 4. `market.go` — implement the Provider interface

`SimplePrice` and `CoinOHLC` are mandatory. Return `model.*` types, not your internal types.
Branch on `c.marketType` when behaviour differs between spot and futures.

```go
func (c *Client) SimplePrice(ctx context.Context, ids []string, vsCurrency string) (model.PriceResponse, error) {
    result := make(model.PriceResponse, len(ids))
    vs := strings.ToLower(vsCurrency)
    for _, id := range ids {
        // call API, parse response, map to model.PriceResponse
        result[id] = map[string]float64{vs: price}
    }
    return result, nil
}

func (c *Client) CoinOHLC(ctx context.Context, id, vsCurrency, days, interval string) (model.OHLCData, error) {
    // Return [][]float64 where each element is [timestamp, open, high, low, close]
}
```

Implement optional interfaces (`TickerProvider`, `OrderBookProvider`, …) only when the exchange
supports them. Commands check at runtime:

```go
tp, ok := client.(provider.TickerProvider)
if !ok {
    return fmt.Errorf("%s provider does not support ticker data", client.ID())
}
```

### 5. `capabilities.go` — declare the capability matrix

Import `internal/capability` (not `internal/provider` — that would create an import cycle).

```go
package myexchange

import "github.com/mdnmdn/bits/internal/capability"

func (c *Client) Capabilities() capability.CapabilityMatrix {
    s := capability.MarketSpot
    f := capability.MarketFutures
    return capability.NewCapabilityMatrix(
        capability.CapabilityKey{Market: s, Feature: capability.FeaturePrice},
        capability.CapabilityKey{Market: s, Feature: capability.FeatureCandles},
        capability.CapabilityKey{Market: s, Feature: capability.FeatureTicker24h},
        // add futures entries if supported
        capability.CapabilityKey{Market: f, Feature: capability.FeaturePrice},
    )
}
```

### 6. `ws.go` — streaming (optional)

Use `ws.BaseClient` from `internal/ws/base_client.go`. It handles dial, reconnect, and
exponential backoff with jitter. Implement only `OnConnect` (subscriptions) and `OnMessage`
(parsing), then push results onto a channel.

```go
package myexchange

import (
    "context"
    "encoding/json"

    "github.com/mdnmdn/bits/internal/model"
    "github.com/mdnmdn/bits/internal/ws"
)

// WatchOrderBook implements provider.OrderBookStreamProvider.
func (c *Client) WatchOrderBook(ctx context.Context, symbols []string, limit int) (<-chan *model.OrderBook, error) {
    url := buildStreamURL(symbols, limit, c.config.UseTestnet)
    updates := make(chan *model.OrderBook, 100)

    base := ws.NewBaseClient(url)
    base.UserAgent = c.userAgent

    // OnConnect: send subscription message(s) after dial.
    base.OnConnect = func(ctx context.Context, conn *websocket.Conn) error {
        return conn.WriteJSON(map[string]any{
            "method": "SUBSCRIBE",
            "params": buildSubscriptionParams(symbols),
            "id":     1,
        })
    }

    // OnMessage: parse raw bytes, push to channel.
    base.OnMessage = func(ctx context.Context, raw []byte) error {
        var msg apiDepthMessage
        if err := json.Unmarshal(raw, &msg); err != nil {
            return err
        }
        ob := convertOrderBook(msg)
        select {
        case updates <- ob:
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

> **Note on `ws.BaseClient` vs `ws.Client`**
>
> Two WebSocket implementations exist in `internal/ws/`:
>
> | | `ws.BaseClient` | `ws.Client` |
> |---|---|---|
> | Purpose | Generic reconnecting WS client | CoinGecko-specific (ActionCable + auth) |
> | Reconnect logic | Built-in | Built-in (independent copy) |
> | Usage | Binance `ws.go`, new providers | `cmd/watch.go` only |
>
> **Always use `ws.BaseClient` for new providers.** It is the shared infrastructure.
> `ws.Client` is protocol-specific to CoinGecko and should not be used as a template.
> If a new provider needs a protocol handshake on connect (auth, subscription), implement
> it in `base.OnConnect` — `BaseClient` calls this callback after every dial (including
> reconnects), so the handshake is re-executed automatically.

### 7. Register the provider

In `internal/provider/registry.go`, add to `AvailableProviders` and to `NewProvider`:

```go
var AvailableProviders = []string{"coingecko", "binance", "bitget", "myexchange"}

func NewProvider(name string, cfg *config.Config) (Provider, error) {
    switch name {
    // … existing cases …
    case "myexchange":
        return myexchange.NewClient(cfg.MyExchange), nil
    default:
        return nil, fmt.Errorf("unknown provider: %s", name)
    }
}
```

Also add the import at the top of the file.

---

## Data Model Rules

- All public method signatures must use types from `internal/model`, never internal API types.
- `model.PriceResponse` is `map[string]map[string]float64` — outer key is coin ID or symbol,
  inner key is currency (lowercase), with an optional `<currency>_24h_change` sibling key.
- `model.OHLCData` is `[][]float64` where each row is `[timestamp_ms, open, high, low, close]`.
- Parse string-encoded floats from exchange APIs with `strconv.ParseFloat(s, 64)`.

---

## Testing Patterns

Look at `internal/provider/binance/client_test.go` and `internal/provider/coingecko/client_test.go`
for the established pattern:

- Use `net/http/httptest.NewServer` to mock HTTP responses.
- Use `testify/assert` for assertions.
- Test `SimplePrice`, `CoinOHLC`, and any optional interfaces in isolation.
- For market-type branching, write one test per market type.

For WebSocket methods, test `OnMessage` in isolation by calling it directly with fixture JSON,
without spinning up a real WebSocket server.

---

## Quick Checklist

```
□ internal/config/config.go       — add MyExchangeConfig, wire env vars
□ internal/provider/myexchange/
  □ client.go                     — Client struct, NewClient, ID, SetUserAgent
  □ types.go                      — internal JSON response structs
  □ market.go                     — SimplePrice, CoinOHLC (+ optional interfaces)
  □ capabilities.go               — Capabilities() using internal/capability
  □ ws.go                         — (optional) streaming via ws.NewBaseClient
□ internal/provider/registry.go   — add to AvailableProviders and NewProvider
□ _docs/architecture.md           — update capability table
□ _docs/features.md               — note new provider in relevant commands
```

Run `go build ./...` and `go test ./...` after each file to catch issues early.
