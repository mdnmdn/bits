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
| **binance** | spot | OK | WARN* | OK | OK | OK | OK | - | KO | OK |
| **binance** | futures | - | OK | - | OK | WARN | OK | - | - | OK |
| **binance** | margin | - | WARN* | - | OK | OK | OK | - | - | - |
| **bitget** | spot | OK | OK | OK | OK | OK | OK | - | KO | OK |
| **bitget** | futures | - | OK | - | OK | WARN | OK | - | WARN | KO |
| **bitget** | margin | - | OK | - | - | KO | - | - | - | - |
| **whitebit** | spot | OK | OK | OK | OK | OK | OK | - | KO | OK |
| **whitebit** | futures | - | OK | - | OK | WARN | OK | - | - | KO |
| **cryptocom** | spot | OK | WARN* | OK | OK | OK | OK | - | OK | OK |
| **cryptocom** | futures | - | WARN* | - | OK | OK | OK | - | - | - |
| **cryptocom** | margin | - | OK | - | - | OK | OK | - | - | - |
| **mexc** | spot | OK | WARN* | OK | OK | OK | OK | - | KO | KO |
| **mexc** | futures | - | OK | - | OK | WARN | OK | - | - | KO |
| **mexc** | margin | - | WARN* | - | OK | OK | OK | - | - | - |

---

## Issues by Provider

### CoinGecko

#### price — KO: no price data returned

`bits price BTCUSDT` fails with "no price data returned". CoinGecko's `/coins/markets` and price endpoints expect CoinGecko coin IDs (e.g. `bitcoin`, `ethereum`), not exchange-style symbols like `BTCUSDT`. The symbol resolver does not currently translate `BTCUSDT` → `bitcoin` for CoinGecko.

#### candles — KO: ExchangeInfo not supported

The symbol engine now validates symbols before calling the provider. For CoinGecko, loading the symbol list requires `ExchangeInfo`, which CoinGecko does not implement. Error: `provider coingecko does not support ExchangeInfo`. Root cause is unchanged: CoinGecko uses coin IDs (`bitcoin`), not exchange-style symbols (`BTCUSDT`), so the symbol engine cannot resolve them.

#### stream_price — KO: ExchangeInfo not supported

Same root cause as candles: symbol resolution fails before the stream is started because CoinGecko does not implement `ExchangeInfo`. Previously the error was `plan restricted: requires paid CoinGecko API key` (from the stream itself); now it surfaces earlier at the symbol-engine stage. The stream still requires a paid plan, but that error is masked by the resolution failure.

---

### Binance

#### exchange_info — WARN*: spot and margin return identical symbol counts (3556)

Both spot and margin `exchange_info` return 3556 symbols, suggesting the margin endpoint either returns the full spot list or the symbol engine is routing both requests to the same endpoint. The margin symbol universe should be a strict subset of spot.

#### ticker_24h (futures) — WARN: volume order of magnitude differs

Futures ticker volume is `176031.56` vs spot reference `2181.56`. This is expected: futures volume is expressed in contracts, not base-asset units. The reference comparison is not applicable to futures markets.

#### stream_price (spot) — KO: stream timeout (symbol issue)

`BTCUSDT` stream times out on spot. Likely a symbol translation issue — the symbol engine may not be mapping `BTCUSDT` to the correct Binance WebSocket stream name. See [`_docs/symbol-engine.md`](../symbol-engine.md).

---

### Bitget

#### ticker_24h (futures) — WARN: volume order of magnitude differs

Futures ticker volume is `52627.03` vs spot reference `2181.58`. Same expected behaviour as binance futures (contract denomination).

#### ticker_24h (margin) — KO: HTTP 404

`[bitget] HTTP 404: {"code":"40404","msg":"Request URL NOT FOUND"}`. The margin ticker endpoint either does not exist or uses a different URL path. The capability is registered but the underlying API call is hitting an invalid route.

#### stream_price (spot) — KO: partial receive then stream timeout (symbol issue)

Stream initially succeeds (3 ticks received) then times out. Bitget spot WebSocket accepts the subscription but drops the connection after a few updates. May be a keepalive or symbol format issue causing the stream to close. See [`_docs/symbol-engine.md`](../symbol-engine.md).

#### stream_price (futures) — WARN: stream timeout (symbol issue)

`BTCUSDT` on futures times out. Bitget futures use a different symbol format (e.g. `BTCUSDT_PERP`). The symbol engine must translate the input symbol to the futures native format before opening the stream. See [`_docs/symbol-engine.md`](../symbol-engine.md).

#### stream_order_book (futures) — KO: stream timeout (symbol issue)

Same root cause as `stream_price` on futures: symbol not translated to Bitget futures native format. Previously a WARN (timeout), now consistently erroring.

---

### WhiteBit

#### ticker_24h (futures) — WARN: volume order of magnitude differs

Futures ticker volume is `34198.23` vs spot reference `2181.58`. Same expected behaviour as other exchanges (contract denomination).

#### stream_price (spot) — KO: timestamp not increasing + stream timeout

Received ticks but their timestamps do not increase monotonically, followed by a stream timeout. The WhiteBit spot WebSocket may deliver updates without server-side timestamps, or the timestamp field is not populated correctly. The connection also drops, suggesting `BTCUSDT` is not resolved to `BTC_USDT` (WhiteBit native format) for the WebSocket subscription. See [`_docs/symbol-engine.md`](../symbol-engine.md).

#### stream_order_book (futures) — KO: stream timeout (symbol issue)

`BTCUSDT` on futures times out. WhiteBit futures use a `BTC_PERP`-style symbol. Symbol translation to the native futures format is needed for stream subscriptions. Previously a WARN, now consistently erroring.

---

### Crypto.com

#### exchange_info — WARN*: spot and futures return identical symbol counts (608)

Both spot and futures `exchange_info` return exactly 608 symbols. Crypto.com uses a single unified instrument list (spot symbols like `BTC_USDT` and perp symbols like `BTCUSD-PERP` are returned together). The market filter in the exchange info implementation may not be filtering by market type correctly.

---

### MEXC

#### exchange_info — WARN*: spot and margin return identical symbol counts (2388)

Both spot and margin return 2388 symbols, indicating the margin endpoint routes to the same spot list. MEXC margin symbols should be a subset.

#### ticker_24h (futures) — WARN: volume order of magnitude differs

Futures ticker volume is `194,047,975` vs spot reference `2181.66` — a difference of five orders of magnitude. This is much larger than other exchanges and may indicate the volume field is in a different unit (e.g. USDT notional rather than base-asset volume), or there is a field mapping error.

#### stream_price (spot) — KO: OriginalSymbol missing + stream timeout

Two errors: `OriginalSymbol missing (Symbol="BTCUSDT")` followed by stream timeout. The symbol engine fails to resolve `BTCUSDT` back to an original symbol for MEXC spot, causing a hard error before the stream is even established. This is a symbol-engine mapping error, not just a timeout.

#### stream_order_book (spot) — KO: stream timeout (symbol issue)

Order book stream for `BTCUSDT` on MEXC spot times out. Same likely root cause as `stream_price` — symbol not correctly resolved for MEXC WebSocket subscriptions.

#### stream_order_book (futures) — KO: stream timeout (symbol issue)

Order book stream for `BTCUSDT` on MEXC futures times out. Symbol translation to MEXC futures native format is needed for stream subscriptions.
