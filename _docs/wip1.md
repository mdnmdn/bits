# Plan: Add Binance & Bitget Providers + Architecture Documentation

## Context

The `bits` CLI completed Phase 1 (extracting CoinGecko into a provider pattern) but the `Provider` interface still returns CoinGecko-specific types. The goal is to:
1. Create generic model types and refactor the interface
2. Port Binance and Bitget reference implementations from `___tmp/providers/`
3. Expose prices, candles/OHLC, tickers, and order books across all providers
4. Port trading code (orders, balances) but leave it internal for future use
5. Support multi-provider config (YAML + env vars)
6. Update architecture docs to reflect current state

---

## Step 1: Generic Model Types

**Create `internal/model/types.go`**

Define provider-agnostic types that mirror the existing CoinGecko shapes (to minimize refactoring):

| Type | Shape | Notes |
|------|-------|-------|
| `PriceResponse` | `map[string]map[string]float64` | Same as coingecko.PriceResponse |
| `MarketCoin` | struct with same fields | Same JSON-friendly fields |
| `SearchResponse`, `SearchCoin` | structs | CoinGecko-only but generic typed |
| `TrendingResponse` + nested | structs | CoinGecko-only but generic typed |
| `HistoricalData` | struct | CoinGecko-only |
| `MarketChartResponse` | struct (Prices, MarketCaps, TotalVolumes) | Shared |
| `OHLCData` | `[][]float64` | Shared across providers |
| `Ticker24h` | NEW struct (Symbol, LastPrice, PriceChange%, High, Low, Volume...) | For Binance/Bitget |
| `OrderBook`, `OrderBookEntry` | NEW structs (Price/Quantity as float64) | For Binance/Bitget |
| `GainersLosersResponse`, `GainerCoin` | structs | CoinGecko-only |
| `CoinDetail`, `CoinDetailMarket` | structs | CoinGecko-only (TUI) |

**Create `internal/model/errors.go`**

Move sentinel errors to model package: `ErrInvalidAPIKey`, `ErrPlanRestricted`, `ErrRateLimited`, `RateLimitError`. Add `ErrNotSupported`.

---

## Step 2: Refactor Provider Interface

**Modify `internal/provider/types.go`** - Split into capability interfaces:

```
Provider (base)         - ID(), SetUserAgent(), SimplePrice(), CoinOHLC()
SymbolPricer            - SimplePriceBySymbols()          [CoinGecko]
MarketLister            - CoinMarkets(), FetchAllMarkets() [CoinGecko]
Searcher                - Search()                         [CoinGecko]
TrendingProvider        - SearchTrending()                 [CoinGecko]
HistoricalProvider      - CoinHistory(), CoinMarketChart(), CoinMarketChartRange(), CoinOHLCRange()
GainersLosersProvider   - TopGainersLosers()               [CoinGecko]
DetailProvider          - CoinDetail()                     [CoinGecko, TUI]
TickerProvider          - Ticker24h()                      [Binance, Bitget] NEW
OrderBookProvider       - OrderBook()                      [Binance, Bitget] NEW
```

All return types use `model.*` instead of `coingecko.*`.

---

## Step 3: CoinGecko Adapter

- **Keep `coingecko/types.go`** as-is for JSON unmarshaling
- **Create `coingecko/convert.go`** with `toModelMarketCoin()`, `toModelSearchResponse()`, etc.
- **Update `coingecko/coins.go`** return types to `model.*`, calling converters
- **Update `coingecko/client.go`** errors to re-export from `model`
- CoinGecko implements ALL capability interfaces

---

## Step 4: Update Consumers

**`cmd/` files** - Change imports from `coingecko.*` to `model.*`:
- `cmd/price.go` - `coingecko.PriceResponse` -> `model.PriceResponse`
- `cmd/output.go` - error types from `model`
- `cmd/history.go` - response types + error types from `model`
- `cmd/markets_test.go`, `cmd/price_test.go`, etc. - update type references

**`internal/tui/` files** - Change imports:
- `tui/markets.go` - `coingecko.MarketCoin` -> `model.MarketCoin`
- `tui/detail.go` - `coingecko.CoinDetail` -> `model.CoinDetail`
- `tui/trending.go` - `coingecko.TrendingResponse` -> `model.TrendingResponse`

**Checkpoint: `go test ./...` must pass before continuing.**

---

## Step 5: Multi-Provider Config

**Modify `internal/config/config.go`**:

```go
type Config struct {
    Provider string        `mapstructure:"provider"` // active provider
    APIKey   string        `mapstructure:"api_key"`  // CoinGecko (backward compat)
    Tier     string        `mapstructure:"tier"`      // CoinGecko
    Binance  BinanceConfig `mapstructure:"binance"`
    Bitget   BitgetConfig  `mapstructure:"bitget"`
}

type BinanceConfig struct {
    APIKey, APISecret, BaseURL string
    UseTestnet                 bool
}

type BitgetConfig struct {
    Key, Secret, Passphrase, BaseURL string
}
```

Env var overrides with `BITS_` prefix:
- `BITS_PROVIDER`, `BITS_BINANCE_API_KEY`, `BITS_BINANCE_API_SECRET`, `BITS_BITGET_KEY`, `BITS_BITGET_SECRET`, `BITS_BITGET_PASSPHRASE`

---

## Step 6: CLI Integration

**`cmd/root.go`** - Add `--provider`/`-p` persistent flag

**`cmd/client_factory.go`** - Resolve provider: flag -> env -> config -> "coingecko" default

**`internal/provider/registry.go`** - Add `"binance"` and `"bitget"` cases

**Add capability checks** to commands that use provider-specific features:
```go
tp, ok := client.(provider.TrendingProvider)
if !ok {
    return fmt.Errorf("%s provider does not support trending data", client.ID())
}
```

Commands needing guards: `trending`, `search`, `markets`, `top-gainers-losers`, `watch` (CoinGecko-only WebSocket)

---

## Step 7: Binance Provider

**Create `internal/provider/binance/`** (adapted from `___tmp/providers/binance/`):

| File | Contents |
|------|----------|
| `client.go` | Client struct using `go-binance/v2`, constructor, ID(), SetUserAgent() |
| `types.go` | Internal types: AccountInfo, Balance, TickerStats, OrderBook, Symbol, Filter, order types |
| `market.go` | Provider interface methods: SimplePrice, CoinOHLC, Ticker24h, OrderBook |
| `convert.go` | Kline->OHLCData, stats->Ticker24h, depth->OrderBook conversions |
| `trading.go` | Internal trading methods (unexported or future): PlaceOrder, CancelOrder, GetAccountInfo, etc. |

**New dependency**: `github.com/adshao/go-binance/v2`

**Symbol handling**: Binance uses `BTCUSDT` format. `SimplePrice` accepts these as `ids`.

**Interval mapping**: `"daily"->"1d"`, `"hourly"->"1h"`, pass-through for `"1m"`, `"5m"`, etc.

---

## Step 8: Bitget Provider

**Create `internal/provider/bitget/`** (adapted from `___tmp/providers/bitget/`):

| File | Contents |
|------|----------|
| `client.go` | HTTP client with HMAC-SHA256 signing, trading pairs cache, constructor |
| `types.go` | API response types: TickerResponse, CandleData, SymbolResponse, order types |
| `market.go` | Provider interface methods: SimplePrice, CoinOHLC, Ticker24h |
| `convert.go` | Candle->OHLCData, ticker->Ticker24h conversions |
| `auth.go` | generateSignature(), auth header helpers |
| `trading.go` | Internal trading methods: PlaceOrder, CancelOrder, GetOrderStatus, etc. |

**No new external dependency** (raw HTTP + crypto/hmac)

**Symbol handling**: Bitget uses `BTCUSDT` (with `_SPBL` suffix fallback internally)

**Granularity mapping**: `"daily"->"1day"`, `"hourly"->"1h"`, `"1m"->"1min"`, etc.

---

## Step 9: New CLI Commands

**Add `cmd/ticker.go`** - `bits ticker BTCUSDT -p binance`
- Uses `TickerProvider` capability interface
- Shows 24h stats (price, change, high, low, volume)
- Works with Binance and Bitget

**Add `cmd/orderbook.go`** - `bits orderbook BTCUSDT -p binance`
- Uses `OrderBookProvider` capability interface
- Shows top N bids/asks
- Works with Binance (Bitget order book can be added later)

---

## Step 10: Documentation

**Replace `_docs/wit-transformations.md`** with `_docs/architecture.md`:
- Current architecture (not "in progress")
- Provider interface and capability system
- How to add a new provider
- Config structure

**Update `_docs/features.md`** with multi-provider commands

**Create `_docs/future-tasks.md`**:
- Trading commands (balance, buy, sell, cancel-order)
- Portfolio tracking
- WebSocket streaming for Binance/Bitget
- Provider-specific search/trending for exchanges
- Output formatters (CSV, MD, YAML, TOON)

**Update `CLAUDE.md`** with new project structure and provider list

---

## Implementation Order

1. `internal/model/` (types + errors) - pure types, no deps
2. `internal/provider/types.go` - new interface using model types
3. `internal/provider/coingecko/convert.go` + update `coins.go` + `client.go`
4. Update `cmd/` and `internal/tui/` imports -> **run tests**
5. `internal/config/config.go` - multi-provider config
6. `cmd/root.go` + `cmd/client_factory.go` - `--provider` flag
7. `internal/provider/registry.go` - add cases
8. `internal/provider/binance/` - full implementation
9. `internal/provider/bitget/` - full implementation
10. Capability guards in existing commands
11. New commands: `ticker`, `orderbook`
12. Tests for new providers
13. Documentation updates

---

## Verification

1. `go build -o bits .` compiles
2. `go test -race ./...` passes (existing + new tests)
3. `./bits price --ids bitcoin` still works (CoinGecko default)
4. `./bits price --ids BTCUSDT -p binance` returns Binance price
5. `./bits ticker BTCUSDT -p binance` shows 24h stats
6. `./bits price --ids BTCUSDT -p bitget` returns Bitget price
7. `./bits trending` with `-p binance` returns clear "not supported" error
8. `make lint` passes
