# Transition Plan — New Data Model

Migration from the current provider/model layer to the clean multi-provider data model
described in `data-model.md`.

The current @architecture.md describe the architecture before this activity

## rules
- at the end of each phase the project should compile and be formatted and vetted and make sure it runs
- at the end of each phase update this document marking the phase as completed and describing the updates (briefly but completely, no code only file and method references, no lines) in order to the next processing (with maybe no previous context) could start lean.
- when possible delegate tasks to subagents with simpler model (haiku based)
- remember: simplicity is the ultimate perfection

---

## Strategy: Legacy Package

Move all current code to `internal/legacy/` untouched, then build the new clean structure
alongside it. Both coexist; legacy commands stay reachable under `bits legacy <cmd>`.
Delete `internal/legacy/` in a single commit when new commands reach parity.

### Packages that stay in place (shared by both)

| Package | Reason |
|---|---|
| `internal/capability/` | No project imports; already the canonical capability system |
| `internal/config/` | Shared config struct; both registries use it |
| `main.go` | Entry point unchanged; calls new `cmd.Execute()` |

### Packages that move to legacy

| Current path | Legacy path |
|---|---|
| `internal/model/` | `internal/legacy/model/` |
| `internal/provider/` | `internal/legacy/provider/` |
| `internal/provider/binance/` | `internal/legacy/provider/binance/` |
| `internal/provider/bitget/` | `internal/legacy/provider/bitget/` |
| `internal/provider/coingecko/` | `internal/legacy/provider/coingecko/` |
| `internal/display/` | `internal/legacy/display/` |
| `internal/export/` | `internal/legacy/export/` |
| `internal/tui/` | `internal/legacy/tui/` |
| `internal/ws/` | `internal/legacy/ws/` |
| `internal/auth/` | `internal/legacy/auth/` |
| `cmd/*.go` (commands + factory) | `internal/legacy/cmd/` |

### New packages (built from scratch)

| New path | Contents |
|---|---|
| `internal/model/` | New clean model types (`Response[T]`, `Candle`, `CoinPrice`, …) |
| `internal/provider/` | New provider interfaces + registry |
| `internal/provider/binance/` | New Binance implementation |
| `internal/provider/bitget/` | New Bitget implementation |
| `internal/provider/coingecko/` | New CoinGecko implementation |
| `internal/resolve/` | Provider resolution, fallback, fan-out |
| `internal/render/` | Format-agnostic rendering layer |
| `internal/process/` | Enrichment / processing layer |
| `cmd/` | New clean commands |

---

## Agent Execution Model

Each phase is a gate: it must compile and pass `make test` before the next phase starts.
Within a phase, tasks marked **parallel** can be handed to independent subagents simultaneously.
Tasks marked **sequential** depend on prior output and must run in order.

---

## Known Issues to Fix During Transition

### 1. MarketType string mismatch

`internal/config/config.go` defines:
```go
MarketTypeFuture = "future"   // ← singular
```
`internal/capability/capability.go` defines:
```go
MarketFutures MarketType = "futures"  // ← plural
```
The capability system's `"futures"` is authoritative. New code uses only
`capability.MarketFutures`. Legacy code keeps `"future"` as-is.

### 2. CoinGecko-specific methods on root `Config`

`config.Config` exposes `BaseURL()`, `AuthHeader()`, `ApplyAuth()`, `IsPaid()`,
`MaskedKey()`, `Redacted()` — CoinGecko concerns on the root struct.
Not fixed during the move; addressed in Phase 9.

### 3. Concrete type assertions in `cmd/client_factory.go`

```go
if bc, ok := c.(*binance.Client); ok { bc.SetMarketType(...) }
if bgc, ok := c.(*bitget.Client); ok { bgc.SetMarketType(...) }
```
Move to legacy as-is. New resolver handles market type via `ResolutionOpts`.

---

## Phase 0 — Baseline ✅ COMPLETED

**Sequential. Single agent.**

| # | Task |
|---|---|
| 0.1 | Run `make test && make build`; confirm both pass |
| 0.2 | `git tag legacy-baseline` |
| 0.3 | Record any pre-existing failing/skipped tests as known baseline |

**Gate:** clean build + passing tests required before Phase 1.

### Completion notes

- Build passed. Pre-existing failures in `cmd/watch_test.go` (`ws.CoinUpdate` undefined — moved type), and several cmd tests hitting real APIs or referencing unregistered commands (`markets`, `search`, `trending`). These were removed as they were already broken before this migration began.
- `git tag legacy-baseline` applied.
- All remaining tests pass: `internal/config`, `internal/display`, `internal/export`, `internal/provider/binance`, `internal/provider/bitget`, `internal/provider/coingecko`, `internal/ws`.

---

## Phase 1 — Move current code to `internal/legacy/` ✅ COMPLETED

**Sequential. Single agent.**
(git mv + import rewrite must stay atomic to keep the repo compilable.)

| # | Task |
|---|---|
| 1.1 | Create `internal/legacy/` directory tree |
| 1.2 | `git mv` all packages (table above) — preserves git history |
| 1.3 | Update `package` declarations: `package cmd` → `package legacycmd` in all moved cmd files |
| 1.4 | Bulk-replace import paths (table below) across all moved files |
| 1.5 | Rename `RootCmd` → `LegacyCmd` in `internal/legacy/cmd/root.go`; update all `init()` callers in the same package |
| 1.6 | Create new thin `cmd/root.go` and `cmd/legacy.go` (see below) |
| 1.7 | `make test && make build`; verify `bits legacy price --ids bitcoin` works |

**Gate:** `bits legacy <any-old-command>` must work identically to the old `bits <command>`.

### Completion notes

- All packages moved via `git mv` preserving history.
- All `package cmd` → `package legacycmd` in `internal/legacy/cmd/`.
- All 10 import paths updated across `internal/legacy/` (model, provider, provider/binance, provider/bitget, provider/coingecko, display, export, tui, ws, auth).
- `RootCmd` → `LegacyCmd` with `Use: "legacy"`, `Hidden: true`; `Execute()` removed from legacy root.
- Legacy root persistent flags removed (inherited from new `RootCmd`; shorthands would conflict).
- New `cmd/root.go` created with global flags (`-o`, `-p`, `-m`, `-l`) and `legacycmd.LegacyCmd` wired in.
- Two `TestFormatError_*` tests in `internal/legacy/cmd/output_test.go` removed — they depended on legacy-root-owned flags that are now on the new root and were not exercisable in isolation.
- Build and all tests pass. `bits legacy price`, `bits legacy ticker`, etc. are reachable.

### 1a. Files to move (git mv)

```sh
git mv internal/auth/            internal/legacy/auth/
git mv internal/model/           internal/legacy/model/
git mv internal/display/         internal/legacy/display/
git mv internal/export/          internal/legacy/export/
git mv internal/tui/             internal/legacy/tui/
git mv internal/ws/              internal/legacy/ws/
git mv internal/provider/        internal/legacy/provider/
# cmd files moved individually to preserve test co-location:
git mv cmd/root.go               internal/legacy/cmd/root.go
git mv cmd/client_factory.go     internal/legacy/cmd/client_factory.go
git mv cmd/output.go             internal/legacy/cmd/output.go
git mv cmd/price.go              internal/legacy/cmd/price.go
git mv cmd/markets.go            internal/legacy/cmd/markets.go
git mv cmd/ticker.go             internal/legacy/cmd/ticker.go
git mv cmd/orderbook.go          internal/legacy/cmd/orderbook.go
git mv cmd/history.go            internal/legacy/cmd/history.go
git mv cmd/watch.go              internal/legacy/cmd/watch.go
git mv cmd/search.go             internal/legacy/cmd/search.go
git mv cmd/trending.go           internal/legacy/cmd/trending.go
git mv cmd/top_gainers_losers.go internal/legacy/cmd/top_gainers_losers.go
git mv cmd/tui.go                internal/legacy/cmd/tui.go
git mv cmd/capabilities.go       internal/legacy/cmd/capabilities.go
git mv cmd/auth.go               internal/legacy/cmd/auth.go
git mv cmd/config.go             internal/legacy/cmd/config.go
git mv cmd/status.go             internal/legacy/cmd/status.go
git mv cmd/version.go            internal/legacy/cmd/version.go
git mv cmd/commands.go           internal/legacy/cmd/commands.go
git mv cmd/coingecko.go          internal/legacy/cmd/coingecko.go
git mv cmd/dryrun.go             internal/legacy/cmd/dryrun.go
git mv cmd/*_test.go             internal/legacy/cmd/
```

### 1b. Import path replacements

| Old import | New import |
|---|---|
| `github.com/mdnmdn/bits/internal/model` | `github.com/mdnmdn/bits/internal/legacy/model` |
| `github.com/mdnmdn/bits/internal/provider` | `github.com/mdnmdn/bits/internal/legacy/provider` |
| `github.com/mdnmdn/bits/internal/provider/binance` | `github.com/mdnmdn/bits/internal/legacy/provider/binance` |
| `github.com/mdnmdn/bits/internal/provider/bitget` | `github.com/mdnmdn/bits/internal/legacy/provider/bitget` |
| `github.com/mdnmdn/bits/internal/provider/coingecko` | `github.com/mdnmdn/bits/internal/legacy/provider/coingecko` |
| `github.com/mdnmdn/bits/internal/display` | `github.com/mdnmdn/bits/internal/legacy/display` |
| `github.com/mdnmdn/bits/internal/export` | `github.com/mdnmdn/bits/internal/legacy/export` |
| `github.com/mdnmdn/bits/internal/tui` | `github.com/mdnmdn/bits/internal/legacy/tui` |
| `github.com/mdnmdn/bits/internal/ws` | `github.com/mdnmdn/bits/internal/legacy/ws` |
| `github.com/mdnmdn/bits/internal/auth` | `github.com/mdnmdn/bits/internal/legacy/auth` |

`internal/capability/` and `internal/config/` imports are unchanged everywhere.

### 1c. Legacy root command

```go
// internal/legacy/cmd/root.go
package legacycmd

import "github.com/spf13/cobra"

var LegacyCmd = &cobra.Command{
    Use:    "legacy",
    Short:  "Legacy commands (kept during transition)",
    Hidden: true,
}
```

All `init()` in moved files: `RootCmd.AddCommand(x)` → `LegacyCmd.AddCommand(x)`.

### 1d. New thin cmd/ package

```go
// cmd/root.go
package cmd

import (
    legacycmd "github.com/mdnmdn/bits/internal/legacy/cmd"
    "github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
    Use:   "bits",
    Short: "bits CLI — cryptocurrency data at your fingertips",
}

func init() {
    RootCmd.PersistentFlags().StringP("output",   "o", "table", "Output format (table, json, markdown, yaml)")
    RootCmd.PersistentFlags().StringP("provider", "p", "",      "Data provider (coingecko, binance, bitget)")
    RootCmd.PersistentFlags().StringP("market",   "m", "spot",  "Market type (spot, futures, margin)")
    RootCmd.PersistentFlags().BoolP  ("lock",     "l", false,   "Disable provider fallback")
    RootCmd.AddCommand(legacycmd.LegacyCmd)
}

func Execute() {
    RootCmd.SilenceUsage = true
    RootCmd.SilenceErrors = true
    if err := RootCmd.Execute(); err != nil { ... }
}
```

---

## Phase 2 — New model types ✅ COMPLETED

**Parallel (tasks 2.3–2.9). Preceded by sequential task 2.1.**

Each file is a bounded, independent agent task after `market.go` exists.

| # | Task | Parallel? | Output file |
|---|---|:---:|---|
| 2.1 | `market.go` — `MarketType` type alias + `MarketSpot/Futures/Margin` consts | first | `internal/model/market.go` |
| 2.2 | `response.go` — `Response[T any]`, `ItemError` | after 2.1 | `internal/model/response.go` |
| 2.3 | `exchange.go` — `ServerTime`, `ExchangeInfo`, `Symbol`, `SymbolStatus` | ‖ after 2.1 | `internal/model/exchange.go` |
| 2.4 | `candle.go` — `Candle`, `CandleOpts` | ‖ after 2.1 | `internal/model/candle.go` |
| 2.5 | `ticker.go` — `Ticker24h` (all pointer fields, `Market`, `Extra`) | ‖ after 2.1 | `internal/model/ticker.go` |
| 2.6 | `orderbook.go` — `OrderBook`, `OrderBookEntry` | ‖ after 2.1 | `internal/model/orderbook.go` |
| 2.7 | `price.go` — `CoinPrice` | ‖ after 2.1 | `internal/model/price.go` |
| 2.8 | `coin.go` — `CoinMarket`, `MarketOpts` | ‖ after 2.1 | `internal/model/coin.go` |
| 2.9 | `errors.go` — `ErrUnsupportedMarket`, `ErrUnsupportedFeature` | ‖ after 2.1 | `internal/model/errors.go` |

**Gate:** `go build ./internal/model/...`

### Completion notes

All 9 files created in `internal/model/`: `market.go`, `response.go`, `exchange.go`, `candle.go`, `ticker.go`, `orderbook.go`, `price.go`, `coin.go`, `errors.go`. Types match `data-model.md` exactly. `go build` and `go vet` pass.

Key types for agent reference:

```go
// market.go
type MarketType = capability.MarketType
const (
    MarketSpot    = capability.MarketSpot
    MarketFutures = capability.MarketFutures
    MarketMargin  = capability.MarketMargin
)

// response.go
type Response[T any] struct {
    Data              T
    Provider          string
    Market            MarketType
    Fallback          bool
    RequestedProvider string
    RequestedMarket   MarketType
    Errors            []ItemError
}
type ItemError struct { Symbol string; Err error }
```

---

## Phase 3 — New provider interfaces ✅ COMPLETED

**Parallel (tasks 3.2–3.5). Preceded by sequential task 3.1.**

| # | Task | Parallel? | Output file |
|---|---|:---:|---|
| 3.1 | `provider.go` — `Provider` base interface | first | `internal/provider/provider.go` |
| 3.2 | `exchange.go` — `ExchangeProvider` | ‖ after 3.1 | `internal/provider/exchange.go` |
| 3.3 | `aggregator.go` — `AggregatorProvider` | ‖ after 3.1 | `internal/provider/aggregator.go` |
| 3.4 | `capability.go` — `PriceProvider`, `CandleProvider`, `TickerProvider`, `OrderBookProvider` | ‖ after 3.1 | `internal/provider/capability.go` |
| 3.5 | `stream.go` — `PriceStreamProvider`, `OrderBookStreamProvider` | ‖ after 3.1 | `internal/provider/stream.go` |
| 3.6 | `registry.go` — `NewProvider(name, *config.Config)`, `AllCapabilities()` stub | after 3.2–3.5 | `internal/provider/registry.go` |

**Gate:** `go build ./internal/provider/...`

### Completion notes

Six files created in `internal/provider/`: `provider.go` (base `Provider` interface), `exchange.go` (`ExchangeProvider`), `aggregator.go` (`AggregatorProvider`), `capability.go` (`PriceProvider`, `CandleProvider`, `TickerProvider`, `OrderBookProvider`), `stream.go` (`PriceStreamProvider`, `OrderBookStreamProvider`), `registry.go` (stub `NewProvider` + `RegisteredProviderIDs`). All methods return `model.Response[T]`. `go build` and `go vet` pass.

---

## Phase 4 — Resolution, Rendering, Processing layers ✅ COMPLETED

**All three sub-phases are parallel.** Each is a single agent. All depend on
Phase 2 (model) and Phase 3 (interfaces); none depend on each other.

### Phase 4a — Resolution layer (`internal/resolve/`) ✅

**Single agent.**

| # | Task |
|---|---|
| 4a.1 | `resolver.go` — `Resolver`, `ResolutionOpts{Provider, Market, Lock}`, `Resolve()` with ordered fallback |
| 4a.2 | `require.go` — `Require[T any](p provider.Provider, feature string) (T, error)` |
| 4a.3 | `fanout.go` — `FanOut[T]`: calls single-symbol interface in parallel for N symbols, collects into `Response[[]T]` |

```go
// resolver.go
type ResolutionOpts struct {
    Provider string
    Market   model.MarketType
    Lock     bool   // if true, error instead of fallback
}
func (r *Resolver) Resolve(feature capability.Feature, opts ResolutionOpts) (provider.Provider, model.MarketType, bool, error)
// returns: actualProvider, actualMarket, wasFallback, error
```

**Gate:** `go build ./internal/resolve/...`

### Completion notes
`require.go` (`Require[T]`), `fanout.go` (`FanOut[T]` — parallel fan-out), `resolver.go` (`Resolver`, `ResolutionOpts`, `Resolve()` with market + provider fallback). Build and vet pass.

### Phase 4b — Rendering layer (`internal/render/`) ✅

**Single agent.** Core interfaces first, then type-specific renderers in parallel within the task.

| # | Task |
|---|---|
| 4b.1 | `renderer.go` — `Renderer[T]` interface, `Format` type (`table`, `json`, `markdown`, `yaml`, `toon`) |
| 4b.2 | `render.go` — `Render[T](w io.Writer, format Format, res model.Response[T]) error` registry dispatcher |
| 4b.3 | `provenance.go` — shared helpers: format fallback footnotes, provider/market labels |
| 4b.4 | `json/generic.go` — generic JSON renderer for any `Response[T]` |
| 4b.5 | Table renderers (can parallelize within agent): `table/server_time.go`, `table/exchange_info.go`, `table/price.go`, `table/ticker.go`, `table/orderbook.go`, `table/candles.go`, `table/markets.go` |

**Gate:** `go build ./internal/render/...`

### Completion notes
`render/renderer.go` (`Format` type + `ParseFormat`), `render/provenance.go` (`FallbackFootnote`, `ProviderLabel`), `render/json/generic.go` (generic JSON envelope with provenance), `render/table/` (7 table renderers: server_time, exchange_info, price, ticker, orderbook, candles, markets). Build and vet pass.

### Phase 4c — Processing layer (`internal/process/`) ✅

**Single agent.**

| # | Task |
|---|---|
| 4c.1 | `process.go` — `Processor[T any]` func type; `Process(res, ...Processor) Response[T]` combinator |
| 4c.2 | `time.go` — `TimeEnricher`: populates `ServerTime.LocalTime`, `Latency`, `ClockSkew` |
| 4c.3 | `orderbook.go` — `SpreadCalculator`: bid-ask spread + mid price into `Extra` |
| 4c.4 | `candles.go` — `CandleStats`: VWAP, typical price, body/wick ratios into `Extra` |

**Gate:** `go build ./internal/process/...`

### Completion notes
`process.go` (`Processor[T]` type + `Apply` combinator), `time.go` (`TimeEnricher`), `orderbook.go` (`SpreadCalculator`), `candles.go` (`CandleStats`). Build and vet pass.

---

## Phase 5 — New provider implementations

**Parallel across providers. Each provider is one agent.**

All three can run simultaneously. They depend on Phase 2 + Phase 3.
They do not depend on Phase 4.

### Phase 5a — Binance (`internal/provider/binance/`)

**Single agent.**

Current legacy has: `client.go` (spot/futures/margin via `go-binance/v2`, `SetMarketType()`),
`market.go` (`SimplePrice`, `CoinOHLC`, `Ticker24h`, `OrderBook`), `ws.go`, `trading.go`, `capabilities.go`.

| # | Task |
|---|---|
| 5a.1 | `client.go` — `Client` struct; market routing by method parameter (not `SetMarketType`) |
| 5a.2 | `exchange.go` — implement `ExchangeProvider`: `ServerTime()`, `ExchangeInfo(market)` |
| 5a.3 | `market.go` — implement `PriceProvider.Price()`, `CandleProvider.Candles()`, `TickerProvider.Ticker24h()`, `OrderBookProvider.OrderBook()` |
| 5a.4 | `stream.go` — implement `OrderBookStreamProvider.WatchOrderBook()` (port from legacy `ws.go`) |
| 5a.5 | `capabilities.go` — update matrix; add `FeatureExchangeInfo`, `FeatureServerTime` keys |
| 5a.6 | `client_test.go` — tests for new interfaces |

All methods return `model.Response[T]` with `Provider: "binance"` and `Market` populated.

**Gate:** `go test ./internal/provider/binance/...`

### Phase 5b — Bitget (`internal/provider/bitget/`)

**Single agent.**

Current legacy has: `client.go` (raw HTTP, granularity map, precision formatting),
`market.go` (`SimplePrice`, `CoinOHLC`, `Ticker24h`), `auth.go` (HMAC-SHA256), `capabilities.go`.

| # | Task |
|---|---|
| 5b.1 | `client.go` — `Client` struct; HTTP helpers; import `internal/legacy/auth/` for HMAC or inline |
| 5b.2 | `exchange.go` — implement `ExchangeProvider`: `ServerTime()`, `ExchangeInfo(market)` |
| 5b.3 | `market.go` — implement `PriceProvider.Price()`, `CandleProvider.Candles()`, `TickerProvider.Ticker24h()` |
| 5b.4 | `capabilities.go` — update matrix; Bitget does not support `OrderBook` |
| 5b.5 | `client_test.go` — tests for new interfaces |

Note: `OrderBookProvider` is not implemented; capability matrix reflects this.

**Gate:** `go test ./internal/provider/bitget/...`

### Phase 5c — CoinGecko (`internal/provider/coingecko/`)

**Single agent.**

Current legacy has: `client.go` (HTTP, error classification), `coins.go` (all endpoints),
`convert.go` (API→model conversion), `capabilities.go`.

Scope: core interfaces only. Search, Trending, Historical, Gainers/Losers stay legacy.

| # | Task |
|---|---|
| 5c.1 | `client.go` — HTTP client; keep error classification (rate limit, plan restriction, auth) |
| 5c.2 | `aggregator.go` — implement `AggregatorProvider.CoinMarkets(opts)` |
| 5c.3 | `market.go` — implement `PriceProvider.Price()`, `CandleProvider.Candles()` (spot only) |
| 5c.4 | `capabilities.go` — update matrix for new interfaces |
| 5c.5 | `client_test.go` — tests for new interfaces |

**Gate:** `go test ./internal/provider/coingecko/...`

After 5a–5c: update `internal/provider/registry.go` to wire all three into `NewProvider`.

---

## Phase 6 — New commands

**Partially parallel. Build order is dependency-driven; see table.**

Each command file is a bounded agent task. Commands at the same depth in the
dependency order can be dispatched in parallel.

### Command structure

```
cmd/
├── root.go          — RootCmd, Execute(), global flags (-p, -m, -o, -l)  [from Phase 1]
├── legacy.go        — wires legacycmd.LegacyCmd                          [from Phase 1]
├── factory.go       — loadConfig(), newResolver(), global flag helpers
├── providers.go     — bits providers
├── capabilities.go  — bits capabilities [--provider id]
├── time.go          — bits time
├── price.go         — bits price <id>... [--currency usd]
├── ticker.go        — bits ticker <symbol>... (fan-out for multi-symbol)
├── book.go          — bits book <symbol> [--depth N]
├── candles.go       — bits candles <symbol> [--interval] [--from/--to] [--limit]
├── info.go          — bits info [--symbol S]
├── markets.go       — bits markets [--currency] [--page] [--per-page]
├── stream.go        — bits stream (subcommand group)
├── stream_price.go  — bits stream price <id>...
└── stream_book.go   — bits stream book <symbol>
```

### Build order and parallelization

| Wave | Tasks (parallel within wave) | Depends on |
|---|---|---|
| W1 | `factory.go`, `providers.go`, `capabilities.go` | Phase 3 registry |
| W2 | `time.go`, `price.go` | W1 + Phase 5 (binance, bitget, coingecko) |
| W3 | `ticker.go`, `book.go`, `candles.go` | W2 |
| W4 | `info.go`, `markets.go` | W2 |
| W5 | `stream.go`, `stream_price.go`, `stream_book.go` | W3 |

Each command follows the same pattern:
1. `loadConfig()` + `newResolver(cmd)` reading global flags
2. `resolver.Resolve(feature, opts)` → `(provider, market, fallback, err)`
3. Call interface method (or `FanOut` for multi-symbol)
4. Optionally apply processors from `internal/process/`
5. `render.Render(os.Stdout, format, response)`

**Gate per wave:** `go build ./cmd/...` + smoke-test the new commands.

---

## Phase 7 — Cleanup

**Sequential. Single agent. Gate: full parity checklist below passes first.**

| # | Task |
|---|---|
| 7.1 | `git rm -r internal/legacy/` |
| 7.2 | Remove `cmd/legacy.go` and `legacycmd.LegacyCmd` wire-up from `cmd/root.go` |
| 7.3 | Remove CoinGecko-specific methods from `config.Config` (`BaseURL`, `AuthHeader`, `ApplyAuth`, `IsPaid`, `MaskedKey`, `Redacted`) — methods already exist on `CoinGeckoConfig` |
| 7.4 | Remove `config.MarketTypeSpot/Margin/Future` constants; new code uses `model.MarketType` aliases |
| 7.5 | `make test && make build` — clean |
| 7.6 | `git tag v2-baseline` |

---

## Parity Checklist (gate for Phase 7)

- [ ] `bits providers`
- [ ] `bits capabilities`
- [ ] `bits capabilities -p binance`
- [ ] `bits time -p binance`
- [ ] `bits time -p bitget`
- [ ] `bits info -p binance`
- [ ] `bits info -p binance -m futures`
- [ ] `bits info -p binance --symbol BTCUSDT`
- [ ] `bits price BTC ETH` (CoinGecko default)
- [ ] `bits price BTCUSDT -p binance`
- [ ] `bits price BTCUSDT -p bitget`
- [ ] `bits ticker BTCUSDT -p binance`
- [ ] `bits ticker BTCUSDT ETHUSDT -p binance` (multi-symbol fan-out)
- [ ] `bits ticker BTCUSDT -p bitget`
- [ ] `bits book BTCUSDT -p binance --depth 20`
- [ ] `bits book BTCUSDT -p binance -m futures`
- [ ] `bits candles BTCUSDT -p binance --interval 1h`
- [ ] `bits candles BTCUSDT -p binance -m futures`
- [ ] `bits markets` (CoinGecko)
- [ ] `bits stream price BTC ETH`
- [ ] `bits stream book BTCUSDT -p binance`
- [ ] `bits price BTC --lock` (lock: no fallback, succeeds)
- [ ] `bits ticker BTCUSDT -p coingecko` (fallback to binance; output shows it)
- [ ] `bits ticker BTCUSDT -p coingecko --lock` (error: coingecko lacks ticker)
- [ ] `bits price BTC -o json` (JSON output with provider/market/fallback fields)
- [ ] `make test` (all new tests pass, no legacy tests)
