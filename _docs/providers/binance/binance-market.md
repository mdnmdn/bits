# Binance Market Data APIs

## Price / Ticker

### Price Ticker

Returns the latest price for a symbol.

**Endpoint**: `GET /api/v3/ticker/price`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | No | Symbol (e.g. BTCUSDT). Omit for all prices |

**Response**:
```json
{
  "symbol": "BTCUSDT",
  "price": "46263.71"
}
```

### 24hr Ticker Price Change

**Endpoint**: `GET /api/v3/ticker/24hr`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | No | Symbol. Omit for all tickers |

**Response**:
```json
{
  "symbol": "BTCUSDT",
  "priceChange": "184.34",
  "priceChangePercent": "0.4",
  "weightedAvgPrice": "46200.00",
  "prevClosePrice": "46079.37",
  "lastPrice": "46263.71",
  "lastQty": "0.001",
  "bidPrice": "46260.38",
  "bidQty": "1.5",
  "askPrice": "46260.41",
  "askQty": "2.3",
  "openPrice": "46079.37",
  "highPrice": "47550.01",
  "lowPrice": "45555.5",
  "volume": "1732.461487",
  "quoteVolume": "80417543.23",
  "openTime": 1641349500000,
  "closeTime": 1641349582808,
  "firstId": 12345,
  "lastId": 12400,
  "count": 55
}
```

### Book Ticker

Best price/qty on the order book for a symbol.

**Endpoint**: `GET /api/v3/ticker/bookTicker`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | No | Symbol. Omit for all book tickers |

**Response**:
```json
{
  "symbol": "BTCUSDT",
  "bidPrice": "46260.38",
  "bidQty": "1.5",
  "askPrice": "46260.41",
  "askQty": "2.3"
}
```

### Average Price

Current average price for a symbol.

**Endpoint**: `GET /api/v3/avgPrice`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Symbol |

**Response**:
```json
{
  "mins": 5,
  "price": "46258.34"
}
```

---

## Candles / Klines

**Endpoint**: `GET /api/v3/klines` (spot) / `GET /fapi/v1/klines` (futures)

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| interval | string | Yes | 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M |
| startTime | long | No | Start time in ms |
| endTime | long | No | End time in ms |
| limit | int | No | Default 500, max 1000 |

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
    "168387.3",       // 7: Quote asset volume
    100,              // 8: Number of trades
    "1000.0",         // 9: Taker buy base volume
    "1000.0",         // 10: Taker buy quote volume
    "0"               // 11: Ignore
  ]
]
```

---

## Order Book / Depth

**Endpoint**: `GET /api/v3/depth` (spot) / `GET /fapi/v1/depth` (futures)

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| limit | int | No | Default 100, max 5000. Valid: 5,10,20,50,100,500,1000,5000 |

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

---

## Recent Trades

**Endpoint**: `GET /api/v3/trades`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| limit | int | No | Default 500, max 1000 |

**Response**:
```json
[
  {
    "id": 12345,
    "price": "46260.41",
    "qty": "0.5",
    "quoteQty": "23130.205",
    "time": 1640830579240,
    "isBuyerMaker": false,
    "isBestMatch": true
  }
]
```

---

## Futures-Specific Endpoints

### Futures Ticker

**Endpoint**: `GET /fapi/v1/ticker/24hr`

Same parameters and response shape as spot 24hr ticker.

### Futures Index Price

**Endpoint**: `GET /fapi/v1/premiumIndex`

**Response**:
```json
{
  "symbol": "BTCUSDT",
  "markPrice": "46250.00",
  "indexPrice": "46248.50",
  "estimatedSettlePrice": "46249.00",
  "lastFundingRate": "0.0001",
  "nextFundingTime": 1641369600000,
  "interestRate": "0.0001",
  "time": 1641349500000
}
```


