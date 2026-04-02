# Crypto.com API Documentation

## Overview

- **Official exchange**: https://crypto.com
- **API Docs**: https://exchange-docs.crypto.com

## Go SDK

- **Community**: `cshep4/crypto-dot-com-exchange-go`
- **SDK Location**: `/tmp/temp/crypto-dot-com-exchange-go/`
- **Module**: `github.com/cshep4/crypto-dot-com-exchange-go`

## SDK Structure

```
crypto-dot-com-exchange-go/
├── client.go               # Main client interface
├── getbook.go              # Order book
├── gettrades.go            # Trades
├── getticker.go            # Ticker
├── getinstruments.go        # Instruments/Symbols
├── createorder.go          # Order creation
├── cancelorder.go          # Order cancellation
├── getopenorders.go        # Open orders
├── getorderhistory.go      # Order history
├── internal/
│   ├── api/               # HTTP requester
│   ├── auth/              # Authentication
│   └── mocks/             # Test mocks
└── errors/                # Error handling
```

## Client Creation

```go
import cdcexchange "github.com/cshep4/crypto-dot-com-exchange-go"

// Create client (public)
client, err := cdcexchange.NewClient(
    cdcexchange.WithTimeout(30 * time.Second),
)

// Create client with auth
client, err = cdcexchange.NewClient(
    cdcexchange.WithAPIKey("api-key"),
    cdcexchange.WithSecretKey("secret-key"),
    cdcexchange.WithTimeout(30 * time.Second),
)
```

## Base URLs

| Environment | URL |
|-------------|-----|
| Production | `https://api.crypto.com/v2/` |
| UAT Sandbox | `https://uat-api.3ona.co/v2/` |

## Authentication

The SDK uses request signing with HMAC-SHA512:

```go
// Implemented internally via internal/auth/auth.go
// Uses API key and secret key for signed requests
```

---

## Market Data API (Public)

### Get Instruments

```go
instruments, err := client.GetInstruments(ctx)
```

**Endpoint**: `public/get-instruments`

Returns all supported instruments (e.g., BTC_USDT).

**Response**:
```json
{
  "id": 1,
  "method": "public/get-instruments",
  "code": 0,
  "result": {
    "instruments": [
      {
        "instrument_name": "BTC_USDT",
        "product_type": "SPOT",
        "quote_currency": "USDT",
        "base_currency": "BTC",
        "min_order_size": "0.0001",
        "max_order_size": "100",
        "price_precision": 2,
        "quantity_precision": 4
      }
    ]
  }
}
```

### Get Order Book

```go
book, err := client.GetBook(ctx, "BTC_USDT", 10)
```

**Endpoint**: `public/get-book`

Parameters:
- `instrument`: Instrument name (e.g., "BTC_USDT")
- `depth`: Depth level (1-50)

**Response**:
```json
{
  "id": 1,
  "method": "public/get-book",
  "code": 0,
  "result": {
    "instrument_name": "BTC_USDT",
    "bids": [[50000.0, 1.5], ...],
    "asks": [[50001.0, 0.5], ...]
  }
}
```

### Get Tickers

```go
// Single ticker
tickers, err := client.GetTickers(ctx, "BTC_USDT")

// All tickers
tickers, err := client.GetTickers(ctx, "")
```

**Endpoint**: `public/get-ticker`

**Response**:
```json
{
  "id": 1,
  "method": "public/get-ticker",
  "code": 0,
  "result": {
    "data": [
      {
        "i": "BTC_USDT",
        "v": 1000000,
        "vv": 50000000000,
        "l": 49000,
        "h": 51000,
        "o": 49000,
        "c": 50000,
        "p": 1000,
        "t": 10000
      }
    ]
  }
}
```

Fields:
- `i`: Instrument name
- `v`: Volume (24h)
- `vv`: Volume value (24h)
- `l`: Lowest price (24h)
- `h`: Highest price (24h)
- `o`: Open price
- `c`: Last price
- `p`: Price change
- `t`: Trade count (24h)

---

## Trading API (Private)

### Create Order

```go
req := cdcexchange.CreateOrderRequest{
    InstrumentName: "BTC_USDT",
    Side:           "BUY", // or "SELL"
    Type:           "LIMIT", // or "MARKET", "STOP_LIMIT"
    Price:          "50000",
    Quantity:       "0.1",
}
result, err := client.CreateOrder(ctx, req)
```

**Endpoint**: `private/create-order`

### Cancel Order

```go
err := client.CancelOrder(ctx, "BTC_USDT", "order-id")
```

**Endpoint**: `private/cancel-order`

### Get Open Orders

```go
result, err := client.GetOpenOrders(ctx, cdcexchange.GetOpenOrdersRequest{
    InstrumentName: "BTC_USDT",
})
```

**Endpoint**: `private/get-open-orders`

### Get Order History

```go
orders, err := client.GetOrderHistory(ctx, cdcexchange.GetOrderHistoryRequest{
    InstrumentName: "BTC_USDT",
})
```

**Endpoint**: `private/get-order-history`

### Get Trades

```go
trades, err := client.GetTrades(ctx, cdcexchange.GetTradesRequest{
    InstrumentName: "BTC_USDT",
})
```

**Endpoint**: `private/get-trades`

---

## Rate Limits

- Public endpoints: Not explicitly documented
- Private endpoints: Rate limited per UID

## Products

- **Spot**: Default
- **Margin**: Supported
- **Derivatives**: Supported (via derivatives transfer API)

## Reference

- [Crypto.com API Docs](https://exchange-docs.crypto.com)
- [SDK GitHub](https://github.com/cshep4/crypto-dot-com-exchange-go)
