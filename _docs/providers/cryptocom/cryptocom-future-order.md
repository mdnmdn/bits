# Crypto.com Exchange Futures/Perpetual Order REST API Documentation

## Reference

- **Official API Docs**: https://exchange-docs.crypto.com/exchange/v1/rest-ws/index.html
- **Base URL**: `https://api.crypto.com/exchange/v1`

## Unified Order System

Crypto.com Exchange API v1 handles **Spot, Margin, and Derivatives (futures/perpetuals)** through the same endpoints. The `instrument_name` parameter determines the market type — there are no separate futures-specific order endpoints.

### Symbol Format

| Type | Format | Example |
|------|--------|---------|
| Spot | `BASE_QUOTE` | `BTC_USDT`, `CRO_USD` |
| Perpetual | `BASEQUOTE-PERP` | `BTCUSD-PERP`, `ETHUSDT-PERP` |
| Futures (dated) | `BASEQUOTE-YYMMDD` | `BTCUSD-231124` |

## Authentication

All order endpoints are private and require HMAC-SHA256 signature in the request body.

### Required Body Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | number | Yes | Request ID (matched in response) |
| `nonce` | number | Yes | Millisecond timestamp |
| `method` | string | Yes | API method name (e.g., `private/create-order`) |
| `api_key` | string | Yes | Your API key |
| `sig` | string | Yes | HMAC-SHA256 signature |
| `params` | object | Yes | Method parameters |

### Signature Method (HMAC-SHA256)

```
signature = HMAC_SHA256(method + id + api_key + paramsString + nonce, secret_key)
```

Where `paramsString` is the JSON-serialized params object (or empty string if no params).

## Place Order

- **Description**: Create a new order for spot, margin, or perpetual futures.
- **Endpoint**: `POST /exchange/v1/private/create-order`
- **Method**: `private/create-order`
- **Rate Limit**: 15 requests per 100ms

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `instrument_name` | string | Yes | e.g. `BTC_USDT` (spot), `BTCUSD-PERP` (perpetual) |
| `side` | string | Yes | `BUY` or `SELL` |
| `type` | string | Yes | `LIMIT` or `MARKET` |
| `price` | string | Yes for LIMIT | Limit price |
| `quantity` | string | Yes | Order quantity |
| `notional` | string | For MARKET BUY | Amount to spend (only for MARKET BUY) |
| `client_oid` | string | No | Client Order ID (max 36 chars) |
| `exec_inst` | array | No | `POST_ONLY`, `SMART_POST_ONLY`, `ISOLATED_MARGIN` |
| `time_in_force` | string | No | `GOOD_TILL_CANCEL`, `IMMEDIATE_OR_CANCEL`, `FILL_OR_KILL` |
| `spot_margin` | string | No | `SPOT` or `MARGIN` (not for futures) |
| `stp_scope` | string | No | `M` (Master/Sub) or `S` (Sub only) |
| `stp_inst` | string | No | `M` (Cancel Maker), `T` (Cancel Taker), `B` (Cancel Both) |
| `stp_id` | string | No | 0 to 32767 |
| `fee_instrument_name` | string | No | Preferred fee token (e.g. `USD`, `USDT`, `EUR`) |
| `isolation_id` | string | No | For isolated margin positions |
| `leverage` | string | No | Max leverage for isolated position |
| `isolated_margin_amount` | string | No | Amount to transfer to isolated position |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `order_id` | string | Newly created order ID |
| `client_oid` | string | Client Order ID (or nonce if not provided) |

### Sample Request (Perpetual)

```json
{
  "id": 1,
  "nonce": 1610905028000,
  "method": "private/create-order",
  "api_key": "YOUR_API_KEY",
  "sig": "HMAC_SHA256_SIGNATURE",
  "params": {
    "instrument_name": "BTCUSD-PERP",
    "side": "SELL",
    "type": "LIMIT",
    "price": "50000.5",
    "quantity": "1",
    "client_oid": "c5f682ed-7108-4f1c-b755-972fcdca0f02",
    "exec_inst": ["POST_ONLY"],
    "time_in_force": "FILL_OR_KILL"
  }
}
```

### Sample Response

```json
{
  "id": 1,
  "method": "private/create-order",
  "code": 0,
  "result": {
    "client_oid": "c5f682ed-7108-4f1c-b755-972fcdca0f02",
    "order_id": "18342311"
  }
}
```

## Cancel Order

- **Description**: Cancel an active order.
- **Endpoint**: `POST /exchange/v1/private/cancel-order`
- **Method**: `private/cancel-order`
- **Rate Limit**: 15 requests per 100ms

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `order_id` | string/number | Depends | Order ID (either `order_id` or `client_oid` required) |
| `client_oid` | string | Depends | Client Order ID |

### Sample Response

```json
{
  "id": 1,
  "method": "private/cancel-order",
  "code": 0,
  "message": "NO_ERROR",
  "result": {
    "client_oid": "c5f682ed-7108-4f1c-b755-972fcdca0f02",
    "order_id": "18342311"
  }
}
```

## Cancel All Orders

- **Description**: Cancel all open orders, optionally filtered by instrument.
- **Endpoint**: `POST /exchange/v1/private/cancel-all-orders`
- **Method**: `private/cancel-all-orders`
- **Rate Limit**: 15 requests per 100ms

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `instrument_name` | string | No | e.g. `BTCUSD-PERP`. Omit to cancel ALL instruments |
| `type` | string | No | `LIMIT`, `TRIGGER`, or `ALL` |

## Query Order

- **Description**: Get detailed order information.
- **Endpoint**: `POST /exchange/v1/private/get-order-detail`
- **Method**: `private/get-order-detail`
- **Rate Limit**: 30 requests per 100ms

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `order_id` | string | Depends | Order ID (either `order_id` or `client_oid` required) |
| `client_oid` | string | Depends | Client Order ID |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `account_id` | string | Account ID |
| `order_id` | string | Order ID |
| `client_oid` | string | Client Order ID |
| `order_type` | string | `MARKET`, `LIMIT`, `STOP_LOSS`, `STOP_LIMIT`, `TAKE_PROFIT`, `TAKE_PROFIT_LIMIT` |
| `time_in_force` | string | TIF value |
| `side` | string | `BUY` or `SELL` |
| `exec_inst` | array | `POST_ONLY`, `SMART_POST_ONLY`, `LIQUIDATION`, `ISOLATED_MARGIN` |
| `quantity` | string | Order quantity |
| `limit_price` | string | Limit price |
| `order_value` | string | Order value |
| `maker_fee_rate` | string | Maker fee rate |
| `taker_fee_rate` | string | Taker fee rate |
| `avg_price` | string | Average fill price |
| `cumulative_quantity` | string | Filled quantity |
| `cumulative_value` | string | Filled value |
| `cumulative_fee` | string | Total fees |
| `status` | string | `REJECTED`, `CANCELED`, `FILLED`, `EXPIRED` |
| `instrument_name` | string | e.g. `BTCUSD-PERP` |
| `fee_instrument_name` | string | Fee currency |
| `isolation_id` | string | Isolated position ID (if applicable) |
| `isolation_type` | string | e.g. `ISOLATED_MARGIN` |
| `reason` | number | Rejection reason code |
| `create_time` | number | Creation timestamp (ms) |
| `create_time_ns` | string | Creation timestamp (ns) |
| `update_time` | number | Update timestamp (ms) |

## Open Orders

- **Description**: Get all open orders, optionally filtered by instrument.
- **Endpoint**: `POST /exchange/v1/private/get-open-orders`
- **Method**: `private/get-open-orders`
- **Rate Limit**: 3 requests per 100ms

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `instrument_name` | string | No | e.g. `BTCUSD-PERP`. Omit for all instruments |

### Response

Returns `result.data` array of order objects (same fields as Query Order response), plus:
- `order_date` (string) - Creation date
- Status values: `NEW`, `PENDING`, `ACTIVE`

## Order History

- **Description**: Get historical orders with pagination.
- **Endpoint**: `POST /exchange/v1/private/get-order-history`
- **Method**: `private/get-order-history`
- **Rate Limit**: 1 request per second

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `instrument_name` | string | No | e.g. `BTCUSD-PERP`. Omit for all |
| `start_time` | number/string | No | Unix timestamp (inclusive). Default: end_time - 1 day |
| `end_time` | number/string | No | Unix timestamp (exclusive). Default: current time |
| `limit` | int | No | Max records. Default: 100, Max: 100 |
| `isolation_id` | string | No | Filter by isolated position |

## Close Position

- **Description**: Close a futures/margin position.
- **Endpoint**: `POST /exchange/v1/private/close-position`
- **Method**: `private/close-position`

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `instrument_name` | string | Yes | e.g. `BTCUSD-PERP` |
| `type` | string | Yes | `LIMIT` or `MARKET` |
| `price` | string | Depends | Required for LIMIT orders |
| `quantity` | string | No | Omit for full close, provide for partial |
| `isolation_id` | string | No | Required for isolated positions |

### Sample Request

```json
{
  "id": 1,
  "nonce": 1610905028000,
  "method": "private/close-position",
  "api_key": "YOUR_API_KEY",
  "sig": "SIGNATURE",
  "params": {
    "instrument_name": "BTCUSD-PERP",
    "type": "MARKET"
  }
}
```

## Position Management

### Change Account Leverage

- **Method**: `private/change-account-leverage`

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `account_id` | string | Yes | Account ID |
| `leverage` | number | Yes | 1-100 (inclusive) |

### Change Isolated Margin Leverage

- **Method**: `private/change-isolated-margin-leverage`

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `isolation_id` | string | Yes | Isolated position ID |
| `leverage` | number | Yes | Max leverage for isolated position |

### Get Positions

- **Method**: `private/get-positions`

Returns all open positions with fields including `isolation_id`, `isolation_type`, `leverage`, `liquidation_price`, `isolated_margin_balance`.

### Isolated Margin Transfer

- **Method**: `private/create-isolated-margin-transfer`

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `isolation_id` | string | Yes | Isolated position ID |
| `direction` | string | Yes | `CREDIT` or `DEBIT` |
| `amount` | string | Yes | USD amount to transfer |

## Order Types

| Type | Description | Futures Support |
|------|-------------|-----------------|
| `LIMIT` | Execute at specified price or better | Yes |
| `MARKET` | Execute at best available price | Yes |

### Execution Instructions (`exec_inst`)

| Value | Description |
|-------|-------------|
| `POST_ONLY` | Ensure order is posted to book (maker only) |
| `SMART_POST_ONLY` | Smart post-only that adjusts price |
| `ISOLATED_MARGIN` | Order for isolated margin position |

## Trade Side

| Value | Description |
|-------|-------------|
| `BUY` | Buy order (bid) |
| `SELL` | Sell order (ask) |

## Time In Force

| Value | Description |
|-------|-------------|
| `GOOD_TILL_CANCEL` | Order remains active until filled or canceled |
| `IMMEDIATE_OR_CANCEL` | Execute immediately; cancel any unfilled portion |
| `FILL_OR_KILL` | Execute fully or cancel entirely |

**Constraints**:
- When `exec_inst` contains `POST_ONLY`, `time_in_force` can **only** be `GOOD_TILL_CANCEL`
- `POST_ONLY` and `SMART_POST_ONLY` cannot coexist in `exec_inst`

## Order Status

| Status | Description |
|--------|-------------|
| `NEW` | Order accepted |
| `PENDING` | Order pending |
| `ACTIVE` | Order active in orderbook |
| `CANCELED` | Order canceled |
| `FILLED` | Order fully filled |
| `EXPIRED` | Order expired |
| `REJECTED` | Order rejected |

## Constraints & Limits

### Rate Limits

| Method | Limit |
|--------|-------|
| `private/create-order` | 15 requests per 100ms |
| `private/cancel-order` | 15 requests per 100ms |
| `private/cancel-all-orders` | 15 requests per 100ms |
| `private/get-order-detail` | 30 requests per 100ms |
| `private/get-order-history` | 1 request per second |
| All other private methods | 3 requests per 100ms |

> Rate limits are **per API method, per API key**.

### Key Constraints

- **Unified API:** Exchange API v1 handles Spot, Margin, and Derivatives (futures/perpetuals) through the same endpoints. The `instrument_name` determines the market type.
- **All numbers must be strings** wrapped in double quotes (e.g., `"12.34"`, not `12.34`)
- **Open order limits**: 200 per trading pair, 1000 total per account
- **Client Order ID**: Max 36 characters; recommended to always specify to avoid nonce collisions
- **Async operations**: `create-order`, `cancel-order`, `cancel-all-orders` are asynchronous — response only confirms request queuing
- **Partial fills**: Detected by `status: "ACTIVE"` and `cumulative_quantity > 0`
- **Perpetual instrument format**: `BASEQUOTE-PERP` (e.g., `BTCUSD-PERP`, `ETHUSDT-PERP`)
- **Margin modes**: Cross margin (default) vs Isolated margin (via `isolation_id`, `exec_inst: ["ISOLATED_MARGIN"]`)

## Notes

- Crypto.com uses a **JSON-RPC-like envelope** for all requests: `{id, nonce, method, api_key, sig, params}`
- Unlike Binance/MEXC, Crypto.com does **not** use HTTP headers for authentication — everything is in the request body.
- All order operations are **asynchronous** — the response confirms the request was queued, not that the order is filled.
- Use `user.order.{instrument_name}` WebSocket channel for real-time order status updates.
