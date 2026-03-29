# Provider Capabilities — Work In Progress

## Overview

This document outlines the plan to implement missing capabilities across exchange providers in `bits`.
The goal is to achieve parity across supported exchanges where API support exists.

### Target Capability Matrix

| FEATURE            | MARKET  | coingecko | binance | bitget | whitebit |
|--------------------|---------|-----------|---------|--------|----------|
| server_time        | spot    | -         | ✓       | ✓      | ✓        |
| exchange_info      | spot    | -         | ✓       | ✓      | ✓        |
| exchange_info      | futures | -         | ✓       | ✓      | ✓        |
| price              | spot    | ✓         | ✓       | ✓      | ✓        |
| price              | futures | -         | ✓       | ✓      | -        |
| price              | margin  | -         | ✓       | ✓      | -        |
| candles            | spot    | ✓         | ✓       | ✓      | ✓        |
| candles            | futures | -         | ✓       | ✓      | -        |
| candles            | margin  | -         | ✓       | -      | -        |
| ticker_24h         | spot    | -         | ✓       | ✓      | ✓        |
| ticker_24h         | futures | -         | ✓       | ✓      | ✓        |
| ticker_24h         | margin  | -         | ✓       | ✓      | -        |
| order_book         | spot    | -         | ✓       | ✓      | ✓        |
| order_book         | futures | -         | ✓       | ✓      | ✓        |
| markets_list       | spot    | ✓         | -       | -      | -        |
| stream_price       | spot    | ✓         | ✓       | ✓      | ✓        |
| stream_order_book  | spot    | -         | ✓       | ✓      | ✓        |
| stream_order_book  | futures | -         | ✓       | ✓      | ✓        |

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

## Phase 1: WebSocket Infrastructure & Bitget Support ✅ COMPLETED

**Goal:** Establish common WS infrastructure and implement Bitget Order Book.

**Completed:**
- Created `internal/ws/base_client.go:Manager` as a common stateful WebSocket infrastructure with command support.
- Implemented `OrderBook(ctx, symbol, market, depth)` for Bitget in `internal/provider/bitget/market.go` (spot and futures).
- Updated `Capabilities()` in `internal/provider/bitget/client.go`.
- Added interface assertions in `internal/provider/bitget/client_test.go`.

---

## Phase 2: Bitget WebSocket & Margin Support ✅ COMPLETED

**Goal:** Implement Bitget WebSocket and Margin market data.

**Completed:**
- Implemented `WatchPrices` and `WatchOrderBook` for Bitget in `internal/provider/bitget/stream.go` using the new `ws.Manager`.
- Implemented Margin `ExchangeInfo` in `internal/provider/bitget/exchange.go`.
- Implemented Margin `Price` and `Ticker24h` in `internal/provider/bitget/market.go`.
- Updated `internal/config/config.go` with Bitget Margin settings.
- Updated Bitget `Capabilities()` for Margin and Streaming features.

---

## Phase 3: WhiteBit Futures & WebSocket Support ✅ COMPLETED

**Goal:** Extend WhiteBit to support Futures and WebSocket streaming.

**Completed:**
- Updated `ExchangeInfo()` for WhiteBit in `internal/provider/whitebit/exchange.go` to support futures.
- Updated REST market data methods in `internal/provider/whitebit/market.go` to respect the `market` parameter.
- Implemented stateful `WatchPrices` and `WatchOrderBook` for WhiteBit in `internal/provider/whitebit/stream.go` using `ws.Manager`.
- Updated WhiteBit `Capabilities()` for Futures and Streaming features.
- Added interface assertions in `internal/provider/whitebit/client_test.go`.

---

## Phase 4: Binance Refactor & Final Alignment ✅ COMPLETED

**Goal:** Refactor Binance WS and verify all changes.

**Completed:**
- Refactored Binance `WatchPrices` and `WatchOrderBook` in `internal/provider/binance/stream.go` to use the common `ws.Manager`.
- Verified all providers and markets (spot, futures, margin) for compilation and interface compliance.
- Updated `_docs/ws-handling.md` with specifications.
- Updated `_docs/providers/` documentation with margin/futures findings.

---

## Phase 5: Feedback Fixes & Final Polish ✅ COMPLETED

**Goal:** Address PR feedback and ensure correctness.

**Completed:**
- Fixed WhiteBit Futures `ExchangeInfo` parsing and `_PERP` symbol translation.
- Fixed WhiteBit Futures `Ticker24h` JSON field mapping (High/Low/Volume).
- Fixed WhiteBit WebSocket `lastprice_update` parsing.
- Refined Bitget `Capabilities()` to strictly match configuration.
- Verified Bitget Margin market data fetching.
- Removed unsupported capabilities (WhiteBit futures candles/price) to maintain parity with API reality.

---

## Verification Results (2026-03-29)

Verified implementation status of declared capabilities for Bitget and WhiteBit against CLI command wiring.

### Bitget

| Feature            | Market  | Declared | Implemented | Notes |
|--------------------|---------|----------|-------------|-------|
| server_time        | spot    | ✓        | ✓           | `exchange.go:ServerTime` |
| exchange_info      | spot    | ✓        | ✓           | `exchange.go:spotExchangeInfo` |
| exchange_info      | futures | ✓        | ✓           | `exchange.go:futuresExchangeInfo` |
| exchange_info      | margin  | ✓        | ✓           | `exchange.go:marginExchangeInfo` |
| price              | spot    | ✓        | ✓           | `market.go:Price()` |
| price              | futures | ✓        | ✗           | Declared but `market.go:Price()` hardcodes `MarketSpot` |
| price              | margin  | ✓        | ✗           | Declared but `market.go:Price()` hardcodes `MarketSpot` |
| candles            | spot    | ✓        | ✓           | `market.go:Candles()` |
| candles            | futures | ✓        | ✓           | `market.go:Candles()` |
| candles            | margin  | ✓        | ✗           | Not implemented (only spot/futures in `Candles()`) |
| ticker_24h         | spot    | ✓        | ✓           | `market.go:Ticker24h()` |
| ticker_24h         | futures | ✓        | ✓           | `market.go:Ticker24h()` |
| ticker_24h         | margin  | ✓        | ✓           | `market.go:Ticker24h()` via `fetchMarginTicker()` |
| order_book         | spot    | ✓        | ✓           | `market.go:OrderBook()` |
| order_book         | futures | ✓        | ✓           | `market.go:OrderBook()` |
| stream_price       | spot    | ✓        | ✓           | `stream.go:WatchPrices()` |
| stream_order_book  | spot    | ✓        | ✓           | `stream.go:WatchOrderBook()` |
| stream_order_book  | futures | ✓        | ✓           | `stream.go:WatchOrderBook()` |

**Issues Found:**
- `Price()` method ignores market parameter and always fetches spot data - needs update to support futures/margin or remove those capabilities from `Capabilities()` matrix
- Margin candles not implemented despite being declared

### WhiteBit

| Feature            | Market  | Declared | Implemented | Notes |
|--------------------|---------|----------|-------------|-------|
| server_time        | spot    | ✓        | ✓           | `exchange.go:ServerTime` |
| exchange_info      | spot    | ✓        | ✓           | `exchange.go:ExchangeInfo()` |
| exchange_info      | futures | ✓        | ✓           | `exchange.go:futuresExchangeInfo()` |
| price              | spot    | ✓        | ✓           | `market.go:Price()` |
| price              | futures | -        | -           | Correctly not declared |
| candles            | spot    | ✓        | ✓           | `market.go:Candles()` |
| candles            | futures | -        | -           | Correctly not declared |
| ticker_24h         | spot    | ✓        | ✓           | `market.go:Ticker24h()` |
| ticker_24h         | futures | ✓        | ✓           | `market.go:futuresTicker24h()` |
| order_book         | spot    | ✓        | ✓           | `market.go:OrderBook()` |
| order_book         | futures | ✓        | ✓           | `market.go:OrderBook()` |
| stream_price       | spot    | ✓        | ✓           | `stream.go:WatchPrices()` |
| stream_order_book  | spot    | ✓        | ✓           | `stream.go:WatchOrderBook()` |
| stream_order_book  | futures | ✓        | ✓           | `stream.go:WatchOrderBook()` |

**Status:** All declared capabilities are correctly implemented.

---

**Summary:**
- Bitget has 3 discrepancies: Price(futures), Price(margin), and Margin candles are declared in `Capabilities()` but not actually implemented in the respective methods
- WhiteBit is fully aligned - all declared capabilities are implemented
- Both providers are wired to CLI commands via the resolver pattern in `cmd/price.go`, `cmd/ticker.go`, `cmd/book.go`, `cmd/candles.go`, etc.
