# Bitget WebSocket Market Documentation (Futures)

## Reference

- **Official API Docs**:
  - USDT-M Futures: https://www.bitget.com/api-doc/contract/websocket/public/Tickers-Channel
  - Coin-M Futures: https://www.bitget.com/api-doc/contract/websocket/public/Tickers-Channel
  - Legacy Docs: https://bitgetlimited.github.io/apidoc/en/mix/
- **WebSocket Base URL**: `wss://ws.bitget.com/v2/ws/public` (unified for USDT-M and Coin-M)

## Protocol Overview

### Connection

- **Protocol Type**: Raw WebSocket
- **Base URL**: `wss://ws.bitget.com/v2/ws/public`
- **Connection Limits**: Maximum 50 connections per IP address
- **Data Format**: JSON

### Keep-Alive / Ping-Pong Mechanism

- **Ping Interval**: Send a ping message every 30 seconds to maintain the connection
- **Ping Format**: `{"op": "ping"}`
- **Pong Response**: Server responds with `{"op": "pong"}`
- **Timeout**: If no pong received within 5 seconds, consider connection dead and reconnect

### Reconnection Guidelines

1. Implement exponential backoff for reconnection attempts
2. Start with 1 second delay, double on each failure up to 64 seconds
3. Re-subscribe to all channels after successful reconnection
4. Reset local order book snapshot on reconnection for depth channels

### Authentication

- **Public Streams**: No authentication required
- **Private Streams**: Require API Key, Secret Key, and Passphrase (not covered in this document)

## Message Format

### Request Format

All subscription requests follow this JSON structure:

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "USDT-FUTURES",
      "channel": "ticker",
      "instId": "BTCUSDT"
    }
  ]
}
```

| Parameter | Type   | Required | Description                           |
|-----------|--------|----------|---------------------------------------|
| op        | String | Yes      | Operation: `subscribe` or `unsubscribe` |
| args      | List   | Yes      | List of channel subscription objects  |
| instType  | String | Yes      | Product type (see below)              |
| channel   | String | Yes      | Channel name                          |
| instId    | String | Yes      | Product ID (e.g., BTCUSDT)            |

### instType Values

| Value          | Description        |
|----------------|--------------------|
| USDT-FUTURES   | USDT-M Futures     |
| COIN-FUTURES   | Coin-M Futures     |
| USDC-FUTURES   | USDC-M Futures     |

### Response Format

**Subscription Confirmation**:

```json
{
  "event": "subscribe",
  "arg": {
    "instType": "USDT-FUTURES",
    "channel": "ticker",
    "instId": "BTCUSDT"
  }
}
```

| Parameter | Type   | Description                    |
|-----------|--------|--------------------------------|
| event     | String | Event type (`subscribe`)       |
| arg       | Object | Subscribed channel information |
| code      | String | Error code (only on error)     |
| msg       | String | Error message (only on error)  |

### Error Format

```json
{
  "event": "error",
  "arg": {
    "instType": "USDT-FUTURES",
    "channel": "ticker",
    "instId": "BTCUSDT"
  },
  "code": 30003,
  "msg": "instType:USDT-FUTURES,channel:ticker,instId:BTCUSDT Symbol not exists",
  "op": "subscribe"
}
```

| Parameter | Type   | Description                    |
|-----------|--------|--------------------------------|
| event     | String | Always `error`                 |
| arg       | Object | The subscription argument that caused the error |
| code      | String | Error code                     |
| msg       | String | Detailed error message         |
| op        | String | The operation that failed      |

## WebSocket Endpoints

### Ticker Channel

- **Description**: Retrieve the latest traded price, bid price, ask price and 24-hour trading volume of the instruments. Pushes updates when there is a change (deal, buy, sell, issue) every 300ms to 400ms.
- **Channel**: `ticker`
- **instType**: `USDT-FUTURES` or `COIN-FUTURES`

#### Subscription Request Format

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "USDT-FUTURES",
      "channel": "ticker",
      "instId": "BTCUSDT"
    }
  ]
}
```

#### Response Fields

| Field           | Type   | Description                                                                 |
|-----------------|--------|-----------------------------------------------------------------------------|
| instId          | String | Product ID (e.g., BTCUSDT)                                                  |
| lastPr          | String | Latest price                                                                |
| askPr           | String | Ask price                                                                   |
| bidPr           | String | Bid price                                                                   |
| high24h         | String | 24h high price                                                              |
| low24h          | String | 24h low price                                                               |
| change24h       | String | 24h price change percentage                                                 |
| fundingRate     | String | Current funding rate                                                        |
| nextFundingTime | String | Next funding rate settlement time (Unix timestamp in milliseconds)          |
| ts              | String | System time (Unix timestamp in milliseconds)                                |
| markPrice       | String | Mark price                                                                  |
| indexPrice      | String | Index price                                                                 |
| holdingAmount   | String | Open interest                                                               |
| baseVolume      | String | Trading volume of the coin                                                  |
| quoteVolume     | String | Trading volume of quote currency                                            |
| openUtc         | String | Price at 00:00 UTC                                                          |
| symbolType      | String | Symbol type: `1` = perpetual, `2` = delivery                                |
| symbol          | String | Trading pair name                                                           |
| deliveryPrice   | String | Delivery price (always 0 for perpetual contracts)                           |
| bidSz           | String | Best bid size                                                               |
| askSz           | String | Best ask size                                                               |
| open24h         | String | Entry price of the last 24 hours                                            |

#### Sample Response

```json
{
  "data": [
    {
      "lastPr": "87673.6",
      "symbol": "BTCUSDT",
      "indexPrice": "87714.0732915359034044",
      "open24h": "87027.0",
      "nextFundingTime": "1766678400000",
      "bidPr": "87673.6",
      "change24h": "0.00743",
      "quoteVolume": "1521198076.61216",
      "deliveryPrice": "0",
      "askSz": "14.333",
      "low24h": "86542.5",
      "symbolType": "1",
      "openUtc": "87628.9",
      "instId": "BTCUSDT",
      "bidSz": "6.9129",
      "markPrice": "87673.7",
      "high24h": "88022.1",
      "askPr": "87673.7",
      "holdingAmount": "28135.5456",
      "baseVolume": "17398.1612",
      "fundingRate": "0.000055",
      "ts": "1766674540816"
    }
  ],
  "arg": {
    "instType": "USDT-FUTURES",
    "instId": "BTCUSDT",
    "channel": "ticker"
  },
  "action": "snapshot",
  "ts": 1766674540817
}
```

### Trade Channel

- **Description**: Get the public trade data (taker orders). Real-time push for all executed trades.
- **Channel**: `trade`
- **instType**: `USDT-FUTURES` or `COIN-FUTURES`

#### Subscription Request Format

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "USDT-FUTURES",
      "channel": "trade",
      "instId": "BTCUSDT"
    }
  ]
}
```

#### Response Fields

| Field    | Type   | Description                                      |
|----------|--------|--------------------------------------------------|
| ts       | String | Fill time (Unix timestamp in milliseconds)         |
| price    | String | Filled price                                     |
| size     | String | Filled amount                                    |
| side     | String | Filled side: `buy` or `sell`                     |
| tradeId  | String | Unique trade ID                                  |

#### Sample Response

```json
{
  "action": "snapshot",
  "arg": {
    "instType": "USDT-FUTURES",
    "channel": "trade",
    "instId": "BTCUSDT"
  },
  "data": [
    {
      "ts": "1695716760565",
      "price": "27000.5",
      "size": "0.001",
      "side": "buy",
      "tradeId": "1111111111"
    },
    {
      "ts": "1695716759514",
      "price": "27000.0",
      "size": "0.001",
      "side": "sell",
      "tradeId": "1111111111"
    }
  ],
  "ts": 1695716761589
}
```

### Depth Channel (Order Book)

- **Description**: Get order book depth data with various levels of granularity.
- **Channels**:
  - `books`: All levels of depth. First update is full snapshot, then incremental updates
  - `books1`: 1st level of depth. Pushes snapshot each time (10ms frequency)
  - `books5`: 5 depth levels. Pushes snapshot each time (150ms frequency)
  - `books15`: 15 depth levels. Pushes snapshot each time (150ms frequency)
- **instType**: `USDT-FUTURES`, `COIN-FUTURES`, or `USDC-FUTURES`

#### Subscription Request Format

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "USDT-FUTURES",
      "channel": "books5",
      "instId": "BTCUSDT"
    }
  ]
}
```

#### Response Fields

| Field     | Type   | Description                                                                 |
|-----------|--------|-----------------------------------------------------------------------------|
| asks      | List   | Seller depth (array of [price, size])                                       |
| bids      | List   | Buyer depth (array of [price, size])                                        |
| ts        | String | Match engine timestamp (milliseconds)                                       |
| checksum  | Long   | CRC32 checksum for data integrity verification                              |
| seq       | Long   | Serial number. Increases when order book is updated (for out-of-order detection) |

#### Sample Response

```json
{
  "action": "snapshot",
  "arg": {
    "instType": "USDT-FUTURES",
    "channel": "books5",
    "instId": "BTCUSDT"
  },
  "data": [
    {
      "asks": [
        ["27000.5", "8.760"],
        ["27001.0", "0.400"]
      ],
      "bids": [
        ["27000.0", "2.710"],
        ["26999.5", "1.460"]
      ],
      "checksum": 0,
      "seq": 123,
      "ts": "1695716059516"
    }
  ],
  "ts": 1695716059516
}
```

#### Snapshot vs Incremental Updates

- **books channel**: First push is a full snapshot (`action: "snapshot"`), subsequent pushes are incremental updates (`action: "update"`)
- **books1, books5, books15 channels**: Always push full snapshots

**Merging update data into snapshot**:

1. If there are levels with the same price in the updates:
   - If amount is 0, delete this depth data
   - If amount changes, replace the depth data
2. If no level exists with the same price, insert the new depth information sorted by price (bids descending, asks ascending)

#### Checksum Verification

The checksum mechanism helps verify the accuracy of order book data:

1. Take the first 25 bids and asks from the local snapshot
2. Build a string in the format: `bid1[price:amount]:ask1[price:amount]:bid2[price:amount]:ask2[price:amount]...`
3. Calculate the CRC32 value (32-bit signed integer)
4. Compare with the checksum value received in the push data

**Example checksum string construction**:

```
bids: [[43231.1, 4], [43231, 6]]
asks: [[43232.8, 9], [43232.9, 8]]

Checksum string: "43231.1:4:43232.8:9:43231:6:43232.9:8"
```

**Important**: If price is '0.5000', use the original value '0.5000' for checksum calculation, not '0.5'.

### Candlestick Channel

- **Description**: Pushes candlestick (K-line) data. After subscription, sends a snapshot followed by updates. Pushes once per second when there are transactions, or once per specified time granularity when there are no transactions.
- **Channels**:
  - `candle1m` (1 minute)
  - `candle5m` (5 minutes)
  - `candle15m` (15 minutes)
  - `candle30m` (30 minutes)
  - `candle1H` (1 hour)
  - `candle4H` (4 hours)
  - `candle6H` (6 hours)
  - `candle12H` (12 hours)
  - `candle1D` (1 day)
  - `candle3D` (3 days)
  - `candle1W` (1 week)
  - `candle1M` (1 month)
  - `candle6Hutc` (6 hours, UTC)
  - `candle12Hutc` (12 hours, UTC)
  - `candle1Dutc` (1 day, UTC)
  - `candle3Dutc` (3 days, UTC)
  - `candle1Wutc` (1 week, UTC)
  - `candle1Mutc` (1 month, UTC)
- **instType**: `USDT-FUTURES` or `COIN-FUTURES`

#### Subscription Request Format

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "USDT-FUTURES",
      "channel": "candle1m",
      "instId": "BTCUSDT"
    }
  ]
}
```

#### Response Fields

Data is returned as an array of strings with the following indices:

| Index | Type   | Description                                      |
|-------|--------|--------------------------------------------------|
| [0]   | String | Start time (Unix timestamp in milliseconds)        |
| [1]   | String | Opening price                                    |
| [2]   | String | Highest price                                    |
| [3]   | String | Lowest price                                     |
| [4]   | String | Closing price                                    |
| [5]   | String | Trading volume (base currency)                   |
| [6]   | String | Trading volume (quote currency)                  |
| [7]   | String | Trading volume (USDT)                            |

#### Sample Response

```json
{
  "action": "snapshot",
  "arg": {
    "instType": "USDT-FUTURES",
    "channel": "candle1m",
    "instId": "BTCUSDT"
  },
  "data": [
    [
      "1695685500000",
      "27000",
      "27000.5",
      "27000",
      "27000.5",
      "0.057",
      "1539.0155",
      "1539.0155"
    ]
  ],
  "ts": 1695715462250
}
```

### Mark Price Channel

- **Description**: Pushes mark price updates for futures contracts. Mark price is used for calculating unrealized PnL and liquidation.
- **Channel**: `mark-price`
- **instType**: `USDT-FUTURES` or `COIN-FUTURES`

#### Subscription Request Format

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "USDT-FUTURES",
      "channel": "mark-price",
      "instId": "BTCUSDT"
    }
  ]
}
```

#### Response Fields

| Field      | Type   | Description                                      |
|------------|--------|--------------------------------------------------|
| instId     | String | Product ID                                       |
| markPrice  | String | Mark price                                       |
| indexPrice | String | Index price                                      |
| ts         | String | Timestamp (Unix timestamp in milliseconds)         |

#### Sample Response

```json
{
  "action": "update",
  "arg": {
    "instType": "USDT-FUTURES",
    "channel": "mark-price",
    "instId": "BTCUSDT"
  },
  "data": [
    {
      "instId": "BTCUSDT",
      "markPrice": "87673.7",
      "indexPrice": "87714.0732915359034044",
      "ts": "1766674540816"
    }
  ],
  "ts": 1766674540817
}
```

### Funding Rate Channel

- **Description**: Pushes funding rate updates for perpetual futures contracts. Funding rate is exchanged between long and short positions periodically.
- **Channel**: `funding-rate`
- **instType**: `USDT-FUTURES` or `COIN-FUTURES`

#### Subscription Request Format

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "USDT-FUTURES",
      "channel": "funding-rate",
      "instId": "BTCUSDT"
    }
  ]
}
```

#### Response Fields

| Field           | Type   | Description                                      |
|-----------------|--------|--------------------------------------------------|
| instId          | String | Product ID                                       |
| fundingRate     | String | Current funding rate                             |
| nextFundingTime | String | Next funding time (Unix timestamp in milliseconds) |
| ts              | String | Timestamp (Unix timestamp in milliseconds)         |

#### Sample Response

```json
{
  "action": "update",
  "arg": {
    "instType": "USDT-FUTURES",
    "channel": "funding-rate",
    "instId": "BTCUSDT"
  },
  "data": [
    {
      "instId": "BTCUSDT",
      "fundingRate": "0.000055",
      "nextFundingTime": "1766678400000",
      "ts": "1766674540816"
    }
  ],
  "ts": 1766674540817
}
```

### Open Interest Channel

- **Description**: Pushes open interest updates for futures contracts. Open interest represents the total number of outstanding derivative contracts.
- **Channel**: `open-interest`
- **instType**: `USDT-FUTURES` or `COIN-FUTURES`

#### Subscription Request Format

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "USDT-FUTURES",
      "channel": "open-interest",
      "instId": "BTCUSDT"
    }
  ]
}
```

#### Response Fields

| Field         | Type   | Description                                      |
|---------------|--------|--------------------------------------------------|
| instId        | String | Product ID                                       |
| holdingAmount | String | Open interest amount                             |
| ts            | String | Timestamp (Unix timestamp in milliseconds)         |

#### Sample Response

```json
{
  "action": "update",
  "arg": {
    "instType": "USDT-FUTURES",
    "channel": "open-interest",
    "instId": "BTCUSDT"
  },
  "data": [
    {
      "instId": "BTCUSDT",
      "holdingAmount": "28135.5456",
      "ts": "1766674540816"
    }
  ],
  "ts": 1766674540817
}
```

### Liquidation Orders Channel

- **Description**: Pushes liquidation order data when positions are forcibly closed due to insufficient margin.
- **Channel**: `liquidation-orders`
- **instType**: `USDT-FUTURES` or `COIN-FUTURES`

#### Subscription Request Format

```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "USDT-FUTURES",
      "channel": "liquidation-orders",
      "instId": "BTCUSDT"
    }
  ]
}
```

#### Response Fields

| Field    | Type   | Description                                      |
|----------|--------|--------------------------------------------------|
| instId   | String | Product ID                                       |
| side     | String | Liquidation side: `buy` or `sell`                |
| price    | String | Liquidation price                                |
| size     | String | Liquidation amount                               |
| ts       | String | Timestamp (Unix timestamp in milliseconds)         |

#### Sample Response

```json
{
  "action": "update",
  "arg": {
    "instType": "USDT-FUTURES",
    "channel": "liquidation-orders",
    "instId": "BTCUSDT"
  },
  "data": [
    {
      "instId": "BTCUSDT",
      "side": "sell",
      "price": "85000.0",
      "size": "0.5",
      "ts": "1766674540816"
    }
  ],
  "ts": 1766674540817
}
```

## Constraints & Limits

### Connection Limits

- Maximum **50 connections** per IP address
- Each connection can subscribe to multiple channels

### Subscription Limits

- Maximum **300 subscriptions** per connection
- Each subscription counts as one channel + one instrument combination

### Message Rate Limits

- **Public channels**: Maximum 20 requests per second per IP
- **Data push frequency**:
  - Ticker: 300ms - 400ms on change
  - Trade: Real-time
  - books1: 10ms
  - books5, books15: 150ms
  - books: 150ms (incremental)
  - Candlesticks: 1 second when active, or per granularity when inactive

### Other Constraints

- All timestamps are in milliseconds (Unix epoch)
- All numeric values are returned as strings
- Symbol names are case-sensitive
- WebSocket connections should be kept alive with ping/pong every 30 seconds
- If connection is lost, all subscriptions must be re-established

## Notes

### Symbol Format Conventions

**USDT-M Futures**:
- Format: `{BASE}{QUOTE}` (e.g., `BTCUSDT`, `ETHUSDT`)
- Margin currency: USDT
- Example: `BTCUSDT`

**Coin-M Futures**:
- Format: `{BASE}{QUOTE}` (e.g., `BTCUSD`, `ETHUSD`)
- Margin currency: The base coin (e.g., BTC, ETH)
- Example: `BTCUSD`

### Differences Between USDT-M and Coin-M

| Aspect           | USDT-M Futures              | Coin-M Futures              |
|------------------|-----------------------------|-----------------------------|
| Margin Currency  | USDT                        | Base coin (BTC, ETH, etc.)  |
| Settlement       | USDT                        | Base coin                   |
| Symbol Format    | BTCUSDT, ETHUSDT            | BTCUSD, ETHUSD              |
| instType         | `USDT-FUTURES`              | `COIN-FUTURES`              |
| PnL Calculation  | In USDT                     | In base coin                |
| Contract Value   | Fixed in USDT               | Fixed in base coin          |

### Implementation Notes

1. **Order Book Management**: When using the `books` channel, maintain a local order book and apply incremental updates. Use the `seq` field to detect out-of-order packets and the `checksum` field to verify data integrity.

2. **Reconnection Strategy**: Always re-subscribe to all channels after reconnection. For order book channels, discard the local snapshot and rebuild from the new snapshot.

3. **Rate Limiting**: Be mindful of the subscription limits and message rate limits. Subscribe only to the instruments and channels you need.

4. **Data Types**: All numeric values in WebSocket messages are returned as strings. Convert to appropriate numeric types in your application.

5. **Time Synchronization**: Ensure your system clock is synchronized with NTP servers to properly handle timestamp-based operations.

### REST API Equivalents

For historical data or data not available via WebSocket, use the REST API:

| WebSocket Channel     | REST API Endpoint                    |
|-----------------------|--------------------------------------|
| Ticker                | `/api/mix/v1/market/ticker`          |
| Trades                | `/api/mix/v1/market/fills`           |
| Order Book            | `/api/mix/v1/market/depth`           |
| Candlesticks          | `/api/mix/v1/market/candles`         |
| Mark Price            | `/api/mix/v1/market/mark-price`      |
| Funding Rate          | `/api/mix/v1/market/funding-rate`    |
| Open Interest         | `/api/mix/v1/market/open-interest`   |
