# KuCoin API Documentation

## Overview

- **Official exchange**: https://www.kucoin.com
- **API Docs**: https://docs.kucoin.com

## Go SDK

- **Official**: `Kucoin/kucoin-universal-sdk` - Multi-language SDK including Go
- **SDK Location**: `/tmp/temp/kucoin-universal-sdk/sdk/golang/`
- **Module**: `github.com/Kucoin/kucoin-universal-sdk/sdk/golang`

## SDK Structure

```
sdk/golang/
├── pkg/
│   ├── api/
│   │   ├── api_rest.go       # REST API client interface
│   │   ├── api_ws.go         # WebSocket API client interface
│   │   └── client.go         # Main Client implementation
│   ├── generate/
│   │   └── service/
│   │       ├── spot_api.go       # Spot service
│   │       ├── futures_api.go    # Futures service
│   │       ├── margin_api.go    # Margin service
│   │       └── ...
│   └── types/
└── internal/
    ├── rest/                 # REST implementation
    └── ws/                  # WebSocket implementation
```

## Client Creation

```go
import "github.com/Kucoin/kucoin-universal-sdk/sdk/golang/pkg/api"
import "github.com/Kucoin/kucoin-universal-sdk/sdk/golang/pkg/types"

option := types.ClientOption{
    Key:           "api-key",
    Secret:        "api-secret",
    Passphrase:    "passphrase",
    Environment:   types.Production, // or types.Sandbox
}
client := api.NewClient(&option)
```

## Market Data API (Spot)

### Get Ticker

```go
req := &market.GetTickerReq{Symbol: "BTC-USDT"}
resp, err := client.RestService().GetSpotService().GetMarketAPI().GetTicker(req, ctx)
```

**Endpoint**: `GET /api/v1/market/orderbook/level1`

**Response**:
```json
{
  "code": "200000",
  "data": {
    "symbol": "BTC-USDT",
    "buy": "50000.00",
    "sell": "50001.00",
    "last": "50000.00",
    "vol": "1000.00",
    "volValue": "50000000.00",
    "time": 1690196141868
  }
}
```

### Get All Tickers

```go
resp, err := client.RestService().GetSpotService().GetMarketAPI().GetAllTickers(ctx)
```

**Endpoint**: `GET /api/v1/market/allTickers`

### Get Order Book (Level 2)

```go
req := &market.GetPartOrderBookReq{
    Symbol: "BTC-USDT",
    Size:   "20",
}
resp, err := client.RestService().GetSpotService().GetMarketAPI().GetPartOrderBook(req, ctx)
```

**Endpoint**: `GET /api/v1/market/orderbook/level2_{size}`

### Get Full Order Book

```go
req := &market.GetFullOrderBookReq{
    Symbol: "BTC-USDT",
    Type:   "step0", // step0, step1, step2
}
resp, err := client.RestService().GetSpotService().GetMarketAPI().GetFullOrderBook(req, ctx)
```

**Endpoint**: `GET /api/v3/market/orderbook/level2`

### Get Klines (Candles)

```go
req := &market.GetKlinesReq{
    Symbol:      "BTC-USDT",
    Type:        "1min", // 1min, 5min, 15min, 30min, 1hour, 4hour, 1day
    StartAt:     1690196000,
    EndAt:       1690200000,
}
resp, err := client.RestService().GetSpotService().GetMarketAPI().GetKlines(req, ctx)
```

**Endpoint**: `GET /api/v1/market/candles`

### Get 24h Stats

```go
req := &market.Get24hrStatsReq{Symbol: "BTC-USDT"}
resp, err := client.RestService().GetSpotService().GetMarketAPI().Get24hrStats(req, ctx)
```

**Endpoint**: `GET /api/v1/market/stats`

### Get Trade History

```go
req := &market.GetTradeHistoryReq{
    Symbol: "BTC-USDT",
}
resp, err := client.RestService().GetSpotService().GetMarketAPI().GetTradeHistory(req, ctx)
```

**Endpoint**: `GET /api/v1/market/histories`

### Get Symbols

```go
req := &market.GetAllSymbolsReq{}
resp, err := client.RestService().GetSpotService().GetMarketAPI().GetAllSymbols(req, ctx)
```

**Endpoint**: `GET /api/v2/symbols`

---

## WebSocket API

### Connect

```go
wsService := client.WsService()
// Get public token first
tokenResp, _ := client.RestService().GetSpotService().GetMarketAPI().GetPublicToken(ctx)
// Connect with token
wsService.Connect(ctx, tokenResp.Data.Token)
```

**WebSocket URL**: `wss://ws-api.kucoin.com?token=<token>`

### Subscribe

```go
wsService.Subscribe("/market/ticker:BTC-USDT", handler)
wsService.Subscribe("/market/level2:BTC-USDT", handler)
wsService.Subscribe("/market/candles:BTC-USDT_1min", handler)
```

### Channels

| Channel | Description |
|---------|-------------|
| `/market/ticker:{symbol}` | Ticker |
| `/market/level2:{symbol}` | Order book L2 |
| `/market/level3:{symbol}` | Order book L3 |
| `/market/candles:{symbol}_{type}` | Klines |
| `/market/match:{symbol}` | Trade deals |

---

## Base URLs

| Environment | URL |
|-------------|-----|
| Production | `https://api.kucoin.com` |
| Sandbox | `https://api-sandbox.kucoin.com` |
| WebSocket | `wss://ws-api.kucoin.com` |

## Authentication

### Headers

```
KC-API-KEY: <api-key>
KC-API-SIGN: <signature>
KC-API-TIMESTAMP: <timestamp>
KC-API-PASSPHRASE: <passphrase>
KC-API-KEY-VERSION: <version>
```

### Signature

```
signature = Base64(HMAC-SHA256(secretKey, timestamp + method + path + body))
```

---

## Rate Limits

- Public endpoints: 1800 req/min
- Private endpoints: 300 req/min
- Order placement: 50 req/10s

## Products

- **Spot**: Default
- **Futures**: `GetFuturesService()`
- **Margin**: `GetMarginService()`

## Reference

- [KuCoin API Docs](https://docs.kucoin.com)
- [SDK GitHub](https://github.com/Kucoin/kucoin-universal-sdk)
