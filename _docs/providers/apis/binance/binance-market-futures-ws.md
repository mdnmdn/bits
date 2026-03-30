# Binance WebSocket Market Documentation (Futures)

## Reference
- **Official API Docs**:
  - USDT-M Futures: https://developers.binance.com/docs/derivatives/usds-margined-futures/websocket-market-streams
  - COIN-M Futures: https://developers.binance.com/docs/derivatives/coin-margined-futures/websocket-market-streams
- **WebSocket Base URLs**:
  - USDT-M: `wss://fstream.binance.com/ws`
  - USDT-M Combined: `wss://fstream.binance.com/stream?streams=`
  - COIN-M: `wss://dstream.binance.com/ws`
  - COIN-M Combined: `wss://dstream.binance.com/stream?streams=`

## Protocol Overview

- **Protocol type**: Raw WebSocket (no authentication required for market streams)
- **Connection limits**:
  - Maximum **1024 streams** per single connection
  - Maximum **10 incoming messages per second** per connection
  - Connections exceeding the message rate limit will be disconnected; IPs that are repeatedly disconnected may be banned
- **Keep-alive / ping-pong mechanism**:
  - Server sends a `ping frame` every **3 minutes**
  - Client must respond with a `pong frame` within **10 minutes** or the connection will be disconnected
  - Unsolicited `pong frames` are allowed (clients can send pong frames at a frequency higher than every 10 minutes to maintain the connection)
  - When receiving a ping, clients should send a pong with a copy of ping's payload as soon as possible
- **Reconnection guidelines**:
  - A single connection is only valid for **24 hours**; expect to be disconnected at the 24-hour mark
  - Implement automatic reconnection with exponential backoff
  - Re-subscribe to all streams after reconnection
- **URL paths**:
  - Raw streams: `/ws/<streamName>`
  - Combined streams: `/stream?streams=<streamName1>/<streamName2>/<streamName3>`
  - Market streams use `/market` or `/public` URL paths internally
- **Differences between USDT-M and COIN-M endpoints**:
  - USDT-M uses `fstream.binance.com`, COIN-M uses `dstream.binance.com`
  - COIN-M has additional streams: Index Price, Index Kline, Mark Price Kline, and pair-based mark price streams
  - COIN-M mark price for all symbols is organized by pair (e.g., `btcusd@markPrice`), not a global `!markPrice@arr`
  - Symbol formats differ: USDT-M uses `btcusdt`, COIN-M uses `btcusd_PERP` for perpetual and `btcusd_201225` for quarterly

## Stream Naming Convention

- All symbols for streams are **lowercase**
- USDT-M symbol format: `<base><quote>` (e.g., `btcusdt`, `ethusdt`, `bnbusdt`)
- COIN-M symbol format:
  - Perpetual: `<base><quote>_PERP` (e.g., `btcusd_PERP`, `ethusd_PERP`)
  - Quarterly delivery: `<base><quote>_YYMMDD` (e.g., `btcusd_201225`, `ethusd_200925`)
- Combined stream format: `{"stream":"<streamName>","data":<rawPayload>}`
- Example combined stream URL: `wss://fstream.binance.com/stream?streams=btcusdt@aggTrade/ethusdt@markPrice`

## WebSocket Endpoints

### Aggregate Trade Stream

- **Description**: Pushes market trade information aggregated for fills with same price and taking side every 100ms. Only market trades are aggregated; insurance fund trades and ADL trades are excluded. RPI orders are aggregated into field `q` without special tags.
- **Stream Name**: `<symbol>@aggTrade`
- **Update Speed**: 100ms
- **Response Fields**:

| Field | Type   | Description                                      |
|-------|--------|--------------------------------------------------|
| e     | string | Event type (`aggTrade`)                          |
| E     | int    | Event time (ms)                                  |
| s     | string | Symbol                                           |
| a     | int    | Aggregate trade ID                               |
| p     | string | Price                                            |
| q     | string | Quantity (all market trades including RPI)         |
| nq    | string | Normal quantity (excluding RPI trades)            |
| f     | int    | First trade ID                                   |
| l     | int    | Last trade ID                                    |
| T     | int    | Trade time (ms)                                  |
| m     | bool   | Is the buyer the market maker?                   |

- **Sample Response**:

```json
{
  "e": "aggTrade",
  "E": 123456789,
  "s": "BTCUSDT",
  "a": 5933014,
  "p": "0.001",
  "q": "100",
  "nq": "100",
  "f": 100,
  "l": 105,
  "T": 123456785,
  "m": true
}
```

### Mark Price Stream

- **Description**: Mark price and funding rate for a single symbol. Futures-specific: includes mark price, estimated settle price, index price, funding rate, and next funding time.
- **Stream Name**: `<symbol>@markPrice` or `<symbol>@markPrice@1s`
- **Update Speed**: 3000ms (default) or 1000ms
- **Response Fields**:

| Field | Type   | Description                                                        |
|-------|--------|--------------------------------------------------------------------|
| e     | string | Event type (`markPriceUpdate`)                                     |
| E     | int    | Event time (ms)                                                    |
| s     | string | Symbol                                                             |
| p     | string | Mark price                                                         |
| ap    | string | Mark price moving average (USDT-M only, added 2026-03-16)          |
| i     | string | Index price                                                        |
| P     | string | Estimated settle price (only useful in last hour before settlement)|
| r     | string | Funding rate                                                       |
| T     | int    | Next funding time (ms)                                             |

- **Sample Response** (USDT-M):

```json
{
  "e": "markPriceUpdate",
  "E": 1562305380000,
  "s": "BTCUSDT",
  "p": "11794.15000000",
  "ap": "11794.15000000",
  "i": "11784.62659091",
  "P": "11784.25641265",
  "r": "0.00038167",
  "T": 1562306400000
}
```

- **Sample Response** (COIN-M):

```json
{
  "e": "markPriceUpdate",
  "E": 1596095725000,
  "s": "BTCUSD_201225",
  "p": "10934.62615417",
  "P": "10962.17178236",
  "i": "10933.62615417",
  "r": "",
  "T": 0
}
```

> **Note**: For COIN-M delivery symbols, `r` (funding rate) and `T` (next funding time) will be empty/0 since delivery contracts do not have funding rates.

### Mark Price Stream for All Symbols

- **Description**: Mark price and funding rate for all symbols. USDT-M uses `!markPrice@arr` for all symbols. COIN-M uses `<pair>@markPrice` for all symbols of a specific pair.
- **Stream Name**:
  - USDT-M: `!markPrice@arr` or `!markPrice@arr@1s`
  - COIN-M: `<pair>@markPrice` or `<pair>@markPrice@1s` (e.g., `btcusd@markPrice`)
- **Update Speed**: 3000ms (default) or 1000ms
- **Response Fields**: Array of mark price objects (same fields as individual mark price stream)
- **Sample Response** (USDT-M):

```json
[
  {
    "e": "markPriceUpdate",
    "E": 1562305380000,
    "s": "BTCUSDT",
    "p": "11185.87786614",
    "ap": "11185.87786614",
    "i": "11784.62659091",
    "P": "11784.25641265",
    "r": "0.00030000",
    "T": 1562306400000
  }
]
```

- **Sample Response** (COIN-M):

```json
[
  {
    "e": "markPriceUpdate",
    "E": 1596095725000,
    "s": "BTCUSD_201225",
    "p": "10934.62615417",
    "P": "10962.17178236",
    "i": "10933.62615417",
    "r": "",
    "T": 0
  },
  {
    "e": "markPriceUpdate",
    "E": 1596095725000,
    "s": "BTCUSD_PERP",
    "p": "11012.31359011",
    "P": "10962.17178236",
    "i": "10933.62615417",
    "r": "0.00000000",
    "T": 1596096000000
  }
]
```

### Kline/Candlestick Stream

- **Description**: Pushes updates to the current klines/candlestick every 250ms.
- **Stream Name**: `<symbol>@kline_<interval>`
- **Available Intervals**: 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M
- **Update Speed**: 250ms
- **Response Fields**:

| Field | Type   | Description                          |
|-------|--------|--------------------------------------|
| e     | string | Event type (`kline`)                 |
| E     | int    | Event time (ms)                      |
| s     | string | Symbol                               |
| k.t   | int    | Kline start time (ms)                |
| k.T   | int    | Kline close time (ms)                |
| k.s   | string | Symbol                               |
| k.i   | string | Interval                             |
| k.f   | int    | First trade ID                       |
| k.L   | int    | Last trade ID                        |
| k.o   | string | Open price                           |
| k.c   | string | Close price                          |
| k.h   | string | High price                           |
| k.l   | string | Low price                            |
| k.v   | string | Base asset volume                    |
| k.n   | int    | Number of trades                     |
| k.x   | bool   | Is this kline closed?                |
| k.q   | string | Quote asset volume                   |
| k.V   | string | Taker buy base asset volume          |
| k.Q   | string | Taker buy quote asset volume         |
| k.B   | string | Ignore                               |

- **Sample Response**:

```json
{
  "e": "kline",
  "E": 1638747660000,
  "s": "BTCUSDT",
  "k": {
    "t": 1638747660000,
    "T": 1638747719999,
    "s": "BTCUSDT",
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

### Continuous Contract Kline Stream

- **Description**: Futures-specific kline streams for continuous contracts (perpetual, current_quarter, next_quarter, tradifi_perpetual). Tracks price across contract rollovers.
- **Stream Name**: `<pair>_<contractType>@continuousKline_<interval>`
- **Contract Types**: `perpetual`, `current_quarter`, `next_quarter`, `tradifi_perpetual`
- **Available Intervals**: 1s, 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M
- **Update Speed**: 250ms
- **Response Fields**:

| Field | Type   | Description                          |
|-------|--------|--------------------------------------|
| e     | string | Event type (`continuous_kline`)      |
| E     | int    | Event time (ms)                      |
| ps    | string | Pair                                 |
| ct    | string | Contract type                        |
| k.t   | int    | Kline start time (ms)                |
| k.T   | int    | Kline close time (ms)                |
| k.i   | string | Interval                             |
| k.f   | int    | First updateId                       |
| k.L   | int    | Last updateId                        |
| k.o   | string | Open price                           |
| k.c   | string | Close price                          |
| k.h   | string | High price                           |
| k.l   | string | Low price                            |
| k.v   | string | Volume                               |
| k.n   | int    | Number of trades                     |
| k.x   | bool   | Is this kline closed?                |
| k.q   | string | Quote asset volume                   |
| k.V   | string | Taker buy volume                     |
| k.Q   | string | Taker buy quote asset volume         |
| k.B   | string | Ignore                               |

- **Sample Response**:

```json
{
  "e": "continuous_kline",
  "E": 1607443058651,
  "ps": "BTCUSDT",
  "ct": "PERPETUAL",
  "k": {
    "t": 1607443020000,
    "T": 1607443079999,
    "i": "1m",
    "f": 116467658886,
    "L": 116468012423,
    "o": "18787.00",
    "c": "18804.04",
    "h": "18804.04",
    "l": "18786.54",
    "v": "197.664",
    "n": 543,
    "x": false,
    "q": "3715253.19494",
    "V": "184.769",
    "Q": "3472925.84746",
    "B": "0"
  }
}
```

### Index Price Stream (COIN-M only)

- **Description**: Index price stream for COIN-M futures pairs.
- **Stream Name**: `<pair>@indexPrice` or `<pair>@indexPrice@1s`
- **Update Speed**: 3000ms (default) or 1000ms
- **Response Fields**:

| Field | Type   | Description                     |
|-------|--------|---------------------------------|
| e     | string | Event type (`indexPriceUpdate`) |
| E     | int    | Event time (ms)                 |
| i     | string | Pair                            |
| p     | string | Index price                     |

- **Sample Response**:

```json
{
  "e": "indexPriceUpdate",
  "E": 1591261236000,
  "i": "BTCUSD",
  "p": "9636.57860000"
}
```

### Index Kline Stream (COIN-M only)

- **Description**: Kline/candlestick stream for index prices of COIN-M futures pairs.
- **Stream Name**: `<pair>@indexPriceKline_<interval>`
- **Available Intervals**: Same as regular kline streams

### Mark Price Kline Stream (COIN-M only)

- **Description**: Kline/candlestick stream for mark prices of COIN-M futures symbols.
- **Stream Name**: `<symbol>@markPriceKline_<interval>`
- **Available Intervals**: Same as regular kline streams

### Individual Symbol Mini Ticker Stream

- **Description**: 24hr rolling window mini-ticker statistics for a single symbol. These are NOT the statistics of the UTC day, but a 24hr rolling window from requestTime to 24hrs before.
- **Stream Name**: `<symbol>@miniTicker`
- **Update Speed**: 2000ms
- **Response Fields**:

| Field | Type   | Description                              |
|-------|--------|------------------------------------------|
| e     | string | Event type (`24hrMiniTicker`)            |
| E     | int    | Event time (ms)                          |
| s     | string | Symbol                                   |
| c     | string | Close price                              |
| o     | string | Open price                               |
| h     | string | High price                               |
| l     | string | Low price                                |
| v     | string | Total traded base asset volume           |
| q     | string | Total traded quote asset volume          |

- **Sample Response**:

```json
{
  "e": "24hrMiniTicker",
  "E": 123456789,
  "s": "BTCUSDT",
  "c": "0.0025",
  "o": "0.0010",
  "h": "0.0025",
  "l": "0.0010",
  "v": "10000",
  "q": "18"
}
```

### All Market Mini Tickers Stream

- **Description**: 24hr rolling window mini-ticker statistics for all symbols. Only tickers that have changed will be present in the array.
- **Stream Name**: `!miniTicker@arr`
- **Update Speed**: 1000ms
- **Response Fields**: Array of mini-ticker objects (same fields as individual mini-ticker stream)
- **Sample Response**:

```json
[
  {
    "e": "24hrMiniTicker",
    "E": 123456789,
    "s": "BTCUSDT",
    "c": "0.0025",
    "o": "0.0010",
    "h": "0.0025",
    "l": "0.0010",
    "v": "10000",
    "q": "18"
  }
]
```

### Individual Symbol Ticker Stream

- **Description**: 24hr rolling window ticker statistics for a single symbol.
- **Stream Name**: `<symbol>@ticker`
- **Update Speed**: 2000ms
- **Response Fields**:

| Field | Type   | Description                              |
|-------|--------|------------------------------------------|
| e     | string | Event type (`24hrTicker`)                |
| E     | int    | Event time (ms)                          |
| s     | string | Symbol                                   |
| p     | string | Price change                             |
| P     | string | Price change percent                     |
| w     | string | Weighted average price                   |
| c     | string | Last price                               |
| Q     | string | Last quantity                            |
| o     | string | Open price                               |
| h     | string | High price                               |
| l     | string | Low price                                |
| v     | string | Total traded base asset volume           |
| q     | string | Total traded quote asset volume          |
| O     | int    | Statistics open time (ms)                |
| C     | int    | Statistics close time (ms)               |
| F     | int    | First trade ID                           |
| L     | int    | Last trade ID                            |
| n     | int    | Total number of trades                   |

- **Sample Response**:

```json
{
  "e": "24hrTicker",
  "E": 123456789,
  "s": "BTCUSDT",
  "p": "0.0015",
  "P": "250.00",
  "w": "0.0018",
  "c": "0.0025",
  "Q": "10",
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

### All Market Tickers Stream

- **Description**: 24hr rolling window ticker statistics for all symbols. Only tickers that have changed will be present in the array.
- **Stream Name**: `!ticker@arr`
- **Update Speed**: 1000ms
- **Response Fields**: Array of ticker objects (same fields as individual ticker stream)
- **Sample Response**:

```json
[
  {
    "e": "24hrTicker",
    "E": 123456789,
    "s": "BTCUSDT",
    "p": "0.0015",
    "P": "250.00",
    "w": "0.0018",
    "c": "0.0025",
    "Q": "10",
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
]
```

### Individual Symbol Book Ticker Stream

- **Description**: Pushes any update to the best bid or ask's price or quantity in real-time for a specified symbol. RPI orders are excluded.
- **Stream Name**: `<symbol>@bookTicker`
- **Update Speed**: Real-time
- **Response Fields**:

| Field | Type   | Description                     |
|-------|--------|---------------------------------|
| e     | string | Event type (`bookTicker`)       |
| u     | int    | Order book updateId             |
| E     | int    | Event time (ms)                 |
| T     | int    | Transaction time (ms)           |
| s     | string | Symbol                          |
| b     | string | Best bid price                  |
| B     | string | Best bid qty                    |
| a     | string | Best ask price                  |
| A     | string | Best ask qty                    |

- **Sample Response**:

```json
{
  "e": "bookTicker",
  "u": 400900217,
  "E": 1568014460893,
  "T": 1568014460891,
  "s": "BNBUSDT",
  "b": "25.35190000",
  "B": "31.21000000",
  "a": "25.36520000",
  "A": "40.66000000"
}
```

### All Book Tickers Stream

- **Description**: Pushes any update to the best bid or ask's price or quantity in real-time for all symbols. RPI orders are excluded.
- **Stream Name**: `!bookTicker`
- **Update Speed**: 5000ms (updated from real-time in December 2023)
- **Response Fields**: Same as individual book ticker stream
- **Sample Response**:

```json
{
  "e": "bookTicker",
  "u": 400900217,
  "E": 1568014460893,
  "T": 1568014460891,
  "s": "BNBUSDT",
  "b": "25.35190000",
  "B": "31.21000000",
  "a": "25.36520000",
  "A": "40.66000000"
}
```

### Partial Book Depth Stream

- **Description**: Top bids and asks at specified depth levels. RPI orders are excluded.
- **Stream Name**: `<symbol>@depth<levels>` or `<symbol>@depth<levels>@500ms` or `<symbol>@depth<levels>@100ms`
  - Valid `<levels>`: 5, 10, 20
- **Update Speed**: 250ms (default), 500ms, or 100ms
- **Response Fields**:

| Field | Type   | Description                                   |
|-------|--------|-----------------------------------------------|
| e     | string | Event type (`depthUpdate`)                    |
| E     | int    | Event time (ms)                               |
| T     | int    | Transaction time (ms)                         |
| s     | string | Symbol                                        |
| U     | int    | First update ID in event                      |
| u     | int    | Final update ID in event                      |
| pu    | int    | Final update ID in last stream (`u` in last)  |
| b     | array  | Bids to be updated [price, qty]               |
| a     | array  | Asks to be updated [price, qty]               |

- **Sample Response**:

```json
{
  "e": "depthUpdate",
  "E": 1571889248277,
  "T": 1571889248276,
  "s": "BTCUSDT",
  "U": 390497796,
  "u": 390497878,
  "pu": 390497794,
  "b": [
    ["7403.89", "0.002"],
    ["7403.90", "3.906"],
    ["7404.00", "1.428"],
    ["7404.85", "5.239"],
    ["7405.43", "2.562"]
  ],
  "a": [
    ["7405.96", "3.340"],
    ["7406.63", "4.525"],
    ["7407.08", "2.475"],
    ["7407.15", "4.800"],
    ["7407.20", "0.175"]
  ]
}
```

### Diff. Book Depth Stream

- **Description**: Full bids and asks updates (not limited to top N levels). RPI orders are excluded.
- **Stream Name**: `<symbol>@depth` or `<symbol>@depth@500ms` or `<symbol>@depth@100ms`
- **Update Speed**: 250ms (default), 500ms, or 100ms
- **Response Fields**: Same as partial book depth stream
- **Sample Response**:

```json
{
  "e": "depthUpdate",
  "E": 123456789,
  "T": 123456788,
  "s": "BTCUSDT",
  "U": 157,
  "u": 160,
  "pu": 149,
  "b": [
    ["0.0024", "10"]
  ],
  "a": [
    ["0.0026", "100"]
  ]
}
```

### Liquidation Order Streams

- **Description**: Futures-specific force liquidation order streams. Pushes force liquidation order information. Only the latest liquidation order within 1000ms is pushed as a snapshot. If no liquidation occurs in the interval, no stream is pushed.
- **Stream Name**:
  - Individual symbol: `<symbol>@forceOrder`
  - All market: `!forceOrder@arr`
- **Update Speed**: 1000ms
- **Response Fields**:

| Field  | Type   | Description                              |
|--------|--------|------------------------------------------|
| e      | string | Event type (`forceOrder`)                |
| E      | int    | Event time (ms)                          |
| o.s    | string | Symbol                                   |
| o.S    | string | Side (`BUY` or `SELL`)                   |
| o.o    | string | Order type (`LIMIT`)                     |
| o.f    | string | Time in force (`IOC`)                    |
| o.q    | string | Original quantity                        |
| o.p    | string | Price                                    |
| o.ap   | string | Average price                            |
| o.X    | string | Order status (`FILLED`)                  |
| o.l    | string | Order last filled quantity               |
| o.z    | string | Order filled accumulated quantity        |
| o.T    | int    | Order trade time (ms)                    |

- **Sample Response**:

```json
{
  "e": "forceOrder",
  "E": 1568014460893,
  "o": {
    "s": "BTCUSDT",
    "S": "SELL",
    "o": "LIMIT",
    "f": "IOC",
    "q": "0.014",
    "p": "9910",
    "ap": "9910",
    "X": "FILLED",
    "l": "0.014",
    "z": "0.014",
    "T": 1568014460893
  }
}
```

### Composite Index Stream (COIN-M only)

- **Description**: Composite index information for COIN-M futures. Provides index constituent data.
- **Stream Name**: `<symbol>@compositeIndex`
- **Update Speed**: Real-time
- **Response Fields**:

| Field | Type   | Description                              |
|-------|--------|------------------------------------------|
| e     | string | Event type (`compositeIndex`)            |
| E     | int    | Event time (ms)                          |
| s     | string | Symbol                                   |
| p     | string | Index price                              |
| i     | array  | Index constituents (array of objects)    |

- **Sample Response**:

```json
{
  "e": "compositeIndex",
  "E": 1591261236000,
  "s": "DEFIUSDT",
  "p": "1.2345",
  "i": [
    {
      "b": "binance",
      "w": "0.50000000"
    },
    {
      "b": "okex",
      "w": "0.30000000"
    },
    {
      "b": "huobi",
      "w": "0.20000000"
    }
  ]
}
```

## Constraints & Limits

| Constraint                  | Limit                              |
|-----------------------------|------------------------------------|
| Max streams per connection  | 1024                               |
| Max incoming messages/sec   | 10 per connection                  |
| Connection validity         | 24 hours                           |
| Ping interval               | Every 3 minutes                    |
| Pong timeout                | 10 minutes                         |
| IP rate limit (market data) | 2400 requests per 5 minutes per IP |
| Subscription method         | Static (via URL) or dynamic (via subscribe/unsubscribe messages) |

## Notes

### Symbol Format Differences

- **USDT-M**: Uses standard spot-like format (e.g., `btcusdt`, `ethusdt`)
- **COIN-M**:
  - Perpetual: `<base><quote>_PERP` (e.g., `btcusd_PERP`)
  - Quarterly: `<base><quote>_YYMMDD` (e.g., `btcusd_201225`)
  - Index/pair streams use just the pair (e.g., `btcusd`)

### Funding Rate Information

- Funding rate (`r`) and next funding time (`T`) are only available for **perpetual** contracts
- Delivery/quarterly contracts show empty string `""` for funding rate and `0` for next funding time
- USDT-M mark price stream includes `ap` (mark price moving average) as of 2026-03-16

### RPI (Retail Price Improvement) Orders

- RPI orders are **excluded** from book ticker and depth streams
- RPI orders are **included** in aggregate trade streams (aggregated into `q` field)
- The `nq` field shows quantity excluding RPI trades

### COIN-M Specific Streams

COIN-M futures has additional streams not available in USDT-M:
- **Index Price Stream**: `<pair>@indexPrice`
- **Index Kline Stream**: `<pair>@indexPriceKline_<interval>`
- **Mark Price Kline Stream**: `<symbol>@markPriceKline_<interval>`
- **Mark Price of All Symbols of a Pair**: `<pair>@markPrice` (instead of global `!markPrice@arr`)

### Local Order Book Management

- Use `U` (first update ID), `u` (final update ID), and `pu` (previous final update ID) to maintain correct order book state
- Fetch initial order book snapshot via REST API `/fapi/v1/depth` or `/dapi/v1/depth`
- Apply updates in sequence, ensuring no gaps in update IDs
- See official docs: [How To Manage A Local Order Book Correctly](https://developers.binance.com/docs/derivatives/usds-margined-futures/websocket-market-streams/How-to-manage-a-local-order-book-correctly)

### Chinese Character Support

- As of 2025-10-09, symbols may contain Chinese characters (e.g., `测试USDT`)
- When placing orders via REST, Chinese symbols must be URL-encoded (UTF-8 percent-encoding)
- WebSocket push messages may also contain Chinese symbols; ensure proper handling
