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
| 1 | High | ✅ fixed | **CoinGecko API key required** — unauthenticated requests return 404. Added early validation with helpful error message pointing to config/env var. Test plan updated to use proper coin IDs (`bitcoin` not `BTC`). |
| 2 | Medium | ✅ fixed | **Bitget candles spot** — `history-candles` endpoint requires `startTime`. Now routes to `/spot/market/candles` (no `From`) or `/spot/market/history-candles` (with `--from`). |
| 3 | Low | ✅ fixed | **`-p unknown_provider`** — now returns clear "unknown provider" error immediately, before any HTTP call. |
| 4 | Low | ✅ fixed | **Partial ticker** — `FanOut` now merges `Response.Errors` so INVALID symbols appear in the `errors` array. |

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

### CoinGecko (coin IDs) [key]

> CoinGecko uses coin IDs (`bitcoin`, `ethereum`), not ticker symbols (`BTC`, `ETH`). A demo API key is required — set `BITS_COINGECKO_API_KEY` or `coingecko.api_key` in config.

| Command | Expected | Result |
|---------|----------|--------|
| `bits price bitcoin` | ✓ data — bitcoin price in USD | ✅ ok |
| `bits price bitcoin ethereum` | ✓ data — two rows | ✅ ok |
| `bits price bitcoin --currency eur` | ✓ data — price in EUR | ✅ ok |
| `bits price bitcoin -o json` | ✓ json — `provider: "coingecko"`, `market: "spot"` | ✅ ok |

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
| `bits price bitcoin --lock` | ✓ data — coingecko supports price, no fallback needed | ✅ ok |
| `bits price BTCUSDT -p binance` | ✓ data — binance spot | ✅ ok |

---

## 4. Candles (`CandleProvider`)

### CoinGecko [key]

| Command | Expected | Result |
|---------|----------|--------|
| `bits candles bitcoin -p coingecko` | ✓ data — OHLCV rows (default interval) | ✅ ok |
| `bits candles bitcoin -p coingecko --limit 10` | ✓ data — ≤10 rows | ✅ ok |

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
| `bits candles BTCUSDT -p bitget` | ✓ data | ✅ ok |
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
| `bits markets` [key] | ✓ data — top 100 coins by market cap (CoinGecko default) | ✅ ok |
| `bits markets --currency eur` [key] | ✓ data — prices in EUR | ✅ ok |
| `bits markets --per-page 10` [key] | ✓ data — 10 rows | ✅ ok |
| `bits markets --page 2 --per-page 10` [key] | ✓ data — page 2 results | ✅ ok |
| `bits markets --order volume_desc` [key] | ✓ data — sorted by volume | ✅ ok |
| `bits markets -o json` [key] | ✓ json | ✅ ok |
| `bits markets -p binance` [key] | ✓ fallback — binance lacks markets_list → falls back to coingecko | ✅ ok |
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
| `bits price bitcoin -o json` [key] | `provider: "coingecko"`, `market: "spot"`, `fallback: false` | ✅ ok |
| `bits ticker BTCUSDT -p binance -o json` | `provider: "binance"`, `market: "spot"` | ✅ ok |
| `bits ticker BTCUSDT -p coingecko -o json` | `fallback: true`, `requested_provider: "coingecko"` | ✅ ok |
| `bits time -p binance -o json` | `data.time`, `data.latency`, `data.clock_skew` | ✅ ok |
| `bits book BTCUSDT -p binance -o json` | `data.bids`, `data.asks` arrays | ✅ ok |
| `bits markets -o json` [key] | data array of coin objects | ✅ ok |

---

## 10. Edge Cases

| Command | Expected | Result |
|---------|----------|--------|
| `bits price INVALID_COIN` | empty data — unknown coin ID returns empty from CoinGecko | ✅ ok — empty table (expected) |
| `bits ticker BTCUSDT INVALID -p binance` | ✓ partial — BTCUSDT succeeds; INVALID in `errors` array | ✅ ok |
| `bits candles BTCUSDT -p binance --from 2099-01-01` | data or empty (future date) | ✅ ok — empty table |
| `bits price BTC -p unknown_provider` | ✗ error — unknown provider | ✅ ok |
| `bits time` (no provider flag, no config) | ✓ fallback — resolves to first exchange provider | ✅ ok — served by binance |
