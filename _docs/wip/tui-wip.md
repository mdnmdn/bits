# TUI Implementation Plan

## Overview

This document outlines a phased approach to implementing the TUI feature. Each phase is designed to be independently compilable and functional.

## rules
- at the end of each phase the project should compile and be formatted and vetted and make sure it runs
- at the end of each phase update this document marking the phase as completed and describing the updates (briefly but completely, no code only file and method references, no lines) in order to the next processing (with maybe no previous context) could start lean.
- when possible delegate tasks to subagents with simpler model (haiku based)
- remember: simplicity is the ultimate perfection

## Dependencies

The project already includes required TUI libraries:
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/charmbracelet/huh` - Form components

## Phase 1: Skeleton & Command Registration ✅ COMPLETED

**Goal**: Create basic TUI package structure and register the `tui` command

### Files Created

1. `cmd/tui.go` - Main TUI command with subcommands and flags
2. `internal/tui/tui.go` - Package init file
3. `internal/tui/app.go` - Bubble Tea app implementing tea.Model

### Verification
- ✅ `make build` compiles successfully
- ✅ `bits tui --help` shows help message
- ✅ `bits tui` launches and exits cleanly with `q`

---

## Phase 2: Provider Integration ✅ COMPLETED

**Goal**: Connect TUI to existing provider system

### Files Created

1. `internal/tui/provider.go` - TUI-specific provider wrapper
   - `NewTUIProvider()` - Creates provider from config
   - `GetProviderForFeature()` - Uses `resolve.Resolver` to select provider
   - `GetAvailableProviders(feature)` - Returns providers supporting a feature
   - `GetAvailableMarkets(provider, feature)` - Returns markets for provider+feature

2. `internal/tui/state.go` - TUI state management
   - `Section` type (prices, ticker, exchange, book, candles, markets, trending)
   - `TUIState` struct with Provider, Market, Filters, Items
   - `RefreshConfig` struct with interval and timer

### Files Created/Modified

1. `cmd/tui.go` - Section subcommands with feature mapping
2. `internal/tui/provider.go` - TUIProvider wrapper
3. `internal/tui/provider_iface.go` - ProviderWrapper interface
4. `internal/tui/state.go` - State types and RefreshConfig

### Verification
- ✅ `make build` compiles successfully
- ✅ Provider selection works via `cycleProvider()` method

---

## Phase 3: Prices Section (MVP) ✅ COMPLETED

**Goal**: Display price list with provider/market selection

### Files Created

1. `internal/tui/section/prices.go` - PricesModel implementation
2. `internal/tui/types.go` - SectionModel interface
   - `Table` struct with columns, rows, cursor
   - `View()` method with styling
   - Helper `NewTable()` constructor

3. `internal/tui/components/statusbar.go` - Status bar
   - `StatusBar` struct with provider, market, refresh info
   - `View()` method

### Files Created/Modified

1. `internal/tui/app.go` - Integrated PricesModel
2. `internal/tui/types.go` - SectionModel interface

### Verification
- ✅ `bits tui prices` shows price table
- ✅ `bits tui prices -p binance` shows Binance prices
- ✅ `j/k` or arrow keys navigate table
- ✅ `q` quits

---

## Phase 4: Multi-Provider Support ✅ COMPLETED

**Goal**: Enable provider switching and multi-provider comparison

### Files Created/Modified

1. `internal/tui/app.go` - Added `cycleProvider()` and `cycleMarket()` methods
   - `p` key cycles through available providers for current section
   - `m` key cycles through available markets

### Verification
- ✅ `p` key cycles providers
- ✅ `m` key cycles markets
- Press `p` to cycle providers
- Press `m` to toggle multi-provider view
- Shows price comparison when multiple providers selected

---

## Phase 5: Auto-Refresh ✅ COMPLETED

**Goal**: Add automatic data refresh with configurable intervals

### Files Created/Modified

1. `internal/tui/app.go` - Added refresh configuration and cycleRefresh()
   - `refreshInterval` field (5s, 10s, 30s, 1m)
   - `refreshEnabled` bool
   - `r` key cycles through intervals

### Verification
- ✅ `r` key cycles through refresh intervals
- ✅ Shows refresh indicator in header

**Goal**: Add 24h ticker display

### Files to Create

1. `internal/tui/section/ticker.go` - Ticker section
   - `TickerModel` with symbol, ticker data
   - Uses `TickerProvider.Ticker24h()`
   - Display: symbol, price, 24h high/low, volume, change

2. `cmd/tui.go` - Map ticker section
   - Section maps to `capability.FeatureTicker24h`
   - Provider filter: Binance, Bitget only

### Verification
- `bits tui ticker -p binance` shows ticker
- Auto-refresh works
- Shows error for unsupported providers

---

## Phase 7: Order Book Section

**Goal**: Add order book depth view

### Files to Create

1. `internal/tui/section/book.go` - Order book section
   - `BookModel` with symbol, bids, asks
   - Uses `OrderBookProvider.OrderBook()`
   - Display: bid/ask levels, spread calculation
   - Toggle bid/ask view with `s`

2. `cmd/tui.go` - Map book section
   - Section maps to `capability.FeatureOrderBook`
   - Provider filter: Binance spot only

### Verification
- `bits tui book BTCUSDT -p binance` shows order book
- Bid/ask toggle works

---

## Phase 8: Candles Section

**Goal**: Add OHLCV chart display

### Files to Create

1. `internal/tui/section/candles.go` - Candles section
   - `CandlesModel` with symbol, interval, candles
   - Uses `CandleProvider.Candles()`
   - Uses existing `ntcharts` for braille chart rendering
   - Interval selector: 1m, 5m, 15m, 1h, 4h, 1d

2. `cmd/tui.go` - Map candles section
   - Section maps to `capability.FeatureCandles`

### Verification
- `bits tui candles BTCUSDT -p binance` shows chart
- Interval selection works

---

## Phase 9: Exchange & Markets Sections

**Goal**: Add remaining sections

### Files to Create

1. `internal/tui/section/exchange.go` - Exchange info
   - Server time, exchange status
   - Uses `ExchangeProvider.ServerTime()`

2. `internal/tui/section/markets.go` - CoinGecko markets
   - Market listings with pagination
   - Uses `AggregatorProvider.CoinMarkets()`

### Files to Modify

1. `cmd/tui.go` - Add remaining section mappings
   - `exchange` → `capability.FeatureServerTime`
   - `markets` → `capability.FeatureMarketsList`

### Verification
- `bits tui exchange -p binance` works
- `bits tui markets -p coingecko` works

---

## Phase 10: Polish & Navigation

**Goal**: Finalize keyboard navigation and help

### Files to Modify

1. `internal/tui/app.go` - Final polish
   - Add `?` help overlay with all keybindings
   - Add `Tab` section cycling
   - Add `g` + number for page jumping
   - Improve error handling and loading states

2. `internal/tui/components/help.go` - Help component
   - Display keyboard shortcuts
   - Context-aware (show relevant keys for current section)

3. `cmd/tui.go` - Direct section jump
   - `bits tui exchange -p binance` jumps to section
   - Pass initial state from command flags

### Verification
- All sections accessible and functional
- Help shows correct shortcuts
- Direct navigation works

---

## Summary

| Phase | Focus | Key Files |
|-------|-------|-----------|
| 1 | Skeleton | `cmd/tui.go`, `internal/tui/app.go` |
| 2 | Provider Integration | `internal/tui/provider.go`, `state.go` |
| 3 | Prices MVP | `internal/tui/section/prices.go`, components |
| 4 | Multi-Provider | `selector.go`, prices enhancement |
| 5 | Auto-Refresh | `app.go` refresh loop |
| 6 | Ticker | `internal/tui/section/ticker.go` |
| 7 | Order Book | `internal/tui/section/book.go` |
| 8 | Candles | `internal/tui/section/candles.go` |
| 9 | Exchange/Markets | `exchange.go`, `markets.go` |
| 10 | Polish | Help, navigation, error handling |

## Implementation Notes

- Use subagents for simpler phases (3-6) with haiku model
- Test each phase with `make build && ./bits tui <section>`
- Keep views simple - leverage existing table renderers from `internal/render/table/`
- Reuse existing model types from `internal/model/`
- Delegate streaming to existing `internal/ws/` package when available
