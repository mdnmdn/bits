# MEXC Spot Market Order REST API Documentation

## Reference

- **Official API Docs**: https://www.mexc.com/api-docs/spot-v3/introduction
- **Base URL**: `https://api.mexc.com`

## Authentication

All order endpoints require signed requests:

| Requirement | Detail |
|-------------|--------|
| **API Key Header** | `X-MEXC-APIKEY: <your_api_key>` |
| **Signature Method** | HMAC-SHA256 |
| **Signature Location** | Query string parameter `signature` |
| **Signature Input** | `totalParams` = query string concatenated with request body (no `&` between them if mixed) |
| **Timestamp** | Millisecond timestamp, required on all signed endpoints |
| **recvWindow** | Optional, default 5000ms, max 60000ms |
| **Signature Case** | Lowercase only |

### Timing Validation

```
timestamp < (serverTime + 1000) AND (serverTime - timestamp) <= recvWindow
```

## Place Order

- **Description**: Create a new order (limit, market, limit maker, IOC, FOK).
- **Endpoint**: `POST /api/v3/order`
- **Permission**: `SPOT_DEAL_WRITE`
- **Rate Limit**: 12 requests/second (shared with cancel order)
- **Docs**: https://www.mexc.com/api-docs/spot-v3/introduction

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair (e.g. `BTCUSDT`) |
| `side` | ENUM | YES | `BUY` or `SELL` |
| `type` | ENUM | YES | Order type (see Order Types section) |
| `quantity` | DECIMAL | NO | Amount of base asset |
| `quoteOrderQty` | DECIMAL | NO | Amount of quote asset to spend |
| `price` | DECIMAL | NO | Limit price |
| `newClientOrderId` | STRING | NO | User-defined client order ID |
| `stpMode` | STRING | NO | Self-trade prevention: `""`, `"cancel_maker"`, `"cancel_taker"`, `"cancel_both"` |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Millisecond timestamp |
| `signature` | STRING | YES | HMAC-SHA256 signature |

### Required Parameters by Order Type

| Type | Required Parameters |
|------|---------------------|
| `LIMIT` | `quantity`, `price` |
| `MARKET` | `quantity` OR `quoteOrderQty` |
| `LIMIT_MAKER` | `quantity`, `price` |
| `IMMEDIATE_OR_CANCEL` | `quantity`, `price` |
| `FILL_OR_KILL` | `quantity`, `price` |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `symbol` | STRING | Trading pair |
| `orderId` | STRING | Order ID |
| `orderListId` | INT | OCO list ID (-1 if not OCO) |
| `price` | STRING | Price |
| `origQty` | STRING | Original quantity |
| `type` | ENUM | Order type |
| `side` | ENUM | BUY or SELL |
| `stpMode` | STRING | Self-trade prevention mode |
| `transactTime` | LONG | Transaction timestamp |

### Sample Request

```bash
curl -H "X-MEXC-APIKEY: mx0aBYs33eIilxBQC5" \
  -X POST 'https://api.mexc.com/api/v3/order' \
  -d 'symbol=BTCUSDT&side=BUY&type=LIMIT&quantity=1&price=11&recvWindow=5000&timestamp=1644489390087&signature=<hmac_sha256_sig>'
```

### Sample Response

```json
{
  "symbol": "MXUSDT",
  "orderId": "06a480e69e604477bfb48dddd5f0b750",
  "orderListId": -1,
  "price": "0.1",
  "origQty": "50",
  "type": "LIMIT",
  "side": "BUY",
  "stpMode": "",
  "transactTime": 1666676533741
}
```

## Test New Order

- **Description**: Creates and validates a new order without sending it to the matching engine.
- **Endpoint**: `POST /api/v3/order/test`
- **Permission**: `SPOT_DEAL_WRITE`
- **Rate Limit**: Weight(IP) = 1

Uses the same parameters as `POST /api/v3/order`.

### Sample Response

```json
{}
```

## Query Order

- **Description**: Check an order's status.
- **Endpoint**: `GET /api/v3/order`
- **Permission**: `SPOT_DEAL_READ`
- **Rate Limit**: Weight(IP) = 2

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair |
| `orderId` | STRING | NO | Order ID |
| `origClientOrderId` | STRING | NO | Original client order ID |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Millisecond timestamp |
| `signature` | STRING | YES | HMAC-SHA256 signature |

**Note**: Either `orderId` or `origClientOrderId` must be sent.

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `symbol` | STRING | Trading pair |
| `orderId` | LONG | Order ID |
| `orderListId` | INT | OCO list ID |
| `clientOrderId` | STRING | Client order ID |
| `price` | STRING | Price |
| `Qty` | STRING | Original quantity |
| `executedQty` | STRING | Executed quantity |
| `cummulativeQuoteQty` | STRING | Cumulative quote quantity |
| `status` | ENUM | Order status |
| `timeInForce` | STRING | Time in force |
| `type` | ENUM | Order type |
| `side` | ENUM | BUY or SELL |
| `stopPrice` | STRING | Stop price |
| `time` | LONG | Order creation time |
| `updateTime` | LONG | Last update time |
| `isWorking` | BOOL | Is order in orderbook |
| `stpMode` | STRING | Self-trade prevention mode |
| `cancelReason` | STRING | Cancel reason (e.g. `stp_cancel`) |
| `origQuoteOrderQty` | STRING | Original quote order quantity |

### Sample Response

```json
{
  "symbol": "LTCBTC",
  "orderId": 1,
  "orderListId": -1,
  "clientOrderId": "myOrder1",
  "price": "0.1",
  "Qty": "1.0",
  "executedQty": "0.0",
  "cummulativeQuoteQty": "0.0",
  "status": "NEW",
  "timeInForce": "GTC",
  "type": "LIMIT",
  "side": "BUY",
  "stopPrice": "0.0",
  "time": 1499827319559,
  "updateTime": 1499827319559,
  "stpMode": "",
  "cancelReason": "stp_cancel",
  "isWorking": true,
  "origQuoteOrderQty": "0.000000"
}
```

## Cancel Order

- **Description**: Cancel an active order.
- **Endpoint**: `DELETE /api/v3/order`
- **Permission**: `SPOT_DEAL_WRITE`
- **Rate Limit**: 12 requests/second (shared with place order)

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair |
| `orderId` | STRING | NO | Order ID |
| `origClientOrderId` | STRING | NO | Original client order ID |
| `newClientOrderId` | STRING | NO | New client order ID for the cancellation |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Millisecond timestamp |
| `signature` | STRING | YES | HMAC-SHA256 signature |

**Note**: Either `orderId` or `origClientOrderId` must be sent.

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `symbol` | STRING | Trading pair |
| `origClientOrderId` | STRING | Original client order ID |
| `orderId` | LONG | Order ID |
| `clientOrderId` | STRING | Client order ID |
| `price` | STRING | Price |
| `origQty` | STRING | Original quantity |
| `executedQty` | STRING | Executed quantity |
| `cummulativeQuoteQty` | STRING | Cumulative quote quantity |
| `status` | ENUM | Order status |
| `timeInForce` | STRING | Time in force |
| `type` | ENUM | Order type |
| `side` | ENUM | BUY or SELL |

### Sample Response

```json
{
  "symbol": "LTCBTC",
  "origClientOrderId": "myOrder1",
  "orderId": 4,
  "clientOrderId": "cancelMyOrder1",
  "price": "2.00000000",
  "origQty": "1.00000000",
  "executedQty": "0.00000000",
  "cummulativeQuoteQty": "0.00000000",
  "status": "CANCELED",
  "timeInForce": "GTC",
  "type": "LIMIT",
  "side": "BUY"
}
```

## Cancel All Orders

- **Description**: Cancel all open orders on a symbol.
- **Endpoint**: `DELETE /api/v3/openOrders`
- **Permission**: `SPOT_DEAL_WRITE`
- **Rate Limit**: 12 requests/second (shared with place order)

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Millisecond timestamp |
| `signature` | STRING | YES | HMAC-SHA256 signature |

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
    "price": "0.089853",
    "origQty": "0.178622",
    "executedQty": "0.000000",
    "cummulativeQuoteQty": "0.000000",
    "status": "CANCELED",
    "timeInForce": "GTC",
    "type": "LIMIT",
    "side": "BUY"
  }
]
```

## Open Orders

- **Description**: Get all open orders.
- **Endpoint**: `GET /api/v3/openOrders`
- **Permission**: `SPOT_DEAL_READ`
- **Rate Limit**: Weight(IP) = 3

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | NO | Trading pair (if omitted, returns all open orders) |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Millisecond timestamp |
| `signature` | STRING | YES | HMAC-SHA256 signature |

### Response

Returns an array of open orders (same fields as Query Order response).

## Order History

- **Description**: Get all orders (active, canceled, filled) with pagination.
- **Endpoint**: `GET /api/v3/allOrders`
- **Permission**: `SPOT_DEAL_READ`
- **Rate Limit**: Weight(IP) = 10

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair |
| `startTime` | LONG | NO | Start timestamp (ms) |
| `endTime` | LONG | NO | End timestamp (ms) |
| `limit` | INT | NO | Default 500, max 1000 |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Millisecond timestamp |
| `signature` | STRING | YES | HMAC-SHA256 signature |

**Query Window**: Default latest 24 hours. Maximum query range: **7 days**.

### Response

Returns an array of orders (same fields as Query Order response).

## Batch Orders

- **Description**: Place up to 20 orders in a single request.
- **Endpoint**: `POST /api/v3/batchOrders`
- **Permission**: `SPOT_DEAL_WRITE`
- **Rate Limit**: 2 times/second

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `batchOrders` | LIST | YES | List of orders (max 20) |
| `timestamp` | LONG | YES | Millisecond timestamp |
| `signature` | STRING | YES | HMAC-SHA256 signature |

Each order in `batchOrders` supports: `symbol`, `side`, `type`, `quantity`, `quoteOrderQty`, `price`, `newClientOrderId`, `stpMode`, `recvWindow`.

**Note**: All orders must be for the **same symbol**.

### Sample Request

```bash
POST /api/v3/batchOrders?batchOrders=[{"type":"LIMIT_ORDER","price":"40000","quantity":"0.0002","symbol":"BTCUSDT","side":"BUY","newClientOrderId":9588234},{"type":"LIMIT_ORDER","price":"4005","quantity":"0.0003","symbol":"BTCUSDT","side":"SELL"}]
```

### Sample Response

```json
[
  {
    "symbol": "BTCUSDT",
    "orderId": "1196315350023612316",
    "orderListId": -1
  },
  {
    "symbol": "BTCUSDT",
    "orderId": "1196315350023612318",
    "orderListId": -1
  }
]
```

## Order Types

| Type | Description | Required Params |
|------|-------------|-----------------|
| `LIMIT` | Standard limit order | `quantity`, `price` |
| `MARKET` | Market order at best available price | `quantity` OR `quoteOrderQty` |
| `LIMIT_MAKER` | Limit order that will not match immediately (post-only) | `quantity`, `price` |
| `IMMEDIATE_OR_CANCEL` | IOC — fill as much as possible, cancel remainder | `quantity`, `price` |
| `FILL_OR_KILL` | FOK — fill entirely or cancel entirely | `quantity`, `price` |
| `STOP_MARKET_ORDER` | Stop market order (query only, cannot be placed) | — |

## Trade Side

| Value | Description |
|-------|-------------|
| `BUY` | Buy order |
| `SELL` | Sell order |

## Time In Force

MEXC does not expose a `timeInForce` parameter on order placement. Instead, the time-in-force behavior is determined by the **order type**:

| Order Type | Implicit TIF |
|------------|--------------|
| `LIMIT` | GTC (Good Till Cancel) — stays in orderbook until filled or canceled |
| `LIMIT_MAKER` | GTC, post-only — will be rejected/canceled if it would match immediately |
| `IMMEDIATE_OR_CANCEL` | IOC — execute immediately, cancel unfilled portion |
| `FILL_OR_KILL` | FOK — execute entirely or cancel entirely |
| `MARKET` | Immediate execution at market price |

The `timeInForce` field appears in query/cancel responses to indicate the effective TIF of the order.

## Order Status

| Status | Description |
|--------|-------------|
| `NEW` | Order accepted, not yet filled |
| `FILLED` | Fully filled |
| `PARTIALLY_FILLED` | Partially filled, still active |
| `CANCELED` | Canceled by user |
| `PARTIALLY_CANCELED` | Partially filled then canceled |

## Self Trade Prevention (STP)

| `stpMode` Value | Behavior |
|-----------------|----------|
| `""` (empty) | No self-trade prevention (default) |
| `cancel_maker` | Cancel the maker order |
| `cancel_taker` | Cancel the taker order |
| `cancel_both` | Cancel both sides |

**STP Requirements**:
1. At least one strategy group must be created
2. `stpMode` must not be empty
3. When `stpMode = ""`, self-trading is not restricted

## Constraints & Limits

### Rate Limits

| Endpoint | Method | Rate Limit |
|----------|--------|------------|
| `/api/v3/order` (place) | POST | 12 req/s (shared with cancel) |
| `/api/v3/order/test` | POST | Weight(IP) = 1 |
| `/api/v3/order` (cancel) | DELETE | 12 req/s (shared with place) |
| `/api/v3/openOrders` (cancel all) | DELETE | 12 req/s (shared with place) |
| `/api/v3/order` (query) | GET | Weight(IP) = 2 |
| `/api/v3/openOrders` (list open) | GET | Weight(IP) = 3 |
| `/api/v3/allOrders` (history) | GET | Weight(IP) = 10 |
| `/api/v3/batchOrders` | POST | 2 req/s |

### Global Rate Limit Rules

- **IP limit**: Each endpoint with IP limits has an independent **300 requests per 10 seconds** limit
- **UID limit**: Each endpoint with UID limits has an independent **500 requests per 10 seconds** limit
- HTTP 429 = rate limit exceeded — must back off or face IP ban (2 min to 3 days)
- `Retry-After` header sent with 418/429 responses

### Key Constraints

- **Market order `quantity`**: Specifies the base asset amount to sell/buy
- **Market order `quoteOrderQty`**: Specifies the quote asset amount to spend (BUY) or receive (SELL)
- **Batch orders**: All orders must be for the same symbol, max 20 per request
- **Order history query window**: Maximum 7 days
- **`stpMode` requires at least one STP strategy group to be created** for it to take effect
- **Timestamp validation**: `timestamp < (serverTime + 1000) AND (serverTime - timestamp) <= recvWindow`

## Error Codes

| Code | Description |
|------|-------------|
| `-2011` | Unknown order sent |
| `30000` | Suspended trading for the symbol |
| `30001` | Current transaction direction not allowed |
| `30002` | Minimum transaction volumes violated |
| `30003` | Maximum transaction volume exceeded |
| `30004` | Insufficient position |
| `30018` | Market order is disabled |
| `30019` | API market order is disabled |
| `30029` | Cannot exceed maximum order limit |
| `700001` | API-key format invalid |
| `700002` | Signature for this request is not valid |
| `700003` | Timestamp outside of recvWindow |
| `700004` | Neither `origClientOrderId` nor `orderId` sent |

## Notes

- MEXC spot API is **Binance-compatible** — endpoint paths, parameter names, and response formats closely mirror Binance's API.
- Symbol format uses **no separator**: `BTCUSDT` (not `BTC_USDT`).
- The `orderId` field in place order response is a **string**, while in query/cancel responses it is a **LONG**.
- All numeric values in responses are returned as strings to preserve precision.
- `STOP_MARKET_ORDER` type exists but is **query only** — cannot be placed directly.
