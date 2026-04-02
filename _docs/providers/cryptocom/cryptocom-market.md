# Crypto.com Market Data APIs

## Price / Ticker

### Tickers

Returns 24h ticker statistics for one or all instruments.

**Endpoint**: `GET /public/get-tickers`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| instrument_name | string | No | Instrument name (e.g. `BTC_USDT`). Omit for all |

**Sample Request**:
```
GET /public/get-tickers?instrument_name=BTC_USDT
```

**Response**:
```json
{
  "id": 1,
  "method": "public/get-tickers",
  "code": 0,
  "result": {
    "data": [{
      "i": "BTC_USDT",
      "v": 1000000,
      "vv": 50000000000,
      "l": 49000,
      "h": 51000,
      "o": 49000,
      "c": 50000,
      "p": 1000,
      "t": 10000
    }]
  }
}
```

**Field mapping**:

| Field | Description |
|-------|-------------|
| `i` | Instrument name |
| `v` | 24h volume (base) |
| `vv` | 24h volume value (quote) |
| `l` | Lowest price (24h) |
| `h` | Highest price (24h) |
| `o` | Open price |
| `c` | Last/close price |
| `p` | Price change |
| `t` | Trade count (24h) |

**Notes**: `c` (change) field is a **ratio** (e.g. 0.02 = 2%). Multiply by 100 for percentage.

---

## Candles / Klines

### Candlestick Data

**Endpoint**: `GET /public/get-candlestick`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| instrument_name | string | Yes | Instrument name |
| timeframe | string | Yes | `1m`, `5m`, `15m`, `30m`, `1h`, `4h`, `6h`, `12h`, `1d`, `1w` |

**Sample Request**:
```
GET /public/get-candlestick?instrument_name=BTC_USDT&timeframe=1h
```

**Response**:
```json
{
  "id": 1,
  "method": "public/get-candlestick",
  "code": 0,
  "result": {
    "data": [{
      "t": 1690196400000,
      "o": "50000",
      "h": "51000",
      "l": "49000",
      "c": "50500",
      "v": "100.5"
    }]
  }
}
```

**Field mapping**:

| Field | Description |
|-------|-------------|
| `t` | Timestamp (ms) |
| `o` | Open price |
| `h` | High price |
| `l` | Low price |
| `c` | Close price |
| `v` | Volume (base) |

---

## Order Book / Depth

### Order Book

**Endpoint**: `GET /public/get-book`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| instrument_name | string | Yes | Instrument name |
| depth | int | No | 1-50 |

**Sample Request**:
```
GET /public/get-book?instrument_name=BTC_USDT&depth=10
```

**Response**:
```json
{
  "id": 1,
  "method": "public/get-book",
  "code": 0,
  "result": {
    "instrument_name": "BTC_USDT",
    "bids": [["50000.0", "1.5"], ["49999.0", "2.3"]],
    "asks": [["50001.0", "0.5"], ["50002.0", "1.0"]]
  }
}
```

**Notes**:
- Entry format: `[price, quantity]`
- Sorted: bids descending, asks ascending

---

## Recent Trades

### Trades

**Endpoint**: `GET /public/get-trades`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| instrument_name | string | Yes | Instrument name |

**Response**:
```json
{
  "id": 1,
  "method": "public/get-trades",
  "code": 0,
  "result": {
    "data": [{
      "d": 12345,
      "s": "BUY",
      "p": "50000",
      "q": "0.5",
      "t": 1690196000000,
      "i": "BTC_USDT"
    }]
  }
}
```

**Field mapping**:

| Field | Description |
|-------|-------------|
| `d` | Trade ID |
| `s` | Side (`BUY`/`SELL`) |
| `p` | Price |
| `q` | Quantity |
| `t` | Timestamp (ms) |
| `i` | Instrument name |


