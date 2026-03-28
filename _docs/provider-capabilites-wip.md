# Provider Capabilities — Work In Progress

## Overview

This document outlines the plan to implement missing capabilities across exchange providers in `bits`.
The goal is to achieve parity across supported exchanges where API support exists.

### Target Capability Matrix

| FEATURE            | MARKET  | coingecko | binance | bitget | whitebit |
|--------------------|---------|-----------|---------|--------|----------|
| server_time        | spot    | -         | ✓       | ✓      | ✓        |
| exchange_info      | spot    | -         | ✓       | ✓      | ✓        |
| price              | spot    | ✓         | ✓       | ✓      | ✓        |
| price              | futures | -         | ✓       | ✓      | -        |
| price              | margin  | -         | ✓       | -      | -        |
| candles            | spot    | ✓         | ✓       | ✓      | ✓        |
| candles            | futures | -         | ✓       | ✓      | -        |
| candles            | margin  | -         | ✓       | -      | -        |
| ticker_24h         | spot    | -         | ✓       | ✓      | ✓        |
| ticker_24h         | futures | -         | ✓       | ✓      | -        |
| ticker_24h         | margin  | -         | ✓       | -      | -        |
| order_book         | spot    | -         | ✓       | **P1** | ✓        |
| order_book         | futures | -         | ✓       | **P1** | **P3**   |
| markets_list       | spot    | ✓         | -       | -      | -        |
| stream_price       | spot    | ✓         | **P5**  | **P2** | **P4**   |
| stream_order_book  | spot    | -         | ✓       | **P2** | **P4**   |
| stream_order_book  | futures | -         | ✓       | **P2** | **P4**   |

**Legend:**
- ✓ : Already implemented
- **PX** : Planned for Phase X
- `-` : Not planned / Not supported by provider API

## Rules
- At the end of each phase, the project must compile, be formatted, and vetted.
- At the end of each phase, update this document marking the phase as completed and describing updates (briefly, method/file references only).
- Delegate tasks to subagents when possible.
- Simplicity is the ultimate perfection.

---

## Phase 1 — Bitget Order Book (Spot & Futures) ✅ COMPLETED

**Goal:** Implement `OrderBookProvider` for Bitget.

**Completed:** Implemented `OrderBook(ctx, symbol, market, depth)` in `internal/provider/bitget/market.go`. It supports both spot and futures markets. Added `FeatureOrderBook` to `Capabilities()` in `internal/provider/bitget/client.go`. Verified the implementation by type assertion in `internal/provider/bitget/client_test.go`.

### Tasks
- [x] Implement `OrderBook(ctx, symbol, market, depth)` in `internal/provider/bitget/market.go`
  - Use `GET /api/v2/spot/market/orderbook` for spot.
  - Use `GET /api/v2/mix/market/depth` for futures.
- [x] Register `FeatureOrderBook` for both spot and futures in `internal/provider/bitget/client.go:Capabilities()`.
- [x] Verify with `./bits book BTCUSDT -p bitget` and `./bits book BTCUSDT -p bitget -m futures`.

---

## Phase 2 — Bitget WebSocket Streaming ✅ COMPLETED

**Goal:** Implement `PriceStreamProvider` and `OrderBookStreamProvider` for Bitget.

**Completed:** Created `internal/provider/bitget/stream.go`. Implemented `WatchPrices` (using `ticker` channel) and `WatchOrderBook` (using `depth` channel). Updated `Capabilities()` in `internal/provider/bitget/client.go` and verified via interface assertions in `internal/provider/bitget/client_test.go`.

### Tasks
- [x] Create `internal/provider/bitget/stream.go`.
- [x] Implement `WatchPrices(ctx, ids)` using Bitget v2 WebSocket (`ticker` channel).
- [x] Implement `WatchOrderBook(ctx, symbol, market, depth)` using Bitget v2 WebSocket (`depth` channel).
- [x] Update `Capabilities()` to include `FeatureStreamPrice` and `FeatureStreamOrderBook`.
- [x] Verify with `./bits watch BTCUSDT -p bitget`.

---

## Phase 3 — WhiteBit Futures Support ✅ COMPLETED

**Goal:** Extend WhiteBit provider to support futures market.

**Completed:** Updated `ExchangeInfo()` in `internal/provider/whitebit/exchange.go` to support `MarketFutures` via the `/api/v4/public/futures` endpoint. Updated all market data methods in `internal/provider/whitebit/market.go` to respect the `market` parameter. Registered futures capabilities in `internal/provider/whitebit/client.go`. Added `internal/provider/whitebit/client_test.go` with interface assertions.

### Tasks
- [x] Update `ExchangeInfo()` in `internal/provider/whitebit/exchange.go` to handle `MarketFutures`.
  - Use `GET /api/v4/public/futures`.
- [x] Implement Futures support for `Price`, `Candles`, `Ticker24h`, and `OrderBook` in `internal/provider/whitebit/market.go`.
- [x] Register futures capabilities in `internal/provider/whitebit/client.go`.
- [x] Verify with `-m futures` flag.

---

## Phase 4 — WhiteBit WebSocket Streaming ✅ COMPLETED

**Goal:** Implement streaming for WhiteBit.

**Completed:** Created `internal/provider/whitebit/stream.go`. Implemented `WatchPrices` and `WatchOrderBook` using the WhiteBit WebSocket API (`ticker_subscribe` and `depth_subscribe` methods). Updated `Capabilities()` in `internal/provider/whitebit/client.go` and verified via interface assertions in `internal/provider/whitebit/client_test.go`.

### Tasks
- [x] Create `internal/provider/whitebit/stream.go`.
- [x] Implement `WatchPrices` and `WatchOrderBook` using WhiteBit WebSocket API.
- [x] Update `Capabilities()` in `internal/provider/whitebit/client.go`.

---

## Phase 5 — Binance & WhiteBit Enhancements

**Goal:** Fill remaining gaps in streaming.

### Tasks
- [ ] Implement `WatchPrices` for Binance in `internal/provider/binance/stream.go`.
- [ ] Implement `WatchPrices` for WhiteBit (if not done in P4).
- [ ] Any other minor alignment identified during previous phases.

---

## Phase 6 — Final Verification & Cleanup

**Goal:** Ensure all implemented features are stable and documented.

### Tasks
- [ ] Run full test suite.
- [ ] Update `_docs/architecture.md` with final capability matrix.
- [ ] Final linting and formatting.
