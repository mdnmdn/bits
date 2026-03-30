# Crypto.com WebSocket Market Documentation (Spot)

## Reference
- **Official API Docs**: https://exchange-docs.crypto.com/exchange/v1/rest-ws/index.html#websocket-subscriptions
- **WebSocket Base URL (Production)**: `wss://stream.crypto.com/exchange/v1/market`
- **WebSocket Base URL (UAT Sandbox)**: `wss://uat-stream.3ona.co/exchange/v1/market`

## Protocol Overview
- **Protocol**: Raw WebSocket with JSON-RPC 2.0 format
- **Connection limits**: No explicit per-IP connection limit documented; use multiple connections for >400 subscriptions
- **Keep-alive / heartbeat**: Server sends `public/heartbeat` every **30 seconds**; client **must** respond with `public/respond-heartbeat` using the same `id` within **5 seconds**, or the connection will be terminated (close code 1000)
- **Reconnection guidelines**: Add a **1-second sleep** after establishing the WebSocket connection before sending any requests to avoid `TOO_MANY_REQUESTS` errors (rate limits are pro-rated based on the calendar-second the connection was opened)
- **Authentication**: **Not required** for public market data streams (trade, ticker, book, candlestick). Authentication via `public/auth` is only needed for private user-specific channels

## Message Format

### Request Format
All requests follow the JSON-RPC 2.0 structure:

```json
{
  "id": 1,
  "method": "subscribe",
  "params": {
    "channels": ["trade.BTC_USDT"]
  }
}
```

| Field  | Type   | Required | Description                                                                 |
|--------|--------|----------|-----------------------------------------------------------------------------|
| id     | long   | Y        | Request identifier (0 to 9,223,372,036,854,775,807). Response echoes this.  |
| method | string | Y        | `subscribe` or `unsubscribe`                                                |
| params | object | N        | Parameters for the method                                                   |

### Response Format (Initial Confirmation)

```json
{
  "id": 1,
  "method": "subscribe",
  "code": 0
}
```

| Field  | Type   | Description                                |
|--------|--------|--------------------------------------------|
| id     | long   | Original request identifier                |
| method | string | Method invoked (`subscribe` / `unsubscribe`) |
| code   | int    | `0` for success                            |
| result | object | Result object (present on data pushes)     |

### Notification Format (Data Push)

```json
{
  "id": -1,
  "method": "subscribe",
  "code": 0,
  "result": {
    "instrument_name": "BTC_USDT",
    "subscription": "trade.BTC_USDT",
    "channel": "trade",
    "data": [...]
  }
}
```

| Field            | Type   | Description                                              |
|------------------|--------|----------------------------------------------------------|
| id               | long   | Always `-1` for data push notifications                  |
| method           | string | Always `"subscribe"`                                     |
| code             | int    | `0` for success                                          |
| result           | object | Contains `instrument_name`, `subscription`, `channel`, `data` |

### Error Format

```json
{
  "id": -1,
  "method": "ERROR",
  "code": 40001,
  "message": "Bad request",
  "original": "{\"id\":1,\"method\":\"subscribe\",...}"
}
```

| Field    | Type   | Description                                          |
|----------|--------|------------------------------------------------------|
| id       | long   | `-1` if original request omitted `id`                |
| method   | string | `"ERROR"` if original request omitted `method`       |
| code     | int    | Error code (see Response and Reason Codes)           |
| message  | string | Human-readable error message                         |
| original | string | Original request as escaped string                   |

### WebSocket Termination Codes

| Code | Description                                                        |
|------|--------------------------------------------------------------------|
| 1000 | Normal disconnection (usually heartbeat not handled properly)      |
| 1006 | Abnormal disconnection                                             |
| 1013 | Server restarting -- try again later                               |

## WebSocket Endpoints

### Price Stream (Trade)

- **Description**: Publishes new trades for an instrument. Returns a snapshot of the last **50 trades** after initial subscription, then streams real-time trades. The `side` field represents the taker order's side.
- **Channel**: `trade.{instrument_name}`
- **Applies To**: Websocket (Market Data Subscriptions)

#### Subscription Request

```json
{
  "id": 1,
  "method": "subscribe",
  "params": {
    "channels": ["trade.BTC_USDT"]
  }
}
```

#### Response Fields

| Field            | Type   | Description                                      |
|------------------|--------|--------------------------------------------------|
| instrument_name  | string | e.g. `BTC_USDT`                                  |
| subscription     | string | `trade.{instrument_name}`                        |
| channel          | string | Always `trade`                                   |
| data             | array  | Array of trade objects (see below)               |

**Data object fields:**

| Field | Type   | Description                                        |
|-------|--------|----------------------------------------------------|
| d     | string | Trade ID                                           |
| t     | number | Trade timestamp (Unix milliseconds)                |
| p     | string | Trade price                                        |
| q     | string | Trade quantity                                     |
| s     | string | Side (`BUY` or `SELL`) — side of the taker order   |
| i     | string | Instrument name                                    |

#### Sample Response

```json
{
  "id": -1,
  "method": "subscribe",
  "code": 0,
  "result": {
    "instrument_name": "BTC_USDT",
    "subscription": "trade.BTC_USDT",
    "channel": "trade",
    "data": [{
      "d": "2030407068",
      "t": 1613581138462,
      "p": "51327.500000",
      "q": "0.000100",
      "s": "SELL",
      "i": "BTC_USDT"
    }]
  }
}
```

### 24h Ticker Stream

- **Description**: Publishes 24-hour rolling ticker statistics for an instrument, including high/low price, volume, best bid/ask, and price change.
- **Channel**: `ticker.{instrument_name}`
- **Applies To**: Websocket (Market Data Subscriptions)

#### Subscription Request

```json
{
  "id": 1,
  "method": "subscribe",
  "params": {
    "channels": ["ticker.BTC_USDT"]
  }
}
```

#### Response Fields

| Field            | Type   | Description                                      |
|------------------|--------|--------------------------------------------------|
| instrument_name  | string | e.g. `BTC_USDT`                                  |
| subscription     | string | `ticker.{instrument_name}`                       |
| channel          | string | Always `ticker`                                  |
| data             | array  | Array of ticker objects (see below)              |

**Data object fields:**

| Field | Type   | Description                                                        |
|-------|--------|--------------------------------------------------------------------|
| h     | string | Price of the 24h highest trade                                     |
| l     | string | Price of the 24h lowest trade (`null` if no trades)                |
| a     | string | Price of the latest trade (`null` if no trades)                    |
| c     | string | 24-hour price change ratio (`null` if no trades)                   |
| b     | string | Current best bid price (`null` if no bids)                         |
| bs    | string | Current best bid size (`null` if no bids)                          |
| k     | string | Current best ask price (`null` if no asks)                         |
| ks    | string | Current best ask size (`null` if no asks)                          |
| i     | string | Instrument name                                                    |
| v     | string | Total 24h traded volume                                            |
| vv    | string | Total 24h traded volume value (in USD)                             |
| oi    | string | Open interest (relevant for derivatives; `0` or absent for spot)   |
| t     | number | Timestamp (Unix milliseconds)                                      |

#### Sample Response

```json
{
  "id": -1,
  "method": "subscribe",
  "code": 0,
  "result": {
    "instrument_name": "BTC_USDT",
    "subscription": "ticker.BTC_USDT",
    "channel": "ticker",
    "data": [{
      "h": "51790.00",
      "l": "47895.50",
      "a": "51174.500000",
      "c": "0.03955106",
      "b": "51170.000000",
      "bs": "0.1000",
      "k": "51180.000000",
      "ks": "0.2000",
      "i": "BTC_USDT",
      "v": "879.5024",
      "vv": "26370000.12",
      "oi": "0",
      "t": 1613580710768
    }]
  }
}
```

### Order Book Depth Stream

- **Description**: Order book / L2 streaming data. Supports two subscription types: **SNAPSHOT** (full book on every update) and **SNAPSHOT_AND_UPDATE** (delta updates after initial snapshot). Delta subscription is recommended for reduced bandwidth/processing.
- **Channel**: `book.{instrument_name}.{depth}`
- **Applies To**: Websocket (Market Data Subscriptions)

#### Subscription Request (Snapshot — default)

```json
{
  "id": 1,
  "method": "subscribe",
  "params": {
    "channels": ["book.BTC_USDT.10"]
  }
}
```

#### Subscription Request (Delta / Snapshot and Update)

```json
{
  "id": 1,
  "method": "subscribe",
  "params": {
    "channels": ["book.BTC_USDT.10"],
    "book_subscription_type": "SNAPSHOT_AND_UPDATE",
    "book_update_frequency": 10
  }
}
```

#### Channel Parameters

| Parameter                  | Type   | Required | Description                                                                 |
|----------------------------|--------|----------|-----------------------------------------------------------------------------|
| instrument_name            | string | Y        | Must be formal symbol, e.g. `BTC_USDT`                                      |
| depth                      | int    | Y        | Allowed values: `10`, `50`                                                  |
| book_subscription_type     | string | N        | `SNAPSHOT` (default) or `SNAPSHOT_AND_UPDATE` (delta)                       |
| book_update_frequency      | int    | N        | Snapshot: `500` (default, ms). Delta: `10` (default) or `100` (ms)          |

#### Response Fields

| Field            | Type   | Description                                      |
|------------------|--------|--------------------------------------------------|
| instrument_name  | string | Same as requested `instrument_name`              |
| subscription     | string | Same as requested channel                        |
| channel          | string | `book` (snapshot) or `book.update` (delta)       |
| depth            | string | Same as requested depth                          |
| data             | array  | Array of book objects (see below)                |

**Snapshot data object (`channel: "book"`):**

| Field | Type    | Description                                      |
|-------|---------|--------------------------------------------------|
| bids  | array   | Array of `[price, size, order_count]` levels     |
| asks  | array   | Array of `[price, size, order_count]` levels     |
| tt    | integer | Epoch millis of last book update                 |
| t     | integer | Epoch millis of message publish                  |
| u     | integer | Update sequence number                           |

**Delta data object (`channel: "book.update"`):**

| Field  | Type    | Description                                      |
|--------|---------|--------------------------------------------------|
| update | object  | Contains `bids` and `asks` delta arrays          |
| tt     | integer | Epoch millis of last book update                 |
| t      | integer | Epoch millis of message publish                  |
| u      | integer | Current update sequence number                   |
| pu     | integer | Previous update sequence number                  |

**Level array format** (`[price, size, order_count]`):

| Index | Type   | Description                              |
|-------|--------|------------------------------------------|
| 0     | string | Price of the level                       |
| 1     | string | Total size of the level                  |
| 2     | string | Number of standing orders at this level  |

#### Sample Snapshot Response

```json
{
  "id": -1,
  "method": "subscribe",
  "code": 0,
  "result": {
    "instrument_name": "BTC_USDT",
    "subscription": "book.BTC_USDT.10",
    "channel": "book",
    "depth": 10,
    "data": [{
      "asks": [
        ["30082.5", "0.1689", "1"],
        ["30083.0", "0.1288", "1"],
        ["30084.5", "0.0171", "1"]
      ],
      "bids": [
        ["30079.0", "0.0505", "1"],
        ["30077.5", "1.0527", "2"],
        ["30076.0", "0.1689", "1"]
      ],
      "t": 1654780033786,
      "tt": 1654780033755,
      "u": 542048017824
    }]
  }
}
```

#### Sample Delta Update Response

```json
{
  "id": -1,
  "method": "subscribe",
  "code": 0,
  "result": {
    "instrument_name": "BTC_USDT",
    "subscription": "book.BTC_USDT.10",
    "channel": "book.update",
    "depth": 10,
    "data": [{
      "update": {
        "asks": [
          ["50126.000000", "0", "0"],
          ["50180.000000", "3.279000", "10"]
        ],
        "bids": [["50097.000000", "0.252000", "1"]]
      },
      "tt": 1647917463003,
      "t": 1647917463003,
      "u": 7845460002,
      "pu": 7845460001
    }]
  }
}
```

#### Snapshot vs Incremental Updates

- **SNAPSHOT** (default): After initial subscription, a full book snapshot is sent. Subsequent updates are full snapshots published at the requested interval (default 500ms), even if no change.
- **SNAPSHOT_AND_UPDATE** (delta): After initial full snapshot, only delta changes are published. Each update has a unique `u` (current sequence) and `pu` (previous sequence). An update should only be applied if `pu` matches the `u` of the last received update. On mismatch, re-subscribe to get a new snapshot.
- **Empty delta heartbeat**: For delta subscriptions, if no book changes occur, an empty delta (`"asks": [], "bids": []`) is sent after 5 seconds. The `u`/`pu` fields must still be processed.
- **Re-subscription**: To re-sync, issue another `subscribe` request. No need to `unsubscribe` first.

### Candlestick/Kline Stream

- **Description**: Publishes candlestick (k-line) data over a given period for an instrument.
- **Channel**: `candlestick.{time_frame}.{instrument_name}`
- **Applies To**: Websocket (Market Data Subscriptions)

#### Available Intervals

| Interval | Description                                    |
|----------|------------------------------------------------|
| `1m`     | One minute (legacy: `M1`)                      |
| `5m`     | Five minutes (legacy: `M5`)                    |
| `15m`    | 15 minutes (legacy: `M15`)                     |
| `30m`    | 30 minutes (legacy: `M30`)                     |
| `1h`     | One hour (legacy: `H1`)                        |
| `2h`     | Two hours (legacy: `H2`)                       |
| `4h`     | 4 hours (legacy: `H4`)                         |
| `12h`    | 12 hours (legacy: `H12`)                       |
| `1D`     | One day (legacy: `D1`, `1d`)                   |
| `7D`     | 1 week, starting Monday 00:00 UTC              |
| `14D`    | 2 weeks, starting Mon Oct 28 2019 00:00 UTC    |
| `1M`     | 1 month, starting first day of month 00:00 UTC |

Legacy formats (e.g., `M1`, `H1`, `D1`) are still supported until further notice.

#### Subscription Request

```json
{
  "id": 1,
  "method": "subscribe",
  "params": {
    "channels": ["candlestick.1D.BTC_USDT"]
  }
}
```

#### Response Fields

| Field            | Type   | Description                                      |
|------------------|--------|--------------------------------------------------|
| instrument_name  | string | e.g. `BTC_USDT`                                  |
| subscription     | string | `candlestick.{time_frame}.{instrument_name}`     |
| channel          | string | Always `candlestick`                             |
| interval         | string | The period (e.g. `1D`)                           |
| data             | array  | Array of candlestick objects (see below)         |

**Data object fields:**

| Field | Type   | Description                                      |
|-------|--------|--------------------------------------------------|
| o     | number | Open price                                       |
| h     | number | High price                                       |
| l     | number | Low price                                        |
| c     | number | Close price                                      |
| v     | number | Volume                                           |
| t     | long   | Start time of candlestick (Unix milliseconds)    |

#### Sample Response

```json
{
  "id": 1,
  "method": "subscribe",
  "code": 0,
  "result": {
    "instrument_name": "BTC_USDT",
    "subscription": "candlestick.1D.BTC_USDT",
    "channel": "candlestick",
    "interval": "1D",
    "data": [{
      "o": "51140.500000",
      "h": "51699.000000",
      "l": "49212.000000",
      "c": "51313.500000",
      "v": "867.9432",
      "t": 1612224000000
    }]
  }
}
```

## Constraints & Limits

| Constraint                              | Limit                                    |
|-----------------------------------------|------------------------------------------|
| Market Data WS rate limit               | 100 requests per second                  |
| Max subscriptions per WS connection     | 400 (error `EXCEED_MAX_SUBSCRIPTIONS`)   |
| Heartbeat interval                      | Server sends every 30 seconds            |
| Heartbeat response timeout              | Client must respond within 5 seconds     |
| Recommended post-connect delay          | 1 second before sending first request    |
| Nonce tolerance                         | Client clock must be within 60 seconds   |

## Notes

### Symbol Format Conventions
- **Spot markets**: `{BASE}_{QUOTE}` format, e.g. `BTC_USDT`, `ETH_USDC`, `CRO_USDT`
- **Perpetual contracts**: `{BASE}USD-PERP` or `{BASE}USDT-PERP`, e.g. `BTCUSD-PERP`
- **Futures**: `{BASE}USD-{YYMMDD}`, e.g. `BTCUSD-231124`
- **Index prices**: `{BASE}USD-INDEX`, e.g. `BTCUSD-INDEX`
- This documentation covers **spot markets only** — use the `{BASE}_{QUOTE}` format.

### Implementation Notes
- All **numbers in requests must be strings** wrapped in double quotes (e.g. `"12.34"`, not `12.34`). Responses may return numbers or strings depending on the field.
- The Market Data WebSocket is on a **separate endpoint** from the User API WebSocket. No authentication is needed for public market streams.
- If a single connection reaches the 400 subscription limit, establish additional connections for more channels.
- The `book.{instrument_name}` (default depth) subscription was **removed** in March 2025. Always use `book.{instrument_name}.{depth}` with an explicit depth.
- The `100ms` full snapshot frequency for book subscriptions was **removed** in February 2025. Use `500ms` for snapshots or switch to delta (`SNAPSHOT_AND_UPDATE`) for `10ms`/`100ms` updates.
- Market data wildcard ticker subscription was **removed** in December 2023. Use instrument-specific subscriptions.
- For delta book updates, always validate the `pu` (previous update) matches the last received `u` (current update). On mismatch, re-subscribe to get a fresh snapshot.
