# Kraken Market Data

## Futures

### Candles

`GET /api/charts/v1/trade/{symbol}/{interval}?from=&to=&count=`

- Page cap: **2000** (`count`, also the default). `from` / `to` are seconds.
- Ordering: **oldest first** starting at `from`; `more_candles` is set when the window holds more, so a range wider than `count` yields the oldest ones. Page with windows of `count` candles.
- Bounds: `from` inclusive; the venue's `to` is inclusive, so the provider sends the last second before `To` (`To` exclusive).
- Intervals: `1m 5m 15m 30m 1h 4h 12h 1d 1w`. Others are an error.
- The charts API has no `result` field (unlike the derivatives API).

### 24h Ticker

`GET /derivatives/api/v3/tickers/{symbol}`. Volume is in the contract unit (base asset for linear `PF_` contracts); `QuoteVolume` is `volumeQuote`.

### Symbols

Native symbols (`PF_`, `PI_`, `FF_`, `FI_` prefixes) are only upper-cased. `BASE-QUOTE`, `BASE_QUOTE`, `BASE/QUOTE`, `BASEQUOTE` map to `PF_<BASE><QUOTE>` with BTC as XBT (`BTC-USD` is `PF_XBTUSD`). Output `Symbol` is `BASE-QUOTE`.

### Price

`GET /derivatives/api/v3/tickers/{symbol}`: last price, bid/ask with sizes (a zero size is left unset), 24h open/high/low/volume/change percent, `time`.

### Order book

`GET /derivatives/api/v3/orderbook?symbol=`. The venue returns the whole book, not sorted best-first (bids ascend); the provider sorts bids high to low and asks low to high, then cuts to `depth` (0 is the whole book). `Time` is the response `serverTime`.

### Funding rates

`GET /derivatives/api/v4/historicalfundingrates?symbol=`. Hourly settlements, ascending, whole history (no range or paging parameters), so `From`, `To` (both inclusive) and `Limit` are applied by the provider (with `From` the earliest `Limit`, without it the most recent). `Rate` is `relativeFundingRate`, the hourly ratio (0.0001 is 0.01%); the absolute `fundingRate` (quote units per contract) is not used. No mark price is published.

## Spot and margin

Base `https://api.kraken.com/0/public/`. Responses are `{"error": [...], "result": {...}}`, keyed by the long pair name (`XXBTZUSD`), not the requested one; the provider takes the single pair key. Error codes map: `EQuery:Unknown asset pair` not found, `EAPI:Rate limit` / `EGeneral:Too many` rate limit, other `EQuery` invalid request, `EService` server error.

### Candles (`OHLC`)

`GET /OHLC?pair=&interval=<minutes>&since=<s>`

- **History limit: the venue returns only the most recent 720 candles of an interval** (about 12 days of 1h, 720 days of 1d). `since` is exclusive and cannot reach further back: an older `since` still returns the latest 720. There is no deep history on this endpoint.
- The provider applies `From`, `To` (exclusive) and `Limit` itself. A range older than the window returns no candles (or only the overlap), never made-up data. For older spot history use the Kraken trade-history download (not covered).
- Intervals: `1m 5m 15m 30m 1h 4h 1d 1w 15d` (1, 5, 15, 30, 60, 240, 1440, 10080, 21600 minutes). There is no `12h`: error.
- Row: `[time, open, high, low, close, vwap, volume, count]`, `time` is the open time in seconds. The last row is the still-forming candle and is returned as sent; callers drop it.

### 24h Ticker (`Ticker`)

`GET /Ticker?pair=`. Last `c[0]`; 24h volume `v[1]` (base), high `h[1]`, low `l[1]`, vwap `p[1]`; bid `b`, ask `a`. Kraken has no price 24h ago: `OpenPrice` is `o`, the open of the current UTC day, and `PriceChange` / `PriceChangePercent` are measured from it (`Extra["open_basis"]` is `utc_day`). `QuoteVolume` is not published and stays unset.

### Price

From `Ticker`: last price `c[0]`, bid/ask with lot sizes, 24h volume/high/low, `Open24h` is the UTC-day open. `ID` is the altname, `Currency` the quote.

### Order book (`Depth`)

`GET /Depth?pair=&count=`. `count` is `depth`, capped at 500 (0 gives 100). Rows `[price, volume, unix seconds]`; `Time` is the newest row time.
