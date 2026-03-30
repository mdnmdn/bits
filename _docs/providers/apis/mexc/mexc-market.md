# MEXC Market Data APIs

## Price / Ticker

### Price Ticker (Spot)

Returns the latest price for a symbol.

**Endpoint**: `GET /api/v3/ticker/price`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | No | Trading pair (e.g. `BTCUSDT`). Omit for all prices |

**Sample Request**:
```
GET /api/v3/ticker/price?symbol=BTCUSDT
```

**Response**:
```json
{
  "symbol": "BTCUSDT",
  "price": "46263.71"
}
```

**Weight**: 10

### 24hr Ticker (Spot)

**Endpoint**: `GET /api/v3/ticker/24hr`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | No | Trading pair. Omit for all tickers |

**Response**:
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

**Weight**: 25 (single symbol), 50 (all symbols)

### Futures Ticker

**Endpoint**: `GET /api/v1/contract/ticker`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | No | Contract symbol (e.g. `BTC_USDT`). Omit for all |

**Rate limit**: 10/2s

**Response**:
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

### Futures Index Price

**Endpoint**: `GET /api/v1/contract/index_price/{symbol}`

**Rate limit**: 20/2s

**Response**:
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

### Futures Funding Rate

**Endpoint**: `GET /api/v1/contract/funding_rate/{symbol}`

**Rate limit**: 20/2s

**Response**:
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

---

## Candles / Klines

### Spot Candles

**Endpoint**: `GET /api/v3/klines`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| interval | string | Yes | `1m`, `5m`, `15m`, `30m`, `60m`, `4h`, `1d`, `1w`, `1M` |
| startTime | long | No | Start time in ms |
| endTime | long | No | End time in ms |
| limit | int | No | Default 500, max 500 |

**Sample Request**:
```
GET /api/v3/klines?symbol=BTCUSDT&interval=1h&limit=100
```

**Response** (array of arrays):
```json
[
  [
    1640804880000,    // 0: Open time (ms)
    "47482.36",       // 1: Open
    "47482.36",       // 2: High
    "47416.57",       // 3: Low
    "47436.1",        // 4: Close
    "3.550717",       // 5: Volume (base)
    1640804940000,    // 6: Close time (ms)
    "168387.3"        // 7: Volume (quote)
  ]
]
```

**Weight**: 1

**Quirks** (documented in provider implementation):

1. **No time filters**: Returns NEWEST 500 candles (correct)
2. **Only startTime**: Returns OLDEST 500 candles (unexpected behavior)
3. **Only endTime**: Returns OLDEST candles up to that time (unexpected)
4. **endTime + limit**: Returns NEWEST N candles up to endTime (correct)
5. **startTime + endTime**: Returns correct range (correct)

**Workaround**: The bits provider fetches all 500 candles and filters client-side for reliability.

### Futures Candles

**Endpoint**: `GET /api/v1/contract/kline/{symbol}`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Contract symbol (in path) |
| interval | string | No | `Min1`, `Min5`, `Min15`, `Min30`, `Min60`, `Hour4`, `Hour8`, `Day1`, `Week1`, `Month1` |
| start | long | No | Start timestamp in **seconds** |
| end | long | No | End timestamp in **seconds** |

**Rate limit**: 20/2s

**Response** (parallel slices):
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

**Notes**:
- Max 2000 data points per request
- Timestamps in **seconds** (not ms)
- Response uses parallel arrays, not array-of-arrays

### Futures Index Price Candles

**Endpoint**: `GET /api/v1/contract/kline/index_price/{symbol}`

Same parameters as standard kline.

---

## Order Book / Depth

### Spot Order Book

**Endpoint**: `GET /api/v3/depth`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| limit | int | No | Default 100, max 5000 |

**Sample Request**:
```
GET /api/v3/depth?symbol=BTCUSDT&limit=20
```

**Response**:
```json
{
  "lastUpdateId": 1112416,
  "bids": [
    ["46260.38", "1.5"],
    ["46260.00", "2.3"]
  ],
  "asks": [
    ["46260.41", "0.8"],
    ["46260.50", "1.2"]
  ]
}
```

**Weight**: 3

### Futures Order Book

**Endpoint**: `GET /api/v1/contract/depth/{symbol}`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Contract symbol (in path) |
| limit | int | No | Number of rows |

**Rate limit**: 10/2s

**Response**:
```json
{
  "success": true,
  "code": 0,
  "data": {
    "asks": [
      [108779.2, 3240, 1],
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

**Notes**: Futures depth entry format is `[price, order_count, quantity]` (3 elements, not 2).

---

## Recent Trades

### Spot Trades

**Endpoint**: `GET /api/v3/trades`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| limit | int | No | Default 500, max 1000 |

**Weight**: 5

**Response**:
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

### Futures Trades

**Endpoint**: `GET /api/v1/contract/deals/{symbol}`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Contract symbol (in path) |
| limit | int | No | Max 100, default 100 |

**Rate limit**: 20/2s

**Response**:
```json
{
  "success": true,
  "code": 0,
  "data": [
    {
      "p": 109177.4,
      "v": 14,
      "T": 1,
      "O": 1,
      "M": 1,
      "t": 1761883066648
    }
  ]
}
```

**Field mapping**:

| Field | Description |
|-------|-------------|
| `p` | Price |
| `v` | Quantity |
| `T` | Side: 1=buy, 2=sell |
| `O` | Open flag: 1=open, 2=reduce, 3=other |
| `M` | Self-trade: 1=yes, 2=no |
| `t` | Timestamp (ms) |
