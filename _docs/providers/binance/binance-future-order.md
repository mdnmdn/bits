# Binance Futures Order REST API Documentation

## Reference

- **Official API Docs**:
  - USDT-M: https://developers.binance.com/docs/derivatives/usds-margined-futures/trade/rest-api
  - COIN-M: https://developers.binance.com/docs/derivatives/coin-margined-futures/trade
- **Base URLs**:
  - USDT-M: `https://fapi.binance.com`
  - COIN-M: `https://dapi.binance.com`

## Authentication

All order endpoints are `SIGNED` endpoints requiring authentication.

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

- **Description**: Create a new futures order (limit, market, stop, take-profit, trailing stop, etc.)
- **Endpoints**:
  - USDT-M: `POST /fapi/v1/order`
  - COIN-M: `POST /dapi/v1/order`
- **Weight**: 0 (IP), 1 (10s order count), 1 (1min order count)
- **Security Type**: `SIGNED`

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair symbol |
| `side` | ENUM | YES | `BUY` or `SELL` |
| `positionSide` | ENUM | NO | `BOTH` (default, one-way mode), `LONG`, `SHORT` (hedge mode). **Must be sent in hedge mode.** |
| `type` | ENUM | YES | Order type (see Order Types section) |
| `timeInForce` | ENUM | NO | `GTC`, `IOC`, `FOK`, `GTD` |
| `quantity` | DECIMAL | NO | Order quantity. **USDT-M**: base asset amount. **COIN-M**: contract number. Cannot be sent with `closePosition=true`. |
| `reduceOnly` | STRING | NO | `"true"` or `"false"` (default: `"false"`). **Cannot be sent in hedge mode.** Cannot be sent with `closePosition=true`. |
| `price` | DECIMAL | NO | Limit price |
| `newClientOrderId` | STRING | NO | Unique client order ID. Pattern: `^[\.A-Z\:/a-z0-9_-]{1,36}$`. Auto-generated if not sent. |
| `stopPrice` | DECIMAL | NO | Trigger price for STOP/TAKE_PROFIT orders |
| `closePosition` | STRING | NO | `"true"` or `"false"`. Close-All, used with `STOP_MARKET` or `TAKE_PROFIT_MARKET`. **COIN-M only for New Order.** |
| `activationPrice` | DECIMAL | NO | Activation price for trailing stop orders (default: latest price) |
| `callbackRate` | DECIMAL | NO | Trailing stop callback rate (min 0.1, max 10 for COIN-M, max 5 for USDT-M). 1 = 1%. |
| `workingType` | ENUM | NO | `"MARK_PRICE"` or `"CONTRACT_PRICE"` (default: `"CONTRACT_PRICE"`) |
| `priceProtect` | STRING | NO | `"TRUE"` or `"FALSE"` (default: `"FALSE"`). Used with STOP/TAKE_PROFIT orders. |
| `newOrderRespType` | ENUM | NO | `"ACK"` or `"RESULT"` (default: `"ACK"`) |
| `priceMatch` | ENUM | NO | `OPPONENT`, `OPPONENT_5`, `OPPONENT_10`, `OPPONENT_20`, `QUEUE`, `QUEUE_5`, `QUEUE_10`, `QUEUE_20`. Only for LIMIT/STOP/TAKE_PROFIT. Cannot be used with `price`. |
| `selfTradePreventionMode` | ENUM | NO | `NONE`, `EXPIRE_TAKER`, `EXPIRE_MAKER`, `EXPIRE_BOTH`. Default `NONE` (test) / `EXPIRE_MAKER` (live). Only effective with IOC/GTC/GTD. |
| `goodTillDate` | LONG | NO | Mandatory when `timeInForce` = `GTD`. Must be > now + 600s and < 253402300799000. |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Current timestamp in milliseconds |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `clientOrderId` | STRING | Client order ID |
| `cumQty` | STRING | Cumulative filled quantity |
| `cumQuote` | STRING | Cumulative quote (USDT-M only) |
| `cumBase` | STRING | Cumulative base (COIN-M only) |
| `executedQty` | STRING | Executed quantity |
| `orderId` | LONG | Server order ID |
| `avgPrice` | STRING | Average fill price |
| `origQty` | STRING | Original quantity |
| `price` | STRING | Order price |
| `reduceOnly` | BOOL | Reduce-only flag |
| `side` | STRING | BUY or SELL |
| `positionSide` | STRING | BOTH, LONG, or SHORT |
| `status` | STRING | Order status |
| `stopPrice` | STRING | Stop price |
| `closePosition` | BOOL | Close-All flag |
| `symbol` | STRING | Trading pair |
| `timeInForce` | STRING | GTC, IOC, FOK, or GTD |
| `type` | STRING | Order type |
| `origType` | STRING | Original order type |
| `activatePrice` | STRING | Activation price (trailing stop only) |
| `priceRate` | STRING | Callback rate (trailing stop only) |
| `updateTime` | LONG | Update timestamp |
| `workingType` | STRING | MARK_PRICE or CONTRACT_PRICE |
| `priceProtect` | BOOL | Price protect flag |
| `priceMatch` | STRING | Price match mode |
| `selfTradePreventionMode` | STRING | STP mode |
| `goodTillDate` | LONG | GTD auto-cancel timestamp |

### Sample Request (USDT-M)

```bash
curl -X POST "https://fapi.binance.com/fapi/v1/order" \
  -H "X-MBX-APIKEY: your_api_key" \
  -d "symbol=BTCUSDT&side=BUY&type=LIMIT&timeInForce=GTC&quantity=0.001&price=50000&recvWindow=5000&timestamp=1644489390087&signature=<hmac_sha256_sig>"
```

### Sample Response

```json
{
  "clientOrderId": "testOrder",
  "cumQty": "0",
  "cumQuote": "0",
  "executedQty": "0",
  "orderId": 22542179,
  "avgPrice": "0.00000",
  "origQty": "10",
  "price": "0",
  "reduceOnly": false,
  "side": "BUY",
  "positionSide": "SHORT",
  "status": "NEW",
  "stopPrice": "9300",
  "closePosition": false,
  "symbol": "BTCUSDT",
  "timeInForce": "GTD",
  "type": "TRAILING_STOP_MARKET",
  "origType": "TRAILING_STOP_MARKET",
  "updateTime": 1566818724722,
  "workingType": "CONTRACT_PRICE",
  "priceProtect": false,
  "priceMatch": "NONE",
  "selfTradePreventionMode": "NONE",
  "goodTillDate": 1693207680000
}
```

## Test New Order

- **Description**: Creates and validates a new order without sending it to the matching engine.
- **Endpoint**: `POST /fapi/v1/order/test` (USDT-M only; COIN-M has no test endpoint)
- **Weight**: 0 (not counted)
- **Security Type**: `SIGNED`

Accepts all parameters from `POST /fapi/v1/order`. Returns the same response format as a real order would, but the order is NOT submitted to the matching engine.

## Query Order

- **Description**: Check an order's status.
- **Endpoints**:
  - USDT-M: `GET /fapi/v1/order`
  - COIN-M: `GET /dapi/v1/order`
- **Weight**: 1

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair |
| `orderId` | LONG | NO* | Order ID (*one required) |
| `origClientOrderId` | STRING | NO* | Original client order ID (*one required) |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Current timestamp |

**Notes:**
- `orderId` is self-incrementing for each specific `symbol`
- Orders not found if: status is CANCELED/EXPIRED AND no fills AND created + 3 days < now; OR created + 90 days < now

### Response

Returns the same fields as Place Order response.

## Cancel Order

- **Description**: Cancel an active order.
- **Endpoints**:
  - USDT-M: `DELETE /fapi/v1/order`
  - COIN-M: `DELETE /dapi/v1/order`
- **Weight**: 1

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair |
| `orderId` | LONG | NO* | Order ID (*one required) |
| `origClientOrderId` | STRING | NO* | Original client order ID (*one required) |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Current timestamp |

### Response

Returns the same fields as Place Order response with `status: "CANCELED"`.

## Cancel All Open Orders

- **Description**: Cancel all open orders on a symbol.
- **Endpoints**:
  - USDT-M: `DELETE /fapi/v1/allOpenOrders`
  - COIN-M: `DELETE /dapi/v1/allOpenOrders`
- **Weight**: 1

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Current timestamp |

### Sample Response

```json
{
  "code": 200,
  "msg": "The operation of cancel all open order is done."
}
```

## Cancel Multiple Orders

- **Description**: Cancel multiple orders by orderId or clientOrderId.
- **Endpoints**:
  - USDT-M: `DELETE /fapi/v1/batchOrders`
  - COIN-M: `DELETE /dapi/v1/batchOrders`
- **Weight**: 1

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair |
| `orderIdList` | LIST<LONG> | NO* | Max length 10. e.g. `[1234567,2345678]` |
| `origClientOrderIdList` | LIST<STRING> | NO* | Max length 10. e.g. `["my_id_1","my_id_2"]` |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Current timestamp |

*Either `orderIdList` or `origClientOrderIdList` must be sent.

### Response

Returns an array of results — each entry is either a canceled order object or an error object:

```json
[
  {
    "clientOrderId": "myOrder1",
    "cumQty": "0",
    "executedQty": "0",
    "orderId": 283194212,
    "origQty": "11",
    "status": "CANCELED",
    "symbol": "BTCUSDT",
    "side": "BUY",
    "type": "LIMIT"
  },
  {
    "code": -2011,
    "msg": "Unknown order sent."
  }
]
```

## Place Multiple Orders

- **Description**: Place up to 5 orders in a single request.
- **Endpoints**:
  - USDT-M: `POST /fapi/v1/batchOrders`
  - COIN-M: `POST /dapi/v1/batchOrders`
- **Weight**: USDT-M: 5 (10s), 1 (1min), 5 (IP) | COIN-M: 5 (IP)

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `batchOrders` | LIST<JSON> | YES | Order list. Max 5 orders |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Current timestamp |

### Batch Order Item Parameters

| Parameter | Type | Mandatory | Description |
|-----------|------|-----------|-------------|
| `symbol` | STRING | YES | Trading pair |
| `side` | ENUM | YES | BUY or SELL |
| `positionSide` | ENUM | NO | BOTH, LONG, or SHORT |
| `type` | ENUM | YES | Order type |
| `timeInForce` | ENUM | NO | GTC, IOC, FOK, GTD |
| `quantity` | DECIMAL | YES | Order quantity |
| `reduceOnly` | STRING | NO | "true" or "false" |
| `price` | DECIMAL | NO | Limit price |
| `newClientOrderId` | STRING | NO | Client order ID |
| `newOrderRespType` | ENUM | NO | "ACK" or "RESULT" |
| `priceMatch` | ENUM | NO | Price match mode |
| `selfTradePreventionMode` | ENUM | NO | STP mode |
| `goodTillDate` | LONG | NO | USDT-M only |
| `stopPrice` | DECIMAL | NO | COIN-M only |
| `activationPrice` | DECIMAL | NO | COIN-M only |
| `callbackRate` | DECIMAL | NO | COIN-M: min 0.1, max 4 |
| `workingType` | ENUM | NO | COIN-M only |
| `priceProtect` | STRING | NO | COIN-M only |

### Response

Returns an array of results — each entry is either a placed order object or an error object:

```json
[
  {
    "clientOrderId": "testOrder",
    "orderId": 22542179,
    "symbol": "BTCUSDT",
    "side": "BUY",
    "type": "LIMIT",
    "status": "NEW",
    "price": "10001",
    "origQty": "0.001"
  },
  {
    "code": -2022,
    "msg": "ReduceOnly Order is rejected."
  }
]
```

## Open Orders

- **Description**: Get all open orders.
- **Endpoints**:
  - USDT-M: `GET /fapi/v1/openOrders`
  - COIN-M: `GET /dapi/v1/openOrders`
- **Weight**: 1 (single symbol), 40 (all symbols)

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | NO | Trading pair (omit for all symbols) |
| `pair` | STRING | NO | COIN-M only |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Current timestamp |

### Response

Returns an array of open orders (same fields as Place Order response).

## All Orders

- **Description**: Get all orders (active, canceled, filled) with pagination.
- **Endpoints**:
  - USDT-M: `GET /fapi/v1/allOrders`
  - COIN-M: `GET /dapi/v1/allOrders`
- **Weight**: USDT-M: 5 | COIN-M: 20 (with symbol), 40 (with pair)

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `symbol` | STRING | YES (USDT-M) / NO (COIN-M) | Trading pair |
| `pair` | STRING | NO | COIN-M only. Either `symbol` or `pair` must be sent |
| `orderId` | LONG | NO | If set, returns orders >= that orderId |
| `startTime` | LONG | NO | Start timestamp |
| `endTime` | LONG | NO | End timestamp |
| `limit` | INT | NO | USDT-M: default 500, max 1000. COIN-M: default 50, max 100 |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Current timestamp |

**Notes:**
- Query time period must be less than 7 days (default: recent 7 days)
- COIN-M: `pair` can't be sent with `orderId`

### Response

Returns an array of orders (same fields as Place Order response).

## Change Position Mode

- **Description**: Switch between one-way mode and hedge mode.
- **Endpoints**:
  - USDT-M: `POST /fapi/v1/positionSide/dual`
  - COIN-M: `POST /dapi/v1/positionSide/dual`
- **Weight**: 1

### Parameters

| Name | Type | Mandatory | Description |
|------|------|-----------|-------------|
| `dualSidePosition` | STRING | YES | `"true"`: Hedge Mode; `"false"`: One-way Mode |
| `recvWindow` | LONG | NO | Max 60000 |
| `timestamp` | LONG | YES | Current timestamp |

### Sample Response

```json
{ "code": 200, "msg": "success" }
```

**Important:** Position mode change applies to **every symbol**. Cannot change position mode while holding positions.

## Order Types

| Type | Description | Required Parameters |
|------|-------------|---------------------|
| `LIMIT` | Basic limit order | `timeInForce`, `quantity`, `price` |
| `MARKET` | Market order at best available price | `quantity` |
| `STOP` | Stop limit order | `quantity`, `price`, `stopPrice` |
| `TAKE_PROFIT` | Take profit limit order | `quantity`, `price`, `stopPrice` |
| `STOP_MARKET` | Stop market order | `stopPrice` |
| `TAKE_PROFIT_MARKET` | Take profit market order | `stopPrice` |
| `TRAILING_STOP_MARKET` | Trailing stop market order | `callbackRate` (min 0.1, max 10) |

### Order Type Triggering Rules

- **STOP / STOP_MARKET:**
  - BUY: latest price >= `stopPrice`
  - SELL: latest price <= `stopPrice`

- **TAKE_PROFIT / TAKE_PROFIT_MARKET:**
  - BUY: latest price <= `stopPrice`
  - SELL: latest price >= `stopPrice`

- **TRAILING_STOP_MARKET:**
  - BUY: lowest price after placement <= `activationPrice`, and latest price >= lowest price * (1 + `callbackRate`)
  - SELL: highest price after placement >= `activationPrice`, and latest price <= highest price * (1 - `callbackRate`)

## Order Status

| Status | Description |
|--------|-------------|
| `NEW` | Order accepted |
| `PARTIALLY_FILLED` | Partially filled |
| `FILLED` | Fully filled |
| `CANCELED` | Canceled by user |
| `REJECTED` | Order rejected |
| `EXPIRED` | Order expired (e.g., GTD auto-cancel) |
| `EXPIRED_IN_MATCH` | Order expired due to STP |

## Time In Force

| Value | Description |
|-------|-------------|
| `GTC` | Good Till Cancel (default) |
| `IOC` | Immediate Or Cancel |
| `FOK` | Fill Or Kill |
| `GTD` | Good Till Date (requires `goodTillDate` param) |

### GTD Notes

- `goodTillDate` is mandatory when `timeInForce` = `GTD`
- Timestamp retains second-level precision, ms part ignored
- Must be > current time + 600 seconds and < 253402300799000
- In extreme market conditions, auto cancel time might be delayed

## Position Side

| Mode | `positionSide` Values |
|------|----------------------|
| One-way Mode (default) | `BOTH` (default, can be omitted) |
| Hedge Mode | `LONG` or `SHORT` (must be sent) |

## Self Trade Prevention

| Mode | Description |
|------|-------------|
| `NONE` | No self-trade prevention |
| `EXPIRE_TAKER` | Expire taker order when STP triggers |
| `EXPIRE_MAKER` | Expire maker order when STP triggers |
| `EXPIRE_BOTH` | Expire both orders when STP triggers |

Only effective when `timeInForce` is IOC, GTC, or GTD.

## Price Match

| Value | Description |
|-------|-------------|
| `OPPONENT` | Best opponent price |
| `OPPONENT_5` | 5th best opponent price |
| `OPPONENT_10` | 10th best opponent price |
| `OPPONENT_20` | 20th best opponent price |
| `QUEUE` | Best same-side price |
| `QUEUE_5` | 5th best same-side price |
| `QUEUE_10` | 10th best same-side price |
| `QUEUE_20` | 20th best same-side price |

Only for LIMIT/STOP/TAKE_PROFIT orders. Cannot be used with `price`.

## Constraints & Limits

### Rate Limits

#### USDT-M Futures

| Endpoint | Method | 10s Order Count | 1min Order Count | IP Weight |
|----------|--------|-----------------|------------------|-----------|
| `/fapi/v1/order` | POST | 1 | 1 | 0 |
| `/fapi/v1/order/test` | POST | — | — | — |
| `/fapi/v1/batchOrders` | POST | 5 | 1 | 5 |
| `/fapi/v1/order` | DELETE | — | — | 1 |
| `/fapi/v1/batchOrders` | DELETE | — | — | 1 |
| `/fapi/v1/allOpenOrders` | DELETE | — | — | 1 |
| `/fapi/v1/order` | GET | — | — | 1 |
| `/fapi/v1/openOrders` | GET | — | — | 1 (single symbol), 40 (all symbols) |
| `/fapi/v1/allOrders` | GET | — | — | 5 |

#### COIN-M Futures

| Endpoint | Method | 1min Order Count | IP Weight |
|----------|--------|------------------|-----------|
| `/dapi/v1/order` | POST | 1 | 0 |
| `/dapi/v1/batchOrders` | POST | — | 5 |
| `/dapi/v1/order` | DELETE | — | 1 |
| `/dapi/v1/batchOrders` | DELETE | — | 1 |
| `/dapi/v1/allOpenOrders` | DELETE | — | 1 |
| `/dapi/v1/order` | GET | — | 1 |
| `/dapi/v1/openOrders` | GET | — | 1 (single symbol), 40 (multiple) |
| `/dapi/v1/allOrders` | GET | — | 20 (with symbol), 40 (with pair) |

### Key Constraints

- **reduceOnly**: Cannot be sent in hedge mode. Cannot be sent with `closePosition=true`.
- **closePosition**: Only for `STOP_MARKET` or `TAKE_PROFIT_MARKET` orders. Closes **all** current long position (if SELL) or short position (if BUY). Cannot be used with `quantity` or `reduceOnly`. In hedge mode: cannot be used with BUY orders in LONG position side, or SELL orders in SHORT position side.
- **priceProtect**: When `true`, the difference rate between MARK_PRICE and CONTRACT_PRICE cannot exceed the symbol's `triggerProtect` value (from exchange info).
- **TRAILING_STOP_MARKET Error -2021**: If `activationPrice` is wrong side: BUY requires `activationPrice` < latest price; SELL requires `activationPrice` > latest price.
- **Batch orders**: Max 5 orders per request. Orders are processed concurrently, matching order not guaranteed.
- **newClientOrderId pattern**: Must match `^[\.A-Z\:/a-z0-9_-]{1,36}$`.

## USDT-M vs COIN-M Differences

| Aspect | USDT-M (`/fapi/v1/`) | COIN-M (`/dapi/v1/`) |
|--------|----------------------|----------------------|
| **Test endpoint** | `/fapi/v1/order/test` exists | No test endpoint |
| **Quantity unit** | Base asset amount | Contract number |
| **Response cumulative field** | `cumQuote` | `cumBase` |
| **Response has `pair` field** | No | Yes |
| **`closePosition` in New Order** | No | Yes |
| **`stopPrice` in batch order items** | No | Yes |
| **`workingType` in batch order items** | No | Yes |
| **`priceProtect` in batch order items** | No | Yes |
| **`activationPrice` in batch order items** | No | Yes |
| **`callbackRate` max** | 5 (test) | 4 (batch) / 10 (single) |
| **`goodTillDate` in batch** | Yes | No |
| **All Orders `limit` max** | 1000 | 100 |
| **All Orders `limit` default** | 500 | 50 |
| **All Orders requires symbol** | Yes | No (can use `pair`) |

## Notes

- All numeric values in responses are returned as strings to preserve precision.
- Symbols for futures are **uppercase** (e.g., `BTCUSDT`).
- Rate limits are based on **IP address**, not API keys.
- Position mode change applies globally to all symbols and requires no open positions.
- The `workingType` parameter determines whether stop orders trigger on mark price or contract (last) price.
