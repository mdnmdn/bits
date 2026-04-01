# Bitget Market Data APIs

## Price / Ticker

### Spot Ticker

Returns 24h ticker statistics for spot pairs.

**Endpoint**: `GET /api/v2/spot/market/tickers`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | No | Trading pair (e.g. BTCUSDT) |

**Response**:
```json
{
  "code": "00000",
  "msg": "success",
  "data": [{
    "symbol": "BTCUSDT",
    "lastPr": "34413.1",
    "high24h": "37775.65",
    "low24h": "34413.1",
    "open": "35134.2",
    "baseVolume": "1234.56",
    "quoteVolume": "45678901.23",
    "bidPr": "34412.0",
    "askPr": "34414.0",
    "change24h": "0.00069",
    "ts": "1625125755277"
  }]
}
```

**Notes**: `change24h` is a ratio (e.g. 0.02 = 2%). Multiply by 100 for percentage.

### Futures Ticker

**Endpoint**: `GET /api/v2/mix/market/ticker`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| productType | string | Yes | `USDT-FUTURES` |

**Response**:
```json
{
  "code": "00000",
  "data": [{
    "symbol": "BTCUSDT",
    "lastPr": "1829.3",
    "askPr": "1829.8",
    "bidPr": "1829.3",
    "bidSz": "0.054",
    "askSz": "0.785",
    "high24h": "1850.0",
    "low24h": "1800.0",
    "ts": "1695794098184",
    "change24h": "0.01",
    "baseVolume": "100000",
    "quoteVolume": "182930000",
    "indexPrice": "1822.15",
    "fundingRate": "0.0001",
    "holdingAmount": "9488.49",
    "markPrice": "1829"
  }]
}
```

**Notes**: Futures uses `openUtc` (open at UTC 0) instead of `open` (24h open).

### Margin Ticker

**Endpoint**: `GET /api/v2/margin/market/tickers`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | No | Trading pair |
| productType | string | No | `isolated` or `cross` |

---

## Candles / Klines

### Spot Candles

**Endpoint**: `GET /api/v2/spot/market/candles`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| granularity | string | Yes | `1min`, `5min`, `15min`, `30min`, `1h`, `4h`, `1day` |
| startTime | long | No | Start time in ms |
| endTime | long | No | End time in ms |
| limit | int | No | Max 1000 |

### Spot Historical Candles

**Endpoint**: `GET /api/v2/spot/market/history-candles`

Same parameters as candles. Requires `startTime`.

### Futures Candles

**Endpoint**: `GET /api/v2/mix/market/history-candles`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| productType | string | Yes | `USDT-FUTURES` |
| granularity | string | Yes | `1m`, `5m`, `15m`, `30m`, `1H`, `4H`, `1D` |
| startTime | long | No | Start time in ms |
| endTime | long | No | End time in ms |
| limit | int | No | Max 1000 |

**Response** (all candle endpoints):
```json
{
  "code": "00000",
  "data": [
    ["1656604800000", "37834.5", "37849.5", "37773.5", "37773.5", "428.3462", "16198849.1079", "16198849.1079"]
  ]
}
```

**Column order**: `[timestamp_ms, open, high, low, close, base_volume, quote_volume, quote_volume_dup]`

---

## Order Book / Depth

### Spot Order Book

**Endpoint**: `GET /api/v2/spot/market/orderbook`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| type | string | No | `step0` to `step5` (aggregation levels) |
| limit | int | No | Max 150 |

**Response**:
```json
{
  "code": "00000",
  "data": {
    "asks": [["34567.15", "0.0131"], ["34567.25", "0.0144"]],
    "bids": [["34567", "0.2917"], ["34566.85", "0.0145"]],
    "ts": "1698303884584"
  }
}
```

### Futures Order Book

**Endpoint**: `GET /api/v2/mix/market/depth`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| productType | string | Yes | `USDT-FUTURES` |
| type | string | No | `step0` to `step5` |
| limit | int | No | Max 150 |

---

## VIP Fee Rate

### Spot VIP Fee Rate

**Endpoint**: `GET /api/v2/spot/market/vip-fee-rate`

**Description**: Returns VIP fee rate tiers. The `level` "0" represents the default rate for non-VIP users.

**Parameters**: None (public endpoint, no auth required)

**Response**:
```json
{
  "code": "00000",
  "msg": "success",
  "data": [
    {
      "level": "0",
      "dealAmount": "0",
      "assetAmount": "0",
      "takerFeeRate": "0.001",
      "makerFeeRate": "0.001",
      "btcWithdrawAmount": "2",
      "usdtWithdrawAmount": "50000"
    },
    {
      "level": "1",
      "dealAmount": "1000000",
      "assetAmount": "50000",
      "takerFeeRate": "0.0008",
      "makerFeeRate": "0.0006",
      "btcWithdrawAmount": "300",
      "usdtWithdrawAmount": "5000000"
    }
  ]
}
```

**Notes**:
- Fee rates are returned as decimal fractions (e.g., `0.001` = 0.1%)
- When `level` is "0", these are the default maker/taker fees for all spot symbols
- Higher VIP levels (1+) have lower fees based on 30-day trading volume and asset holdings

---

## Recent Trades

### Spot Trades

**Endpoint**: `GET /api/v2/spot/market/fills`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading pair |
| limit | int | No | Max 100 |

**Response**:
```json
{
  "code": "00000",
  "data": [{
    "symbol": "BTCUSDT",
    "tradeId": "12345",
    "price": "34567.15",
    "quantity": "0.0131",
    "side": "buy",
    "ts": "1698303884584"
  }]
}
```


