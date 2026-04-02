# Provider Implementation Guide

> ⚠️ **Status: Work in Progress** — The library API may change rapidly.

This guide explains how the multi-provider architecture works and how to add a new provider.
Work through the steps in order; run `go build ./...` after each file to catch issues early.

---

## Reference Documents

| Document | Purpose |
|----------|---------|
| `_docs/architecture.md` | High-level architecture, interface hierarchy, fallback policy |
| `_docs/symbol-engine.md` | Symbol resolution, normalization, disk caching |
| `_docs/data-model.md` | All model types (`Response[T]`, `Ticker24h`, `OrderBook`, …) |
| `_docs/config.md` | Config file format, env var conventions, viper setup |
| `_docs/ws-handling.md` | WebSocket infrastructure (`ws.Manager`, `ws.BaseClient`) |
| `_docs/providers/` | Per-provider API docs (endpoints, auth, response formats) |

Reference implementations:

| Provider | Location | Notable features |
|----------|----------|-----------------|
| Bitget | `provider/bitget/` | Raw HTTP, HMAC auth, spot + futures + margin |
| WhiteBit | `provider/whitebit/` | Raw HTTP, no auth for public, streaming |
| Binance | `provider/binance/` | Go SDK wrapper, spot + futures, streaming |
| CoinGecko | `provider/coingecko/` | Aggregator, paid/demo tiers, streaming |
| Crypto.com | `provider/cryptocom/` | Raw HTTP, spot only, streaming |
| MEXC | `provider/mexc/` | Raw HTTP, spot only, protobuf parsing |

---

## Architecture Overview

```
cmd/                         ← command handlers
  └── factory.go             ← loadConfig(), newResolver(), flag helpers
        ↓
resolve/
  ├── resolver.go            ← Resolve() — selects provider+market, handles fallback
  ├── require.go             ← Require[T] — type-asserts to a capability interface
  └── fanout.go              ← FanOut[T] — parallel multi-symbol calls
        ↓
provider/                ← capability interfaces + implementations
  ├── provider.go            ← Provider base interface
  ├── exchange.go            ← ExchangeProvider
  ├── aggregator.go          ← AggregatorProvider
  ├── capability.go          ← PriceProvider, CandleProvider, TickerProvider, OrderBookProvider
  ├── stream.go              ← PriceStreamProvider, OrderBookStreamProvider
  ├── binance/               ← exchange provider (spot + futures)
  ├── bitget/                ← exchange provider (spot + futures)
  ├── coingecko/             ← aggregator + streaming provider
  ├── whitebit/              ← exchange provider (spot, streaming)
  ├── cryptocom/             ← exchange provider (spot, streaming)
  └── registry/
        └── registry.go      ← NewProvider() factory — separate pkg to avoid import cycle
        ↓
model/                   ← provider-agnostic data types
capability/              ← MarketType / Feature matrix types (no project imports)
internal/ws/                 ← WebSocket infrastructure (CLI-only internal)
  ├── base_client.go         ← reusable WebSocket client (reconnect + backoff)
  └── client.go              ← CoinGecko-specific streaming client (ActionCable protocol)
```

Commands **never** import a provider package directly. They always receive a `provider.Provider`
from `newResolver(cfg).Resolve(...)`, then type-assert to a capability interface via
`resolve.Require[T]`.

---

## Interface Hierarchy

All interfaces live in `provider/`.

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
    StartPriceStream(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error)
    SubscribePrice(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error)
    UnsubscribePrice(ctx context.Context, ids []string) error
    SubscribedPrices() []string
    StopPriceStream() error
    PriceStreamStatus() StreamStatus
    GetLastPrice(id string) (*model.CoinPrice, error)
    ReconnectPrice(ctx context.Context) error
    GetDataChannelPrice() <-chan *model.CoinPrice
}

type OrderBookStreamProvider interface {
    StartOrderBookStream(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error)
    SubscribeOrderBook(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error)
    UnsubscribeOrderBook(ctx context.Context, symbols []string) error
    SubscribedOrderBooks() []string
    StopOrderBookStream() error
    OrderBookStreamStatus() StreamStatus
    GetLastOrderBook(symbol string) (*model.OrderBook, error)
    ReconnectOrderBook(ctx context.Context) error
    GetDataChannelOrderBook() <-chan *model.OrderBook
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

A provider lives in `provider/<name>/`. Follow this file-per-concern convention:

```
provider/<name>/
├── client.go          — Client struct, NewClient(), ID(), SetUserAgent(), Capabilities(), doRequest()
├── errors.go          — providerErr, httpErr, apiErr, httpStatusToKind, apiCodeToKind helpers
├── types.go           — provider-specific JSON structs (internal; never exported to model)
├── market.go          — PriceProvider, CandleProvider, TickerProvider, OrderBookProvider methods
├── exchange.go        — (optional) ExchangeProvider methods (ServerTime, ExchangeInfo)
└── stream.go          — (optional) streaming methods via ws.BaseClient
```

Additional files (`auth.go`, `trading.go`, etc.) are fine. Keep `client.go` focused on
construction, identity, and capabilities; keep API logic in `market.go` / `exchange.go`.
Keep error construction helpers in `errors.go` — never use `fmt.Errorf` at provider boundaries.

---

## Step-by-Step: Adding a New Provider

### 1. Config — `config/config.go`

Add a config struct with `mapstructure` tags. Include `Spot`/`Futures`/`Margin` sub-structs
as needed, with `IsXxxEnabled()` helpers:

```go
type MyExchangeConfig struct {
    APIKey    string       `mapstructure:"api_key"`
    APISecret string       `mapstructure:"api_secret"`
    BaseURL   string       `mapstructure:"base_url"`
    Spot      MarketConfig `mapstructure:"spot"`
}

func (c MyExchangeConfig) IsSpotEnabled() bool { return c.Spot.Enabled }
```

Add to the top-level `Config` struct:

```go
MyExchange MyExchangeConfig `mapstructure:"myexchange"`
```

Also update:
- `applyEnvMap()` — handle `"myexchange.*"` keys from `.env` files
- `applyEnvOverrides()` — handle `BITS_MYEXCHANGE_*` env vars:
  ```go
  cfg.MyExchange.APIKey    = getEnv("BITS_MYEXCHANGE_API_KEY",    cfg.MyExchange.APIKey)
  cfg.MyExchange.APISecret = getEnv("BITS_MYEXCHANGE_API_SECRET", cfg.MyExchange.APISecret)
  ```
- `Redacted()` — mask `APIKey` and `APISecret`
- `ConfigTemplate` — add a commented `[myexchange]` section

### 2. `client.go` — struct, constructor, identity, capabilities

```go
package myexchange

import (
    "net/http"
    "time"

    "github.com/mdnmdn/bits/capability"
    "github.com/mdnmdn/bits/config"
)

const providerID = "myexchange"
const defaultBaseURL = "https://api.myexchange.com"

type Client struct {
    cfg        config.MyExchangeConfig
    httpClient *http.Client
    userAgent  string
}

func NewClient(cfg config.MyExchangeConfig) *Client {
    if cfg.BaseURL == "" {
        cfg.BaseURL = defaultBaseURL
    }
    return &Client{
        cfg:        cfg,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
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

Add a `doRequest` helper for all HTTP calls. It must check the HTTP status code and wrap
every error path with a typed helper from `errors.go`:

```go
func (c *Client) doRequest(ctx context.Context, path, query string) ([]byte, error) {
    url := c.cfg.BaseURL + "/" + path
    if query != "" {
        url += "?" + query
    }
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, providerErr(model.ErrKindUnknown, "build request: "+err.Error(), err)
    }
    req.Header.Set("Content-Type", "application/json")
    if c.userAgent != "" {
        req.Header.Set("User-Agent", c.userAgent)
    }
    // add auth headers here if needed (see internal/auth/signature.go for HMAC helpers)
    resp, err := c.httpClient.Do(req)
    if err != nil {
        if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
            return nil, providerErr(model.ErrKindCanceled, err.Error(), err)
        }
        return nil, providerErr(model.ErrKindNetwork, err.Error(), err)
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, providerErr(model.ErrKindNetwork, "read response: "+err.Error(), err)
    }
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, httpErr(resp.StatusCode, string(body))
    }
    return body, nil
}
```

### 3. `errors.go` — error construction helpers

Every provider package must have an `errors.go` with unexported helpers that construct
`*model.ProviderError` values. Never use `fmt.Errorf` at provider boundaries.

```go
package myexchange

import "github.com/mdnmdn/bits/model"

func providerErr(kind model.ErrorKind, msg string, cause error) *model.ProviderError {
    return &model.ProviderError{
        Kind:            kind,
        ProviderID:      providerID,
        ProviderMessage: msg,
        Cause:           cause,
    }
}

func httpErr(status int, body string) *model.ProviderError {
    return &model.ProviderError{
        Kind:            httpStatusToKind(status),
        ProviderID:      providerID,
        ProviderMessage: body,
        HTTPStatus:      status,
    }
}

// apiErr maps the provider's native error code to a ProviderError.
// Adjust the signature to match the provider's envelope: string code, int code, etc.
func apiErr(code, msg string) *model.ProviderError {
    return &model.ProviderError{
        Kind:            apiCodeToKind(code),
        ProviderID:      providerID,
        ProviderCode:    code,
        ProviderMessage: msg,
    }
}

func httpStatusToKind(status int) model.ErrorKind {
    switch {
    case status == 401 || status == 403:
        return model.ErrKindAuth
    case status == 404:
        return model.ErrKindNotFound
    case status == 429:
        return model.ErrKindRateLimit
    case status == 400:
        return model.ErrKindInvalidRequest
    case status >= 500:
        return model.ErrKindServerError
    default:
        return model.ErrKindUnknown
    }
}

func apiCodeToKind(code string) model.ErrorKind {
    switch code {
    case "40001", "40003": // auth errors — adjust per provider
        return model.ErrKindAuth
    default:
        return model.ErrKindUnknown
    }
}
```

Only define `apiErr` / `apiCodeToKind` if the provider returns structured error codes in the
response body (Bitget, Crypto.com, WhiteBit). HTTP-only providers (MEXC, CoinGecko) need only
`providerErr`, `httpErr`, and `httpStatusToKind`.

### 4. `types.go` — internal JSON shapes

Define private structs for deserialising the exchange's API responses. Never expose them
outside the package; convert to `model.*` types in `market.go`.

- Name structs with a provider prefix: `type myexchangeTickerData struct { … }`
- Use `string` fields for price/qty if the API returns quoted numbers; `float64` for bare numbers
- Use exact API field names in `json` tags

```go
package myexchange

type myexchangeTickerData struct {
    Symbol    string `json:"symbol"`
    LastPrice string `json:"lastPrice"`   // quoted float — parse with strconv.ParseFloat
    // …
}
```

### 5. `market.go` — implement capability interfaces

All methods return `model.Response[T]` with `Provider`, `Market`, and `Kind` populated.
Collect per-symbol failures in `[]model.ItemError` rather than returning a top-level error.
All error values must be `*model.ProviderError` — use the helpers from `errors.go`.

```go
func (c *Client) Price(ctx context.Context, ids []string, currency string) (model.Response[[]model.CoinPrice], error) {
    prices := make([]model.CoinPrice, 0, len(ids))
    var errs []model.ItemError

    for _, id := range ids {
        data, err := c.fetchTicker(id) // doRequest + JSON decode inside
        if err != nil {
            // WrapError preserves *ProviderError as-is; wraps plain errors as ErrKindUnknown.
            errs = append(errs, model.ItemError{Symbol: id, Err: model.WrapError(providerID, err)})
            continue
        }
        prices = append(prices, convertPrice(data))
    }

    return model.Response[[]model.CoinPrice]{
        Kind:     model.KindPrice,
        Data:     prices,
        Provider: providerID,
        Market:   model.MarketSpot,
        Errors:   errs,
    }, nil
}

func (c *Client) Ticker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
    body, err := c.doRequest(ctx, "ticker", "symbol="+symbol)
    if err != nil {
        // Total failure — return as the method error, not in Errors slice.
        return model.Response[model.Ticker24h]{}, err
    }
    var envelope struct {
        Code string          `json:"code"`
        Msg  string          `json:"msg"`
        Data myexchangeTicker `json:"data"`
    }
    if err := json.Unmarshal(body, &envelope); err != nil {
        return model.Response[model.Ticker24h]{}, providerErr(model.ErrKindParse, err.Error(), err)
    }
    if envelope.Code != "0" {
        return model.Response[model.Ticker24h]{}, apiErr(envelope.Code, envelope.Msg)
    }
    return model.Response[model.Ticker24h]{
        Kind:     model.KindTicker,
        Data:     convertTicker(envelope.Data, market),
        Provider: providerID,
        Market:   market,
    }, nil
}
```

Rules for all methods:
- Always set `Response.Provider = providerID`
- Always set `Response.Market = market` (or `model.MarketSpot` for spot-only)
- Always set `Response.Kind = model.KindXxx`
- Use `strconv.ParseFloat(s, 64)` for string-encoded numbers
- Use `time.UnixMilli(ms)` for millisecond timestamps
- Map standard intervals (`1m`, `5m`, `1h`, `4h`, `1d`) to the provider's format in `Candles`
- Total call failures → return as `error`; per-symbol failures → append to `Response.Errors`
- Never use `fmt.Errorf` — always use `providerErr`, `httpErr`, `apiErr`, or `model.WrapError`

Implement only the interfaces the exchange actually supports. Commands check at runtime via
`resolve.Require[T]`; unsupported interfaces cause a clean error or fallback.

### 6. `exchange.go` — ExchangeProvider (optional)

Implement if the exchange supports `bits time` and `bits info`:

```go
func (c *Client) ServerTime(ctx context.Context) (model.Response[model.ServerTime], error) {
    // Call dedicated endpoint (e.g. /api/v1/time) or estimate from round-trip.
    // Set Latency field if measuring the round-trip duration.
    return model.Response[model.ServerTime]{
        Kind:     model.KindServerTime,
        Data:     model.ServerTime{Time: serverTime, Latency: &latency},
        Provider: providerID,
        Market:   model.MarketSpot,
    }, nil
}

func (c *Client) ExchangeInfo(ctx context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error) {
    // Route to spot/futures/margin sub-method based on market.
    // Map provider status strings to model.SymbolStatus constants.
    // Set PricePrecision, QtyPrecision, MinQty, MaxQty where available.
}
```

### 7. `stream.go` — streaming (optional)

Use `ws.BaseClient` from `internal/ws/base_client.go`. Implement `OnConnect` (subscriptions)
and `OnMessage` (parsing), then push results onto a channel.

```go
package myexchange

import (
    "context"
    "encoding/json"

    "github.com/mdnmdn/bits/model"
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

### 8. Register the provider — `provider/registry/registry.go`

Add a case to `NewProvider`, and the provider ID to `AllProviderIDs` and `AllProviderIDsWithAliases`.
Also add any short aliases to the `providerAliases` map:

```go
var providerAliases = map[string]string{
    // …
    "me": "myexchange",
}

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

func AllProviderIDsWithAliases() []string {
    return []string{"coingecko", "binance", "bitget", "myexchange", "me"}
}
```

### 9. Documentation

- Create `_docs/providers/myexchange-api.md` — capabilities table, API notes, auth details, config examples, usage examples
- Update `_docs/architecture.md` — add a row to the provider capabilities table

---

## Data Model

All public method signatures use types from `model/`:

| Type | Description |
|------|-------------|
| `Response[T]` | Generic envelope: `Kind`, `Data T`, `Provider`, `Market`, `Fallback`, `Errors` |
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

## Fee Representation

All fees are stored as **decimal fractions** (not basis points or percentages):

| Field | Type | Example | Meaning |
|-------|------|---------|---------|
| `Symbol.MakerFee` | `*float64` | `0.001` | 0.1% maker fee |
| `Symbol.TakerFee` | `*float64` | `0.001` | 0.1% taker fee |

**Conversion rules:**
- If API returns basis points (e.g., `10` for 0.1%), divide by 10000
- If API returns percentage (e.g., `0.1` for 0.1%), divide by 100
- If API returns ratio (e.g., `0.001`), use as-is

**Priority:** Symbol-specific fees > Default trading fees. When an exchange provides per-symbol fees, use those; otherwise apply the account-level default fees to all symbols.

**Sources by provider:**
- **Binance**: Use `/api/v3/exchangeInfo` → `filters.COMMISSION_RATE` or account-level `/api/v3/account` → `commissionRates`
- **Bitget**: Use `/api/v2/spot/market/vip-fee-rate` (spot) or `/api/v2/margin/currencies` (margin)
- **WhiteBit**: Use `/api/v4/fees` endpoint
- **Other providers**: See `_docs/providers/` for specific endpoints

---

## Testing Patterns

Look at `provider/binance/client_test.go` and `provider/coingecko/client_test.go`
for the established pattern:

- Use `net/http/httptest.NewServer` to mock HTTP responses.
- Use `testify/assert` for assertions.
- Test each capability method in isolation.
- For market-type branching, write one test per market type.

For WebSocket methods, test `OnMessage` in isolation by calling it directly with fixture JSON,
without spinning up a real WebSocket server.

---

## Verification

```sh
# Compile the whole project
go build ./...

# Run all tests (race detector on)
go test -race ./...

# Check formatting
gofmt -l ./provider/myexchange/

# Sanity-check the provider appears
go run . providers
go run . capabilities -p myexchange

# Live smoke tests (requires network + valid symbols)
go run . time -p myexchange
go run . info -p myexchange
go run . price BTC_USDT -p myexchange
go run . ticker BTC_USDT -p myexchange
go run . book BTC_USDT -p myexchange
```

---

## Common Pitfalls

| Issue | Fix |
|-------|-----|
| `resolve.Require` fails at runtime | Verify the provider struct implements the interface method signatures exactly |
| `bits capabilities` shows no entries | Check `Capabilities()` returns non-empty matrix |
| Prices always 0 | Exchange returns string prices — use `strconv.ParseFloat` |
| Wrong percent change | Multiply ratio by 100 if exchange returns e.g. `0.02` for 2% |
| ExchangeInfo returns all symbols | Filter by product type / status in the loop |
| `bits time -p myexchange` falls back | Implement `ExchangeProvider` and register `FeatureServerTime` |
| Import cycle | Never import registry from a provider package; only registry imports providers |
| Linter: `func httpErr is unused` | Remove helpers not called anywhere — only define what is actually used |
| HTTP 500 silently parsed as success | Check `resp.StatusCode` in `doRequest` and return `httpErr(status, body)` |
| `ItemError.Err` compile error | `Err` field is `*model.ProviderError`, not `error` — use `model.WrapError` to adapt |

---

## File Checklist

```
□ config/config.go
  □ MyExchangeConfig struct (APIKey, APISecret, BaseURL, Spot/Futures MarketConfig)
  □ IsSpotEnabled() / IsFuturesEnabled() helpers
  □ Add MyExchange to top-level Config struct
  □ applyEnvMap() — ".env" file keys
  □ applyEnvOverrides() — BITS_MYEXCHANGE_* env vars
  □ Redacted() — mask APIKey / APISecret
  □ ConfigTemplate — add commented [myexchange] section

□ provider/myexchange/
  □ client.go     — Client struct, NewClient, defaultBaseURL, ID, SetUserAgent, Capabilities, doRequest (with HTTP status check)
  □ errors.go     — providerErr, httpErr, httpStatusToKind; add apiErr/apiCodeToKind if provider has structured error codes
  □ types.go      — private JSON response structs (prefixed names, string fields for quoted numbers)
  □ market.go     — Price / Ticker24h / Candles / OrderBook (set Kind, Provider, Market on every Response; use error helpers, not fmt.Errorf)
  □ exchange.go   — (optional) ServerTime (with Latency) + ExchangeInfo (with SymbolStatus mapping)
  □ stream.go     — (optional) WatchPrices / WatchOrderBook via ws.NewBaseClient
  □ client_test.go — httptest mocks for each capability method

□ provider/registry/registry.go
  □ providerAliases map entry
  □ NewProvider case
  □ AllProviderIDs / AllProviderIDsWithAliases

□ _docs/providers/myexchange-api.md  — capabilities, API notes, config examples
□ _docs/architecture.md              — add row to provider capabilities table
```
