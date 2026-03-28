# CoinMarketCap API Documentation

## Overview

- **Official website**: https://coinmarketcap.com
- **API Docs**: https://coinmarketcap.com/api/documentation/v1/
- **Pro API**: https://pro.coinmarketcap.com/api/v1

## Go SDK

- **Community**: `tigusigalpa/coinmarketcap-go` - Full CoinMarketCap Pro API support
- **Alternative**: `andyle182810/gocmc` - Lightweight Go SDK
- **Legacy**: `sharath/go-coinmarketcap` - Older client

## Base URLs

| Environment | URL |
|-------------|-----|
| Production (Pro) | `https://pro-api.coinmarketcap.com` |
| Sandbox | `https://sandbox.coinmarketcap.com` |

## Authentication

### Headers

```
X-CMC_PRO_API_KEY: <api-key>
Accept: application/json
Accept-Encoding: deflate, gzip
```

### Alternative (Query String)

```
https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest?CMC_PRO_API_KEY=<api-key>
```

**Note**: Header method is recommended for security (protects API key from being visible in URLs)

---

## API Categories

### 1. Cryptocurrency Endpoints

| Endpoint | Description |
|----------|-------------|
| `/v1/cryptocurrency/map` | Get cryptocurrency ID map |
| `/v1/cryptocurrency/info` | Get cryptocurrency metadata |
| `/v1/cryptocurrency/listings/latest` | List all cryptocurrencies (latest) |
| `/v1/cryptocurrency/listings/historical` | List cryptocurrencies (historical) |
| `/v1/cryptocurrency/quotes/latest` | Get market quotes (latest) |
| `/v1/cryptocurrency/quotes/historical` | Get market quotes (historical) |
| `/v1/cryptocurrency/ohlcv/historical` | Get OHLCV data |
| `/v1/cryptocurrency/market-pairs/latest` | Get market pairs |

### 2. Exchange Endpoints

| Endpoint | Description |
|----------|-------------|
| `/v1/exchange/map` | Get exchange ID map |
| `/v1/exchange/info` | Get exchange metadata |
| `/v1/exchange/listings/latest` | List all exchanges |
| `/v1/exchange/quotes/latest` | Get exchange quotes |
| `/v1/exchange/market-pairs/latest` | Get exchange market pairs |

### 3. Global Metrics Endpoints

| Endpoint | Description |
|----------|-------------|
| `/v1/global-metrics/quotes/latest` | Get global market metrics |
| `/v1/global-metrics/quotes/historical` | Get historical global metrics |

### 4. Tools Endpoints

| Endpoint | Description |
|----------|-------------|
| `/v1/tools/price-conversion` | Price conversion tool |

---

## Market Data API Examples

### Get Latest Cryptocurrency Listings

```
GET https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest?start=1&limit=10&convert=USD
```

**Parameters**:
- `start`: Pagination start (default: 1)
- `limit`: Number of results (max: 5000)
- `convert`: Currency to convert to (USD, EUR, BTC, etc.)
- `sort`: Sort field (market_cap, name, price, etc.)
- `cryptocurrency_type`: Filter (all, tokens, coins)

**Response**:
```json
{
  "status": {
    "timestamp": "2024-01-01T00:00:00.000Z",
    "credit_count": 1,
    "elapsed": 10
  },
  "data": [
    {
      "id": 1,
      "name": "Bitcoin",
      "symbol": "BTC",
      "slug": "bitcoin",
      "cmc_rank": 1,
      "num_market_pairs": 10000,
      "circulating_supply": 19000000,
      "total_supply": 21000000,
      "max_supply": 21000000,
      "quote": {
        "USD": {
          "price": 50000.00,
          "volume_24h": 50000000000,
          "market_cap": 950000000000,
          "percent_change_1h": 0.5,
          "percent_change_24h": 2.5,
          "percent_change_7d": 5.0,
          "last_updated": "2024-01-01T00:00:00.000Z"
        }
      }
    }
  ]
}
```

### Get Latest Quotes for Specific Cryptocurrencies

```
GET https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?id=1,1027&convert=USD,EUR
```

**Parameters**:
- `id`: Comma-separated CoinMarketCap IDs (e.g., 1=Bitcoin, 1027=Ethereum)
- `symbol`: Alternatively use symbols (BTC,ETH)
- `convert`: Multiple currencies (USD,EUR,GBP,BTC,ETH)

### Get Cryptocurrency OHLCV Historical

```
GET https://pro-api.coinmarketcap.com/v1/cryptocurrency/ohlcv/historical?id=1&interval=daily&time_start=2023-01-01&time_end=2023-12-31
```

**Parameters**:
- `id`: CoinMarketCap ID
- `interval`: Time interval (hourly, daily, weekly, monthly, yearly)
- `time_start`: Start date (ISO 8601 or Unix)
- `time_end`: End date
- `count`: Number of data points

### Get Market Pairs

```
GET https://pro-api.coinmarketcap.com/v1/cryptocurrency/market-pairs/latest?id=1&convert=USD
```

Returns all trading pairs for a cryptocurrency across exchanges.

### Get Cryptocurrency Metadata

```
GET https://pro-api.coinmarketcap.com/v1/cryptocurrency/info?id=1,1027
```

Returns:
- Logo URL
- Description
- Website URL
- Explorer URLs
- Tags
- Platform info (for tokens)

### Get Cryptocurrency ID Map

```
GET https://pro-api.coinmarketcap.com/v1/cryptocurrency/map?listing_status=active&start=0&limit=100
```

**Parameters**:
- `listing_status`: active, inactive, untracked
- `start`: Pagination start
- `limit`: Number of results

---

## Exchange API Examples

### Get Exchange Listings

```
GET https://pro-api.coinmarketcap.com/v1/exchange/listings/latest?limit=10&convert=USD
```

### Get Exchange Quotes

```
GET https://pro-api.coinmarketcap.com/v1/exchange/quotes/latest?id=270&convert=USD
```

### Get Exchange Market Pairs

```
GET https://pro-api.coinmarketcap.com/v1/exchange/market-pairs/latest?slug=binance&convert=USD
```

---

## Global Metrics API Examples

### Get Global Market Metrics

```
GET https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest?convert=USD
```

**Response**:
```json
{
  "data": {
    "BTC_dominance": 45.5,
    "ETH_dominance": 18.2,
    "active_cryptocurrencies": 10000,
    "active_market_pairs": 50000,
    "active_exchanges": 500,
    "total_market_cap": 2000000000000,
    "total_volume_24h": 100000000000,
    "quote": {
      "USD": {
        "total_market_cap": 2000000000000,
        "total_volume_24h": 100000000000
      }
    }
  }
}
```

---

## Tools API Examples

### Price Conversion

```
GET https://pro-api.coinmarketcap.com/v1/tools/price-conversion?symbol=BTC&amount=1&convert=USD,EUR,GBP
```

**Parameters**:
- `symbol`: Cryptocurrency symbol
- `amount`: Amount to convert
- `convert`: Target currencies

---

## Supported Fiat Currencies

| Currency | Code |
|----------|------|
| US Dollar | USD |
| Euro | EUR |
| British Pound | GBP |
| Japanese Yen | JPY |
| Chinese Yuan | CNY |
| Canadian Dollar | CAD |
| Australian Dollar | AUD |
| Swiss Franc | CHF |
| And more... | |

---

## Rate Limits

- **Free/Trial plans**: 10 calls/minute
- **Paid plans**: 60 calls/minute
- **Enterprise**: Custom limits

## Pricing Plans

| Plan | Monthly Credits | Price |
|------|----------------|-------|
| Starter | 10,000 | Free |
| Hobbyist | 100,000 | $29 |
| Startup | 1,000,000 | $79 |
| Enterprise | Unlimited | Custom |

**Note**: Paginated endpoints count as 1 credit + 1 credit per 100 items beyond default.

---

## Error Codes

| Code | Description |
|------|-------------|
| 400 | Bad Request - Invalid parameters |
| 401 | Unauthorized - Invalid API key |
| 403 | Forbidden - Plan restrictions |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error |

---

## Go SDK Usage Example

```go
import "github.com/tigusigalpa/coinmarketcap-go"

// Create client
client := coinmarketcap.NewClient("YOUR_API_KEY")

// Get latest listings
listings, err := client.Cryptocurrency.GetListingsLatest(
    coinmarketcap.ListingsLatestParams{
        Start: 1,
        Limit: 10,
        Convert: "USD",
    },
)

// Get quotes
quotes, err := client.Cryptocurrency.GetQuotesLatest(
    coinmarketcap.QuoteLatestParams{
        ID: "1,1027",
        Convert: "USD",
    },
)

// Get OHLCV
ohlcv, err := client.Cryptocurrency.GetOHLCVHistorical(
    coinmarketcap.OHLCVHistoricalParams{
        ID: "1",
        Interval: "daily",
        TimeStart: "2023-01-01",
        TimeEnd: "2023-12-31",
    },
)
```

---

## Reference

- [CoinMarketCap API Docs](https://coinmarketcap.com/api/documentation/v1/)
- [Pro API](https://pro.coinmarketcap.com/api/v1)
- [Pricing](https://coinmarketcap.com/api/pricing/)
- [Go SDK](https://github.com/tigusigalpa/coinmarketcap-go)
