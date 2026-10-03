# Bybit EU Market Data

## Candles

`GET /v5/market/kline?category=spot&symbol=&interval=&start=&end=&limit=`

- Page cap: **1000** (`limit`, also the default).
- Ordering: the venue returns rows **newest first** in `[start, end]` (both inclusive), at most `limit`; a wider window yields the newest ones. The provider sends `end = To - 1 ms` (so `To` is exclusive) and returns oldest first. Page with windows of `limit` candles.
- Intervals: `1m 3m 5m 15m 30m 1h 2h 4h 6h 12h 1d 1w 1M` (native `1 3 5 15 30 60 120 240 360 720 D W M`).
- Row: `[start, o, h, l, c, volume, turnover]`.

## 24h Ticker

`GET /v5/market/tickers?category=spot&symbol=`. A missing symbol is `not_found`.

## Symbols

`BTC-USDT`, `btc_usdt`, `BTC/USDT` all become `BTCUSDT`.

## Price

`GET /v5/market/tickers?category=spot&symbol=` per id. `ids` are symbols (`BTCUSDT`, `BTC-USDT`); `currency` is ignored and `Currency` stays empty. Fills 24h change (percent), high, low, open, volume, bid/ask and sizes. A failing symbol is an item error.

## Order Book

`GET /v5/market/orderbook?category=spot&symbol=&limit=`: default depth 20 in the provider, clamped to **200** (spot cap). `u` becomes `LastUpdateID`, `cts` becomes `Time` (UTC).

## Margin

`market=margin` serves the spot data above with `Market` set to `margin`.
