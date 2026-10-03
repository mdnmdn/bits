# Bitvavo Market Data

## Candles

`GET /{market}/candles?interval=&start=&end=&limit=`

- Page cap: **1440** (`limit`, also the default).
- Ordering (observed on the live API): rows come **newest first**; `start` inclusive, `end` **exclusive**. When the range holds more than `limit` rows, the newest `limit` are returned, so page a range by moving `To` back or by windows of at most `limit` candles. The provider returns oldest first.
- Minutes without a trade have no candle (gaps).
- Intervals: `1m 5m 15m 30m 1h 2h 4h 6h 8h 12h 1d`. Others are an error.
- Row: `[ts, o, h, l, c, volume]`; a row without a volume is rejected.

## 24h Ticker

`GET /ticker/24h?market=`: last, open, high, low, bid, ask, volume (base), `volumeQuote`, open and close timestamps. Change is computed from open and last.

## Price

`GET /ticker/price?market=` per id, or `GET /ticker/price` (all markets, one call) when no ids are given. `ids` are market names (`BTC-EUR`, `BTCEUR`, `BTC_EUR`); `currency` is ignored (a market names its quote, returned in `Currency`). A failing id is an item error.

## Order Book

`GET /{market}/book?depth=`: default depth 20 in the provider, clamped to **1000**. `nonce` becomes `LastUpdateID`; `timestamp` is in nanoseconds and becomes `Time` (UTC). Levels are `[price, qty]` strings.
