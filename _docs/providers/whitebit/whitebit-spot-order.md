# WhiteBit Spot Market Order REST API Documentation

## Reference

- **Official API Docs**: https://docs.zondacrypto.exchange/reference/introduction
- **Base URLs**:
  - Global: `https://whitebit.com/api/v4`
  - EU: `https://whitebit.eu/api/v4`

## Authentication

All order endpoints are private and require authentication.

### Required Headers

| Header | Value | Description |
|--------|-------|-------------|
| `Content-type` | `application/json` | Specifies JSON format |
| `X-TXC-APIKEY` | `YOUR_API_KEY` | Public API key |
| `X-TXC-PAYLOAD` | `base64_encoded_payload` | Base64-encoded request body |
| `X-TXC-SIGNATURE` | `signature` | HMAC-SHA512 signature (hex encoded) |

### Signature Method (HMAC-SHA512)

```
signature = hex(HMAC_SHA512(payload, key=api_secret))
```

Where `payload` is the raw JSON request body (before base64 encoding).

### Request Body Format

Every private request body must include:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `request` | string | Yes | Request path without domain (e.g., `/api/v4/order/new`) |
| `nonce` | string | Yes | Incrementing number larger than previous requests (use Unix timestamp in ms) |
| `nonceWindow` | boolean | No | Enable time-based nonce validation (±5 seconds) |

### Nonce Management

- Use Unix timestamp in milliseconds
- Each nonce must be larger than previous requests
- When `nonceWindow: true`: timestamp must be within ±5 seconds of server time

## Place Limit Order

- **Description**: Create a new limit order.
- **Endpoint**: `POST /api/v4/order/new`
- **Rate Limit**: 10,000 requests / 10 sec
- **Docs**: https://docs.zondacrypto.exchange/reference/neworder

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `market` | string | Yes | Trading pair (e.g., `BTC_USDT`) |
| `side` | string | Yes | `buy` or `sell` |
| `amount` | string | Yes | Amount in stock currency (e.g., `"0.001"`) |
| `price` | string | Yes | Price in money currency (e.g., `"9800"`) |
| `client_order_id` | string | No | Custom unique identifier (letters, numbers, dashes, dots, underscores only) |
| `postOnly` | boolean | No | Default: `false`. Guarantees maker order |
| `ioc` | boolean | No | Default: `false`. Immediate-or-cancel |
| `bboRole` | integer | No | `1` = Queue Method, `2` = Counterparty Method (use with IOC) |
| `stp` | string | No | Self-trade prevention: `no`, `cancel_both`, `cancel_new`, `cancel_old` |
| `rpi` | boolean | No | Retail Price Improvement mode (post-only by design) |
| `request` | string | Yes | `/api/v4/order/new` |
| `nonce` | string | Yes | Unique incrementing value |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `order_id` | integer | Order ID |
| `client_order_id` | string | Custom client order ID (empty if not specified) |
| `market` | string | Trading pair |
| `side` | string | `buy` or `sell` |
| `type` | string | Order type (`limit`) |
| `timestamp` | number | Creation timestamp |
| `deal_money` | string | Filled amount in money currency |
| `deal_stock` | string | Filled amount in stock currency |
| `amount` | string | Original order amount |
| `left` | string | Remaining unfilled amount |
| `deal_fee` | string | Fee in money currency |
| `price` | string | Order price |
| `postOnly` | boolean | Post-only flag |
| `ioc` | boolean | IOC flag |
| `status` | string | Order status |
| `stp` | string | Self-trade prevention mode |
| `rpi` | boolean | RPI mode flag |

### Sample Request

```json
{
  "request": "/api/v4/order/new",
  "nonce": 1594297865000,
  "market": "BTC_USDT",
  "side": "buy",
  "amount": "0.001",
  "price": "40000",
  "client_order_id": "order1987111",
  "postOnly": false
}
```

### Sample Response

```json
{
  "order_id": 4180284841,
  "client_order_id": "order1987111",
  "market": "BTC_USDT",
  "side": "buy",
  "type": "limit",
  "timestamp": 1595792396.165973,
  "deal_money": "0",
  "deal_stock": "0",
  "amount": "0.01",
  "left": "0.001",
  "deal_fee": "0",
  "price": "40000",
  "postOnly": false,
  "ioc": false,
  "status": "NEW",
  "stp": "no",
  "rpi": false
}
```

## Place Market Order

- **Description**: Create a new market order.
- **Endpoint**: `POST /api/v4/order/market`
- **Rate Limit**: 10,000 requests / 10 sec
- **Docs**: https://docs.zondacrypto.exchange/reference/newmarketorder

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `market` | string | Yes | Trading pair (e.g., `BTC_USDT`) |
| `side` | string | Yes | `buy` or `sell` |
| `amount` | string | Yes | **Buy**: amount in money currency (min total). **Sell**: amount in stock currency |
| `client_order_id` | string | No | Custom unique identifier |
| `stp` | string | No | Self-trade prevention: `no`, `cancel_both`, `cancel_new`, `cancel_old` |
| `request` | string | Yes | `/api/v4/order/market` |
| `nonce` | string | Yes | Unique incrementing value |

### Sample Request

```json
{
  "request": "/api/v4/order/market",
  "nonce": 1594297865000,
  "market": "BTC_USDT",
  "side": "buy",
  "amount": "100",
  "client_order_id": "market_order_001"
}
```

## Place Stop-Limit Order

- **Description**: Create a new stop-limit order.
- **Endpoint**: `POST /api/v4/order/stop_limit`
- **Rate Limit**: 10,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `market` | string | Yes | Trading pair |
| `side` | string | Yes | `buy` or `sell` |
| `amount` | string | Yes | Amount in stock currency |
| `price` | string | Yes | Limit price in money currency |
| `activation_price` | string | Yes | Trigger price in money currency |
| `client_order_id` | string | No | Custom unique identifier |
| `bboRole` | integer | No | `1` = Queue, `2` = Counterparty |
| `stp` | string | No | Self-trade prevention mode |
| `request` | string | Yes | `/api/v4/order/stop_limit` |
| `nonce` | string | Yes | Unique incrementing value |

### Sample Request

```json
{
  "request": "/api/v4/order/stop_limit",
  "nonce": 1594297865000,
  "market": "BTC_USDT",
  "side": "sell",
  "amount": "0.001",
  "price": "39000",
  "activation_price": "39500",
  "client_order_id": "stop_limit_001"
}
```

## Place Stop-Market Order

- **Description**: Create a new stop-market order.
- **Endpoint**: `POST /api/v4/order/stop_market`
- **Rate Limit**: 10,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `market` | string | Yes | Trading pair |
| `side` | string | Yes | `buy` or `sell` |
| `amount` | string | Yes | **Buy**: money currency. **Sell**: stock currency |
| `activation_price` | string | Yes | Trigger price in money currency |
| `client_order_id` | string | No | Custom unique identifier |
| `stp` | string | No | Self-trade prevention mode |
| `request` | string | Yes | `/api/v4/order/stop_market` |
| `nonce` | string | Yes | Unique incrementing value |

## Bulk Limit Orders

- **Description**: Place up to 20 limit orders in a single request.
- **Endpoint**: `POST /api/v4/order/bulk`
- **Rate Limit**: 10,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `orders` | array | Yes | Array of 1-20 limit orders |
| `stopOnFail` | boolean | No | Default: `false`. Stop processing on first failure |
| `request` | string | Yes | `/api/v4/order/bulk` |
| `nonce` | string | Yes | Unique incrementing value |

Each order in `orders` array:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `market` | string | Yes | Trading pair |
| `side` | string | Yes | `buy` or `sell` |
| `amount` | string | Yes | Amount in stock currency |
| `price` | string | Yes | Price in money currency |
| `client_order_id` | string | No | Custom identifier |
| `postOnly` | boolean | No | Post-only flag |
| `ioc` | boolean | No | Immediate-or-cancel |
| `rpi` | boolean | No | Retail Price Improvement |

### Sample Response

```json
[
  {
    "result": {
      "order_id": 4180284841,
      "client_order_id": "bulk_1",
      "market": "BTC_USDT",
      "side": "buy",
      "type": "limit",
      "timestamp": 1595792396.165973,
      "deal_money": "0",
      "deal_stock": "0",
      "amount": "0.02",
      "left": "0.02",
      "deal_fee": "0",
      "price": "40000",
      "postOnly": false,
      "ioc": false,
      "status": "NEW",
      "stp": "no",
      "rpi": false
    },
    "error": null
  },
  {
    "result": null,
    "error": {
      "code": 30,
      "message": "Validation failed",
      "errors": {
        "amount": ["Not enough balance."]
      }
    }
  }
]
```

## Query Order (Active Orders)

- **Description**: Query unexecuted (active) orders.
- **Endpoint**: `POST /api/v4/orders`
- **Rate Limit**: 12,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `market` | string | No | Trading pair (omit for all markets) |
| `order_id` | integer | No | Specific order ID |
| `client_order_id` | string | No | Specific client order ID |
| `offset` | integer | No | Default: 0, Min: 0, Max: 10,000 |
| `limit` | integer | No | Default: 50, Min: 1, Max: 100 |
| `request` | string | Yes | `/api/v4/orders` |
| `nonce` | string | Yes | Unique incrementing value |

**Note:** Search across all markets is available only if `client_order_id` and `order_id` are not provided.

### Response

Returns an array of order objects (same fields as Place Limit Order response).

### Sample Request

```json
{
  "request": "/api/v4/orders",
  "nonce": 1594297865000,
  "market": "BTC_USDT",
  "offset": 0,
  "limit": 50
}
```

### Sample Response

```json
[
  {
    "order_id": 4180284841,
    "client_order_id": "order1987111",
    "market": "BTC_USDT",
    "side": "buy",
    "type": "limit",
    "timestamp": 1595792396.165973,
    "deal_money": "0",
    "deal_stock": "0",
    "amount": "0.01",
    "left": "0.001",
    "deal_fee": "0",
    "price": "40000",
    "postOnly": false,
    "ioc": false,
    "status": "NEW",
    "stp": "no",
    "rpi": false
  }
]
```

## Query Order Deals

- **Description**: Query executed order deals (fills).
- **Endpoint**: `POST /api/v4/trade-account/order`
- **Rate Limit**: 12,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `order_id` | integer | Yes | Order ID to query deals for |
| `offset` | integer | No | Default: 0 |
| `limit` | integer | No | Default: 50 |
| `request` | string | Yes | `/api/v4/trade-account/order` |
| `nonce` | string | Yes | Unique incrementing value |

### Response Fields (per deal)

| Field | Type | Description |
|-------|------|-------------|
| `id` | integer | Deal identifier |
| `dealOrderId` | integer | Parent order ID |
| `client_order_id` | string | Custom order ID |
| `time` | number | Execution timestamp |
| `side` | string | `buy` or `sell` |
| `role` | integer | `1` = maker, `2` = taker |
| `amount` | string | Amount in stock currency |
| `price` | string | Deal price |
| `deal` | string | Amount in money currency |
| `fee` | string | Fee paid |
| `feeAsset` | string | Fee currency |
| `rpi` | boolean | RPI flag |

### Sample Response

```json
{
  "records": [
    {
      "id": 149156519,
      "dealOrderId": 3134995325,
      "client_order_id": "customId11",
      "time": 1593342324.613711,
      "side": "buy",
      "role": 2,
      "amount": "598",
      "price": "0.00000701",
      "deal": "0.00419198",
      "fee": "0.00000419198",
      "feeAsset": "USDT",
      "rpi": true
    }
  ],
  "offset": 0,
  "limit": 100
}
```

## Cancel Order

- **Description**: Cancel a single active order.
- **Endpoint**: `POST /api/v4/order/cancel`
- **Rate Limit**: 10,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `market` | string | Yes | Trading pair |
| `order_id` | integer | No* | Order ID to cancel |
| `client_order_id` | string | No* | Client order ID to cancel |
| `request` | string | Yes | `/api/v4/order/cancel` |
| `nonce` | string | Yes | Unique incrementing value |

*Either `order_id` OR `client_order_id` is required. Do not pass both. `client_order_id` takes priority.

### Response

Returns the cancelled order details in the same format as Place Limit Order response.

### Sample Request

```json
{
  "request": "/api/v4/order/cancel",
  "nonce": 1594297865000,
  "market": "BTC_USDT",
  "order_id": 4180284841
}
```

## Cancel All Orders

- **Description**: Cancel all active orders, optionally filtered by market.
- **Endpoint**: `POST /api/v4/order/cancel/all`
- **Rate Limit**: 10,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `market` | string | No | Trading pair (omit for all markets) |
| `type` | array | No | Order types: `spot`, `margin`, `futures` |
| `request` | string | Yes | `/api/v4/order/cancel/all` |
| `nonce` | string | Yes | Unique incrementing value |

### Sample Request

```json
{
  "request": "/api/v4/order/cancel/all",
  "nonce": 1594297865000,
  "market": "BTC_USDT",
  "type": ["spot"]
}
```

## Order History

- **Description**: Query executed order history.
- **Endpoint**: `POST /api/v4/trade-account/order/history`
- **Rate Limit**: 12,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `market` | string | No | Trading pair filter |
| `offset` | integer | No | Default: 0 |
| `limit` | integer | No | Default: 50 |
| `request` | string | Yes | `/api/v4/trade-account/order/history` |
| `nonce` | string | Yes | Unique incrementing value |

### Response Format

Returns object with market names as keys, each containing an array of orders:

```json
{
  "BTC_USDT": [
    {
      "id": 4986126152,
      "client_order_id": "customId11",
      "amount": "0.0009",
      "price": "40000",
      "type": "limit",
      "side": "sell",
      "ctime": 1597486960.311311,
      "ftime": 1597486960.311332,
      "takerFee": "0.001",
      "makerFee": "0.001",
      "deal_fee": "0.041258268",
      "deal_stock": "0.0009",
      "deal_money": "41.258268",
      "postOnly": false,
      "ioc": false,
      "status": "CANCELED",
      "feeAsset": "USDT",
      "stp": "no",
      "rpi": false
    }
  ]
}
```

## Executed Order History (Deals)

- **Description**: Query executed deal history.
- **Endpoint**: `POST /api/v4/trade-account/executed-history`
- **Rate Limit**: 12,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `market` | string | No | Trading pair filter |
| `client_order_id` | string | No | Filter by client order ID |
| `startDate` | integer | No | Start date (Unix timestamp) |
| `endDate` | integer | No | End date (Unix timestamp) |
| `offset` | integer | No | Default: 0, Min: 0 |
| `limit` | integer | No | Default: 50, Min: 1, Max: 100 |
| `request` | string | Yes | `/api/v4/trade-account/executed-history` |
| `nonce` | string | Yes | Unique incrementing value |

**Note:** Can retrieve data not older than 6 months from current month.

## Order Types

| Type | Endpoint | Required Parameters | Description |
|------|----------|---------------------|-------------|
| **Limit** | `POST /api/v4/order/new` | `market`, `side`, `amount`, `price` | Execute at specified price or better |
| **Market** | `POST /api/v4/order/market` | `market`, `side`, `amount` | Execute immediately at best available price |
| **Stop-Limit** | `POST /api/v4/order/stop_limit` | `market`, `side`, `amount`, `price`, `activation_price` | Trigger limit order when activation price is reached |
| **Stop-Market** | `POST /api/v4/order/stop_market` | `market`, `side`, `amount`, `activation_price` | Trigger market order when activation price is reached |

## Trade Side

| Value | Description |
|-------|-------------|
| `buy` | Buy order (acquire stock currency, spend money currency) |
| `sell` | Sell order (sell stock currency, receive money currency) |

## Time In Force

WhiteBit uses boolean flags instead of traditional GTC/IOC/FOK terminology:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `postOnly` | boolean | `false` | Order is guaranteed to be maker. Cancelled if it would match immediately. |
| `ioc` | boolean | `false` | Immediate-or-Cancel. Executes all or part immediately, cancels unfilled portion. |

**Implicit behavior:** Orders without `postOnly` or `ioc` flags behave as **Good-Til-Cancelled (GTC)** — they remain active until filled or manually cancelled.

### BBO (Best Bid/Offer) Role

| Value | Method | Description |
|-------|--------|-------------|
| `1` | Queue Method | Order joins the queue at BBO price |
| `2` | Counterparty Method | Order takes liquidity from counterparty (use with IOC) |

## Self Trade Prevention

| Value | Description |
|-------|-------------|
| `no` | No self-trade prevention (default) |
| `cancel_both` | Cancel both orders |
| `cancel_new` | Cancel the new (incoming) order |
| `cancel_old` | Cancel the old (existing) order |

## Constraints & Limits

### Rate Limits

| Endpoint | Rate Limit |
|----------|------------|
| `POST /api/v4/order/new` (Limit) | 10,000 requests / 10 sec |
| `POST /api/v4/order/market` (Market) | 10,000 requests / 10 sec |
| `POST /api/v4/order/stop_limit` (Stop-Limit) | 10,000 requests / 10 sec |
| `POST /api/v4/order/stop_market` (Stop-Market) | 10,000 requests / 10 sec |
| `POST /api/v4/order/bulk` (Bulk Limit) | 10,000 requests / 10 sec |
| `POST /api/v4/order/cancel` (Cancel) | 10,000 requests / 10 sec |
| `POST /api/v4/order/cancel/all` (Cancel All) | 10,000 requests / 10 sec |
| `POST /api/v4/orders` (Active Orders) | 12,000 requests / 10 sec |
| `POST /api/v4/trade-account/order/history` (Order History) | 12,000 requests / 10 sec |
| `POST /api/v4/trade-account/executed-history` (Deal History) | 12,000 requests / 10 sec |
| `POST /api/v4/trade-account/order` (Order Deals) | 12,000 requests / 10 sec |

### Key Constraints

- **Minimum Order Value:** Total (`amount * price`) must be at least 5.05 USDT equivalent
- **Client Order ID:** Must contain only Latin letters, numbers, and dashes. Must be unique per account
- **IOC + PostOnly Conflict:** Cannot use `ioc=true` with `postOnly=true` or `rpi=true` (error code 37)
- **RPI Orders:** Retail Price Improvement orders are post-only by design, not visible in public order book feeds
- **Cancel Order Priority:** `client_order_id` takes priority over `order_id`. Do not pass both simultaneously
- **Bulk Orders:** 1-20 orders per request. `stopOnFail` controls failure behavior
- **History Retention:** Executed order history available for 6 months only
- **Market Format:** Must be `STOCK_MONEY` format (e.g., `BTC_USDT`)
- **Numeric Strings:** All amounts and prices should be sent as numeric strings to preserve precision

## Error Codes

### Validation Error Codes

| Code | Description |
|------|-------------|
| `30` | Default validation error |
| `31` | Market validation failed |
| `32` | Amount validation failed |
| `33` | Price validation failed |
| `35` | Maker fee validation failed |
| `36` | Client order ID validation failed |
| `37` | IOC + PostOnly/RPI conflict |

### HTTP Status Codes

| Status | Meaning |
|--------|---------|
| `200` | Success |
| `400` | Bad request - invalid parameters |
| `401` | Unauthorized - missing/invalid authentication |
| `403` | Forbidden - insufficient permissions |
| `422` | Request validation failed |
| `429` | Too Many Requests - rate limit exceeded |
| `503` | Service temporarily unavailable |

## Notes

- WhiteBit uses **separate endpoints** for different order types (limit, market, stop-limit, stop-market) rather than a unified order endpoint with a `type` parameter.
- Symbol format uses underscore separator: `BTC_USDT` (not `BTCUSDT`).
- Timestamps are in Unix-time format with microseconds (e.g., `1595792396.165973`).
- All private endpoints use POST with JSON body, even for queries.
- When rate limit is exceeded, returns HTTP `429 Too Many Requests`. Use exponential backoff: 1s → 2s → 4s → 8s.
