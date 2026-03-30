# Manual Test Plan (Historical)

This document contains the historical manual test plan and known issues that were manually verified before the automated capability test tool was implemented.

### Known issues & tasks (History)

| # | Severity | Status | Description |
|---|----------|--------|-------------|
| 1 | High | ✅ fixed | **CoinGecko API key required** — unauthenticated requests return 404. Added early validation with helpful error message pointing to config/env var. Test plan updated to use proper coin IDs (`bitcoin` not `BTC`). |
| 2 | Medium | ✅ fixed | **Bitget candles spot** — `history-candles` endpoint requires `startTime`. Now routes to `/spot/market/candles` (no `From`) or `/spot/market/history-candles` (with `--from`). |
| 3 | Low | ✅ fixed | **`-p unknown_provider`** — now returns clear "unknown provider" error immediately, before any HTTP call. |
| 4 | Low | ✅ fixed | **Partial ticker** — `FanOut` now merges `Response.Errors` so INVALID symbols appear in the `errors` array. |

---

## Historical Results (2024-03-28)

### 0. Sanity
**Results: ✅ all ok**

### 1. Server Time (`ExchangeProvider`)
**Results: ✅ all ok**

### 2. Exchange Info (`ExchangeProvider`)
**Results: ✅ all ok**

### 3. Price (`PriceProvider`)
**Results: ✅ all ok**

### 4. Candles (`CandleProvider`)
**Results: ✅ all ok**

### 5. Ticker 24h (`TickerProvider`)
**Results: ✅ all ok**

### 6. Order Book (`OrderBookProvider`)
**Results: ✅ all ok**

### 7. Markets (`AggregatorProvider`)
**Results: ✅ all ok**

### 8. Streaming (`PriceStreamProvider` / `OrderBookStreamProvider`)
**Results: ✅ all ok (binance) / ⚠️ coingecko paid required**

### 9. Output Formats
**Results: ✅ all ok**

### 10. Edge Cases
**Results: ✅ all ok**
