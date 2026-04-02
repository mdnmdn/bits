# Crypto.com Exchange Spot Market Order REST API Documentation

## Reference

- **Official API Docs**: https://exchange-docs.crypto.com/exchange/v1/rest-ws/index.html
- **Base URL**: `https://api.crypto.com/exchange/v1`

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

- **Description**: Create a new spot order (limit or market).
- **Endpoint**: `POST /exchange/v1/private/create-order`
- **Method**: `private/create-order`
- **Rate Limit**: 15 requests per 100ms
- **Docs**: https://exchange-docs.crypto.com/exchange/v1/rest-ws/index.html#private-create-order

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `instrument_name` | string | Yes | Trading pair, e.g. `BTC_USDT` |
| `side` | string | Yes | `BUY` or `SELL` |
| `type` | string | Yes | `LIMIT` or `MARKET` |
| `price` | string | Yes for LIMIT | Limit price |
| `quantity` | string | Yes | Order quantity |
| `notional` | string | For MARKET BUY | Amount to spend (only for MARKET BUY) |
| `client_oid` | string | No | Client Order ID (max 36 chars) |
| `exec_inst` | array | No | Execution instructions: `POST_ONLY`, `SMART_POST_ONLY`, `ISOLATED_MARGIN` |
| `time_in_force` | string | No | `GOOD_TILL_CANCEL`, `IMMEDIATE_OR_CANCEL`, `FILL_OR_KILL` |
| `spot_margin` | string | No | `SPOT` (non-margin) or `MARGIN` |
| `stp_scope` | string | No | `M` (master+sub) or `S` (sub only) |
| `stp_inst` | string | No | `M` (cancel maker), `T` (cancel taker), `B` (cancel both) |
| `stp_id` | string | No | 0–32767 |
| `fee_instrument_name` | string | No | Preferred fee token |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `order_id` | string | Newly created order ID |
| `client_oid` | string | Client Order ID (or nonce if not provided) |

### Sample Request

```json
{
  "id": 1,
  "nonce": 1610905028000,
  "method": "private/create-order",
  "api_key": "YOUR_API_KEY",
  "sig": "HMAC_SHA256_SIGNATURE",
  "params": {
    "instrument_name": "BTC_USDT",
    "side": "BUY",
    "type": "LIMIT",
    "price": "50000.5",
    "quantity": "0.01",
    "client_oid": "my-order-001",
    "time_in_force": "GOOD_TILL_CANCEL"
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
    "client_oid": "my-order-001",
    "order_id": "18342311"
  }
}
```

## Query Order

- **Description**: Get detailed order information.
- **Endpoint**: `POST /exchange/v1/private/get-order-detail`
- **Method**: `private/get-order-detail`
- **Rate Limit**: 30 requests per 100ms

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `order_id` | string | Depends | Order ID (string format recommended) |
| `client_oid` | string | Depends | Client Order ID |

> **Note**: Either `order_id` or `client_oid` must be specified.

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `account_id` | string | Account ID |
| `order_id` | string | Order ID |
| `client_oid` | string | Client Order ID |
| `order_type` | string | `MARKET`, `LIMIT` |
| `time_in_force` | string | TIF value |
| `side` | string | `BUY` or `SELL` |
| `exec_inst` | array | Execution instructions |
| `quantity` | string | Original quantity |
| `limit_price` | string | Limit price |
| `order_value` | string | Order value |
| `maker_fee_rate` | string | Maker fee rate |
| `taker_fee_rate` | string | Taker fee rate |
| `avg_price` | string | Average fill price |
| `cumulative_quantity` | string | Filled quantity |
| `cumulative_value` | string | Filled value |
| `cumulative_fee` | string | Total fees |
| `status` | string | Order status |
| `instrument_name` | string | Trading pair |
| `fee_instrument_name` | string | Fee currency |
| `create_time` | number | Creation timestamp (ms) |
| `update_time` | number | Update timestamp (ms) |
| `reason` | number | Rejection reason code (if applicable) |

### Sample Response

```json
{
  "id": 1,
  "method": "private/get-order-detail",
  "code": 0,
  "result": {
    "account_id": "52e7c00f-1324-5a6z-bfgt-de445bde21a5",
    "order_id": "19848525",
    "client_oid": "1613571154900",
    "order_type": "LIMIT",
    "time_in_force": "GOOD_TILL_CANCEL",
    "side": "BUY",
    "exec_inst": [],
    "quantity": "0.0100",
    "limit_price": "50000.0",
    "order_value": "500.000000",
    "maker_fee_rate": "0.000250",
    "taker_fee_rate": "0.000400",
    "avg_price": "0.0",
    "cumulative_quantity": "0.0000",
    "cumulative_value": "0.000000",
    "cumulative_fee": "0.000000",
    "status": "ACTIVE",
    "instrument_name": "BTC_USDT",
    "fee_instrument_name": "USDT",
    "create_time": 1613575617173,
    "update_time": 1613575617173
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
| `order_id` | string | Depends | Order ID (string format recommended) |
| `client_oid` | string | Depends | Client Order ID |

> **Note**: Either `order_id` or `client_oid` must be present.

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `order_id` | string | Order ID |
| `client_oid` | string | Client Order ID |

### Sample Response

```json
{
  "id": 1,
  "method": "private/cancel-order",
  "code": 0,
  "message": "NO_ERROR",
  "result": {
    "client_oid": "my-order-001",
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
| `instrument_name` | string | No | e.g. `BTC_USDT`. Omit to cancel ALL instruments |
| `type` | string | No | `LIMIT`, `TRIGGER`, or `ALL` |

### Sample Request

```json
{
  "id": 1,
  "nonce": 1611169184000,
  "method": "private/cancel-all-orders",
  "api_key": "YOUR_API_KEY",
  "sig": "SIGNATURE",
  "params": {
    "instrument_name": "BTC_USDT"
  }
}
```

### Sample Response

```json
{
  "id": 1,
  "method": "private/cancel-all-orders",
  "code": 0
}
```

## Open Orders

- **Description**: Get all open orders, optionally filtered by instrument.
- **Endpoint**: `POST /exchange/v1/private/get-open-orders`
- **Method**: `private/get-open-orders`
- **Rate Limit**: 3 requests per 100ms

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `instrument_name` | string | No | e.g. `BTC_USDT`. Omit for all instruments |

### Response

Returns `result.data` array of order objects (same fields as Query Order response).

### Sample Response

```json
{
  "id": 1,
  "method": "private/get-open-orders",
  "code": 0,
  "result": {
    "data": [
      {
        "account_id": "52e7c00f-1324-5a6z-bfgt-de445bde21a5",
        "order_id": "19848525",
        "client_oid": "1613571154900",
        "order_type": "LIMIT",
        "time_in_force": "GOOD_TILL_CANCEL",
        "side": "BUY",
        "exec_inst": [],
        "quantity": "0.0100",
        "limit_price": "50000.0",
        "order_value": "500.000000",
        "maker_fee_rate": "0.000250",
        "taker_fee_rate": "0.000400",
        "avg_price": "0.0",
        "cumulative_quantity": "0.0000",
        "cumulative_value": "0.000000",
        "cumulative_fee": "0.000000",
        "status": "ACTIVE",
        "instrument_name": "BTC_USDT",
        "fee_instrument_name": "USDT",
        "create_time": 1613575617173,
        "update_time": 1613575617173
      }
    ]
  }
}
```

## Order History

- **Description**: Get historical orders with pagination.
- **Endpoint**: `POST /exchange/v1/private/get-order-history`
- **Method**: `private/get-order-history`
- **Rate Limit**: 1 request per second

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `instrument_name` | string | No | e.g. `BTC_USDT` |
| `start_time` | number | No | Start timestamp (ms) |
| `end_time` | number | No | End timestamp (ms) |
| `page_size` | number | No | Page size |
| `page` | number | No | Page number |
| `isolation_id` | string | No | Filter by isolated position |

### Response

Returns `result.data` array of order objects (same fields as Query Order response).

## Order Types

| Type | Description | Required Params |
|------|-------------|-----------------|
| `LIMIT` | Execute at specified price or better | `price`, `quantity` |
| `MARKET` | Execute at best available price | `quantity` (SELL) or `notional` (BUY) |

> **Note**: As of 2026-02-20, `STOP_LOSS`, `STOP_LIMIT`, `TAKE_PROFIT`, `TAKE_PROFIT_LIMIT` were **removed** from `private/create-order`. These are now only available via `private/advanced/create-order`.

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

### Execution Instructions (`exec_inst`)

| Value | Description |
|-------|-------------|
| `POST_ONLY` | Ensures order is placed on the book as a maker |
| `SMART_POST_ONLY` | Smart post-only with better price matching |
| `ISOLATED_MARGIN` | Use isolated margin |

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

## Self Trade Prevention

| `stp_scope` | Description |
|-------------|-------------|
| `M` | Master + sub accounts |
| `S` | Sub accounts only |

| `stp_inst` | Description |
|------------|-------------|
| `M` | Cancel maker |
| `T` | Cancel taker |
| `B` | Cancel both |

| `stp_id` | Description |
|----------|-------------|
| `0–32767` | Self-trade prevention group ID |

## Constraints & Limits

### Rate Limits

| Endpoint | Limit |
|----------|-------|
| `private/create-order` | 15 requests per 100ms |
| `private/cancel-order` | 15 requests per 100ms |
| `private/cancel-all-orders` | 15 requests per 100ms |
| `private/get-order-detail` | 30 requests per 100ms |
| `private/get-order-history` | 1 request per second |
| All other private endpoints | 3 requests per 100ms |

> Rate limits are **per API method, per API key**.

### Key Constraints

- **All numbers must be strings** wrapped in double quotes (e.g., `"12.34"`, not `12.34`)
- **Open order limits**: 200 per trading pair, 1000 total per account
- **Client Order ID**: Max 36 characters; recommended to always specify to avoid nonce collisions
- **Async operations**: `create-order`, `cancel-order`, `cancel-all-orders` are asynchronous — response only confirms request queuing
- **Partial fills**: Detected by `status: "ACTIVE"` and `cumulative_quantity > 0`
- **Spot instrument format**: `BASE_QUOTE` (e.g., `BTC_USDT`, `ETH_USDT`)
- **`spot_margin`**: Use `SPOT` for non-margin, `MARGIN` for margin orders

## Notes

- Crypto.com uses a **JSON-RPC-like envelope** for all requests: `{id, nonce, method, api_key, sig, params}`
- Unlike Binance/MEXC, Crypto.com does **not** use HTTP headers for authentication — everything is in the request body.
- Symbol format uses underscore separator: `BTC_USDT` (not `BTCUSDT`).
- All order operations are **asynchronous** — the response confirms the request was queued, not that the order is filled.
- Stop order types were moved to a separate `private/advanced/create-order` endpoint.
