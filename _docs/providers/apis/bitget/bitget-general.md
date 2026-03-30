# Bitget API Documentation

## Overview

- **Official exchange**: https://www.bitget.com
- **API Docs**: https://www.bitget.com/api-doc/common/intro
- **Provider ID**: `bitget`
- **Implementation**: Manual HTTP client (no external SDK)

## Base URLs

| Type | URL |
|------|-----|
| REST API | `https://api.bitget.com` |
| WebSocket (Public) | `wss://ws.bitget.com/v2/ws/public` |
| WebSocket (Private) | `wss://ws.bitget.com/v2/ws/private` |

## Authentication

Market data endpoints are public but the implementation sends auth headers on all requests.

### Required Headers

```
ACCESS-KEY: <api-key>
ACCESS-SIGN: <base64-hmac-sha256-signature>
ACCESS-TIMESTAMP: <unix-ms-timestamp>
ACCESS-PASSPHRASE: <api-passphrase>
Content-Type: application/json
```

### Signature

```
signature = Base64(HMAC-SHA256(secretKey, timestamp + METHOD + requestPath + body))
```

## Rate Limits

| Type | Limit |
|------|-------|
| IP request | 20/2s (market data) |
| WebSocket connect | 300/IP/5min |
| WebSocket subscriptions | 240/hour/connection |

## Products

| Product | Market Type | Supported |
|---------|-------------|-----------|
| Spot | `spot` | Yes |
| USDT-M Futures | `futures` | Yes |
| Margin (Isolated) | `margin` | Yes |
| Coin-M Futures | N/A | No |

## WebSocket

*Streaming endpoints exist but are not documented here.*

---

## Exchange Info APIs

### Server Time

Returns the server time in milliseconds as a string.

**Endpoint**: `GET /api/v2/public/time`

**Parameters**: None

**Response**:
```json
{
  "code": "00000",
  "msg": "success",
  "data": {
    "serverTime": "1690196000000"
  }
}
```

### Spot Symbols

Returns all spot trading pairs.

**Endpoint**: `GET /api/v2/spot/public/symbols`

**Parameters**: None

**Response**:
```json
{
  "code": "00000",
  "data": [{
    "symbol": "BTCUSDT",
    "baseCoin": "BTC",
    "quoteCoin": "USDT",
    "status": "online",
    "pricePrecision": "2",
    "quantityPrecision": "6",
    "minTradeAmount": "0.00001",
    "maxTradeAmount": "1000000"
  }]
}
```

### Futures Contracts

Returns all futures contracts.

**Endpoint**: `GET /api/v2/mix/market/contracts`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| productType | string | Yes | `USDT-FUTURES`, `COIN-FUTURES`, `USDC-FUTURES` |

**Response**:
```json
{
  "code": "00000",
  "data": [{
    "symbol": "BTCUSDT",
    "baseCoin": "BTC",
    "quoteCoin": "USDT",
    "symbolStatus": "normal",
    "pricePlace": "1",
    "volumePlace": "2",
    "minTradeNum": "0.01",
    "maxOrderQty": "1000000"
  }]
}
```

### Margin Currencies

Returns margin trading pairs.

**Endpoint**: `GET /api/v2/margin/currencies`

**Response**:
```json
{
  "code": "00000",
  "data": [{
    "symbol": "ETHUSDT",
    "baseCoin": "ETH",
    "quoteCoin": "USDT",
    "status": "online",
    "pricePrecision": "2",
    "quantityPrecision": "4",
    "minTradeUSDT": "5",
    "makerFeeRate": "0.001",
    "takerFeeRate": "0.001",
    "isBorrowable": true
  }]
}
```
