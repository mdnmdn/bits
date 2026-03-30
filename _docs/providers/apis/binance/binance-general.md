# Binance API Documentation

## Overview

- **Official exchange**: https://www.binance.com
- **API Docs**: https://binance-docs.github.io/apidocs/
- **SDK used**: `github.com/adshao/go-binance/v2`
- **Provider ID**: `binance`

## Base URLs

| Environment | REST URL | WebSocket URL |
|-------------|----------|---------------|
| Production | `https://api.binance.com` | `wss://stream.binance.com:9443/stream` |
| Testnet | `https://testnet.binance.vision` | `wss://testnet.binance.vision/stream` |
| Demo | `https://demo-api.binance.com` | N/A |
| Futures (prod) | `https://fapi.binance.com` | `wss://fstream.binance.com/stream` |
| Futures (testnet) | `https://testnet.binancefuture.com` | `wss://stream.binancefuture.com/stream` |

## Authentication

Market data endpoints are public. Trading/account endpoints require API key authentication.

### Headers

```
X-MBX-APIKEY: <api-key>
```

### Signature

```
signature = HMAC-SHA256(secretKey, queryString)
```

## Rate Limits

| Type | Limit |
|------|-------|
| IP request weight | 6000/minute |
| Order rate | 50/10s per account |
| Raw requests | 6100/minute |

## Products

| Product | Market Type | Supported |
|---------|-------------|-----------|
| Spot | `spot` | Yes |
| Margin | `margin` | Yes |
| USDT-M Futures | `futures` | Yes |
| Coin-M Futures | N/A | Via delivery client |
| Options | N/A | Via options client |

## WebSocket

*Streaming endpoints exist but are not documented here.*

---

## Exchange Info APIs

### Server Time

Returns the current server time in milliseconds.

**Endpoint**: `GET /api/v3/time` (spot) / `GET /fapi/v1/time` (futures)

**Parameters**: None

**Response**:
```json
{
  "serverTime": 1499827319559
}
```

### Exchange Info (Spot)

Returns trading rules, symbol filters, and rate limits.

**Endpoint**: `GET /api/v3/exchangeInfo`

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | No | Single symbol |
| symbols | string | No | URL-encoded array of symbols |

**Response** (key fields):
```json
{
  "symbols": [{
    "symbol": "BTCUSDT",
    "status": "TRADING",
    "baseAsset": "BTC",
    "quoteAsset": "USDT",
    "baseAssetPrecision": 8,
    "quotePrecision": 8,
    "filters": [
      {"filterType": "PRICE_FILTER", "minPrice": "0.01", "maxPrice": "1000000", "tickSize": "0.01"},
      {"filterType": "LOT_SIZE", "minQty": "0.00001", "maxQty": "9000", "stepSize": "0.00001"}
    ]
  }]
}
```

### Exchange Info (Futures)

**Endpoint**: `GET /fapi/v1/exchangeInfo`

Same response structure as spot, with additional fields like `contractType`, `deliveryDate`, `onboardDate`.
