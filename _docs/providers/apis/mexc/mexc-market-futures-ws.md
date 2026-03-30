# MEXC WebSocket Market Documentation (Futures)

## Reference
- **Official API Docs**:
  - https://www.mexc.com/api-docs/futures/websocket-api
  - https://mexcdevelop.github.io/apidocs/contract_v1/en/ (legacy)
- **WebSocket Base URL**: `wss://contract.mexc.com/edge`

## Protocol Overview
- **Protocol type**: Raw WebSocket (no framing wrapper)
- **Connection limits**: Not explicitly documented; recommend 1–3 connections per IP
- **Keep-alive / ping-pong mechanism**:
  - Client sends: `{"method": "ping"}`
  - Server responds: `{"channel": "pong", "data": <timestamp>}`
  - If no ping received within **60 seconds**, the server disconnects the client
  - Recommended: send ping every **10–20 seconds**
- **Reconnection guidelines**: On disconnect, reconnect and re-subscribe to all channels
- **Authentication**: Not required for public market streams; required only for private channels (orders, positions, assets, etc.)
- **Compression**: Responses are gzip-compressed by default. Add `"gzip": false` to subscription requests to receive plaintext

## Message Format

### Request Format
```json
{
  "method": "sub.<channel>",
  "param": {
    "symbol": "BTC_USDT"
  },
  "gzip": false
}
```

### Response Format
```json
{
  "channel": "push.<channel>",
  "data": { ... },
  "symbol": "BTC_USDT",
  "ts": 1587442022003
}
```

### Error Format
```json
{
  "channel": "rs.error",
  "data": "error message",
  "ts": 1587442022003
}
```

## WebSocket Endpoints

### Trade Stream (Price)
- **Description**: Real-time trade/execution data for a specific futures contract. Pushed whenever trades occur. Aggregation (compression) is enabled by default; set `compress: false` to disable.
- **Method/Topic**: `sub.deal` / `push.deal`
- **Subscription Request**:
```json
{
  "method": "sub.deal",
  "param": {
    "symbol": "BTC_USDT"
  }
}
```
- **Unsubscribe Request**:
```json
{
  "method": "unsub.deal",
  "param": {
    "symbol": "BTC_USDT"
  }
}
```
- **Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| p | decimal | Trade price |
| v | decimal | Trade quantity (contracts) |
| T | int | Trade side: 1 = buy, 2 = sell |
| O | int | Open/close flag: 1 = new position, 2 = reduce position, 3 = position unchanged |
| M | int | Self-trade: 1 = yes, 2 = no |
| i | int | Transaction ID |
| t | long | Trade timestamp (ms) |

- **Sample Response**:
```json
{
  "symbol": "BTC_USDT",
  "data": [
    {
      "p": 115309.8,
      "v": 55,
      "T": 2,
      "O": 3,
      "M": 1,
      "t": 1755487578276,
      "i": 13064218826
    },
    {
      "p": 115309.8,
      "v": 11,
      "T": 1,
      "O": 3,
      "M": 1,
      "t": 1755487578275,
      "i": 13064218827
    }
  ],
  "channel": "push.deal",
  "ts": 1755487578276
}
```

### Ticker Stream (24h)
- **Description**: 24-hour ticker statistics. Two variants: `sub.tickers` (all perpetual contracts, pushed every 1s) and `sub.ticker` (single contract, pushed every 1s when trades occur).
- **Method/Topic**: `sub.ticker` / `push.ticker` (single) or `sub.tickers` / `push.tickers` (all)
- **Subscription Request (single)**:
```json
{
  "method": "sub.ticker",
  "param": {
    "symbol": "BTC_USDT"
  }
}
```
- **Subscription Request (all)**:
```json
{
  "method": "sub.tickers",
  "param": {}
}
```
- **Response Fields (single ticker)**:

| Field | Type | Description |
|-------|------|-------------|
| symbol | string | Contract name |
| timestamp | long | Trade timestamp (ms) |
| lastPrice | decimal | Last traded price |
| bid1 | decimal | Best bid price |
| ask1 | decimal | Best ask price |
| holdVol | decimal | Open interest (total held volume) |
| fundingRate | decimal | Current funding rate |
| riseFallRate | decimal | 24h price change rate |
| riseFallValue | decimal | 24h price change amount |
| volume24 | decimal | 24h volume (in contracts) |
| amount24 | decimal | 24h turnover (in currency) |
| fairPrice | decimal | Fair price |
| indexPrice | decimal | Index price |
| maxBidPrice | decimal | Max buy price |
| minAskPrice | decimal | Min sell price |
| lower24Price | decimal | 24h low |
| high24Price | decimal | 24h high |
| contractId | int | Contract ID |

- **Sample Response (single ticker)**:
```json
{
  "channel": "push.ticker",
  "data": {
    "ask1": 6866.5,
    "bid1": 6865,
    "contractId": 1,
    "fairPrice": 6867.4,
    "fundingRate": 0.0008,
    "high24Price": 7223.5,
    "indexPrice": 6861.6,
    "lastPrice": 6865.5,
    "lower24Price": 6756,
    "maxBidPrice": 7073.42,
    "minAskPrice": 6661.37,
    "riseFallRate": -0.0424,
    "riseFallValue": -304.5,
    "symbol": "BTC_USDT",
    "timestamp": 1587442022003,
    "holdVol": 2284742,
    "volume24": 164586129
  },
  "symbol": "BTC_USDT"
}
```

### Depth Stream (Order Book)
- **Description**: Real-time order book updates for a specific contract. Pushed every **200ms** after subscription.
- **Method/Topic**: `sub.depth` / `push.depth`
- **Available Variants**:
  - `sub.depth` — full depth updates
  - `sub.depth.step` — depth aggregated by notional step (e.g., `step: "10"` groups prices into buckets of 10)
- **Subscription Request**:
```json
{
  "method": "sub.depth",
  "param": {
    "symbol": "BTC_USDT"
  }
}
```
- **Subscription Request (with step)**:
```json
{
  "method": "sub.depth.step",
  "param": {
    "symbol": "BTC_USDT",
    "step": "10"
  }
}
```
- **Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| asks | `List<Numeric[]>` | Ask depth levels |
| bids | `List<Numeric[]>` | Bid depth levels |
| version | long | Version number for ordering |
| askMarketLevelPrice | decimal | Highest willing ask (step variant only) |
| bidMarketLevelPrice | decimal | Highest willing bid (step variant only) |
| ct | long | Timestamp (step variant only) |

Each depth level entry: `[price, contractCount, orderCount]`
- `price` — price level
- `contractCount` — total contracts at this price
- `orderCount` — number of orders at this price

- **Sample Response**:
```json
{
  "channel": "push.depth",
  "data": {
    "asks": [[6859.5, 3251, 1]],
    "bids": [],
    "version": 96801927
  },
  "symbol": "BTC_USDT",
  "ts": 1587442022003
}
```
- **Snapshot vs incremental updates**: The depth stream sends incremental updates. Each message contains only changed levels with a `version` field for ordering. Clients should maintain a local order book and apply updates in version order.

### Kline/Candlestick Stream
- **Description**: Real-time K-line (candlestick) updates. Pushed whenever a candle updates.
- **Method/Topic**: `sub.kline` / `push.kline`
- **Available Intervals**: `Min1`, `Min5`, `Min15`, `Min30`, `Min60`, `Hour4`, `Hour8`, `Day1`, `Week1`, `Month1`
- **Subscription Request**:
```json
{
  "method": "sub.kline",
  "param": {
    "symbol": "BTC_USDT",
    "interval": "Min60"
  },
  "gzip": false
}
```
- **Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| symbol | string | Contract name |
| interval | string | Candle interval |
| o | decimal | Open price |
| c | decimal | Close price |
| h | decimal | High price |
| l | decimal | Low price |
| a | decimal | Total traded amount |
| q | decimal | Total traded volume |
| v | decimal | Total volume |
| ro | decimal | Real open price |
| rc | decimal | Real close price |
| rh | decimal | Real high price |
| rl | decimal | Real low price |
| t | long | Window start timestamp (seconds) |

- **Sample Response**:
```json
{
  "channel": "push.kline",
  "data": {
    "a": 233.740269343644737245,
    "c": 6885,
    "h": 6910.5,
    "interval": "Min60",
    "l": 6885,
    "o": 6894.5,
    "q": 1611754,
    "symbol": "BTC_USDT",
    "t": 1587448800
  },
  "symbol": "BTC_USDT"
}
```

### Mark Price Stream
- **Description**: MEXC does not provide a dedicated "mark price" stream. The closest equivalents are:
  - **Fair price** (`sub.fair.price`) — used for liquidation calculations
  - **Index price** (`sub.index.price`) — composite index from multiple exchanges
- **Method/Topic**: `sub.fair.price` / `push.fair.price`
- **Subscription Request**:
```json
{
  "method": "sub.fair.price",
  "param": {
    "symbol": "BTC_USDT"
  },
  "gzip": false
}
```
- **Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| symbol | string | Contract name |
| price | decimal | Fair price |

- **Sample Response**:
```json
{
  "channel": "push.fair.price",
  "data": {
    "price": 6867.4,
    "symbol": "BTC_USDT"
  },
  "symbol": "BTC_USDT",
  "ts": 1587442022003
}
```

### Funding Rate Stream
- **Description**: Real-time funding rate updates for perpetual contracts. Pushed whenever the funding rate changes.
- **Method/Topic**: `sub.funding.rate` / `push.funding.rate`
- **Subscription Request**:
```json
{
  "method": "sub.funding.rate",
  "param": {
    "symbol": "BTC_USDT"
  },
  "gzip": false
}
```
- **Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| symbol | string | Contract name |
| rate | decimal | Current funding rate |
| nextSettleTime | long | Next settlement timestamp (ms) |

- **Sample Response**:
```json
{
  "channel": "push.funding.rate",
  "data": {
    "rate": 0.001,
    "symbol": "BTC_USDT"
  },
  "symbol": "BTC_USDT",
  "ts": 1587442022003
}
```

### Index Price Stream
- **Description**: Real-time index price updates. Pushed whenever the index price changes. Works for both linear (USDT-settled) and inverse contracts — use the same symbol format (e.g., `BTC_USDT` for both).
- **Method/Topic**: `sub.index.price` / `push.index.price`
- **Subscription Request**:
```json
{
  "method": "sub.index.price",
  "param": {
    "symbol": "BTC_USDT"
  },
  "gzip": false
}
```
- **Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| symbol | string | Contract name |
| price | decimal | Index price |

- **Sample Response**:
```json
{
  "channel": "push.index.price",
  "data": {
    "price": 6861.6,
    "symbol": "BTC_USDT"
  },
  "symbol": "BTC_USDT",
  "ts": 1587442022003
}
```

### Open Interest Stream
- **Description**: MEXC does not provide a standalone open interest stream. Open interest (`holdVol`) is included in the **single ticker** stream (`sub.ticker`) and updated every 1s when trades occur.
- **Workaround**: Subscribe to `sub.ticker` and extract `holdVol` from each update.

## Additional Public Channels

### Contract Data Stream
- **Description**: Contract specification updates (fees, limits, leverage, etc.). Pushed whenever contract data changes.
- **Method/Topic**: `sub.contract` / `push.contract`
- **Subscription Request**:
```json
{
  "method": "sub.contract"
}
```
- **Key Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| symbol | string | Contract name |
| contractSize | decimal | Contract size |
| minVol / maxVol | decimal | Min/max order size |
| priceUnit | decimal | Min price tick |
| volUnit | decimal | Min volume step |
| makerFeeRate / takerFeeRate | decimal | Fee rates |
| maxLeverage / minLeverage | int | Leverage limits |
| priceScale / volScale / amountScale | int | Decimal precision |
| state | int | 0=enabled, 1=delivering, 2=delivered, 3=delisted, 4=paused |
| futureType | int | 1=PERPETUAL, 2=DAILY |

## Constraints & Limits
- **Ping timeout**: Connection is closed if no ping is received within **60 seconds**
- **Recommended ping interval**: Every **10–20 seconds**
- **Depth update frequency**: Every **200ms**
- **Ticker update frequency**: Every **1s**
- **Compression**: Responses are gzip-compressed by default; use `"gzip": false` for plaintext
- **Subscription limits**: Not explicitly documented; avoid subscribing to excessive symbols on a single connection
- **No authentication required** for public market streams

## Notes
- **Symbol format**: Futures contracts use underscore separator, e.g., `BTC_USDT`, `ETH_USDT`
- **Differences from spot WebSocket**:
  - Different base URL (`wss://contract.mexc.com/edge` vs spot endpoint)
  - Futures include `holdVol` (open interest), `fundingRate`, `fairPrice`, `indexPrice` in ticker responses
  - Trade stream includes `O` (open/close flag) field specific to futures positions
  - Depth entries include order count as third element: `[price, contracts, orders]`
- **Gzip handling**: Most responses are gzip-compressed. Clients must decompress unless `gzip: false` is specified in the subscription request
- **Private channels**: Require HMAC-SHA256 authentication via `login` method with `apiKey`, `signature`, and `reqTime` parameters
- **Subscription filtering**: After login, use `personal.filter` to selectively receive private data (orders, positions, assets, etc.)
