# OKX API Documentation

## Overview

- **Official exchange**: https://www.okx.com
- **API Docs**: https://www.okx.com/docs-v5/en/

## Go SDK

- **Community**: `tigusigalpa/okx-go` - Full OKX v5 API support
- **SDK Location**: `/tmp/temp/okx-go/`
- **Module**: `github.com/tigusigalpa/okx-go`

## SDK Structure

```
okx-go/
├── okx.go                  # Main client with services
├── client.go               # HTTP client
├── rest/
│   ├── market/            # Market data (public)
│   │   └── market.go      # Ticker, OrderBook, Candles, Trades
│   ├── public/            # Public endpoints
│   ├── trade/             # Trading (private)
│   ├── account/           # Account (private)
│   └── ...
├── websocket.go           # WebSocket client
└── models/
    ├── market.go          # Market data models
    └── ...
```

## Client Creation

```go
import "github.com/tigusigalpa/okx-go"

// Create REST client
client := okx.NewRestClient(apiKey, secretKey, passphrase)

// Access market data (public)
marketClient := client.Market

// Access trading (private)
tradeClient := client.Trade
```

## Base URLs

| Environment | URL |
|-------------|-----|
| Production | `https://www.okx.com` |
| US/AU users | `https://us.okx.com` |
| EU users | `https://eea.okx.com` |
| WebSocket | `wss://ws.okx.com:8443` |

## Authentication

### Headers

```
OK-ACCESS-KEY: <api-key>
OK-ACCESS-SIGN: <signature>
OK-ACCESS-TIMESTAMP: <timestamp>
OK-ACCESS-PASSPHRASE: <passphrase>
```

### Signature

```
sign = Base64(HMAC-SHA256(secretKey, timestamp + method + requestPath + body))
```

Timestamp format: ISO 8601 with milliseconds, e.g., `2020-12-08T09:08:57.715Z`

---

## Market Data API

### Get Ticker

```go
// Single ticker
ticker, err := marketClient.GetTicker(ctx, "BTC-USDT")

// Multiple tickers (by instrument type)
tickers, err := marketClient.GetTickers(ctx, "SPOT", nil, nil)
```

**Endpoint**: `GET /api/v5/market/ticker?instId=BTC-USDT`

**Response**:
```json
[
  {
    "instId": "BTC-USDT",
    "last": "50000.0",
    "lastSz": "0.1",
    "askPx": "50001.0",
    "askSz": "0.5",
    "bidPx": "49999.0",
    "bidSz": "0.5",
    "open24h": "49000.0",
    "high24h": "51000.0",
    "low24h": "48000.0",
    "volCcy24h": "1000000.0",
    "vol24h": "20000.0",
    "sodUtc0": "49000.0",
    "sodUtc8": "49500.0"
  }
]
```

### Get Order Book

```go
// Standard order book
book, err := marketClient.GetOrderBook(ctx, "BTC-USDT", okx.Ptr("20"))

// Full order book
bookFull, err := marketClient.GetOrderBookFull(ctx, "BTC-USDT", okx.Ptr("20"))

// Lite order book (only best bid/ask)
bookLite, err := marketClient.GetOrderBookLite(ctx, "BTC-USDT")
```

**Endpoints**:
- Standard: `GET /api/v5/market/books?instId=BTC-USDT&sz=20`
- Full: `GET /api/v5/market/books-full?instId=BTC-USDT&sz=20`
- Lite: `GET /api/v5/market/books-lite?instId=BTC-USDT`

### Get Candlesticks

```go
// Current candles
candles, err := marketClient.GetCandles(ctx, "BTC-USDT", 
    okx.Ptr("1m"),  // bar
    nil,            // after
    nil,            // before  
    okx.Ptr("100")) // limit

// Historical candles
historyCandles, err := marketClient.GetHistoryCandles(ctx, "BTC-USDT",
    okx.Ptr("1m"),
    nil,
    nil,
    okx.Ptr("100"))
```

**Endpoints**:
- Current: `GET /api/v5/market/candles`
- History: `GET /api/v5/market/history-candles`

**Bar intervals**: `1m`, `3m`, `5m`, `15m`, `30m`, `1h`, `2h`, `4h`, `6h`, `12h`, `1D`, `1W`, `1M`

### Get Trades

```go
// Recent trades
trades, err := marketClient.GetTrades(ctx, "BTC-USDT", okx.Ptr("100"))

// Historical trades
historyTrades, err := marketClient.GetHistoryTrades(ctx, "BTC-USDT", 
    nil,  // type
    nil,  // after
    nil,  // before
    okx.Ptr("100"))
```

**Endpoints**:
- Recent: `GET /api/v5/market/trades`
- History: `GET /api/v5/market/history-trades`

### Get Index Candles

```go
candles, err := marketClient.GetIndexCandles(ctx, "BTC-USDT", okx.Ptr("1m"))
```

**Endpoint**: `GET /api/v5/market/index-candles`

### Get Mark Price Candles

```go
candles, err := marketClient.GetMarkPriceCandles(ctx, "BTC-USDT", okx.Ptr("1m"))
```

**Endpoint**: `GET /api/v5/market/mark-price-candles`

### Get 24h Volume

```go
volume, err := marketClient.Get24hVolume(ctx)
```

**Endpoint**: `GET /api/v5/market/platform-24-volume`

### Get Exchange Rate

```go
rate, err := marketClient.GetExchangeRate(ctx)
```

**Endpoint**: `GET /api/v5/market/exchange-rate`

---

## WebSocket API

### Connect

```go
wsClient, err := okx.NewWebSocketClient(token, false)
```

**WebSocket URL**: `wss://ws.okx.com:8443/ws/v5/public`

### Subscribe

```json
{
  "op": "subscribe",
  "args": [
    {
      "channel": "tickers",
      "instId": "BTC-USDT"
    }
  ]
}
```

### Channels

| Channel | Description |
|---------|-------------|
| `tickers` | Tickers |
| `books` | Order book |
| `candles` | Klines |
| `trades` | Trades |
| `bbo-tickers` | Best bid/ask ticker |

---

## Rate Limits

- Trading: 600 req/2s (UID)
- Market data: 20 req/s

## Products

- **Spot**: `SPOT`
- **Futures**: `SWAP` (perpetual), `FUTURES`
- **Options**: `OPTION`

## Reference

- [OKX API Docs](https://www.okx.com/docs-v5/en/)
- [SDK GitHub](https://github.com/tigusigalpa/okx-go)
