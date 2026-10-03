# OKX Market Data

## Candles

`GET /api/v5/market/history-candles?instId=&bar=&limit=&after=`

- Page cap: **300** (`limit`). Cursor `after` returns records older than the given ms timestamp. Rows come newest first; the provider pages backwards and returns oldest first.
- Bars: `1m 3m 5m 15m 30m`, `1H 2H 4H`, UTC variants `6Hutc 12Hutc 1Dutc 2Dutc 3Dutc 1Wutc 1Mutc` (bits intervals `6h 12h 1d 2d 3d 1w 1M` map to them, so they open at 00:00 UTC). Other intervals are an error.
- Bounds: `From` and `To` are both **inclusive** on the candle open time. With no `From`, the newest `Limit` candles at or before `To` (default 100). With `From` and `Limit`, the first `Limit` candles from `From`.
- Candles still forming (`confirm != 1`) are dropped.
- Volume: spot is base asset; SWAP uses `volCcy` (base coin), not contracts.
- Row: `[ts, o, h, l, c, vol, volCcy, volCcyQuote, confirm]`.

## 24h Ticker

`GET /api/v5/market/ticker?instId=`

- Change and percent are computed from `open24h` and `last`.
- Spot: `Volume` = `vol24h`, `QuoteVolume` = `volCcy24h`.
- SWAP: `Volume` = `volCcy24h` (base), `QuoteVolume` = base volume times last (approximation).
- `OpenTime` / `CloseTime` derive from `ts` (close) minus 24 h (open).

## Price

Uses `GET /api/v5/market/ticker` per id (no batch endpoint call). `ids` are trading symbols, `currency` is echoed only. Price has no market argument: an id with a `-SWAP` suffix is a perpetual swap, any other id is spot (margin has the same prices). Fields: last price, 24h open/high/low, bid/ask, 24h volume (spot `vol24h`, SWAP `volCcy24h` in base), change percent from open. A failing id goes to `Response.Errors`.

## Order Book

`GET /api/v5/market/books?instId=&sz=`

- `depth` is levels per side: default 20, capped at **400** (the endpoint maximum; `books-full` allows 5000 and is not used).
- Level: `[price, size, "0" (deprecated), number of orders]`. SWAP size is in contracts.
- `Time` is the snapshot `ts`; no update id is set.

## Funding Rates (SWAP)

`GET /api/v5/public/funding-rate-history?instId=&after=&limit=`

- Rows come newest first; the cursor `after` returns records older than the given `fundingTime` (ms). The provider pages with `limit=100` and returns ascending by time.
- `Rate` is `realizedRate` (falls back to `fundingRate`). No mark price is published.
- `opts.From` and `opts.To` are inclusive. With no `From`, the newest `Limit` entries (default 100). With `From` and `Limit`, the earliest `Limit` entries from `From`.
- OKX keeps only about the last 3 months of history. An empty window returns an empty list, not an error.

## Symbols

`BTC-USDT`, `BTCUSDT` and `BTC-USDT-SWAP` are accepted; `-SWAP` is added for the futures market. Spot and margin use the plain instId. Output `Symbol` is `BASE-QUOTE`, `OriginalSymbol` the `instId`.
