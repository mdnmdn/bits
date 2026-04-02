# Binance Spot Market Order REST API Documentation

## Reference

- **Official API Docs**: https://developers.binance.com/docs/binance-spot-api-docs/rest-api/trading-endpoints
- **Base URL**: `https://api.binance.com`

## Authentication

All order endpoints are `SIGNED` or `USER_DATA` endpoints requiring authentication.

### Required Headers/Params

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `X-MBX-APIKEY` | STRING | YES | API key sent in request header |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |
| `signature` | STRING | YES | HMAC-SHA256 signature of the query string + body |
| `recvWindow` | LONG | NO | Request validity window in ms (default: 5000, max: 60000) |

### Signature Method (HMAC-SHA256)

1. Construct the payload: `param1=value1&param2=value2...`
2. Percent-encode any non-ASCII characters before signing
3. Compute HMAC-SHA256 using the `secretKey` as the signing key
4. Append the hex-encoded signature as the `signature` parameter

## Place Order

- **Description**: Create a new order (limit, market, stop-loss, take-profit, etc.)
- **Endpoint**: `POST /api/v3/order`
- **Weight**: 1
- **Security Type**: `SIGNED`
- **Docs**: https://developers.binance.com/docs/binance-spot-api-docs/rest-api/trading-endpoints#new-order-trade

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair symbol |
| `side` | ENUM | YES | `BUY` or `SELL` |
| `type` | ENUM | YES | Order type (see Order Types section) |
| `timeInForce` | ENUM | NO | `GTC`, `IOC`, `FOK` (required for LIMIT orders) |
| `quantity` | DECIMAL | NO | Amount of base asset |
| `quoteOrderQty` | DECIMAL | NO | Amount of quote asset to spend/receive |
| `price` | DECIMAL | NO | Limit price |
| `newClientOrderId` | STRING | NO | Unique client order ID (auto-generated if not sent) |
| `stopPrice` | DECIMAL | NO | Trigger price for stop orders |
| `trailingDelta` | LONG | NO | Delta for trailing stop orders |
| `icebergQty` | DECIMAL | NO | Visible quantity for iceberg orders |
| `newOrderRespType` | ENUM | NO | `ACK`, `RESULT`, or `FULL` |
| `selfTradePreventionMode` | ENUM | NO | STP mode (see STP section) |
| `pegPriceType` | ENUM | NO | `PRIMARY_PEG` or `MARKET_PEG` |
| `pegOffsetValue` | INT | NO | Price level offset (max: 100) |
| `recvWindow` | DECIMAL | NO | Max 60000, supports 3 decimal places |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |

### Type-Specific Required Parameters

| Type | Required Parameters |
|------|---------------------|
| `LIMIT` | `timeInForce`, `quantity`, `price` |
| `MARKET` | `quantity` or `quoteOrderQty` |
| `STOP_LOSS` | `quantity`, (`stopPrice` or `trailingDelta`) |
| `STOP_LOSS_LIMIT` | `timeInForce`, `quantity`, `price`, (`stopPrice` or `trailingDelta`) |
| `TAKE_PROFIT` | `quantity`, (`stopPrice` or `trailingDelta`) |
| `TAKE_PROFIT_LIMIT` | `timeInForce`, `quantity`, `price`, (`stopPrice` or `trailingDelta`) |
| `LIMIT_MAKER` | `quantity`, `price` |

### Response Fields (ACK)

| Field | Type | Description |
|-------|------|-------------|
| `symbol` | STRING | Trading pair |
| `orderId` | LONG | Order ID |
| `orderListId` | LONG | OCO list ID (-1 if not OCO) |
| `clientOrderId` | STRING | Client order ID |
| `transactTime` | LONG | Transaction timestamp |

### Response Fields (RESULT)

Includes all ACK fields plus:

| Field | Type | Description |
|-------|------|-------------|
| `price` | STRING | Order price |
| `origQty` | STRING | Original quantity |
| `executedQty` | STRING | Filled quantity |
| `cummulativeQuoteQty` | STRING | Cumulative quote quantity |
| `status` | STRING | Order status |
| `timeInForce` | STRING | Time in force |
| `type` | STRING | Order type |
| `side` | STRING | Order side |
| `workingTime` | LONG | Working timestamp |
| `selfTradePreventionMode` | STRING | STP mode |

### Response Fields (FULL)

Includes all RESULT fields plus:

| Field | Type | Description |
|-------|------|-------------|
| `fills` | ARRAY | Array of fill objects |

**Fill Object Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `price` | STRING | Fill price |
| `qty` | STRING | Fill quantity |
| `commission` | STRING | Commission amount |
| `commissionAsset` | STRING | Commission asset |
| `tradeId` | LONG | Trade ID |

### Sample Request

```bash
curl -X POST "https://api.binance.com/api/v3/order" \
  -H "X-MBX-APIKEY: your_api_key" \
  -d "symbol=BTCUSDT&side=BUY&type=LIMIT&timeInForce=GTC&quantity=0.001&price=50000&recvWindow=5000&timestamp=1644489390087&signature=<hmac_sha256_sig>"
```

### Sample Response (RESULT)

```json
{
  "symbol": "BTCUSDT",
  "orderId": 28,
  "orderListId": -1,
  "clientOrderId": "6gCrw2kRUAF9CvJDGP16IP",
  "transactTime": 1507725176595,
  "price": "50000.00000000",
  "origQty": "0.00100000",
  "executedQty": "0.00100000",
  "cummulativeQuoteQty": "50.00000000",
  "status": "FILLED",
  "timeInForce": "GTC",
  "type": "LIMIT",
  "side": "BUY",
  "workingTime": 1507725176595,
  "selfTradePreventionMode": "NONE"
}
```

## Test New Order

- **Description**: Creates and validates a new order without sending it to the matching engine.
- **Endpoint**: `POST /api/v3/order/test`
- **Weight**: 1 (without `computeCommissionRates`), 20 (with `computeCommissionRates`)
- **Security Type**: `SIGNED`

### Parameters

Accepts all parameters from `POST /api/v3/order` plus:

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `computeCommissionRates` | BOOLEAN | NO | Default: `false`. If `true`, returns estimated commission rates. |

### Sample Response (with computeCommissionRates=true)

```json
{
  "standardCommissionForOrder": {
    "maker": "0.00000112",
    "taker": "0.00000114"
  },
  "specialCommissionForOrder": {
    "maker": "0.05000000",
    "taker": "0.06000000"
  },
  "taxCommissionForOrder": {
    "maker": "0.00000112",
    "taker": "0.00000114"
  },
  "discount": {
    "enabledForAccount": true,
    "enabledForSymbol": true,
    "discountAsset": "BNB",
    "discount": "0.25000000"
  }
}
```

## Query Order

- **Description**: Check an order's status.
- **Endpoint**: `GET /api/v3/order`
- **Weight**: 1
- **Security Type**: `SIGNED`

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair |
| `orderId` | LONG | NO* | Order ID (*one required) |
| `origClientOrderId` | STRING | NO* | Original client order ID (*one required) |
| `recvWindow` | DECIMAL | NO | Max 60000 |
| `timestamp` | LONG | YES | Current timestamp |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `symbol` | STRING | Trading pair |
| `orderId` | LONG | Order ID |
| `orderListId` | LONG | OCO list ID |
| `clientOrderId` | STRING | Client order ID |
| `price` | STRING | Order price |
| `origQty` | STRING | Original quantity |
| `executedQty` | STRING | Filled quantity |
| `cummulativeQuoteQty` | STRING | Cumulative quote quantity |
| `status` | STRING | Order status |
| `timeInForce` | STRING | Time in force |
| `type` | STRING | Order type |
| `side` | STRING | Order side |
| `stopPrice` | STRING | Stop price |
| `icebergQty` | STRING | Iceberg quantity |
| `time` | LONG | Order creation time |
| `updateTime` | LONG | Last update time |
| `isWorking` | BOOLEAN | Is order in orderbook |
| `workingTime` | LONG | Working timestamp |
| `selfTradePreventionMode` | STRING | STP mode |

### Sample Response

```json
{
  "symbol": "LTCBTC",
  "orderId": 1,
  "orderListId": -1,
  "clientOrderId": "myOrder1",
  "price": "0.1",
  "origQty": "1.0",
  "executedQty": "0.0",
  "cummulativeQuoteQty": "0.0",
  "status": "NEW",
  "timeInForce": "GTC",
  "type": "LIMIT",
  "side": "BUY",
  "stopPrice": "0.0",
  "icebergQty": "0.0",
  "time": 1499827319559,
  "updateTime": 1499827319559,
  "isWorking": true,
  "workingTime": 1499827319559,
  "selfTradePreventionMode": "NONE"
}
```

## Cancel Order

- **Description**: Cancel an active order.
- **Endpoint**: `DELETE /api/v3/order`
- **Weight**: 1
- **Security Type**: `SIGNED`

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair |
| `orderId` | LONG | NO* | Order ID (*one required) |
| `origClientOrderId` | STRING | NO* | Original client order ID (*one required) |
| `newClientOrderId` | STRING | NO | Unique ID for this cancel |
| `cancelRestrictions` | ENUM | NO | `ONLY_NEW` or `ONLY_PARTIALLY_FILLED` |
| `recvWindow` | DECIMAL | NO | Max 60000 |
| `timestamp` | LONG | YES | Current timestamp |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `symbol` | STRING | Trading pair |
| `origClientOrderId` | STRING | Original client order ID |
| `orderId` | LONG | Order ID |
| `orderListId` | LONG | OCO list ID |
| `clientOrderId` | STRING | Client order ID for this cancel |
| `transactTime` | LONG | Transaction timestamp |
| `price` | STRING | Order price |
| `origQty` | STRING | Original quantity |
| `executedQty` | STRING | Filled quantity |
| `cummulativeQuoteQty` | STRING | Cumulative quote quantity |
| `status` | STRING | Order status |
| `timeInForce` | STRING | Time in force |
| `type` | STRING | Order type |
| `side` | STRING | Order side |
| `selfTradePreventionMode` | STRING | STP mode |

### Sample Response

```json
{
  "symbol": "LTCBTC",
  "origClientOrderId": "myOrder1",
  "orderId": 4,
  "orderListId": -1,
  "clientOrderId": "cancelMyOrder1",
  "transactTime": 1684804350068,
  "price": "2.00000000",
  "origQty": "1.00000000",
  "executedQty": "0.00000000",
  "cummulativeQuoteQty": "0.00000000",
  "status": "CANCELED",
  "timeInForce": "GTC",
  "type": "LIMIT",
  "side": "BUY",
  "selfTradePreventionMode": "NONE"
}
```

## Cancel All Open Orders

- **Description**: Cancel all open orders on a symbol.
- **Endpoint**: `DELETE /api/v3/openOrders`
- **Weight**: 1
- **Security Type**: `SIGNED`

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair |
| `recvWindow` | DECIMAL | NO | Max 60000 |
| `timestamp` | LONG | YES | Current timestamp |

### Response

Returns an array of canceled orders (same fields as Cancel Order response).

### Sample Response

```json
[
  {
    "symbol": "BTCUSDT",
    "origClientOrderId": "E6APeyTJvkMvLMYMqu1KQ4",
    "orderId": 11,
    "orderListId": -1,
    "clientOrderId": "pXLV6Hz6mprAcVYpVMTGgx",
    "transactTime": 1684804350068,
    "price": "0.089853",
    "origQty": "0.178622",
    "executedQty": "0.000000",
    "status": "CANCELED",
    "timeInForce": "GTC",
    "type": "LIMIT",
    "side": "BUY",
    "selfTradePreventionMode": "NONE"
  }
]
```

## Open Orders

- **Description**: Get all open orders on a symbol.
- **Endpoint**: `GET /api/v3/openOrders`
- **Weight**: 1
- **Security Type**: `SIGNED`

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | NO | Trading pair (omit for all symbols) |
| `recvWindow` | DECIMAL | NO | Max 60000 |
| `timestamp` | LONG | YES | Current timestamp |

### Response

Returns an array of open orders (same fields as Query Order response).

## All Orders

- **Description**: Get all orders (active, canceled, filled) with pagination.
- **Endpoint**: `GET /api/v3/allOrders`
- **Weight**: 10
- **Security Type**: `SIGNED`

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair |
| `orderId` | LONG | NO | Order ID to start from |
| `startTime` | LONG | NO | Start timestamp |
| `endTime` | LONG | NO | End timestamp |
| `limit` | INT | NO | Default 500, max 1000 |
| `recvWindow` | DECIMAL | NO | Max 60000 |
| `timestamp` | LONG | YES | Current timestamp |

### Response

Returns an array of orders (same fields as Query Order response).

## Order Types

| Type | Description | Required Parameters |
|------|-------------|---------------------|
| `LIMIT` | Basic limit order | `timeInForce`, `quantity`, `price` |
| `MARKET` | Market order at best available price | `quantity` or `quoteOrderQty` |
| `STOP_LOSS` | Market order triggered when stopPrice is reached | `quantity`, (`stopPrice` or `trailingDelta`) |
| `STOP_LOSS_LIMIT` | Limit order triggered when stopPrice is reached | `timeInForce`, `quantity`, `price`, (`stopPrice` or `trailingDelta`) |
| `TAKE_PROFIT` | Market order triggered when stopPrice is reached | `quantity`, (`stopPrice` or `trailingDelta`) |
| `TAKE_PROFIT_LIMIT` | Limit order triggered when stopPrice is reached | `timeInForce`, `quantity`, `price`, (`stopPrice` or `trailingDelta`) |
| `LIMIT_MAKER` | Limit order that will be rejected if it would immediately match (POST-ONLY) | `quantity`, `price` |

## Order Status

| Status | Description |
|--------|-------------|
| `NEW` | Order accepted by engine |
| `PENDING_NEW` | Pending until working order of order list is filled |
| `PARTIALLY_FILLED` | Part of order has been filled |
| `FILLED` | Order completed |
| `CANCELED` | Order canceled by user |
| `PENDING_CANCEL` | Currently unused |
| `REJECTED` | Order not accepted by engine |
| `EXPIRED` | Order canceled per type rules or by exchange |
| `EXPIRED_IN_MATCH` | Order expired due to STP |

## Time In Force

| Value | Description |
|-------|-------------|
| `GTC` | Good Til Canceled — remains on book until canceled |
| `IOC` | Immediate Or Cancel — fills as much as possible then expires |
| `FOK` | Fill or Kill — expires if full order cannot be filled |

## Self Trade Prevention Modes

| Mode | Description |
|------|-------------|
| `NONE` | No self-trade prevention |
| `EXPIRE_MAKER` | Expire the maker order |
| `EXPIRE_TAKER` | Expire the taker order |
| `EXPIRE_BOTH` | Expire both orders |
| `DECREMENT` | Decrement quantity to prevent self-trade |
| `TRANSFER` | Transfer to avoid self-trade |

## Constraints & Limits

### Rate Limits

| Scope | Limit | Window |
|-------|-------|--------|
| REQUEST_WEIGHT | 6,000 | 1 minute |
| ORDERS | 50 | 10 seconds |
| ORDERS | 160,000 | 1 day |
| RAW_REQUESTS | 300,000 | 5 minutes |

### Endpoint-Specific Weights

| Endpoint | Weight |
|----------|--------|
| `POST /api/v3/order` | 1 |
| `POST /api/v3/order/test` | 1 (20 with computeCommissionRates) |
| `DELETE /api/v3/order` | 1 |
| `DELETE /api/v3/openOrders` | 1 |
| `GET /api/v3/openOrders` | 1 |
| `GET /api/v3/order` | 1 |
| `GET /api/v3/allOrders` | 10 |

### Key Constraints

- **Iceberg Orders**: Any order with `icebergQty` MUST have `timeInForce` set to `GTC`
- **LIMIT_MAKER Orders**: Will be rejected if order would immediately match as taker (POST-ONLY)
- **Market Orders with quoteOrderQty**: `BUY` spends specified quote asset amount; `SELL` receives specified quote asset amount
- **Stop Order Trigger Rules**: Price above market: `STOP_LOSS BUY`, `TAKE_PROFIT SELL`. Price below market: `STOP_LOSS SELL`, `TAKE_PROFIT BUY`
- **Client Order ID**: Orders with same `newClientOrderId` accepted only when previous one is filled, otherwise rejected
- **Pegged Orders**: `PRIMARY_PEG` pegs to best price on same side; `MARKET_PEG` pegs to best price on opposite side. `pegOffsetValue` max is 100.
- **Rate limits are based on IP address, not API keys**. Repeated violations result in IP bans (HTTP 418) scaling from 2 minutes to 3 days.
- **Unfilled order count is tracked per account**. Filled orders decrement the unfilled order count.

## Notes

- All numeric values in responses are returned as strings to preserve precision.
- Symbols for order endpoints are **uppercase** (e.g., `BTCUSDT`), unlike WebSocket streams which use lowercase.
- The `newOrderRespType` parameter defaults to `ACK` for non-LIMIT/MARKET orders and `FULL` for LIMIT/MARKET orders.
- Using `orderId` for cancel/query is faster than using `origClientOrderId`.
- `recvWindow` supports up to 3 decimal places for microsecond precision (e.g., `6000.346`).
