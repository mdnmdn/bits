# Crypto.com API Documentation

## Overview

- **Official exchange**: https://crypto.com
- **API Docs**: https://exchange-docs.crypto.com
- **Provider ID**: `cryptocom`
- **Implementation**: Manual HTTP client (no external SDK)
- **API Version**: v2 REST API

## Base URLs

| Environment | URL |
|-------------|-----|
| Production | `https://api.crypto.com/exchange/v1` |
| UAT Sandbox | `https://uat-api.3ona.co/v2` |

## Authentication

Market data endpoints are public (no auth required). Trading endpoints require API key authentication.

### Headers (Private Endpoints)

```
Content-Type: application/json
```

### Signature

```
signature = HMAC-SHA512(secretKey, request_body)
```

## Rate Limits

| Type | Limit |
|------|-------|
| Public endpoints | Not explicitly documented |
| Private endpoints | Rate limited per UID |

## Products

| Product | Market Type | Supported |
|---------|-------------|-----------|
| Spot | `spot` | Yes |
| Margin | `margin` | Yes |
| Derivatives | `futures` | Yes |

## WebSocket

*Streaming endpoints exist but are not documented here.*

---

## Exchange Info APIs

### Instruments

Returns all supported trading instruments.

**Endpoint**: `GET /public/get-instruments`

**Parameters**: None

**Response**:
```json
{
  "id": 1,
  "method": "public/get-instruments",
  "code": 0,
  "result": {
    "instruments": [{
      "instrument_name": "BTC_USDT",
      "product_type": "SPOT",
      "quote_currency": "USDT",
      "base_currency": "BTC",
      "min_order_size": "0.0001",
      "max_order_size": "100",
      "price_precision": 2,
      "quantity_precision": 4
    }]
  }
}
```

### Server Time

Crypto.com v2 REST does **not** expose a dedicated server-time endpoint.
The bits provider estimates server time by measuring round-trip latency of a ticker call.

---

## API Response Envelope

All Crypto.com API responses follow a standard envelope:

```json
{
  "id": <request-id>,
  "method": "<method-name>",
  "code": 0,
  "result": { ... },
  "message": ""
}
```

- `code`: `0` = success, non-zero = error
- `result`: The actual response data (varies by endpoint)
- `message`: Error message (if code != 0)
