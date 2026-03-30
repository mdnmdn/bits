# WhiteBit API Documentation

## Overview

- **Official exchange**: https://whitebit.com
- **API Docs**: https://docs.zondacrypto.exchange/reference/introduction
- **Provider ID**: `whitebit`
- **Implementation**: Manual HTTP client (no external SDK)

## Base URLs

| Type | URL |
|------|-----|
| REST API | `https://whitebit.com` |
| WebSocket | `wss://api.whitebit.com/ws` |

## Authentication

Market data endpoints are public (no auth required). Trading endpoints require API key authentication.

### Headers (Private Endpoints)

```
X-TXC-APIKEY: <api-key>
X-TXC-SIGNATURE: <hmac-sha512-signature>
X-TXC-TIMESTAMP: <unix-timestamp>
```

### Signature

```
signature = HMAC-SHA512(secretKey, timestamp + METHOD + path + body)
```

## Rate Limits

| Type | Limit |
|------|-------|
| Public endpoints | 100 req/s |
| Private endpoints | 20 req/s |

## Products

| Product | Market Type | Supported |
|---------|-------------|-----------|
| Spot | `spot` | Yes |
| Futures | `futures` | Yes (perpetual) |
| Margin | N/A | No |

## WebSocket

*Streaming endpoints exist but are not documented here.*

---

## Exchange Info APIs

### Server Time

Returns the server time in Unix seconds.

**Endpoint**: `GET /api/v4/public/time`

**Parameters**: None

**Response**:
```json
{
  "time": 1690196000
}
```

### Markets (Spot)

Returns all spot trading pairs.

**Endpoint**: `GET /api/v4/public/markets`

**Parameters**: None

**Response**:
```json
[
  {
    "name": "BTC_USD",
    "stock": "BTC",
    "money": "USD",
    "stockPrec": "8",
    "moneyPrec": "2",
    "minAmount": "0.00001",
    "minTotal": "1",
    "maxTotal": "1000000",
    "makerFee": "0.001",
    "takerFee": "0.001",
    "tradesEnabled": true
  }
]
```

### Futures Markets

Returns all perpetual futures contracts.

**Endpoint**: `GET /api/v4/public/futures`

**Response**:
```json
{
  "success": true,
  "result": [{
    "ticker_id": "BTC_PERP",
    "stock_currency": "BTC",
    "money_currency": "USDT",
    "brackets": {"1": 100, "10": 50, "20": 30},
    "max_leverage": 100
  }]
}
```

**Notes**: Futures ticker IDs use `_PERP` suffix (e.g. `BTC_USDT` -> `BTC_PERP`).
