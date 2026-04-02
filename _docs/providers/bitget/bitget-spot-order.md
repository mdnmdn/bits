# Bitget Spot Market Order REST API Documentation

## Reference

- **Official API Docs**: https://www.bitget.com/api-doc/spot/trade/
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

## Place Order

- **Description**: Place a new spot order (limit or market).
- **Endpoint**: `POST /api/v2/spot/trade/place-order`
- **Rate Limit**: 10 req/s/UID (1 req/s for copy trading traders)
- **Docs**: https://www.bitget.com/api-doc/spot/trade/Place-Order

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `symbol` | String | Yes | Trading pair, e.g. `BTCUSDT` |
| `side` | String | Yes | `buy` or `sell` |
| `orderType` | String | Yes | `limit` or `market` |
| `force` | String | Yes | Execution strategy: `gtc`, `post_only`, `fok`, `ioc` (ignored for market orders) |
| `price` | String | No | Limit price (required for limit orders) |
| `size` | String | Yes | For limit & market-sell: base coin amount. For market-buy: quote coin amount. |
| `clientOid` | String | No | Custom order ID (invalid when `tpslType=tpsl`) |
| `tpslType` | String | No | `normal` (default) or `tpsl` |
| `triggerPrice` | String | No | TP/SL trigger price (only for SPOT TP/SL) |
| `stpMode` | String | No | Self-trade prevention: `none`, `cancel_taker`, `cancel_maker`, `cancel_both` |
| `presetTakeProfitPrice` | String | No | Take profit price |
| `executeTakeProfitPrice` | String | No | Take profit execute price |
| `presetStopLossPrice` | String | No | Stop loss price |
| `executeStopLossPrice` | String | No | Stop loss execute price |
| `receiveWindow` | String | No | Valid time window (Unix ms) |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `orderId` | String | Order ID |
| `clientOid` | String | Client order ID |

### Sample Request

```bash
curl -X POST "https://api.bitget.com/api/v2/spot/trade/place-order" \
  -H "ACCESS-KEY:your_api_key" \
  -H "ACCESS-SIGN:your_signature" \
  -H "ACCESS-PASSPHRASE:your_passphrase" \
  -H "ACCESS-TIMESTAMP:1659076670000" \
  -H "Content-Type: application/json" \
  -d '{"symbol":"BTCUSDT","side":"buy","orderType":"limit","force":"gtc","price":"23222.5","size":"1","clientOid":"121211212122"}'
```

### Sample Response

```json
{
  "code": "00000",
  "msg": "success",
  "requestTime": 1695808949356,
  "data": {
    "orderId": "1001",
    "clientOid": "121211212122"
  }
}
```

## Batch Place Orders

- **Description**: Place up to 50 orders in a single request.
- **Endpoint**: `POST /api/v2/spot/trade/batch-orders`
- **Rate Limit**: 5 req/s/UID (1 req/s for traders)
- **Docs**: https://www.bitget.com/api-doc/spot/trade/Batch-Place-Order

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `symbol` | String | No | Trading pair (ignored in `multiple` mode) |
| `batchMode` | String | No | `single` (default, single currency) or `multiple` (cross-currency) |
| `orderList` | Array | Yes | Max 50 orders |

**orderList item fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `symbol` | String | No (required in `multiple` mode) | Trading pair |
| `side` | String | Yes | `buy` or `sell` |
| `orderType` | String | Yes | `limit` or `market` |
| `force` | String | Yes | `gtc`, `post_only`, `fok`, `ioc` |
| `price` | String | No | Limit price |
| `size` | String | Yes | Amount |
| `clientOid` | String | No | Custom order ID |
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
  "msg": "success",
  "requestTime": 1666336231317,
  "data": {
    "successList": [
      { "orderId": "121211212122", "clientOid": "1" }
    ],
    "failureList": [
      { "orderId": "121211212122", "clientOid": "1", "errorCode": "...", "errorMsg": "clientOrderId duplicate" }
    ]
  }
}
```

## Query Order

- **Description**: Get order details by orderId or clientOid.
- **Endpoint**: `GET /api/v2/spot/trade/orderInfo`
- **Rate Limit**: 20 req/s/UID
- **Docs**: https://www.bitget.com/api-doc/spot/trade/Get-Order-Info

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `orderId` | String | No* | Either `orderId` or `clientOid` required |
| `clientOid` | String | No* | Either `clientOid` or `orderId` required |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `userId` | String | Account ID |
| `symbol` | String | Trading pair |
| `orderId` | String | Order ID |
| `clientOid` | String | Custom order ID |
| `price` | String | Order price |
| `size` | String | Amount |
| `orderType` | String | `limit` or `market` |
| `side` | String | `buy` or `sell` |
| `status` | String | `live`, `partially_filled`, `filled`, `cancelled` |
| `priceAvg` | String | Average fill price |
| `baseVolume` | String | Filled quantity (base coin) |
| `quoteVolume` | String | Filled amount (quote coin) |
| `enterPointSource` | String | Client type: `WEB`, `API`, `SYS`, `ANDROID`, `IOS` |
| `orderSource` | String | Order source |
| `feeDetail` | String | Fee breakdown JSON string |
| `cancelReason` | String | `normal_cancel` or `stp_cancel` |
| `cTime` | String | Creation time (Unix ms) |
| `uTime` | String | Update time (Unix ms) |

### Sample Response

```json
{
  "code": "00000",
  "msg": "success",
  "requestTime": 1695865476577,
  "data": [
    {
      "userId": "**********",
      "symbol": "BTCUSDT",
      "orderId": "121211212122",
      "clientOid": "121211212122",
      "price": "0",
      "size": "10.0000000000000000",
      "orderType": "market",
      "side": "buy",
      "status": "filled",
      "priceAvg": "13000.0000000000000000",
      "baseVolume": "0.0007000000000000",
      "quoteVolume": "9.1000000000000000",
      "enterPointSource": "API",
      "feeDetail": "{\"newFees\":{\"c\":0,\"d\":0,\"deduction\":false,\"r\":-0.112079256,\"t\":-0.112079256,\"totalDeductionFee\":0}}",
      "orderSource": "market",
      "cancelReason": "",
      "cTime": "1695865232127",
      "uTime": "1695865233051"
    }
  ]
}
```

## Cancel Order

- **Description**: Cancel an active order.
- **Endpoint**: `POST /api/v2/spot/trade/cancel-order`
- **Rate Limit**: 10 req/s/UID
- **Docs**: https://www.bitget.com/api-doc/spot/trade/Cancel-Order

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `symbol` | String | Yes | Trading pair (not required when `tpslType=tpsl`) |
| `orderId` | String | No* | Either `orderId` or `clientOid` required |
| `clientOid` | String | No* | Either `clientOid` or `orderId` required |
| `tpslType` | String | No | `normal` (default) or `tpsl` |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `orderId` | String | Order ID |
| `clientOid` | String | Client order ID |

### Sample Response

```json
{
  "code": "00000",
  "msg": "success",
  "requestTime": 1234567891234,
  "data": {
    "orderId": "121211212122",
    "clientOid": "xx001"
  }
}
```

## Batch Cancel Orders

- **Description**: Cancel up to 50 orders in a single request.
- **Endpoint**: `POST /api/v2/spot/trade/batch-cancel-order`
- **Rate Limit**: 10 req/s/UID
- **Docs**: https://www.bitget.com/api-doc/spot/trade/Batch-Cancel-Order

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `symbol` | String | No | Trading pair (ignored in `multiple` mode) |
| `batchMode` | String | No | `single` (default) or `multiple` |
| `orderList` | Array | Yes | Max 50 orders |

**orderList item fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `symbol` | String | No | Trading pair (required in `multiple` mode) |
| `orderId` | String | No* | Either `orderId` or `clientOid` required |
| `clientOid` | String | No* | Either `clientOid` or `orderId` required |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `successList` | Array | Successfully canceled orders |
| `failureList` | Array | Failed cancellations (`orderId`, `clientOid`, `errorCode`, `errorMsg`) |

### Important Constraint

For batch cancellation, **you cannot mix `orderId` and `clientOid`** across orders in the same batch. All orders must use the same identifier type.

## Open Orders

- **Description**: Get current unfilled orders.
- **Endpoint**: `GET /api/v2/spot/trade/unfilled-orders`
- **Rate Limit**: 20 req/s/UID
- **Docs**: https://www.bitget.com/api-doc/spot/trade/Get-Current-Orders

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `symbol` | String | No | Trading pair |
| `startTime` | String | No | Start time (Unix ms, past 90 days only) |
| `endTime` | String | No | End time (Unix ms, within 90 days of startTime) |
| `idLessThan` | String | No | Pagination: return data before this orderId |
| `limit` | String | No | Default 100, max 100 |
| `orderId` | String | No | Filter by orderId |
| `tpslType` | String | No | `normal` (default) or `tpsl` |

### Response

Returns array of order objects with same fields as Query Order response.

## History Orders

- **Description**: Get historical orders (filled, canceled).
- **Endpoint**: `GET /api/v2/spot/trade/history-orders`
- **Rate Limit**: 20 req/s/UID
- **Docs**: https://www.bitget.com/api-doc/spot/trade/Get-History-Orders

### Parameters

| Name | Type | Required | Description |
|------|------|-----------|-------------|
| `symbol` | String | No | Trading pair |
| `startTime` | String | No | Start time (Unix ms) |
| `endTime` | String | No | End time (Unix ms) |
| `idLessThan` | String | No | Pagination: return data before this orderId |
| `limit` | String | No | Default 100, max 100 |
| `orderId` | String | No | Filter by orderId |
| `tpslType` | String | No | `normal` (default) or `tpsl` |

### Response

Returns array of order objects with same fields as Query Order response, plus `cancelReason` field.

## Order Types

| Value | Description | Required Params |
|-------|-------------|-----------------|
| `limit` | Limit order | `price` required |
| `market` | Market order | `price` ignored; `size` meaning changes based on side |

**Market order `size` semantics:**
- **Market Buy**: `size` = quote coin amount (e.g., USDT to spend)
- **Market Sell**: `size` = base coin amount (e.g., BTC to sell)

## Trade Side

| Value | Description |
|-------|-------------|
| `buy` | Buy (long) |
| `sell` | Sell (short) |

## Time In Force (`force` parameter)

| Value | Description | Notes |
|-------|-------------|-------|
| `gtc` | Good Till Cancelled | Standard limit order, remains active until filled or cancelled |
| `post_only` | Post Only | Ensures order is placed on the book as a maker; cancelled if it would match immediately |
| `fok` | Fill or Kill | Must be filled entirely immediately or cancelled |
| `ioc` | Immediate or Cancel | Fill as much as possible immediately, cancel remainder |

**Note:** `force` is **ignored** when `orderType` is `market`.

## Order Status

| Value | Description |
|-------|-------------|
| `live` | Pending match (unfilled) |
| `partially_filled` | Partially filled |
| `filled` | Fully filled |
| `cancelled` | Cancelled |

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
| Place Order | POST | 10 req/s/UID (1 req/s for copy traders) |
| Batch Place Orders | POST | 5 req/s/UID (1 req/s for traders) |
| Cancel Order | POST | 10 req/s/UID |
| Batch Cancel Orders | POST | 10 req/s/UID |
| Get Order Info | GET | 20 req/s/UID |
| Get Unfilled Orders | GET | 20 req/s/UID |
| Get History Orders | GET | 20 req/s/UID |

### Key Constraints

- **Batch limits**: Both batch place and batch cancel support max **50 orders** per request.
- **Batch identifier consistency**: In batch cancel, do not mix `orderId` and `clientOid` across orders.
- **History data window**: History orders and unfilled orders only return data from the past **90 days**.
- **TP/SL orders**: When `tpslType=tpsl`, `symbol` is not required for place/cancel, and `clientOid` is invalid for place order.
- **Price/amount precision**: Decimal places and step sizes are defined per symbol via the Get Symbol Info endpoint.
- **All numeric values are returned as strings** to preserve precision.

## Notes

- Bitget uses lowercase enum values (`buy`, `sell`, `limit`, `market`, `gtc`, etc.) unlike Binance/MEXC which use uppercase.
- The `size` parameter has different meanings depending on side for market orders.
- API broker rebate tracking available via `X-CHANNEL-API-CODE` header.
- `receiveWindow` validates that the request is received within the specified time window from the request timestamp.
