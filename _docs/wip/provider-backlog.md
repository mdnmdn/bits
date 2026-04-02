# Provider Backlog

Tracks capability issues found by running `cmd/test-providers`. Generated on 2026-04-02.

## How to Update This Document

Run the full integration test and capture output:

```sh
go run ./cmd/test-providers 2>&1 | tee /tmp/test-providers.txt
```

Then review the output and update the table and issue sections below. Key interpretation rules:

- **server_time** `warn clock skew not calculated` — ignore; clock skew computation is not implemented for most providers.
- **stream timeout** — likely a symbol-engine mismatch, not a provider bug. The default symbol `BTCUSDT` may not be translated correctly to the provider-native format for futures/margin markets. See [`_docs/symbol-engine.md`](../symbol-engine.md).
- **exchange_info WARN\*** — spot and margin (or spot and futures) return identical symbol counts, suggesting the exchange info endpoint is not market-aware and returns the same data regardless of the requested market.
- **ticker_24h volume WARN (futures)** — futures volumes are denominated in contracts, not base-asset units. The volume magnitude will always differ from the spot reference. This is expected behaviour for futures markets and not a bug.

To run a targeted re-test for a specific provider/capability:

```sh
go run ./cmd/test-providers --provider cryptocom --capabilities stream_price,stream_order_book
go run ./cmd/test-providers --provider bitget --markets margin --capabilities ticker_24h
```

---

## Summary Matrix

Legend: `OK` = passing · `WARN` = warning · `WARN*` = same symbol count across markets (suspect) · `KO` = error · `-` = capability not available for this market

| Provider | Market | server_time | exchange_info | price | candles | ticker_24h | order_book | markets_list | stream_price | stream_order_book |
|----------|--------|:-----------:|:-------------:|:-----:|:-------:|:----------:|:----------:|:------------:|:------------:|:-----------------:|
| **coingecko** | spot | - | - | KO | KO | - | - | OK | KO | - |
| **binance** | spot | OK | WARN* | OK | OK | OK | OK | - | WARN | OK |
| **binance** | futures | - | OK | - | OK | WARN | OK | - | - | OK |
| **binance** | margin | - | WARN* | - | OK | OK | OK | - | - | - |
| **bitget** | spot | OK | OK | OK | OK | OK | OK | - | WARN | OK |
| **bitget** | futures | - | OK | - | OK | WARN | OK | - | WARN | WARN |
| **bitget** | margin | - | OK | - | - | KO | - | - | - | - |
| **whitebit** | spot | OK | OK | OK | OK | OK | OK | - | WARN | OK |
| **whitebit** | futures | - | OK | - | OK | WARN | OK | - | - | WARN |
| **cryptocom** | spot | OK | WARN* | OK | OK | OK | OK | - | OK | OK |
| **cryptocom** | futures | - | WARN* | - | OK | OK | OK | - | - | - |
| **cryptocom** | margin | - | OK | - | - | OK | OK | - | - | - |
| **mexc** | spot | OK | WARN* | OK | OK | OK | OK | - | WARN | WARN |
| **mexc** | futures | - | OK | - | OK | WARN | OK | - | - | - |
| **mexc** | margin | - | WARN* | - | OK | OK | OK | - | - | - |

---

## Issues by Provider

### CoinGecko

#### price — KO: no price data returned

`bits price BTCUSDT` fails with "no price data returned". CoinGecko's `/coins/markets` and price endpoints expect CoinGecko coin IDs (e.g. `bitcoin`, `ethereum`), not exchange-style symbols like `BTCUSDT`. The symbol resolver does not currently translate `BTCUSDT` → `bitcoin` for CoinGecko.

#### candles — KO: coin not found (HTTP 404)

Same root cause as price above. The candles endpoint also requires a CoinGecko coin ID. Request: `[coingecko] HTTP 404: {"error":"coin not found"}`.

#### stream_price — KO: plan restricted

Streaming requires a paid CoinGecko API key. Error: `plan restricted: requires paid CoinGecko API key`. This is a known limitation; the capability should either be gated on plan detection or documented as requiring Pro tier.

---

### Binance

#### exchange_info — WARN*: spot and margin return identical symbol counts (3556)

Both spot and margin `exchange_info` return 3556 symbols, suggesting the margin endpoint either returns the full spot list or the symbol engine is routing both requests to the same endpoint. The margin symbol universe should be a strict subset of spot.

#### ticker_24h (futures) — WARN: volume order of magnitude differs

Futures ticker volume is `170786.35` vs spot reference `2197.49`. This is expected: futures volume is expressed in contracts, not base-asset units. The reference comparison is not applicable to futures markets.

#### stream_price (spot) — WARN: stream timeout (symbol issue)

`BTCUSDT` stream times out on spot. Likely a symbol translation issue — the symbol engine may not be mapping `BTCUSDT` to the correct Binance WebSocket stream name. See [`_docs/symbol-engine.md`](../symbol-engine.md).

---

### Bitget

#### ticker_24h (futures) — WARN: volume order of magnitude differs

Futures ticker volume is `50957.24` vs spot reference `2197.49`. Same expected behaviour as binance futures (contract denomination).

#### ticker_24h (margin) — KO: HTTP 404

`[bitget] HTTP 404: {"code":"40404","msg":"Request URL NOT FOUND"}`. The margin ticker endpoint either does not exist or uses a different URL path. The capability is registered but the underlying API call is hitting an invalid route.

#### stream_price (futures) — WARN: stream timeout (symbol issue)

`BTCUSDT` on futures times out. Bitget futures use a different symbol format (e.g. `BTCUSDT_PERP`). The symbol engine must translate the input symbol to the futures native format before opening the stream. See [`_docs/symbol-engine.md`](../symbol-engine.md).

#### stream_order_book (futures) — WARN: stream timeout (symbol issue)

Same root cause as `stream_price` on futures: symbol not translated to Bitget futures native format.

---

### WhiteBit

#### ticker_24h (futures) — WARN: volume order of magnitude differs

Futures ticker volume is `34261.26` vs spot reference `2197.62`. Same expected behaviour as other exchanges (contract denomination).

#### stream_price (spot) — WARN: timestamp not increasing

Received ticks but their timestamps do not increase monotonically. The WhiteBit spot WebSocket may deliver updates without server-side timestamps, or the timestamp field is not populated correctly.

#### stream_price (spot) — WARN: stream timeout (symbol issue)

The stream eventually times out after the non-monotonic timestamp warning, suggesting `BTCUSDT` is not resolved to `BTC_USDT` (WhiteBit native format) for the WebSocket subscription. See [`_docs/symbol-engine.md`](../symbol-engine.md).

#### stream_order_book (futures) — WARN: stream timeout (symbol issue)

`BTCUSDT` on futures times out. WhiteBit futures use a `BTC_PERP`-style symbol. Symbol translation to the native futures format is needed for stream subscriptions.

---

### Crypto.com

#### exchange_info — WARN*: spot and futures return identical symbol counts (608)

Both spot and futures `exchange_info` return exactly 608 symbols. Crypto.com uses a single unified instrument list (spot symbols like `BTC_USDT` and perp symbols like `BTCUSD-PERP` are returned together). The market filter in the exchange info implementation may not be filtering by market type correctly.

---

### MEXC

#### exchange_info — WARN*: spot and margin return identical symbol counts (2388)

Both spot and margin return 2388 symbols, indicating the margin endpoint routes to the same spot list. MEXC margin symbols should be a subset.

#### ticker_24h (futures) — WARN: volume order of magnitude differs

Futures ticker volume is `192,309,418` vs spot reference `2197.63` — a difference of five orders of magnitude. This is much larger than other exchanges and may indicate the volume field is in a different unit (e.g. USDT notional rather than base-asset volume), or there is a field mapping error.

#### stream_price (spot) — WARN: stream timeout (symbol issue)

`BTCUSDT` stream times out after successfully receiving 3 ticks. Suggests the symbol is accepted initially but the subscription drops or the connection closes. May be a keepalive or reconnect issue in addition to or instead of a symbol problem.

#### stream_order_book (spot) — WARN: stream timeout (symbol issue)

Order book stream for `BTCUSDT` on MEXC spot times out. Same likely root cause as `stream_price`.
