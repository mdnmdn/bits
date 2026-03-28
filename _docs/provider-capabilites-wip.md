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
| price              | futures | -         | ✓       | ✓      | ✓        |
| price              | margin  | -         | ✓       | ✓      | -        |
| candles            | spot    | ✓         | ✓       | ✓      | ✓        |
| candles            | futures | -         | ✓       | ✓      | ✓        |
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
