# MEXC Futures Order REST API Documentation

## Reference

- **Official API Docs**: https://www.mexc.com/api-docs/futures-v2/introduction
- **Base URL**: `https://api.mexc.com/api/v1/private/`

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

- **Description**: Create a new futures order (limit, market, post-only, IOC, FOK).
- **Endpoint**: `POST /api/v1/private/order/submit`
- **Rate Limit**: 10 times / 2 seconds
- **Permission**: Order Placing

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `symbol` | string | true | Contract name (e.g., `BTC_USDT`) |
| `price` | decimal | true | Order price (not required for market orders) |
| `vol` | decimal | true | Order quantity (contracts) |
| `leverage` | int | false | Leverage (required when opening position) |
| `side` | int | true | 1: open long, 2: close short, 3: open short, 4: close long |
| `type` | int | true | Order type (see Order Types section) |
| `openType` | int | true | 1: isolated, 2: cross |
| `externalOid` | string | false | External order ID (client-generated) |
| `positionId` | int | false | Position ID |
| `stopLossPrice` | decimal | false | Stop-loss price |
| `takeProfitPrice` | decimal | false | Take-profit price |
| `lossTrend` | int | false | SL price type: 1: latest (default), 2: fair, 3: index |
| `profitTrend` | int | false | TP price type: 1: latest (default), 2: fair, 3: index |
| `priceProtect` | int | false | Trigger protection: 0: off, 1: on |
| `positionMode` | int | false | 1: dual-side (default), 2: one-way |
| `reduceOnly` | boolean | false | Reduce-only (one-way mode only) |
| `marketCeiling` | boolean | false | 100% market open |
| `flashClose` | boolean | false | Flash close |
| `bboTypeNum` | int | false | BBO type: 0: none, 1: opp-1, 2: opp-5, 3: same-1, 4: same-5 |
| `stpMode` | number | false | Self-trade prevention: 0: none, 1: cancel both, 2: cancel maker, 3: cancel taker |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `orderId` | string | Order ID |
| `ts` | long | Timestamp |

### Sample Request

```bash
curl -H "X-MEXC-APIKEY: your_api_key" \
  -X POST 'https://api.mexc.com/api/v1/private/order/submit' \
  -d 'symbol=BTC_USDT&price=50000&vol=1&side=1&type=1&openType=1&leverage=10&timestamp=1644489390087&signature=<hmac_sha256_sig>'
```

### Sample Response

```json
{
  "success": true,
  "code": 0,
  "data": {
    "orderId": "739113577038255616",
    "ts": 1761888808839
  }
}
```

## Query Order

### By Order ID

- **Endpoint**: `GET /api/v1/private/order/get/{orderId}`
- **Rate Limit**: 20 times / 2 seconds

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `orderId` | string | true | Order ID |

### By External Order ID

- **Endpoint**: `GET /api/v1/private/order/external/{symbol}/{external_oid}`
- **Rate Limit**: 20 times / 2 seconds

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `symbol` | string | true | Contract |
| `external_oid` | string | true | External order ID |

### Batch Query by Order IDs

- **Endpoint**: `GET /api/v1/private/order/batch_query`
- **Rate Limit**: 5 times / 2 seconds

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `order_ids` | long | true | Comma-separated order IDs (max 50) |

### Response Fields (all query endpoints)

| Field | Type | Description |
|-------|------|-------------|
| `orderId` | long | Order ID |
| `symbol` | string | Contract |
| `positionId` | long | Position ID |
| `price` | decimal | Order price |
| `vol` | decimal | Order quantity |
| `leverage` | long | Leverage |
| `side` | int | 1: open long, 2: close short, 3: open short, 4: close long |
| `category` | int | 1: limit, 2: liquidation custody, 3: custody close, 4: ADL reduction |
| `orderType` | int | 1: limit, 2: post only, 3: IOC, 4: FOK, 5: market |
| `dealAvgPrice` | decimal | Average fill price |
| `dealVol` | decimal | Filled quantity |
| `orderMargin` | decimal | Order margin |
| `takerFee` | decimal | Taker fee |
| `makerFee` | decimal | Maker fee |
| `profit` | decimal | Close PnL |
| `feeCurrency` | string | Fee currency |
| `openType` | int | 1: isolated, 2: cross |
| `state` | int | 1: pending, 2: unfilled, 3: filled, 4: canceled, 5: invalid |
| `externalOid` | string | External order ID |
| `createTime` | long | Created time (ms) |
| `updateTime` | long | Updated time (ms) |
| `positionMode` | int | 1: dual-side, 2: one-way |
| `reduceOnly` | boolean | Reduce-only flag |
| `bboTypeNum` | int | BBO type |
| `stopLossPrice` | decimal | Stop-loss price |
| `takeProfitPrice` | decimal | Take-profit price |
| `errorCode` | int | Error code (0 = normal) |

### Sample Response

```json
{
  "success": true,
  "code": 0,
  "data": {
    "orderId": 739113577038255616,
    "symbol": "BTC_USDT",
    "positionId": 123456,
    "price": "50000",
    "vol": "1",
    "leverage": 10,
    "side": 1,
    "category": 1,
    "orderType": 1,
    "dealAvgPrice": "0",
    "dealVol": "0",
    "orderMargin": "500",
    "takerFee": "0.0006",
    "makerFee": "0.0002",
    "profit": "0",
    "feeCurrency": "USDT",
    "openType": 1,
    "state": 2,
    "externalOid": "ext_11",
    "createTime": 1644489390087,
    "updateTime": 1644489390087,
    "positionMode": 1,
    "reduceOnly": false,
    "bboTypeNum": 0,
    "stopLossPrice": "0",
    "takeProfitPrice": "0",
    "errorCode": 0
  }
}
```

## Cancel Orders

### Batch Cancel by Order IDs

- **Endpoint**: `POST /api/v1/private/order/cancel`
- **Rate Limit**: 20 times / 2 seconds

Body: List of order IDs (max 50)

### Cancel by External Order ID

- **Endpoint**: `POST /api/v1/private/order/cancel_with_external`
- **Rate Limit**: 20 times / 2 seconds

Body: `[{"symbol": "BTC_USDT", "externalOid": "ext_11"}]`

### Batch Cancel by External Order ID

- **Endpoint**: `POST /api/v1/private/order/batch_cancel_with_external`
- **Rate Limit**: 20 times / 2 seconds

Body: List of `{symbol, externalOid}` pairs

### Response

```json
{
  "success": true,
  "code": 0,
  "data": [
    { "orderId": 108886241042563584, "errorCode": 0, "errorMsg": "success" },
    { "orderId": 101716841474621953, "errorCode": 2040, "errorMsg": "order not exist" }
  ]
}
```

## Cancel All Orders

- **Endpoint**: `POST /api/v1/private/order/cancel_all`
- **Rate Limit**: 20 times / 2 seconds

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `symbol` | string | false | Contract (omit to cancel all contracts) |

## Open Orders

- **Endpoint**: `GET /api/v1/private/order/list/open_orders`
- **Rate Limit**: 20 times / 2 seconds

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `page_num` | int | true | Page number (default 1) |
| `page_size` | int | true | Page size (default 20, max 100) |

Returns paginated list of all active orders across all contracts.

## Order History

- **Endpoint**: `GET /api/v1/private/order/list/history_orders`
- **Rate Limit**: 20 times / 2 seconds

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `symbol` | string | false | Contract filter |
| `states` | string | false | Comma-separated states: 1,2,3,4,5 |
| `category` | int | false | 1: limit, 2: liquidation, 3: custody close, 4: ADL |
| `startTime` | long | false | Start time (ms) |
| `endTime` | long | false | End time (ms) |
| `orderId` | long | false | Order ID filter |
| `page_num` | int | true | Page number (default 1) |
| `page_size` | int | true | Page size (default 20, max 100) |

Response is paginated with `pageSize`, `totalCount`, `totalPage`, `currentPage`, `resultList`.

## Order Types

| Value | Name | Description |
|-------|------|-------------|
| 1 | Limit | Standard limit order |
| 2 | Post Only | Maker-only, rejected if would match immediately |
| 3 | IOC | Immediate-Or-Cancel |
| 4 | FOK | Fill-Or-Kill |
| 5 | Market | Market order |

## Position Side / Mode

### Side Values (used in order placement)

| Value | Meaning |
|-------|---------|
| 1 | Open Long |
| 2 | Close Short |
| 3 | Open Short |
| 4 | Close Long |

### Position Modes

| Value | Mode | Description |
|-------|------|-------------|
| 1 | Dual-side (Hedge) | Separate long/short positions (default) |
| 2 | One-way | Single net position per contract |

**Important constraints:**
- `reduceOnly` is only applicable in one-way mode
- To switch position modes, you must have no active orders, plan orders, or open positions
- Switching from dual-side to one-way resets risk limit to level 1
- In dual-side mode, `side` explicitly determines open vs close direction
- In one-way mode with `reduceOnly=true`, orders can only reduce position size

### Open Type (Margin Mode)

| Value | Type |
|-------|------|
| 1 | Isolated |
| 2 | Cross |

## Time In Force

TIF is encoded in the `type` parameter:

| TIF | Type Value | Behavior |
|-----|------------|----------|
| GTC (Good-Til-Canceled) | 1 | Standard limit order |
| Post Only | 2 | Maker-only, cancels if would match |
| IOC | 3 | Fill immediately, cancel remainder |
| FOK | 4 | Fill entirely or cancel |
| Market | 5 | Immediate market execution |

## Self Trade Prevention (STP)

| Value | Behavior |
|-------|----------|
| 0 | No self-trade prevention (default) |
| 1 | Cancel both sides |
| 2 | Cancel maker order |
| 3 | Cancel taker order |

## Constraints & Limits

### Rate Limits

| Endpoint | Limit |
|----------|-------|
| Place Order (`/submit`) | 10 req / 2s |
| Batch Place (`/submit_batch`) | 10 req / 2s (market makers only) |
| Cancel Orders (`/cancel`) | 20 req / 2s |
| Cancel All (`/cancel_all`) | 20 req / 2s |
| Query Order (`/get/{id}`, `/external/...`) | 20 req / 2s |
| Batch Query (`/batch_query`) | 5 req / 2s |
| Open Orders (`/list/open_orders`) | 20 req / 2s |
| History Orders (`/list/history_orders`) | 20 req / 2s |

### Error Codes (in order response)

| Code | Description |
|------|-------------|
| 0 | Normal |
| 1 | Parameter error |
| 2 | Insufficient balance |
| 3 | Position not found |
| 4 | Insufficient available position |
| 5 | Price vs liquidation price conflict |
| 6 | Liquidation price vs fair price conflict |
| 7 | Exceeds risk limit |

### Key Constraints

- **External Order ID**: Must start with a custom prefix. Orders with `externalOid` starting with `CHASE_LIMIT` indicate chase-limit orders.
- **Batch Place**: Restricted to market maker accounts only.
- **Symbol format**: Uses underscore separator: `BTC_USDT` (not `BTCUSDT`).
- **Max open orders** per contract is returned in contract info as `maxNumOrders: [hedged_max, one_way_max]`.
- **All numeric values in responses are returned as strings** to preserve precision.

## Notes

- MEXC futures uses **integer enums** for order types, sides, and position modes — unlike spot which uses string enums.
- The `side` parameter combines both direction (buy/sell) and action (open/close) into a single integer value.
- Time in force is encoded in the `type` parameter rather than as a separate field.
- The official docs label most order endpoints as "Under Maintenance" — this is a documentation artifact from the 2022-07-25 maintenance event. The endpoints are functional.
- STP and Batch Order APIs were added as of the 2026-01-19 update log.
