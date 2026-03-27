# Bitget API Documentation

Bitget provides APIs for Spot, Margin, and Futures trading. This document covers the main trading endpoints.

## Base URLs

| Type | URL |
|------|-----|
| REST API | `https://api.bitget.com` |
| WebSocket (Public) | `wss://ws.bitget.com/v2/ws/public` |
| WebSocket (Private) | `wss://ws.bitget.com/v2/ws/private` |

## Authentication

### Required Headers

```
ACCESS-KEY: API key
ACCESS-SIGN: Base64 encoded signature
ACCESS-TIMESTAMP: Unix timestamp in milliseconds
ACCESS-PASSPHRASE: API key password
Content-Type: application/json
locale: en-US
```

### Signature Generation

```
signature = Base64(HMAC-SHA256(secretKey, timestamp + method + requestPath + "?" + queryString + body))
```

If queryString is empty:
```
signature = Base64(HMAC-SHA256(secretKey, timestamp + method + requestPath + body))
```

---

## Spot API

### Market Data

#### Get Coin Info

```
GET /api/v2/spot/public/coins
```

Query: `coin` (optional) - Coin name

Response:
```json
{
  "data": [{
    "coinId": "1",
    "coin": "BTC",
    "transfer": "true",
    "chains": [{
      "chain": "BTC",
      "withdrawFee": "0.005",
      "minDepositAmount": "0.001",
      "minWithdrawAmount": "0.001"
    }]
  }]
}
```

#### Get Ticker Information

```
GET /api/v2/spot/market/tickers
```

Query: `symbol` (optional) - Trading pair (e.g., BTCUSDT)

Response:
```json
{
  "data": [{
    "symbol": "BTCUSDT",
    "high24h": "37775.65",
    "open": "35134.2",
    "low24h": "34413.1",
    "lastPr": "34413.1",
    "quoteVolume": "0",
    "baseVolume": "0",
    "bidPr": "0",
    "askPr": "0",
    "bidSz": "0.0663",
    "askSz": "0.0119",
    "ts": "1625125755277",
    "change24h": "0.00069"
  }]
}
```

#### Get OrderBook Depth

```
GET /api/v2/spot/market/orderbook
```

Query:
- `symbol` (required) - Trading pair
- `type` (optional) - step0 to step5 (depth levels)
- `limit` (optional) - Number of results, max 150

Response:
```json
{
  "data": {
    "asks": [["34567.15", "0.0131"], ["34567.25", "0.0144"]],
    "bids": [["34567", "0.2917"], ["34566.85", "0.0145"]],
    "ts": "1698303884584"
  }
}
```

#### Get Candlestick Data

```
GET /api/v2/spot/market/candles
```

Query:
- `symbol` (required) - Trading pair
- `granularity` (required) - Time interval: 1m, 3m, 5m, 15m, 30m, 1h, 4h, 6h, 12h, 1day, 3day, 1week, 1M
- `startTime` (optional) - Unix ms
- `endTime` (optional) - Unix ms
- `limit` (optional) - Max 1000

Response:
```json
{
  "data": [[
    "1656604800000",  // timestamp
    "37834.5",        // open
    "37849.5",        // high
    "37773.5",        // low
    "37773.5",        // close
    "428.3462",       // volume base
    "16198849.1079",  // volume usdt
    "16198849.1079"   // volume quote
  ]]
}
```

---

### Spot Trading

#### Place Order

```
POST /api/v2/spot/trade/place-order
```

Body:
```json
{
  "symbol": "BTCUSDT",
  "side": "buy",        // "buy" or "sell"
  "orderType": "limit", // "limit" or "market"
  "force": "gtc",       // "gtc", "post_only", "fok", "ioc"
  "price": "23222.5",
  "size": "1",
  "clientOid": "121211212122",
  "tpslType": "normal"  // "normal" or "tpsl"
}
```

Response:
```json
{
  "data": {
    "orderId": "1001",
    "clientOid": "121211212122"
  }
}
```

#### Cancel Order

```
POST /api/v2/spot/trade/cancel-order
```

Body:
```json
{
  "orderId": "1001",
  "symbol": "BTCUSDT"
}
```

#### Get Order Info

```
GET /api/v2/spot/trade/orderinfo
```

Query:
- `orderId` or `clientOid`
- `symbol`

#### Get Current Orders (Unfilled)

```
GET /api/v2/spot/trade/unfilled-orders
```

Query: `symbol`

#### Get History Orders

```
GET /api/v2/spot/trade/history-orders
```

Query:
- `symbol`
- `startTime`, `endTime`
- `limit`

#### Get Fills

```
GET /api/v2/spot/trade/fills
```

Query:
- `orderId` (optional)
- `startTime`, `endTime`

---

### Spot Account

#### Get Account Information

```
GET /api/v2/spot/account/info
```

Response:
```json
{
  "data": {
    "userId": "**********",
    "authorities": ["stow", "smow", "wtow"],
    "traderType": "trader"
  }
}
```

#### Get Account Assets

```
GET /api/v2/spot/account/assets
```

Query: `coin` (optional)

Response:
```json
{
  "data": [{
    "coin": "BTC",
    "available": "1.5",
    "locked": "0.5",
    "frozen": "0"
  }]
}
```

---

## Margin API

### Common

#### Get Support Currencies

```
GET /api/v2/margin/currencies
```

Response:
```json
{
  "data": [{
    "symbol": "ETHUSDT",
    "baseCoin": "ETH",
    "quoteCoin": "USDT",
    "maxCrossedLeverage": "3",
    "maxIsolatedLeverage": "10",
    "takerFeeRate": "0.001",
    "makerFeeRate": "0.001",
    "isCrossBorrowable": true,
    "isIsolatedBaseBorrowable": true,
    "isIsolatedQuoteBorrowable": true
  }]
}
```

#### Get Leverage Interest Rate

```
GET /api/v2/margin/interest-rate
```

Query: `coin` (optional)

---

## Futures API

### Product Types

| Type | Description |
|------|-------------|
| `USDT-FUTURES` | USDT-M Futures |
| `COIN-FUTURES` | Coin-M Futures |
| `USDC-FUTURES` | USDC-M Futures |

---

### Market Data

#### Get Ticker

```
GET /api/v2/mix/market/ticker
```

Query:
- `symbol` (required) - Trading pair
- `productType` (required) - Product type

Response:
```json
{
  "data": [{
    "symbol": "BTCUSDT",
    "lastPr": "1829.3",
    "askPr": "1829.8",
    "bidPr": "1829.3",
    "bidSz": "0.054",
    "askSz": "0.785",
    "high24h": "0",
    "low24h": "0",
    "ts": "1695794098184",
    "change24h": "0",
    "baseVolume": "0",
    "quoteVolume": "0",
    "indexPrice": "1822.15",
    "fundingRate": "0",
    "holdingAmount": "9488.49",
    "markPrice": "1829"
  }]
}
```

#### Get All Tickers

```
GET /api/v2/mix/market/tickers
```

Query: `productType` (required)

#### Get Contract Config

```
GET /api/v2/mix/market/contracts
```

Query:
- `symbol` (optional)
- `productType` (required)

Response:
```json
{
  "data": [{
    "symbol": "BTCUSDT",
    "baseCoin": "BTC",
    "quoteCoin": "USDT",
    "makerFeeRate": "0.0004",
    "takerFeeRate": "0.0006",
    "minTradeNum": "0.01",
    "priceEndStep": "1",
    "volumePlace": "2",
    "pricePlace": "1",
    "minLever": "1",
    "maxLever": "125",
    "symbolType": "perpetual"
  }]
}
```

#### Get Candlestick Data

```
GET /api/v2/mix/market/candles
```

Query:
- `symbol` (required)
- `productType` (required)
- `granularity` (required) - 1m, 3m, 5m, 15m, 30m, 1h, 4h, 6h, 12h, 1d
- `startTime`, `endTime`
- `limit`

#### Get OrderBook Depth

```
GET /api/v2/mix/market/depth
```

Query:
- `symbol` (required)
- `productType` (required)
- `type` (optional) - step0 to step5
- `limit` (optional)

#### Get Open Interest

```
GET /api/v2/mix/market/open-interest
```

Query:
- `symbol` (required)
- `productType` (required)

#### Get Funding Rate

```
GET /api/v2/mix/market/current-funding-rate
```

Query:
- `symbol` (required)
- `productType` (required)

#### Get Historical Funding Rates

```
GET /api/v2/mix/market/history-funding-rate
```

Query:
- `symbol` (required)
- `productType` (required)
- `startTime`, `endTime`

---

### Futures Trading

#### Place Order

```
POST /api/v2/mix/order/place-order
```

Body:
```json
{
  "symbol": "ETHUSDT",
  "productType": "USDT-FUTURES",
  "marginMode": "isolated",  // "isolated" or "crossed"
  "marginCoin": "USDT",
  "size": "0.1",
  "price": "2000",
  "side": "buy",        // "buy" or "sell"
  "tradeSide": "open",  // "open" or "close" (hedge mode only)
  "orderType": "limit", // "limit" or "market"
  "force": "gtc",       // "gtc", "ioc", "fok", "post_only"
  "clientOid": "121211212122",
  "reduceOnly": "NO"    // "YES" or "NO"
}
```

Response:
```json
{
  "data": {
    "orderId": "121211212122",
    "clientOid": "121211212122"
  }
}
```

#### Cancel Order

```
POST /api/v2/mix/order/cancel-order
```

Body:
```json
{
  "orderId": "121211212122",
  "symbol": "ETHUSDT",
  "productType": "USDT-FUTURES"
}
```

#### Modify Order

```
POST /api/v2/mix/order/modify-order
```

Body:
```json
{
  "orderId": "121211212122",
  "symbol": "ETHUSDT",
  "productType": "USDT-FUTURES",
  "newPrice": "2100",
  "newSize": "0.2"
}
```

#### Get Order Detail

```
GET /api/v2/mix/order/detail
```

Query:
- `orderId`
- `symbol`

#### Get Pending Orders

```
GET /api/v2/mix/order/orders-pending
```

Query:
- `symbol`
- `productType`

#### Get History Orders

```
GET /api/v2/mix/order/orders-history
```

Query:
- `symbol`
- `productType`
- `startTime`, `endTime`
- `limit`

#### Get Order Fills

```
GET /api/v2/mix/order/fills
```

Query:
- `symbol`
- `productType`
- `orderId` (optional)

---

### Futures Account

#### Get Single Account

```
GET /api/v2/mix/account/account
```

Query:
- `symbol`
- `productType`
- `marginCoin`

Response:
```json
{
  "data": {
    "marginCoin": "USDT",
    "locked": "0",
    "available": "82755.86971252",
    "accountEquity": "-4767.74182959",
    "usdtEquity": "-4767.74182959",
    "btcEquity": "-0.05309929",
    "crossedRiskRate": "0.0028",
    "crossedMarginLeverage": 14,
    "marginMode": "crossed",
    "posMode": "hedge_mode",
    "unrealizedPL": "-267.63713333"
  }
}
```

#### Get Account List

```
GET /api/v2/mix/account/accounts
```

Query: `productType`

#### Change Leverage

```
POST /api/v2/mix/account/leverage
```

Body:
```json
{
  "symbol": "BTCUSDT",
  "productType": "USDT-FUTURES",
  "leverage": "20",
  "marginCoin": "USDT"
}
```

#### Change Margin Mode

```
POST /api/v2/mix/account/set-margin-mode
```

Body:
```json
{
  "symbol": "BTCUSDT",
  "productType": "USDT-FUTURES",
  "marginMode": "crossed"  // "crossed" or "isolated"
}
```

#### Change Position Mode

```
POST /api/v2/mix/account/set-position-mode
```

Body:
```json
{
  "productType": "USDT-FUTURES",
  "posMode": "hedge_mode"  // "hedge_mode" or "one_way_mode"
}
```

#### Adjust Position Margin

```
POST /api/v2/mix/account/adjust-position-margin
```

Body:
```json
{
  "symbol": "BTCUSDT",
  "productType": "USDT-FUTURES",
  "marginCoin": "USDT",
  "amount": "100",       // Positive = add, Negative = reduce
  "type": "add"         // "add" or "reduce"
}
```

---

### Futures Position

#### Get Single Position

```
GET /api/v2/mix/position/single-position
```

Query:
- `symbol`
- `productType`

Response:
```json
{
  "data": {
    "symbol": "ETHUSDT",
    "marginCoin": "USDT",
    "holdSide": "long",    // "long" or "short"
    "positionAmt": "0.5",
    "availableAmt": "0.5",
    "frozenAmt": "0",
    "openPriceAvg": "2000",
    "margin": "100",
    "unrealizedPL": "50",
    "leverage": "10",
    "markPrice": "2100"
  }
}
```

#### Get All Positions

```
GET /api/v2/mix/position/all-position
```

Query: `productType`

#### Get Position Tier

```
GET /api/v2/mix/market/query-position-lever
```

Query:
- `symbol`
- `productType`

Response:
```json
{
  "data": [{
    "symbol": "BTCUSDT",
    "level": "1",
    "startUnit": "0",
    "endUnit": "50000",
    "leverage": "125",
    "keepMarginRate": "0.004"
  }]
}
```

---

## WebSocket API

### Connection

**Limits:**
- 300 connections/IP/5min
- 100 connections/IP
- 240 subscriptions/hour/connection
- 1000 channel subscriptions/connection

**Keepalive:** Send "ping" every 30s, expect "pong"

### Authentication (Private WS)

```json
{
  "op": "login",
  "args": [{
    "apiKey": "xxx",
    "passphrase": "xxx",
    "timestamp": "1538054050000",
    "sign": "signature"
  }]
}
```

Sign = Base64(HMAC-SHA256(secretKey, timestamp + "GET" + "/user/verify"))

### Subscribe

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

### Unsubscribe

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

---

### WebSocket Channels

#### Spot Channels

| Channel | Description |
|---------|-------------|
| `ticker` | Price ticker (200-300ms) |
| `candle5m`, `candle1m`, etc. | K-line/candlestick |
| `trade` | Real-time trades |
| `depth` | Orderbook depth |
| `depth5`, `depth10`, etc. | Aggregated depth |

#### Futures Channels

| Channel | Description |
|---------|-------------|
| `ticker` | Price ticker (300-400ms) |
| `candle1m`, `candle5m`, etc. | K-line/candlestick |
| `depth` | Orderbook depth |
| `trade` | Public trade channel |

---

### Instrument Types

- `SPOT` - Spot
- `USDT-FUTURES` - USDT-M Futures
- `COIN-FUTURES` - Coin-M Futures
- `USDC-FUTURES` - USDC-M Futures
