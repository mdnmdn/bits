# WhiteBit Futures & Margin Order REST API Documentation

## Reference

- **Official API Docs**: https://docs.zondacrypto.exchange/reference/introduction
- **Base URLs**:
  - Global: `https://whitebit.com/api/v4`
  - EU: `https://whitebit.eu/api/v4`

## Unified Collateral Trading System

WhiteBit uses a **unified "Collateral Trading" system** for margin and perpetual futures. There is no separate futures-specific order API — the same collateral trading endpoints work for both margin and perpetual futures, distinguished only by the `market` parameter.

### Symbol Format

| Market Type | Format | Examples |
|-------------|--------|----------|
| Spot | `BASE_QUOTE` | `BTC_USDT`, `ETH_BTC` |
| Margin | `BASE_QUOTE` | `BTC_USDT` |
| Futures/Perpetuals | `BASE_PERP` | `BTC_PERP`, `ETH_PERP` |

## Authentication

All collateral trading endpoints are private and require authentication.

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
| `request` | string | Yes | Request path without domain (e.g., `/api/v4/order/collateral/limit`) |
| `nonce` | string | Yes | Incrementing number larger than previous requests (use Unix timestamp in ms) |
| `nonceWindow` | boolean | No | Enable time-based nonce validation (±5 seconds) |

## Place Collateral Limit Order

- **Description**: Create a new limit order for margin or perpetual futures.
- **Endpoint**: `POST /api/v4/order/collateral/limit`
- **Rate Limit**: 10,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `market` | string | Yes | Market name (`BTC_PERP` for futures, `BTC_USDT` for margin) |
| `side` | string | Yes | `buy` or `sell` |
| `amount` | string | Yes | Amount in stock currency |
| `price` | string | Yes | Price in money currency |
| `client_order_id` | string | No | Custom unique identifier |
| `stopLoss` | string | No | Stop loss price (creates OTO order) |
| `takeProfit` | string | No | Take profit price (creates OTO order) |
| `postOnly` | boolean | No | Default: `false`. Guarantees maker order |
| `ioc` | boolean | No | Default: `false`. Immediate-or-cancel |
| `rpi` | boolean | No | Retail Price Improvement mode (post-only by design) |
| `positionSide` | string | No | `LONG`, `SHORT`, or `BOTH` (for hedge mode) |
| `request` | string | Yes | `/api/v4/order/collateral/limit` |
| `nonce` | string | Yes | Unique incrementing value |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `order_id` | integer | Order ID |
| `client_order_id` | string | Custom client order ID |
| `market` | string | Market name |
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
| `status` | string | `NEW`, `FILLED`, `PARTIALLY_FILLED`, `CANCELLED` |
| `stp` | string | Self-trade prevention mode |
| `positionSide` | string | `LONG`, `SHORT`, or `BOTH` |
| `rpi` | boolean | RPI mode flag |
| `oto` | object | OTO order data (if stopLoss/takeProfit provided) |

### Sample Request

```json
{
  "request": "/api/v4/order/collateral/limit",
  "nonce": 1594297865000,
  "market": "BTC_PERP",
  "side": "buy",
  "amount": "0.01",
  "price": "40000",
  "client_order_id": "order1987111",
  "stopLoss": "38000",
  "takeProfit": "45000",
  "positionSide": "LONG",
  "postOnly": false,
  "ioc": false,
  "rpi": true
}
```

### Sample Response

```json
{
  "order_id": 4180284841,
  "client_order_id": "order1987111",
  "market": "BTC_PERP",
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
  "status": "FILLED",
  "stp": "no",
  "positionSide": "LONG",
  "rpi": true,
  "oto": {
    "otoId": 29457221,
    "stopLoss": "38000",
    "takeProfit": "45000"
  }
}
```

## Place Collateral Market Order

- **Description**: Create a new market order for margin or perpetual futures.
- **Endpoint**: `POST /api/v4/order/collateral/market`
- **Rate Limit**: 10,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `market` | string | Yes | Market name |
| `side` | string | Yes | `buy` or `sell` |
| `amount` | string | Yes | Amount in stock currency |
| `client_order_id` | string | No | Custom unique identifier |
| `stopLoss` | string | No | Stop loss price |
| `takeProfit` | string | No | Take profit price |
| `positionSide` | string | No | `LONG`, `SHORT`, or `BOTH` |
| `request` | string | Yes | `/api/v4/order/collateral/market` |
| `nonce` | string | Yes | Unique incrementing value |

## Place Collateral Stop-Limit Order

- **Description**: Create a new stop-limit order for margin or perpetual futures.
- **Endpoint**: `POST /api/v4/order/collateral/stop-limit`
- **Rate Limit**: 10,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `market` | string | Yes | Market name |
| `side` | string | Yes | `buy` or `sell` |
| `amount` | string | Yes | Amount in stock currency |
| `price` | string | Yes | Limit price |
| `activation_price` | string | Yes | Trigger price |
| `stopLoss` | string | No | Stop loss price |
| `takeProfit` | string | No | Take profit price |
| `client_order_id` | string | No | Custom unique identifier |
| `positionSide` | string | No | `LONG`, `SHORT`, or `BOTH` |
| `request` | string | Yes | `/api/v4/order/collateral/stop-limit` |
| `nonce` | string | Yes | Unique incrementing value |

## Place Collateral Trigger Market Order (Stop-Market)

- **Description**: Create a new stop-market order for margin or perpetual futures.
- **Endpoint**: `POST /api/v4/order/collateral/trigger-market`
- **Rate Limit**: 10,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `market` | string | Yes | Market name |
| `side` | string | Yes | `buy` or `sell` |
| `amount` | string | Yes | Amount in stock currency |
| `activation_price` | string | Yes | Trigger price |
| `client_order_id` | string | No | Custom unique identifier |
| `stopLoss` | string | No | Stop loss price |
| `takeProfit` | string | No | Take profit price |
| `positionSide` | string | No | `LONG`, `SHORT`, or `BOTH` |
| `request` | string | Yes | `/api/v4/order/collateral/trigger-market` |
| `nonce` | string | Yes | Unique incrementing value |

## Place Collateral OCO Order

- **Description**: Create a One-Cancels-Other order for margin or perpetual futures.
- **Endpoint**: `POST /api/v4/order/collateral/oco`
- **Rate Limit**: 10,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `market` | string | Yes | Market name |
| `side` | string | Yes | `buy` or `sell` |
| `amount` | string | Yes | Amount |
| `price` | string | Yes | Take profit limit price |
| `activation_price` | string | Yes | Stop loss trigger price |
| `stop_limit_price` | string | Yes | Stop loss limit price |
| `client_order_id` | string | No | Custom unique identifier |
| `request` | string | Yes | `/api/v4/order/collateral/oco` |
| `nonce` | string | Yes | Unique incrementing value |

## Place Collateral Bulk Limit Orders

- **Description**: Place up to 20 limit orders in a single request for margin or perpetual futures.
- **Endpoint**: `POST /api/v4/order/collateral/bulk`
- **Rate Limit**: 10,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `orders` | array | Yes | Array of order objects (up to 20) |
| `stopOnFail` | boolean | No | Default: `false`. Stop processing on first failure |
| `request` | string | Yes | `/api/v4/order/collateral/bulk` |
| `nonce` | string | Yes | Unique incrementing value |

Each order in `orders` array supports: `market`, `side`, `amount`, `price`, `client_order_id`, `stopLoss`, `takeProfit`, `postOnly`, `ioc`, `rpi`, `positionSide`.

### Sample Request

```json
{
  "request": "/api/v4/order/collateral/bulk",
  "nonce": 1594297865000,
  "orders": [
    {
      "market": "BTC_PERP",
      "side": "buy",
      "amount": "0.02",
      "price": "40000",
      "positionSide": "LONG",
      "rpi": true
    },
    {
      "market": "ETH_PERP",
      "side": "sell",
      "amount": "0.5",
      "price": "2500",
      "positionSide": "SHORT",
      "postOnly": true
    }
  ],
  "stopOnFail": true
}
```

## Cancel Order

- **Description**: Cancel an active order.
- **Endpoint**: `POST /api/v4/order/cancel`
- **Rate Limit**: 10,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `market` | string | Yes | Market name |
| `order_id` | integer | No* | Order ID to cancel |
| `client_order_id` | string | No* | Client order ID to cancel |
| `request` | string | Yes | `/api/v4/order/cancel` |
| `nonce` | string | Yes | Unique incrementing value |

*Either `order_id` OR `client_order_id` is required. Do not pass both. `client_order_id` takes priority.

## Cancel Conditional Order

- **Description**: Cancel a conditional order (stop-limit, trigger market, OCO).
- **Endpoint**: `POST /api/v4/order/conditional-cancel`
- **Rate Limit**: 10,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `market` | string | Yes | Market name |
| `id` | integer | Yes | Conditional order ID |
| `request` | string | Yes | `/api/v4/order/conditional-cancel` |
| `nonce` | string | Yes | Unique incrementing value |

## Cancel All Orders

- **Description**: Cancel all active orders, optionally filtered by market.
- **Endpoint**: `POST /api/v4/order/cancel/all`
- **Rate Limit**: 10,000 requests / 10 sec

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `market` | string | No | Market name (omit for all markets) |
| `type` | array | No | Order types: `spot`, `margin`, `futures` |
| `request` | string | Yes | `/api/v4/order/cancel/all` |
| `nonce` | string | Yes | Unique incrementing value |

## Query Order (Active Orders)

- **Description**: Query unexecuted (active) orders for margin or perpetual futures.
- **Endpoint**: `POST /api/v4/orders`
- **Rate Limit**: 12,000 requests / 10 sec

### Parameters

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `market` | string | No | — | Filter by market (e.g., `BTC_PERP`) |
| `order_id` | integer | No | — | Filter by specific order ID |
| `client_order_id` | string | No | — | Filter by custom order ID |
| `offset` | integer | No | 0 | Starting index (max: 10,000) |
| `limit` | integer | No | 50 | Records per page (max: 100) |
| `request` | string | Yes | — | `/api/v4/orders` |
| `nonce` | string | Yes | — | Unique incrementing value |

**Note:** Search across all markets is available only if `client_order_id` and `order_id` are not provided.

## Query Conditional Orders

- **Description**: Query unexecuted conditional orders (stop-limit, trigger market, OCO).
- **Endpoint**: `POST /api/v4/orders/conditional`
- **Rate Limit**: 12,000 requests / 10 sec

### Parameters

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `market` | string | No | — | Filter by market |
| `offset` | integer | No | 0 | Starting index |
| `limit` | integer | No | 50 | Records per page (max: 100) |
| `request` | string | Yes | — | `/api/v4/orders/conditional` |
| `nonce` | string | Yes | — | Unique incrementing value |

## Order History

- **Description**: Query executed order history for margin or perpetual futures.
- **Endpoint**: `POST /api/v4/trade-account/order/history`
- **Rate Limit**: 12,000 requests / 10 sec

### Parameters

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `market` | string | No | — | Filter by market |
| `offset` | integer | No | 0 | Starting index |
| `limit` | integer | No | 50 | Records per page (max: 100) |
| `request` | string | Yes | — | `/api/v4/trade-account/order/history` |
| `nonce` | string | Yes | — | Unique incrementing value |

### Response Format

Returns object with market names as keys, each containing an array of orders:

```json
{
  "BTC_PERP": [
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

## Position Management

### Change Leverage

- **Description**: Set the collateral account leverage level.
- **Endpoint**: `POST /api/v4/collateral-account/leverage`
- **Rate Limit**: 1,000 requests / 10 sec

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `leverage` | integer | Yes | Leverage level (1-100) |
| `request` | string | Yes | `/api/v4/collateral-account/leverage` |
| `nonce` | string | Yes | Unique incrementing value |

### Get Hedge Mode Status

- **Endpoint**: `POST /api/v4/collateral-account/hedge-mode`
- **Rate Limit**: 12,000 requests / 10 sec

### Update Hedge Mode

- **Endpoint**: `POST /api/v4/collateral-account/hedge-mode/update`
- **Rate Limit**: 1,000 requests / 10 sec

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `hedgeMode` | boolean | Yes | Enable/disable hedge mode |
| `request` | string | Yes | `/api/v4/collateral-account/hedge-mode/update` |
| `nonce` | string | Yes | Unique incrementing value |

### Open Positions

- **Endpoint**: `POST /api/v4/collateral-account/positions/open`
- **Rate Limit**: 12,000 requests / 10 sec

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `market` | string | No | Filter by market (e.g., `BTC_PERP`) |
| `request` | string | Yes | `/api/v4/collateral-account/positions/open` |
| `nonce` | string | Yes | Unique incrementing value |

### Close Position

- **Endpoint**: `POST /api/v4/collateral-account/position/close`
- **Rate Limit**: 10,000 requests / 10 sec

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `positionId` | integer | Yes | Position ID to close |
| `market` | string | Yes | Market name |
| `positionSide` | string | No | `LONG`, `SHORT`, or `BOTH` |
| `request` | string | Yes | `/api/v4/collateral-account/position/close` |
| `nonce` | string | Yes | Unique incrementing value |

### Positions History

- **Endpoint**: `POST /api/v4/collateral-account/positions/history`
- **Rate Limit**: 12,000 requests / 10 sec

### Funding History

- **Endpoint**: `POST /api/v4/collateral-account/funding-history`
- **Rate Limit**: 12,000 requests / 10 sec

### Collateral Account Balance

- **Endpoint**: `POST /api/v4/collateral-account/balance`
- **Rate Limit**: 12,000 requests / 10 sec

### Collateral Account Summary

- **Endpoint**: `POST /api/v4/collateral-account/summary`
- **Rate Limit**: 12,000 requests / 10 sec

## Order Types

| Order Type | Endpoint | Key Parameters |
|------------|----------|----------------|
| Limit | `POST /api/v4/order/collateral/limit` | `market`, `side`, `amount`, `price` |
| Market | `POST /api/v4/order/collateral/market` | `market`, `side`, `amount` |
| Stop-Limit | `POST /api/v4/order/collateral/stop-limit` | `market`, `side`, `amount`, `price`, `activation_price` |
| Stop-Market | `POST /api/v4/order/collateral/trigger-market` | `market`, `side`, `amount`, `activation_price` |
| OCO | `POST /api/v4/order/collateral/oco` | `market`, `side`, `amount`, `price`, `activation_price`, `stop_limit_price` |
| Bulk Limit | `POST /api/v4/order/collateral/bulk` | `orders[]`, `stopOnFail` |

## Trade Side

| Value | Description |
|-------|-------------|
| `buy` | Buy order |
| `sell` | Sell order |

## Position Side

| Value | Description |
|-------|-------------|
| `LONG` | Long position (hedge mode) |
| `SHORT` | Short position (hedge mode) |
| `BOTH` | One-way mode (default) |

## Time In Force

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `postOnly` | boolean | `false` | Order is guaranteed to be maker. Cancelled if it would match immediately. |
| `ioc` | boolean | `false` | Immediate-or-Cancel. Executes all or part immediately, cancels unfilled portion. |

**Implicit behavior:** Orders without `postOnly` or `ioc` flags behave as **Good-Til-Cancelled (GTC)**.

## Constraints & Limits

### Rate Limits

| Endpoint Category | Rate Limit |
|-------------------|------------|
| Order placement (all types) | 10,000 requests / 10 sec |
| Order cancellation | 10,000 requests / 10 sec |
| Query orders (active/history) | 12,000 requests / 10 sec |
| Position management | 10,000-12,000 requests / 10 sec |
| Leverage change | 1,000 requests / 10 sec |
| Hedge mode update | 1,000 requests / 10 sec |
| Balance/summary queries | 12,000 requests / 10 sec |

### Key Constraints

- **Unified system:** There is no separate "futures" order API. The collateral trading system handles both margin and perpetual futures. The `market` parameter determines the product type (`BTC_USDT` for margin, `BTC_PERP` for perpetuals).
- **Minimum Order Value:** Total (`amount * price`) must be at least 5.05 USDT equivalent
- **Client Order ID:** Must contain only Latin letters, numbers, dashes, dots, and underscores. Must be unique per account
- **IOC + PostOnly Conflict:** Cannot use `ioc=true` with `postOnly=true` or `rpi=true` (error code 37)
- **RPI Orders:** Retail Price Improvement orders are post-only by design, not visible in public order book feeds
- **Cancel Order Priority:** `client_order_id` takes priority over `order_id`. Do not pass both simultaneously
- **Bulk Orders:** 1-20 orders per request. `stopOnFail` controls failure behavior
- **Position side:** Use `positionSide: "LONG"` or `"SHORT"` when hedge mode is enabled. In one-way mode, use `"BOTH"`.
- **All private endpoints use POST** with JSON body, even for queries.
- **Futures markets list:** Use `GET /api/v4/public/futures` to get available perpetual markets with their `ticker_id` (e.g., `BTC_PERP`), max leverage, funding rates, etc.

## Notes

- WhiteBit's collateral trading system unifies margin and perpetual futures order management under a single set of endpoints.
- Symbol format uses underscore separator: `BTC_USDT` (spot/margin), `BTC_PERP` (perpetual futures).
- Timestamps are in Unix-time format with microseconds (e.g., `1595792396.165973`).
- All numeric values in responses are returned as strings to preserve precision.
- The `stopLoss` and `takeProfit` parameters on limit/market orders create OTO (One-Triggers-Other) conditional orders automatically.
- When rate limit is exceeded, returns HTTP `429 Too Many Requests`. Use exponential backoff: 1s → 2s → 4s → 8s.
