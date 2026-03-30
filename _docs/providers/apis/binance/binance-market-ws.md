# Binance WebSocket Market Documentation (Spot)

## Reference
- **Official API Docs**: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams
- **WebSocket Base URL**: wss://stream.binance.com:9443/ws
- **Combined Stream URL**: wss://stream.binance.com:9443/stream?streams=

## Protocol Overview

| Aspect | Details |
|--------|---------|
| Protocol | Raw WebSocket (no authentication required for market data) |
| Base Endpoints | `wss://stream.binance.com:9443` or `wss://stream.binance.com:443` |
| Market Data Only | `wss://data-stream.binance.vision` (user data streams NOT available) |
| Connection Validity | Single connection valid for 24 hours; expect disconnection at 24h mark |
| Keep-Alive | Server sends `ping frame` every 20 seconds; client must respond with `pong` within 60 seconds or connection is dropped |
| Ping/Pong | Pong must include copy of ping's payload. Unsolicited pongs allowed but should have empty payload |
| Reconnection | Implement exponential backoff; reconnect on disconnect or 24h expiry |

## Stream Naming Convention

- All symbols for streams are **lowercase** (e.g., `btcusdt`, not `BTCUSDT`)
- Raw streams: `/ws/<streamName>`
- Combined streams: `/stream?streams=<streamName1>/<streamName2>/<streamName3>`
- Combined stream events are wrapped as: `{"stream":"<streamName>","data":<rawPayload>}`
- Time unit: All timestamps are **milliseconds** by default. Add `timeUnit=MICROSECOND` parameter for microsecond precision

## Live Subscription Management

Streams can be subscribed/unsubscribed dynamically after connection:

### Subscribe
```json
{
  "method": "SUBSCRIBE",
  "params": ["btcusdt@aggTrade", "btcusdt@depth"],
  "id": 1
}
```

### Unsubscribe
```json
{
  "method": "UNSUBSCRIBE",
  "params": ["btcusdt@depth"],
  "id": 312
}
```

### List Subscriptions
```json
{
  "method": "LIST_SUBSCRIPTIONS",
  "id": 3
}
```

Response `result: null` indicates success for non-query requests.

## WebSocket Endpoints

### Aggregate Trade Stream
- **Description**: Pushes trade information aggregated for a single taker order
- **Stream Name**: `<symbol>@aggTrade`
- **Update Speed**: Real-time
- **Official Docs**: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#aggregate-trade-streams

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| e | string | Event type ("aggTrade") |
| E | int64 | Event time (ms) |
| s | string | Symbol |
| a | int64 | Aggregate trade ID |
| p | string | Price |
| q | string | Quantity |
| f | int64 | First trade ID |
| l | int64 | Last trade ID |
| T | int64 | Trade time (ms) |
| m | bool | Is the buyer the market maker? |
| M | bool | Ignore |

**Sample Response:**
```json
{
  "e": "aggTrade",
  "E": 1672515782136,
  "s": "BNBBTC",
  "a": 12345,
  "p": "0.001",
  "q": "100",
  "f": 100,
  "l": 105,
  "T": 1672515782136,
  "m": true,
  "M": true
}
```

### Price Stream (Trade)
- **Description**: Pushes raw trade information; each trade has a unique buyer and seller
- **Stream Name**: `<symbol>@trade`
- **Update Speed**: Real-time
- **Official Docs**: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#trade-streams

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| e | string | Event type ("trade") |
| E | int64 | Event time (ms) |
| s | string | Symbol |
| t | int64 | Trade ID |
| p | string | Price |
| q | string | Quantity |
| T | int64 | Trade time (ms) |
| m | bool | Is the buyer the market maker? |
| M | bool | Ignore |

**Sample Response:**
```json
{
  "e": "trade",
  "E": 1672515782136,
  "s": "BNBBTC",
  "t": 12345,
  "p": "0.001",
  "q": "100",
  "T": 1672515782136,
  "m": true,
  "M": true
}
```

### 24h Ticker Stream
- **Description**: 24hr rolling window ticker statistics for a single symbol (NOT UTC day statistics)
- **Stream Name**: `<symbol>@ticker`
- **Update Speed**: 1000ms
- **Official Docs**: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#individual-symbol-ticker-streams

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| e | string | Event type ("24hrTicker") |
| E | int64 | Event time (ms) |
| s | string | Symbol |
| p | string | Price change |
| P | string | Price change percent |
| w | string | Weighted average price |
| x | string | First trade(F)-1 price |
| c | string | Last price |
| Q | string | Last quantity |
| b | string | Best bid price |
| B | string | Best bid quantity |
| a | string | Best ask price |
| A | string | Best ask quantity |
| o | string | Open price |
| h | string | High price |
| l | string | Low price |
| v | string | Total traded base asset volume |
| q | string | Total traded quote asset volume |
| O | int64 | Statistics open time (ms) |
| C | int64 | Statistics close time (ms) |
| F | int64 | First trade ID |
| L | int64 | Last trade ID |
| n | int64 | Total number of trades |

**Sample Response:**
```json
{
  "e": "24hrTicker",
  "E": 1672515782136,
  "s": "BNBBTC",
  "p": "0.0015",
  "P": "250.00",
  "w": "0.0018",
  "x": "0.0009",
  "c": "0.0025",
  "Q": "10",
  "b": "0.0024",
  "B": "10",
  "a": "0.0026",
  "A": "100",
  "o": "0.0010",
  "h": "0.0025",
  "l": "0.0010",
  "v": "10000",
  "q": "18",
  "O": 0,
  "C": 86400000,
  "F": 0,
  "L": 18150,
  "n": 18151
}
```

### Individual Symbol Mini Ticker Stream
- **Description**: 24hr rolling window mini-ticker statistics (NOT UTC day statistics)
- **Stream Name**: `<symbol>@miniTicker`
- **Update Speed**: 1000ms
- **Official Docs**: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#individual-symbol-mini-ticker-stream

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| e | string | Event type ("24hrMiniTicker") |
| E | int64 | Event time (ms) |
| s | string | Symbol |
| c | string | Close price |
| o | string | Open price |
| h | string | High price |
| l | string | Low price |
| v | string | Total traded base asset volume |
| q | string | Total traded quote asset volume |

**Sample Response:**
```json
{
  "e": "24hrMiniTicker",
  "E": 1672515782136,
  "s": "BNBBTC",
  "c": "0.0025",
  "o": "0.0010",
  "h": "0.0025",
  "l": "0.0010",
  "v": "10000",
  "q": "18"
}
```

### Individual Symbol Book Ticker Stream
- **Description**: Pushes any update to the best bid or ask's price or quantity in real-time
- **Stream Name**: `<symbol>@bookTicker`
- **Update Speed**: Real-time
- **Official Docs**: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#individual-symbol-book-ticker-streams

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| u | int64 | Order book update ID |
| s | string | Symbol |
| b | string | Best bid price |
| B | string | Best bid quantity |
| a | string | Best ask price |
| A | string | Best ask quantity |

**Sample Response:**
```json
{
  "u": 400900217,
  "s": "BNBUSDT",
  "b": "25.35190000",
  "B": "31.21000000",
  "a": "25.36520000",
  "A": "40.66000000"
}
```

### All Market Mini Tickers Stream
- **Description**: 24hr rolling window mini-ticker statistics for all symbols that changed (array)
- **Stream Name**: `!miniTicker@arr`
- **Update Speed**: 1000ms
- **Official Docs**: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#all-market-mini-tickers-stream
- **Note**: Only tickers that have changed will be present in the array

### Partial Book Depth Stream
- **Description**: Top bids and asks, pushed at configurable intervals
- **Stream Name**: `<symbol>@depth<levels>` or `<symbol>@depth<levels>@100ms`
- **Levels Available**: 5, 10, 20
- **Update Speed**: 1000ms or 100ms (with `@100ms` suffix)
- **Official Docs**: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#partial-book-depth-streams

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| lastUpdateId | int64 | Last update ID |
| bids | array | Bids to be updated [price, quantity] |
| asks | array | Asks to be updated [price, quantity] |

**Sample Response:**
```json
{
  "lastUpdateId": 160,
  "bids": [
    ["0.0024", "10"]
  ],
  "asks": [
    ["0.0026", "100"]
  ]
}
```

### Diff. Depth Stream
- **Description**: Order book price and quantity depth updates used to locally manage an order book
- **Stream Name**: `<symbol>@depth` or `<symbol>@depth@100ms`
- **Update Speed**: 1000ms or 100ms (with `@100ms` suffix)
- **Official Docs**: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#diff-depth-stream

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| e | string | Event type ("depthUpdate") |
| E | int64 | Event time (ms) |
| s | string | Symbol |
| U | int64 | First update ID in event |
| u | int64 | Final update ID in event |
| b | array | Bids to be updated [price, quantity] |
| a | array | Asks to be updated [price, quantity] |

**Sample Response:**
```json
{
  "e": "depthUpdate",
  "E": 1672515782136,
  "s": "BNBBTC",
  "U": 157,
  "u": 160,
  "b": [
    ["0.0024", "10"]
  ],
  "a": [
    ["0.0026", "100"]
  ]
}
```

### Candlestick/Kline Stream
- **Description**: Pushes updates to the current klines/candlestick every second in UTC+0 timezone
- **Stream Name**: `<symbol>@kline_<interval>`
- **Update Speed**: 1000ms for `1s`, 2000ms for other intervals
- **Official Docs**: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#klinecandlestick-streams-for-utc

**Available Intervals:**

| Interval | Description |
|----------|-------------|
| 1s | 1 second |
| 1m | 1 minute |
| 3m | 3 minutes |
| 5m | 5 minutes |
| 15m | 15 minutes |
| 30m | 30 minutes |
| 1h | 1 hour |
| 2h | 2 hours |
| 4h | 4 hours |
| 6h | 6 hours |
| 8h | 8 hours |
| 12h | 12 hours |
| 1d | 1 day |
| 3d | 3 days |
| 1w | 1 week |
| 1M | 1 month |

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| e | string | Event type ("kline") |
| E | int64 | Event time (ms) |
| s | string | Symbol |
| k.t | int64 | Kline start time (ms) |
| k.T | int64 | Kline close time (ms) |
| k.s | string | Symbol |
| k.i | string | Interval |
| k.f | int64 | First trade ID |
| k.L | int64 | Last trade ID |
| k.o | string | Open price |
| k.c | string | Close price |
| k.h | string | High price |
| k.l | string | Low price |
| k.v | string | Base asset volume |
| k.n | int64 | Number of trades |
| k.x | bool | Is this kline closed? |
| k.q | string | Quote asset volume |
| k.V | string | Taker buy base asset volume |
| k.Q | string | Taker buy quote asset volume |
| k.B | string | Ignore |

**Sample Response:**
```json
{
  "e": "kline",
  "E": 1672515782136,
  "s": "BNBBTC",
  "k": {
    "t": 1672515780000,
    "T": 1672515839999,
    "s": "BNBBTC",
    "i": "1m",
    "f": 100,
    "L": 200,
    "o": "0.0010",
    "c": "0.0020",
    "h": "0.0025",
    "l": "0.0015",
    "v": "1000",
    "n": 100,
    "x": false,
    "q": "1.0000",
    "V": "500",
    "Q": "0.500",
    "B": "123456"
  }
}
```

### Rolling Window Statistics Stream
- **Description**: Rolling window ticker statistics for a single symbol, computed over multiple windows
- **Stream Name**: `<symbol>@ticker_<window_size>`
- **Window Sizes**: 1h, 4h, 1d
- **Update Speed**: 1000ms
- **Official Docs**: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#individual-symbol-rolling-window-statistics-streams
- **Note**: Open time `"O"` always starts on a minute; closing time `"C"` is current time. Effective window may be up to 59999ms wider than `<window_size>`.

### Reference Price Stream
- **Description**: Reference price updates
- **Stream Name**: `<symbol>@referencePrice`
- **Update Speed**: 1000ms
- **Official Docs**: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#reference-price-streams

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| e | string | Event type ("referencePrice") |
| s | string | Symbol |
| r | string | Reference price (null if no reference price) |
| t | int64 | Engine timestamp when reference price was valid |

### Average Price Stream
- **Description**: Average price changes over a fixed time interval
- **Stream Name**: `<symbol>@avgPrice`
- **Update Speed**: 1000ms
- **Official Docs**: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#average-price

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| e | string | Event type ("avgPrice") |
| E | int64 | Event time (ms) |
| s | string | Symbol |
| i | string | Average price interval |
| w | string | Average price |
| T | int64 | Last trade time (ms) |

## How to Manage a Local Order Book Correctly

1. Open a WebSocket connection to `wss://stream.binance.com:9443/ws/<symbol>@depth`
2. Buffer events received from the stream. Note the `U` of the first event received
3. Get a depth snapshot from `https://api.binance.com/api/v3/depth?symbol=<SYMBOL>&limit=5000`
4. If the `lastUpdateId` from the snapshot is strictly less than `U` from step 2, go back to step 3
5. In buffered events, discard any event where `u` <= `lastUpdateId` of the snapshot
6. Set local order book to the snapshot. Its update ID is `lastUpdateId`
7. Apply the update procedure to all buffered events, then to all subsequent events

**Update Procedure:**
1. If event `u` < local order book update ID, ignore the event
2. If event `U` > local order book update ID + 1, you have missed events - discard and restart
3. For each price level in bids/asks:
   - If price level doesn't exist, insert with new quantity
   - If quantity is zero, remove the price level
4. Set order book update ID to `u` in the processed event

## Constraints & Limits

| Constraint | Limit |
|------------|-------|
| Incoming messages | 5 per second per connection (PING, PONG, JSON control messages) |
| Max streams per connection | 1024 |
| Connection attempts | 300 per 5 minutes per IP |
| Connection validity | 24 hours (expect disconnection) |
| Disconnect penalty | IPs repeatedly disconnected may be banned |

## Notes

- **Symbol Format**: All stream symbols must be **lowercase**. `BTCUSDT` must be specified as `btcusdt`
- **Combined Streams**: When using `/stream?streams=`, each event is wrapped as `{"stream":"<streamName>","data":<rawPayload>}`
- **Time Precision**: Timestamps are milliseconds by default. Add `timeUnit=MICROSECOND` to URL for microsecond precision
- **UTF-8 Support**: Symbols with non-ASCII characters are supported (UTF-8 encoded)
- **Depth Stream Snapshots**: API depth snapshots have a limit of 5000 price levels per side. Levels outside the initial snapshot won't be visible until they change
- **Rolling Window vs UTC Day**: Ticker statistics are 24hr rolling windows, NOT UTC day statistics
- **Deprecated Stream**: `!ticker@arr` (All Market Tickers) is deprecated; use `<symbol>@ticker` or `!miniTicker@arr` instead
- **Market Data Only Endpoint**: `wss://data-stream.binance.vision` can be used for market data only (no user data streams)
