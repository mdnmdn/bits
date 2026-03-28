# Kraken API Documentation

## Overview

- **Official exchange**: https://www.kraken.com
- **API Docs**: https://docs.kraken.com/rest/

## Go SDK

- **Community**: `Beldur/kraken-go-api-client`
- **SDK Location**: `/tmp/temp/kraken-go-api-client/`
- **Module**: `github.com/Beldur/kraken-go-api-client`

## SDK Structure

```
kraken-go-api-client/
├── krakenapi.go    # Main API client
├── types.go        # Data types
└── krakenapi_test.go
```

## Client Creation

```go
import "github.com/Beldur/kraken-go-api-client"

// Public client only
client := krakenapi.New("", "")

// Authenticated client
client := krakenapi.New(apiKey, apiSecret)

// With custom HTTP client
client := krakenapi.NewWithClient(apiKey, apiSecret, httpClient)
```

## Base URLs

| Environment | URL |
|-------------|-----|
| Production | `https://api.kraken.com` |

## Authentication

### Headers (for private endpoints)

```
API-Key: <api-key>
API-Sign: <signature>
```

### Signature

```
signature = Base64(HMAC-SHA512(secretKey, SHA256(nonce + post_data)))
```

---

## Market Data API (Public)

### Get Server Time

```go
time, err := client.Time()
```

**Endpoint**: `GET /0/public/Time`

**Response**:
```json
{
  "error": [],
  "result": {
    "unixtime": 1690196000,
    "rfc1123": "Fri, 24 Aug 2023 12:00:00 +0000"
  }
}
```

### Get Assets

```go
assets, err := client.Assets()
```

**Endpoint**: `GET /0/public/Assets`

Returns all available Kraken assets (e.g., XBT, ETH, USDT).

### Get Asset Pairs

```go
pairs, err := client.AssetPairs()
```

**Endpoint**: `GET /0/public/AssetPairs`

Returns all available asset pairs (e.g., XBT/USD, ETH/USD).

### Get Ticker

```go
ticker, err := client.Ticker("XBTUSD", "ETHUSD")
```

**Endpoint**: `GET /0/public/Ticker?pair=XBTUSD,ETHUSD`

**Response**:
```json
{
  "error": [],
  "result": {
    "XBTUSD": {
      "a": ["50000.0", "", "0.5"],   // Ask [price, whole lot volume, lot volume]
      "b": ["49999.0", "", "0.5"],   // Bid
      "c": ["50000.0", "0.1"],       // Last trade closed [price, volume]
      "v": ["1000", "2000"],         // Volume [today, last 24h]
      "p": ["49000", "48000"],       // Volume weighted avg price
      "t": [100, 200],               // Number of trades [today, last 24h]
      "l": ["49000", "47000"],       // Low [today, last 24h]
      "h": ["51000", "52000"],       // High [today, last 24h]
      "o": 49000                     // Today's opening price
    }
  }
}
```

### Get OHLC (Candles)

```go
// Default 1 minute
ohlc, err := client.OHLC("XBTUSD")

// With interval
ohlc, err := client.OHLCWithInterval("XBTUSD", "5") // 1, 5, 15, 30, 60, 240, 1440, 10080, 21600
```

**Endpoint**: `GET /0/public/OHLC?pair=XBTUSD&interval=5`

**Intervals**:
| Interval | Code |
|----------|------|
| 1m | `1` |
| 5m | `5` |
| 15m | `15` |
| 30m | `30` |
| 1h | `60` |
| 4h | `240` |
| 1D | `1440` |
| 1W | `10080` |
| 2W | `21600` |

**Response**:
```json
{
  "error": [],
  "result": {
    "XBTUSD": [
      [1690196000, 49000, 50000, 48000, 49000, 1000], // [time, open, high, low, close, volume]
      ...
    ],
    "last": 1690200000
  }
}
```

### Get Order Book

```go
orderbook, err := client.Depth("XBTUSD", 10)
```

**Endpoint**: `GET /0/public/Depth?pair=XBTUSD&count=10`

**Response**:
```json
{
  "error": [],
  "result": {
    "XBTUSD": {
      "asks": [[50000, "1.5", "100"], ...],  // [price, volume, timestamp]
      "bids": [[49999, "1.0", "100"], ...]
    }
  }
}
```

### Get Recent Trades

```go
trades, err := client.Trades("XBTUSD", 0) // 0 for all recent
```

**Endpoint**: `GET /0/public/Trades?pair=XBTUSD`

**Response**:
```json
{
  "error": [],
  "result": {
    "XBTUSD": [
      ["50000.0", "0.1", 1690196000.123, "b", "m", ""],
      // [price, volume, time, buy/sell, market/limit, misc]
    ],
    "last": 1690200000
  }
}
```

### Get Spread

```
// Available in SDK but requires additional implementation
// Endpoint: GET /0/public/Spread?pair=XBTUSD
```

---

## Private API (Authenticated)

### Get Balance

```go
balance, err := client.Balance()
```

**Endpoint**: `POST /0/private/Balance`

### Get Trade Balance

```go
tradeBalance, err := client.TradeBalance(map[string]string{})
```

**Endpoint**: `POST /0/private/TradeBalance`

### Get Open Orders

```go
orders, err := client.OpenOrders(map[string]string{})
```

**Endpoint**: `POST /0/private/OpenOrders`

### Get Closed Orders

```go
orders, err := client.ClosedOrders(map[string]string{})
```

**Endpoint**: `POST /0/private/ClosedOrders`

### Get Trades History

```go
trades, err := client.TradesHistory(0, 0, map[string]string{})
```

**Endpoint**: `POST /0/private/TradesHistory`

### Add Order

```go
result, err := client.AddOrder(
    "XBTUSD",    // pair
    "buy",       // direction
    "limit",     // order type
    "0.1",       // volume
    map[string]string{
        "price": "50000",
    },
)
```

**Endpoint**: `POST /0/private/AddOrder`

### Cancel Order

```go
result, err := client.CancelOrder("order-id")
```

**Endpoint**: `POST /0/private/CancelOrder`

---

## Rate Limits

- Public endpoints: Rate limited (varies by endpoint)
- Private endpoints: Rate limited

## Products

- **Spot**: Default (e.g., XBT/USD, ETH/USD)
- **Futures**: Available (via separate endpoints)
- **Margin**: Supported

## Notes

- Kraken uses non-standard asset codes:
  - XBT = Bitcoin (not BTC)
  - ZUSD = USD
  - XETH = ETH
  - etc.
- Minimum order sizes exist for various assets

## Reference

- [Kraken API Docs](https://docs.kraken.com/rest/)
- [SDK GitHub](https://github.com/Beldur/kraken-go-api-client)
