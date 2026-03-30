# MEXC API Documentation

MEXC is a global cryptocurrency exchange offering Spot, Futures, and ETF trading. The MEXC API provides comprehensive market data and trading capabilities through REST and WebSocket interfaces.

## Base URLs

| Type | URL |
|------|-----|
| REST API | `https://api.mexc.com` |
| WebSocket (Market) | `wss://wbs-api.mexc.com/ws` |
| WebSocket (User Data) | `wss://wbs-api.mexc.com/ws/v3` |

## Go SDK

MEXC provides an official Go SDK as part of their multi-language SDK package:

- **Repository**: https://github.com/mexcdevelop/mexc-api-sdk
- **Location**: `dist/go/` in the SDK
- **Alternative**: Community Go SDK at https://github.com/Sagleft/mexcsdk (MIT licensed)

### Go SDK Initialization

```go
package main

import (
	"fmt"
	"mexc-sdk/mexcsdk"
)

func main() {
	apiKey := "your-api-key"
	apiSecret := "your-api-secret"
	spot := mexcsdk.NewSpot(apiKey, apiSecret)
	
	// Get ticker
	ticker := spot.Ticker24hr("BTCUSDT")
	fmt.Println(ticker)
}
```

---

## Authentication

### Required Headers

For authenticated endpoints:

```
X-MEXC-APIKEY: <your-api-key>
Content-Type: application/json
```

### Signature Generation (HMAC-SHA256)

SIGNED endpoints require a signature parameter sent in the query string or request body.

```
signature = HMAC-SHA256(secretKey, totalParams)
```

Where `totalParams` = query string concatenated with request body.

**Important**: 
- Use millisecond timestamp (`timestamp` parameter)
- Default `recvWindow` is 5000ms (max 60000ms)
- Timestamp must be within `serverTime + 1000ms` and `serverTime + recvWindow`

### Timing Security

```
if (timestamp < (serverTime + 1000) && (serverTime - timestamp) <= recvWindow) {
    // process request
} else {
    // reject request
}
```

### Signature Example

```go
import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

func generateSignature(secretKey string, params string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(params))
	return hex.EncodeToString(h.Sum(nil))
}

func main() {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	params := fmt.Sprintf("symbol=BTCUSDT&side=BUY&type=LIMIT&quantity=1&price=11&recvWindow=5000&timestamp=%s", timestamp)
	signature := generateSignature("your-secret-key", params)
	fmt.Println("Signature:", signature)
}
```

---

## Rate Limits

### REST API Limits

| Limit Type | Value |
|------------|-------|
| IP-based | 300 requests / 10 seconds |
| UID-based | 500 requests / 10 seconds |

### WebSocket Limits

| Limit Type | Value |
|------------|-------|
| Requests | 100 requests / second |
| Max subscriptions | 30 streams per connection |
| Connection validity | 24 hours |
| Idle disconnect | 30 seconds (no subscription), 60 seconds (no data) |

### HTTP Return Codes

| Code | Description |
|------|-------------|
| 4XX | Malformed request (client error) |
| 403 | WAF limit violated |
| 429 | Rate limit exceeded |
| 5XX | Internal server error (retry) |

---

## Market Data Endpoints

All market data endpoints are public and don't require authentication.

### Test Connectivity

```
GET /api/v3/ping
```

Response: `{}`

Weight: 1

### Check Server Time

```
GET /api/v3/time
```

Response:
```json
{
  "serverTime": 1645539742000
}
```

Weight: 1

### Exchange Information

```
GET /api/v3/exchangeInfo
```

Query parameters:
- `symbol` - Single symbol (e.g., `BTCUSDT`)
- `symbols` - Comma-separated symbols (e.g., `BTCUSDT,ETHUSDT`)

Response includes: timezone, serverTime, rateLimits, symbols with status, precision, commission, filters

Weight: 25

### Order Book (Depth)

```
GET /api/v3/depth
```

Query parameters:
- `symbol` (required) - Trading pair
- `limit` (optional) - Number of results, default 100, max 5000

Response:
```json
{
  "lastUpdateId": 1112416,
  "bids": [["15.00000", "49999.00000"]],
  "asks": [["14.0000", "1.0000"]]
}
```

Weight: 3

### Recent Trades List

```
GET /api/v3/trades
```

Query parameters:
- `symbol` (required)
- `limit` (optional) - Default 500, max 1000

Response:
```json
[{
  "id": null,
  "price": "23",
  "qty": "0.478468",
  "quoteQty": "11.004764",
  "time": 1640830579240,
  "isBuyerMaker": true,
  "isBestMatch": true
}]
```

Weight: 5

### Aggregate Trades List

```
GET /api/v3/aggTrades
```

Query parameters:
- `symbol` (required)
- `startTime` (optional) - Timestamp in ms
- `endTime` (optional) - Timestamp in ms
- `limit` (optional) - Default 500, max 1000

Response:
```json
[{
  "a": null,
  "f": null,
  "l": null,
  "p": "46782.67",
  "q": "0.0038",
  "T": 1641380483000,
  "m": false,
  "M": true
}]
```

Weight: 1

### Kline/Candlestick Data

```
GET /api/v3/klines
```

Query parameters:
- `symbol` (required)
- `interval` (required) - 1m, 5m, 15m, 30m, 60m, 4h, 1d, 1w, 1M
- `startTime` (optional)
- `endTime` (optional)
- `limit` (optional) - Default 500, max 500

Response:
```json
[[
  1640804880000,    // Open time
  "47482.36",       // Open
  "47482.36",       // High
  "47416.57",       // Low
  "47436.1",        // Close
  "3.550717",       // Volume
  1640804940000,    // Close time
  "168387.3"        // Quote asset volume
]]
```

Weight: 1

### Current Average Price

```
GET /api/v3/avgPrice
```

Query parameters:
- `symbol` (required)

Response:
```json
{
  "mins": 5,
  "price": "9.35751834"
}
```

Weight: 1

### 24hr Ticker Price Change Statistics

```
GET /api/v3/ticker/24hr
```

Query parameters:
- `symbol` (optional) - If omitted, returns all symbols

Response:
```json
{
  "symbol": "BTCUSDT",
  "priceChange": "184.34",
  "priceChangePercent": "0.00400048",
  "prevClosePrice": "46079.37",
  "lastPrice": "46263.71",
  "bidPrice": "46260.38",
  "bidQty": "",
  "askPrice": "46260.41",
  "askQty": "",
  "openPrice": "46079.37",
  "highPrice": "47550.01",
  "lowPrice": "45555.5",
  "volume": "1732.461487",
  "quoteVolume": null,
  "openTime": 1641349500000,
  "closeTime": 1641349582808,
  "count": null
}
```

Weight: 25 (single symbol), 50 (all symbols)

### Symbol Price Ticker

```
GET /api/v3/ticker/price
```

Query parameters:
- `symbol` (optional)

Response:
```json
{
  "symbol": "BTCUSDT",
  "price": "184.34"
}
```

Weight: 10

### Symbol Order Book Ticker

```
GET /api/v3/ticker/bookTicker
```

Query parameters:
- `symbol` (optional)

Response:
```json
{
  "symbol": "BTCUSDT",
  "bidPrice": "46260.38",
  "bidQty": "1.5",
  "askPrice": "46260.41",
  "askQty": "2.3"
}
```

Weight: 10

---

## WebSocket Market Streams

MEXC WebSocket uses Protocol Buffers (protobuf) for data serialization.

### Connection

```
wss://wbs-api.mexc.com/ws
```

### Subscribe/Unubscribe

```json
// Subscribe
{
  "method": "SUBSCRIPTION",
  "params": ["spot@public.kline.v3.api.pb@BTCUSDT@Min15"]
}

// Unsubscribe
{
  "method": "UNSUBSCRIPTION",
  "params": ["spot@public.kline.v3.api.pb@BTCUSDT@Min15"]
}

// Ping
{"method": "PING"}

// Response
{
  "id": 0,
  "code": 0,
  "msg": "spot@public.kline.v3.api.pb@BTCUSDT@Min15"
}
```

### Stream Channels

| Channel | Description |
|---------|-------------|
| `spot@public.aggre.deals.v3.api.pb@100ms@<symbol>` | Trade streams (100ms or 10ms) |
| `spot@public.kline.v3.api.pb@<symbol>@<interval>` | K-line streams |
| `spot@public.aggre.depth.v3.api.pb@100ms@<symbol>` | Diff depth (incremental) |
| `spot@public.limit.depth.v3.api.pb@<symbol>@<level>` | Partial book depth (5, 10, 20 levels) |
| `spot@public.aggre.bookTicker.v3.api.pb@100ms@<symbol>` | Book ticker (best bid/ask) |
| `spot@public.miniTicker.v3.api.pb@<symbol>@<timezone>` | Mini ticker for symbol |
| `spot@public.miniTickers.v3.api.pb@<timezone>` | Mini tickers for all symbols |

### K-line Intervals

- `Min1`, `Min5`, `Min15`, `Min30`, `Min60`
- `Hour4`, `Hour8`
- `Day1`
- `Week1`
- `Month1`

### Trade Stream Response

```json
{
  "channel": "spot@public.aggre.deals.v3.api.pb@100ms@BTCUSDT",
  "publicdeals": {
    "dealsList": [{
      "price": "93220.00",
      "quantity": "0.04438243",
      "tradetype": 2,
      "time": 1736409765051
    }]
  },
  "symbol": "BTCUSDT",
  "sendtime": 1736409765052
}
```

### K-line Stream Response

```json
{
  "channel": "spot@public.kline.v3.api.pb@BTCUSDT@Min15",
  "publicspotkline": {
    "interval": "Min15",
    "windowstart": 1736410500,
    "openingprice": "92925",
    "closingprice": "93158.47",
    "highestprice": "93158.47",
    "lowestprice": "92800",
    "volume": "36.83803224",
    "amount": "3424811.05",
    "windowend": 1736411400
  },
  "symbol": "BTCUSDT",
  "symbolid": "2fb942154ef44a4ab2ef98c8afb6a4a7",
  "createtime": 1736410707571
}
```

### Diff Depth Stream Response

```json
{
  "channel": "spot@public.aggre.depth.v3.api.pb@100ms@BTCUSDT",
  "publicincreasedepths": {
    "asksList": [],
    "bidsList": [{
      "price": "92877.58",
      "quantity": "0.00000000"
    }],
    "eventtype": "spot@public.aggre.depth.v3.api.pb@100ms",
    "fromVersion": "10589632359",
    "toVersion": "10589632359"
  },
  "symbol": "BTCUSDT",
  "sendtime": 1736411507002
}
```

Note: If quantity is 0, the price level should be removed.

### Order Book Local Maintenance

1. Connect to WebSocket and subscribe to diff depth
2. Request snapshot: `GET /api/v3/depth?symbol=BTCUSDT&limit=5000`
3. Compare `lastUpdateId` with first `fromVersion`
4. Apply incremental updates sequentially
5. Reinitialize if version gap detected

---

## Error Codes

| Code | Description |
|------|-------------|
| -2011 | Unknown order sent |
| 400 | API key required |
| 401 | No authority |
| 403 | Access denied |
| 429 | Too many requests |
| 500 | Internal error |
| 503 | Service not available |
| 602 | Signature verification failed |
| 10001 | User does not exist |
| 10007 | Bad symbol |
| 10072 | Invalid access key |
| 10073 | Invalid Request-Time |
| 10101 | Insufficient balance |
| 30000 | Suspended transaction |
| 30016 | Trading disabled |
| 700001 | API-key format invalid |
| 700002 | Signature not valid |
| 700003 | Timestamp outside recvWindow |
| 700006 | IP not in whitelist |

---

## ENUM Definitions

### Order Side
- `BUY`
- `SELL`

### Order Type
- `LIMIT`
- `MARKET`
- `LIMIT_MAKER`
- `IMMEDIATE_OR_CANCEL`
- `FILL_OR_KILL`
- `STOP_MARKET_ORDER` (Query only)

### Order Status
- `NEW` - Uncompleted
- `FILLED` - Filled
- `PARTIALLY_FILLED` - Partially filled
- `CANCELED` - Canceled
- `PARTIALLY_CANCELED` - Partially canceled

### Kline Interval
- `1m`, `5m`, `15m`, `30m`, `60m`
- `4h`, `1d`, `1w`, `1M`

---

## Integration with bits CLI

For adding MEXC as a provider in bits:

1. Create `internal/provider/mexc/client.go` implementing `provider.Provider`
2. Implement capability interfaces: `PriceProvider`, `TickerProvider`, `CandleProvider`, `OrderBookProvider`
3. Return `model.Response[T]` with `Provider` and `Market` populated
4. Register in `internal/registry/registry.go`

### Market Data Mapping

| bits Capability | MEXC Endpoint |
|-----------------|---------------|
| Price | `/api/v3/ticker/price` |
| Ticker | `/api/v3/ticker/24hr` |
| OrderBook | `/api/v3/depth` |
| Candles | `/api/v3/klines` |
| Exchange Info | `/api/v3/exchangeInfo` |

### Symbol Format

MEXC uses uppercase symbol format with quote asset suffix (e.g., `BTCUSDT`, `ETHUSDT`)

### WebSocket for Streaming

For streaming capabilities, use `wss://wbs-api.mexc.com/ws` with protobuf deserialization

---

## API Key Setup

1. Go to [MEXC API Management](https://www.mexc.com/user/openapi)
2. Create API key with appropriate permissions
3. Set IP restrictions for security
4. Configure trading pair permissions

---

## Resources

- [MEXC API Documentation](https://www.mexc.com/api-docs/spot-v3/introduction)
- [Official SDK Repository](https://github.com/mexcdevelop/mexc-api-sdk)
- [WebSocket Proto Files](https://github.com/mexcdevelop/websocket-proto)
- [API Support](https://t.me/MEXCAPIsupport)
