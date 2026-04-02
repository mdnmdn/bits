# Bybit API Documentation

## Overview

- **Official exchange**: https://www.bybit.com
- **API Docs**: https://bybit-exchange.github.io/docs/

## Go SDK

- **Official**: `bybit-exchange/bybit.go.api`
- **SDK Location**: `/tmp/temp/bybit.go.api/`
- **Module**: `github.com/bybit-exchange/bybit.go.api`

## SDK Structure

```
bybit.go.api/
├── bybit_api_client.go   # Main HTTP client
├── market_service.go     # Market data endpoints
├── trade.go              # Trading endpoints
├── position.go           # Position endpoints
├── account.go            # Account endpoints
├── websocket.go          # WebSocket client
├── consts.go             # Constants
├── models/               # Data models
│   └── marketResponse.go
└── examples/             # Usage examples
    └── Market/
        ├── market_ticker.go
        ├── market_kline.go
        └── orderbook_info.go
```

## Client Creation

```go
import bybit_connector "github.com/bybit-exchange/bybit.go.api"

// Create client
client := bybit_connector.NewBybitHttpClient(apiKey, apiSecret)

// Or with options
client = bybit_connector.NewBybitHttpClient(
    apiKey, 
    apiSecret,
    bybit_connector.WithBaseURL(bybit_connector.MAINNET),
    bybit_connector.WithDebug(true),
)

// Access market data
marketService := client.NewUtaBybitServiceWithParams(params)
```

## Base URLs

| Environment | URL |
|-------------|-----|
| Production | `https://api.bybit.com` |
| WebSocket | `wss://stream.bybit.com/v5/public` |

## Authentication

### Headers

```
X-BAPI-API-KEY: <api-key>
X-BAPI-SIGN: <signature>
X-BAPI-TIMESTAMP: <timestamp>
X-BAPI-RECV-WINDOW: <window>
```

### Signature

```
sign = HMAC-SHA256(secretKey, timestamp + apiKey + recvWindow + body)
```

---

## Market Data API (V5)

### Get Ticker

```go
params := map[string]interface{}{
    "category": "spot",
    "symbol":   "BTCUSDT",
}
resp, err := client.NewUtaBybitServiceWithParams(params).GetMarketTickers(ctx)
```

**Endpoint**: `GET /v5/market/tickers`

**Categories**: `spot`, `linear`, `inverse`, `option`

**Response**:
```json
{
  "retCode": 0,
  "retMsg": "OK",
  "result": {
    "list": [
      {
        "symbol": "BTCUSDT",
        "lastPrice": "50000",
        "bid1Price": "49999",
        "bid1Size": "0.5",
        "ask1Price": "50001",
        "ask1Size": "0.5",
        "volume24h": "100000",
        "turnover24h": "5000000000"
      }
    ]
  }
}
```

### Get Order Book

```go
params := map[string]interface{}{
    "category": "spot",
    "symbol":   "BTCUSDT",
    "limit":    20,
}
resp, err := client.NewUtaBybitServiceWithParams(params).GetOrderBookInfo(ctx)
```

**Endpoint**: `GET /v5/market/orderbook`

### Get Klines

```go
params := map[string]interface{}{
    "category":  "spot",
    "symbol":    "BTCUSDT",
    "interval":  "1",      // 1, 3, 5, 15, 30, 60, 120, 240, 360, 720, D, W, M
    "startTime": 1690196000000,
    "limit":     100,
}
resp, err := client.NewUtaBybitServiceWithParams(params).GetMarketKline(ctx)
```

**Endpoint**: `GET /v5/market/kline`

### Get Trades

```go
params := map[string]interface{}{
    "category": "spot",
    "symbol":   "BTCUSDT",
    "limit":    100,
}
resp, err := client.NewUtaBybitServiceWithParams(params).GetPublicRecentTrades(ctx)
```

**Endpoint**: `GET /v5/market/recent-trade`

### Get Index Price Klines

```go
params := map[string]interface{}{
    "category": "linear",
    "symbol":   "BTCUSDT",
    "interval":  "1",
}
resp, err := client.NewUtaBybitServiceWithParams(params).GetIndexPriceKline(ctx)
```

**Endpoint**: `GET /v5/market/index-price-kline`

### Get Mark Price Klines

```go
params := map[string]interface{}{
    "category": "linear",
    "symbol":   "BTCUSDT",
    "interval":  "1",
}
resp, err := client.NewUtaBybitServiceWithParams(params).GetMarkPriceKline(ctx)
```

**Endpoint**: `GET /v5/market/mark-price-kline`

### Get Funding Rate History

```go
params := map[string]interface{}{
    "category": "linear",
    "symbol":   "BTCUSDT",
    "limit":    100,
}
resp, err := client.NewUtaBybitServiceWithParams(params).GetFundingRateHistory(ctx)
```

**Endpoint**: `GET /v5/market/funding/history`

### Get Open Interest

```go
params := map[string]interface{}{
    "category": "linear",
    "symbol":   "BTCUSDT",
    "interval":  "15m",
    "limit":     100,
}
resp, err := client.NewUtaBybitServiceWithParams(params).GetOpenInterests(ctx)
```

**Endpoint**: `GET /v5/market/open-interest`

### Get Instrument Info

```go
params := map[string]interface{}{
    "category": "spot",
    "symbol":   "BTCUSDT",
}
resp, err := client.NewUtaBybitServiceWithParams(params).GetInstrumentInfo(ctx)
```

**Endpoint**: `GET /v5/market/instruments-info`

### Get Server Time

```go
resp, err := client.NewUtaBybitServiceNoParams().GetServerTime(ctx)
```

**Endpoint**: `GET /v5/market/time`

---

## WebSocket API (V5)

### Connect (Public)

```go
// Spot: wss://stream.bybit.com/v5/public/spot
// Linear: wss://stream.bybit.com/v5/public/linear
// Inverse: wss://stream.bybit.com/v5/public/inverse
```

### Subscribe

```json
{
  "op": "subscribe",
  "args": ["tickers.BTCUSDT"]
}
```

### Channels

| Channel | Description |
|---------|-------------|
| `tickers.{symbol}` | Ticker |
| `kline.{interval}.{symbol}` | Klines |
| `trade.{symbol}` | Trades |
| `books.{symbol}` | Order book |

---

## Products

- **Spot**: `category=spot`
- **USDT Perpetual**: `category=linear`
- **Inverse**: `category=inverse`
- **Options**: `category=option`

---

## Rate Limits

- Public: 600 req/s
- Private: 60 req/s
- Order placement: 50 req/s

## Reference

- [Bybit API Docs](https://bybit-exchange.github.io/docs/)
- [SDK GitHub](https://github.com/bybit-exchange/bybit.go.api)
