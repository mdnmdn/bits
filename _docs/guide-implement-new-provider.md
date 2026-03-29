# Guide: Implementing a New Provider

This checklist covers every step required to add a new exchange or aggregator
provider to `bits`. Work through the steps in order; run `go build ./...` after
each file to catch issues early.

---

## Reference Documents

| Document | Purpose |
|----------|---------|
| `_docs/architecture.md` | High-level architecture, interface hierarchy, fallback policy |
| `_docs/provider-structure.md` | Step-by-step provider guide, file layout, interface contracts |
| `_docs/data-model.md` | All model types (`Response[T]`, `Ticker24h`, `OrderBook`, …) |
| `_docs/config.md` | Config file format, env var conventions, viper setup |
| `_docs/ws-handling.md` | WebSocket infrastructure (`ws.Manager`, `ws.BaseClient`) |
| `_docs/providers/` | Per-provider API docs (endpoints, auth, response formats) |

Reference implementations:

| Provider | Location | Notable features |
|----------|----------|-----------------|
| Bitget | `pkg/provider/bitget/` | Raw HTTP, HMAC auth, spot + futures + margin |
| WhiteBit | `pkg/provider/whitebit/` | Raw HTTP, no auth for public, streaming |
| Binance | `pkg/provider/binance/` | Go SDK wrapper, spot + futures, streaming |
| CoinGecko | `pkg/provider/coingecko/` | Aggregator, paid/demo tiers, streaming |
| Crypto.com | `pkg/provider/cryptocom/` | Raw HTTP, spot only, no candles yet |

---

## Checklist

### 1. Config — `pkg/config/config.go`

```
□ Add MyExchangeConfig struct with mapstructure tags
  □ APIKey, APISecret, BaseURL (string fields)
  □ Spot / Margin / Futures (MarketConfig sub-structs as needed)
  □ Add IsSpotEnabled() / IsMarginEnabled() / IsFuturesEnabled() helpers

□ Add MyExchange MyExchangeConfig to the top-level Config struct
  (mapstructure:"myexchange")

□ applyEnvMap() — handle "myexchange.*" keys from .env files

□ applyEnvOverrides() — handle BITS_MYEXCHANGE_* env vars

□ Redacted() — mask APIKey and APISecret

□ ConfigTemplate — add commented [myexchange] section
```

Key patterns to follow:
```go
type MyExchangeConfig struct {
    APIKey    string       `mapstructure:"api_key"`
    APISecret string       `mapstructure:"api_secret"`
    BaseURL   string       `mapstructure:"base_url"`
    Spot      MarketConfig `mapstructure:"spot"`
}
func (c MyExchangeConfig) IsSpotEnabled() bool { return c.Spot.Enabled }
```

---

### 2. Provider Directory — `pkg/provider/myexchange/`

Create the directory and the following files:

#### 2a. `client.go` — identity, HTTP client, capabilities

```
□ Package declaration: package myexchange
□ const providerID = "myexchange"
□ const defaultBaseURL = "https://api.myexchange.com"
□ Client struct: cfg, httpClient, userAgent
□ NewClient(cfg config.MyExchangeConfig) *Client
  □ Set cfg.BaseURL = defaultBaseURL if empty
  □ &http.Client{Timeout: 30 * time.Second}
□ ID() string — returns providerID
□ SetUserAgent(ua string) — stores userAgent
□ Capabilities() capability.CapabilityMatrix
  □ Use capability.NewCapabilityMatrix(keys...) for static sets
  □ Or build manually for config-driven sets (enabled markets)
□ doRequest(path, query string) ([]byte, error)
  □ Construct full URL: cfg.BaseURL + "/" + path (+ "?" + query if non-empty)
  □ Set Content-Type: application/json
  □ Set User-Agent if set
  □ For authenticated endpoints: add auth headers (HMAC-SHA256/512, timestamps)
```

See `internal/auth/signature.go` for HMAC helpers.

#### 2b. `types.go` — internal JSON shapes

```
□ Private structs for every API response (never exported)
□ Name with provider prefix: type myexchangeTickerData struct { ... }
□ Cover: ticker, order book, candles, exchange info, server time
□ Use string fields for price/qty if the API returns quoted numbers
□ Use float64 if the API returns bare numbers
□ Add json tags matching the exact API field names
```

#### 2c. `market.go` — capability methods

Implement whichever interfaces the exchange supports:

```
□ provider.PriceProvider
  func (c *Client) Price(ctx, ids []string, currency string) (Response[[]CoinPrice], error)
  □ Return model.KindPrice in Kind field
  □ Collect per-symbol errors in []model.ItemError

□ provider.TickerProvider
  func (c *Client) Ticker24h(ctx, symbol string, market MarketType) (Response[Ticker24h], error)
  □ Return model.KindTicker

□ provider.CandleProvider
  func (c *Client) Candles(ctx, symbol string, market MarketType, interval string, opts CandleOpts) (Response[[]Candle], error)
  □ Map standard intervals (1m, 5m, 1h, 4h, 1d) to provider format
  □ Handle opts.From, opts.To, opts.Limit if supported

□ provider.OrderBookProvider
  func (c *Client) OrderBook(ctx, symbol string, market MarketType, depth int) (Response[OrderBook], error)
  □ Return model.KindOrderBook

Rules for all methods:
□ Always set Response.Provider = providerID
□ Always set Response.Market = market (or model.MarketSpot for spot-only)
□ Always set Response.Kind = model.KindXxx
□ Use strconv.ParseFloat(s, 64) for string-encoded numbers
□ Use time.UnixMilli(ms) for millisecond timestamps
```

#### 2d. `exchange.go` — ExchangeProvider (optional)

Implement if the exchange supports `bits time` and `bits info`:

```
□ provider.ExchangeProvider
  func (c *Client) ServerTime(ctx) (Response[ServerTime], error)
  □ Call dedicated endpoint (e.g. /api/v1/time) OR estimate from round-trip
  □ Set Latency field if measuring round-trip

  func (c *Client) ExchangeInfo(ctx, market MarketType) (Response[ExchangeInfo], error)
  □ Route to spot/futures/margin sub-method based on market
  □ Map provider status strings to model.SymbolStatus constants
  □ Set PricePrecision, QtyPrecision, MinQty, MaxQty where available
```

#### 2e. `stream.go` — streaming (optional)

Implement if the exchange supports WebSocket data:

```
□ Use ws.BaseClient from internal/ws/base_client.go
□ Implement OnConnect (subscriptions) and OnMessage (parsing)
□ WatchPrices() — implements provider.PriceStreamProvider
□ WatchOrderBook() — implements provider.OrderBookStreamProvider
□ Always use ws.BaseClient (NOT ws.Client which is CoinGecko-specific)
```

---

### 3. Registry — `pkg/provider/registry/registry.go`

```
□ Add import: "github.com/mdnmdn/bits/pkg/provider/myexchange"
□ Add aliases to providerAliases map (e.g., "me": "myexchange")
□ Add case to NewProvider switch: case "myexchange": return myexchange.NewClient(cfg.MyExchange), nil
□ Add "myexchange" to AllProviderIDs() slice
□ Add "myexchange", "me" to AllProviderIDsWithAliases() slice
```

---

### 4. Documentation

```
□ Create _docs/myexchange-provider-wip.md
  □ Summary, capabilities table, API notes, config examples, usage examples
□ Update _docs/architecture.md — add row to provider capabilities table
```

---

### 5. Verification

```sh
# Compile the whole project
go build ./...

# Run all tests (race detector on)
go test -race ./...

# Check formatting
gofmt -l ./pkg/provider/myexchange/

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

---

## File Summary

```
pkg/config/config.go                   ← MyExchangeConfig + Config wiring
pkg/provider/myexchange/
  client.go                            ← Client, NewClient, ID, Capabilities, doRequest
  types.go                             ← Private JSON structs
  market.go                            ← Price / Ticker / Candles / OrderBook
  exchange.go                          ← ServerTime + ExchangeInfo (optional)
  stream.go                            ← WatchPrices / WatchOrderBook (optional)
  client_test.go                       ← Unit tests with httptest mocks
pkg/provider/registry/registry.go      ← NewProvider case + AllProviderIDs
_docs/myexchange-provider-wip.md       ← Progress tracker
```
