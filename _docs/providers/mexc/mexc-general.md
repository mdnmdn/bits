# MEXC API Documentation

## Overview

- **Official exchange**: https://www.mexc.com
- **API Docs**: https://www.mexc.com/api-docs/spot-v3/introduction
- **Provider ID**: `mexc`
- **Implementation**: Manual HTTP client (no external SDK)

## Base URLs

| Type | URL |
|------|-----|
| Spot REST API | `https://api.mexc.com/api/v3` |
| Futures REST API | `https://api.mexc.com/api/v1/contract` |
| Spot WebSocket | `wss://wbs-api.mexc.com/ws` |
| Futures WebSocket | `wss://contract.mexc.com/edge` |

## Authentication

Market data endpoints are public (no auth required). Trading endpoints require API key.

### Headers (Optional for Market Data)

```
X-MEXC-APIKEY: <api-key>
```

## Rate Limits

| Type | Limit |
|------|-------|
| IP-based (Spot) | 300/10s |
| UID-based (Spot) | 500/10s |
| Futures general | 20/2s |
| Futures ticker | 10/2s |
| Futures depth | 10/2s |
| Futures kline | 20/2s |
| WebSocket requests | 100/s |
| WebSocket max subscriptions | 30/connection |

## Products

| Product | Market Type | Supported |
|---------|-------------|-----------|
| Spot | `spot` | Yes |
| Margin | `margin` | Yes (via Spot endpoints) |
| Perpetual Futures | `futures` | Yes |

## Symbol Formats

| Market | Format | Example |
|--------|--------|---------|
| Spot | Uppercase, no separator | `BTCUSDT` |
| Futures | Uppercase, underscore | `BTC_USDT` |

## WebSocket

*Streaming endpoints exist but are not documented here.*

---

## Exchange Info APIs

### Server Time

Returns the server time in milliseconds.

**Endpoint**: `GET /api/v3/time`

**Parameters**: None

**Response**:
```json
{
  "serverTime": 1690196000000
}
```

### Exchange Info (Spot)

Returns trading rules and symbol configuration.

**Endpoint**: `GET /api/v3/exchangeInfo`

**Parameters**: None

**Response** (key fields):
```json
{
  "timezone": "UTC",
  "serverTime": 1690196000000,
  "symbols": [{
    "symbol": "BTCUSDT",
    "status": "ENABLED",
    "baseAsset": "BTC",
    "quoteAsset": "USDT",
    "quotePrecision": 8,
    "baseAssetPrecision": 8,
    "isSpotTradingAllowed": true,
    "isMarginTradingAllowed": true,
    "filters": [
      {"filterType": "PRICE_FILTER", "minPrice": "0.01", "maxPrice": "1000000", "tickSize": "0.01"},
      {"filterType": "LOT_SIZE", "minQty": "0.00001", "maxQty": "9000", "stepSize": "0.00001"}
    ]
  }]
}
```

### Contract Details (Futures)

**Endpoint**: `GET /api/v1/contract/detail`

**Parameters**: None

**Response**:
```json
{
  "success": true,
  "code": 0,
  "data": [{
    "symbol": "BTC_USDT",
    "baseCoin": "BTC",
    "quoteCoin": "USDT",
    "pricePrecision": 2,
    "volUnit": 1,
    "maxLeverage": 200,
    "makerFeeRate": "0.0002",
    "takerFeeRate": "0.0006"
  }]
}
```
