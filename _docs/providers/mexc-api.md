# MEXC API Documentation

MEXC is a global cryptocurrency exchange offering Spot and Futures (Perpetual) trading. This document focuses on market data endpoints for quotes, candles, and order books.

## Base URLs

| Type | URL |
|------|-----|
| Spot REST API | `https://api.mexc.com` |
| Futures REST API | `https://api.mexc.com` |
| Spot WebSocket | `wss://wbs-api.mexc.com/ws` |
| Futures WebSocket | `wss://contract.mexc.com/edge` |

---

## Spot Market Data

### Symbol Format

- Format: `BTCUSDT`, `ETHUSDT` (uppercase)
- Base asset + Quote asset concatenated

---

## Spot Quotes (Price Data)

### Price Ticker

```
GET /api/v3/ticker/price
```

Query parameters:
- `symbol` (optional) - Omit for all symbols

Response:
```json
{
  "symbol": "BTCUSDT",
  "price": "46263.71"
}
```

Weight: 10

### 24hr Ticker

```
GET /api/v3/ticker/24hr
```

Query parameters:
- `symbol` (optional) - Omit for all symbols

Response:
```json
{
  "symbol": "BTCUSDT",
  "priceChange": "184.34",
  "priceChangePercent": "0.4",
  "prevClosePrice": "46079.37",
  "lastPrice": "46263.71",
  "bidPrice": "46260.38",
  "bidQty": "",
  "askPrice": "46260.41",
  "askQty": "",
  "openPrice": "46079.37",
  "highPrice": "47550.01",
  "lowPrice": "45555.5",
  "volume": "1732.461487",
  "quoteVolume": "80417543.23",
  "openTime": 1641349500000,
  "closeTime": 1641349582808,
  "count": null
}
```

Weight: 25 (single), 50 (all)

### Book Ticker (Best Bid/Ask)

```
GET /api/v3/ticker/bookTicker
```

Query parameters:
- `symbol` (optional)

Response:
```json
{
  "symbol": "BTCUSDT",
  "bidPrice": "46260.38",
  "bidQty": "1.5",
  "askPrice": "46260.41",
  "askQty": "2.3"
}
```

Weight: 10

### Average Price

```
GET /api/v3/avgPrice?symbol=BTCUSDT
```

Response:
```json
{
  "mins": 5,
  "price": "46258.34"
}
```

Weight: 1

---

## Spot Candles (Kline)

### Candlestick Data

```
GET /api/v3/klines
```

Query parameters:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair (e.g., BTCUSDT) |
| interval | string | Yes | 1m, 5m, 15m, 30m, 60m, 4h, 1d, 1w, 1M |
| startTime | long | No | Start time in ms |
| endTime | long | No | End time in ms |
| limit | int | No | Default 500, max 500 |

Response (array of arrays):
```json
[
  [
    1640804880000,    // Open time (ms)
    "47482.36",       // Open
    "47482.36",       // High
    "47416.57",       // Low
    "47436.1",        // Close
    "3.550717",       // Volume (base)
    1640804940000,    // Close time (ms)
    "168387.3"       // Volume (quote)
  ]
]
```

Weight: 1

---

## Spot Order Book

### Depth (Order Book)

```
GET /api/v3/depth
```

Query parameters:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| limit | int | No | Default 100, max 5000 |

Response:
```json
{
  "lastUpdateId": 1112416,
  "bids": [
    ["46260.38", "1.5"],      // [price, quantity]
    ["46260.00", "2.3"],
    ["46259.50", "5.0"]
  ],
  "asks": [
    ["46260.41", "0.8"],
    ["46260.50", "1.2"],
    ["46261.00", "3.5"]
  ]
}
```

Weight: 3

### Recent Trades

```
GET /api/v3/trades
```

Query parameters:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| limit | int | No | Default 500, max 1000 |

Response:
```json
[
  {
    "id": null,
    "price": "46260.41",
    "qty": "0.5",
    "quoteQty": "23130.205",
    "time": 1640830579240,
    "isBuyerMaker": false,
    "isBestMatch": true
  }
]
```

Weight: 5

### Aggregate Trades

```
GET /api/v3/aggTrades
```

Query parameters:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| startTime | long | No | Start time in ms |
| endTime | long | No | End time in ms |
| limit | int | No | Default 500, max 1000 |

Response:
```json
[
  {
    "a": 12345,              // Aggregate trade ID
    "p": "46260.41",         // Price
    "q": "0.5",              // Quantity
    "T": 1640830579240,      // Timestamp
    "m": false,              // Is buyer maker
    "M": true                // Is best match
  }
]
```

Weight: 1

---

## Futures Market Data

### Symbol Format

- Format: `BTC_USDT`, `ETH_USDT` (underscore separator)
- Perpetual contracts only

---

## Futures Quotes

### Ticker

```
GET /api/v1/contract/ticker
```

Query parameters:
- `symbol` (optional) - Omit for all contracts

Response:
```json
{
  "success": true,
  "code": 0,
  "data": {
    "symbol": "BTC_USDT",
    "lastPrice": 109167.1,
    "bid1": 109167,
    "ask1": 109167.1,
    "volume24": 954830625,
    "amount24": 10374579341.00211,
    "holdVol": 381485808,
    "lower24Price": 106226,
    "high24Price": 111553.8,
    "riseFallRate": 0.014,
    "riseFallValue": 1510.6,
    "indexPrice": 109235,
    "fairPrice": 109168.9,
    "fundingRate": 0,
    "maxBidPrice": 120158.5,
    "minAskPrice": 98311.5,
    "timestamp": 1761883095759
  }
}
```

Rate limit: 10/2s

### Index Price

```
GET /api/v1/contract/index_price/{symbol}
```

Response:
```json
{
  "success": true,
  "code": 0,
  "data": {
    "symbol": "BTC_USDT",
    "indexPrice": 31103.4,
    "timestamp": 1609829705178
  }
}
```

Rate limit: 20/2s

### Fair Price

```
GET /api/v1/contract/fair_price/{symbol}
```

Response:
```json
{
  "success": true,
  "code": 0,
  "data": {
    "symbol": "BTC_USDT",
    "fairPrice": 31103.4,
    "timestamp": 1609829705178
  }
}
```

Rate limit: 20/2s

### Funding Rate

```
GET /api/v1/contract/funding_rate/{symbol}
```

Response:
```json
{
  "success": true,
  "code": 0,
  "data": {
    "symbol": "BTC_USDT",
    "fundingRate": 0.000018,
    "maxFundingRate": 0.0018,
    "minFundingRate": -0.0018,
    "collectCycle": 8,
    "nextSettleTime": 1761897600000,
    "timestamp": 1761879755894
  }
}
```

Rate limit: 20/2s

---

## Futures Candles

### Candlestick Data

```
GET /api/v1/contract/kline/{symbol}
```

Query parameters:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Contract symbol |
| interval | string | No | Min1, Min5, Min15, Min30, Min60, Hour4, Hour8, Day1, Week1, Month1 |
| start | long | No | Start timestamp (seconds) |
| end | long | No | End timestamp (seconds) |

Response:
```json
{
  "success": true,
  "code": 0,
  "data": {
    "time": [1761876000, 1761876900, 1761877800],
    "open": [109573.9, 109006.4, 109301.5],
    "close": [109006.4, 109301.5, 108725.9],
    "high": [109628.1, 109426.2, 109350.2],
    "low": [108953.3, 109006.4, 108666.2],
    "vol": [5587051.0, 5739575.0, 5945477.0],
    "amount": [6.106243567181E7, ...]
  }
}
```

Notes:
- Max 2000 data points per request
- Timestamp in seconds (not ms)

Rate limit: 20/2s

### Index Price Candles

```
GET /api/v1/contract/kline/index_price/{symbol}
```

Same parameters as standard kline.

### Fair Price Candles

```
GET /api/v1/contract/kline/fair_price/{symbol}
```

Same parameters as standard kline.

---

## Futures Order Book

### Depth (Order Book)

```
GET /api/v1/contract/depth/{symbol}
```

Query parameters:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Contract symbol |
| limit | int | No | Number of rows |

Response:
```json
{
  "success": true,
  "code": 0,
  "data": {
    "asks": [
      [108779.2, 3240, 1],   // [price, orders, quantity]
      [108779.3, 3884, 1]
    ],
    "bids": [
      [108779.1, 3240, 1],
      [108779, 3884, 1]
    ],
    "version": 28111438870,
    "timestamp": 1761879567135
  }
}
```

Note: `[price, order_count, quantity]`

Rate limit: 10/2s

### Depth Snapshots

```
GET /api/v1/contract/depth_commits/{symbol}/{limit}
```

Returns last N depth snapshots.

Response:
```json
{
  "success": true,
  "code": 0,
  "data": [
    {
      "asks": [...],
      "bids": [[3818.91, 272, 1]],
      "version": 26457599299
    }
  ]
}
```

Rate limit: 20/2s

### Recent Trades

```
GET /api/v1/contract/deals/{symbol}
```

Query parameters:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Contract symbol |
| limit | int | No | Max 100, default 100 |

Response:
```json
{
  "success": true,
  "code": 0,
  "data": [
    {
      "p": 109177.4,         // Price
      "v": 14,               // Quantity
      "T": 1,                // Side: 1=buy, 2=sell
      "O": 1,                // Open flag: 1=open, 2=reduce, 3=other
      "M": 1,                // Self-trade: 1=yes, 2=no
      "t": 1761883066648     // Timestamp
    }
  ]
}
```

Rate limit: 20/2s

---

## WebSocket Streams

### Spot WebSocket

Base: `wss://wbs-api.mexc.com/ws`

Uses Protocol Buffers (protobuf). See: https://github.com/mexcdevelop/websocket-proto

#### Subscribe Format

```json
{
  "method": "SUBSCRIPTION",
  "params": ["spot@public.kline.v3.api.pb@BTCUSDT@Min15"]
}
```

#### Unsubscribe

```json
{
  "method": "UNSUBSCRIPTION",
  "params": ["spot@public.kline.v3.api.pb@BTCUSDT@Min15"]
}
```

#### Spot K-line Stream

```
spot@public.kline.v3.api.pb@<symbol>@<interval>
```

Intervals: Min1, Min5, Min15, Min30, Min60, Hour4, Hour8, Day1, Week1, Month1

#### Spot Depth Stream

```
spot@public.aggre.depth.v3.api.pb@100ms@<symbol>
```

Push frequency: 100ms or 10ms

#### Spot Book Ticker

```
spot@public.aggre.bookTicker.v3.api.pb@100ms@<symbol>
```

#### Spot Trade Stream

```
spot@public.aggre.deals.v3.api.pb@100ms@<symbol>
```

---

### Futures WebSocket

Base: `wss://contract.mexc.com/edge`

Native JSON (no protobuf)

#### Subscribe

```json
{
  "method": "sub.ticker",
  "param": {"symbol": "BTC_USDT"}
}
```

#### Unsubscribe

```json
{
  "method": "unsub.ticker",
  "param": {"symbol": "BTC_USDT"}
}
```

#### Ping

```json
{"method": "ping"}
```

Note: Connection closes after 60s without ping. Send ping every 10-20s.

#### Futures Ticker Stream

```
sub.ticker / unsub.ticker
```

Response:
```json
{
  "channel": "push.ticker",
  "data": {
    "symbol": "BTC_USDT",
    "lastPrice": 109167.1,
    "bid1": 109167,
    "ask1": 109167.1,
    "volume24": 954830625,
    "holdVol": 381485808,
    "riseFallRate": 0.014,
    "indexPrice": 109235,
    "fairPrice": 109168.9,
    "timestamp": 1761883095759
  }
}
```

Push frequency: 1s (on trade)

#### Futures K-line Stream

```
sub.kline / unsub.kline
```

```json
{
  "method": "sub.kline",
  "param": {"symbol": "BTC_USDT", "interval": "Min60"}
}
```

Intervals: Min1, Min5, Min15, Min30, Min60, Hour4, Hour8, Day1, Week1, Month1

Response:
```json
{
  "channel": "push.kline",
  "data": {
    "symbol": "BTC_USDT",
    "interval": "Min60",
    "o": 6894.5,
    "c": 6885,
    "h": 6910.5,
    "l": 6885,
    "v": 1611754,
    "a": 233.74,
    "t": 1587448800
  }
}
```

#### Futures Depth Stream

```
sub.depth / unsub.depth
```

```json
{
  "method": "sub.depth",
  "param": {"symbol": "BTC_USDT"}
}
```

Response:
```json
{
  "channel": "push.depth",
  "data": {
    "asks": [[6859.5, 3251, 1]],
    "bids": [],
    "version": 96801927
  },
  "symbol": "BTC_USDT",
  "ts": 1587442022003
}
```

Note: `[price, order_count, quantity]`

Push frequency: 200ms

#### Futures Trade Stream

```
sub.deal / unsub.deal
```

```json
{
  "method": "sub.deal",
  "param": {"symbol": "BTC_USDT"}
}
```

Response:
```json
{
  "channel": "push.deal",
  "data": [
    {
      "p": 115309.8,
      "v": 55,
      "T": 2,
      "O": 3,
      "M": 1,
      "t": 1755487578276,
      "i": 13064218826
    }
  ]
}
```

---

## Exchange Information

### Spot

```
GET /api/v3/exchangeInfo
```

Response includes: symbols, status, precision, commission, filters

Weight: 25

### Futures

```
GET /api/v1/contract/detail
```

Response includes: contract details, leverage, fees, risk limits

Rate limit: 10/2s

---

## Rate Limits

### Spot REST

| Type | Limit |
|------|-------|
| IP-based | 300/10s |
| UID-based | 500/10s |

### Futures REST

| Endpoint | Limit |
|----------|-------|
| General | 20/2s |
| Ticker | 10/2s |
| Depth | 10/2s |
| Kline | 20/2s |

### WebSocket

| Type | Limit |
|------|-------|
| Requests | 100/s |
| Max subscriptions | 30/connection |
| Connection validity | 24h |

---

## Integration with bits CLI

### Market Data Mapping

| bits Capability | Spot Endpoint | Futures Endpoint |
|-----------------|---------------|------------------|
| Price | `/api/v3/ticker/price` | `/api/v1/contract/ticker` |
| Ticker | `/api/v3/ticker/24hr` | `/api/v1/contract/ticker` |
| OrderBook | `/api/v3/depth` | `/api/v1/contract/depth/{symbol}` |
| Candles | `/api/v3/klines` | `/api/v1/contract/kline/{symbol}` |
| Exchange Info | `/api/v3/exchangeInfo` | `/api/v1/contract/detail` |

### Symbol Conversion

| Direction | Format | Example |
|-----------|--------|---------|
| bits → MEXC Spot | uppercase | BTCUSDT |
| bits → MEXC Futures | underscore | BTC_USDT |

### WebSocket Streams

| bits Capability | Spot WS | Futures WS |
|-----------------|---------|------------|
| Stream Price | `spot@public.bookTicker.v3.api.pb@100ms@<symbol>` | `sub.ticker` |
| Stream Book | `spot@public.aggre.depth.v3.api.pb@100ms@<symbol>` | `sub.depth` |
| Stream Candles | `spot@public.kline.v3.api.pb@<symbol>@<interval>` | `sub.kline` |

---

## Error Codes

| Code | Description |
|------|-------------|
| -2011 | Unknown order |
| 400 | API key required |
| 401 | No authority |
| 403 | Access denied |
| 429 | Rate limit |
| 500 | Internal error |
| 10007 | Bad symbol |
| 30016 | Trading disabled |
| 700002 | Signature invalid |
| 700003 | Timestamp outside recvWindow |

---

## Resources

- [Spot API Docs](https://www.mexc.com/api-docs/spot-v3/introduction)
- [Futures API Docs](https://www.mexc.com/api-docs/futures/market-endpoints)
- [WebSocket Proto](https://github.com/mexcdevelop/websocket-proto)
- [Official SDK](https://github.com/mexcdevelop/mexc-api-sdk)
