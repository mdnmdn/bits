# Capabilities Test Plan

CLI smoke tests for all supported provider × feature × market combinations.
Run each command and verify the expected outcome.

**Legend**
- `[key]` — requires a valid API key in config / env
- `[paid]` — requires CoinGecko paid plan
- `✓ data` — expect tabular output with at least one row
- `✓ json` — expect JSON with `provider`, `market`, `data` keys
- `✓ fallback` — expect data + footnote `† served by …`
- `✗ error` — expect a non-zero exit and descriptive error message

**Result legend**
- `✅ ok` — passed as expected
- `❌ fail` — unexpected failure (see notes)
- `⚠️ expected` — failed for known/acceptable reason (e.g. paid plan)

## rules

- make sure to limit the time/response length to not fill the context
- in the check phase, if ok, mark as ok, if not ok, mark the exact command and the relevant error informations
- make the same agent run the same test on different provider in order to check the differences and mantain the coherence


---

## Last run: 2026-03-28

### Known issues & tasks

| # | Severity | Status | Description |
|---|----------|--------|-------------|
| 1 | High | 🔴 open | **CoinGecko API 404** — all direct CoinGecko price/markets/candles calls fail. Demo tier rate limit or endpoint change. Affects sections 3, 4, 7, 9, 10. |
| 2 | Medium | 🔴 open | **Bitget candles spot** — `Parameter verification failed`. Futures works. Investigate spot request params (symbol format or granularity field). |
| 3 | Low | 🔴 open | **`-p unknown_provider`** — returns 404 instead of a clear "unknown provider" error before any HTTP call. Registry lookup should fail early. |
| 4 | Low | 🔴 open | **Partial ticker** — `bits ticker BTCUSDT INVALID -p binance` silently drops INVALID instead of surfacing it in the `errors` array of the response. |

---

## 0. Sanity

```sh
bits --help
bits providers
bits capabilities
bits caps -p binance
bits caps -p coingecko
bits caps -p bitget
```

Expected: help text / tables with no errors.

**Results: ✅ all ok**

---

## 1. Server Time (`ExchangeProvider`)

| Command | Expected | Result |
|---------|----------|--------|
| `bits time -p binance` | ✓ data — server time + latency + clock skew | ✅ ok |
| `bits time -p bitget` | ✓ data — server time | ✅ ok |
| `bits time -p binance -m futures` | ✓ data — (market flag accepted, feature is market-agnostic) | ✅ ok |
| `bits time -p binance -o json` | ✓ json — `data.time`, `provider: "binance"` | ✅ ok |
| `bits time -p coingecko` | ✓ fallback — coingecko lacks server_time → falls back to binance/bitget | ✅ ok |
| `bits time -p coingecko --lock` | ✗ error — coingecko does not support server_time | ✅ ok |

---

## 2. Exchange Info (`ExchangeProvider`)

| Command | Expected | Result |
|---------|----------|--------|
| `bits info -p binance` | ✓ data — symbol table (spot) | ✅ ok (3549 symbols) |
| `bits info -p binance -m futures` | ✓ data — futures symbols | ✅ ok |
| `bits info -p binance --symbol BTCUSDT` | ✓ data — single row for BTCUSDT | ✅ ok |
| `bits info -p binance --symbol INVALID` | ✓ data — empty symbol list (no error) | ✅ ok |
| `bits info -p bitget` | ✓ data — bitget spot symbols | ✅ ok (715 symbols) |
| `bits info -p bitget -m futures` | ✓ data — bitget futures symbols | ✅ ok |
| `bits info -p binance -o json` | ✓ json — `data.exchange_id`, `data.symbols` array | ✅ ok |
| `bits info -p coingecko` | ✓ fallback — no exchange_info on coingecko | ✅ ok |
| `bits info -p coingecko --lock` | ✗ error | ✅ ok |

---

## 3. Price (`PriceProvider`)

### CoinGecko (coin IDs)

| Command | Expected | Result |
|---------|----------|--------|
| `bits price BTC` | ✓ data — bitcoin price in USD | ❌ fail — 404 Not Found |
| `bits price BTC ETH` | ✓ data — two rows | ❌ fail — 404 Not Found |
| `bits price BTC --currency eur` | ✓ data — price in EUR | ❌ fail — 404 Not Found |
| `bits price BTC -o json` | ✓ json — `provider: "coingecko"`, `market: "spot"` | ❌ fail — 404 Not Found |

### Binance (trading symbols)

| Command | Expected | Result |
|---------|----------|--------|
| `bits price BTCUSDT -p binance` | ✓ data — spot price | ✅ ok |
| `bits price BTCUSDT ETHUSDT -p binance` | ✓ data — two rows | ✅ ok |
| `bits price BTCUSDT -p binance -m futures` | ✓ data — futures price | ✅ ok |
| `bits price BTCUSDT -p binance -m margin` | ✓ data — margin price | ✅ ok |
| `bits price BTCUSDT -p binance -o json` | ✓ json | ✅ ok |

### Bitget (trading symbols)

| Command | Expected | Result |
|---------|----------|--------|
| `bits price BTCUSDT -p bitget` | ✓ data — spot price | ✅ ok |
| `bits price BTCUSDT -p bitget -m futures` | ✓ data — futures price | ✅ ok |

### Lock / fallback

| Command | Expected | Result |
|---------|----------|--------|
| `bits price BTC --lock` | ✓ data — coingecko supports price, no fallback needed | ❌ fail — 404 Not Found (CoinGecko API issue) |
| `bits price BTC -p binance` | ✓ data — binance spot (symbol resolved as coin ID) | ❌ fail — empty result (BTC not a valid Binance symbol) |

---

## 4. Candles (`CandleProvider`)

### CoinGecko

| Command | Expected | Result |
|---------|----------|--------|
| `bits candles bitcoin -p coingecko` | ✓ data — OHLCV rows (default interval) | ❌ fail — 404 Not Found |
| `bits candles bitcoin -p coingecko --limit 10` | ✓ data — ≤10 rows | ❌ fail — 404 Not Found |

### Binance

| Command | Expected | Result |
|---------|----------|--------|
| `bits candles BTCUSDT -p binance` | ✓ data — 1h candles (default) | ✅ ok (19 candles) |
| `bits candles BTCUSDT -p binance --interval 15m` | ✓ data — 15-minute candles | ✅ ok |
| `bits candles BTCUSDT -p binance --interval 1d --limit 7` | ✓ data — ≤7 daily candles | ✅ ok (7 candles) |
| `bits candles BTCUSDT -p binance --from 2024-01-01 --to 2024-01-07` | ✓ data — date-range | ✅ ok (169 hourly candles) |
| `bits candles BTCUSDT -p binance -m futures` | ✓ data — futures candles | ✅ ok |
| `bits candles BTCUSDT -p binance -o json` | ✓ json | ✅ ok |

### Bitget

| Command | Expected | Result |
|---------|----------|--------|
| `bits candles BTCUSDT -p bitget` | ✓ data | ❌ fail — `Parameter verification failed` |
| `bits candles BTCUSDT -p bitget -m futures` | ✓ data | ✅ ok |

---

## 5. Ticker 24h (`TickerProvider`)

### Binance

| Command | Expected | Result |
|---------|----------|--------|
| `bits ticker BTCUSDT -p binance` | ✓ data — last price, change%, high, low, volume | ✅ ok |
| `bits ticker BTCUSDT ETHUSDT -p binance` | ✓ data — two rows (fan-out) | ✅ ok |
| `bits ticker BTCUSDT -p binance -m futures` | ✓ data | ✅ ok |
| `bits ticker BTCUSDT -p binance -m margin` | ✓ data | ✅ ok |
| `bits ticker BTCUSDT -p binance -o json` | ✓ json — array under `data` | ✅ ok |

### Bitget

| Command | Expected | Result |
|---------|----------|--------|
| `bits ticker BTCUSDT -p bitget` | ✓ data | ✅ ok |
| `bits ticker BTCUSDT -p bitget -m futures` | ✓ data | ✅ ok |

### Fallback / lock

| Command | Expected | Result |
|---------|----------|--------|
| `bits ticker BTCUSDT -p coingecko` | ✓ fallback — footnote shows actual provider | ✅ ok — `† served by binance (requested: coingecko)` |
| `bits ticker BTCUSDT -p coingecko --lock` | ✗ error — coingecko does not support ticker_24h | ✅ ok |

---

## 6. Order Book (`OrderBookProvider`)

### Binance

| Command | Expected | Result |
|---------|----------|--------|
| `bits book BTCUSDT -p binance` | ✓ data — 20 bid/ask rows (default depth) | ✅ ok |
| `bits book BTCUSDT -p binance --depth 5` | ✓ data — 5 rows | ✅ ok |
| `bits book BTCUSDT -p binance --depth 50` | ✓ data — 50 rows | ✅ ok |
| `bits book BTCUSDT -p binance -m futures` | ✓ data — futures order book | ✅ ok |
| `bits book BTCUSDT -p binance -o json` | ✓ json — `data.bids`, `data.asks` | ✅ ok |

### Fallback / lock

| Command | Expected | Result |
|---------|----------|--------|
| `bits book BTCUSDT -p bitget` | ✓ fallback — bitget lacks order_book → falls back to binance | ✅ ok |
| `bits book BTCUSDT -p bitget --lock` | ✗ error | ✅ ok |
| `bits book BTCUSDT -p coingecko --lock` | ✗ error | ✅ ok |

---

## 7. Markets (`AggregatorProvider`)

| Command | Expected | Result |
|---------|----------|--------|
| `bits markets` | ✓ data — top 100 coins by market cap (CoinGecko default) | ❌ fail — 404 Not Found |
| `bits markets --currency eur` | ✓ data — prices in EUR | ❌ fail — 404 Not Found |
| `bits markets --per-page 10` | ✓ data — 10 rows | ❌ fail — 404 Not Found |
| `bits markets --page 2 --per-page 10` | ✓ data — page 2 results | ❌ fail — 404 Not Found |
| `bits markets --order volume_desc` | ✓ data — sorted by volume | ❌ fail — 404 Not Found |
| `bits markets -o json` | ✓ json | ❌ fail — 404 Not Found |
| `bits markets -p binance` | ✓ fallback — binance lacks markets_list → falls back to coingecko | ❌ fail — fallback also hits CoinGecko 404 |
| `bits markets -p binance --lock` | ✗ error | ✅ ok |

---

## 8. Streaming (`PriceStreamProvider` / `OrderBookStreamProvider`)

Tested with `timeout 5 ./bits stream ... 2>&1 | head -10`. OK = at least one update received.

| Command | Expected | Result |
|---------|----------|--------|
| `bits stream price BTC ETH` [paid] | ✓ live price lines for bitcoin / ethereum | ⚠️ expected — paid plan required |
| `bits stream price BTC -o json` [paid] | ✓ JSON lines with `id`, `price` | ⚠️ expected — paid plan required |
| `bits stream book BTCUSDT -p binance` | ✓ live order book update lines | ✅ ok (5 updates before timeout) |
| `bits stream book BTCUSDT -p binance --depth 5` | ✓ live updates | ✅ ok |
| `bits stream book BTCUSDT -p binance -m futures` | ✓ futures stream | ✅ ok |
| `bits stream price BTC -p binance --lock` | ✗ error — binance lacks stream_price | ✅ ok |
| `bits stream book BTCUSDT -p coingecko --lock` | ✗ error — coingecko lacks stream_order_book | ✅ ok |

---

## 9. Output Formats

Verify JSON provenance fields for one command per provider:

| Command | Expected | Result |
|---------|----------|--------|
| `bits price BTC -o json` | `provider: "coingecko"`, `market: "spot"`, `fallback: false` | ❌ fail — CoinGecko 404 |
| `bits ticker BTCUSDT -p binance -o json` | `provider: "binance"`, `market: "spot"` | ✅ ok |
| `bits ticker BTCUSDT -p coingecko -o json` | `fallback: true`, `requested_provider: "coingecko"` | ✅ ok |
| `bits time -p binance -o json` | `data.time`, `data.latency`, `data.clock_skew` | ✅ ok |
| `bits book BTCUSDT -p binance -o json` | `data.bids`, `data.asks` arrays | ✅ ok |
| `bits markets -o json` | data array of coin objects | ❌ fail — CoinGecko 404 |

---

## 10. Edge Cases

| Command | Expected | Result |
|---------|----------|--------|
| `bits price INVALID_COIN` | ✗ error or empty data (provider-dependent) | ❌ fail — CoinGecko 404 (not a clean "not found" error) |
| `bits ticker BTCUSDT INVALID -p binance` | ✓ partial — BTCUSDT succeeds; INVALID in `errors` array | ⚠️ partial — BTCUSDT succeeds but INVALID silently skipped (not in errors array) |
| `bits candles BTCUSDT -p binance --from 2099-01-01` | data or empty (future date) | ✅ ok — empty table |
| `bits price BTC -p unknown_provider` | ✗ error — unknown provider | ❌ fail — returns 404 instead of "unknown provider" error |
| `bits time` (no provider flag, no config) | ✓ fallback — resolves to first exchange provider | ✅ ok — served by binance |
