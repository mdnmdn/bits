# WhiteBit Market Data APIs

## Price / Ticker

### Ticker (v1)

Returns 24h ticker statistics for a single market.

**Endpoint**: `GET /api/v1/public/ticker`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| market | string | Yes | Trading pair (e.g. `BTC_USD`) |

**Response**:
```json
{
  "success": true,
  "result": {
    "open": "49000",
    "high": "51000",
    "low": "48500",
    "last": "50000",
    "volume": "1000.5",
    "deal": "50000000",
    "bid": "49999",
    "ask": "50001",
    "change": "2.04"
  }
}
```

**Notes**:
- `change` is a percentage value
- `volume` is base asset volume
- `deal` is quote asset volume
- This is the primary price endpoint used by the bits provider

### Market Activity (v4)

Returns 24h activity for all markets.

**Endpoint**: `GET /api/v4/public/markets`

Returns the same data as the markets endpoint, with additional 24h volume/price info.

---

## Candles / Klines

### Kline Data

**Endpoint**: `GET /api/v1/public/kline`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| market | string | Yes | Trading pair |
| interval | string | Yes | `1m`, `3m`, `5m`, `15m`, `30m`, `1h`, `2h`, `4h`, `6h`, `8h`, `12h`, `1d`, `3d`, `1w`, `1M` |
| limit | int | No | Number of candles |
| start | long | No | Start time in Unix seconds |
| end | long | No | End time in Unix seconds |

**Sample Request**:
```
GET /api/v1/public/kline?market=BTC_USD&interval=1h&limit=100
```

**Response**:
```json
{
  "success": true,
  "result": [
    [1690196400, 50000, 50500, 51000, 49000, 100.5, 5050000]
  ]
}
```

**Column order**: `[timestamp_s, open, close, high, low, volume, amount]`

**Notes**:
- Timestamp is in **seconds** (not milliseconds)
- Column order is non-standard: `open, close, high, low` (not OHLC)
- `volume` is base asset, `amount` is quote asset

---

## Order Book / Depth

### Order Book

**Endpoint**: `GET /api/v4/public/orderbook/{market}`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| market | string | Yes | Trading pair (in path) |
| limit | int | No | 1-100 |

**Sample Request**:
```
GET /api/v4/public/orderbook/BTC_USD?limit=10
```

**Response**:
```json
{
  "timestamp": 1690196000,
  "ask": [
    [50001.0, 0.5],
    [50002.0, 1.0]
  ],
  "bid": [
    [49999.0, 0.5],
    [49998.0, 1.0]
  ]
}
```

**Notes**:
- Timestamp is in **seconds**
- Entry format: `[price, amount]`
- Response uses `ask`/`bid` (singular), not `asks`/`bids`

### Depth (Within 2%)

Returns price levels within +/-2% of last price.

**Endpoint**: `GET /api/v4/public/orderbook/{market}/depth`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| market | string | Yes | Trading pair (in path) |
| limit | int | No | Number of levels |

---

## Recent Trades

**Endpoint**: `GET /api/v4/public/trades/{market}`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| market | string | Yes | Trading pair (in path) |
| limit | int | No | Max 100 |

**Response**:
```json
[
  {
    "tradeID": 12345,
    "price": "50000",
    "amount": "0.5",
    "total": "25000",
    "timestamp": 1690196000,
    "type": "buy"
  }
]
```

---

## Futures-Specific Endpoints

### Futures Ticker

Futures data uses the same `/api/v4/public/futures` endpoint as the markets list.
Spot symbol `BTC_USDT` maps to futures ticker ID `BTC_PERP`.

**Endpoint**: `GET /api/v4/public/futures`

**Response**:
```json
{
  "success": true,
  "result": [{
    "ticker_id": "BTC_PERP",
    "last_price": "50000",
    "high": "51000",
    "low": "49000",
    "bid": "49999",
    "ask": "50001",
    "stock_volume": "1000.5",
    "money_volume": "50000000"
  }]
}
```


