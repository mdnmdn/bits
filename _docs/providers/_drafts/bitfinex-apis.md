# Bitfinex API Documentation

## Overview

- **Official exchange**: https://www.bitfinex.com
- **API Docs**: https://docs.bitfinex.com/docs/introduction

## Go SDK

- **Official**: `bitfinexcom/bitfinex-api-go`
- **SDK Location**: `/tmp/temp/bitfinex-api-go/`
- **Module**: `github.com/bitfinexcom/bitfinex-api-go/v2`

## SDK Structure

```
v2/
├── rest/
│   ├── client.go         # Main REST client
│   ├── tickers.go        # Ticker service
│   ├── book.go           # Order book service
│   ├── candles.go        # Candles service
│   ├── trades.go         # Trades service
│   ├── currencies.go     # Currencies service
│   ├── platform_status.go # Platform status
│   └── ...
├── websocket/
│   ├── client.go         # WebSocket client
│   ├── channels.go       # Channel handlers
│   └── orderbook.go      # Order book handling
└── pkg/
    ├── models/
    │   ├── ticker/        # Ticker models
    │   ├── book/         # Order book models
    │   ├── candle/       # Candle models
    │   └── ...
    └── convert/          # Data converters
```

## Client Creation

```go
import "github.com/bitfinexcom/bitfinex-api-go/v2/rest"

// Create client (public)
client := rest.NewClient()

// Create client with authentication
client = rest.NewClient().Credentials(apiKey, apiSecret)
```

## Base URLs

| Environment | URL |
|-------------|-----|
| Production REST | `https://api-pub.bitfinex.com/v2/` |
| Production WS | `wss://api-pub.bitfinex.com/ws/2` |
| Auth WS | `wss://api.bitfinex.com/ws/2` |

## Authentication (REST)

### Headers

```
BFX-APIKEY: <api-key>
BFX-SIGNATURE: <signature>
BFX-PAYLOAD: <payload>
```

### Signature

```
signature = HMAC-SHA384(secretKey, payload)
payload = Base64(JSON.stringify({...}))
```

---

## Market Data API (Public)

### Get Ticker

```go
// Single ticker
ticker, err := client.Tickers.Get("tBTCUSD")

// Multiple tickers
tickers, err := client.Tickers.GetMulti([]string{"tBTCUSD", "tETHUSD"})

// All tickers
tickers, err := client.Tickers.All()
```

**Endpoint**: `GET /v2/tickers?symbols=ALL` or `GET /v2/ticker/{symbol}`

**Response**:
```json
[
  "tBTCUSD",
  45000.0,
  1234.5,
  1690196000000,
  45100.0,
  45200.0,
  44900.0,
  45050.0,
  567.89,
  12345.67
]
```

### Get Order Book

```go
import "github.com/bitfinexcom/bitfinex-api-go/v2/pkg/models/common"

// Precision levels: P0, P1, P2, P3, P4 (price precision)
// Or F0, F1, F2, F3, F4 (full price precision)
book, err := client.Book.All("tBTCUSD", common.BookPrecisionP0, 25)
```

**Endpoint**: `GET /v2/book/{symbol}/{precision}?len={limit}`

**Precisions**:
- P0-P4: Aggregated (0-4 decimal places)
- F0-F4: Full (0-4 decimal places)

**Response**:
```json
[
  [45000.5, 1.5, 100],  // [price, count, amount]
  [45001.0, 2.0, 50],
  ...
]
```

### Get Candles

```go
import "github.com/bitfinexcom/bitfinex-api-go/v2/pkg/models/common"

// Get last candle
candle, err := client.Candles.Last("tBTCUSD", common.CandleResolution1m)

// Get history
candles, err := client.Candles.History("tBTCUSD", common.CandleResolution1m)

// Get history with time range
candles, err := client.Candles.HistoryWithQuery(
    "tBTCUSD",
    common.CandleResolution1m,
    1690196000000,  // start
    1690200000,    // end
    100,           // limit
    1,             // sort order
)
```

**Endpoint**: `GET /v2/candles/trade:{resolution}:{symbol}/{period}`

**Resolutions**:
| Resolution | Code |
|------------|------|
| 1m | `1m` |
| 5m | `5m` |
| 15m | `15m` |
| 30m | `30m` |
| 1h | `1h` |
| 3h | `3h` |
| 6h | `6h` |
| 12h | `12h` |
| 1D | `1D` |
| 1W | `1W` |
| 2W | `2W` |
| 1M | `1M` |

**Response**:
```json
[
  [1690196000000, 45000, 45100, 44900, 45000, 1000.5],
  // [timestamp, open, high, low, close, volume]
]
```

### Get Trades

```go
// Recent trades
trades, err := client.Trades.Get("tBTCUSD", 100)

// Trades with time range
trades, err := client.Trades.GetWithParams("tBTCUSD", map[string]string{
    "limit": "100",
    "start": "1690196000000",
    "end": "1690200000000",
})
```

**Endpoint**: `GET /v2/trades/{symbol}/hist`

**Response**:
```json
[
  [123456, 1690196000000, 45000.5, 1.5],
  // [id, timestamp, price, amount]
]
```

### Get Platform Status

```go
status, err := client.Platform.Status()
```

**Endpoint**: `GET /v2/platform/status`

---

## WebSocket API

### Connect (Public)

```go
import "github.com/bitfinexcom/bitfinex-api-go/v2/websocket"

ws, err := websocket.NewClient(nil, nil).Connect(context.Background())
```

**WebSocket URL**: `wss://api-pub.bitfinex.com/ws/2`

### Subscribe

```go
ws.Subscribe(websocket.NewTickerChannel("tBTCUSD"))
ws.Subscribe(websocket.NewBookChannel("tBTCUSD", websocket.PrecisionP0, 25))
ws.Subscribe(websocket.NewCandlesChannel("tBTCUSD", "1m"))
ws.Subscribe(websocket.NewTradesChannel("tBTCUSD"))
```

### Channels

| Channel | Description |
|---------|-------------|
| `ticker` | Ticker |
| `book` | Order book |
| `trades` | Trades |
| `candles` | Klines |

---

## Rate Limits

- Public: 60 req/min
- Authenticated: 90 req/min

## Products

- **Spot**: `t{symbol}` prefix (e.g., `tBTCUSD`)
- **Futures**: `t{symbol}:{currency}` prefix (e.g., `tBTCF0:USTF0`)
- **Loan**: `l{symbol}` prefix

## Reference

- [Bitfinex API Docs](https://docs.bitfinex.com/docs/introduction)
- [SDK GitHub](https://github.com/bitfinexcom/bitfinex-api-go)
