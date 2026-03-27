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

## rules

- make sure to limit the time/response length to not fill the context
- in the check phase, if ok, mark as ok, if not ok, mark the exact command and the relevant error informations
- make the same agent run the same test on different provider in order to check the differences and mantain the coherence


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

---

## 1. Server Time (`ExchangeProvider`)

| Command | Expected |
|---------|----------|
| `bits time -p binance` | ✓ data — server time + latency + clock skew |
| `bits time -p bitget` | ✓ data — server time |
| `bits time -p binance -m futures` | ✓ data — (market flag accepted, feature is market-agnostic) |
| `bits time -p binance -o json` | ✓ json — `data.time`, `provider: "binance"` |
| `bits time -p coingecko` | ✓ fallback — coingecko lacks server_time → falls back to binance/bitget |
| `bits time -p coingecko --lock` | ✗ error — coingecko does not support server_time |

---

## 2. Exchange Info (`ExchangeProvider`)

| Command | Expected |
|---------|----------|
| `bits info -p binance` | ✓ data — symbol table (spot) |
| `bits info -p binance -m futures` | ✓ data — futures symbols |
| `bits info -p binance --symbol BTCUSDT` | ✓ data — single row for BTCUSDT |
| `bits info -p binance --symbol INVALID` | ✓ data — empty symbol list (no error) |
| `bits info -p bitget` | ✓ data — bitget spot symbols |
| `bits info -p bitget -m futures` | ✓ data — bitget futures symbols |
| `bits info -p binance -o json` | ✓ json — `data.exchange_id`, `data.symbols` array |
| `bits info -p coingecko` | ✓ fallback — no exchange_info on coingecko |
| `bits info -p coingecko --lock` | ✗ error |

---

## 3. Price (`PriceProvider`)

### CoinGecko (coin IDs)

| Command | Expected |
|---------|----------|
| `bits price BTC` | ✓ data — bitcoin price in USD |
| `bits price BTC ETH` | ✓ data — two rows |
| `bits price BTC --currency eur` | ✓ data — price in EUR |
| `bits price BTC -o json` | ✓ json — `provider: "coingecko"`, `market: "spot"` |

### Binance (trading symbols)

| Command | Expected |
|---------|----------|
| `bits price BTCUSDT -p binance` | ✓ data — spot price |
| `bits price BTCUSDT ETHUSDT -p binance` | ✓ data — two rows |
| `bits price BTCUSDT -p binance -m futures` | ✓ data — futures price |
| `bits price BTCUSDT -p binance -m margin` | ✓ data — margin price |
| `bits price BTCUSDT -p binance -o json` | ✓ json |

### Bitget (trading symbols)

| Command | Expected |
|---------|----------|
| `bits price BTCUSDT -p bitget` | ✓ data — spot price |
| `bits price BTCUSDT -p bitget -m futures` | ✓ data — futures price |

### Lock / fallback

| Command | Expected |
|---------|----------|
| `bits price BTC --lock` | ✓ data — coingecko supports price, no fallback needed |
| `bits price BTC -p binance` | ✓ data — binance spot (symbol resolved as coin ID) |

---

## 4. Candles (`CandleProvider`)

### CoinGecko

| Command | Expected |
|---------|----------|
| `bits candles bitcoin -p coingecko` | ✓ data — OHLCV rows (default interval) |
| `bits candles bitcoin -p coingecko --limit 10` | ✓ data — ≤10 rows |

### Binance

| Command | Expected |
|---------|----------|
| `bits candles BTCUSDT -p binance` | ✓ data — 1h candles (default) |
| `bits candles BTCUSDT -p binance --interval 15m` | ✓ data — 15-minute candles |
| `bits candles BTCUSDT -p binance --interval 1d --limit 7` | ✓ data — ≤7 daily candles |
| `bits candles BTCUSDT -p binance --from 2024-01-01 --to 2024-01-07` | ✓ data — date-range |
| `bits candles BTCUSDT -p binance -m futures` | ✓ data — futures candles |
| `bits candles BTCUSDT -p binance -o json` | ✓ json |

### Bitget

| Command | Expected |
|---------|----------|
| `bits candles BTCUSDT -p bitget` | ✓ data |
| `bits candles BTCUSDT -p bitget -m futures` | ✓ data |

---

## 5. Ticker 24h (`TickerProvider`)

### Binance

| Command | Expected |
|---------|----------|
| `bits ticker BTCUSDT -p binance` | ✓ data — last price, change%, high, low, volume |
| `bits ticker BTCUSDT ETHUSDT -p binance` | ✓ data — two rows (fan-out) |
| `bits ticker BTCUSDT -p binance -m futures` | ✓ data |
| `bits ticker BTCUSDT -p binance -m margin` | ✓ data |
| `bits ticker BTCUSDT -p binance -o json` | ✓ json — array under `data` |

### Bitget

| Command | Expected |
|---------|----------|
| `bits ticker BTCUSDT -p bitget` | ✓ data |
| `bits ticker BTCUSDT -p bitget -m futures` | ✓ data |

### Fallback / lock

| Command | Expected |
|---------|----------|
| `bits ticker BTCUSDT -p coingecko` | ✓ fallback — footnote shows actual provider |
| `bits ticker BTCUSDT -p coingecko --lock` | ✗ error — coingecko does not support ticker_24h |

---

## 6. Order Book (`OrderBookProvider`)

### Binance

| Command | Expected |
|---------|----------|
| `bits book BTCUSDT -p binance` | ✓ data — 20 bid/ask rows (default depth) |
| `bits book BTCUSDT -p binance --depth 5` | ✓ data — 5 rows |
| `bits book BTCUSDT -p binance --depth 50` | ✓ data — 50 rows |
| `bits book BTCUSDT -p binance -m futures` | ✓ data — futures order book |
| `bits book BTCUSDT -p binance -o json` | ✓ json — `data.bids`, `data.asks` |

### Fallback / lock

| Command | Expected |
|---------|----------|
| `bits book BTCUSDT -p bitget` | ✓ fallback — bitget lacks order_book → falls back to binance |
| `bits book BTCUSDT -p bitget --lock` | ✗ error |
| `bits book BTCUSDT -p coingecko --lock` | ✗ error |

---

## 7. Markets (`AggregatorProvider`)

| Command | Expected |
|---------|----------|
| `bits markets` | ✓ data — top 100 coins by market cap (CoinGecko default) |
| `bits markets --currency eur` | ✓ data — prices in EUR |
| `bits markets --per-page 10` | ✓ data — 10 rows |
| `bits markets --page 2 --per-page 10` | ✓ data — page 2 results |
| `bits markets --order volume_desc` | ✓ data — sorted by volume |
| `bits markets -o json` | ✓ json |
| `bits markets -p binance` | ✓ fallback — binance lacks markets_list → falls back to coingecko |
| `bits markets -p binance --lock` | ✗ error |

---

## 8. Streaming (`PriceStreamProvider` / `OrderBookStreamProvider`)

Streaming commands run until interrupted (Ctrl-C). Verify at least one update is printed.

| Command | Expected |
|---------|----------|
| `bits stream price BTC ETH` [paid] | ✓ live price lines for bitcoin / ethereum |
| `bits stream price BTC -o json` [paid] | ✓ JSON lines with `id`, `price` |
| `bits stream book BTCUSDT -p binance` | ✓ live order book update lines |
| `bits stream book BTCUSDT -p binance --depth 5` | ✓ live updates |
| `bits stream book BTCUSDT -p binance -m futures` | ✓ futures stream |
| `bits stream price BTC -p binance --lock` | ✗ error — binance lacks stream_price |
| `bits stream book BTCUSDT -p coingecko --lock` | ✗ error — coingecko lacks stream_order_book |

---

## 9. Output Formats

Verify JSON provenance fields for one command per provider:

```sh
bits price BTC -o json          # provider: "coingecko", market: "spot", fallback: false
bits ticker BTCUSDT -p binance -o json   # provider: "binance", market: "spot"
bits ticker BTCUSDT -p coingecko -o json # fallback: true, requested_provider: "coingecko"
bits time -p binance -o json             # data.time, data.latency, data.clock_skew
bits book BTCUSDT -p binance -o json     # data.bids, data.asks arrays
bits markets -o json                     # data array of coin objects
```

---

## 10. Edge Cases

| Command | Expected |
|---------|----------|
| `bits price INVALID_COIN` | ✗ error or empty data (provider-dependent) |
| `bits ticker BTCUSDT INVALID -p binance` | ✓ partial — BTCUSDT succeeds; INVALID in `errors` array |
| `bits candles BTCUSDT -p binance --from 2099-01-01` | data or empty (future date) |
| `bits price BTC -p unknown_provider` | ✗ error — unknown provider |
| `bits time` (no provider flag, no config) | ✓ fallback — resolves to first exchange provider |
