# Bitget WebSocket Market Documentation (Spot)

## Reference
- **Official API Docs**: https://www.bitget.com/api-doc/spot/websocket/public/Tickers-Channel
- **WebSocket Base URL (V2)**: `wss://ws.bitget.com/v2/ws/public` (recommended)
- **WebSocket Base URL (V1)**: `wss://ws.bitget.com/spot/v1/stream` (legacy)
- **Channel Docs**:
  - [Market (Ticker) Channel](https://www.bitget.com/api-doc/spot/websocket/public/Tickers-Channel)
  - [Candlestick Channel](https://www.bitget.com/api-doc/spot/websocket/public/Candlesticks-Channel)
  - [Trading Channel](https://www.bitget.com/api-doc/spot/websocket/public/Trades-Channel)
  - [Depth Channel](https://www.bitget.com/api-doc/spot/websocket/public/Depth-Channel)
  - [Auction Channel](https://www.bitget.com/api-doc/spot/websocket/public/Auction-Channel)

## Protocol Overview

- **Protocol**: Raw WebSocket (full-duplex)
- **Connection limits**: 300 connection requests/IP/5min, max 100 concurrent connections/IP
- **Subscription limits**: 240 subscription requests/hour/connection, max 1000 channels/connection
- **Recommended**: Subscribe to fewer than 50 channels per connection for stability
- **Keep-alive**: Send string `"ping"` every 30 seconds; expect `"pong"` response. Server disconnects if no ping received for 2 minutes
- **Message rate limit**: Max 10 messages/second (includes ping and JSON messages). Exceeding this disconnects the connection
- **Authentication**: Not required for public market streams
- **Compression**: Not documented; payloads are plain JSON

## Message Format

### Subscription Request

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "SPOT",
      "channel": "ticker",
      "instId": "BTCUSDT"
    }
  ]
}
```

### Unsubscription Request

```json
{
  "op": "unsubscribe",
  "args": [
    {
      "instType": "SPOT",
      "channel": "ticker",
      "instId": "BTCUSDT"
    }
  ]
}
```

### Subscription Confirmation

```json
{
  "event": "subscribe",
  "arg": {
    "instType": "SPOT",
    "channel": "ticker",
    "instId": "BTCUSDT"
  }
}
```

### Error Response

```json
{
  "event": "error",
  "arg": {
    "instType": "SPOT",
    "channel": "ticker",
    "instId": "INVALID"
  },
  "code": "30003",
  "msg": "instType:SP,channel:ticker,instId:INVALID Symbol not exists",
  "op": "subscribe"
}
```

### Push Data Envelope

All market data pushes share this structure:

```json
{
  "action": "snapshot",
  "arg": {
    "instType": "SPOT",
    "channel": "ticker",
    "instId": "BTCUSDT"
  },
  "data": [...],
  "ts": 1695702438029
}
```

| Field | Type | Description |
|-------|------|-------------|
| `action` | string | `"snapshot"` or `"update"` |
| `arg` | object | Subscription argument identifying the channel |
| `arg.instType` | string | Product type, always `"SPOT"` for spot markets |
| `arg.channel` | string | Channel name |
| `arg.instId` | string | Instrument ID (e.g., `BTCUSDT`) |
| `data` | array | Channel-specific payload |
| `ts` | number | Server push timestamp (ms) |

## WebSocket Endpoints

### 24h Ticker Stream (Market Channel)

- **Description**: Real-time 24h ticker data including last price, 24h high/low, volume, and best bid/ask. Push frequency: 200ms–300ms.
- **Channel**: `ticker`
- **Subscription Request**:

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "SPOT",
      "channel": "ticker",
      "instId": "ETHUSDT"
    }
  ]
}
```

- **Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `instId` | string | Product ID (e.g., `ETHUSDT`) |
| `lastPr` | string | Latest traded price |
| `bidPr` | string | Best bid price |
| `askPr` | string | Best ask price |
| `bidSz` | string | Best bid size |
| `askSz` | string | Best ask size |
| `open24h` | string | Price 24h ago |
| `high24h` | string | 24h high price |
| `low24h` | string | 24h low price |
| `change24h` | string | 24h change ratio (0.01 = 1%) |
| `baseVolume` | string | 24h volume in base currency |
| `quoteVolume` | string | 24h volume in quote currency |
| `openUtc` | string | UTC±00:00 open price |
| `changeUtc24h` | string | UTC 24h change ratio (0.01 = 1%) |
| `ts` | string | Data generation timestamp (ms) |

- **Sample Response**:

```json
{
  "action": "snapshot",
  "arg": {
    "instType": "SPOT",
    "channel": "ticker",
    "instId": "ETHUSDT"
  },
  "data": [
    {
      "instId": "ETHUSDT",
      "lastPr": "2200.10",
      "open24h": "2150.00",
      "high24h": "2250.00",
      "low24h": "2100.00",
      "change24h": "0.0233",
      "bidPr": "2200.00",
      "askPr": "2200.10",
      "bidSz": "0.0084",
      "askSz": "19740.8811",
      "baseVolume": "12345.67",
      "quoteVolume": "27160485.12",
      "openUtc": "2180.00",
      "changeUtc24h": "0.0092",
      "ts": "1695702438018"
    }
  ],
  "ts": 1695702438029
}
```

### Price Stream (Trade Channel)

- **Description**: Real-time trade data pushed whenever a trade is matched (taker orders). After initial subscription, recent snapshot data is pushed first, then live updates.
- **Channel**: `trade`
- **Subscription Request**:

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "SPOT",
      "channel": "trade",
      "instId": "BTCUSDT"
    }
  ]
}
```

- **Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `tradeId` | string | Unique trade ID |
| `price` | string | Trade execution price |
| `size` | string | Trade quantity (base currency) |
| `side` | string | Trade direction: `"buy"` or `"sell"` |
| `ts` | string | Trade timestamp (ms) |

- **Sample Response**:

```json
{
  "action": "snapshot",
  "arg": {
    "instType": "SPOT",
    "channel": "trade",
    "instId": "BTCUSDT"
  },
  "data": [
    {
      "ts": "1695709835822",
      "price": "26293.4",
      "size": "0.0013",
      "side": "buy",
      "tradeId": "1000000000"
    }
  ],
  "ts": 1695711090682
}
```

### Order Book Depth Stream

- **Description**: Order book data with multiple depth options. The `books` channel provides full depth with incremental updates; other channels push snapshots at fixed depths.
- **Channels**:

| Channel | Depth | Push Frequency | Mode |
|---------|-------|----------------|------|
| `books` | All levels | 200ms | Snapshot + incremental updates |
| `books1` | 1 level | 10ms | Snapshot only |
| `books5` | 5 levels | 200ms | Snapshot only |
| `books15` | 15 levels | 200ms | Snapshot only |

- **Subscription Request**:

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "SPOT",
      "channel": "books5",
      "instId": "BTCUSDT"
    }
  ]
}
```

- **Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `asks` | array | Ask entries, each `[price, size]` |
| `bids` | array | Bid entries, each `[price, size]` |
| `ts` | string | Matching engine timestamp (ms) |
| `checksum` | number | CRC32 checksum (for `books` channel) |
| `seq` | number | Sequence number, increments with each update |

- **Sample Response** (`books5`):

```json
{
  "action": "snapshot",
  "arg": {
    "instType": "SPOT",
    "channel": "books5",
    "instId": "BTCUSDT"
  },
  "data": [
    {
      "asks": [
        ["26274.9", "0.0009"],
        ["26275.0", "0.0500"]
      ],
      "bids": [
        ["26274.8", "0.0009"],
        ["26274.7", "0.0027"]
      ],
      "checksum": 0,
      "seq": 123,
      "ts": "1695710946294"
    }
  ],
  "ts": 1695710946294
}
```

- **Snapshot vs Incremental Updates**:
  - `books` channel: First push is a full snapshot (`action: "snapshot"`), subsequent pushes are incremental (`action: "update"`). Clients must merge updates into their local snapshot.
  - `books1`, `books5`, `books15`: Always push full snapshots. No merging required.
  - **Merging rules**: If an update has a price level with size `0`, remove that level. If size changes, replace the level. If a new price appears, insert it in sorted order (bids descending, asks ascending).
  - **Checksum**: The `books` channel includes a CRC32 checksum computed from the first 25 bid/ask levels. Use this to verify local snapshot integrity.

### Candlestick/Kline Stream

- **Description**: OHLCV candlestick data. Pushes every 1 second when trades occur, or once per granularity interval when idle.
- **Channels**:

| Channel | Interval | Description |
|---------|----------|-------------|
| `candle1m` | 1 minute | 1-minute candles |
| `candle5m` | 5 minutes | 5-minute candles |
| `candle15m` | 15 minutes | 15-minute candles |
| `candle30m` | 30 minutes | 30-minute candles |
| `candle1H` | 1 hour | 1-hour candles |
| `candle4H` | 4 hours | 4-hour candles |
| `candle6H` | 6 hours | 6-hour candles |
| `candle12H` | 12 hours | 12-hour candles |
| `candle1D` | 1 day | Daily candles |
| `candle3D` | 3 days | 3-day candles |
| `candle1W` | 1 week | Weekly candles |
| `candle1M` | 1 month | Monthly candles |
| `candle6Hutc` | 6 hours UTC | 6-hour candles aligned to UTC |
| `candle12Hutc` | 12 hours UTC | 12-hour candles aligned to UTC |
| `candle1Dutc` | 1 day UTC | Daily candles aligned to UTC |
| `candle3Dutc` | 3 days UTC | 3-day candles aligned to UTC |
| `candle1Wutc` | 1 week UTC | Weekly candles aligned to UTC |
| `candle1Mutc` | 1 month UTC | Monthly candles aligned to UTC |

- **Subscription Request**:

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "SPOT",
      "channel": "candle1m",
      "instId": "ETHUSDT"
    }
  ]
}
```

- **Response Fields** (array format):

| Index | Type | Description |
|-------|------|-------------|
| `[0]` | string | Candle start time (ms timestamp) |
| `[1]` | string | Open price |
| `[2]` | string | High price |
| `[3]` | string | Low price |
| `[4]` | string | Close price |
| `[5]` | string | Base currency volume |
| `[6]` | string | Quote currency volume |
| `[7]` | string | Quote currency volume (USDT) |

- **Sample Response**:

```json
{
  "action": "snapshot",
  "arg": {
    "instType": "SPOT",
    "channel": "candle1m",
    "instId": "ETHUSDT"
  },
  "data": [
    [
      "1695672780000",
      "2200.1",
      "2200.1",
      "2200.1",
      "2200.1",
      "0",
      "0",
      "0"
    ]
  ],
  "ts": 1695702747821
}
```

### Best Bid/Ask Stream (BBO)

- **Description**: Best bid/ask data is available through the depth channels. Use `books1` for the fastest updates (10ms push frequency) at the top-of-book level.
- **Channel**: `books1`
- **Subscription Request**:

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "SPOT",
      "channel": "books1",
      "instId": "BTCUSDT"
    }
  ]
}
```

- **Response Fields**: Same as [Order Book Depth Stream](#order-book-depth-stream) but always contains exactly 1 bid and 1 ask level.
- **Sample Response**:

```json
{
  "action": "snapshot",
  "arg": {
    "instType": "SPOT",
    "channel": "books1",
    "instId": "BTCUSDT"
  },
  "data": [
    {
      "asks": [
        ["26275.0", "0.0500"]
      ],
      "bids": [
        ["26274.8", "0.0009"]
      ],
      "checksum": 0,
      "seq": 456,
      "ts": "1695710946304"
    }
  ],
  "ts": 1695710946304
}
```

### Auction Channel

- **Description**: Call auction information for spot pairs. Push frequency: 100ms–300ms. Only active during auction periods.
- **Channel**: `auction`
- **Subscription Request**:

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "SPOT",
      "channel": "auction",
      "instId": "BTCUSDT"
    }
  ]
}
```

- **Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `stage` | string | Auction stage: `pre_market`, `stage_1`, `stage_2`, `stage_3`, `success`, `failure` |
| `stageEndTime` | string | Current stage end time (ms) |
| `estOpeningPrice` | string | Estimated opening price (null if not available) |
| `matchedVolume` | string | Matched volume in base currency (null if not available) |
| `auctionEndTime` | string | Call auction end time (ms) |

- **Sample Response**:

```json
{
  "action": "snapshot",
  "arg": {
    "instType": "SPOT",
    "channel": "auction",
    "instId": "BTCUSDT"
  },
  "data": [
    {
      "stage": "stage_1",
      "stageEndTime": "1728876900000",
      "estOpeningPrice": null,
      "matchedVolume": null,
      "auctionEndTime": "1728877500000"
    }
  ],
  "ts": 1695711090682
}
```

## Constraints & Limits

| Constraint | Value |
|------------|-------|
| Max connections per IP | 100 |
| Connection requests per IP per 5 min | 300 |
| Subscription requests per hour per connection | 240 |
| Max channels per connection | 1000 |
| Recommended channels per connection | < 50 |
| Max messages per second | 10 |
| Ping timeout | 2 minutes (server disconnects) |
| Subscribe payload max length | 4096 bytes per request |
| Reconnect backoff | Exponential backoff recommended |

## Notes

- **Symbol format**: Spot instrument IDs use the format `BASEQUOTE` (e.g., `BTCUSDT`, `ETHUSDT`). No separator between base and quote currencies.
- **V1 vs V2 API**: The V2 endpoint (`wss://ws.bitget.com/v2/ws/public`) uses `instType: "SPOT"` in subscription args. The legacy V1 endpoint (`wss://ws.bitget.com/spot/v1/stream`) may use different channel naming conventions.
- **All values are strings**: Numeric fields in push data are returned as strings. Parse them client-side as needed.
- **Timestamps**: All timestamps are Unix epoch in milliseconds.
- **Seq field**: The `seq` field in depth channel updates is monotonically increasing (except during symbol maintenance). Use it to detect out-of-order packets.
- **Checksum verification**: For the `books` channel, compute CRC32 from the first 25 bid/ask levels interleaved as `bid1[price:size]:ask1[price:size]:bid2[price:size]:ask2[price:size]...`. Use the original string representation (e.g., `0.5000` not `0.5`).
- **Error code 30003**: Symbol does not exist. Verify the instrument ID format.
- **Channel subscription batching**: Multiple channels can be subscribed to in a single request by adding multiple objects to the `args` array. Total payload must not exceed 4096 bytes.
