# WhiteBit WebSocket Market Documentation (Spot)

## Reference
- **Official API Docs**: https://docs.whitebit.com/websocket/overview
- **WebSocket Base URL**: `wss://api.whitebit.com/ws`
- **EU Region URL**: `wss://api.whitebit.eu/ws`
- **Protocol**: JSON-RPC 2.0 over WebSocket

## Protocol Overview
- **Protocol type**: Raw WebSocket with JSON-RPC 2.0 message framing
- **Connection limits**: 1,000 new connections per minute (global), reuse single connection for multiple subscriptions
- **Keep-alive**: Server closes connection after 60 seconds of inactivity. Send `ping` every 50 seconds.
- **Reconnection guidelines**: Use exponential backoff with jitter (1s, 2s, 4s, 8s). Re-subscribe to all channels after reconnect.
- **Authentication**: Not required for public market streams

### Ping/Pong Mechanism

```json
// Request
{ "id": 0, "method": "ping", "params": [] }

// Response
{ "id": 0, "result": "pong", "error": null }
```

## Message Format

### Request Format

All requests follow JSON-RPC structure:

| Field    | Type    | Description                                            |
| -------- | ------- | ------------------------------------------------------ |
| `id`     | Integer | Unique identifier to match response to request         |
| `method` | String  | Method name (e.g., `lastprice_subscribe`)              |
| `params` | Array   | Parameters for the method                              |

### Response Format

| Field    | Type          | Description                                              |
| -------- | ------------- | -------------------------------------------------------- |
| `id`     | Integer       | ID of the original request (null for update events)      |
| `result` | Object/null   | Result payload on success                                |
| `error`  | Object/null   | Error object on failure (null on success)                |

### Error Format

```json
{
  "id": 1,
  "result": null,
  "error": {
    "code": 1,
    "message": "invalid argument"
  }
}
```

| Code | Message             |
| ---- | ------------------- |
| `1`  | invalid argument    |
| `2`  | internal error      |
| `3`  | service unavailable |
| `4`  | method not found    |
| `5`  | service timeout     |
| `7`  | too many requests   |

**Note**: The server closes the connection immediately if invalid JSON is sent.

## WebSocket Endpoints

### Price Stream (Last Price)
- **Description**: Real-time last traded price updates for one or more markets. Update interval: 1 second.
- **Method**: `lastprice_subscribe`
- **Docs**: https://docs.whitebit.com/websocket/market-streams/lastprice

**Subscription Request:**
```json
{
  "id": 1,
  "method": "lastprice_subscribe",
  "params": ["BTC_USDT", "ETH_USDT"]
}
```

**Subscription Response:**
```json
{
  "id": 1,
  "result": { "status": "success" },
  "error": null
}
```

**Update Event:**
```json
{
  "id": null,
  "method": "lastprice_update",
  "params": ["BTC_USDT", "67500.50"]
}
```

**Response Fields (params array):**

| Index | Field       | Type   | Description        |
| ----- | ----------- | ------ | ------------------ |
| 0     | `market`    | String | Market name        |
| 1     | `last_price`| String | Last traded price  |

**One-time Query:**
Use `lastprice_request` with params `["BTC_USDT"]` for a single price fetch.

---

### 24h Ticker Stream (Market Statistics)
- **Description**: 24-hour market statistics including OHLCV data and price changes. Update interval: 1 second.
- **Method**: `market_subscribe`
- **Docs**: https://docs.whitebit.com/websocket/market-streams/market

**Subscription Request:**
```json
{
  "id": 2,
  "method": "market_subscribe",
  "params": ["BTC_USDT", "ETH_USDT"]
}
```

**Update Event:**
```json
{
  "id": null,
  "method": "market_update",
  "params": [
    "ETH_BTC",
    {
      "period": 86400,
      "last": "0.020964",
      "open": "0.020349",
      "close": "0.020964",
      "high": "0.020997",
      "low": "0.020281",
      "volume": "135574.476",
      "deal": "2784.413999488"
    }
  ]
}
```

**Response Fields (MarketStatistics object):**

| Field    | Type    | Description                              |
| -------- | ------- | ---------------------------------------- |
| `period` | Integer | Time window in seconds (fixed 86400)     |
| `last`   | String  | Current last trade price                 |
| `open`   | String  | Price at beginning of period             |
| `close`  | String  | Closing price (typically same as `last`) |
| `high`   | String  | Highest price in period                  |
| `low`    | String  | Lowest price in period                   |
| `volume` | String  | Volume in base (stock) currency          |
| `deal`   | String  | Volume in quote (money) currency         |

**Note**: Subscriptions are fixed at 24-hour rolling window. Custom periods are only supported via `market_request` query.

---

### Order Book Depth Stream
- **Description**: Real-time order book depth with bids and asks. Supports snapshot + incremental updates.
- **Method**: `depth_subscribe`
- **Docs**: https://docs.whitebit.com/websocket/market-streams/depth

**Subscription Request:**
```json
{
  "id": 3,
  "method": "depth_subscribe",
  "params": ["BTC_USDT", 100, "0", true]
}
```

**Params:**

| Index | Field            | Type    | Description                                            |
| ----- | ---------------- | ------- | ------------------------------------------------------ |
| 0     | `market`         | String  | Market name                                            |
| 1     | `limit`          | Integer | Depth limit: 1, 5, 10, 20, 30, 50, or 100             |
| 2     | `price_interval` | String  | Price grouping ("0" for no grouping)                   |
| 3     | `add`            | Boolean | `true` = add subscription, `false` = unsubscribe all   |

**Update Event (full snapshot - first message):**
```json
{
  "id": null,
  "method": "depth_update",
  "params": [
    true,
    {
      "timestamp": 1689600180.516447,
      "asks": [["0.020846", "29.369"]],
      "bids": [["0.02083", "9.598"]],
      "update_id": 214403,
      "event_time": 1749026542.817343
    },
    "ETH_BTC"
  ]
}
```

**Update Event (incremental):**
```json
{
  "id": null,
  "method": "depth_update",
  "params": [
    false,
    {
      "timestamp": 1689600180.616447,
      "past_update_id": 214403,
      "asks": [["0.020850", "15.200"]],
      "bids": [["0.020825", "0"]],
      "update_id": 214404,
      "event_time": 1749026542.917343
    },
    "ETH_BTC"
  ]
}
```

**Response Fields (DepthUpdateData object):**

| Field            | Type    | Description                                           |
| ---------------- | ------- | ----------------------------------------------------- |
| `timestamp`      | Number  | Timestamp from matchengine                            |
| `update_id`      | Integer | Current update sequence ID                            |
| `past_update_id` | Integer | Previous update ID (absent in snapshots)              |
| `asks`           | Array   | Ask levels `[price, amount]`, sorted ascending        |
| `bids`           | Array   | Bid levels `[price, amount]`, sorted descending       |
| `event_time`     | Number  | Event timestamp                                       |

**Snapshot vs Incremental Updates:**
- **First message** (no `past_update_id`): Full order book snapshot. Replace local book.
- **Subsequent messages** (with `past_update_id`): Incremental updates. Apply changes:
  - If amount is `"0"` → remove the price level
  - If amount is not `"0"` → update or insert the level
- **Keepalive**: If 10 seconds pass without changes, a full snapshot is pushed again.
- **RPI orders**: Public depth updates exclude Retail Price Improvement orders.

---

### Market Trades Stream
- **Description**: Real-time trade execution feed. Supports specific markets or all markets.
- **Method**: `trades_subscribe`
- **Docs**: https://docs.whitebit.com/websocket/market-streams/trades

**Subscription Request (specific markets):**
```json
{
  "id": 4,
  "method": "trades_subscribe",
  "params": ["BTC_USDT", "ETH_USDT"]
}
```

**Subscription Request (all markets):**
```json
{
  "id": 4,
  "method": "trades_subscribe",
  "params": []
}
```

**Update Event:**
```json
{
  "id": null,
  "method": "trades_update",
  "params": [
    "ETH_BTC",
    [
      {
        "id": 41358530,
        "time": 1580905394.70332,
        "price": "0.020857",
        "amount": "5.511",
        "type": "sell"
      }
    ]
  ]
}
```

**Trade Object Fields:**

| Field    | Type    | Description                                    |
| -------- | ------- | ---------------------------------------------- |
| `id`     | Integer | Trade ID                                       |
| `time`   | Number  | Trade time (Unix timestamp with milliseconds)  |
| `price`  | String  | Trade price                                    |
| `amount` | String  | Trade amount                                   |
| `type`   | String  | Trade side: `"buy"` or `"sell"`                |
| `rpi`    | Boolean | Optional. `true` if RPI order involved         |

**Note**: Each new subscription replaces the existing one. Use empty `params` for all markets.

---

### Candlestick/Kline Stream
- **Description**: OHLCV candlestick data with configurable intervals. Update interval: 0.5 seconds for subscriptions.
- **Method**: `candles_subscribe`
- **Docs**: https://docs.whitebit.com/websocket/market-streams/kline

**Available Intervals (seconds):**

| Range              | Valid Intervals                                    |
| ------------------ | -------------------------------------------------- |
| < 60s              | 1, 2, 3, 4, 5, 6, 10, 12, 15, 20, 30              |
| 60s - 1h           | 60, 120, 180, 300, 900, 1800 (1m, 2m, 3m, 5m, 15m, 30m) |
| 1h - 1d            | 3600, 7200, 14400, 21600, 43200 (1h, 2h, 4h, 6h, 12h) |
| 1d - 1w            | 86400, 172800, 259200 (1d, 2d, 3d)                |
| Special            | 604800 (1w), 2592000 (1m)                         |

**Subscription Request:**
```json
{
  "id": 5,
  "method": "candles_subscribe",
  "params": ["BTC_USDT", 300]
}
```

**Update Event:**
```json
{
  "id": null,
  "method": "candles_update",
  "params": [
    [
      1580895000,
      "0.020683",
      "0.020683",
      "0.020683",
      "0.020666",
      "504.701",
      "10.433600491",
      "ETH_BTC"
    ]
  ]
}
```

**Candlestick Tuple Fields:**

| Index | Field           | Type    | Description                        |
| ----- | --------------- | ------- | ---------------------------------- |
| 0     | `time`          | Integer | Open time (Unix timestamp)         |
| 1     | `open`          | String  | Open price                         |
| 2     | `close`         | String  | Close price                        |
| 3     | `high`          | String  | High price                         |
| 4     | `low`           | String  | Low price                          |
| 5     | `volume`        | String  | Volume in base (stock) currency    |
| 6     | `deal`          | String  | Volume in quote (money) currency   |
| 7     | `market`        | String  | Market pair name                   |

**One-time Query:**
Use `candles_request` with params `["ETH_BTC", start_time, end_time, interval]` for historical data.

---

### Book Ticker Stream
- **Description**: Real-time best bid and best ask price/amount updates. Instant updates on order book changes.
- **Method**: `bookTicker_subscribe`
- **Docs**: https://docs.whitebit.com/websocket/market-streams/book-ticker

**Subscription Request (specific market):**
```json
{
  "id": 6,
  "method": "bookTicker_subscribe",
  "params": ["BTC_USDT"]
}
```

**Subscription Request (all markets):**
```json
{
  "id": 6,
  "method": "bookTicker_subscribe",
  "params": []
}
```

**Update Event:**
```json
{
  "id": null,
  "method": "bookTicker_update",
  "params": [
    [
      1751958383.593387,
      1751958383.593557,
      "BTC_USDT",
      80670102,
      "67500.00",
      "1.500",
      "67501.00",
      "2.300"
    ]
  ]
}
```

**Book Ticker Tuple Fields:**

| Index | Field             | Type    | Description                                |
| ----- | ----------------- | ------- | ------------------------------------------ |
| 0     | `transaction_time`| Number  | Timestamp from matchengine                 |
| 1     | `message_time`    | Number  | Timestamp from WebSocket server            |
| 2     | `market`          | String  | Market name                                |
| 3     | `update_id`       | Integer | Update sequence ID                         |
| 4     | `best_bid_price`  | String  | Best bid price                             |
| 5     | `best_bid_amount` | String  | Best bid quantity                          |
| 6     | `best_ask_price`  | String  | Best ask price                             |
| 7     | `best_ask_amount` | String  | Best ask quantity                          |

**Note**: Public book ticker updates exclude RPI orders.

## Constraints & Limits

| Constraint                | Limit                              |
| ------------------------- | ---------------------------------- |
| New connections           | 1,000 per minute (global)          |
| JSON-RPC requests         | 200 per minute per connection      |
| Connection timeout        | 60 seconds of inactivity           |
| Ping interval             | Recommended every 50 seconds       |
| Depth subscription limits | 1, 5, 10, 20, 30, 50, or 100      |
| Depth keepalive           | Full snapshot every 10s if idle    |
| Last price update         | Every 1 second                     |
| Market stats update       | Every 1 second                     |
| Candlestick update        | Every 0.5 seconds                  |
| Order book update         | Every ~100ms                       |

## Notes

### Symbol Format Conventions
- Spot markets use `BASE_QUOTE` format with underscore separator
- Examples: `BTC_USDT`, `ETH_BTC`, `SOL_USD`
- Always uppercase

### Subscription Behavior
- Repeating a subscription for the same data type **replaces** the previous subscription
- Use a single WebSocket connection for all subscriptions
- Unsubscribe methods clear all subscriptions of that type (no per-market unsubscribe)

### RPI (Retail Price Improvement)
- RPI orders provide better execution prices for retail traders
- RPI orders are **not** visible in public `depth` or `bookTicker` streams
- RPI orders **are** visible in trade streams (via `rpi` field) and in the exchange UI
- Private WebSocket streams include RPI orders

### Reconnection Best Practices
1. Implement exponential backoff with jitter (1s, 2s, 4s, 8s)
2. Re-subscribe to all channels after reconnect
3. For depth streams, discard local order book and wait for new snapshot
4. Monitor ping latency; reconnect if no pong within 10 seconds
5. Queue subscription requests to avoid hitting rate limits during setup

### Regional Endpoints
- Global: `wss://api.whitebit.com/ws`
- EU: `wss://api.whitebit.eu/ws`
- Choose based on your geographic location and regulatory requirements
