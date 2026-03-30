# MEXC WebSocket Market Documentation (Spot)

## Reference
- **Official API Docs**: https://www.mexc.com/api-docs/spot-v3/websocket-market-streams
- **WebSocket Base URL**: `wss://wbs-api.mexc.com/ws`

## Protocol Overview
- **Protocol type**: Raw WebSocket (JSON messages)
- **Connection validity**: Each connection is valid for no more than 24 hours
- **Keep-alive**: Client can send `{"method": "PING"}` to keep the connection alive; server responds with `{"id": 0, "code": 0, "msg": "PONG"}`
- **Idle disconnection**:
  - No valid subscription: server disconnects after **30 seconds**
  - Subscription active but no data flow: server disconnects after **1 minute**
- **Reconnection guidelines**: Handle disconnections and reconnections properly; each connection expires after 24 hours
- **Authentication**: Not required for public market streams

## Message Format

### Request Format
All subscription/unsubscription requests use JSON with the following structure:

```json
{
  "method": "SUBSCRIPTION",
  "params": ["<topic>"]
}
```

### Response Format
Subscription/unsubscription responses:

```json
{
  "id": 0,
  "code": 0,
  "msg": "<topic>"
}
```

- `id`: Unsigned integer serving as unique identifier for communication
- `code`: `0` indicates success
- `msg`: Echoes the requested topic, confirming successful subscription

### Error Format
Non-zero `code` values indicate errors. Refer to the [Public API Definitions](https://www.mexc.com/api-docs/spot-v3/public-api-definitions) for error codes.

## WebSocket Endpoints

> **Important**: All MEXC Spot WebSocket streams use **Protocol Buffers (protobuf)** format. The `.proto` definition files are available at: https://github.com/mexcdevelop/websocket-proto
>
> Topic naming convention: `spot@public.<type>.v3.api.pb@<symbol>` or `spot@public.<type>.v3.api.pb@<interval>@<symbol>`
>
> All trading pair names (symbols) must be in **UPPERCASE** (e.g., `BTCUSDT`).

### Price Stream (Trade)
- **Description**: Pushes raw trade information; each trade has a unique buyer and seller
- **Method/Topic**: `spot@public.aggre.deals.v3.api.pb@(100ms|10ms)@<symbol>`
- **Push frequency**: 10ms or 100ms aggregation

#### Subscription Request Format

```json
{
  "method": "SUBSCRIPTION",
  "params": ["spot@public.aggre.deals.v3.api.pb@100ms@BTCUSDT"]
}
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `channel` | string | Channel identifier |
| `symbol` | string | Trading pair (e.g., BTCUSDT) |
| `sendtime` | long | Event time (Unix ms) |
| `publicdeals.dealsList` | array | List of trades |
| `publicdeals.dealsList[].price` | string | Trade price |
| `publicdeals.dealsList[].quantity` | string | Trade quantity |
| `publicdeals.dealsList[].tradetype` | int | Trade type: `1` = Buy, `2` = Sell |
| `publicdeals.dealsList[].time` | long | Trade time (Unix ms) |
| `publicdeals.eventtype` | string | Event type identifier |

#### Sample Response

```json
{
  "channel": "spot@public.aggre.deals.v3.api.pb@100ms@BTCUSDT",
  "publicdeals": {
    "dealsList": [
      {
        "price": "93220.00",
        "quantity": "0.04438243",
        "tradetype": 2,
        "time": 1736409765051
      }
    ],
    "eventtype": "spot@public.aggre.deals.v3.api.pb@100ms"
  },
  "symbol": "BTCUSDT",
  "sendtime": 1736409765052
}
```

### 24h Ticker Stream (MiniTicker)
- **Description**: Pushes 24h mini ticker statistics for a specified trading pair in a specified timezone
- **Method/Topic**: `spot@public.miniTicker.v3.api.pb@<symbol>@<UTC-TIMEZONE>`
- **Push frequency**: Every 3 seconds

#### Available Timezones
`24H`, `UTC-10`, `UTC-8`, `UTC-7`, `UTC-6`, `UTC-5`, `UTC-4`, `UTC-3`, `UTC+0`, `UTC+1`, `UTC+2`, `UTC+3`, `UTC+4`, `UTC+4:30`, `UTC+5`, `UTC+5:30`, `UTC+6`, `UTC+7`, `UTC+8`, `UTC+9`, `UTC+10`, `UTC+11`, `UTC+12`, `UTC+12:45`, `UTC+13`

#### Subscription Request Format

```json
{
  "method": "SUBSCRIPTION",
  "params": ["spot@public.miniTicker.v3.api.pb@BTCUSDT@UTC+8"]
}
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `channel` | string | Channel identifier |
| `symbol` | string | Trading pair |
| `sendTime` | long | Event time (Unix ms) |
| `publicMiniTicker.symbol` | string | Trading pair name |
| `publicMiniTicker.price` | string | Latest price |
| `publicMiniTicker.rate` | string | Price change % (UTC+8 timezone) |
| `publicMiniTicker.zonedRate` | string | Price change % (local timezone) |
| `publicMiniTicker.high` | string | Rolling highest price |
| `publicMiniTicker.low` | string | Rolling lowest price |
| `publicMiniTicker.volume` | string | Rolling turnover amount (quote) |
| `publicMiniTicker.quantity` | string | Rolling trading volume (base) |
| `publicMiniTicker.lastCloseRate` | string | Previous close change % (UTC+8) |
| `publicMiniTicker.lastCloseZonedRate` | string | Previous close change % (local) |
| `publicMiniTicker.lastCloseHigh` | string | Previous close rolling high |
| `publicMiniTicker.lastCloseLow` | string | Previous close rolling low |

#### Sample Response

```json
{
  "channel": "spot@public.miniTicker.v3.api.pb@MXUSDT@UTC+8",
  "symbol": "MXUSDT",
  "sendTime": "1755076752201",
  "publicMiniTicker": {
    "symbol": "MXUSDT",
    "price": "2.5174",
    "rate": "0.0766",
    "zonedRate": "0.0766",
    "high": "2.6299",
    "low": "2.302",
    "volume": "11336518.0264",
    "quantity": "4638390.17",
    "lastCloseRate": "0.0767",
    "lastCloseZonedRate": "0.0767",
    "lastCloseHigh": "2.6299",
    "lastCloseLow": "2.302"
  }
}
```

### All Symbols MiniTickers
- **Description**: Pushes 24h mini ticker statistics for **all** trading pairs in the specified timezone
- **Method/Topic**: `spot@public.miniTickers.v3.api.pb@<UTC-TIMEZONE>`
- **Push frequency**: Every 3 seconds

#### Subscription Request Format

```json
{
  "method": "SUBSCRIPTION",
  "params": ["spot@public.miniTickers.v3.api.pb@UTC+8"]
}
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `channel` | string | Channel identifier |
| `sendTime` | long | Event time (Unix ms) |
| `publicMiniTickers.items[]` | array | List of ticker items |
| `publicMiniTickers.items[].symbol` | string | Trading pair name |
| `publicMiniTickers.items[].price` | string | Latest price |
| `publicMiniTickers.items[].rate` | string | Price change % (UTC+8) |
| `publicMiniTickers.items[].zonedRate` | string | Price change % (local timezone) |
| `publicMiniTickers.items[].high` | string | Rolling highest price |
| `publicMiniTickers.items[].low` | string | Rolling lowest price |
| `publicMiniTickers.items[].volume` | string | Rolling turnover amount |
| `publicMiniTickers.items[].quantity` | string | Rolling trading volume |
| `publicMiniTickers.items[].lastCloseRate` | string | Previous close change % (UTC+8) |
| `publicMiniTickers.items[].lastCloseZonedRate` | string | Previous close change % (local) |
| `publicMiniTickers.items[].lastCloseHigh` | string | Previous close rolling high |
| `publicMiniTickers.items[].lastCloseLow` | string | Previous close rolling low |

### Order Book Depth Stream

#### Diff. Depth Stream (Incremental)
- **Description**: Pushes incremental order book depth updates. If `quantity` is `0`, the price level should be removed.
- **Method/Topic**: `spot@public.aggre.depth.v3.api.pb@(100ms|10ms)@<symbol>`
- **Push frequency**: 10ms or 100ms aggregation

##### Subscription Request Format

```json
{
  "method": "SUBSCRIPTION",
  "params": ["spot@public.aggre.depth.v3.api.pb@100ms@BTCUSDT"]
}
```

##### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `channel` | string | Channel identifier |
| `symbol` | string | Trading pair |
| `sendtime` | long | Event time (Unix ms) |
| `publicincreasedepths.asksList[]` | array | Ask (sell) order updates |
| `publicincreasedepths.bidsList[]` | array | Bid (buy) order updates |
| `publicincreasedepths.[].price` | string | Price level of change |
| `publicincreasedepths.[].quantity` | string | Quantity (`0` = remove level) |
| `publicincreasedepths.eventtype` | string | Event type identifier |
| `publicincreasedepths.fromVersion` | string | From version number |
| `publicincreasedepths.toVersion` | string | To version number |

##### Sample Response

```json
{
  "channel": "spot@public.aggre.depth.v3.api.pb@100ms@BTCUSDT",
  "publicincreasedepths": {
    "asksList": [],
    "bidsList": [
      {
        "price": "92877.58",
        "quantity": "0.00000000"
      }
    ],
    "eventtype": "spot@public.aggre.depth.v3.api.pb@100ms",
    "fromVersion": "10589632359",
    "toVersion": "10589632359"
  },
  "symbol": "BTCUSDT",
  "sendtime": 1736411507002
}
```

#### Partial Book Depth Stream (Snapshot)
- **Description**: Pushes limited level depth snapshots for buy and sell orders
- **Method/Topic**: `spot@public.limit.depth.v3.api.pb@<symbol>@<level>`
- **Available Levels**: `5`, `10`, `20`

##### Subscription Request Format

```json
{
  "method": "SUBSCRIPTION",
  "params": ["spot@public.limit.depth.v3.api.pb@BTCUSDT@5"]
}
```

##### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `channel` | string | Channel identifier |
| `symbol` | string | Trading pair |
| `sendtime` | long | Event time (Unix ms) |
| `publiclimitdepths.asksList[]` | array | Ask (sell) orders |
| `publiclimitdepths.bidsList[]` | array | Bid (buy) orders |
| `publiclimitdepths.[].price` | string | Price level |
| `publiclimitdepths.[].quantity` | string | Quantity |
| `publiclimitdepths.eventtype` | string | Event type identifier |
| `publiclimitdepths.version` | string | Version number |

##### Sample Response

```json
{
  "channel": "spot@public.limit.depth.v3.api.pb@BTCUSDT@5",
  "publiclimitdepths": {
    "asksList": [
      {
        "price": "93180.18",
        "quantity": "0.21976424"
      }
    ],
    "bidsList": [
      {
        "price": "93179.98",
        "quantity": "2.82651000"
      }
    ],
    "eventtype": "spot@public.limit.depth.v3.api.pb",
    "version": "36913565463"
  },
  "symbol": "BTCUSDT",
  "sendtime": 1736411838730
}
```

#### Snapshot vs Incremental Updates
- **Incremental** (`aggre.depth`): Pushes only changes. Use `fromVersion`/`toVersion` to maintain ordering. A quantity of `0` means the price level should be removed.
- **Snapshot** (`limit.depth`): Pushes full top-N levels on each update. Simpler to use but limited to 5/10/20 levels.
- **Maintaining a local order book**: See the [Order Book Maintenance](#order-book-maintenance) section below.

### Candlestick/Kline Stream
- **Description**: Pushes updates to the current klines/candlestick every second
- **Method/Topic**: `spot@public.kline.v3.api.pb@<symbol>@<interval>`

#### Available Intervals

| Interval | Description |
|----------|-------------|
| `Min1` | 1 minute |
| `Min5` | 5 minutes |
| `Min15` | 15 minutes |
| `Min30` | 30 minutes |
| `Min60` | 60 minutes |
| `Hour4` | 4 hours |
| `Hour8` | 8 hours |
| `Day1` | 1 day |
| `Week1` | 1 week |
| `Month1` | 1 month |

#### Subscription Request Format

```json
{
  "method": "SUBSCRIPTION",
  "params": ["spot@public.kline.v3.api.pb@BTCUSDT@Min15"]
}
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `channel` | string | Channel identifier |
| `symbol` | string | Trading pair |
| `symbolid` | string | Trading pair ID |
| `createtime` | long | Event time (Unix ms) |
| `publicspotkline.interval` | string | K-line interval |
| `publicspotkline.windowstart` | long | Start time of the K-line (Unix s) |
| `publicspotkline.openingprice` | bigDecimal | Opening price |
| `publicspotkline.closingprice` | bigDecimal | Closing price |
| `publicspotkline.highestprice` | bigDecimal | Highest price |
| `publicspotkline.lowestprice` | bigDecimal | Lowest price |
| `publicspotkline.volume` | bigDecimal | Trade volume (base) |
| `publicspotkline.amount` | bigDecimal | Trade amount (quote) |
| `publicspotkline.windowend` | long | End time of the K-line (Unix s) |

#### Sample Response

```json
{
  "channel": "spot@public.kline.v3.api.pb@BTCUSDT@Min15",
  "publicspotkline": {
    "interval": "Min15",
    "windowstart": 1736410500,
    "openingprice": "92925",
    "closingprice": "93158.47",
    "highestprice": "93158.47",
    "lowestprice": "92800",
    "volume": "36.83803224",
    "amount": "3424811.05",
    "windowend": 1736411400
  },
  "symbol": "BTCUSDT",
  "symbolid": "2fb942154ef44a4ab2ef98c8afb6a4a7",
  "createtime": 1736410707571
}
```

### Individual Symbol Book Ticker Stream
- **Description**: Pushes any update to the best bid or ask's price or quantity in real-time for a specified symbol
- **Method/Topic**: `spot@public.aggre.bookTicker.v3.api.pb@(100ms|10ms)@<symbol>`

#### Subscription Request Format

```json
{
  "method": "SUBSCRIPTION",
  "params": ["spot@public.aggre.bookTicker.v3.api.pb@100ms@BTCUSDT"]
}
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `channel` | string | Channel identifier |
| `symbol` | string | Trading pair |
| `sendtime` | long | Event time (Unix ms) |
| `publicbookticker.bidprice` | string | Best bid price |
| `publicbookticker.bidquantity` | string | Best bid quantity |
| `publicbookticker.askprice` | string | Best ask price |
| `publicbookticker.askquantity` | string | Best ask quantity |

#### Sample Response

```json
{
  "channel": "spot@public.aggre.bookTicker.v3.api.pb@100ms@BTCUSDT",
  "publicbookticker": {
    "bidprice": "93387.28",
    "bidquantity": "3.73485",
    "askprice": "93387.29",
    "askquantity": "7.669875"
  },
  "symbol": "BTCUSDT",
  "sendtime": 1736412092433
}
```

### Individual Symbol Book Ticker Stream (Batch Aggregation)
- **Description**: Pushes the best order information for a specified trading pair in batch format
- **Method/Topic**: `spot@public.bookTicker.batch.v3.api.pb@<symbol>`

#### Subscription Request Format

```json
{
  "method": "SUBSCRIPTION",
  "params": ["spot@public.bookTicker.batch.v3.api.pb@BTCUSDT"]
}
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `channel` | string | Channel identifier |
| `symbol` | string | Trading pair |
| `sendTime` | long | Event time (Unix ms) |
| `publicBookTickerBatch.items[]` | array | Batch items |
| `publicBookTickerBatch.items[].bidPrice` | string | Best bid price |
| `publicBookTickerBatch.items[].bidQuantity` | string | Best bid quantity |
| `publicBookTickerBatch.items[].askPrice` | string | Best ask price |
| `publicBookTickerBatch.items[].askQuantity` | string | Best ask quantity |

#### Sample Response

```json
{
  "channel": "spot@public.bookTicker.batch.v3.api.pb@BTCUSDT",
  "symbol": "BTCUSDT",
  "sendTime": "1739503249114",
  "publicBookTickerBatch": {
    "items": [
      {
        "bidPrice": "96567.37",
        "bidQuantity": "3.362925",
        "askPrice": "96567.38",
        "askQuantity": "1.545255"
      }
    ]
  }
}
```

## Constraints & Limits

| Constraint | Value |
|------------|-------|
| Max subscriptions per connection | 30 |
| Connection validity | 24 hours |
| Idle timeout (no subscription) | 30 seconds |
| Idle timeout (no data flow) | 1 minute |
| Symbol format | UPPERCASE (e.g., `BTCUSDT`) |

## Order Book Maintenance

To properly maintain a local copy of the order book:

1. Open a WebSocket connection and subscribe to `spot@public.aggre.depth.v3.api.pb@(100ms|10ms)@{symbol}`
2. Start caching received incremental depth updates; record the `fromVersion` of the first update
3. Request an order book snapshot from `GET /api/v3/depth?symbol={symbol}&limit=5000` and record the `lastUpdateId`
4. If `lastUpdateId` < first cached `fromVersion`, re-fetch the snapshot (step 3)
5. Discard all cached updates where `toVersion` <= `lastUpdateId`
6. Take the first remaining update. If `fromVersion` > `lastUpdateId + 1`, data is discontinuous; reinitialize from step 3
7. Initialize the local order book with the snapshot data, set local version to `lastUpdateId`, then sequentially apply all remaining cached updates and subsequent incoming updates. After each update, set local version to the update's `toVersion`. If any update does not satisfy `fromVersion = previous toVersion + 1`, reinitialize from step 3

**Note**: The depth snapshot is limited to 5000 levels. Price levels outside the initial snapshot that have not changed in quantity will not appear in incremental push messages. The local order book may differ slightly from the real order book, but 5000 levels is sufficient for most use cases.

## Notes

### Protocol Buffers (Protobuf)
All WebSocket streams use **protobuf** format, not plain JSON. Integration steps:

1. Get `.proto` files from: https://github.com/mexcdevelop/websocket-proto
2. Generate deserialization code using `protoc` (supports Go, Java, Python, C++, etc.)
3. Deserialize incoming binary messages using the generated code

Example Go protobuf deserialization:
```go
// Deserialize incoming binary message
result, err := PushDataV3ApiWrapper_ParseFrom(data)
```

### Symbol Format Conventions
- All symbols must be **UPPERCASE** (e.g., `BTCUSDT`, not `btcusdt`)
- Spot trading pairs follow the `<BASE><QUOTE>` format (e.g., `BTCUSDT`, `ETHBTC`)

### Topic Naming Convention
The general pattern is:
```
spot@public.<type>.v3.api.pb@[<interval>@]<symbol>
spot@public.<type>.v3.api.pb@[<push-rate>@]<symbol>
spot@public.<type>.v3.api.pb@<symbol>@<level>
spot@public.<type>.v3.api.pb@<symbol>@<timezone>
```

Where `<type>` is one of:
- `aggre.deals` — Trade stream
- `kline` — Candlestick stream
- `aggre.depth` — Incremental depth stream
- `limit.depth` — Partial book depth stream
- `aggre.bookTicker` — Best bid/ask stream
- `bookTicker.batch` — Batch book ticker stream
- `miniTicker` — Single symbol 24h ticker
- `miniTickers` — All symbols 24h ticker
