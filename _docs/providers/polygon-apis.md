# Polygon.io (Massive) API Documentation

## Overview

- **Official website**: https://massive.com (formerly polygon.io)
- **API Docs**: https://massive.com/docs
- **Note**: Polygon.io rebranded to Massive.com in Oct 2025, but API still works at both endpoints

## Go SDK

- **Official**: `polygon-io/client-go` (now `massive-com/client-go`)
- **SDK Location**: `/tmp/temp/client-go/`
- **Module**: `github.com/massive-com/client-go/v3`

## SDK Structure

```
client-go/
├── rest/
│   ├── client.go           # Main REST client
│   ├── iterator.go         # Pagination iterator
│   └── gen/
│       └── client.gen.go  # Generated API client (65K+ lines)
├── websocket/
│   ├── client.go          # WebSocket client
│   ├── config.go          # Subscription topics
│   ├── models/
│   │   └── models.go     # WebSocket data models
│   └── subscription.go    # Subscription handling
└── README.md
```

## Client Creation

```go
import "github.com/massive-com/client-go/v3/rest"

// Create client with options
c := rest.NewWithOptions("YOUR_API_KEY",
    rest.WithTrace(false),      // Enable for request/response logging
    rest.WithPagination(true),  // Enable automatic pagination
)

// Or simply
c := rest.New("YOUR_API_KEY")
```

## Base URLs

| Environment | URL |
|-------------|-----|
| Production (new) | `https://api.massive.com` |
| Production (legacy) | `https://api.polygon.io` |

## Authentication

### Header

```
Authorization: Bearer <api-key>
```

Or via environment variable: `MASSIVE_API_KEY`

---

## Crypto Market Data API

### Get Crypto Aggregates (OHLCV)

```go
params := &gen.GetCryptoAggregatesParams{
    Adjusted: rest.Ptr(true),
    Sort:     "asc",
    Limit:    rest.Ptr(100),
}

resp, err := c.GetCryptoAggregatesWithResponse(
    ctx,
    "BTC-USD",     // ticker
    1,             // multiplier
    "day",         // timespan (minute, hour, day, week, month, quarter, year)
    "2024-01-01", // from
    "2024-12-31", // to
    params,
)
```

**Endpoint**: `GET /v2/aggs/ticker/{ticker}/range/{multiplier}/{timespan}/{from}/{to}`

### Get Crypto Trades

```go
params := &gen.GetCryptoTradesParams{
    Limit: rest.Ptr(100),
    Sort:  "asc",
}

resp, err := c.GetCryptoTradesWithResponse(
    ctx,
    "BTC-USD",
    params,
)
```

**Endpoint**: `GET /v3/trades/{cryptoTicker}`

### Get Last Crypto Trade

```go
resp, err := c.GetLastCryptoTradeWithResponse(ctx, "BTC", "USD")
```

**Endpoint**: `GET /v2/last/trade/{from}/{to}`

### Get Crypto Snapshot Ticker

```go
resp, err := c.GetCryptoSnapshotTickerWithResponse(ctx, "BTC-USD")
```

**Endpoint**: `GET /v3/snapshot/locale/global/markets/crypto/tickers/{ticker}`

### Get Crypto Snapshot Tickers

```go
params := &gen.GetCryptoSnapshotTickersParams{
    Limit: rest.Ptr(10),
}

resp, err := c.GetCryptoSnapshotTickersWithResponse(ctx, params)
```

**Endpoint**: `GET /v3/snapshot/locale/global/markets/crypto/tickers`

### Get Crypto Open/Close

```go
resp, err := c.GetCryptoOpenCloseWithResponse(
    ctx,
    "BTC",
    "USD",
    "2024-01-01", // date
    nil,
)
```

**Endpoint**: `GET /v1/open-close/crypto/{from}/{date}`

### Get Grouped Crypto Aggregates

```go
params := &gen.GetGroupedCryptoAggregatesParams{
    Adjusted: rest.Ptr(true),
    Limit:    rest.Ptr(100),
}

resp, err := c.GetGroupedCryptoAggregatesWithResponse(
    ctx,
    "2024-01-01", // date
    params,
)
```

**Endpoint**: `GET /v2/aggs/grouped/locale/global/market/crypto/{date}`

### Get Previous Crypto Aggregates

```go
params := &gen.GetPreviousCryptoAggregatesParams{
    Adjusted: rest.Ptr(true),
    Limit:    rest.Ptr(100),
}

resp, err := c.GetPreviousCryptoAggregatesWithResponse(
    ctx,
    "BTC-USD",
    params,
)
```

**Endpoint**: `GET /v2/aggs/ticker/{ticker}/prev`

---

## Technical Indicators API

### Get Crypto SMA (Simple Moving Average)

```go
params := &gen.GetCryptoSMAParams{
    Timespan:  "day",
    SeriesType: "close",
    Limit:     100,
}

resp, err := c.GetCryptoSMAWithResponse(ctx, "BTC-USD", params)
```

### Get Crypto EMA (Exponential Moving Average)

```go
params := &gen.GetCryptoEMAParams{
    Timespan:  "day",
    SeriesType: "close",
    Limit:     100,
}

resp, err := c.GetCryptoEMAWithResponse(ctx, "BTC-USD", params)
```

### Get Crypto MACD

```go
params := &gen.GetCryptoMACDParams{
    Timespan:   "day",
    SeriesType: "close",
    FastPeriod:  rest.Ptr(12),
    SlowPeriod: rest.Ptr(26),
    SignalPeriod: rest.Ptr(9),
}

resp, err := c.GetCryptoMACDWithResponse(ctx, "BTC-USD", params)
```

### Get Crypto RSI (Relative Strength Index)

```go
params := &gen.GetCryptoRSIParams{
    Timespan:  "day",
    SeriesType: "close",
    Period:    rest.Ptr(14),
    Limit:     100,
}

resp, err := c.GetCryptoRSIWithResponse(ctx, "BTC-USD", params)
```

---

## Crypto Exchanges API

### Get Crypto Exchanges

```go
params := &gen.GetCryptoV1ExchangesParams{
    Limit: rest.Ptr(100),
}

resp, err := c.GetCryptoV1ExchangesWithResponse(ctx, params)
```

**Endpoint**: `GET /v1/meta/crypto-exchanges`

---

## WebSocket API

### Connect

```go
import massivews "github.com/massive-com/client-go/v3/websocket"

c, err := massivews.New(massivews.Config{
    APIKey: "YOUR_API_KEY",
    Feed:   massivews.RealTime,
    Market: massivews.Crypto,
})
```

**WebSocket URL**: `wss://socket.massive.com`

### Subscribe

```go
// Subscribe to crypto trades
c.Subscribe(massivews.CryptoTrades, "BTC-USD", "ETH-USD")

// Subscribe to crypto quotes
c.Subscribe(massivews.CryptoQuotes, "BTC-USD")

// Subscribe to crypto aggregates
c.Subscribe(massivews.CryptoSecAggs, "*")  // Second aggregates
c.Subscribe(massivews.CryptoMinAggs, "*")  // Minute aggregates

// Subscribe to order book (L2)
c.Subscribe(massivews.CryptoL2Book, "BTC-USD")
```

### WebSocket Topics

| Topic | Description |
|-------|-------------|
| `CryptoSecAggs` | Second aggregates |
| `CryptoMinAggs` | Minute aggregates |
| `CryptoTrades` | Trades |
| `CryptoQuotes` | Quotes (bid/ask) |
| `CryptoL2Book` | Level 2 order book |
| `CryptoLaunchpadMinAggs` | Launchpad minute aggregates |
| `CryptoLaunchpadValue` | Launchpad values |

### Receive Data

```go
for {
    select {
    case err := <-c.Error():
        log.Fatal(err)
    case out, more := <-c.Output():
        if !more {
            return
        }
        switch v := out.(type) {
        case models.CryptoTrade:
            fmt.Printf("Trade: %+v\n", v)
        case models.CryptoQuote:
            fmt.Printf("Quote: %+v\n", v)
        case models.CurrencyAgg:
            fmt.Printf("Aggregate: %+v\n", v)
        }
    }
}
```

---

## WebSocket Models

```go
// CryptoTrade
type CryptoTrade struct {
    Ticker             string    // Crypto pair (e.g., "BTC-USD")
    ExchangeID         int       // Exchange ID
    Price              float64   // Trade price
    Size               float64   // Trade size
    ID                 int64     // Trade ID
    ParticipantTimestamp time.Time // Trade timestamp
    TradedDate         string    // Trade date
    TradedTime         string    // Trade time
}

// CryptoQuote
type CryptoQuote struct {
    Ticker             string
    ExchangeID         int
    BidPrice           float64
    BidSize            float64
    AskPrice           float64
    AskSize            float64
    ParticipantTimestamp time.Time
}

// Level2Book
type Level2Book struct {
    Ticker             string
    ExchangeID         int
    Bid                []PriceLevel  // Bids
    Ask                []PriceLevel  // Asks
}

// CurrencyAgg (Aggregates)
type CurrencyAgg struct {
    Ticker             string
    ExchangeID         int
    Open               float64
    High               float64
    Low                float64
    Close              float64
    Volume             float64
    VWAP              float64
    Timestamp          time.Time
}
```

---

## Rate Limits

- **Starter**: 5 requests/minute
- **Developer**: 100 requests/minute
- **Advanced**: 1,000 requests/minute
- **Enterprise**: Custom limits

## Pricing Plans

| Plan | Monthly Cost | API Calls |
|------|-------------|-----------|
| Starter | Free | Limited |
| Developer | $99/mo | 10,000/day |
| Advanced | $499/mo | 100,000/day |
| Enterprise | Custom | Unlimited |

---

## Go SDK Usage Examples

### Get Crypto Price Data

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/massive-com/client-go/v3/rest"
    "github.com/massive-com/client-go/v3/rest/gen"
)

func main() {
    c := rest.New("YOUR_API_KEY")
    ctx := context.Background()

    // Get daily aggregates for BTC
    resp, err := c.GetCryptoAggregatesWithResponse(
        ctx,
        "BTC-USD",
        1,
        "day",
        "2024-01-01",
        "2024-12-31",
        &gen.GetCryptoAggregatesParams{
            Limit: rest.Ptr(365),
        },
    )
    if err != nil {
        log.Fatal(err)
    }
    if err := rest.CheckResponse(resp); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Results: %+v\n", resp.JSON200.Results)
}
```

### Real-time Crypto WebSocket

```go
package main

import (
    "log"
    massivews "github.com/massive-com/client-go/v3/websocket"
    "github.com/massive-com/client-go/v3/websocket/models"
)

func main() {
    c, err := massivews.New(massivews.Config{
        APIKey: "YOUR_API_KEY",
        Feed:   massivews.RealTime,
        Market: massivews.Crypto,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    if err := c.Connect(); err != nil {
        log.Fatal(err)
    }

    // Subscribe to BTC-USD trades
    if err := c.Subscribe(massivews.CryptoTrades, "BTC-USD"); err != nil {
        log.Fatal(err)
    }

    for {
        select {
        case err := <-c.Error():
            log.Fatal(err)
        case out, more := <-c.Output():
            if !more {
                return
            }
            if trade, ok := out.(models.CryptoTrade); ok {
                log.Printf("BTC trade: $%.2f", trade.Price)
            }
        }
    }
}
```

---

## Reference

- [Massive API Docs](https://massive.com/docs)
- [Polygon.io (Legacy) Docs](https://polygon.io/docs)
- [Go SDK GitHub](https://github.com/massive-com/client-go)
- [API Dashboard](https://massive.com/dashboard)
