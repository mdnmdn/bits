# Bitget Futures Order REST API Documentation

## Reference

- **Official API Docs**:
  - USDT-M: https://www.bitget.com/api-doc/usdt-futures/trade/
  - Coin-M: https://www.bitget.com/api-doc/coin-futures/trade/
- **Base URL**: `https://api.bitget.com`

## Authentication

All trade endpoints require signed requests with the following headers:

| Header | Description |
|--------|-------------|
| `ACCESS-KEY` | Your API Key |
| `ACCESS-SIGN` | HMAC-SHA256 signature, Base64 encoded |
| `ACCESS-TIMESTAMP` | Unix millisecond timestamp |
| `ACCESS-PASSPHRASE` | Password set when creating API Key |
| `Content-Type` | `application/json` (for POST requests) |
| `locale` | `en-US` or `zh-CN` (optional) |

### Signature Method

**Payload construction:**
- No query string: `timestamp + method.toUpperCase() + requestPath + body`
- With query string: `timestamp + method.toUpperCase() + requestPath + "?" + queryString + body`

Sign with HMAC-SHA256 using `secretKey`, then Base64 encode.

## Unified API Pattern

Bitget uses a **unified V2 API** (`/api/v2/mix/`) for both USDT-M and Coin-M futures. The product type is distinguished via the `productType` parameter, not separate endpoints.

| Product Type | `productType` Value | `marginCoin` |
|--------------|---------------------|--------------|
| USDT-M | `USDT-FUTURES` | `USDT` |
| Coin-M | `COIN-FUTURES` | Underlying coin (e.g., `BTC`) |
| USDC-M | `USDC-FUTURES` | `USDC` |

## Place Order

- **Description**: Place a new futures order (limit or market).
- **Endpoint**: `POST /api/v2/mix/order/place-order`
- **Rate Limit**: 10 requests/second/UID
- **Docs**: https://www.bitget.com/api-doc/usdt-futures/trade/Place-Order

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `symbol` | String | Yes | Trading pair, e.g. `BTCUSDT` |
| `productType` | String | Yes | `USDT-FUTURES`, `COIN-FUTURES`, `USDC-FUTURES` |
| `marginMode` | String | Yes | `isolated` or `crossed` |
| `marginCoin` | String | Yes | Margin coin (capitalized), e.g. `USDT` |
| `size` | String | Yes | Order amount (contract quantity) |
| `price` | String | No | Order price. **Required if `orderType` is `limit`** |
| `side` | String | Yes | `buy` or `sell` |
| `tradeSide` | String | No | **Only required in hedge-mode**: `open` or `close` |
| `orderType` | String | Yes | `limit` or `market` |
| `force` | String | No | Time in force: `gtc` (default), `ioc`, `fok`, `post_only`. **Required for limit orders** |
| `clientOid` | String | No | Custom order ID (idempotent for 30 days) |
| `reduceOnly` | String | No | `YES` or `NO` (default). **Only applicable in one-way position mode** |
| `presetStopSurplusPrice` | String | No | Take-profit trigger price |
| `presetStopLossPrice` | String | No | Stop-loss trigger price |
| `presetStopSurplusExecutePrice` | String | No | Take-profit execution price (0 = market, >0 = limit) |
| `presetStopLossExecutePrice` | String | No | Stop-loss execution price (0 = market, >0 = limit) |
| `stpMode` | String | No | Self-trade prevention: `none` (default), `cancel_taker`, `cancel_maker`, `cancel_both` |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `clientOid` | String | Client order ID |
| `orderId` | String | Order ID |

### Sample Request

```json
{
  "symbol": "BTCUSDT",
  "productType": "USDT-FUTURES",
  "marginMode": "isolated",
  "marginCoin": "USDT",
  "size": "0.01",
  "price": "42000",
  "side": "buy",
  "tradeSide": "open",
  "orderType": "limit",
  "force": "gtc",
  "clientOid": "my-order-001"
}
```

### Sample Response

```json
{
  "code": "00000",
  "msg": "success",
  "requestTime": 1695806875837,
  "data": {
    "clientOid": "my-order-001",
    "orderId": "1234567890123"
  }
}
```

## Batch Place Orders

- **Description**: Place up to 50 orders in a single request.
- **Endpoint**: `POST /api/v2/mix/order/batch-place-order`
- **Rate Limit**: 5 requests/second/UID (1 req/s for copy trading traders)
- **Max batch size**: 50 orders

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `symbol` | String | Yes | Trading pair |
| `productType` | String | Yes | `USDT-FUTURES`, `COIN-FUTURES`, `USDC-FUTURES` |
| `marginMode` | String | Yes | `isolated` or `crossed` |
| `marginCoin` | String | Yes | Margin coin (capitalized) |
| `orderList` | List<Object> | Yes | Order list (max 50) |

**orderList item fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `size` | String | Yes | Order amount |
| `price` | String | No | Required for limit orders |
| `side` | String | Yes | `buy` or `sell` |
| `tradeSide` | String | No | Required in hedge-mode: `open`/`close` |
| `orderType` | String | Yes | `limit` or `market` |
| `force` | String | No | `gtc`, `ioc`, `fok`, `post_only` |
| `clientOid` | String | No | Custom order ID |
| `reduceOnly` | String | No | `YES`/`NO` (one-way mode only) |
| `presetStopSurplusPrice` | String | No | TP trigger price |
| `presetStopLossPrice` | String | No | SL trigger price |
| `stpMode` | String | No | `none`, `cancel_taker`, `cancel_maker`, `cancel_both` |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `successList` | Array | Successfully placed orders (`orderId`, `clientOid`) |
| `failureList` | Array | Failed orders (`orderId`, `clientOid`, `errorCode`, `errorMsg`) |

### Sample Response

```json
{
  "code": "00000",
  "data": {
    "successList": [
      { "orderId": "12345", "clientOid": "order-1" }
    ],
    "failureList": [
      { "orderId": "", "clientOid": "order-2", "errorCode": "40001", "errorMsg": "Insufficient balance" }
    ]
  },
  "msg": "success",
  "requestTime": 1627293504612
}
```

## Query Order

- **Description**: Get order details by orderId or clientOid.
- **Endpoint**: `GET /api/v2/mix/order/detail`
- **Rate Limit**: 10 requests/second/UID

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `symbol` | String | Yes | Trading pair (must be capitalized) |
| `productType` | String | Yes | `USDT-FUTURES`, `COIN-FUTURES`, `USDC-FUTURES` |
| `orderId` | String | No* | Order ID (*one required) |
| `clientOid` | String | No* | Custom order ID (*one required) |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `symbol` | String | Trading pair |
| `size` | String | Order amount |
| `orderId` | String | Order ID |
| `clientOid` | String | Custom order ID |
| `baseVolume` | String | Filled amount |
| `priceAvg` | String | Average fill price |
| `price` | String | Order price |
| `state` | String | `live`, `partially_filled`, `filled`, `canceled` |
| `side` | String | `buy` / `sell` |
| `force` | String | `gtc`, `ioc`, `fok`, `post_only` |
| `totalProfits` | String | Total PnL |
| `posSide` | String | `long`, `short`, `net` |
| `marginCoin` | String | Margin coin |
| `quoteVolume` | String | Quote volume traded |
| `orderType` | String | `limit` / `market` |
| `leverage` | String | Leverage |
| `marginMode` | String | `isolated` / `crossed` |
| `reduceOnly` | String | `YES` / `NO` |
| `enterPointSource` | String | Order source (WEB, API, SYS, ANDROID, IOS) |
| `tradeSide` | String | `open` / `close` (hedge mode) |
| `posMode` | String | `one_way_mode` / `hedge_mode` |
| `orderSource` | String | Order source type |
| `cancelReason` | String | `normal_cancel`, `stp_cancel` |
| `cTime` | String | Creation time (ms) |
| `uTime` | String | Update time (ms) |
| `presetStopSurplusPrice` | String | TP trigger price |
| `presetStopSurplusType` | String | `fill_price` / `mark_price` |
| `presetStopSurplusExecutePrice` | String | TP execution price |
| `presetStopLossPrice` | String | SL trigger price |
| `presetStopLossType` | String | `fill_price` / `mark_price` |
| `presetStopLossExecutePrice` | String | SL execution price |

### Sample Response

```json
{
  "code": "00000",
  "msg": "success",
  "requestTime": 1695865476577,
  "data": [
    {
      "symbol": "BTCUSDT",
      "size": "0.1",
      "orderId": "12345",
      "clientOid": "my-order-1",
      "baseVolume": "0.05",
      "priceAvg": "42100",
      "price": "42000",
      "state": "partially_filled",
      "side": "buy",
      "force": "gtc",
      "totalProfits": "0",
      "posSide": "long",
      "marginCoin": "USDT",
      "quoteVolume": "2105",
      "orderType": "limit",
      "leverage": "10",
      "marginMode": "crossed",
      "reduceOnly": "NO",
      "enterPointSource": "API",
      "tradeSide": "open",
      "posMode": "hedge_mode",
      "cTime": "1627293504612",
      "uTime": "1627293505612"
    }
  ]
}
```

## Cancel Order

- **Description**: Cancel an active order.
- **Endpoint**: `POST /api/v2/mix/order/cancel-order`
- **Rate Limit**: 10 requests/second/UID

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `symbol` | String | Yes | Trading pair |
| `productType` | String | Yes | `USDT-FUTURES`, `COIN-FUTURES`, `USDC-FUTURES` |
| `marginCoin` | String | No | Margin coin (capitalized) |
| `orderId` | String | No* | Order ID (*one required; `orderId` takes precedence) |
| `clientOid` | String | No* | Custom order ID (*one required) |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `orderId` | String | Order ID |
| `clientOid` | String | Client order ID |

### Sample Response

```json
{
  "code": "00000",
  "data": { "orderId": "12345", "clientOid": "" },
  "msg": "success",
  "requestTime": 1627293504612
}
```

## Batch Cancel Orders

- **Description**: Cancel up to 50 orders in a single request.
- **Endpoint**: `POST /api/v2/mix/order/batch-cancel-orders`
- **Rate Limit**: 10 requests/second/UID
- **Max batch size**: 50 orders

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `productType` | String | Yes | `USDT-FUTURES`, `COIN-FUTURES`, `USDC-FUTURES` |
| `symbol` | String | No | Trading pair (required when `orderIdList` is set) |
| `marginCoin` | String | No | Margin coin (capitalized) |
| `orderIdList` | List | No | Order list (max 50). If filled, `symbol` must not be null |

**orderIdList item fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `orderId` | String | No* | Order ID (*one required; `orderId` takes precedence) |
| `clientOid` | String | No* | Custom order ID (*one required) |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `successList` | Array | Successfully canceled orders |
| `failureList` | Array | Failed cancellations (`orderId`, `clientOid`, `errorCode`, `errorMsg`) |

## Open Orders

- **Description**: Get pending (unfilled) orders.
- **Endpoint**: `GET /api/v2/mix/order/orders-pending`
- **Rate Limit**: 10 requests/second/UID

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `productType` | String | Yes | `USDT-FUTURES`, `COIN-FUTURES`, `USDC-FUTURES` |
| `symbol` | String | No | Trading pair |
| `orderId` | String | No | Filter by order ID |
| `clientOid` | String | No | Filter by custom order ID |
| `status` | String | No | `live` (default), `partially_filled` |
| `idLessThan` | String | No | Cursor pagination (use `endId` from previous response) |
| `startTime` | String | No | Start timestamp (ms). Max span: 3 months |
| `endTime` | String | No | End timestamp (ms). Default: 3 months from start |
| `limit` | String | No | Max 100, default 100 |

### Response Fields

Returns `{entrustedList: [...], endId: "..."}` where each order has the same fields as Query Order response.

## History Orders

- **Description**: Get historical orders (filled, canceled).
- **Endpoint**: `GET /api/v2/mix/order/orders-history`
- **Rate Limit**: 10 requests/second/UID
- **Data range**: Up to 90 days

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `productType` | String | Yes | `USDT-FUTURES`, `COIN-FUTURES`, `USDC-FUTURES` |
| `symbol` | String | No | Trading pair |
| `orderId` | String | No | Filter by order ID |
| `clientOid` | String | No | Filter by custom order ID |
| `idLessThan` | String | No | Cursor pagination |
| `orderSource` | String | No | Filter by order source (e.g. `normal`, `liquidation`) |
| `startTime` | String | No | Start timestamp (ms) |
| `endTime` | String | No | End timestamp (ms) |
| `limit` | String | No | Max 100, default 100 |

### Response

Same structure as Open Orders, with additional fields:
- `status`: `filled` or `canceled` (history orders only)
- `liqPrice`: Liquidation price
- `posAvg`: Average position price

## Order Types

| Order Type | Value | `price` Required | `force` Required | Description |
|------------|-------|-------------------|-------------------|-------------|
| Limit | `limit` | Yes | Yes | Order at specified price |
| Market | `market` | No | No | Order at best available price |

## Trade Side (`tradeSide`)

**Only required in hedge mode.**

| Value | Description |
|-------|-------------|
| `open` | Open a new position |
| `close` | Close an existing position |

### Side + TradeSide Mapping (Hedge Mode)

| Action | `side` | `tradeSide` |
|--------|--------|-------------|
| Open Long | `buy` | `open` |
| Close Long | `buy` | `close` |
| Open Short | `sell` | `open` |
| Close Short | `sell` | `close` |

**One-way mode:** `tradeSide` is ignored. Use `buy`/`sell` for `side`.

## Position Modes

| Mode | `posMode` Value | Description |
|------|-----------------|-------------|
| One-way | `one_way_mode` | Single position per symbol (netted) |
| Hedge | `hedge_mode` | Separate long/short positions per symbol |

### Position Direction (`posSide`)

| Value | Description |
|-------|-------------|
| `long` | Hedge mode long position |
| `short` | Hedge mode short position |
| `net` | One-way position |

### `reduceOnly`

- Only applicable in **one-way position mode**
- Values: `YES` / `NO` (default: `NO`)
- In one-way mode, if reduce-only order size exceeds position size, the system cancels existing reduce-only orders sequentially

## Time In Force (`force` parameter)

| Value | Name | Description |
|-------|------|-------------|
| `gtc` | Good Till Canceled | Order stays until filled or canceled (default) |
| `ioc` | Immediate Or Cancel | Fill immediately; cancel unfilled portion |
| `fok` | Fill Or Kill | Fill entirely or cancel entirely |
| `post_only` | Post Only | Must rest on orderbook as maker; cancels if would match immediately |

## Order Status

| Value | Description |
|-------|-------------|
| `live` | Pending match (unfilled) |
| `partially_filled` | Partially filled |
| `filled` | Fully filled |
| `canceled` | Cancelled |

## Self Trade Prevention (STP)

| Value | Description |
|-------|-------------|
| `none` | No STP (default) |
| `cancel_taker` | Cancel taker order |
| `cancel_maker` | Cancel maker order |
| `cancel_both` | Cancel both taker and maker |

## Constraints & Limits

### Rate Limits

| Endpoint | Method | Rate Limit |
|----------|--------|------------|
| Place Order | POST | 10 req/s/UID |
| Batch Place Orders | POST | 5 req/s/UID (1 req/s for copy traders) |
| Cancel Order | POST | 10 req/s/UID |
| Batch Cancel Orders | POST | 10 req/s/UID |
| Get Order Detail | GET | 10 req/s/UID |
| Get Pending Orders | GET | 10 req/s/UID |
| Get History Orders | GET | 10 req/s/UID |

**General rules:**
- Default rate limit for unspecified endpoints: 10 req/s
- Batch operations: 10 orders per currency pair count as 1 request
- Returns HTTP 429 when rate limited
- Rate limits are per UID (not per IP)

### Key Constraints

- **Symbol format:** Use `BTCUSDT` (not `BTCUSDT_UMCBL` which is the V1 format). Case-sensitive.
- **`tradeSide` requirement:** Only required in hedge mode. In one-way mode, it is ignored.
- **`reduceOnly` scope:** Only applicable in one-way position mode.
- **clientOid idempotency:** `clientOid` is idempotent for 30 days.
- **Hedge mode close order behavior:** If a limit close order occupies part of a position and a market close order exceeds remaining size, the system does NOT error. It closes only the remaining quantity.
- **One-way mode reduce-only overflow:** System cancels existing reduce-only orders sequentially (by creation order) until the new order fits.
- **Timestamp validation:** Request timestamp must be within 30 seconds of server time.
- **Margin coin:** Must be capitalized (e.g., `USDT`, `BTC`).
- **Batch limits**: Both batch place and batch cancel support max **50 orders** per request.
- **History data window**: History orders only return data from the past **90 days**.

## USDT-M vs Coin-M Differences

There are **no endpoint differences** between USDT-M and Coin-M futures. Both use the same V2 `/api/v2/mix/` endpoints. The only difference is the `productType` parameter:

| Aspect | USDT-M | Coin-M |
|--------|--------|--------|
| `productType` | `USDT-FUTURES` | `COIN-FUTURES` |
| `marginCoin` | `USDT` | Underlying coin (e.g., `BTC`) |
| Size unit | Contract quantity (USDT value) | Contract quantity (coin value) |

All order types, parameters, rate limits, and response structures are identical.

## Notes

- Bitget uses lowercase enum values (`buy`, `sell`, `limit`, `market`, `gtc`, etc.) unlike Binance/MEXC which use uppercase.
- The `tradeSide` parameter is futures-specific and determines whether an order opens or closes a position (required in hedge mode).
- TP/SL preset is supported on both single and batch order endpoints via `presetStopSurplusPrice`/`presetStopLossPrice` and their execution price variants.
- All numeric values are returned as strings to preserve precision.
- API broker rebate tracking available via `X-CHANNEL-API-CODE` header.
