# WhiteBit Provider — Work In Progress

## Overview

**Status: ✅ ALL PHASES COMPLETE**

Add WhiteBit as a new data provider to `bits`, implementing the following capabilities (spot only):

| Capability         | Interface              | Endpoint |
|--------------------|------------------------|----------|
| Server time        | `ExchangeProvider`     | `GET /api/v4/public/time` |
| Exchange info      | `ExchangeProvider`     | `GET /api/v4/public/markets` |
| Price              | `PriceProvider`        | `GET /api/v4/public/ticker` |
| Candles (OHLCV)    | `CandleProvider`       | `GET /api/v4/public/kline` |
| 24h ticker         | `TickerProvider`       | `GET /api/v4/public/ticker` |
| Order book         | `OrderBookProvider`    | `GET /api/v4/public/orderbook/{market}` |

WhiteBit uses underscore-separated market symbols (e.g. `BTC_USDT`).

## rules
- at the end of each phase the project should compile and be formatted and vetted and make sure it runs
- at the end of each phase update this document marking the phase as completed and describing the updates (briefly but completely, no code only file and method references, no lines) in order to the next processing (with maybe no previous context) could start lean.
- when possible delegate tasks to subagents with simpler model (haiku based)
- remember: simplicity is the ultimate perfection

---

## Phase 1 — Config & Registration ✅ COMPLETED

**Goal:** Wire WhiteBit into the config and registry so it can be selected as a provider.

**Completed:** Added `WhiteBitConfig` struct with `APIKey`, `APISecret`, `BaseURL`, and `Spot MarketConfig` fields to `internal/config/config.go`, plus `IsSpotEnabled()` method. Added `WhiteBit WhiteBitConfig` field to root `Config` struct. Wired env var support in `applyEnvMap()` and `applyEnvOverrides()` for `.env` and `BITS_*` environment variables. Added WhiteBit redaction to `Redacted()` method and configuration template to `ConfigTemplate`. Registered `"whitebit"` case in `NewProvider()` and `AllProviderIDs()` in `internal/registry/registry.go`. Created stub `internal/provider/whitebit/client.go` with `Client` type, `NewClient()`, `ID()`, `SetUserAgent()`, and `Capabilities()` methods.

**Files to modify:**
- `internal/config/config.go` — `WhiteBitConfig` struct, `Config` field, `applyEnvMap`, `applyEnvOverrides`, `Redacted`, `ConfigTemplate`
- `internal/registry/registry.go` — `NewProvider` switch, `AllProviderIDs`

**Reference implementations:**
- `BitgetConfig` struct and its env wiring in `internal/config/config.go:87-94` (no passphrase variant)
- `case "bitget"` in `internal/registry/registry.go:25`

### Tasks

- [ ] Add `WhiteBitConfig` struct to `internal/config/config.go`
  - Fields: `APIKey`, `APISecret`, `BaseURL`, `Spot MarketConfig`
  - No passphrase needed (unlike Bitget)
- [ ] Add `WhiteBit WhiteBitConfig` field to root `Config` struct (`internal/config/config.go:106`)
- [ ] Add WhiteBit entries to `applyEnvMap` (`internal/config/config.go:362`) — `.env` file support
  - `whitebit.api_key`, `whitebit.api_secret`, `whitebit.base_url`, `whitebit.spot.enabled`
- [ ] Add WhiteBit entries to `applyEnvOverrides` (`internal/config/config.go:420`) — env var support
  - `BITS_WHITEBIT_API_KEY`, `BITS_WHITEBIT_API_SECRET`, `BITS_WHITEBIT_BASE_URL`, `BITS_WHITEBIT_SPOT_ENABLED`
- [ ] Add WhiteBit to `Redacted()` method (`internal/config/config.go:535`)
- [ ] Add WhiteBit section to `ConfigTemplate` string (`internal/config/config.go:222`)
- [ ] Register in `internal/registry/registry.go`
  - Add `case "whitebit": return whitebit.NewClient(cfg.WhiteBit), nil` in `NewProvider` (`registry.go:18`)
  - Add `"whitebit"` to `AllProviderIDs()` (`registry.go:32`)

---

## Phase 2 — Provider Client ✅ COMPLETED

**Goal:** Create the WhiteBit provider package with HTTP client scaffolding.

**Completed:** Rewrote `internal/provider/whitebit/client.go` with full `Capabilities()` registering all six spot features (`FeatureServerTime`, `FeatureExchangeInfo`, `FeaturePrice`, `FeatureCandles`, `FeatureTicker24h`, `FeatureOrderBook`) and `doRequest()` performing plain GET requests with optional `User-Agent` header.

**File to create:** `internal/provider/whitebit/client.go`

**Reference implementations:**
- `internal/provider/bitget/client.go` — raw HTTP client pattern (no SDK), `doRequest`, `Capabilities` matrix building
- `internal/capability/capability.go` — `CapabilityMatrix`, feature/market constants (`FeatureServerTime`, `FeatureExchangeInfo`, etc.)
- `internal/provider/provider.go` — `Provider` base interface (`ID`, `SetUserAgent`, `Capabilities`)

### Tasks

- [x] Create `internal/provider/whitebit/` directory
- [x] Implement `Client` struct
  - Fields: `cfg config.WhiteBitConfig`, `httpClient *http.Client`, `userAgent string`
- [x] Implement `NewClient(cfg config.WhiteBitConfig) *Client`
  - Default `BaseURL` to `https://whitebit.com`
  - 30s HTTP timeout
- [x] Implement `ID() string` → `"whitebit"`
- [x] Implement `SetUserAgent(string)`
- [x] Implement `Capabilities() capability.CapabilityMatrix`
  - Spot: `FeatureServerTime`, `FeatureExchangeInfo`, `FeaturePrice`, `FeatureCandles`, `FeatureTicker24h`, `FeatureOrderBook`
- [x] Implement private `doRequest(path string) ([]byte, error)`
  - Plain GET, no auth needed for public API v4
  - Sets `Content-Type: application/json` and `User-Agent` if set

---

## Phase 3 — Exchange Provider (Time + Info) ✅ COMPLETED

**Goal:** Implement `ExchangeProvider` interface.

**Completed:** Created `internal/provider/whitebit/exchange.go` implementing `ExchangeProvider`. `ServerTime()` calls `GET /api/v4/public/time` and parses Unix seconds via `time.Unix`. `ExchangeInfo()` calls `GET /api/v4/public/markets`, maps each market to `model.Symbol` with `tradesEnabled` → `SymbolStatusTrading`/`SymbolStatusBreak`, always returns `model.MarketSpot`.

**File created:** `internal/provider/whitebit/exchange.go`

**Reference implementations:**
- `internal/provider/bitget/exchange.go` — `ServerTime` and `ExchangeInfo` pattern, response struct definitions, status conversion helpers
- `internal/provider/exchange.go` — `ExchangeProvider` interface signature
- `internal/model/exchange.go` — `ServerTime`, `ExchangeInfo`, `Symbol`, `SymbolStatus` types
- `internal/model/response.go` — `Response[T]` envelope

---

## Phase 4 — Market Data (Price, Ticker, Candles, OrderBook) ✅ COMPLETED

**Goal:** Implement `PriceProvider`, `TickerProvider`, `CandleProvider`, `OrderBookProvider`.

**Completed:** Created `internal/provider/whitebit/market.go`. `fetchAllTickers()` fetches `GET /api/v4/public/ticker` (returns map of all markets). `Price()` and `Ticker24h()` both use `fetchAllTickers()`. `Candles()` calls `GET /api/v4/public/kline` — note non-standard column order `[ts, open, close, high, low, vol]`. `OrderBook()` calls `GET /api/v4/public/orderbook/{symbol}`. All methods return `model.Response[T]` with `Provider: providerID` and `Market: model.MarketSpot`.

**File created:** `internal/provider/whitebit/market.go`

**Reference implementations:**
- `internal/provider/bitget/market.go` — `Price`, `Ticker24h`, `Candles`, `fetchTicker` patterns; `ItemError` collection; `convertGranularity*` helpers
- `internal/provider/capability.go` — `PriceProvider`, `TickerProvider`, `CandleProvider`, `OrderBookProvider` interface signatures
- `internal/model/price.go` — `CoinPrice` struct
- `internal/model/ticker.go` — `Ticker24h` struct (pointer fields for optional values)
- `internal/model/candle.go` — `Candle`, `CandleOpts` structs
- `internal/model/orderbook.go` — `OrderBook`, `OrderBookEntry` structs
- `internal/model/response.go` — `Response[T]`, `ItemError`

### Tasks

#### Ticker / Price
- [ ] Define `whitebitTicker` struct: `Open`, `High`, `Low`, `Last`, `Volume`, `Deal`, `Change`
- [ ] Define `whitebitTickerMap` as `map[string]whitebitTicker`
- [ ] Implement `Price(ctx, ids, currency)` using `GET /api/v4/public/ticker`
  - Fetch all tickers, filter by requested `ids`
  - `Change` field is percent (e.g. `"2.00"` = 2%)
  - Collect misses as `model.ItemError`
- [ ] Implement `Ticker24h(ctx, symbol, market)` using `GET /api/v4/public/ticker`
  - Fetch all tickers, find `symbol`
  - Populate: `LastPrice`, `HighPrice`, `LowPrice`, `OpenPrice`, `Volume`, `QuoteVolume`, `PriceChange`, `PriceChangePercent`
  - `PriceChange = Last - Open`

#### Candles
- [ ] Implement `Candles(ctx, symbol, market, interval, opts)`
  - Endpoint: `GET /api/v4/public/kline?market={symbol}&interval={interval}`
  - Optional params: `limit`, `start` (Unix seconds), `end` (Unix seconds)
  - Response: `{"success": true, "result": [[ts, open, close, high, low, vol, amount], ...]}`
  - **Note:** WhiteBit kline column order is `[time, open, close, high, low, volume]` — close is index 2, high is index 3, low is index 4
  - Timestamps are Unix seconds (not milliseconds) — use `time.Unix`, not `time.UnixMilli`
- [ ] Implement `convertInterval(interval string) string`
  - Map `1m`, `5m`, `15m`, `30m`, `1h`, `4h`, `1d`, `1w` → WhiteBit names (same format as input)

#### Order Book
- [ ] Implement `OrderBook(ctx, symbol, market, depth)`
  - Endpoint: `GET /api/v4/public/orderbook/{symbol}?limit={depth}`
  - Response fields: `timestamp`, `ask [][]float64`, `bid [][]float64`
  - Entry format: `[price, amount]`
  - `timestamp` is in milliseconds — use `time.UnixMilli`
- [ ] Define `whitebitOrderBookResponse` struct: `Timestamp int64`, `Asks [][]float64`, `Bids [][]float64`

---

## Phase 5 — Verification ✅ COMPLETED

**Goal:** Verify WhiteBit provider compilation, formatting, linting, testing, and functional operation.

**Completed:**
- Build check: `go build ./...` — no errors
- Format check: `gofmt -w .` — fixed 14 files (existing codebase formatting issues, not WhiteBit-specific)
- Vet check: `go vet ./...` — no errors
- Test check: `go test -race ./...` — all tests pass
- Binary build: `go build -o bits .` — successful
- Capability matrix: `./bits capabilities -p whitebit` — all six spot features registered (server_time, exchange_info, price, candles, ticker_24h, order_book) with ✓ marks
- Smoke tests:
  - `./bits time -p whitebit` — returns server time with latency and clock skew
  - `./bits info -p whitebit` — returns 1061 symbols from exchange info endpoint
  - `./bits capabilities -p whitebit` — capability matrix displays correctly

All checks passed without errors or panics.

---

## API Reference

- **Base URL:** `https://whitebit.com`
- **Docs:** https://docs.whitebit.com (v4 public API)
- **Rate limits:** 100 req/s public, no auth required for public endpoints
- **Symbol format:** `BASE_QUOTE` (e.g. `BTC_USDT`, `ETH_BTC`)

### Key Endpoint Formats

| Endpoint | Params | Notes |
|----------|--------|-------|
| `GET /api/v4/public/time` | — | Returns Unix seconds |
| `GET /api/v4/public/markets` | — | Returns array of market objects |
| `GET /api/v4/public/ticker` | — | Returns map of all tickers |
| `GET /api/v4/public/kline` | `market`, `interval`, `limit`, `start`, `end` | Kline timestamps are Unix seconds |
| `GET /api/v4/public/orderbook/{market}` | `limit` | Timestamp in milliseconds |

### Kline Column Order

```
[timestamp, open, close, high, low, volume, amount]
  idx 0       1     2      3     4    5       6
```

⚠️ Non-standard order: `close` (idx 2) comes before `high` (idx 3) and `low` (idx 4).
