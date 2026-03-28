# WhiteBit API Documentation

## Overview

- **Official exchange**: https://whitebit.com
- **API Docs**: https://docs.zondacrypto.exchange/reference/introduction (formerly Zonda/WhiteBit)

## Go SDK

- **Official**: `whitebit-exchange/go-sdk`
- **SDK Location**: `/tmp/temp/go-sdk/`
- **Module**: `github.com/whitebit-exchange/go-sdk`

## SDK Structure

```
go-sdk/
├── sdk/
│   └── sdk.go              # Main SDK with all services
├── publicapiv4/
│   └── client.go           # Public API v4 client
├── client/
│   └── client.go           # HTTP client with auth
├── module/
│   ├── market/             # Market data
│   ├── depth/              # Order book
│   ├── tickers/            # Tickers
│   ├── kline/              # Klines/Candles
│   ├── deals/              # Trades
│   └── ...
└── example_ws/            # WebSocket examples
```

## Client Creation

```go
import "github.com/whitebit-exchange/go-sdk/sdk"
import "github.com/whitebit-exchange/go-sdk/publicapiv4"

// Create full SDK with auth
sdk := sdk.New(apiKey, apiSecret)

// Create public API client only
publicClient := publicapiv4.NewClient(nil) // nil uses default options
```

## Base URLs

| Environment | URL |
|-------------|-----|
| Production | `https://whitebit.com` |
| WebSocket | `wss://ws.whitebit.com` |

## Authentication

### Headers

```
X-TXC-APIKEY: <api-key>
X-TXC-SIGNATURE: <signature>
X-TXC-TIMESTAMP: <timestamp>
```

### Signature

```
signature = HMAC-SHA512(secretKey, timestamp + method + path + body)
```

---

## Market Data API (Public v4)

### Get Market Info

```go
markets, err := publicClient.MarketInfo(ctx)
```

**Endpoint**: `GET /api/v4/public/markets`

### Get Market Activity (Ticker)

```go
tickers, err := publicClient.MarketActivity(ctx)
```

**Endpoint**: `GET /api/v4/public/markets`

Returns 24h pricing and volume summary for each market.

### Get Order Book

```go
request := &gosdk.GetAPIV4PublicOrderbookMarketRequest{
    Market: "BTC_USD",
    Limit:  10, // 1-100
}
orderbook, err := publicClient.Orderbook(ctx, request)
```

**Endpoint**: `GET /api/v4/public/orderbook/{market}`

**Response**:
```json
{
  "timestamp": 1690196000000,
  "ask": [
    [50001.0, 0.5],   // [price, amount]
    [50002.0, 1.0]
  ],
  "bid": [
    [49999.0, 0.5],
    [49998.0, 1.0]
  ]
}
```

### Get Depth (Within 2%)

```go
request := &gosdk.GetAPIV4PublicOrderbookDepthMarketRequest{
    Market: "BTC_USD",
    Limit:  10,
}
depth, err := publicClient.Depth(ctx, request)
```

**Endpoint**: `GET /api/v4/public/orderbook/{market}/depth`

Returns price levels within ±2% of market last price.

### Get Recent Trades

```go
request := &gosdk.GetAPIV4PublicTradesMarketRequest{
    Market: "BTC_USD",
    Limit:  100,
}
trades, err := publicClient.RecentTrades(ctx, request)
```

**Endpoint**: `GET /api/v4/public/trades/{market}`

### Get Fee

```go
fees, err := publicClient.Fee(ctx)
```

**Endpoint**: `GET /api/v4/public/fee`

Returns fees and min/max amounts for deposits/withdrawals.

### Get Server Time

```go
time, err := publicClient.ServerTime(ctx)
```

**Endpoint**: `GET /api/v4/public/time`

### Get Collateral Markets

```go
markets, err := publicClient.CollateralMarketsList(ctx)
```

**Endpoint**: `GET /api/v4/public/collateral-markets`

### Get Available Futures Markets

```go
markets, err := publicClient.AvailableFuturesMarketsList(ctx)
```

**Endpoint**: `GET /api/v4/public/futures`

### Get Funding History

```go
request := &gosdk.GetAPIV4PublicFundingHistoryMarketRequest{
    Market: "BTC_USD",
}
funding, err := publicClient.FundingHistory(ctx, request)
```

**Endpoint**: `GET /api/v4/public/funding-history/{market}`

---

## Using SDK Services

### Get Tickers via SDK

```go
tickers, err := sdk.Tickers.GetTickers(ctx)
```

### Get Order Book via SDK

```go
depth, err := sdk.Depth.GetDepth(ctx, "BTC_USD")
```

### Get Klines via SDK

```go
klines, err := sdk.Kline.GetKlines(ctx, "BTC_USD", "1m", 100)
```

---

## WebSocket API

### Connect

```
wss://ws.whitebit.com
```

### Subscribe

```json
{
  "method": "subscribe",
  "params": {
    "channels": ["ticker.BTC_USD"]
  },
  "id": 1
}
```

### Channels

| Channel | Description |
|---------|-------------|
| `ticker.{market}` | Ticker |
| `trades.{market}` | Trades |
| `kline.{market}.{interval}` | Klines |
| `depth.{market}` | Order book |

---

## Rate Limits

- Public: 100 req/s
- Private: 20 req/s

## Products

- **Spot**: Default
- **Futures**: Available
- **Collateral**: Available

## Reference

- [WhiteBit API Docs](https://docs.zondacrypto.exchange/reference/introduction)
- [SDK GitHub](https://github.com/whitebit-exchange/go-sdk)
